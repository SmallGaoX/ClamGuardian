package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"ClamGuardian/config"
	"ClamGuardian/internal/logger"
	"ClamGuardian/internal/matcher"
	"ClamGuardian/internal/monitor"
	"ClamGuardian/internal/position"
	"ClamGuardian/internal/status"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	// 配置文件路径
	cfgFile string

	// Monitor 配置
	monitorPaths    []string
	monitorPatterns []string

	// Position 配置
	positionStorePath      string
	positionUpdateInterval int

	// System 配置
	systemMemoryLimit int64
	systemBufferSize  int

	// Matcher 配置
	matcherRules  []string
	matcherLevels []string
)

var rootCmd = &cobra.Command{
	Use:   "clamguardian",
	Short: "日志文件监控和告警工具",
	RunE:  run,
}

func init() {
	cobra.OnInitialize(initConfig)

	// 配置文件
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认为 ./config.yaml)")

	// Monitor 配置
	rootCmd.Flags().StringSliceVar(&monitorPaths, "paths", []string{}, "要监控的目录路径列表")
	rootCmd.Flags().StringSliceVar(&monitorPatterns, "patterns", []string{"*.log"}, "文件匹配模式列表")

	// Position 配置
	rootCmd.Flags().StringVar(&positionStorePath, "position-store", "positions.json", "位置存储文件路径")
	rootCmd.Flags().IntVar(&positionUpdateInterval, "position-interval", 5, "位置更新间隔(秒)")

	// System 配置
	rootCmd.Flags().Int64Var(&systemMemoryLimit, "memory-limit", 100, "内存使用限制(MB)")
	rootCmd.Flags().IntVar(&systemBufferSize, "buffer-size", 4096, "读取缓冲区大小(bytes)")

	// Matcher 配置
	rootCmd.Flags().StringSliceVar(&matcherRules, "rules", []string{}, "匹配规则列表")
	rootCmd.Flags().StringSliceVar(&matcherLevels, "levels", []string{}, "规则对应的告警级别")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("使用配置文件: %s\n", viper.ConfigFileUsed())
	}

	// 命令行参数覆盖配置文件
	if len(monitorPaths) > 0 {
		viper.Set("monitor.paths", monitorPaths)
	}
	if len(monitorPatterns) > 0 {
		viper.Set("monitor.patterns", monitorPatterns)
	}
	if positionStorePath != "positions.json" {
		viper.Set("position.store_path", positionStorePath)
	}
	if positionUpdateInterval != 5 {
		viper.Set("position.update_interval", positionUpdateInterval)
	}
	if systemMemoryLimit != 100 {
		viper.Set("system.memory_limit", systemMemoryLimit)
	}
	if systemBufferSize != 4096 {
		viper.Set("system.buffer_size", systemBufferSize)
	}

	// 处理匹配规则
	if len(matcherRules) > 0 && len(matcherRules) == len(matcherLevels) {
		rules := make([]matcher.MatchRule, len(matcherRules))
		for i := range matcherRules {
			rules[i] = matcher.MatchRule{
				Pattern: matcherRules[i],
				Level:   matcherLevels[i],
			}
		}
		viper.Set("matcher.rules", rules)
	}
}

func Execute() error {
	return rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 初始化日志系统
	logger.InitLogger(
		cfg.Log.Path,
		cfg.Log.MaxSize,
		cfg.Log.MaxBackups,
		cfg.Log.MaxAge,
	)
	defer logger.Logger.Sync()

	// 现在可以安全地使用 logger
	logger.Logger.Info("应用启动",
		zap.Strings("monitor_paths", cfg.Monitor.Paths),
		zap.Strings("patterns", cfg.Monitor.Patterns))

	// 创建位置管理器
	pm, err := position.NewManager(cfg.Position.StorePath, cfg.Position.UpdateInterval)
	if err != nil {
		return fmt.Errorf("创建位置管理器失败: %v", err)
	}

	// 创建匹配器
	m, err := matcher.NewMatcher(cfg.Matcher.Rules, cfg.System.BufferSize)
	if err != nil {
		return fmt.Errorf("创建匹配器失败: %v", err)
	}

	// 创建监控器
	mon, err := monitor.NewMonitor(cfg.Monitor.Paths, cfg.Monitor.Patterns, m, pm, cfg.System.BufferSize)
	if err != nil {
		return fmt.Errorf("创建监控器失败: %v", err)
	}

	// 启动内存监控
	go monitorMemory(cfg.System.MemoryLimit)

	// 创建一个带取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动文件监控
	if err := mon.Start(ctx); err != nil {
		return fmt.Errorf("启动监控失败: %v", err)
	}

	// 启动状态监控
	statusMonitor, err := status.NewMonitor(
		time.Duration(cfg.Status.Interval)*time.Second,
		cfg,
	)
	if err != nil {
		logger.Logger.Error("创建状态监控失败", zap.Error(err))
		return fmt.Errorf("创建状态监控失败: %v", err)
	}
	statusMonitor.Start()
	defer statusMonitor.Stop()

	// 如果启用了指标收集，启动 Prometheus 服务器
	if cfg.Metrics.Enabled {
		go func() {
			http.Handle(cfg.Metrics.Path, promhttp.Handler())
			addr := fmt.Sprintf(":%d", cfg.Metrics.Port)
			logger.Logger.Info("启动metrics服务",
				zap.String("address", addr),
				zap.String("path", cfg.Metrics.Path))
			if err := http.ListenAndServe(addr, nil); err != nil {
				logger.Logger.Error("metrics服务器启动失败", zap.Error(err))
			}
		}()
	}

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logger.Logger.Info("接收到退出信号", zap.String("signal", sig.String()))

	return nil
}

// monitorMemory 监控内存使用
func monitorMemory(limit int64) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)

			if stats.Alloc > uint64(limit*1024*1024) {
				fmt.Printf("警告: 内存使用超过限制 (当前: %dMB, 限制: %dMB)\n",
					stats.Alloc/(1024*1024), limit)
				runtime.GC()
			}

		case <-time.After(5 * time.Second):
			return
		}
	}
}
