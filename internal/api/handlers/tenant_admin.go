package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/tenant"
	"github.com/gorax/gorax/internal/validation"
)

// TenantAdminHandler handles tenant administration endpoints
type TenantAdminHandler struct {
	tenantService *tenant.Service
	logger        *slog.Logger
}

// NewTenantAdminHandler creates a new tenant admin handler
func NewTenantAdminHandler(tenantService *tenant.Service, logger *slog.Logger) *TenantAdminHandler {
	return &TenantAdminHandler{
		tenantService: tenantService,
		logger:        logger,
	}
}

// CreateTenant handles POST /api/v1/admin/tenants
func (h *TenantAdminHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var input tenant.CreateTenantInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode create tenant request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := ValidateCreateTenantInput(input); err != nil {
		h.logger.Warn("invalid create tenant input", "error", err)
		http.Error(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	t, err := h.tenantService.Create(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to create tenant", "error", err)
		http.Error(w, "failed to create tenant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// ListTenants handles GET /api/v1/admin/tenants
func (h *TenantAdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters with overflow protection
	limit, _ := validation.ParsePaginationLimit(
		r.URL.Query().Get("limit"),
		validation.DefaultPaginationLimit,
		100, // Admin endpoint has a lower max of 100
	)
	offset, _ := validation.ParsePaginationOffset(r.URL.Query().Get("offset"))

	tenants, err := h.tenantService.List(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("failed to list tenants", "error", err)
		http.Error(w, "failed to list tenants", http.StatusInternalServerError)
		return
	}

	// Get total count for pagination
	total, err := h.tenantService.Count(r.Context())
	if err != nil {
		h.logger.Warn("failed to get tenant count", "error", err)
		total = 0
	}

	response := map[string]interface{}{
		"data": tenants,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTenant handles GET /api/v1/admin/tenants/{id}
func (h *TenantAdminHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	t, err := h.tenantService.GetByID(r.Context(), tenantID)
	if err != nil {
		if err == tenant.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get tenant", "error", err, "tenant_id", tenantID)
		http.Error(w, "failed to get tenant", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// UpdateTenant handles PUT /api/v1/admin/tenants/{id}
func (h *TenantAdminHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	var input tenant.UpdateTenantInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode update tenant request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := ValidateUpdateTenantInput(input); err != nil {
		h.logger.Warn("invalid update tenant input", "error", err)
		http.Error(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	t, err := h.tenantService.Update(r.Context(), tenantID, input)
	if err != nil {
		if err == tenant.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to update tenant", "error", err, "tenant_id", tenantID)
		http.Error(w, "failed to update tenant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// DeleteTenant handles DELETE /api/v1/admin/tenants/{id}
func (h *TenantAdminHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	err := h.tenantService.Delete(r.Context(), tenantID)
	if err != nil {
		if err == tenant.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to delete tenant", "error", err, "tenant_id", tenantID)
		http.Error(w, "failed to delete tenant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateTenantQuotas handles PUT /api/v1/admin/tenants/{id}/quotas
func (h *TenantAdminHandler) UpdateTenantQuotas(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	var quotas tenant.TenantQuotas
	if err := json.NewDecoder(r.Body).Decode(&quotas); err != nil {
		h.logger.Error("failed to decode quotas request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	t, err := h.tenantService.UpdateQuotas(r.Context(), tenantID, quotas)
	if err != nil {
		if err == tenant.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to update tenant quotas", "error", err, "tenant_id", tenantID)
		http.Error(w, "failed to update quotas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant": t,
		"quotas": quotas,
	})
}

// GetTenantUsage handles GET /api/v1/admin/tenants/{id}/usage
func (h *TenantAdminHandler) GetTenantUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	// Get tenant to verify it exists
	t, err := h.tenantService.GetByID(r.Context(), tenantID)
	if err != nil {
		if err == tenant.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get tenant", "error", err, "tenant_id", tenantID)
		http.Error(w, "failed to get tenant", http.StatusInternalServerError)
		return
	}

	// Get usage statistics
	stats, err := h.tenantService.GetExecutionStats(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to get usage stats", "error", err, "tenant_id", tenantID)
		http.Error(w, "failed to get usage stats", http.StatusInternalServerError)
		return
	}

	// Parse quotas for comparison
	var quotas tenant.TenantQuotas
	if err := json.Unmarshal(t.Quotas, &quotas); err != nil {
		h.logger.Error("failed to parse quotas", "error", err, "tenant_id", tenantID)
		quotas = tenant.DefaultQuotas(t.Tier)
	}

	response := map[string]interface{}{
		"tenant_id": tenantID,
		"usage":     stats,
		"quotas":    quotas,
		"utilization": map[string]interface{}{
			"workflows_percentage":             calculatePercentage(stats.WorkflowCount, quotas.MaxWorkflows),
			"executions_today_percentage":      calculatePercentage(stats.ExecutionsToday, quotas.MaxExecutionsPerDay),
			"concurrent_executions_percentage": calculatePercentage(stats.ConcurrentExecutions, quotas.MaxConcurrentExecutions),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// calculatePercentage calculates usage percentage, handling unlimited quotas (-1)
func calculatePercentage(current, max int) float64 {
	if max == -1 {
		return 0.0 // Unlimited
	}
	if max == 0 {
		return 0.0
	}
	percentage := (float64(current) / float64(max)) * 100.0
	if percentage > 100.0 {
		return 100.0
	}
	return percentage
}

// SwitchTenantRequest represents the request to switch tenant context
type SwitchTenantRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
}

// SwitchTenant handles POST /api/v1/admin/switch-tenant
// Allows admin users to switch their tenant context
func (h *TenantAdminHandler) SwitchTenant(w http.ResponseWriter, r *http.Request) {
	var req SwitchTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode switch tenant request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.TenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	// Switch tenant context
	_, targetTenant, err := h.tenantService.SwitchTenant(r.Context(), req.TenantID)
	if err != nil {
		h.logger.Error("failed to switch tenant", "error", err, "target_tenant_id", req.TenantID)
		if err == tenant.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to switch tenant: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"tenant":  targetTenant,
		"message": "Successfully switched to tenant: " + targetTenant.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetTenantStatus handles PUT /api/v1/admin/tenants/{id}/status
func (h *TenantAdminHandler) SetTenantStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	var input struct {
		Status string `json:"status" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode status request", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !tenant.IsValidStatus(input.Status) {
		http.Error(w, "invalid status: must be one of active, inactive, suspended", http.StatusBadRequest)
		return
	}

	t, err := h.tenantService.SetStatus(r.Context(), tenantID, tenant.TenantStatus(input.Status))
	if err != nil {
		if err == tenant.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to set tenant status", "error", err, "tenant_id", tenantID)
		http.Error(w, "failed to set tenant status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// ActivateTenant handles POST /api/v1/admin/tenants/{id}/activate
func (h *TenantAdminHandler) ActivateTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	t, err := h.tenantService.Activate(r.Context(), tenantID)
	if err != nil {
		if err == tenant.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to activate tenant", "error", err, "tenant_id", tenantID)
		http.Error(w, "failed to activate tenant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// SuspendTenant handles POST /api/v1/admin/tenants/{id}/suspend
func (h *TenantAdminHandler) SuspendTenant(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantID")
	if tenantID == "" {
		http.Error(w, "tenant ID is required", http.StatusBadRequest)
		return
	}

	t, err := h.tenantService.Suspend(r.Context(), tenantID)
	if err != nil {
		if err == tenant.ErrNotFound {
			http.Error(w, "tenant not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to suspend tenant", "error", err, "tenant_id", tenantID)
		http.Error(w, "failed to suspend tenant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}
