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
				// Verify user has access to this tenant
				if !VerifyTenantAccess(user, headerTenantID) {
					http.Error(w, "access denied: user does not have access to this tenant", http.StatusForbidden)
					return
				}
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

// VerifyTenantAccess checks if a user has access to a specific tenant
// Returns true if:
// - The tenant ID matches the user's tenant ID
// - The user is an admin (admins have access to all tenants)
func VerifyTenantAccess(user *User, tenantID string) bool {
	if user == nil {
		return false
	}

	// Admins have access to all tenants
	if IsAdmin(user) {
		return true
	}

	// Users can access their own tenant
	return user.TenantID == tenantID
}
