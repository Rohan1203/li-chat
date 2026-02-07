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
)

func main() {
	logger.Info("[APP::START] ============== Chat Application Starting ==============")
	logger.Debug("[APP::CONFIG] Loading application configuration...")
	cfg := config.Load()
	logger.Info("[APP::CONFIG] Configuration loaded successfully")

	logger.Debug("[APP::LOGGER] Initializing structured logger...")
	logger.Init(logger.Config{
		Level:      "debug",
		Filename:   "logs/chat-app.log",
		MaxSizeMB:  50,
		MaxBackups: 10,
		MaxAgeDays: 30,
		Compress:   true,
	})
	defer logger.Sync()
	logger.Info("[APP::LOGGER] Logger initialized successfully")

	logger.Debug("[APP::DB] Establishing database connection...")
	repo, err := db.NewRepository("chat.db")
	if err != nil {
		logger.Error("[APP::DB] Failed to initialize database repository: %v", err)
		logger.Warn("[APP::DB] Application cannot start without database connection")
		panic(err)
	}
	logger.Info("[APP::DB] Database repository initialized successfully")

	logger.Debug("[APP::WS] Creating and starting WebSocket hub...")
	hub := websocket.NewHub(repo)
	go hub.Run()
	logger.Info("[APP::WS] WebSocket hub started")

	logger.Debug("[APP::HTTP] Setting up HTTP routes and handlers...")
	router := httpserver.NewRouter(hub, repo)
	server := httpserver.New(cfg, router)
	logger.Info("[APP::HTTP] HTTP server initialized on port %s", cfg.Port)

	logger.Debug("[APP::HTTP] Starting HTTP server in background...")
	go func() {
		logger.Info("[APP::HTTP] Server starting on %s", cfg.Port)
		if err := server.Start(); err != nil {
			logger.Error("[APP::HTTP] Server encountered error: %v", err)
			logger.Warn("[APP::HTTP] HTTP server stopped")
		}
	}()

	logger.Info("[APP::START] ============== Chat Application Ready ==============")
	logger.Info("[APP::RUN] Application is running. Press Ctrl+C to shutdown...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("[APP::SIGNAL] Shutdown signal received")
	logger.Info("[APP::SHUTDOWN] Initiating graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	logger.Debug("[APP::SHUTDOWN] Closing HTTP server gracefully...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("[APP::SHUTDOWN] Error during server shutdown: %v", err)
		logger.Warn("[APP::SHUTDOWN] Server shutdown with errors")
	} else {
		logger.Info("[APP::SHUTDOWN] Server shutdown completed successfully")
	}

	logger.Info("[APP::END] ============== Chat Application Terminated ==============")
}
