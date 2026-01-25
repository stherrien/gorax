package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/api/response"
	"github.com/gorax/gorax/internal/schedule"
	"github.com/gorax/gorax/internal/validation"
)

// ScheduleService defines the interface for schedule operations
type ScheduleService interface {
	Create(ctx context.Context, tenantID, workflowID, userID string, input schedule.CreateScheduleInput) (*schedule.Schedule, error)
	GetByID(ctx context.Context, tenantID, id string) (*schedule.Schedule, error)
	Update(ctx context.Context, tenantID, id string, input schedule.UpdateScheduleInput) (*schedule.Schedule, error)
	Delete(ctx context.Context, tenantID, id string) error
	List(ctx context.Context, tenantID, workflowID string, limit, offset int) ([]*schedule.Schedule, error)
	ListAll(ctx context.Context, tenantID string, limit, offset int) ([]*schedule.ScheduleWithWorkflow, error)
	ParseNextRunTime(expression, timezone string) (time.Time, error)
	GetNextRunTimes(expression, timezone string, count int) ([]time.Time, error)
	ListExecutionLogs(ctx context.Context, tenantID, scheduleID string, limit, offset int) ([]*schedule.ExecutionLog, error)
	GetExecutionLog(ctx context.Context, tenantID, logID string) (*schedule.ExecutionLog, error)
	CountExecutionLogs(ctx context.Context, tenantID, scheduleID string) (int, error)
}

// ScheduleHandler handles schedule-related HTTP requests
type ScheduleHandler struct {
	service ScheduleService
	logger  *slog.Logger
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(service ScheduleService, logger *slog.Logger) *ScheduleHandler {
	return &ScheduleHandler{
		service: service,
		logger:  logger,
	}
}

// Create creates a new schedule for a workflow
func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)
	workflowID := chi.URLParam(r, "workflowID")

	var input schedule.CreateScheduleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	sched, err := h.service.Create(r.Context(), tenantID, workflowID, user.ID, input)
	if err != nil {
		if _, ok := err.(*schedule.ValidationError); ok {
			_ = response.BadRequest(w, err.Error())
			return
		}
		h.logger.Error("failed to create schedule", "error", err)
		_ = response.InternalError(w, "failed to create schedule")
		return
	}

	_ = response.Created(w, sched)
}

// List returns all schedules for a workflow
func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")

	limit, _ := validation.ParsePaginationLimit(
		r.URL.Query().Get("limit"),
		validation.DefaultPaginationLimit,
		validation.MaxPaginationLimit,
	)
	offset, _ := validation.ParsePaginationOffset(r.URL.Query().Get("offset"))

	schedules, err := h.service.List(r.Context(), tenantID, workflowID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list schedules", "error", err)
		_ = response.InternalError(w, "failed to list schedules")
		return
	}

	_ = response.Paginated(w, schedules, limit, offset, 0)
}

// ListAll returns all schedules for the tenant
func (h *ScheduleHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	limit, _ := validation.ParsePaginationLimit(
		r.URL.Query().Get("limit"),
		validation.DefaultPaginationLimit,
		validation.MaxPaginationLimit,
	)
	offset, _ := validation.ParsePaginationOffset(r.URL.Query().Get("offset"))

	schedules, err := h.service.ListAll(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list all schedules", "error", err)
		_ = response.InternalError(w, "failed to list schedules")
		return
	}

	_ = response.Paginated(w, schedules, limit, offset, 0)
}

// Get retrieves a single schedule
func (h *ScheduleHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	scheduleID := chi.URLParam(r, "scheduleID")

	sched, err := h.service.GetByID(r.Context(), tenantID, scheduleID)
	if err != nil {
		if err == schedule.ErrNotFound {
			_ = response.NotFound(w, "schedule not found")
			return
		}
		h.logger.Error("failed to get schedule", "error", err)
		_ = response.InternalError(w, "failed to get schedule")
		return
	}

	_ = response.OK(w, sched)
}

