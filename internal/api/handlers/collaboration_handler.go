package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/collaboration"
	"github.com/gorax/gorax/internal/config"
	ws "github.com/gorax/gorax/internal/websocket"
)

// CollaborationHandler handles real-time collaboration WebSocket connections
type CollaborationHandler struct {
	collabHub *collaboration.Hub
	wsHub     *ws.Hub
	logger    *slog.Logger
	upgrader  websocket.Upgrader
	wsConfig  config.WebSocketConfig
}

// NewCollaborationHandler creates a new collaboration handler
func NewCollaborationHandler(collabHub *collaboration.Hub, wsHub *ws.Hub, wsConfig config.WebSocketConfig, logger *slog.Logger) *CollaborationHandler {
	return &CollaborationHandler{
		collabHub: collabHub,
		wsHub:     wsHub,
		logger:    logger,
		wsConfig:  wsConfig,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     wsConfig.CheckOrigin(),
		},
	}
}

// HandleWorkflowCollaboration handles WebSocket connections for workflow collaboration
func (h *CollaborationHandler) HandleWorkflowCollaboration(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user
	user := middleware.GetUser(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "id")

	if workflowID == "" {
		http.Error(w, "workflow_id required", http.StatusBadRequest)
		return
	}

	// Validate workflow ID format (prevent injection attacks)
	if !isValidID(workflowID) {
		h.logger.Warn("invalid workflow ID format", "workflow_id", workflowID, "user_id", user.ID)
		http.Error(w, "invalid workflow_id format", http.StatusBadRequest)
		return
	}

	// Check connection limit per workflow
	currentConnections := h.collabHub.GetClientCount(workflowID)
	if currentConnections >= h.wsConfig.MaxConnectionsPerWorkflow {
		h.logger.Warn("workflow connection limit exceeded",
			"workflow_id", workflowID,
			"current_connections", currentConnections,
			"max_connections", h.wsConfig.MaxConnectionsPerWorkflow,
		)
		http.Error(w, "workflow connection limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade connection", "error", err)
		return
	}

	// Create WebSocket client
	client := &ws.Client{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		Conn:          conn,
		Hub:           h.wsHub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	// Register with WebSocket hub
	h.wsHub.Register <- client

	// Subscribe to collaboration room
	room := "collaboration:" + workflowID
	h.wsHub.SubscribeClient(client, room)

	// Register with collaboration hub
	h.collabHub.RegisterClient(workflowID, user.ID, client)

	// Start client pumps
	go h.writePump(client, workflowID, user.ID)
	go h.readPump(client, workflowID, user.ID, user.Email)

	h.logger.Info("collaboration connection established",
		"workflow_id", workflowID,
		"user_id", user.ID,
		"client_id", client.ID,
	)
}

// readPump handles incoming WebSocket messages
func (h *CollaborationHandler) readPump(client *ws.Client, workflowID, userID, userName string) {
	defer func() {
		h.handleClientDisconnect(client, workflowID, userID)
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Set read limit from configuration to prevent memory exhaustion
	client.Conn.SetReadLimit(h.wsConfig.MaxMessageSize)

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("websocket error", "error", err)
			}
			break
		}

		h.handleMessage(client, workflowID, userID, userName, message)
	}
}

// writePump handles outgoing WebSocket messages
func (h *CollaborationHandler) writePump(client *ws.Client, workflowID, userID string) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming collaboration messages
func (h *CollaborationHandler) handleMessage(client *ws.Client, workflowID, userID, userName string, data []byte) {
	var msg collaboration.WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		h.logger.Error("failed to parse message", "error", err)
		h.sendError(client, "invalid message format")
		return
	}

	ctx := context.Background()

	switch msg.Type {
	case collaboration.MessageTypeJoin:
		h.handleJoin(ctx, client, workflowID, userID, userName)

	case collaboration.MessageTypeLeave:
		h.handleLeave(ctx, workflowID, userID)

	case collaboration.MessageTypePresence:
		h.handlePresence(ctx, workflowID, userID, msg.Payload)

	case collaboration.MessageTypeLockAcquire:
		h.handleLockAcquire(ctx, workflowID, userID, msg.Payload)

	case collaboration.MessageTypeLockRelease:
		h.handleLockRelease(ctx, workflowID, userID, msg.Payload)

	case collaboration.MessageTypeChange:
		h.handleChange(ctx, workflowID, userID, msg.Payload)

	default:
		h.logger.Warn("unknown message type", "type", msg.Type)
		h.sendError(client, "unknown message type")
	}
}

