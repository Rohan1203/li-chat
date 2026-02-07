package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *zap.Logger

type Config struct {
	Level      string
	Filename   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
}

func Init(cfg Config) {
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	})

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder

	level := zapcore.InfoLevel
	_ = level.Set(cfg.Level)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.NewMultiWriteSyncer(
			writer,
			zapcore.AddSync(os.Stdout),
		),
		level,
	)

	log = zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(2), // IMPORTANT for correct caller
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

func Debug(format string, args ...any) {
	if log == nil {
		return
	}
	log.Debug(fmt.Sprintf(format, args...))
}

func Info(format string, args ...any) {
	if log == nil {
		return
	}
	log.Info(fmt.Sprintf(format, args...))
}

func Warn(format string, args ...any) {
	if log == nil {
		return
	}
	log.Warn(fmt.Sprintf(format, args...))
}

func Error(format string, args ...any) {
	if log == nil {
		return
	}
	log.Error(fmt.Sprintf(format, args...))
}
