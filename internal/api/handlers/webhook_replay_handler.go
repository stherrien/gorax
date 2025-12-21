package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/tracing"
	"github.com/gorax/gorax/internal/webhook"
)

// WebhookReplayHandler handles webhook event replay HTTP requests
type WebhookReplayHandler struct {
	replayService *webhook.ReplayService
	logger        *slog.Logger
}

// NewWebhookReplayHandler creates a new webhook replay handler
func NewWebhookReplayHandler(replayService *webhook.ReplayService, logger *slog.Logger) *WebhookReplayHandler {
	return &WebhookReplayHandler{
		replayService: replayService,
		logger:        logger,
	}
}

// ReplayEvent replays a single webhook event
// POST /api/v1/events/{id}/replay
func (h *WebhookReplayHandler) ReplayEvent(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	eventID := chi.URLParam(r, "eventID")

	var input webhook.ReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// Body is optional, so it's OK if it's empty
		input.EventID = eventID
	}

	// Use modified payload if provided
	var modifiedPayload json.RawMessage
	if len(input.ModifiedPayload) > 0 {
		modifiedPayload = input.ModifiedPayload
	}

	var result *webhook.ReplayResult

	// Wrap replay operation with tracing
	_, _ = tracing.TraceWebhookReplay(r.Context(), tenantID, "", eventID, func(ctx context.Context) (string, error) {
		result = h.replayService.ReplayEvent(ctx, tenantID, eventID, modifiedPayload)
		if !result.Success {
			return "", fmt.Errorf("replay failed: %s", result.Error)
		}
		return result.ExecutionID, nil
	})

	if !result.Success {
		status := http.StatusInternalServerError
		if result.Error == "event not found: webhook not found" {
			status = http.StatusNotFound
		}
		h.respondJSON(w, status, result)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// BatchReplayEvents replays multiple webhook events
// POST /api/v1/webhooks/{id}/events/replay
func (h *WebhookReplayHandler) BatchReplayEvents(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "webhookID")

	var input webhook.BatchReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(input.EventIDs) == 0 {
		h.respondError(w, http.StatusBadRequest, "eventIds is required and cannot be empty")
		return
	}

	var results *webhook.BatchReplayResponse

	// Wrap batch replay operation with tracing
	_ = tracing.TraceWebhookBatchReplay(r.Context(), tenantID, webhookID, len(input.EventIDs), func(ctx context.Context) (int, int, error) {
		results = h.replayService.BatchReplayEvents(ctx, tenantID, webhookID, input.EventIDs)

		// Count successes and failures
		successCount, failureCount := 0, 0
		for _, result := range results.Results {
			if result.Success {
				successCount++
			} else {
				failureCount++
			}
		}
		return successCount, failureCount, nil
	})

	h.respondJSON(w, http.StatusOK, results)
}

// Helper methods

func (h *WebhookReplayHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *WebhookReplayHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
