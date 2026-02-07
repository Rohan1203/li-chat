package websocket

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"li-chat/internal/auth"
	"li-chat/internal/db"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // tighten this in production
	},
}

func HandleWS(hub *Hub, repo *db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// AUTH CHECK - Extract JWT from Authorization header or query parameter
		var token string

		// Try Authorization header first
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		// Fall back to query parameter
		if token == "" {
			token = r.URL.Query().Get("token")
		}

		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate JWT token
		claims, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &Client{
			hub:      hub,
			conn:     conn,
			send:     make(chan []byte, 256),
			userID:   claims.UserID,
			username: claims.Username,
		}

		hub.register <- client

		go client.writePump()
		go client.readPump()
	}
}
