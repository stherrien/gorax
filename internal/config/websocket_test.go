package config

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOrigin(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		want           bool
	}{
		{
			name:           "exact match - localhost",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:3000",
			want:           true,
		},
		{
			name:           "exact match - production domain",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://example.com",
			want:           true,
		},
		{
			name:           "no match - wrong port",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:5173",
			want:           false,
		},
		{
			name:           "no match - wrong protocol",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "http://example.com",
			want:           false,
		},
		{
			name:           "wildcard subdomain - match",
			allowedOrigins: []string{"https://*.example.com"},
			requestOrigin:  "https://app.example.com",
			want:           true,
		},
		{
			name:           "wildcard subdomain - nested match",
			allowedOrigins: []string{"https://*.example.com"},
			requestOrigin:  "https://dev.app.example.com",
			want:           true,
		},
		{
			name:           "wildcard subdomain - no match base domain",
			allowedOrigins: []string{"https://*.example.com"},
			requestOrigin:  "https://example.com",
			want:           false,
		},
		{
			name:           "wildcard subdomain - no match different domain",
			allowedOrigins: []string{"https://*.example.com"},
			requestOrigin:  "https://app.evil.com",
			want:           false,
		},
		{
			name:           "multiple allowed origins - first matches",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com"},
			requestOrigin:  "http://localhost:3000",
			want:           true,
		},
		{
			name:           "multiple allowed origins - second matches",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com"},
			requestOrigin:  "https://example.com",
			want:           true,
		},
		{
			name:           "multiple allowed origins - no match",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com"},
			requestOrigin:  "https://evil.com",
			want:           false,
		},
		{
			name:           "empty origin - denied",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "",
			want:           false,
		},
		{
			name:           "empty allowed origins - denied",
			allowedOrigins: []string{},
			requestOrigin:  "http://localhost:3000",
			want:           false,
		},
		{
			name:           "wildcard port - match",
			allowedOrigins: []string{"http://localhost:*"},
			requestOrigin:  "http://localhost:3000",
			want:           true,
		},
		{
			name:           "wildcard port - different port match",
			allowedOrigins: []string{"http://localhost:*"},
			requestOrigin:  "http://localhost:5173",
			want:           true,
		},
		{
			name:           "case insensitive domain",
			allowedOrigins: []string{"https://Example.COM"},
			requestOrigin:  "https://example.com",
			want:           true,
		},
		{
			name:           "invalid origin format - denied",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "not-a-valid-origin",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WebSocketConfig{
				AllowedOrigins: tt.allowedOrigins,
			}

			// Create request with origin header
			req, err := http.NewRequest("GET", "http://example.com/ws", nil)
			require.NoError(t, err)
			req.Header.Set("Origin", tt.requestOrigin)

			got := ValidateOrigin(config, req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWebSocketConfig_CheckOrigin(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WebSocketConfig{
				AllowedOrigins: tt.allowedOrigins,
			}

			req, err := http.NewRequest("GET", "http://example.com/ws", nil)
			require.NoError(t, err)
			req.Header.Set("Origin", tt.requestOrigin)

			checkOrigin := config.CheckOrigin()
			got := checkOrigin(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadWebSocketConfig(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     []string
	}{
		{
			name:     "default localhost origins",
			envValue: "",
			want: []string{
				"http://localhost:3000",
				"http://localhost:5173",
				"http://localhost:5174",
			},
		},
		{
			name:     "custom single origin",
			envValue: "https://example.com",
			want:     []string{"https://example.com"},
		},
		{
			name:     "multiple custom origins",
			envValue: "https://example.com,https://app.example.com,http://localhost:3000",
			want:     []string{"https://example.com", "https://app.example.com", "http://localhost:3000"},
		},
		{
			name:     "origins with whitespace",
			envValue: " https://example.com , https://app.example.com ",
			want:     []string{"https://example.com", "https://app.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				t.Setenv("WEBSOCKET_ALLOWED_ORIGINS", tt.envValue)
			}

			config := loadWebSocketConfig()
			assert.Equal(t, tt.want, config.AllowedOrigins)
		})
	}
}
