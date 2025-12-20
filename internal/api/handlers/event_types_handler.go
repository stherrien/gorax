package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

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
		h.respondError(w, http.StatusInternalServerError, "failed to list event types")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": eventTypes,
	})
}

// Helper methods

func (h *EventTypesHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *EventTypesHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
