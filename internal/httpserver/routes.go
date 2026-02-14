package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"li-chat/internal/auth"
	"li-chat/internal/db"
	"li-chat/internal/websocket"
)

func NewRouter(hub *websocket.Hub, repo *db.Repository) http.Handler {
	mux := http.NewServeMux()
	authHandler := NewAuthHandler(repo)

	mux.HandleFunc("/ws", websocket.HandleWS(hub, repo))

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/logout", authHandler.Logout)
	mux.HandleFunc("/whoami", authHandler.WhoAmI)
	mux.HandleFunc("/refresh-token", authHandler.RefreshToken)
	// mux.HandleFunc("/messages", authHandler.LoadMessage())

	// Serve embedded web assets properly
	webFS := getWebFS()
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	return mux
}

func getMessages(repo *db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Verify JWT token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		_, err := auth.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch messages (limit to last 100)
		messages, err := repo.GetMessages(100)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}