// handleJoin processes join messages
func (h *CollaborationHandler) handleJoin(ctx context.Context, client *ws.Client, workflowID, userID, userName string) {
	presence, err := h.collabHub.GetService().JoinSession(ctx, workflowID, userID, userName)
	if err != nil {
		h.logger.Error("failed to join session", "error", err)
		h.sendError(client, "failed to join session")
		return
	}

	// Send session state to joining user
	session := h.collabHub.GetService().GetSession(workflowID)
	statePayload, _ := json.Marshal(collaboration.SessionState{Session: session})
	h.collabHub.BroadcastToUser(workflowID, userID, collaboration.WebSocketMessage{
		Type:      collaboration.MessageTypeUserJoined,
		Payload:   statePayload,
		Timestamp: time.Now(),
	})

	// Broadcast user joined to others
	joinedPayload, _ := json.Marshal(collaboration.UserJoinedPayload{User: presence})
	h.collabHub.BroadcastToOthers(workflowID, userID, collaboration.WebSocketMessage{
		Type:      collaboration.MessageTypeUserJoined,
		Payload:   joinedPayload,
		Timestamp: time.Now(),
	})

	h.logger.Info("user joined collaboration",
		"workflow_id", workflowID,
		"user_id", userID,
	)
}

// handleLeave processes leave messages
func (h *CollaborationHandler) handleLeave(ctx context.Context, workflowID, userID string) {
	err := h.collabHub.GetService().LeaveSession(ctx, workflowID, userID)
	if err != nil {
		h.logger.Error("failed to leave session", "error", err)
		return
	}

	// Broadcast user left
	leftPayload, _ := json.Marshal(collaboration.UserLeftPayload{UserID: userID})
	h.collabHub.BroadcastToOthers(workflowID, userID, collaboration.WebSocketMessage{
		Type:      collaboration.MessageTypeUserLeft,
		Payload:   leftPayload,
		Timestamp: time.Now(),
	})

	h.logger.Info("user left collaboration",
		"workflow_id", workflowID,
		"user_id", userID,
	)
}

// handlePresence processes presence update messages
func (h *CollaborationHandler) handlePresence(ctx context.Context, workflowID, userID string, payload json.RawMessage) {
	var presence collaboration.PresencePayload
	if err := json.Unmarshal(payload, &presence); err != nil {
		h.logger.Error("failed to parse presence payload", "error", err)
		return
	}

	err := h.collabHub.GetService().UpdatePresence(ctx, workflowID, userID, presence.Cursor, presence.Selection)
	if err != nil {
		h.logger.Error("failed to update presence", "error", err)
		return
	}

	// Broadcast presence update to others
	h.collabHub.BroadcastToOthers(workflowID, userID, collaboration.WebSocketMessage{
		Type:      collaboration.MessageTypePresenceUpdate,
		Payload:   payload,
		Timestamp: time.Now(),
	})
}

// handleLockAcquire processes lock acquisition messages
func (h *CollaborationHandler) handleLockAcquire(ctx context.Context, workflowID, userID string, payload json.RawMessage) {
	var lockReq collaboration.LockAcquirePayload
	if err := json.Unmarshal(payload, &lockReq); err != nil {
		h.logger.Error("failed to parse lock payload", "error", err)
		return
	}

	// Validate element ID and type
	if !isValidID(lockReq.ElementID) {
		h.logger.Warn("invalid element ID in lock request",
			"element_id", lockReq.ElementID,
			"user_id", userID,
			"workflow_id", workflowID,
		)
		return
	}

	if !isValidElementType(lockReq.ElementType) {
		h.logger.Warn("invalid element type in lock request",
			"element_type", lockReq.ElementType,
			"user_id", userID,
			"workflow_id", workflowID,
		)
		return
	}

	lock, err := h.collabHub.GetService().AcquireLock(ctx, workflowID, userID, lockReq.ElementID, lockReq.ElementType)
	if err != nil {
		// Lock acquisition failed, send failure to requester
		session := h.collabHub.GetService().GetSession(workflowID)
		var currentLock *collaboration.EditLock
		if session != nil {
			currentLock = session.Locks[lockReq.ElementID]
		}

		failedPayload, _ := json.Marshal(collaboration.LockFailedPayload{
			ElementID:   lockReq.ElementID,
			Reason:      err.Error(),
			CurrentLock: currentLock,
		})
		h.collabHub.BroadcastToUser(workflowID, userID, collaboration.WebSocketMessage{
			Type:      collaboration.MessageTypeLockFailed,
			Payload:   failedPayload,
			Timestamp: time.Now(),
		})
		return
	}

	// Broadcast lock acquired to all users
	acquiredPayload, _ := json.Marshal(collaboration.LockAcquiredPayload{Lock: lock})
	h.collabHub.BroadcastToWorkflow(workflowID, collaboration.WebSocketMessage{
		Type:      collaboration.MessageTypeLockAcquired,
		Payload:   acquiredPayload,
		Timestamp: time.Now(),
	})

	h.logger.Info("lock acquired",
		"workflow_id", workflowID,
		"user_id", userID,
		"element_id", lockReq.ElementID,
	)
}

