package cmd

import (
	"fmt"
	"strings"
	"time"

	"ClamGuardian/config"
	"ClamGuardian/internal/metrics"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "显示当前运行状态",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	stateManager := metrics.GetStateManager()
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 基本信息
	fmt.Println("\n=== ClamGuardian 运行状态 ===")
	fmt.Printf("配置文件: %s\n", stateManager.GetConfigPath())
	fmt.Printf("运行时间: %s\n", time.Since(stateManager.GetStartTime()).Round(time.Second))

	// 监控配置
	fmt.Println("\n=== 监控配置 ===")
	fmt.Printf("监控路径: %v\n", stateManager.GetMonitoringPaths())
	fmt.Printf("文件模式: %v\n", cfg.Monitor.Patterns)

	// 系统配置
	fmt.Println("\n=== 系统配置 ===")
	fmt.Printf("内存限制: %dMB\n", cfg.System.MemoryLimit)
	fmt.Printf("缓冲大小: %d bytes\n", cfg.System.BufferSize)
	fmt.Printf("PID文件: %s\n", cfg.System.PidFile)

	// 位置管理配置
	fmt.Println("\n=== 位置管理 ===")
	fmt.Printf("存储路径: %s\n", cfg.Position.StorePath)
	fmt.Printf("更新间隔: %d秒\n", cfg.Position.UpdateInterval)

	// 指标配置
	fmt.Println("\n=== 指标配置 ===")
	fmt.Printf("指标启用: %v\n", cfg.Metrics.Enabled)
	if cfg.Metrics.Enabled {
		fmt.Printf("指标端口: %d\n", cfg.Metrics.Port)
		fmt.Printf("指标路径: %s\n", cfg.Metrics.Path)
	}

	// 日志配置
	fmt.Println("\n=== 日志配置 ===")
	fmt.Printf("日志路径: %s\n", cfg.Log.Path)
	fmt.Printf("日志格式: %s\n", cfg.Log.Format)
	fmt.Printf("日志级别: %s\n", cfg.Log.Level)
	fmt.Printf("单文件大小限制: %dMB\n", cfg.Log.MaxSize)
	fmt.Printf("最大备份数: %d\n", cfg.Log.MaxBackups)
	fmt.Printf("保留天数: %d\n", cfg.Log.MaxAge)

	// 匹配规则
	fmt.Println("\n=== 匹配规则 ===")
	if len(cfg.Matcher.Rules) == 0 {
		fmt.Println("未配置匹配规则")
	} else {
		for i, rule := range cfg.Matcher.Rules {
			fmt.Printf("%d. 模式: %s, 级别: %s\n", i+1, rule.Pattern, rule.Level)
		}
	}

	// 系统状态
	memory, cpu := stateManager.GetSystemMetrics()
	fmt.Println("\n=== 系统状态 ===")
	fmt.Printf("内存使用: %s\n", formatBytes(memory))
	fmt.Printf("CPU使用率: %.2f%%\n", cpu)

	// 匹配统计
	matches := stateManager.GetTotalMatches()
	fmt.Println("\n=== 匹配统计 ===")
	if len(matches) == 0 {
		fmt.Println("暂无匹配记录")
	} else {
		for level, count := range matches {
			fmt.Printf("%s级别: %d次\n", level, count)
		}
	}

	// 文件监控状态
	files := stateManager.GetAllFileStatus()
	fmt.Println("\n=== 文件监控状态 ===")
	if len(files) == 0 {
		fmt.Println("当前没有监控的文件")
		return nil
	}

	fmt.Printf("\n%-50s %-15s %-15s %-10s %-20s %-10s\n",
		"文件名", "当前位置", "文件大小", "进度", "最后修改时间", "匹配数")
	fmt.Println(strings.Repeat("-", 120))

	for _, file := range files {
		filename := file.Filename
		if len(filename) > 50 {
			filename = "..." + filename[len(filename)-47:]
		}
		fmt.Printf("%-50s %-15s %-15s %-10.1f%% %-20s %-10d\n",
			filename,
			formatBytes(file.Position),
			formatBytes(file.Size),
			file.Progress*100,
			file.LastModified.Format("2006-01-02 15:04:05"),
			file.MatchCount)
	}

	return nil
}

// formatBytes 格式化字节数，支持 uint64 和 int64
func formatBytes(bytes interface{}) string {
	var bytesInt uint64

	switch v := bytes.(type) {
	case uint64:
		bytesInt = v
	case int64:
		bytesInt = uint64(v)
	default:
		return "unknown"
	}

	const unit = 1024
	if bytesInt < unit {
		return fmt.Sprintf("%d B", bytesInt)
	}

	div, exp := uint64(unit), 0
	for n := bytesInt / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB",
		float64(bytesInt)/float64(div), "KMGTPE"[exp])
}
