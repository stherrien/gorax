package websocket

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID            string
	TenantID      string
	Conn          *websocket.Conn
	Hub           *Hub
	Send          chan []byte
	Subscriptions map[string]bool
	mu            sync.RWMutex
}

// Hub manages all WebSocket connections and message broadcasting
type Hub struct {
	// Registered clients by client ID
	clients map[string]*Client

	// Clients grouped by subscription room
	rooms map[string]map[string]*Client

	// Register client (exported for handler)
	Register chan *Client

	// Unregister client (exported for handler)
	Unregister chan *Client

	// Broadcast message to specific room
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex

	logger *slog.Logger
}

// BroadcastMessage represents a message to broadcast to a room
type BroadcastMessage struct {
	Room    string
	Message []byte
}

// NewHub creates a new WebSocket hub
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
		logger:     logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToRoom(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client
	h.logger.Info("client registered",
		"client_id", client.ID,
		"tenant_id", client.TenantID,
	)
}

// unregisterClient removes a client and cleans up their subscriptions
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[client.ID]; exists {
		delete(h.clients, client.ID)

		// Remove from all rooms
		client.mu.RLock()
		for room := range client.Subscriptions {
			if clients, exists := h.rooms[room]; exists {
				delete(clients, client.ID)
				if len(clients) == 0 {
					delete(h.rooms, room)
				}
			}
		}
		client.mu.RUnlock()

		close(client.Send)

		h.logger.Info("client unregistered",
			"client_id", client.ID,
			"tenant_id", client.TenantID,
		)
	}
}

// SubscribeClient subscribes a client to a room
func (h *Hub) SubscribeClient(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	defer client.mu.Unlock()

	// Add to client's subscriptions
	client.Subscriptions[room] = true

	// Add to room's clients
	if _, exists := h.rooms[room]; !exists {
		h.rooms[room] = make(map[string]*Client)
	}
	h.rooms[room][client.ID] = client

	h.logger.Info("client subscribed to room",
		"client_id", client.ID,
		"room", room,
	)
}

// UnsubscribeClient unsubscribes a client from a room
func (h *Hub) UnsubscribeClient(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	defer client.mu.Unlock()

	// Remove from client's subscriptions
	delete(client.Subscriptions, room)

	// Remove from room's clients
	if clients, exists := h.rooms[room]; exists {
		delete(clients, client.ID)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}

	h.logger.Info("client unsubscribed from room",
		"client_id", client.ID,
		"room", room,
	)
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(room string, message []byte) {
	h.broadcast <- &BroadcastMessage{
		Room:    room,
		Message: message,
	}
}

// broadcastToRoom performs the actual broadcast
func (h *Hub) broadcastToRoom(msg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, exists := h.rooms[msg.Room]; exists {
		h.logger.Debug("broadcasting to room",
			"room", msg.Room,
			"client_count", len(clients),
		)

		for _, client := range clients {
			select {
			case client.Send <- msg.Message:
			default:
				// Client's send channel is full, skip
				h.logger.Warn("client send channel full, dropping message",
					"client_id", client.ID,
					"room", msg.Room,
				)
			}
		}
	}
}

// Client read and write pumps

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.Hub.logger.Warn("failed to set read deadline", "error", err, "client_id", c.ID)
	}
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	c.Conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error("websocket error", "error", err, "client_id", c.ID)
			}
			break
		}

		// Handle incoming messages (subscriptions, etc.)
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.Hub.logger.Warn("failed to set write deadline", "error", err, "client_id", c.ID)
				return
			}
			if !ok {
				// The hub closed the channel
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{}) //nolint:errcheck // best effort close
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				c.Hub.logger.Warn("failed to write message", "error", err, "client_id", c.ID)
				return
			}

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				if _, err := w.Write([]byte{'\n'}); err != nil {
					break
				}
				if _, err := w.Write(<-c.Send); err != nil {
					break
				}
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from clients
func (c *Client) handleMessage(message []byte) {
	// Parse message and handle subscription requests
	// Message format: {"type": "subscribe", "room": "execution:123"}
	// For now, we'll handle subscriptions through HTTP params
	// This can be extended later for dynamic subscriptions
}
