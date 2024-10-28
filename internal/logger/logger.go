package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// InitLogger 初始化日志配置
func InitLogger(logPath string, maxSize, maxBackups, maxAge int) {
	// 配置 lumberjack
	w := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    maxSize,    // 每个日志文件最大尺寸（MB）
		MaxBackups: maxBackups, // 保留的旧文件最大数量
		MaxAge:     maxAge,     // 保留的旧文件最大天数
		Compress:   true,       // 是否压缩旧文件
	}

	// 配置 encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建 core
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(w),
		zap.InfoLevel,
	)

	// 创建 logger
	Logger = zap.New(core, zap.AddCaller())
}
