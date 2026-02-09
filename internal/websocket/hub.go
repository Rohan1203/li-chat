package websocket

import (
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"li-chat/internal/db"
	"li-chat/internal/model"
	"li-chat/pkg/logger"
)

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	repo       *db.Repository
}

func NewHub(repo *db.Repository) *Hub {
	logger.Debug("Initializing WebSocket hub")
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		repo:       repo,
	}
}

func (h *Hub) Run() {
	logger.Info("WebSocket hub started and running")
	logger.Debug("Hub event loop initialized")

	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
			logger.Debug("Client registered", zap.String("username", c.username), zap.Int64("user_id", c.userID))
			logger.Info("Connected clients updated", zap.Int("count", len(h.clients)))

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				logger.Debug("Client unregistered", zap.String("username", c.username), zap.Int64("user_id", c.userID))
				logger.Info("Connected clients updated", zap.Int("count", len(h.clients)))
			} else {
				logger.Warn("Attempted to unregister non-existent client")
			}
		}
	}
}

func (h *Hub) handleMessage(msg IncomingMessage) {
	logger.Info("Handling incoming message", zap.String("username", msg.Username))
	logger.Debug("Message details", zap.Int("content_length", len(msg.Content)))

	if msg.Username == "" {
		logger.Warn("Empty username in message")
		return
	}

	if msg.Content == "" {
		logger.Warn("Empty content in message", zap.String("username", msg.Username))
		return
	}

	logger.Debug("Getting or creating user", zap.String("username", msg.Username))
	userID, err := h.repo.GetOrCreateUser(msg.Username)
	if err != nil {
		logger.Error("Error getting or creating user", zap.String("username", msg.Username), zap.Error(err))
		logger.Warn("Message discarded due to user operation failure")
		return
	}

	logger.Debug("Saving message for user", zap.String("username", msg.Username), zap.Int64("user_id", userID))
	err = h.repo.SaveMessage(userID, msg.Content)
	if err != nil {
		logger.Error("Error saving message", zap.String("username", msg.Username), zap.Error(err))
		logger.Warn("Message save failed - broadcast cancelled")
		return
	}
	logger.Debug("Message persisted successfully")

	logger.Debug("Preparing message broadcast", zap.Int("client_count", len(h.clients)))
	out := model.Message{
		Username:  msg.Username,
		Content:   msg.Content,
		CreatedAt: time.Now().Format("15:04"),
	}

	data, err := json.Marshal(out)
	if err != nil {
		logger.Error("Error marshaling message", zap.Error(err))
		logger.Warn("Broadcast cancelled due to JSON marshaling error")
		return
	}
	logger.Debug("Message serialized successfully", zap.Int("payload_size", len(data)))

	var sentCount int
	var failedCount int

	for c := range h.clients {
		select {
		case c.send <- data:
			sentCount++
			logger.Debug("Message sent to client", zap.String("username", c.username))
		default:
			failedCount++
			logger.Warn("Failed to send message to client", zap.String("username", c.username))
		}
	}

	logger.Info("Message broadcasted", zap.String("username", msg.Username), zap.Int("sent", sentCount), zap.Int("failed", failedCount), zap.Int("total_clients", len(h.clients)))

	if failedCount > 0 && failedCount == len(h.clients) {
		logger.Error("Broadcast failed for all connected clients")
	}
}
