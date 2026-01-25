package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/tenant"
)

// TenantHandler handles tenant-related endpoints for regular users
type TenantHandler struct {
	tenantService *tenant.Service
	logger        *slog.Logger
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *tenant.Service, logger *slog.Logger) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		logger:        logger,
	}
}

// TenantInfoResponse represents the response for tenant info endpoint
type TenantInfoResponse struct {
	Tenant *tenant.Tenant       `json:"tenant"`
	Quotas *tenant.TenantQuotas `json:"quotas"`
	Usage  *tenant.UsageStats   `json:"usage"`
}

// GetCurrentTenant handles GET /api/v1/tenant/info
// Returns information about the current tenant from the request context
func (h *TenantHandler) GetCurrentTenant(w http.ResponseWriter, r *http.Request) {
	// Get tenant from middleware context
	t := middleware.GetTenant(r)
	if t == nil {
		h.logger.Error("no tenant in context")
		http.Error(w, "no tenant context", http.StatusBadRequest)
		return
	}

	// Parse quotas
	quotas, err := t.GetQuotas()
	if err != nil {
		h.logger.Warn("failed to parse tenant quotas", "error", err, "tenant_id", t.ID)
		// Use default quotas based on tier
		defaultQuotas := tenant.DefaultQuotas(t.Tier)
		quotas = &defaultQuotas
	}

	// Get usage stats
	usage, err := h.tenantService.GetExecutionStats(r.Context(), t.ID)
	if err != nil {
		h.logger.Warn("failed to get usage stats", "error", err, "tenant_id", t.ID)
		// Return nil usage if stats can't be retrieved
		usage = nil
	}

	response := TenantInfoResponse{
		Tenant: t,
		Quotas: quotas,
		Usage:  usage,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// GetTenantSettings handles GET /api/v1/tenant/settings
// Returns the settings for the current tenant
func (h *TenantHandler) GetTenantSettings(w http.ResponseWriter, r *http.Request) {
	t := middleware.GetTenant(r)
	if t == nil {
		http.Error(w, "no tenant context", http.StatusBadRequest)
		return
	}

	settings, err := t.GetSettings()
	if err != nil {
		h.logger.Warn("failed to parse tenant settings", "error", err, "tenant_id", t.ID)
		// Return default settings
		settings = &tenant.TenantSettings{
			DefaultTimezone: "UTC",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// GetTenantQuotas handles GET /api/v1/tenant/quotas
// Returns the quotas for the current tenant
func (h *TenantHandler) GetTenantQuotas(w http.ResponseWriter, r *http.Request) {
	t := middleware.GetTenant(r)
	if t == nil {
		http.Error(w, "no tenant context", http.StatusBadRequest)
		return
	}

	quotas, err := t.GetQuotas()
	if err != nil {
		h.logger.Warn("failed to parse tenant quotas", "error", err, "tenant_id", t.ID)
		defaultQuotas := tenant.DefaultQuotas(t.Tier)
		quotas = &defaultQuotas
	}

	// Get current usage for comparison
	usage, err := h.tenantService.GetExecutionStats(r.Context(), t.ID)
	if err != nil {
		h.logger.Warn("failed to get usage stats", "error", err, "tenant_id", t.ID)
		usage = &tenant.UsageStats{}
	}

	response := map[string]any{
		"quotas": quotas,
		"usage":  usage,
		"utilization": map[string]float64{
			"workflows_percentage":             calculatePercentage(usage.WorkflowCount, quotas.MaxWorkflows),
			"executions_today_percentage":      calculatePercentage(usage.ExecutionsToday, quotas.MaxExecutionsPerDay),
			"concurrent_executions_percentage": calculatePercentage(usage.ConcurrentExecutions, quotas.MaxConcurrentExecutions),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
