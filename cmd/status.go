package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"ClamGuardian/internal/status"
	"github.com/spf13/cobra"
)

type StatusInfo struct {
	Timestamp   time.Time `json:"timestamp"`
	MemoryUsage uint64    `json:"memory_usage"`
	CPUPercent  float64   `json:"cpu_percent"`
	NumFiles    int       `json:"num_files"`
	NumMatches  int64     `json:"num_matches"`
	UptimeHours float64   `json:"uptime_hours"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "显示当前运行状态",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// 创建状态监控器
	monitor, err := status.NewMonitor(time.Second)
	if err != nil {
		return fmt.Errorf("创建状态监控器失败: %v", err)
	}

	// 获取当前状态
	info, err := monitor.GetCurrentStatus()
	if err != nil {
		return fmt.Errorf("获取状态信息失败: %v", err)
	}

	// 将状态信息格式化为 JSON
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("格式化状态信息失败: %v", err)
	}

	fmt.Println(string(data))
	return nil
}
