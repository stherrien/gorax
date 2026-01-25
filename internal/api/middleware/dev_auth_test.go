package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDevAuth(t *testing.T) {
	tests := []struct {
		name             string
		envSetup         func()
		headers          map[string]string
		expectedStatus   int
		checkContext     func(*testing.T, *http.Request)
		contextCaptured  bool
	}{
		{
			name: "not allowed in non-development mode",
			envSetup: func() {
				os.Setenv("APP_ENV", "production")
			},
			headers:        map[string]string{},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "not allowed when APP_ENV not set",
			envSetup: func() {
				os.Unsetenv("APP_ENV")
			},
			headers:        map[string]string{},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "allowed in development mode with defaults",
			envSetup: func() {
				os.Setenv("APP_ENV", "development")
			},
			headers:         map[string]string{},
			expectedStatus:  http.StatusOK,
			contextCaptured: true,
			checkContext: func(t *testing.T, r *http.Request) {
				user := GetUser(r)
				assert.NotNil(t, user)
				assert.Equal(t, "00000000-0000-0000-0000-000000000001", user.TenantID)
				assert.Equal(t, "00000000-0000-0000-0000-000000000002", user.ID)
				assert.Equal(t, "dev@example.com", user.Email)
			},
		},
		{
			name: "custom tenant ID from header",
			envSetup: func() {
				os.Setenv("APP_ENV", "development")
			},
			headers: map[string]string{
				"X-Tenant-ID": "custom-tenant-123",
			},
			expectedStatus:  http.StatusOK,
			contextCaptured: true,
			checkContext: func(t *testing.T, r *http.Request) {
				user := GetUser(r)
				assert.NotNil(t, user)
				assert.Equal(t, "custom-tenant-123", user.TenantID)
			},
		},
		{
			name: "custom user ID from header",
			envSetup: func() {
				os.Setenv("APP_ENV", "development")
			},
			headers: map[string]string{
				"X-User-ID": "custom-user-456",
			},
			expectedStatus:  http.StatusOK,
			contextCaptured: true,
			checkContext: func(t *testing.T, r *http.Request) {
				user := GetUser(r)
				assert.NotNil(t, user)
				assert.Equal(t, "custom-user-456", user.ID)
			},
		},
		{
			name: "custom role from header",
			envSetup: func() {
				os.Setenv("APP_ENV", "development")
			},
			headers: map[string]string{
				"X-User-Role": "admin",
			},
			expectedStatus:  http.StatusOK,
			contextCaptured: true,
			checkContext: func(t *testing.T, r *http.Request) {
				user := GetUser(r)
				assert.NotNil(t, user)
				assert.Equal(t, "admin", user.Traits["role"])
			},
		},
		{
			name: "all custom headers",
			envSetup: func() {
				os.Setenv("APP_ENV", "development")
			},
			headers: map[string]string{
				"X-Tenant-ID": "tenant-abc",
				"X-User-ID":   "user-xyz",
				"X-User-Role": "super-admin",
			},
			expectedStatus:  http.StatusOK,
			contextCaptured: true,
			checkContext: func(t *testing.T, r *http.Request) {
				user := GetUser(r)
				assert.NotNil(t, user)
				assert.Equal(t, "tenant-abc", user.TenantID)
				assert.Equal(t, "user-xyz", user.ID)
				assert.Equal(t, "super-admin", user.Traits["role"])
				assert.Equal(t, "tenant-abc", user.Traits["tenant_id"])
				assert.Equal(t, "dev@example.com", user.Traits["email"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env
			originalEnv := os.Getenv("APP_ENV")
			defer os.Setenv("APP_ENV", originalEnv)

			// Set up test environment
			tt.envSetup()

			// Create request with headers
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()

			// Capture request for context checking
			var capturedReq *http.Request
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedReq = r
				w.WriteHeader(http.StatusOK)
			})

			// Execute middleware
			middleware := DevAuth()
			middleware(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.checkContext != nil && capturedReq != nil {
				tt.checkContext(t, capturedReq)
			}
		})
	}
}

func TestDevAuth_TraitsContainsExpectedFields(t *testing.T) {
	// Save original env
	originalEnv := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", originalEnv)

	os.Setenv("APP_ENV", "development")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	var capturedReq *http.Request
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	})

	middleware := DevAuth()
	middleware(nextHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	user := GetUser(capturedReq)
	assert.NotNil(t, user)

	// Verify traits structure
	assert.Contains(t, user.Traits, "email")
	assert.Contains(t, user.Traits, "tenant_id")
	assert.Equal(t, "dev@example.com", user.Traits["email"])
	assert.Equal(t, user.TenantID, user.Traits["tenant_id"])
}
