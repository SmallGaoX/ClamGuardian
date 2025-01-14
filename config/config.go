package config

import (
	"fmt"

	"ClamGuardian/internal/matcher"
	"github.com/spf13/viper"
)

type Config struct {
	Monitor struct {
		Paths    []string `mapstructure:"paths"`
		Patterns []string `mapstructure:"patterns"`
	} `mapstructure:"monitor"`
	Matcher struct {
		Rules []matcher.MatchRule `mapstructure:"rules"`
	} `mapstructure:"matcher"`
	Position struct {
		StorePath      string `mapstructure:"store_path"`
		UpdateInterval int    `mapstructure:"update_interval"`
	} `mapstructure:"position"`
	System struct {
		MemoryLimit int64  `mapstructure:"memory_limit"`
		BufferSize  int    `mapstructure:"buffer_size"`
		PidFile     string `mapstructure:"pid_file"` // 新增：PID文件路径
	} `mapstructure:"system"`
	Metrics struct {
		Enabled bool   `mapstructure:"enabled"`
		Port    int    `mapstructure:"port"`
		Path    string `mapstructure:"path"`
	} `mapstructure:"metrics"`
	Log struct {
		Path       string `mapstructure:"path"`
		Format     string `mapstructure:"format"` // 新增：日志格式配置
		Level      string `mapstructure:"level"`  // 新增：日志级别配置
		MaxSize    int    `mapstructure:"max_size"`
		MaxBackups int    `mapstructure:"max_backups"`
		MaxAge     int    `mapstructure:"max_age"`
	} `mapstructure:"log"`
	Status struct {
		Interval int `mapstructure:"interval"` // 状态收集间隔(秒)
	} `mapstructure:"status"`
}

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("无法解析配置: %v", err)
	}

	// 设置默认PID文件路径
	if config.System.PidFile == "" {
		config.System.PidFile = "/var/run/clamguardian.pid"
	}

	// 验证必要的配置
	if len(config.Monitor.Paths) == 0 {
		return nil, fmt.Errorf("未指定监控路径")
	}

	return &config, nil
}
