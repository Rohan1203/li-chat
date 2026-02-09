package config

import (
	"time"

	"li-chat/pkg/logger"

	"go.uber.org/zap"
)

type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func Load() *Config {
	logger.Debug("Loading application configuration")
	cfg := &Config{
		Port:            ":8080",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
	logger.Debug("Configuration loaded",
		zap.String("port", cfg.Port),
		zap.Duration("read_timeout", cfg.ReadTimeout),
		zap.Duration("write_timeout", cfg.WriteTimeout),
		zap.Duration("shutdown_timeout", cfg.ShutdownTimeout))
	return cfg
}
