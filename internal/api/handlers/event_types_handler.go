package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gorax/gorax/internal/api/response"
	"github.com/gorax/gorax/internal/eventtypes"
)

// EventTypeService defines the interface for event type business logic
type EventTypeService interface {
	ListEventTypes(ctx context.Context) ([]eventtypes.EventType, error)
}

// EventTypesHandler handles event type HTTP requests
type EventTypesHandler struct {
	service EventTypeService
	logger  *slog.Logger
}

// NewEventTypesHandler creates a new event types handler
func NewEventTypesHandler(service EventTypeService, logger *slog.Logger) *EventTypesHandler {
	return &EventTypesHandler{
		service: service,
		logger:  logger,
	}
}

// List returns all event types from the registry
func (h *EventTypesHandler) List(w http.ResponseWriter, r *http.Request) {
	eventTypes, err := h.service.ListEventTypes(r.Context())
	if err != nil {
		h.logger.Error("failed to list event types", "error", err)
		_ = response.InternalError(w, "failed to list event types")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": eventTypes,
	})
}
