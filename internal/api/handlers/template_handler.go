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
	"github.com/gorax/gorax/internal/template"
)

// TemplateHandler handles template HTTP requests
type TemplateHandler struct {
	service  TemplateService
	logger   *slog.Logger
	validate *validator.Validate
}

// TemplateService defines the interface for template business logic
type TemplateService interface {
	CreateTemplate(ctx context.Context, tenantID, userID string, input template.CreateTemplateInput) (*template.Template, error)
	GetTemplate(ctx context.Context, tenantID, id string) (*template.Template, error)
	ListTemplates(ctx context.Context, tenantID string, filter template.TemplateFilter) ([]*template.Template, error)
	UpdateTemplate(ctx context.Context, tenantID, id string, input template.UpdateTemplateInput) error
	DeleteTemplate(ctx context.Context, tenantID, id string) error
	CreateFromWorkflow(ctx context.Context, tenantID, userID string, input template.CreateTemplateFromWorkflowInput) (*template.Template, error)
	InstantiateTemplate(ctx context.Context, tenantID, templateID string, input template.InstantiateTemplateInput) (*template.InstantiateTemplateResult, error)
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(service TemplateService, logger *slog.Logger) *TemplateHandler {
	return &TemplateHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// ListTemplates returns all templates for the tenant
func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	filter := template.TemplateFilter{
		Category:    r.URL.Query().Get("category"),
		SearchQuery: r.URL.Query().Get("search"),
	}

	if tags := r.URL.Query().Get("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
	}

	if isPublic := r.URL.Query().Get("is_public"); isPublic != "" {
		isPublicBool := isPublic == "true"
		filter.IsPublic = &isPublicBool
	}

	templates, err := h.service.ListTemplates(r.Context(), tenantID, filter)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to list templates")
		return
	}

	h.respondJSON(w, http.StatusOK, templates)
}

// GetTemplate retrieves a single template
func (h *TemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	templateID := chi.URLParam(r, "id")

	tmpl, err := h.service.GetTemplate(r.Context(), tenantID, templateID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "template not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get template")
		return
	}

	h.respondJSON(w, http.StatusOK, tmpl)
}

// CreateTemplate creates a new template
func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)

	var input template.CreateTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	tmpl, err := h.service.CreateTemplate(r.Context(), tenantID, userID, input)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.respondError(w, http.StatusConflict, "template with this name already exists")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to create template")
		return
	}

	h.respondJSON(w, http.StatusCreated, tmpl)
}

// UpdateTemplate updates an existing template
func (h *TemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	templateID := chi.URLParam(r, "id")

	var input template.UpdateTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.UpdateTemplate(r.Context(), tenantID, templateID, input); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "template not found")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to update template")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "template updated successfully"})
}

// DeleteTemplate deletes a template
func (h *TemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	templateID := chi.URLParam(r, "id")

	if err := h.service.DeleteTemplate(r.Context(), tenantID, templateID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "template not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to delete template")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateFromWorkflow creates a template from an existing workflow
func (h *TemplateHandler) CreateFromWorkflow(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)
	workflowID := chi.URLParam(r, "workflowId")

	var input template.CreateTemplateFromWorkflowInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.WorkflowID = workflowID

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	tmpl, err := h.service.CreateFromWorkflow(r.Context(), tenantID, userID, input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "workflow not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to create template from workflow")
		return
	}

	h.respondJSON(w, http.StatusCreated, tmpl)
}

// InstantiateTemplate creates a workflow definition from a template
func (h *TemplateHandler) InstantiateTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	templateID := chi.URLParam(r, "id")

	var input template.InstantiateTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.InstantiateTemplate(r.Context(), tenantID, templateID, input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "template not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to instantiate template")
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *TemplateHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *TemplateHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
