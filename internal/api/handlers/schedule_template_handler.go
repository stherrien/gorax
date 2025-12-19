package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/schedule"
)

// ScheduleTemplateHandler handles schedule template HTTP requests
type ScheduleTemplateHandler struct {
	service  ScheduleTemplateService
	logger   *slog.Logger
	validate *validator.Validate
}

// ScheduleTemplateService defines the interface for schedule template business logic
type ScheduleTemplateService interface {
	ListTemplates(ctx context.Context, filter schedule.ScheduleTemplateFilter) ([]*schedule.ScheduleTemplate, error)
	GetTemplate(ctx context.Context, id string) (*schedule.ScheduleTemplate, error)
	ApplyTemplate(ctx context.Context, tenantID, userID, templateID string, input schedule.ApplyTemplateInput) (*schedule.Schedule, error)
}

// NewScheduleTemplateHandler creates a new schedule template handler
func NewScheduleTemplateHandler(service ScheduleTemplateService, logger *slog.Logger) *ScheduleTemplateHandler {
	return &ScheduleTemplateHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// ListTemplates returns all schedule templates
func (h *ScheduleTemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	filter := schedule.ScheduleTemplateFilter{
		Category:    r.URL.Query().Get("category"),
		SearchQuery: r.URL.Query().Get("search"),
	}

	if tags := r.URL.Query().Get("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
	}

	if isSystem := r.URL.Query().Get("is_system"); isSystem != "" {
		isSystemBool := isSystem == "true"
		filter.IsSystem = &isSystemBool
	}

	if err := filter.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	templates, err := h.service.ListTemplates(r.Context(), filter)
	if err != nil {
		h.logError("failed to list schedule templates", err)
		respondError(w, http.StatusInternalServerError, "failed to list schedule templates")
		return
	}

	respondJSON(w, http.StatusOK, templates)
}

// GetTemplate retrieves a single schedule template
func (h *ScheduleTemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "id")

	template, err := h.service.GetTemplate(r.Context(), templateID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "template not found")
			return
		}
		h.logError("failed to get schedule template", err)
		respondError(w, http.StatusInternalServerError, "failed to get schedule template")
		return
	}

	respondJSON(w, http.StatusOK, template)
}

// ApplyTemplate applies a template to create a schedule
func (h *ScheduleTemplateHandler) ApplyTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)
	templateID := chi.URLParam(r, "id")

	var input schedule.ApplyTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := input.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	createdSchedule, err := h.service.ApplyTemplate(r.Context(), tenantID, userID, templateID, input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		h.logError("failed to apply schedule template", err)
		respondError(w, http.StatusInternalServerError, "failed to apply schedule template")
		return
	}

	respondJSON(w, http.StatusCreated, createdSchedule)
}

func (h *ScheduleTemplateHandler) logError(message string, err error) {
	if h.logger != nil {
		h.logger.Error(message, "error", err)
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