// Update updates a schedule
func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	scheduleID := chi.URLParam(r, "scheduleID")

	var input schedule.UpdateScheduleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	sched, err := h.service.Update(r.Context(), tenantID, scheduleID, input)
	if err != nil {
		if err == schedule.ErrNotFound {
			_ = response.NotFound(w, "schedule not found")
			return
		}
		if _, ok := err.(*schedule.ValidationError); ok {
			_ = response.BadRequest(w, err.Error())
			return
		}
		h.logger.Error("failed to update schedule", "error", err)
		_ = response.InternalError(w, "failed to update schedule")
		return
	}

	_ = response.OK(w, sched)
}

// Delete deletes a schedule
func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	scheduleID := chi.URLParam(r, "scheduleID")

	err := h.service.Delete(r.Context(), tenantID, scheduleID)
	if err != nil {
		if err == schedule.ErrNotFound {
			_ = response.NotFound(w, "schedule not found")
			return
		}
		h.logger.Error("failed to delete schedule", "error", err)
		_ = response.InternalError(w, "failed to delete schedule")
		return
	}

	response.NoContent(w)
}

// ParseCron validates a cron expression and returns next run times
func (h *ScheduleHandler) ParseCron(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CronExpression string `json:"cron_expression"`
		Timezone       string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	// Default timezone to UTC
	if input.Timezone == "" {
		input.Timezone = "UTC"
	}

	// Calculate next 5 run times
	nextRun, err := h.service.ParseNextRunTime(input.CronExpression, input.Timezone)
	if err != nil {
		_ = response.BadRequest(w, "invalid cron expression: "+err.Error())
		return
	}

	_ = response.JSON(w, http.StatusOK, map[string]interface{}{
		"valid":    true,
		"next_run": nextRun,
	})
}

// PreviewSchedule returns next N execution times for a cron expression
func (h *ScheduleHandler) PreviewSchedule(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CronExpression string `json:"cron_expression"`
		Timezone       string `json:"timezone"`
		Count          int    `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	// Default timezone to UTC
	if input.Timezone == "" {
		input.Timezone = "UTC"
	}

	// Default count to 10, max 50
	if input.Count <= 0 {
		input.Count = 10
	}
	if input.Count > 50 {
		input.Count = 50
	}

	// Get next N run times
	nextRuns, err := h.service.GetNextRunTimes(input.CronExpression, input.Timezone, input.Count)
	if err != nil {
		_ = response.BadRequest(w, "invalid cron expression: "+err.Error())
		return
	}

	_ = response.JSON(w, http.StatusOK, map[string]interface{}{
		"valid":     true,
		"next_runs": nextRuns,
		"count":     len(nextRuns),
		"timezone":  input.Timezone,
	})
}

// ListExecutionHistory returns execution history for a schedule
func (h *ScheduleHandler) ListExecutionHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	scheduleID := chi.URLParam(r, "scheduleID")

	limit, _ := validation.ParsePaginationLimit(
		r.URL.Query().Get("limit"),
		validation.DefaultPaginationLimit,
		validation.MaxPaginationLimit,
	)
	offset, _ := validation.ParsePaginationOffset(r.URL.Query().Get("offset"))

	logs, err := h.service.ListExecutionLogs(r.Context(), tenantID, scheduleID, limit, offset)
	if err != nil {
		if err == schedule.ErrNotFound {
			_ = response.NotFound(w, "schedule not found")
			return
		}
		h.logger.Error("failed to list execution history", "error", err)
		_ = response.InternalError(w, "failed to list execution history")
		return
	}

	// Get total count
	total, err := h.service.CountExecutionLogs(r.Context(), tenantID, scheduleID)
	if err != nil {
		h.logger.Error("failed to count execution logs", "error", err)
		// Don't fail the request, just set total to 0
		total = 0
	}

	_ = response.Paginated(w, logs, limit, offset, total)
}

// GetExecutionLog retrieves a specific execution log
func (h *ScheduleHandler) GetExecutionLog(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	logID := chi.URLParam(r, "logID")

	log, err := h.service.GetExecutionLog(r.Context(), tenantID, logID)
	if err != nil {
		if err == schedule.ErrNotFound {
			_ = response.NotFound(w, "execution log not found")
			return
		}
		h.logger.Error("failed to get execution log", "error", err)
		_ = response.InternalError(w, "failed to get execution log")
		return
	}

	_ = response.OK(w, log)
}
