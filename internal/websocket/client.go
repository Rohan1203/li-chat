package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"

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
	logger.Info("[CLIENT::RPUMP] Read pump started for user: %s (ID: %d)", c.username, c.userID)
	logger.Debug("[CLIENT::RPUMP] Setting up read deadline and handlers")

	defer func() {
		logger.Debug("[CLIENT::RPUMP] Cleaning up - unregistering client and closing connection")
		c.hub.unregister <- c
		c.conn.Close()
		logger.Info("[CLIENT::RPUMP] Read pump ended for user: %s (ID: %d)", c.username, c.userID)
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		logger.Debug("[CLIENT::PING] Pong message received from user: %s", c.username)
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("[CLIENT::RPUMP] WebSocket error from user %s (ID: %d): %v", c.username, c.userID, err)
				logger.Warn("[CLIENT::RPUMP] Unexpected connection close")
			} else {
				logger.Debug("[CLIENT::RPUMP] WebSocket connection closed normally - User: %s (ID: %d)", c.username, c.userID)
			}
			break
		}

		logger.Debug("[CLIENT::RCV] Raw WebSocket message received - User: %s, Size: %d bytes", c.username, len(data))

		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			logger.Warn("[CLIENT::PARSE] Failed to unmarshal message from user %s: %v", c.username, err)
			continue
		}

		logger.Debug("[CLIENT::PARSE] Message parsed successfully - From: %s, Content length: %d bytes", msg.Username, len(msg.Content))
		logger.Debug("[CLIENT::HANDLE] Forwarding message to hub handler")
		c.hub.handleMessage(msg)
	}
}

func (c *Client) writePump() {
	logger.Info("[CLIENT::WPUMP] Write pump started for user: %s (ID: %d)", c.username, c.userID)
	logger.Debug("[CLIENT::WPUMP] Setting up ping ticker with period: %v", pingPeriod)

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		logger.Debug("[CLIENT::WPUMP] Cleaning up - stopping ticker and closing connection")
		ticker.Stop()
		c.conn.Close()
		logger.Info("[CLIENT::WPUMP] Write pump ended for user: %s (ID: %d)", c.username, c.userID)
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				logger.Debug("[CLIENT::WPUMP] Send channel closed by hub, sending close message")
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					logger.Error("[CLIENT::WPUMP] Failed to send close message: %v", err)
				}
				logger.Info("[CLIENT::WPUMP] Connection closing initiated for user: %s", c.username)
				return
			}

			logger.Debug("[CLIENT::SEND] Sending message to client - User: %s, Size: %d bytes", c.username, len(message))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Error("[CLIENT::SEND] Failed to write message to user %s: %v", c.username, err)
				logger.Warn("[CLIENT::SEND] Connection error - terminating write pump")
				return
			}
			logger.Debug("[CLIENT::SEND] Message sent successfully to user: %s", c.username)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			logger.Debug("[CLIENT::PING] Sending ping message to user: %s", c.username)
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("[CLIENT::PING] Failed to write ping message to user %s: %v", c.username, err)
				logger.Warn("[CLIENT::PING] Ping send failed - terminating connection")
				return
			}
			logger.Debug("[CLIENT::PING] Ping message sent to user: %s", c.username)
		}
	}
}
