package status

import (
	"context"
	"os"
	"time"

	"ClamGuardian/config"
	"ClamGuardian/internal/logger"
	"ClamGuardian/internal/metrics"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
)

// Monitor 状态监控器
type Monitor struct {
	proc     *process.Process
	ctx      context.Context
	cancel   context.CancelFunc
	interval time.Duration
	config   *config.Config // 添加配置信息
}

// NewMonitor 创建新的状态监控器
func NewMonitor(interval time.Duration, cfg *config.Config) (*Monitor, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		proc:     proc,
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
		config:   cfg,
	}, nil
}

// Start 开始监控
func (m *Monitor) Start() {
	go m.collect()
}

// Stop 停止监控
func (m *Monitor) Stop() {
	m.cancel()
}

// collect 收集系统指标
func (m *Monitor) collect() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// 收集内存使用情况
			memInfo, err := m.proc.MemoryInfo()
			if err == nil {
				metrics.MemoryUsage.Set(float64(memInfo.RSS))
				logger.Logger.Debug("内存使用情况更新",
					zap.Uint64("rss", memInfo.RSS),
					zap.Uint64("vms", memInfo.VMS))
			}

			// 收集CPU使用情况
			cpuPercent, err := m.proc.CPUPercent()
			if err == nil {
				metrics.CPUUsage.Set(cpuPercent)
				logger.Logger.Debug("CPU使用情况更新",
					zap.Float64("cpu_percent", cpuPercent))
			}

			// 收集系统CPU使用情况
			if systemCPU, err := cpu.Percent(0, false); err == nil && len(systemCPU) > 0 {
				logger.Logger.Debug("系统CPU使用情况",
					zap.Float64("system_cpu_percent", systemCPU[0]))
			}
		}
	}
}

// StatusInfo 状态信息结构
type StatusInfo struct {
	Timestamp      time.Time
	MemoryUsage    uint64
	CPUPercent     float64
	NumFiles       int
	NumMatches     int64
	UptimeHours    float64
	MetricsEnabled bool
	MetricsPort    int
	MetricsPath    string
}

// GetCurrentStatus 获取当前状态信息
func (m *Monitor) GetCurrentStatus() (*StatusInfo, error) {
	info := &StatusInfo{
		Timestamp:      time.Now(),
		MetricsEnabled: m.config.Metrics.Enabled,
		MetricsPort:    m.config.Metrics.Port,
		MetricsPath:    m.config.Metrics.Path,
	}

	// 获取内存使用情况
	if memInfo, err := m.proc.MemoryInfo(); err == nil {
		info.MemoryUsage = memInfo.RSS
	}

	// 获取 CPU 使用情况
	if cpuPercent, err := m.proc.CPUPercent(); err == nil {
		info.CPUPercent = cpuPercent
	}

	// 获取运行时间
	if createTime, err := m.proc.CreateTime(); err == nil {
		uptime := time.Since(time.Unix(createTime/1000, 0))
		info.UptimeHours = uptime.Hours()
	}

	// 获取监控文件数和匹配数
	// 这里需要从 monitor 和 matcher 组件获取相关统计信息
	// TODO: 实现文件数和匹配数的统计

	return info, nil
}
