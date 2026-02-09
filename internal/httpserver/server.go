package httpserver

import (
	"context"
	"net/http"

	"li-chat/internal/config"
	"li-chat/pkg/logger"

	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg *config.Config, handler http.Handler) *Server {
	logger.Debug("Creating HTTP server", zap.String("address", cfg.Port))
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Port,
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	}
}

func (s *Server) Start() error {
	logger.Info("Starting HTTP server")
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Error("HTTP server error", zap.Error(err))
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("Initiating graceful server shutdown")
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		logger.Error("Error during server shutdown", zap.Error(err))
	} else {
		logger.Info("Server shutdown completed successfully")
	}
	return err
}
