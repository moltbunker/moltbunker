package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/moltbunker/moltbunker/internal/logging"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, check origin properly
		return true
	},
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Channel string      `json:"channel,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	hub          *WebSocketHub
	conn         *websocket.Conn
	send         chan []byte
	subscribed   map[string]bool
	mu           sync.RWMutex
}

// WebSocketHub manages WebSocket clients and broadcasting
type WebSocketHub struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan *WebSocketMessage
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mu         sync.RWMutex
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan *WebSocketMessage, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Close all clients
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			logging.Debug("WebSocket client connected",
				"total_clients", len(h.clients),
				logging.Component("websocket"))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			logging.Debug("WebSocket client disconnected",
				"total_clients", len(h.clients),
				logging.Component("websocket"))

		case msg := <-h.broadcast:
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				// Check if client is subscribed to this channel
				if msg.Channel != "" {
					client.mu.RLock()
					subscribed := client.subscribed[msg.Channel]
					client.mu.RUnlock()
					if !subscribed {
						continue
					}
				}

				select {
				case client.send <- data:
				default:
					// Client's buffer is full, close connection
					h.mu.RUnlock()
					h.mu.Lock()
					close(client.send)
					delete(h.clients, client)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all clients subscribed to the channel
func (h *WebSocketHub) Broadcast(eventType string, data interface{}) {
	msg := &WebSocketMessage{
		Type: eventType,
		Data: data,
	}

	select {
	case h.broadcast <- msg:
	default:
		logging.Warn("WebSocket broadcast buffer full",
			logging.Component("websocket"))
	}
}

// BroadcastToChannel sends a message to clients subscribed to a specific channel
func (h *WebSocketHub) BroadcastToChannel(channel string, eventType string, data interface{}) {
	msg := &WebSocketMessage{
		Type:    eventType,
		Channel: channel,
		Data:    data,
	}

	select {
	case h.broadcast <- msg:
	default:
		logging.Warn("WebSocket broadcast buffer full",
			"channel", channel,
			logging.Component("websocket"))
	}
}

// ClientCount returns the number of connected clients
func (h *WebSocketHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// newWebSocketClient creates a new WebSocket client
func newWebSocketClient(hub *WebSocketHub, conn *websocket.Conn) *WebSocketClient {
	return &WebSocketClient{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		subscribed: make(map[string]bool),
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *WebSocketClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logging.Debug("WebSocket read error",
					"error", err.Error(),
					logging.Component("websocket"))
			}
			break
		}

		// Parse message
		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Handle subscription requests
		c.handleMessage(&msg)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second) // Ping interval
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming WebSocket messages
func (c *WebSocketClient) handleMessage(msg *WebSocketMessage) {
	switch msg.Type {
	case "subscribe":
		c.handleSubscribe(msg)
	case "unsubscribe":
		c.handleUnsubscribe(msg)
	case "ping":
		c.sendMessage(&WebSocketMessage{Type: "pong"})
	}
}

// handleSubscribe handles subscription requests
func (c *WebSocketClient) handleSubscribe(msg *WebSocketMessage) {
	// Extract channels from data
	var req struct {
		Channels []string `json:"channels"`
	}

	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	if channels, ok := data["channels"].([]interface{}); ok {
		for _, ch := range channels {
			if channel, ok := ch.(string); ok {
				c.mu.Lock()
				c.subscribed[channel] = true
				c.mu.Unlock()

				logging.Debug("WebSocket client subscribed",
					"channel", channel,
					logging.Component("websocket"))
			}
		}
	}

	// Also handle if channels is in the main data
	if jsonData, err := json.Marshal(msg.Data); err == nil {
		json.Unmarshal(jsonData, &req)
		for _, channel := range req.Channels {
			c.mu.Lock()
			c.subscribed[channel] = true
			c.mu.Unlock()
		}
	}

	// Send confirmation
	c.sendMessage(&WebSocketMessage{
		Type: "subscribed",
		Data: map[string]interface{}{
			"channels": c.getSubscribedChannels(),
		},
	})
}

// handleUnsubscribe handles unsubscription requests
func (c *WebSocketClient) handleUnsubscribe(msg *WebSocketMessage) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	if channels, ok := data["channels"].([]interface{}); ok {
		for _, ch := range channels {
			if channel, ok := ch.(string); ok {
				c.mu.Lock()
				delete(c.subscribed, channel)
				c.mu.Unlock()
			}
		}
	}

	// Send confirmation
	c.sendMessage(&WebSocketMessage{
		Type: "unsubscribed",
		Data: map[string]interface{}{
			"channels": c.getSubscribedChannels(),
		},
	})
}

// sendMessage sends a message to the client
func (c *WebSocketClient) sendMessage(msg *WebSocketMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.send <- data:
	default:
		// Buffer full
	}
}

// getSubscribedChannels returns the list of subscribed channels
func (c *WebSocketClient) getSubscribedChannels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	channels := make([]string, 0, len(c.subscribed))
	for ch := range c.subscribed {
		channels = append(channels, ch)
	}
	return channels
}

// handleWebSocket handles WebSocket upgrade requests
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Warn("WebSocket upgrade failed",
			"error", err.Error(),
			logging.Component("websocket"))
		return
	}

	client := newWebSocketClient(s.wsHub, conn)
	s.wsHub.register <- client

	// Start pumps
	go client.writePump()
	go client.readPump()
}
