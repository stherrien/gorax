package collaboration

import (
	"encoding/json"
	"log/slog"
	"sync"

	ws "github.com/gorax/gorax/internal/websocket"
)

// Hub manages collaboration WebSocket connections and broadcasts
type Hub struct {
	service *Service
	wsHub   *ws.Hub
	clients map[string]map[string]*ws.Client // workflowID -> userID -> client
	mu      sync.RWMutex
	logger  *slog.Logger
}

// NewHub creates a new collaboration hub
func NewHub(service *Service, wsHub *ws.Hub, logger *slog.Logger) *Hub {
	return &Hub{
		service: service,
		wsHub:   wsHub,
		clients: make(map[string]map[string]*ws.Client),
		logger:  logger,
	}
}

// RegisterClient registers a client for collaboration on a workflow
func (h *Hub) RegisterClient(workflowID, userID string, client *ws.Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[workflowID] == nil {
		h.clients[workflowID] = make(map[string]*ws.Client)
	}

	h.clients[workflowID][userID] = client

	h.logger.Info("collaboration client registered",
		"workflow_id", workflowID,
		"user_id", userID,
		"client_id", client.ID,
	)
}

// UnregisterClient unregisters a client from collaboration
func (h *Hub) UnregisterClient(workflowID, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[workflowID] != nil {
		delete(h.clients[workflowID], userID)

		if len(h.clients[workflowID]) == 0 {
			delete(h.clients, workflowID)
		}
	}

	h.logger.Info("collaboration client unregistered",
		"workflow_id", workflowID,
		"user_id", userID,
	)
}

// BroadcastToWorkflow broadcasts a message to all collaborators on a workflow
func (h *Hub) BroadcastToWorkflow(workflowID string, message WebSocketMessage) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("failed to marshal message", "error", err)
		return err
	}

	// Broadcast to workflow room
	room := "collaboration:" + workflowID
	h.wsHub.BroadcastToRoom(room, data)

	h.logger.Debug("broadcast to workflow",
		"workflow_id", workflowID,
		"message_type", message.Type,
		"client_count", len(h.clients[workflowID]),
	)

	return nil
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(workflowID, userID string, message WebSocketMessage) error {
	h.mu.RLock()
	client := h.clients[workflowID][userID]
	h.mu.RUnlock()

	if client == nil {
		h.logger.Warn("client not found for user",
			"workflow_id", workflowID,
			"user_id", userID,
		)
		return nil
	}

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("failed to marshal message", "error", err)
		return err
	}

	select {
	case client.Send <- data:
		h.logger.Debug("message sent to user",
			"workflow_id", workflowID,
			"user_id", userID,
			"message_type", message.Type,
		)
	default:
		h.logger.Warn("client send channel full",
			"workflow_id", workflowID,
			"user_id", userID,
		)
	}

	return nil
}

// BroadcastToOthers broadcasts a message to all collaborators except the sender
func (h *Hub) BroadcastToOthers(workflowID, excludeUserID string, message WebSocketMessage) error {
	h.mu.RLock()
	clients := h.clients[workflowID]
	h.mu.RUnlock()

	if clients == nil {
		return nil
	}

	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("failed to marshal message", "error", err)
		return err
	}

	count := 0
	for userID, client := range clients {
		if userID == excludeUserID {
			continue
		}

		select {
		case client.Send <- data:
			count++
		default:
			h.logger.Warn("client send channel full",
				"workflow_id", workflowID,
				"user_id", userID,
			)
		}
	}

	h.logger.Debug("broadcast to others",
		"workflow_id", workflowID,
		"exclude_user_id", excludeUserID,
		"message_type", message.Type,
		"sent_count", count,
	)

	return nil
}

// GetClientCount returns the number of clients connected to a workflow
func (h *Hub) GetClientCount(workflowID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.clients[workflowID] == nil {
		return 0
	}

	return len(h.clients[workflowID])
}

// GetService returns the underlying collaboration service
func (h *Hub) GetService() *Service {
	return h.service
}
