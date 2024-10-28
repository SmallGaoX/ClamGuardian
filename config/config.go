package config

import (
	"ClamGuardian/internal/matcher"
	"fmt"
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
		MemoryLimit int64 `mapstructure:"memory_limit"`
		BufferSize  int   `mapstructure:"buffer_size"`
	} `mapstructure:"system"`
}

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("无法解析配置: %v", err)
	}

	// 验证必要的配置
	if len(config.Monitor.Paths) == 0 {
		return nil, fmt.Errorf("未指定监控路径")
	}

	return &config, nil
}
