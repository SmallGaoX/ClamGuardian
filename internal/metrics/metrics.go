package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// MemoryUsage 内存使用指标
	MemoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "clamguardian_memory_usage_bytes",
		Help: "当前应用程序使用的内存量(字节)",
	})

	// CPUUsage CPU使用指标
	CPUUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "clamguardian_cpu_usage_percent",
		Help: "当前应用程序的CPU使用率",
	})

	// ProcessedFiles 已处理文件数
	ProcessedFiles = promauto.NewCounter(prometheus.CounterOpts{
		Name: "clamguardian_processed_files_total",
		Help: "已处理的文件总数",
	})

	// RuleMatches 匹配规则命中数
	RuleMatches = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "clamguardian_rule_matches_total",
			Help: "规则匹配命中总数",
		},
		[]string{"level"},
	)
)
