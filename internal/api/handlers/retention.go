package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	apiMiddleware "github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/retention"
)

// RetentionService defines the interface for retention service operations
type RetentionService interface {
	GetRetentionPolicy(ctx context.Context, tenantID string) (*retention.RetentionPolicy, error)
	CleanupOldExecutions(ctx context.Context, tenantID string) (*retention.CleanupResult, error)
	CleanupAllTenants(ctx context.Context) (*retention.CleanupResult, error)
}

// RetentionRepository defines the interface for retention repository operations
type RetentionRepository interface {
	SetRetentionPolicy(ctx context.Context, tenantID string, retentionDays int, enabled bool) error
}

// RetentionHandler handles retention policy endpoints
type RetentionHandler struct {
	service RetentionService
	repo    RetentionRepository
	logger  *slog.Logger
}

// NewRetentionHandler creates a new retention handler
func NewRetentionHandler(service RetentionService, repo RetentionRepository, logger *slog.Logger) *RetentionHandler {
	return &RetentionHandler{
		service: service,
		repo:    repo,
		logger:  logger,
	}
}

// GetRetentionPolicyInput represents the request body for updating retention policy
type UpdateRetentionPolicyInput struct {
	RetentionDays int  `json:"retention_days"`
	Enabled       bool `json:"enabled"`
}

// GetPolicy handles GET /api/v1/retention/policy
// Returns the retention policy for the current tenant
func (h *RetentionHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := apiMiddleware.GetTenantID(r)
	if tenantID == "" {
		h.logger.Error("tenant ID not found in context")
		http.Error(w, "tenant context required", http.StatusInternalServerError)
		return
	}

	policy, err := h.service.GetRetentionPolicy(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to get retention policy",
			"tenant_id", tenantID,
			"error", err,
		)
		http.Error(w, "failed to get retention policy", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": policy,
	}); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// UpdatePolicy handles PUT /api/v1/retention/policy
// Updates the retention policy for the current tenant
func (h *RetentionHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := apiMiddleware.GetTenantID(r)
	if tenantID == "" {
		h.logger.Error("tenant ID not found in context")
		http.Error(w, "tenant context required", http.StatusInternalServerError)
		return
	}

	var input UpdateRetentionPolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("failed to decode update retention policy request",
			"tenant_id", tenantID,
			"error", err,
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate retention days
	if input.RetentionDays < retention.MinRetentionDays {
		http.Error(w, "retention days must be at least 7", http.StatusBadRequest)
		return
	}
	if input.RetentionDays > retention.MaxRetentionDays {
		http.Error(w, "retention days must be at most 3650", http.StatusBadRequest)
		return
	}

	if err := h.repo.SetRetentionPolicy(r.Context(), tenantID, input.RetentionDays, input.Enabled); err != nil {
		if err == retention.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to update retention policy",
			"tenant_id", tenantID,
			"error", err,
		)
		http.Error(w, "failed to update retention policy", http.StatusInternalServerError)
		return
	}

	h.logger.Info("retention policy updated",
		"tenant_id", tenantID,
		"retention_days", input.RetentionDays,
		"enabled", input.Enabled,
	)

	// Return updated policy
	policy, err := h.service.GetRetentionPolicy(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to get updated retention policy",
			"tenant_id", tenantID,
			"error", err,
		)
		http.Error(w, "failed to get updated retention policy", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": policy,
	}); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// TriggerCleanup handles POST /api/v1/retention/cleanup
// Triggers an immediate cleanup for the current tenant (admin only)
func (h *RetentionHandler) TriggerCleanup(w http.ResponseWriter, r *http.Request) {
	tenantID := apiMiddleware.GetTenantID(r)
	if tenantID == "" {
		h.logger.Error("tenant ID not found in context")
		http.Error(w, "tenant context required", http.StatusInternalServerError)
		return
	}

	h.logger.Info("manual cleanup triggered",
		"tenant_id", tenantID,
	)

	result, err := h.service.CleanupOldExecutions(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("cleanup failed",
			"tenant_id", tenantID,
			"error", err,
		)
		http.Error(w, "cleanup failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("manual cleanup completed",
		"tenant_id", tenantID,
		"executions_deleted", result.ExecutionsDeleted,
		"executions_archived", result.ExecutionsArchived,
		"step_executions_deleted", result.StepExecutionsDeleted,
	)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": result,
	}); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// AdminGetPolicy handles GET /api/v1/admin/tenants/{tenantID}/retention
// Returns the retention policy for a specific tenant (admin only)
func (h *RetentionHandler) AdminGetPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	policy, err := h.service.GetRetentionPolicy(r.Context(), tenantID)
	if err != nil {
		if err == retention.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get retention policy",
			"tenant_id", tenantID,
			"error", err,
		)
		http.Error(w, "failed to get retention policy", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": policy,
	}); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// AdminUpdatePolicy handles PUT /api/v1/admin/tenants/{tenantID}/retention
// Updates the retention policy for a specific tenant (admin only)
func (h *RetentionHandler) AdminUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	var input UpdateRetentionPolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("failed to decode update retention policy request",
			"tenant_id", tenantID,
			"error", err,
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate retention days
	if input.RetentionDays < retention.MinRetentionDays {
		http.Error(w, "retention days must be at least 7", http.StatusBadRequest)
		return
	}
	if input.RetentionDays > retention.MaxRetentionDays {
		http.Error(w, "retention days must be at most 3650", http.StatusBadRequest)
		return
	}

	if err := h.repo.SetRetentionPolicy(r.Context(), tenantID, input.RetentionDays, input.Enabled); err != nil {
		if err == retention.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to update retention policy",
			"tenant_id", tenantID,
			"error", err,
		)
		http.Error(w, "failed to update retention policy", http.StatusInternalServerError)
		return
	}

	h.logger.Info("retention policy updated by admin",
		"tenant_id", tenantID,
		"retention_days", input.RetentionDays,
		"enabled", input.Enabled,
	)

	// Return updated policy
	policy, err := h.service.GetRetentionPolicy(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to get updated retention policy",
			"tenant_id", tenantID,
			"error", err,
		)
		http.Error(w, "failed to get updated retention policy", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": policy,
	}); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// AdminTriggerCleanup handles POST /api/v1/admin/tenants/{tenantID}/retention/cleanup
// Triggers an immediate cleanup for a specific tenant (admin only)
func (h *RetentionHandler) AdminTriggerCleanup(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("admin triggered cleanup",
		"tenant_id", tenantID,
	)

	result, err := h.service.CleanupOldExecutions(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("cleanup failed",
			"tenant_id", tenantID,
			"error", err,
		)
		http.Error(w, "cleanup failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("admin cleanup completed",
		"tenant_id", tenantID,
		"executions_deleted", result.ExecutionsDeleted,
		"executions_archived", result.ExecutionsArchived,
		"step_executions_deleted", result.StepExecutionsDeleted,
	)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": result,
	}); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// AdminTriggerAllTenantsCleanup handles POST /api/v1/admin/retention/cleanup-all
// Triggers cleanup for all tenants (admin only)
func (h *RetentionHandler) AdminTriggerAllTenantsCleanup(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("admin triggered cleanup for all tenants")

	result, err := h.service.CleanupAllTenants(r.Context())
	if err != nil {
		h.logger.Error("cleanup for all tenants failed",
			"error", err,
		)
		http.Error(w, "cleanup failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("admin cleanup for all tenants completed",
		"executions_deleted", result.ExecutionsDeleted,
		"executions_archived", result.ExecutionsArchived,
		"step_executions_deleted", result.StepExecutionsDeleted,
	)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": result,
	}); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
