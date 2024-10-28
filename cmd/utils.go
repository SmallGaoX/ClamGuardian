package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"ClamGuardian/config"
)

var defaultPidFile = "/var/run/clamguardian.pid"

// getPidFile 获取PID文件路径
func getPidFile() string {
	// 如果是 stop 或 restart 命令，先尝试从配置文件获取
	cfg, err := config.LoadConfig()
	if err == nil && cfg.System.PidFile != "" {
		return cfg.System.PidFile
	}
	return defaultPidFile
}

// isProcessRunning 检查指定 PID 的进程是否在运行
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 发送空信号来检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// getPID 从 PID 文件中读取进程 ID
func getPID() (int, error) {
	pidFile := getPidFile()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, fmt.Errorf("无法读取PID文件(%s): %v", pidFile, err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return 0, fmt.Errorf("无效的PID文件内容: %v", err)
	}

	return pid, nil
}

// writePID 写入PID到文件
func writePID(pidFile string) error {
	// 确保目录存在
	dir := filepath.Dir(pidFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建PID文件目录失败: %v", err)
	}

	pid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return fmt.Errorf("写入PID文件失败: %v", err)
	}
	return nil
}
