package cmd

import (
	"bufio"
	"encoding/json"
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

	// 输出基本状态信息
	fmt.Println("=== ClamGuardian 运行状态 ===")
	fmt.Printf("内存使用: %.2f MB\n", memoryUsage/(1024*1024))
	fmt.Printf("CPU使用率: %.1f%%\n", cpuUsage)
	fmt.Printf("监控文件数: %d\n", fileCount)
	fmt.Printf("规则匹配数: %d\n", matchCount)
	fmt.Printf("最后更新时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// 获取文件监控状态
	fmt.Println("\n=== 文件监控状态 ===")
	fileResp, err := http.Get(fmt.Sprintf("http://localhost:%d/files", 2112))
	if err != nil {
		fmt.Println("无法获取文件监控状态")
		return nil // 继续显示其他信息
	}
	defer fileResp.Body.Close()

	if fileResp.StatusCode != http.StatusOK {
		fmt.Println("获取文件监控状态失败")
		return nil
	}

	var files []struct {
		Filename string  `json:"filename"`
		Position int64   `json:"position"`
		Size     int64   `json:"size"`
		Progress float64 `json:"progress"`
	}

	if err := json.NewDecoder(fileResp.Body).Decode(&files); err != nil {
		fmt.Println("解析文件监控状态失败")
		return nil
	}

	// 输出文件状态
	if len(files) == 0 {
		fmt.Println("当前没有监控的文件")
		return nil
	}

	fmt.Printf("\n%-50s %-15s %-15s %s\n", "文件名", "当前位置", "文件大小", "进度")
	fmt.Println(strings.Repeat("-", 90))
	for _, file := range files {
		filename := file.Filename
		if len(filename) > 50 {
			filename = "..." + filename[len(filename)-47:]
		}
		fmt.Printf("%-50s %-15s %-15s %.1f%%\n",
			filename,
			formatBytes(file.Position),
			formatBytes(file.Size),
			file.Progress*100,
		)
	}

	return nil
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(bytes)/float64(div), "KMGTPE"[exp])
}
