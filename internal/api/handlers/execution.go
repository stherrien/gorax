package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/workflow"
)

// ExecutionService defines the methods needed from workflow service
type ExecutionService interface {
	ListExecutionsAdvanced(ctx context.Context, tenantID string, filter workflow.ExecutionFilter, cursor string, limit int) (*workflow.ExecutionListResult, error)
	GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*workflow.ExecutionWithSteps, error)
	GetExecutionStats(ctx context.Context, tenantID string, filter workflow.ExecutionFilter) (*workflow.ExecutionStats, error)
}

// ExecutionHandler handles execution-related HTTP requests
type ExecutionHandler struct {
	service ExecutionService
	logger  *slog.Logger
}

// NewExecutionHandler creates a new execution handler
func NewExecutionHandler(service ExecutionService, logger *slog.Logger) *ExecutionHandler {
	return &ExecutionHandler{
		service: service,
		logger:  logger,
	}
}

// ListExecutionsAdvanced returns executions with advanced filtering and cursor pagination
// GET /api/v1/executions
func (h *ExecutionHandler) ListExecutionsAdvanced(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant ID not found")
		return
	}

	filter, err := h.parseExecutionFilter(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid filter parameters: "+err.Error())
		return
	}

	cursor := r.URL.Query().Get("cursor")
	limit := h.parseLimit(r)

	result, err := h.service.ListExecutionsAdvanced(r.Context(), tenantID, filter, cursor, limit)
	if err != nil {
		if _, ok := err.(*workflow.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("failed to list executions",
			"error", err,
			"tenant_id", tenantID,
		)
		h.respondError(w, http.StatusInternalServerError, "failed to list executions")
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// GetExecutionWithSteps retrieves an execution with all its step executions
// GET /api/v1/executions/:executionID/steps
func (h *ExecutionHandler) GetExecutionWithSteps(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant ID not found")
		return
	}

	executionID := chi.URLParam(r, "executionID")
	if executionID == "" {
		h.respondError(w, http.StatusBadRequest, "execution ID is required")
		return
	}

	result, err := h.service.GetExecutionWithSteps(r.Context(), tenantID, executionID)
	if err != nil {
		if err == workflow.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "execution not found")
			return
		}
		h.logger.Error("failed to get execution with steps",
			"error", err,
			"tenant_id", tenantID,
			"execution_id", executionID,
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get execution")
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// GetExecutionStats returns execution statistics grouped by status
// GET /api/v1/executions/stats
func (h *ExecutionHandler) GetExecutionStats(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant ID not found")
		return
	}

	filter, err := h.parseExecutionFilter(r)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid filter parameters: "+err.Error())
		return
	}

	stats, err := h.service.GetExecutionStats(r.Context(), tenantID, filter)
	if err != nil {
		if _, ok := err.(*workflow.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("failed to get execution stats",
			"error", err,
			"tenant_id", tenantID,
		)
		h.respondError(w, http.StatusInternalServerError, "failed to get execution stats")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// parseExecutionFilter parses execution filter from query parameters
func (h *ExecutionHandler) parseExecutionFilter(r *http.Request) (workflow.ExecutionFilter, error) {
	filter := workflow.ExecutionFilter{
		WorkflowID:  r.URL.Query().Get("workflow_id"),
		Status:      r.URL.Query().Get("status"),
		TriggerType: r.URL.Query().Get("trigger_type"),
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return filter, err
		}
		filter.StartDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return filter, err
		}
		filter.EndDate = &endDate
	}

	return filter, nil
}

// parseLimit parses and validates the limit query parameter
func (h *ExecutionHandler) parseLimit(r *http.Request) int {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return 0 // Will use service default
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		return 0 // Will use service default
	}

	return limit
}

// Helper methods

func (h *ExecutionHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *ExecutionHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
