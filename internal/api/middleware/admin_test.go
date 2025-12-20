package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name:     "nil user returns false",
			user:     nil,
			expected: false,
		},
		{
			name: "user without role returns false",
			user: &User{
				ID:     "user-1",
				Email:  "user@example.com",
				Traits: map[string]interface{}{},
			},
			expected: false,
		},
		{
			name: "user with admin role returns true",
			user: &User{
				ID:    "admin-1",
				Email: "admin@example.com",
				Traits: map[string]interface{}{
					"role": "admin",
				},
			},
			expected: true,
		},
		{
			name: "user with non-admin role returns false",
			user: &User{
				ID:    "user-2",
				Email: "user@example.com",
				Traits: map[string]interface{}{
					"role": "user",
				},
			},
			expected: false,
		},
		{
			name: "user with admin in roles array returns true",
			user: &User{
				ID:    "admin-2",
				Email: "admin@example.com",
				Traits: map[string]interface{}{
					"roles": []interface{}{"user", "admin"},
				},
			},
			expected: true,
		},
		{
			name: "user with non-admin roles array returns false",
			user: &User{
				ID:    "user-3",
				Email: "user@example.com",
				Traits: map[string]interface{}{
					"roles": []interface{}{"user", "viewer"},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAdmin(tt.user)
			if result != tt.expected {
				t.Errorf("IsAdmin() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	// Handler that returns 200 OK
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		user           *User
		expectedStatus int
	}{
		{
			name:           "no user returns 401",
			user:           nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "non-admin user returns 403",
			user: &User{
				ID:     "user-1",
				Email:  "user@example.com",
				Traits: map[string]interface{}{},
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user returns 200",
			user: &User{
				ID:    "admin-1",
				Email: "admin@example.com",
				Traits: map[string]interface{}{
					"role": "admin",
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/admin/tenants", nil)
			if tt.user != nil {
				ctx := context.WithValue(req.Context(), UserContextKey, tt.user)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Apply middleware
			handler := RequireAdmin()(successHandler)
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}