// handleLockRelease processes lock release messages
func (h *CollaborationHandler) handleLockRelease(ctx context.Context, workflowID, userID string, payload json.RawMessage) {
	var lockRel collaboration.LockReleasePayload
	if err := json.Unmarshal(payload, &lockRel); err != nil {
		h.logger.Error("failed to parse lock release payload", "error", err)
		return
	}

	// Validate element ID
	if !isValidID(lockRel.ElementID) {
		h.logger.Warn("invalid element ID in lock release",
			"element_id", lockRel.ElementID,
			"user_id", userID,
			"workflow_id", workflowID,
		)
		return
	}

	err := h.collabHub.GetService().ReleaseLock(ctx, workflowID, userID, lockRel.ElementID)
	if err != nil {
		h.logger.Error("failed to release lock", "error", err)
		return
	}

	// Broadcast lock released to all users
	releasedPayload, _ := json.Marshal(collaboration.LockReleasedPayload{ElementID: lockRel.ElementID})
	h.collabHub.BroadcastToWorkflow(workflowID, collaboration.WebSocketMessage{
		Type:      collaboration.MessageTypeLockReleased,
		Payload:   releasedPayload,
		Timestamp: time.Now(),
	})

	h.logger.Info("lock released",
		"workflow_id", workflowID,
		"user_id", userID,
		"element_id", lockRel.ElementID,
	)
}

// handleChange processes change messages
func (h *CollaborationHandler) handleChange(ctx context.Context, workflowID, userID string, payload json.RawMessage) {
	// Broadcast change to other users
	h.collabHub.BroadcastToOthers(workflowID, userID, collaboration.WebSocketMessage{
		Type:      collaboration.MessageTypeChangeApplied,
		Payload:   payload,
		Timestamp: time.Now(),
	})

	h.logger.Debug("change broadcast",
		"workflow_id", workflowID,
		"user_id", userID,
	)
}

// handleClientDisconnect handles client disconnection
func (h *CollaborationHandler) handleClientDisconnect(client *ws.Client, workflowID, userID string) {
	h.collabHub.UnregisterClient(workflowID, userID)
	h.wsHub.Unregister <- client
	client.Conn.Close()

	ctx := context.Background()
	h.handleLeave(ctx, workflowID, userID)

	h.logger.Info("collaboration client disconnected",
		"workflow_id", workflowID,
		"user_id", userID,
	)
}

// sendError sends an error message to a client
func (h *CollaborationHandler) sendError(client *ws.Client, message string) {
	errorPayload, _ := json.Marshal(collaboration.ErrorPayload{Message: message})
	msg := collaboration.WebSocketMessage{
		Type:      collaboration.MessageTypeError,
		Payload:   errorPayload,
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(msg)
	select {
	case client.Send <- data:
	default:
		h.logger.Warn("failed to send error message, channel full")
	}
}

// isValidID validates that an ID contains only allowed characters
// Prevents injection attacks by allowing only alphanumeric, hyphens, and underscores
func isValidID(id string) bool {
	if id == "" || len(id) > 256 {
		return false
	}

	// Allow alphanumeric, hyphens, and underscores only
	for _, ch := range id {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_') {
			return false
		}
	}

	return true
}

// isValidElementType validates that an element type is one of the allowed types
func isValidElementType(elementType string) bool {
	allowedTypes := map[string]bool{
		"node": true,
		"edge": true,
	}
	return allowedTypes[elementType]
}
