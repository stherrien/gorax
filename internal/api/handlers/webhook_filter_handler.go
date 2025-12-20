package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/webhook"
)

// WebhookFilterHandler handles webhook filter HTTP requests
type WebhookFilterHandler struct {
	service  WebhookFilterService
	logger   *slog.Logger
	validate *validator.Validate
}

// WebhookFilterService defines the interface for filter business logic
type WebhookFilterService interface {
	ListFilters(ctx context.Context, tenantID, webhookID string) ([]*webhook.WebhookFilter, error)
	GetFilter(ctx context.Context, tenantID, webhookID, filterID string) (*webhook.WebhookFilter, error)
	CreateFilter(ctx context.Context, tenantID, webhookID string, filter *webhook.WebhookFilter) (*webhook.WebhookFilter, error)
	UpdateFilter(ctx context.Context, tenantID, webhookID, filterID string, filter *webhook.WebhookFilter) (*webhook.WebhookFilter, error)
	DeleteFilter(ctx context.Context, tenantID, webhookID, filterID string) error
	TestFilters(ctx context.Context, tenantID, webhookID string, payload map[string]interface{}) (*webhook.FilterResult, error)
}

// NewWebhookFilterHandler creates a new webhook filter handler
func NewWebhookFilterHandler(service WebhookFilterService, logger *slog.Logger) *WebhookFilterHandler {
	return &WebhookFilterHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

// CreateFilterRequest represents the request to create a filter
type CreateFilterRequest struct {
	FieldPath  string      `json:"fieldPath" validate:"required"`
	Operator   string      `json:"operator" validate:"required,oneof=equals not_equals contains not_contains starts_with ends_with regex gt lt in not_in exists not_exists"`
	Value      interface{} `json:"value"`
	LogicGroup int         `json:"logicGroup" validate:"min=0"`
	Enabled    bool        `json:"enabled"`
}

// UpdateFilterRequest represents the request to update a filter
type UpdateFilterRequest struct {
	FieldPath  string      `json:"fieldPath" validate:"required"`
	Operator   string      `json:"operator" validate:"required,oneof=equals not_equals contains not_contains starts_with ends_with regex gt lt in not_in exists not_exists"`
	Value      interface{} `json:"value"`
	LogicGroup int         `json:"logicGroup" validate:"min=0"`
	Enabled    bool        `json:"enabled"`
}

// TestFiltersRequest represents the request to test filters
type TestFiltersRequest struct {
	Payload map[string]interface{} `json:"payload" validate:"required"`
}

// List returns all filters for a webhook
func (h *WebhookFilterHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")

	filters, err := h.service.ListFilters(r.Context(), tenantID, webhookID)
	if err != nil {
		if err == webhook.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "webhook not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to list filters")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": filters,
	})
}

// Get retrieves a single filter
func (h *WebhookFilterHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")
	filterID := chi.URLParam(r, "filterID")

	filter, err := h.service.GetFilter(r.Context(), tenantID, webhookID, filterID)
	if err != nil {
		if err == webhook.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "filter not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get filter")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": filter,
	})
}

// Create creates a new filter
func (h *WebhookFilterHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")

	var input CreateFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := &webhook.WebhookFilter{
		FieldPath:  input.FieldPath,
		Operator:   webhook.FilterOperator(input.Operator),
		Value:      input.Value,
		LogicGroup: input.LogicGroup,
		Enabled:    input.Enabled,
	}

	created, err := h.service.CreateFilter(r.Context(), tenantID, webhookID, filter)
	if err != nil {
		if err == webhook.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "webhook not found")
			return
		}
		h.logger.Error("failed to create filter", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to create filter")
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data": created,
	})
}

// Update updates a filter
func (h *WebhookFilterHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")
	filterID := chi.URLParam(r, "filterID")

	var input UpdateFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := &webhook.WebhookFilter{
		FieldPath:  input.FieldPath,
		Operator:   webhook.FilterOperator(input.Operator),
		Value:      input.Value,
		LogicGroup: input.LogicGroup,
		Enabled:    input.Enabled,
	}

	updated, err := h.service.UpdateFilter(r.Context(), tenantID, webhookID, filterID, filter)
	if err != nil {
		if err == webhook.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "filter not found")
			return
		}
		h.logger.Error("failed to update filter", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to update filter")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": updated,
	})
}

// Delete deletes a filter
func (h *WebhookFilterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")
	filterID := chi.URLParam(r, "filterID")

	err := h.service.DeleteFilter(r.Context(), tenantID, webhookID, filterID)
	if err != nil {
		if err == webhook.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "filter not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to delete filter")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Test tests filters against a sample payload
func (h *WebhookFilterHandler) Test(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	webhookID := chi.URLParam(r, "id")

	var input TestFiltersRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.TestFilters(r.Context(), tenantID, webhookID, input.Payload)
	if err != nil {
		if err == webhook.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "webhook not found")
			return
		}
		h.logger.Error("failed to test filters", "error", err)
		h.respondError(w, http.StatusInternalServerError, "failed to test filters")
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// Helper methods

func (h *WebhookFilterHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *WebhookFilterHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
