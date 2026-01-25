package middleware

import (
	"context"
	"net/http"

	"github.com/gorax/gorax/internal/config"
	"github.com/gorax/gorax/internal/pkg/tenantctx"
	"github.com/gorax/gorax/internal/tenant"
)

const (
	// TenantContextKey is the context key for the current tenant
	TenantContextKey contextKey = "tenant"
)

// TenantMiddlewareConfig holds configuration for the tenant middleware
type TenantMiddlewareConfig struct {
	// TenantConfig from application config
	TenantConfig config.TenantConfig
}

// TenantContext middleware extracts and validates the tenant from the request
// For backwards compatibility, this version assumes multi-tenant mode
func TenantContext(tenantService *tenant.Service) func(next http.Handler) http.Handler {
	return TenantContextWithConfig(tenantService, TenantMiddlewareConfig{
		TenantConfig: config.TenantConfig{
			Mode: "multi",
		},
	})
}

// TenantContextWithConfig middleware extracts and validates the tenant from the request
// with support for single-tenant mode configuration
func TenantContextWithConfig(tenantService *tenant.Service, cfg TenantMiddlewareConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get authenticated user
			user := GetUser(r)
			if user == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			var t *tenant.Tenant
			var err error

			// Handle single-tenant mode
			if cfg.TenantConfig.IsSingleTenantMode() {
				t, err = resolveSingleTenantMode(r.Context(), tenantService, cfg.TenantConfig)
				if err != nil {
					http.Error(w, "failed to resolve default tenant", http.StatusInternalServerError)
					return
				}
			} else {
				// Multi-tenant mode: resolve tenant from user or header
				tenantID := resolveTenantID(user, r, cfg.TenantConfig)
				if tenantID == "" {
					http.Error(w, "tenant not found", http.StatusBadRequest)
					return
				}

				// Validate tenant exists and is active
				t, err = tenantService.GetByID(r.Context(), tenantID)
				if err != nil {
					http.Error(w, "tenant not found", http.StatusNotFound)
					return
				}
			}

			if t.Status != string(tenant.StatusActive) {
				http.Error(w, "tenant is not active", http.StatusForbidden)
				return
			}

			// Add tenant to context using both the middleware key and tenantctx package
			// This ensures compatibility with both middleware.GetTenantID() and tenantctx.GetTenantID()
			ctx := context.WithValue(r.Context(), TenantContextKey, t)
			ctx = tenantctx.WithTenantID(ctx, t.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// resolveSingleTenantMode resolves the tenant for single-tenant deployments
func resolveSingleTenantMode(ctx context.Context, tenantService *tenant.Service, cfg config.TenantConfig) (*tenant.Tenant, error) {
	// If a specific default tenant ID is configured, use it
	if cfg.DefaultTenantID != "" {
		return tenantService.GetByID(ctx, cfg.DefaultTenantID)
	}

	// Otherwise, get or create the default tenant
	return tenantService.GetOrCreateDefault(ctx)
}

// resolveTenantID determines the tenant ID based on configuration and request
func resolveTenantID(user *User, r *http.Request, cfg config.TenantConfig) string {
	var tenantID string

	// Resolution based on configured strategy
	switch cfg.ResolutionStrategy {
	case "header":
		// Header-only resolution (for API-first deployments)
		tenantID = r.Header.Get("X-Tenant-ID")
	case "subdomain":
		// Subdomain-based resolution (e.g., acme.gorax.com)
		tenantID = extractTenantFromSubdomain(r)
	case "path":
		// Path-based resolution (e.g., /api/v1/tenant/{tenantID}/...)
		// This would be handled by the router, so fall back to user tenant
		tenantID = user.TenantID
	default:
		// Default "user" strategy: use tenant from authenticated user
		tenantID = user.TenantID
	}

	// Allow admin users to override tenant via header if cross-tenant access is allowed
	if cfg.AllowCrossTenantAccess {
		if headerTenantID := r.Header.Get("X-Tenant-ID"); headerTenantID != "" {
			// Verify user has access to this tenant
			if VerifyTenantAccess(user, headerTenantID) {
				tenantID = headerTenantID
			}
		}
	}

	return tenantID
}

// extractTenantFromSubdomain extracts tenant identifier from request subdomain
// For example, "acme.gorax.com" would return "acme"
// Requires at least 3 parts (subdomain.domain.tld) to identify a subdomain
func extractTenantFromSubdomain(r *http.Request) string {
	host := r.Host

	// Remove port if present
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			host = host[:i]
			break
		}
		if host[i] == '.' || host[i] == '[' {
			break
		}
	}

	// Count dots to determine if we have a subdomain
	// We need at least 2 dots for a subdomain (e.g., acme.gorax.com has 2 dots)
	dotCount := 0
	for i := 0; i < len(host); i++ {
		if host[i] == '.' {
			dotCount++
		}
	}

	// Need at least 2 dots for a true subdomain
	// e.g., "gorax.com" has 1 dot -> no subdomain
	// e.g., "acme.gorax.com" has 2 dots -> "acme" is subdomain
	if dotCount < 2 {
		return ""
	}

	// Extract the first part as subdomain
	for i := 0; i < len(host); i++ {
		if host[i] == '.' {
			subdomain := host[:i]
			// Ignore common non-tenant subdomains
			if subdomain != "www" && subdomain != "api" && subdomain != "app" {
				return subdomain
			}
			return ""
		}
	}

	return ""
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
