package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "重启 ClamGuardian",
	RunE:  runRestart,
}

func init() {
	rootCmd.AddCommand(restartCmd)
}

func runRestart(cmd *cobra.Command, args []string) error {
	// 先停止当前实例
	if err := runStop(cmd, args); err != nil {
		fmt.Printf("停止服务时出错: %v\n", err)
	}

	// 等待进程完全停止
	time.Sleep(2 * time.Second)

	// 启动新实例
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法获取可执行文件路径: %v", err)
	}

	// 构建启动命令
	command := exec.Command(executable)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// 启动新进程
	if err := command.Start(); err != nil {
		return fmt.Errorf("启动新实例失败: %v", err)
	}

	// 保存新的 PID
	pidFile := "/var/run/clamguardian.pid"
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", command.Process.Pid)), 0644); err != nil {
		return fmt.Errorf("无法写入PID文件: %v", err)
	}

	fmt.Println("ClamGuardian 已重启")
	return nil
}
