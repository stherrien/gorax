package middleware

import (
	"context"
	"net/http"

	"github.com/gorax/gorax/internal/tenant"
)

const (
	// TenantContextKey is the context key for the current tenant
	TenantContextKey contextKey = "tenant"
)

// TenantContext middleware extracts and validates the tenant from the request
func TenantContext(tenantService *tenant.Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get authenticated user
			user := GetUser(r)
			if user == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			var tenantID string

			// Priority 1: User's tenant from Kratos session
			if user.TenantID != "" {
				tenantID = user.TenantID
			}

			// Priority 2: X-Tenant-ID header (for admin/multi-tenant users)
			if headerTenantID := r.Header.Get("X-Tenant-ID"); headerTenantID != "" {
				// TODO: Verify user has access to this tenant
				tenantID = headerTenantID
			}

			if tenantID == "" {
				http.Error(w, "tenant not found", http.StatusBadRequest)
				return
			}

			// Validate tenant exists and is active
			t, err := tenantService.GetByID(r.Context(), tenantID)
			if err != nil {
				http.Error(w, "tenant not found", http.StatusNotFound)
				return
			}

			if t.Status != "active" {
				http.Error(w, "tenant is not active", http.StatusForbidden)
				return
			}

			// Add tenant to context
			ctx := context.WithValue(r.Context(), TenantContextKey, t)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetTenant extracts the tenant from the request context
func GetTenant(r *http.Request) *tenant.Tenant {
	t, _ := r.Context().Value(TenantContextKey).(*tenant.Tenant)
	return t
}

// GetTenantID extracts just the tenant ID from the request context
func GetTenantID(r *http.Request) string {
	t := GetTenant(r)
	if t != nil {
		return t.ID
	}
	return ""
}
