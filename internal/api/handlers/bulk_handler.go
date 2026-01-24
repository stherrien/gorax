package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/api/response"
	"github.com/gorax/gorax/internal/schedule"
)

// BulkOperationError represents an error for a single item in a bulk operation
type BulkOperationError struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	Success []string             `json:"success"`
	Failed  []BulkOperationError `json:"failed"`
}

// BulkScheduleService defines bulk operations for schedules
type BulkScheduleService interface {
	BulkUpdate(ctx context.Context, tenantID string, ids []string, enabled bool) ([]string, []BulkOperationError)
	BulkDelete(ctx context.Context, tenantID string, ids []string) ([]string, []BulkOperationError)
	ExportSchedules(ctx context.Context, tenantID string, ids []string) ([]*schedule.Schedule, error)
}

// BulkExecutionService defines bulk operations for executions
type BulkExecutionService interface {
	BulkDelete(ctx context.Context, tenantID string, ids []string) ([]string, []BulkOperationError)
	BulkRetry(ctx context.Context, tenantID string, ids []string) ([]string, []BulkOperationError)
}

// BulkHandler handles bulk operation HTTP requests
type BulkHandler struct {
	scheduleService  BulkScheduleService
	executionService BulkExecutionService
	logger           *slog.Logger
}

// NewBulkHandler creates a new bulk operations handler
func NewBulkHandler(
	scheduleService BulkScheduleService,
	executionService BulkExecutionService,
	logger *slog.Logger,
) *BulkHandler {
	return &BulkHandler{
		scheduleService:  scheduleService,
		executionService: executionService,
		logger:           logger,
	}
}

// BulkUpdateSchedulesInput represents input for bulk schedule updates
type BulkUpdateSchedulesInput struct {
	IDs    []string `json:"ids"`
	Action string   `json:"action"` // "enable", "disable", "delete"
}

// BulkUpdateSchedules handles bulk schedule updates
// PATCH /api/v1/schedules/bulk
func (h *BulkHandler) BulkUpdateSchedules(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		_ = response.InternalError(w, "tenant ID not found")
		return
	}

	var input BulkUpdateSchedulesInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validateBulkScheduleInput(input); err != nil {
		_ = response.BadRequest(w, err.Error())
		return
	}

	var success []string
	var failed []BulkOperationError

	switch input.Action {
	case "enable":
		success, failed = h.scheduleService.BulkUpdate(r.Context(), tenantID, input.IDs, true)
	case "disable":
		success, failed = h.scheduleService.BulkUpdate(r.Context(), tenantID, input.IDs, false)
	case "delete":
		success, failed = h.scheduleService.BulkDelete(r.Context(), tenantID, input.IDs)
	default:
		_ = response.BadRequest(w, "invalid action: must be enable, disable, or delete")
		return
	}

	result := BulkOperationResult{
		Success: success,
		Failed:  failed,
	}

	h.logger.Info("bulk schedule operation completed",
		"action", input.Action,
		"tenant_id", tenantID,
		"success_count", len(success),
		"failed_count", len(failed),
	)

	_ = response.OK(w, result)
}

// BulkDeleteExecutionsInput represents input for bulk execution deletion
type BulkDeleteExecutionsInput struct {
	IDs []string `json:"ids"`
}

// BulkDeleteExecutions handles bulk execution deletion
// DELETE /api/v1/executions/bulk
func (h *BulkHandler) BulkDeleteExecutions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		_ = response.InternalError(w, "tenant ID not found")
		return
	}

	var input BulkDeleteExecutionsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	if len(input.IDs) == 0 {
		_ = response.BadRequest(w, "at least one execution ID is required")
		return
	}

	success, failed := h.executionService.BulkDelete(r.Context(), tenantID, input.IDs)

	result := BulkOperationResult{
		Success: success,
		Failed:  failed,
	}

	h.logger.Info("bulk execution deletion completed",
		"tenant_id", tenantID,
		"success_count", len(success),
		"failed_count", len(failed),
	)

	_ = response.OK(w, result)
}

// BulkRetryExecutionsInput represents input for bulk execution retry
type BulkRetryExecutionsInput struct {
	IDs []string `json:"ids"`
}

// BulkRetryExecutions handles bulk execution retry
// POST /api/v1/executions/bulk/retry
func (h *BulkHandler) BulkRetryExecutions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		_ = response.InternalError(w, "tenant ID not found")
		return
	}

	var input BulkRetryExecutionsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	if len(input.IDs) == 0 {
		_ = response.BadRequest(w, "at least one execution ID is required")
		return
	}

	success, failed := h.executionService.BulkRetry(r.Context(), tenantID, input.IDs)

	result := BulkOperationResult{
		Success: success,
		Failed:  failed,
	}

	h.logger.Info("bulk execution retry completed",
		"tenant_id", tenantID,
		"success_count", len(success),
		"failed_count", len(failed),
	)

	_ = response.OK(w, result)
}

// ExportSchedules exports schedules as JSON
// GET /api/v1/schedules/export?ids=id1,id2
func (h *BulkHandler) ExportSchedules(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	if tenantID == "" {
		_ = response.InternalError(w, "tenant ID not found")
		return
	}

	idsParam := r.URL.Query().Get("ids")
	var ids []string
	if idsParam != "" {
		ids = strings.Split(idsParam, ",")
	}

	schedules, err := h.scheduleService.ExportSchedules(r.Context(), tenantID, ids)
	if err != nil {
		h.logger.Error("failed to export schedules",
			"error", err,
			"tenant_id", tenantID,
		)
		_ = response.InternalError(w, "failed to export schedules")
		return
	}

	h.logger.Info("schedules exported",
		"tenant_id", tenantID,
		"count", len(schedules),
	)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=schedules-export.json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(schedules)
}

// validateBulkScheduleInput validates bulk schedule input
func (h *BulkHandler) validateBulkScheduleInput(input BulkUpdateSchedulesInput) error {
	if len(input.IDs) == 0 {
		return &ValidationError{Message: "at least one schedule ID is required"}
	}

	if input.Action == "" {
		return &ValidationError{Message: "action is required"}
	}

	validActions := map[string]bool{"enable": true, "disable": true, "delete": true}
	if !validActions[input.Action] {
		return &ValidationError{Message: "invalid action: must be enable, disable, or delete"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
