package httpserver

import (
	"net/http"

	"li-chat/internal/db"
	"li-chat/internal/websocket"
	"li-chat/pkg/logger"
)

func NewRouter(hub *websocket.Hub, repo *db.Repository) http.Handler {
	mux := http.NewServeMux()
	auth := NewAuthHandler(repo)

	mux.HandleFunc("/ws", websocket.HandleWS(hub, repo))

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		logger.Info("Health check endpoint called")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/register", auth.Register)
	mux.HandleFunc("/login", auth.Login)
	mux.HandleFunc("/logout", auth.Logout)
	mux.HandleFunc("/whoami", auth.WhoAmI)

	// Serve embedded web assets
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	return mux
}
