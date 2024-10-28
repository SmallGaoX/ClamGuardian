package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "停止运行中的 ClamGuardian",
	RunE:  runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	pid, err := getPID()
	if err != nil {
		return err
	}

	if !isProcessRunning(pid) {
		return fmt.Errorf("进程未运行")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("无法找到进程: %v", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("无法停止进程: %v", err)
	}

	fmt.Println("ClamGuardian 已停止")
	return nil
}
