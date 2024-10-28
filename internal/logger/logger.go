package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// LogFormat 日志格式类型
type LogFormat string

const (
	FormatText LogFormat = "text"
	FormatJSON LogFormat = "json"
)

// LogConfig 日志配置
type LogConfig struct {
	Path       string    // 日志文件路径
	Format     LogFormat // 日志格式：text 或 json
	MaxSize    int       // 每个日志文件的最大大小（MB）
	MaxBackups int       // 保留的旧文件最大数量
	MaxAge     int       // 保留的旧文件最大天数
	Level      string    // 日志级别：debug, info, warn, error
}

func init() {
	// 创建一个默认的控制台 logger
	Logger, _ = zap.NewDevelopment()
}

// InitLogger 初始化日志配置
func InitLogger(config LogConfig) error {
	// 配置 lumberjack
	w := &lumberjack.Logger{
		Filename:   config.Path,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   true,
	}

	// 设置日志级别
	var level zapcore.Level
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 根据日志级别配置 encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	// 只在 debug 级别时添加 caller 和 func 信息
	if level == zapcore.DebugLevel {
		encoderConfig.CallerKey = "caller"
		encoderConfig.FunctionKey = "func"
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	// 根据配置选择编码器
	var encoder zapcore.Encoder
	switch config.Format {
	case FormatJSON:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default: // FormatText
		if level == zapcore.DebugLevel {
			encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		} else {
			encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		}
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建 core
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(w)),
		level,
	)

	// 创建 logger，只在 debug 级别时添加 caller
	var options []zap.Option
	if level == zapcore.DebugLevel {
		options = append(options,
			zap.AddCaller(),
			zap.AddCallerSkip(1),
		)
	}
	// 添加错误级别的堆栈跟踪
	options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))

	// 创建 logger
	Logger = zap.New(core, options...)

	return nil
}
