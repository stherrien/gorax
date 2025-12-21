package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/config"
)

func TestValidateCORSConfig(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.CORSConfig
		env         string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid development config with localhost",
			cfg: config.CORSConfig{
				AllowedOrigins:   []string{"http://localhost:3000"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: false,
				MaxAge:           300,
			},
			env:         "development",
			expectError: false,
		},
		{
			name: "valid production config",
			cfg: config.CORSConfig{
				AllowedOrigins:   []string{"https://app.gorax.io"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: true,
				MaxAge:           300,
			},
			env:         "production",
			expectError: false,
		},
		{
			name: "production rejects wildcard origin",
			cfg: config.CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
			},
			env:         "production",
			expectError: true,
			errorMsg:    "wildcard (*) origin not allowed in production",
		},
		{
			name: "production rejects localhost",
			cfg: config.CORSConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
			},
			env:         "production",
			expectError: true,
			errorMsg:    "localhost origins not allowed in production",
		},
		{
			name: "production rejects 127.0.0.1",
			cfg: config.CORSConfig{
				AllowedOrigins: []string{"http://127.0.0.1:3000"},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
			},
			env:         "production",
			expectError: true,
			errorMsg:    "localhost origins not allowed in production",
		},
		{
			name: "production with credentials and broad origins",
			cfg: config.CORSConfig{
				AllowedOrigins:   []string{"https://app1.com", "https://app2.com", "https://app3.com"},
				AllowedMethods:   []string{"GET"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: true,
			},
			env:         "production",
			expectError: false, // Should not error, but will log warning
		},
		{
			name: "empty origins",
			cfg: config.CORSConfig{
				AllowedOrigins: []string{},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
			},
			env:         "production",
			expectError: true,
			errorMsg:    "at least one allowed origin must be specified",
		},
		{
			name: "empty methods",
			cfg: config.CORSConfig{
				AllowedOrigins: []string{"https://app.com"},
				AllowedMethods: []string{},
				AllowedHeaders: []string{"Content-Type"},
			},
			env:         "production",
			expectError: true,
			errorMsg:    "at least one allowed method must be specified",
		},
		{
			name: "negative max age",
			cfg: config.CORSConfig{
				AllowedOrigins: []string{"https://app.com"},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
				MaxAge:         -1,
			},
			env:         "production",
			expectError: true,
			errorMsg:    "max age must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCORSConfig(tt.cfg, tt.env)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name          string
		cfg           config.CORSConfig
		env           string
		requestOrigin string
		requestMethod string
		isPreflight   bool
		expectOrigin  string
		expectMethods string
		expectHeaders string
		expectMaxAge  string
		expectCreds   string
		expectStatus  int
	}{
		{
			name: "simple request with allowed origin",
			cfg: config.CORSConfig{
				AllowedOrigins: []string{"https://app.gorax.io"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
				MaxAge:         300,
			},
			env:           "production",
			requestOrigin: "https://app.gorax.io",
			requestMethod: "GET",
			isPreflight:   false,
			expectOrigin:  "https://app.gorax.io",
			expectStatus:  http.StatusOK,
		},
		{
			name: "simple request with disallowed origin",
			cfg: config.CORSConfig{
				AllowedOrigins: []string{"https://app.gorax.io"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
			},
			env:           "production",
			requestOrigin: "https://malicious.com",
			requestMethod: "GET",
			isPreflight:   false,
			expectOrigin:  "", // No CORS headers should be set
			expectStatus:  http.StatusOK,
		},
		{
			name: "preflight request with allowed origin",
			cfg: config.CORSConfig{
				AllowedOrigins: []string{"https://app.gorax.io"},
				AllowedMethods: []string{"GET", "POST", "DELETE"},
				AllowedHeaders: []string{"Content-Type", "Authorization"},
				MaxAge:         300,
			},
			env:           "production",
			requestOrigin: "https://app.gorax.io",
			requestMethod: "DELETE",
			isPreflight:   true,
			expectOrigin:  "https://app.gorax.io",
			expectMethods: "GET, POST, DELETE",
			expectHeaders: "Content-Type, Authorization",
			expectMaxAge:  "300",
			expectStatus:  http.StatusOK,
		},
		{
			name: "request with credentials",
			cfg: config.CORSConfig{
				AllowedOrigins:   []string{"https://app.gorax.io"},
				AllowedMethods:   []string{"GET"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: true,
			},
			env:           "production",
			requestOrigin: "https://app.gorax.io",
			requestMethod: "GET",
			isPreflight:   false,
			expectOrigin:  "https://app.gorax.io",
			expectCreds:   "true",
			expectStatus:  http.StatusOK,
		},
		{
			name: "wildcard origin in development",
			cfg: config.CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET"},
				AllowedHeaders: []string{"Content-Type"},
			},
			env:           "development",
			requestOrigin: "http://localhost:3000",
			requestMethod: "GET",
			isPreflight:   false,
			expectOrigin:  "*",
			expectStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CORS middleware
			corsMiddleware, err := NewCORSMiddleware(tt.cfg, tt.env)
			require.NoError(t, err)

			// Wrap handler
			wrappedHandler := corsMiddleware(handler)

			// Create request
			var req *http.Request
			if tt.isPreflight {
				req = httptest.NewRequest(http.MethodOptions, "/test", nil)
				req.Header.Set("Access-Control-Request-Method", tt.requestMethod)
			} else {
				req = httptest.NewRequest(tt.requestMethod, "/test", nil)
			}
			req.Header.Set("Origin", tt.requestOrigin)

			// Record response
			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			// Assert status
			assert.Equal(t, tt.expectStatus, rr.Code)

			// Assert CORS headers
			if tt.expectOrigin != "" {
				assert.Equal(t, tt.expectOrigin, rr.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
			}

			if tt.expectMethods != "" {
				assert.Equal(t, tt.expectMethods, rr.Header().Get("Access-Control-Allow-Methods"))
			}

			if tt.expectHeaders != "" {
				assert.Equal(t, tt.expectHeaders, rr.Header().Get("Access-Control-Allow-Headers"))
			}

			if tt.expectMaxAge != "" {
				assert.Equal(t, tt.expectMaxAge, rr.Header().Get("Access-Control-Max-Age"))
			}

			if tt.expectCreds != "" {
				assert.Equal(t, tt.expectCreds, rr.Header().Get("Access-Control-Allow-Credentials"))
			}
		})
	}
}

func TestNewCORSMiddleware_InvalidConfig(t *testing.T) {
	cfg := config.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}

	// Should fail in production with wildcard
	_, err := NewCORSMiddleware(cfg, "production")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wildcard (*) origin not allowed in production")
}
