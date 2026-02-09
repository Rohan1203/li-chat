package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"li-chat/pkg/logger"
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   int64
	username string
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type IncomingMessage struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

func (c *Client) readPump() {
	logger.Info("Read pump started for user", zap.String("username", c.username), zap.Int64("user_id", c.userID))
	logger.Debug("Setting up read deadline and handlers")

	defer func() {
		logger.Debug("Cleaning up - unregistering client and closing connection")
		c.hub.unregister <- c
		c.conn.Close()
		logger.Info("Read pump ended for user", zap.String("username", c.username), zap.Int64("user_id", c.userID))
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		logger.Debug("Pong message received from user", zap.String("username", c.username))
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket error from user", zap.String("username", c.username), zap.Int64("user_id", c.userID), zap.Error(err))
				logger.Warn("Unexpected connection close")
			} else {
				logger.Debug("WebSocket connection closed normally", zap.String("username", c.username), zap.Int64("user_id", c.userID))
			}
			break
		}

		logger.Debug("Raw WebSocket message received", zap.String("username", c.username), zap.Int("size", len(data)))

		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			logger.Warn("Failed to unmarshal message from user", zap.String("username", c.username), zap.Error(err))
			continue
		}

		logger.Debug("Message parsed successfully", zap.String("username", msg.Username), zap.Int("content_length", len(msg.Content)))
		logger.Debug("Forwarding message to hub handler")
		c.hub.handleMessage(msg)
	}
}

func (c *Client) writePump() {
	logger.Info("Write pump started for user", zap.String("username", c.username), zap.Int64("user_id", c.userID))
	logger.Debug("Setting up ping ticker", zap.Duration("period", pingPeriod))

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		logger.Debug("Cleaning up - stopping ticker and closing connection")
		ticker.Stop()
		c.conn.Close()
		logger.Info("Write pump ended for user", zap.String("username", c.username), zap.Int64("user_id", c.userID))
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				logger.Debug("Send channel closed by hub, sending close message")
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					logger.Error("Failed to send close message", zap.Error(err))
				}
				logger.Info("Connection closing initiated for user", zap.String("username", c.username))
				return
			}

			logger.Debug("Sending message to client", zap.String("username", c.username), zap.Int("message_size", len(message)))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Error("Failed to write message to user", zap.String("username", c.username), zap.Error(err))
				logger.Warn("Connection error - terminating write pump")
				return
			}
			logger.Debug("Message sent successfully to user", zap.String("username", c.username))

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			logger.Debug("Sending ping message to user", zap.String("username", c.username))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("Failed to write ping message to user", zap.String("username", c.username), zap.Error(err))
				logger.Warn("Ping send failed - terminating connection")
				return
			}
			logger.Debug("Ping message sent to user", zap.String("username", c.username))
		}
	}
}
