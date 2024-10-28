package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	// 尝试连接到运行中的实例获取状态
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", 2112)) // 使用默认的 metrics 端口
	if err != nil {
		return fmt.Errorf("无法连接到运行中的实例: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("获取状态失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 解析 metrics 响应
	var (
		memoryUsage float64
		cpuUsage    float64
		fileCount   int
		matchCount  int64
	)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// 跳过注释行
		if strings.HasPrefix(line, "#") {
			continue
		}

		// 解析指标行
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			continue
		}

		name := parts[0]
		valueStr := parts[1]
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}

		switch name {
		case "clamguardian_memory_usage_bytes":
			memoryUsage = value
		case "clamguardian_cpu_usage_percent":
			cpuUsage = value
		case "clamguardian_processed_files_total":
			fileCount = int(value)
		case "clamguardian_rule_matches_total":
			matchCount = int64(value)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取metrics响应失败: %v", err)
	}

	// 输出状态信息到控制台
	fmt.Println("=== ClamGuardian 运行状态 ===")
	fmt.Printf("内存使用: %.2f MB\n", memoryUsage/(1024*1024))
	fmt.Printf("CPU使用率: %.1f%%\n", cpuUsage)
	fmt.Printf("监控文件数: %d\n", fileCount)
	fmt.Printf("规则匹配数: %d\n", matchCount)
	fmt.Printf("最后更新时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return nil
}
