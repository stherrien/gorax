package rbac

import (
	"context"
	"encoding/json"
	"net/http"
)

// PermissionCheck represents a permission requirement
type PermissionCheck struct {
	Resource string
	Action   string
}

// RequirePermission returns middleware that checks for a specific permission
func RequirePermission(repo RepositoryInterface, resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("user_id").(string)
			if !ok || userID == "" {
				respondError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			tenantID, ok := r.Context().Value("tenant_id").(string)
			if !ok || tenantID == "" {
				respondError(w, http.StatusBadRequest, "Missing tenant context")
				return
			}

			has, err := repo.HasPermission(r.Context(), userID, tenantID, resource, action)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to check permission")
				return
			}

			if !has {
				respondError(w, http.StatusForbidden, "Permission denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission returns middleware that checks if user has ANY of the specified permissions
func RequireAnyPermission(repo RepositoryInterface, permissions ...PermissionCheck) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("user_id").(string)
			if !ok || userID == "" {
				respondError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			tenantID, ok := r.Context().Value("tenant_id").(string)
			if !ok || tenantID == "" {
				respondError(w, http.StatusBadRequest, "Missing tenant context")
				return
			}

			// Check if user has any of the required permissions
			for _, perm := range permissions {
				has, err := repo.HasPermission(r.Context(), userID, tenantID, perm.Resource, perm.Action)
				if err != nil {
					respondError(w, http.StatusInternalServerError, "Failed to check permission")
					return
				}
				if has {
					next.ServeHTTP(w, r)
					return
				}
			}

			respondError(w, http.StatusForbidden, "Permission denied")
		})
	}
}

// RequireAllPermissions returns middleware that checks if user has ALL of the specified permissions
func RequireAllPermissions(repo RepositoryInterface, permissions ...PermissionCheck) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("user_id").(string)
			if !ok || userID == "" {
				respondError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			tenantID, ok := r.Context().Value("tenant_id").(string)
			if !ok || tenantID == "" {
				respondError(w, http.StatusBadRequest, "Missing tenant context")
				return
			}

			// Check if user has all of the required permissions
			for _, perm := range permissions {
				has, err := repo.HasPermission(r.Context(), userID, tenantID, perm.Resource, perm.Action)
				if err != nil {
					respondError(w, http.StatusInternalServerError, "Failed to check permission")
					return
				}
				if !has {
					respondError(w, http.StatusForbidden, "Permission denied")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole returns middleware that checks if user has a specific role
func RequireRole(repo RepositoryInterface, roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("user_id").(string)
			if !ok || userID == "" {
				respondError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			tenantID, ok := r.Context().Value("tenant_id").(string)
			if !ok || tenantID == "" {
				respondError(w, http.StatusBadRequest, "Missing tenant context")
				return
			}

			roles, err := repo.GetUserRoles(r.Context(), userID, tenantID)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to check role")
				return
			}

			for _, role := range roles {
				if role.Name == roleName {
					next.ServeHTTP(w, r)
					return
				}
			}

			respondError(w, http.StatusForbidden, "Permission denied")
		})
	}
}

// WithUserPermissions adds user permissions to request context
func WithUserPermissions(repo RepositoryInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("user_id").(string)
			if !ok || userID == "" {
				next.ServeHTTP(w, r)
				return
			}

			tenantID, ok := r.Context().Value("tenant_id").(string)
			if !ok || tenantID == "" {
				next.ServeHTTP(w, r)
				return
			}

			permissions, err := repo.GetUserPermissions(r.Context(), userID, tenantID)
			if err != nil {
				// Don't fail request, just continue without permissions in context
				next.ServeHTTP(w, r)
				return
			}

			// Add permissions to context
			ctx := context.WithValue(r.Context(), "user_permissions", permissions)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserPermissionsFromContext retrieves user permissions from request context
func GetUserPermissionsFromContext(ctx context.Context) []Permission {
	permissions, ok := ctx.Value("user_permissions").([]Permission)
	if !ok {
		return []Permission{}
	}
	return permissions
}

// respondError writes an error response
func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck // can't recover from encode error
		"error": message,
	})
}
