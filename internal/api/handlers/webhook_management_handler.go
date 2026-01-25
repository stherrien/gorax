package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/api/response"
	"github.com/gorax/gorax/internal/tracing"
	"github.com/gorax/gorax/internal/validation"
	"github.com/gorax/gorax/internal/webhook"
)

// WebhookManagementHandler handles webhook management HTTP requests
type WebhookManagementHandler struct {
	service  WebhookManagementService
	logger   *slog.Logger
	validate *validator.Validate
}

// WebhookManagementService defines the interface for webhook business logic
type WebhookManagementService interface {
	List(ctx context.Context, tenantID string, limit, offset int) ([]*webhook.Webhook, int, error)
	GetByID(ctx context.Context, tenantID, webhookID string) (*webhook.Webhook, error)
	CreateWithDetails(ctx context.Context, tenantID, workflowID, name, path, authType, description string, priority int) (*webhook.Webhook, error)
	Update(ctx context.Context, tenantID, webhookID, name, authType, description string, priority int, enabled bool) (*webhook.Webhook, error)
	DeleteByID(ctx context.Context, tenantID, webhookID string) error
	RegenerateSecret(ctx context.Context, tenantID, webhookID string) (*webhook.Webhook, error)
	TestWebhook(ctx context.Context, tenantID, webhookID, method string, headers map[string]string, body json.RawMessage) (*webhook.TestResult, error)
	GetEventHistory(ctx context.Context, tenantID, webhookID string, limit, offset int) ([]*webhook.Event, int, error)
}

