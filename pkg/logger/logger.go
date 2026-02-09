package logger

import (
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

// Structured logging methods for production-grade logging
func Debug(msg string, fields ...zap.Field) {
	if log == nil {
		return
	}
	log.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	if log == nil {
		return
	}
	log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	if log == nil {
		return
	}
	log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	if log == nil {
		return
	}
	log.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	if log == nil {
		return
	}
	log.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	if log == nil {
		return
	}
	log.Fatal(msg, fields...)
}
