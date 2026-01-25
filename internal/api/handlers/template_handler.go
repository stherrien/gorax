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
	"github.com/gorax/gorax/internal/api/response"
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
		_ = response.InternalError(w, "failed to list templates")
		return
	}

	_ = response.OK(w, templates)
}

// GetTemplate retrieves a single template
func (h *TemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	templateID := chi.URLParam(r, "id")

	tmpl, err := h.service.GetTemplate(r.Context(), tenantID, templateID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			_ = response.NotFound(w, "template not found")
			return
		}
		_ = response.InternalError(w, "failed to get template")
		return
	}

	_ = response.OK(w, tmpl)
}

// CreateTemplate creates a new template
func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)

	var input template.CreateTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = response.BadRequest(w, err.Error())
		return
	}

	tmpl, err := h.service.CreateTemplate(r.Context(), tenantID, userID, input)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			_ = response.Conflict(w, "template with this name already exists")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			_ = response.BadRequest(w, err.Error())
			return
		}
		_ = response.InternalError(w, "failed to create template")
		return
	}

	_ = response.Created(w, tmpl)
}

// UpdateTemplate updates an existing template
func (h *TemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	templateID := chi.URLParam(r, "id")

	var input template.UpdateTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.service.UpdateTemplate(r.Context(), tenantID, templateID, input); err != nil {
		if strings.Contains(err.Error(), "not found") {
			_ = response.NotFound(w, "template not found")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			_ = response.BadRequest(w, err.Error())
			return
		}
		_ = response.InternalError(w, "failed to update template")
		return
	}

	_ = response.OK(w, map[string]string{"message": "template updated successfully"})
}

// DeleteTemplate deletes a template
func (h *TemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	templateID := chi.URLParam(r, "id")

	if err := h.service.DeleteTemplate(r.Context(), tenantID, templateID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			_ = response.NotFound(w, "template not found")
			return
		}
		_ = response.InternalError(w, "failed to delete template")
		return
	}

	response.NoContent(w)
}

// CreateFromWorkflow creates a template from an existing workflow
func (h *TemplateHandler) CreateFromWorkflow(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	userID := middleware.GetUserID(r)
	workflowID := chi.URLParam(r, "workflowId")

	var input template.CreateTemplateFromWorkflowInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	input.WorkflowID = workflowID

	if err := h.validate.Struct(input); err != nil {
		_ = response.BadRequest(w, err.Error())
		return
	}

	tmpl, err := h.service.CreateFromWorkflow(r.Context(), tenantID, userID, input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			_ = response.NotFound(w, "workflow not found")
			return
		}
		_ = response.InternalError(w, "failed to create template from workflow")
		return
	}

	_ = response.Created(w, tmpl)
}

// InstantiateTemplate creates a workflow definition from a template
func (h *TemplateHandler) InstantiateTemplate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	templateID := chi.URLParam(r, "id")

	var input template.InstantiateTemplateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = response.BadRequest(w, err.Error())
		return
	}

	result, err := h.service.InstantiateTemplate(r.Context(), tenantID, templateID, input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			_ = response.NotFound(w, "template not found")
			return
		}
		_ = response.InternalError(w, "failed to instantiate template")
		return
	}

	_ = response.OK(w, result)
}
