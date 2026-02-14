package httpserver

import (
	"embed"
	"io/fs"
	"li-chat/pkg/logger"

	"go.uber.org/zap"
)

//go:embed web/*
var webFS embed.FS

func getWebFS() fs.FS {
	fsys, err := fs.Sub(webFS, "web")
	if err != nil {
		logger.Fatal("Error: ", zap.Error(err))
	}
	return fsys
}