// NewWebhookManagementHandler creates a new webhook management handler
func NewWebhookManagementHandler(service WebhookManagementService, logger *slog.Logger) *WebhookManagementHandler {
	return &WebhookManagementHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// CreateWebhookRequest represents the request to create a webhook
type CreateWebhookRequest struct {
	Name        string `json:"name" validate:"required"`
	WorkflowID  string `json:"workflowId" validate:"required"`
	Path        string `json:"path" validate:"required"`
	AuthType    string `json:"authType" validate:"oneof=none signature basic api_key"`
	Description string `json:"description"`
	Priority    int    `json:"priority" validate:"min=0,max=3"`
}

// UpdateWebhookRequest represents the request to update a webhook
type UpdateWebhookRequest struct {
	Name        string `json:"name"`
	AuthType    string `json:"authType" validate:"omitempty,oneof=none signature basic api_key"`
	Description string `json:"description"`
	Priority    int    `json:"priority" validate:"min=0,max=3"`
	Enabled     bool   `json:"enabled"`
}

// TestWebhookRequest represents the request to test a webhook
type TestWebhookRequest struct {
	Method  string            `json:"method" validate:"required"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
}

// List returns all webhooks for the tenant
// @Summary List webhooks
// @Description Returns a paginated list of webhooks for the authenticated tenant
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Security TenantID
// @Security UserID
// @Success 200 {object} map[string]any "List of webhooks"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /webhooks [get]
func (h *WebhookManagementHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	limit, _ := validation.ParsePaginationLimit(
		r.URL.Query().Get("limit"),
		validation.DefaultPaginationLimit,
		validation.MaxPaginationLimit,
	)
	offset, _ := validation.ParsePaginationOffset(r.URL.Query().Get("offset"))

	webhooks, total, err := h.service.List(r.Context(), tenantID, limit, offset)
	if err != nil {
		_ = response.InternalError(w, "failed to list webhooks")
		return
	}

	_ = response.OK(w, map[string]any{
		"data":   webhooks,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Get retrieves a single webhook
// @Summary Get webhook
// @Description Retrieves a webhook by ID
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID"
// @Security TenantID
// @Security UserID
// @Success 200 {object} map[string]any "Webhook details"
// @Failure 404 {object} map[string]string "Webhook not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /webhooks/{id} [get]
func (h *WebhookManagementHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")

	wh, err := h.service.GetByID(r.Context(), tenantID, webhookID)
	if err != nil {
		if err == webhook.ErrNotFound {
			_ = response.NotFound(w, "webhook not found")
			return
		}
		_ = response.InternalError(w, "failed to get webhook")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": wh,
	})
}

// Create creates a new webhook
// @Summary Create webhook
// @Description Creates a new webhook endpoint for a workflow
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param webhook body CreateWebhookRequest true "Webhook creation request"
// @Security TenantID
// @Security UserID
// @Success 201 {object} map[string]any "Created webhook"
// @Failure 400 {object} map[string]string "Invalid request or validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /webhooks [post]
func (h *WebhookManagementHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	var input CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = response.BadRequest(w, err.Error())
		return
	}

	wh, err := h.service.CreateWithDetails(
		r.Context(),
		tenantID,
		input.WorkflowID,
		input.Name,
		input.Path,
		input.AuthType,
		input.Description,
		input.Priority,
	)
	if err != nil {
		h.logger.Error("failed to create webhook", "error", err)
		_ = response.InternalError(w, "failed to create webhook")
		return
	}

	_ = response.Created(w, map[string]any{
		"data": wh,
	})
}

// Update updates a webhook
func (h *WebhookManagementHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")

	var input UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = response.BadRequest(w, err.Error())
		return
	}

	wh, err := h.service.Update(
		r.Context(),
		tenantID,
		webhookID,
		input.Name,
		input.AuthType,
		input.Description,
		input.Priority,
		input.Enabled,
	)
	if err != nil {
		if err == webhook.ErrNotFound {
			_ = response.NotFound(w, "webhook not found")
			return
		}
		_ = response.InternalError(w, "failed to update webhook")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": wh,
	})
}

// Delete deletes a webhook
func (h *WebhookManagementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")

	err := h.service.DeleteByID(r.Context(), tenantID, webhookID)
	if err != nil {
		if err == webhook.ErrNotFound {
			_ = response.NotFound(w, "webhook not found")
			return
		}
		_ = response.InternalError(w, "failed to delete webhook")
		return
	}

	response.NoContent(w)
}

// RegenerateSecret regenerates the webhook secret
// @Summary Regenerate webhook secret
// @Description Regenerates the secret key for webhook signature verification
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID"
// @Security TenantID
// @Security UserID
// @Success 200 {object} map[string]any "Webhook with new secret"
// @Failure 404 {object} map[string]string "Webhook not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /webhooks/{id}/regenerate-secret [post]
func (h *WebhookManagementHandler) RegenerateSecret(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")

	wh, err := h.service.RegenerateSecret(r.Context(), tenantID, webhookID)
	if err != nil {
		if err == webhook.ErrNotFound {
			_ = response.NotFound(w, "webhook not found")
			return
		}
		_ = response.InternalError(w, "failed to regenerate secret")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": wh,
	})
}

// TestWebhook tests a webhook with sample payload
func (h *WebhookManagementHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")

	var input TestWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		_ = response.BadRequest(w, err.Error())
		return
	}

	var result *webhook.TestResult
	var testErr error

	// Wrap webhook test operation with tracing
	_ = tracing.TraceWebhookReceive(r.Context(), tenantID, webhookID, input.Method, "/webhooks/test", func(ctx context.Context) error {
		result, testErr = h.service.TestWebhook(
			ctx,
			tenantID,
			webhookID,
			input.Method,
			input.Headers,
			input.Body,
		)
		return testErr
	})

	if testErr != nil {
		if testErr == webhook.ErrNotFound {
			_ = response.NotFound(w, "webhook not found")
			return
		}
		_ = response.InternalError(w, "failed to test webhook")
		return
	}

	_ = response.OK(w, result)
}

// GetEventHistory retrieves webhook event history
func (h *WebhookManagementHandler) GetEventHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")

	limit, _ := validation.ParsePaginationLimit(
		r.URL.Query().Get("limit"),
		validation.DefaultPaginationLimit,
		validation.MaxPaginationLimit,
	)
	offset, _ := validation.ParsePaginationOffset(r.URL.Query().Get("offset"))

	events, total, err := h.service.GetEventHistory(r.Context(), tenantID, webhookID, limit, offset)
	if err != nil {
		if err == webhook.ErrNotFound {
			_ = response.NotFound(w, "webhook not found")
			return
		}
		_ = response.InternalError(w, "failed to get event history")
		return
	}

	_ = response.OK(w, map[string]any{
		"data":   events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
