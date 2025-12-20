package middleware

import (
	"net/http"
)

// AdminRole is the role value that grants admin access
const AdminRole = "admin"

// RequireAdmin returns middleware that checks if the user has admin role
func RequireAdmin() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r)
			if user == nil {
				http.Error(w, "unauthorized: not authenticated", http.StatusUnauthorized)
				return
			}

			if !IsAdmin(user) {
				http.Error(w, "forbidden: admin access required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IsAdmin checks if a user has admin role
func IsAdmin(user *User) bool {
	if user == nil {
		return false
	}

	// Check role in traits
	role, ok := user.Traits["role"].(string)
	if ok && role == AdminRole {
		return true
	}

	// Check roles array in traits (for systems that use arrays)
	roles, ok := user.Traits["roles"].([]interface{})
	if ok {
		for _, r := range roles {
			if roleStr, ok := r.(string); ok && roleStr == AdminRole {
				return true
			}
		}
	}

	return false
}
