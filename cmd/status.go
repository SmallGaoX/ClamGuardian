package cmd

import (
	"fmt"
	"time"

	"ClamGuardian/config"
	"ClamGuardian/internal/status"
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
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 创建状态监控器
	monitor, err := status.NewMonitor(time.Second, cfg)
	if err != nil {
		return fmt.Errorf("创建状态监控器失败: %v", err)
	}

	// 获取当前状态
	info, err := monitor.GetCurrentStatus()
	if err != nil {
		return fmt.Errorf("获取状态信息失败: %v", err)
	}

	// 格式化输出状态信息
	fmt.Println("=== ClamGuardian 运行状态 ===")
	fmt.Printf("运行时间: %.1f 小时\n", info.UptimeHours)
	fmt.Printf("内存使用: %.2f MB\n", float64(info.MemoryUsage)/(1024*1024))
	fmt.Printf("CPU使用率: %.1f%%\n", info.CPUPercent)
	fmt.Printf("监控文件数: %d\n", info.NumFiles)
	fmt.Printf("规则匹配数: %d\n", info.NumMatches)
	fmt.Printf("最后更新时间: %s\n", info.Timestamp.Format("2006-01-02 15:04:05"))

	// 如果启用了指标服务
	if info.MetricsEnabled {
		fmt.Printf("指标服务地址: http://localhost:%d%s\n", info.MetricsPort, info.MetricsPath)
	}

	return nil
}
