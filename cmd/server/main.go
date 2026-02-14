package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"li-chat/internal/config"
	"li-chat/internal/db"
	"li-chat/internal/httpserver"
	"li-chat/internal/websocket"
	"li-chat/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	logger.Debug("Initializing logger")
	logger.Init(logger.Config{
		Level:      "debug",
		Filename:   "logs/chat-app.log",
		MaxSizeMB:  50,
		MaxBackups: 10,
		MaxAgeDays: 30,
		Compress:   true,
	})
	defer logger.Sync()

	logger.Debug("Initiating database connection")
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		logger.Error("DATABASE_URL environment variable is not set")
		panic("DATABASE_URL is required")
	}
	repo, err := db.NewRepository(dbConnStr)
	if err != nil {
		logger.Error("Failed to initialize database repository", zap.Error(err))
		logger.Warn("Application cannot start without database connection")
		panic(err)
	}
	logger.Info("Logger initialized successfully")

	logger.Debug("Creating and starting WebSocket hub")
	hub := websocket.NewHub(repo)
	go hub.Run()

	logger.Debug("Setting up HTTP routes and handlers")
	router := httpserver.NewRouter(hub, repo)
	server := httpserver.New(cfg, router)
	logger.Info("HTTP server initialized", zap.String("port", cfg.Port))

	logger.Debug("Starting HTTP server in background")
	go func() {
		logger.Info("Server starting", zap.String("port", cfg.Port))
		if err := server.Start(); err != nil {
			logger.Error("Server encountered error", zap.Error(err))
			logger.Warn("HTTP server stopped")
		}
	}()
	logger.Debug("Application is running. Press Ctrl+C to shutdown")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received")
	logger.Info("Initiating graceful shutdown")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	logger.Debug("Closing HTTP server gracefully")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Error during server shutdown", zap.Error(err))
		logger.Warn("Server shutdown with errors")
	} else {
		logger.Info("Server shutdown completed successfully")
	}
}
