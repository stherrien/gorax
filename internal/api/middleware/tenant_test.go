package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorax/gorax/internal/config"
)

func TestExtractTenantFromSubdomain(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{
			name:     "valid subdomain",
			host:     "acme.gorax.com",
			expected: "acme",
		},
		{
			name:     "subdomain with port",
			host:     "acme.gorax.com:8080",
			expected: "acme",
		},
		{
			name:     "www subdomain should be ignored",
			host:     "www.gorax.com",
			expected: "",
		},
		{
			name:     "api subdomain should be ignored",
			host:     "api.gorax.com",
			expected: "",
		},
		{
			name:     "app subdomain should be ignored",
			host:     "app.gorax.com",
			expected: "",
		},
		{
			name:     "no subdomain - needs at least 2 dots",
			host:     "gorax.com",
			expected: "", // Single domain without subdomain
		},
		{
			name:     "localhost with port",
			host:     "localhost:8080",
			expected: "",
		},
		{
			name:     "tenant with hyphen",
			host:     "acme-corp.gorax.com",
			expected: "acme-corp",
		},
		{
			name:     "numeric subdomain",
			host:     "123.gorax.com",
			expected: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://"+tt.host+"/api/v1/test", nil)
			result := extractTenantFromSubdomain(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveTenantID(t *testing.T) {
	tests := []struct {
		name       string
		user       *User
		headerID   string
		cfg        config.TenantConfig
		expectedID string
	}{
		{
			name: "user strategy - uses user tenant ID",
			user: &User{
				ID:       "user-1",
				TenantID: "tenant-from-user",
			},
			cfg: config.TenantConfig{
				ResolutionStrategy: "user",
			},
			expectedID: "tenant-from-user",
		},
		{
			name: "header strategy - uses X-Tenant-ID header",
			user: &User{
				ID:       "user-1",
				TenantID: "tenant-from-user",
			},
			headerID: "tenant-from-header",
			cfg: config.TenantConfig{
				ResolutionStrategy: "header",
			},
			expectedID: "tenant-from-header",
		},
		{
			name: "header strategy - empty header returns empty",
			user: &User{
				ID:       "user-1",
				TenantID: "tenant-from-user",
			},
			headerID: "",
			cfg: config.TenantConfig{
				ResolutionStrategy: "header",
			},
			expectedID: "",
		},
		{
			name: "path strategy - falls back to user tenant",
			user: &User{
				ID:       "user-1",
				TenantID: "tenant-from-user",
			},
			cfg: config.TenantConfig{
				ResolutionStrategy: "path",
			},
			expectedID: "tenant-from-user",
		},
		{
			name: "admin can override with header when cross-tenant allowed",
			user: &User{
				ID:       "admin-1",
				TenantID: "tenant-1",
				Traits: map[string]any{
					"role": "admin",
				},
			},
			headerID: "tenant-2",
			cfg: config.TenantConfig{
				ResolutionStrategy:     "user",
				AllowCrossTenantAccess: true,
			},
			expectedID: "tenant-2",
		},
		{
			name: "non-admin cannot override with header",
			user: &User{
				ID:       "user-1",
				TenantID: "tenant-1",
			},
			headerID: "tenant-2",
			cfg: config.TenantConfig{
				ResolutionStrategy:     "user",
				AllowCrossTenantAccess: true,
			},
			expectedID: "tenant-1", // stays as user's tenant since they can't access tenant-2
		},
		{
			name: "empty resolution strategy defaults to user",
			user: &User{
				ID:       "user-1",
				TenantID: "tenant-from-user",
			},
			cfg: config.TenantConfig{
				ResolutionStrategy: "",
			},
			expectedID: "tenant-from-user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/test", nil)
			if tt.headerID != "" {
				req.Header.Set("X-Tenant-ID", tt.headerID)
			}

			result := resolveTenantID(tt.user, req, tt.cfg)
			assert.Equal(t, tt.expectedID, result)
		})
	}
}

func TestTenantConfig_IsSingleTenantMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{
			name:     "single mode",
			mode:     "single",
			expected: true,
		},
		{
			name:     "multi mode",
			mode:     "multi",
			expected: false,
		},
		{
			name:     "empty mode defaults to multi",
			mode:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.TenantConfig{
				Mode: tt.mode,
			}
			assert.Equal(t, tt.expected, cfg.IsSingleTenantMode())
		})
	}
}
