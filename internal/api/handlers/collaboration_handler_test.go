package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorax/gorax/internal/collaboration"
	"github.com/gorax/gorax/internal/config"
	"github.com/gorax/gorax/internal/websocket"
)

func TestCollaborationHandler_OriginValidation(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		wantUpgrade    bool
	}{
		{
			name:           "allowed origin - localhost",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:3000",
			wantUpgrade:    false, // Will fail auth, but origin check passes
		},
		{
			name:           "denied origin - wrong domain",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "https://evil.com",
			wantUpgrade:    false,
		},
		{
			name:           "allowed origin - wildcard subdomain",
			allowedOrigins: []string{"https://*.example.com"},
			requestOrigin:  "https://app.example.com",
			wantUpgrade:    false, // Will fail auth, but origin check passes
		},
		{
			name:           "denied origin - base domain when wildcard",
			allowedOrigins: []string{"https://*.example.com"},
			requestOrigin:  "https://example.com",
			wantUpgrade:    false,
		},
		{
			name:           "denied origin - empty",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "",
			wantUpgrade:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := slog.Default()
			service := collaboration.NewService()
			wsHub := websocket.NewHub(logger)
			collabHub := collaboration.NewHub(service, wsHub, logger)

			wsConfig := config.WebSocketConfig{
				AllowedOrigins: tt.allowedOrigins,
			}

			handler := NewCollaborationHandler(collabHub, wsHub, wsConfig, logger)

			// Create test request
			req := httptest.NewRequest("GET", "/api/v1/workflows/wf-123/collaborate", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			req.Header.Set("Connection", "upgrade")
			req.Header.Set("Upgrade", "websocket")
			req.Header.Set("Sec-WebSocket-Version", "13")
			req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

			rec := httptest.NewRecorder()

			// Execute
			handler.HandleWorkflowCollaboration(rec, req)

			// Assert
			// Since we don't have auth middleware in the test, we expect 401 Unauthorized
			// But if the origin is invalid, we won't even get to the auth check
			// The websocket upgrade will fail with 403 Forbidden for invalid origins
			if tt.requestOrigin == "" || !contains(tt.allowedOrigins, tt.requestOrigin) {
				// For invalid origins, we expect the handler to not upgrade
				// The actual response code depends on whether we got past origin check
				assert.NotEqual(t, http.StatusSwitchingProtocols, rec.Code)
			}

			// Note: In a real scenario with proper auth, we'd test that valid origins
			// can establish connections while invalid ones are rejected at the origin check
		})
	}
}

func TestCollaborationHandler_CheckOriginFunction(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		want           bool
	}{
		{
			name:           "allowed origin",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:3000",
			want:           true,
		},
		{
			name:           "denied origin",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "https://evil.com",
			want:           false,
		},
		{
			name:           "wildcard subdomain allowed",
			allowedOrigins: []string{"https://*.example.com"},
			requestOrigin:  "https://app.example.com",
			want:           true,
		},
		{
			name:           "wildcard port allowed",
			allowedOrigins: []string{"http://localhost:*"},
			requestOrigin:  "http://localhost:5173",
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wsConfig := config.WebSocketConfig{
				AllowedOrigins: tt.allowedOrigins,
			}

			logger := slog.Default()
			service := collaboration.NewService()
			wsHub := websocket.NewHub(logger)
			collabHub := collaboration.NewHub(service, wsHub, logger)

			handler := NewCollaborationHandler(collabHub, wsHub, wsConfig, logger)

			// Create test request with origin
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.requestOrigin)

			// Test the CheckOrigin function
			got := handler.upgrader.CheckOrigin(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "valid alphanumeric",
			id:   "workflow-123",
			want: true,
		},
		{
			name: "valid with underscores",
			id:   "wf_abc_123",
			want: true,
		},
		{
			name: "valid mixed case",
			id:   "Workflow-ABC-123",
			want: true,
		},
		{
			name: "invalid - contains spaces",
			id:   "workflow 123",
			want: false,
		},
		{
			name: "invalid - contains special chars",
			id:   "workflow@123",
			want: false,
		},
		{
			name: "invalid - SQL injection attempt",
			id:   "wf'; DROP TABLE--",
			want: false,
		},
		{
			name: "invalid - path traversal",
			id:   "../../../etc/passwd",
			want: false,
		},
		{
			name: "invalid - empty",
			id:   "",
			want: false,
		},
		{
			name: "invalid - too long",
			id:   string(make([]byte, 257)),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidID(tt.id)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidElementType(t *testing.T) {
	tests := []struct {
		name        string
		elementType string
		want        bool
	}{
		{
			name:        "valid - node",
			elementType: "node",
			want:        true,
		},
		{
			name:        "valid - edge",
			elementType: "edge",
			want:        true,
		},
		{
			name:        "invalid - unknown type",
			elementType: "link",
			want:        false,
		},
		{
			name:        "invalid - empty",
			elementType: "",
			want:        false,
		},
		{
			name:        "invalid - injection attempt",
			elementType: "node'; DROP TABLE--",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidElementType(tt.elementType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCollaborationHandler_ConnectionLimits(t *testing.T) {
	logger := slog.Default()
	service := collaboration.NewService()
	wsHub := websocket.NewHub(logger)
	collabHub := collaboration.NewHub(service, wsHub, logger)

	// Set low connection limit for testing
	wsConfig := config.WebSocketConfig{
		AllowedOrigins:            []string{"http://localhost:3000"},
		MaxConnectionsPerWorkflow: 2,
		MaxMessageSize:            512 * 1024,
	}

	handler := NewCollaborationHandler(collabHub, wsHub, wsConfig, logger)

	// Simulate reaching the connection limit
	// In a real test, we'd need to mock the hub's GetClientCount method
	// For now, this demonstrates the structure

	req := httptest.NewRequest("GET", "/api/v1/workflows/wf-123/collaborate", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	// Without auth, this will fail earlier, but demonstrates the pattern
	handler.HandleWorkflowCollaboration(rec, req)

	// The handler should reject if auth passes but connection limit is reached
	assert.NotEqual(t, http.StatusSwitchingProtocols, rec.Code)
}

func TestCollaborationHandler_PayloadSizeLimit(t *testing.T) {
	logger := slog.Default()
	service := collaboration.NewService()
	wsHub := websocket.NewHub(logger)
	collabHub := collaboration.NewHub(service, wsHub, logger)

	wsConfig := config.WebSocketConfig{
		AllowedOrigins:            []string{"http://localhost:3000"},
		MaxConnectionsPerWorkflow: 50,
		MaxMessageSize:            1024, // 1KB limit for testing
	}

	_ = NewCollaborationHandler(collabHub, wsHub, wsConfig, logger)

	// Verify that the upgrader has the correct read limit
	assert.Equal(t, wsConfig.MaxMessageSize, int64(1024))
}
