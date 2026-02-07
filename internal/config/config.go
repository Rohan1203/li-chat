package config

import (
	"time"

	"li-chat/pkg/logger"
)

type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func Load() *Config {
	logger.Debug("Loading application configuration...")
	cfg := &Config{
		Port:            ":8080",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
	logger.Debug("Configuration loaded: Port=%s, ReadTimeout=%v, WriteTimeout=%v, ShutdownTimeout=%v",
		cfg.Port, cfg.ReadTimeout, cfg.WriteTimeout, cfg.ShutdownTimeout)
	return cfg
}
