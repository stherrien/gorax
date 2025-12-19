package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/gorax/gorax/internal/api/middleware"
	ws "github.com/gorax/gorax/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate origin against allowed domains
		// For now, allow all origins (matches CORS config)
		return true
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub    *ws.Hub
	logger *slog.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *ws.Hub, logger *slog.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
	}
}

// HandleConnection handles WebSocket connection upgrades
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// User is already authenticated via middleware
	user := middleware.GetUser(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tenantID := middleware.GetTenantID(r)

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade connection", "error", err)
		return
	}

	// Create client
	client := &ws.Client{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		Conn:          conn,
		Hub:           h.hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	// Register client with hub
	h.hub.Register <- client

	// Parse subscription parameters
	h.handleSubscriptions(r, client)

	// Start client pumps
	go client.WritePump()
	go client.ReadPump()

	h.logger.Info("websocket connection established",
		"client_id", client.ID,
		"tenant_id", tenantID,
		"user_id", user.ID,
	)
}

// handleSubscriptions processes subscription query parameters
func (h *WebSocketHandler) handleSubscriptions(r *http.Request, client *ws.Client) {
	// Get subscription parameters from query string
	executionID := r.URL.Query().Get("execution_id")
	workflowID := r.URL.Query().Get("workflow_id")
	subscribeTenant := r.URL.Query().Get("subscribe_tenant")

	// Subscribe to specific execution
	if executionID != "" {
		room := "execution:" + executionID
		h.hub.SubscribeClient(client, room)
		h.logger.Info("client subscribed to execution",
			"client_id", client.ID,
			"execution_id", executionID,
		)
	}

	// Subscribe to workflow updates
	if workflowID != "" {
		room := "workflow:" + workflowID
		h.hub.SubscribeClient(client, room)
		h.logger.Info("client subscribed to workflow",
			"client_id", client.ID,
			"workflow_id", workflowID,
		)
	}

	// Subscribe to tenant-wide updates (for dashboard)
	if subscribeTenant == "true" {
		room := "tenant:" + client.TenantID
		h.hub.SubscribeClient(client, room)
		h.logger.Info("client subscribed to tenant",
			"client_id", client.ID,
			"tenant_id", client.TenantID,
		)
	}
}

// HandleExecutionConnection is a convenience endpoint for execution-specific subscriptions
func (h *WebSocketHandler) HandleExecutionConnection(w http.ResponseWriter, r *http.Request) {
	executionID := chi.URLParam(r, "executionID")
	if executionID == "" {
		http.Error(w, "execution_id required", http.StatusBadRequest)
		return
	}

	// Add execution_id to query params for the main handler
	q := r.URL.Query()
	q.Set("execution_id", executionID)
	r.URL.RawQuery = q.Encode()

	h.HandleConnection(w, r)
}

// HandleWorkflowConnection is a convenience endpoint for workflow-specific subscriptions
func (h *WebSocketHandler) HandleWorkflowConnection(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowID")
	if workflowID == "" {
		http.Error(w, "workflow_id required", http.StatusBadRequest)
		return
	}

	// Add workflow_id to query params for the main handler
	q := r.URL.Query()
	q.Set("workflow_id", workflowID)
	r.URL.RawQuery = q.Encode()

	h.HandleConnection(w, r)
}
