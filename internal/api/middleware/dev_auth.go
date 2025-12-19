package middleware

import (
	"context"
	"net/http"
	"os"
)

// DevAuth returns middleware for development that bypasses Kratos
// IMPORTANT: This should ONLY be used in development (APP_ENV=development)
func DevAuth() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only allow in development mode
			if os.Getenv("APP_ENV") != "development" {
				http.Error(w, "development auth only available in development mode", http.StatusForbidden)
				return
			}

			// Get tenant ID from header (required)
			tenantID := r.Header.Get("X-Tenant-ID")
			if tenantID == "" {
				// Default to test tenant for convenience
				tenantID = "00000000-0000-0000-0000-000000000001"
			}

			// Get user ID from header (optional)
			userID := r.Header.Get("X-User-ID")
			if userID == "" {
				userID = "00000000-0000-0000-0000-000000000002"
			}

			// Create development user
			user := &User{
				ID:       userID,
				Email:    "dev@example.com",
				TenantID: tenantID,
				Traits: map[string]interface{}{
					"email":     "dev@example.com",
					"tenant_id": tenantID,
				},
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
