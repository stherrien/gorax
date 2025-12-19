package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/schedule"
)

// ScheduleHandler handles schedule-related HTTP requests
type ScheduleHandler struct {
	service *schedule.Service
	logger  *slog.Logger
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(service *schedule.Service, logger *slog.Logger) *ScheduleHandler {
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
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sched, err := h.service.Create(r.Context(), tenantID, workflowID, user.ID, input)
	if err != nil {
		if _, ok := err.(*schedule.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("failed to create schedule", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to create schedule")
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data": sched,
	})
}

// List returns all schedules for a workflow
func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	workflowID := chi.URLParam(r, "workflowID")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	schedules, err := h.service.List(r.Context(), tenantID, workflowID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list schedules", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to list schedules")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   schedules,
		"limit":  limit,
		"offset": offset,
	})
}

// ListAll returns all schedules for the tenant
func (h *ScheduleHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	schedules, err := h.service.ListAll(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list all schedules", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to list schedules")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   schedules,
		"limit":  limit,
		"offset": offset,
	})
}

// Get retrieves a single schedule
func (h *ScheduleHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	scheduleID := chi.URLParam(r, "scheduleID")

	sched, err := h.service.GetByID(r.Context(), tenantID, scheduleID)
	if err != nil {
		if err == schedule.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "schedule not found")
			return
		}
		h.logger.Error("failed to get schedule", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to get schedule")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": sched,
	})
}

// Update updates a schedule
func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	scheduleID := chi.URLParam(r, "scheduleID")

	var input schedule.UpdateScheduleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sched, err := h.service.Update(r.Context(), tenantID, scheduleID, input)
	if err != nil {
		if err == schedule.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "schedule not found")
			return
		}
		if _, ok := err.(*schedule.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("failed to update schedule", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to update schedule")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": sched,
	})
}

// Delete deletes a schedule
func (h *ScheduleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	scheduleID := chi.URLParam(r, "scheduleID")

	err := h.service.Delete(r.Context(), tenantID, scheduleID)
	if err != nil {
		if err == schedule.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "schedule not found")
			return
		}
		h.logger.Error("failed to delete schedule", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to delete schedule")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ParseCron validates a cron expression and returns next run times
func (h *ScheduleHandler) ParseCron(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CronExpression string `json:"cron_expression"`
		Timezone       string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Default timezone to UTC
	if input.Timezone == "" {
		input.Timezone = "UTC"
	}

	// Calculate next 5 run times
	nextRun, err := h.service.ParseNextRunTime(input.CronExpression, input.Timezone)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid cron expression: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid":    true,
		"next_run": nextRun,
	})
}

// Helper methods

func (h *ScheduleHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ScheduleHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
