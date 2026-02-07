package websocket

import (
	"encoding/json"
	"time"

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
	logger.Info("[HUB::START] WebSocket hub started and running")
	logger.Debug("[HUB::LOOP] Hub event loop initialized")

	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
			logger.Debug("[HUB::REG] Client registered - User: %s (ID: %d)", c.username, c.userID)
			logger.Info("[HUB::STATS] Total connected clients: %d", len(h.clients))

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				logger.Debug("[HUB::UNREG] Client unregistered - User: %s (ID: %d)", c.username, c.userID)
				logger.Info("[HUB::STATS] Total connected clients: %d", len(h.clients))
			} else {
				logger.Warn("[HUB::UNREG] Attempted to unregister non-existent client")
			}
		}
	}
}

func (h *Hub) handleMessage(msg IncomingMessage) {
	logger.Info("[HUB::MSG] Handling incoming message from user: %s", msg.Username)
	logger.Debug("[HUB::MSG] Message content length: %d bytes", len(msg.Content))

	if msg.Username == "" {
		logger.Warn("[HUB::MSG] Empty username in message")
		return
	}

	if msg.Content == "" {
		logger.Warn("[HUB::MSG] Empty content in message from user: %s", msg.Username)
		return
	}

	logger.Debug("[HUB::DB] Getting or creating user: %s", msg.Username)
	userID, err := h.repo.GetOrCreateUser(msg.Username)
	if err != nil {
		logger.Error("[HUB::DB] Error getting or creating user %s: %v", msg.Username, err)
		logger.Warn("[HUB::MSG] Message discarded due to user operation failure")
		return
	}

	logger.Debug("[HUB::DB] Saving message for user %s (ID: %d)", msg.Username, userID)
	err = h.repo.SaveMessage(userID, msg.Content)
	if err != nil {
		logger.Error("[HUB::DB] Error saving message for user %s: %v", msg.Username, err)
		logger.Warn("[HUB::MSG] Message save failed - broadcast cancelled")
		return
	}
	logger.Debug("[HUB::DB] Message persisted successfully")

	logger.Debug("[HUB::BROADCAST] Preparing message broadcast to %d connected clients", len(h.clients))
	out := model.Message{
		Username:  msg.Username,
		Content:   msg.Content,
		CreatedAt: time.Now().Format("15:04"),
	}

	data, err := json.Marshal(out)
	if err != nil {
		logger.Error("[HUB::JSON] Error marshaling message: %v", err)
		logger.Warn("[HUB::BROADCAST] Broadcast cancelled due to JSON marshaling error")
		return
	}
	logger.Debug("[HUB::JSON] Message serialized successfully, payload size: %d bytes", len(data))

	var sentCount int
	var failedCount int

	for c := range h.clients {
		select {
		case c.send <- data:
			sentCount++
			logger.Debug("[HUB::SEND] Message sent to client - User: %s", c.username)
		default:
			failedCount++
			logger.Warn("[HUB::SEND] Failed to send message to client - User: %s (send channel full)", c.username)
		}
	}

	logger.Info("[HUB::BROADCAST] Message from %s broadcasted - Sent: %d, Failed: %d, Total: %d", msg.Username, sentCount, failedCount, len(h.clients))

	if failedCount > 0 && failedCount == len(h.clients) {
		logger.Error("[HUB::BROADCAST] Broadcast failed for all connected clients")
	}
}
