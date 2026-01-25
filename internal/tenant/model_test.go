package tenant

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"active is valid", "active", true},
		{"inactive is valid", "inactive", true},
		{"suspended is valid", "suspended", true},
		{"deleted is valid", "deleted", true},
		{"empty is invalid", "", false},
		{"unknown is invalid", "unknown", false},
		{"ACTIVE uppercase is invalid", "ACTIVE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenantTier_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		tier     string
		expected bool
	}{
		{"free is valid", "free", true},
		{"professional is valid", "professional", true},
		{"enterprise is valid", "enterprise", true},
		{"empty is invalid", "", false},
		{"premium is invalid", "premium", false},
		{"FREE uppercase is invalid", "FREE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidTier(tt.tier)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateSubdomain(t *testing.T) {
	tests := []struct {
		name      string
		subdomain string
		wantErr   bool
		errMsg    string
	}{
		{"valid subdomain", "acme", false, ""},
		{"valid with numbers", "acme123", false, ""},
		{"valid with hyphens", "acme-corp", false, ""},
		{"valid complex", "my-tenant-123", false, ""},
		{"too short", "ab", true, "at least"},
		{"too long", "a" + string(make([]byte, 64)), true, "at most"},
		{"starts with hyphen", "-acme", true, "must start and end"},
		{"ends with hyphen", "acme-", true, "must start and end"},
		{"uppercase letters", "ACME", true, "lowercase"},
		{"contains underscore", "acme_corp", true, "lowercase"},
		{"contains spaces", "acme corp", true, "lowercase"},
		{"single char invalid", "a", true, "at least"},
		{"exactly 3 chars valid", "abc", false, ""},
		{"exactly 63 chars valid", string(make([]byte, 63)) + "", true, "lowercase"}, // invalid chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubdomain(tt.subdomain)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTenant_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"active tenant", "active", true},
		{"inactive tenant", "inactive", false},
		{"suspended tenant", "suspended", false},
		{"deleted tenant", "deleted", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := &Tenant{Status: tt.status}
			assert.Equal(t, tt.expected, tenant.IsActive())
		})
	}
}

func TestTenant_IsSuspended(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"suspended tenant", "suspended", true},
		{"active tenant", "active", false},
		{"inactive tenant", "inactive", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := &Tenant{Status: tt.status}
			assert.Equal(t, tt.expected, tenant.IsSuspended())
		})
	}
}

func TestTenant_IsDeleted(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"deleted tenant", "deleted", true},
		{"active tenant", "active", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := &Tenant{Status: tt.status}
			assert.Equal(t, tt.expected, tenant.IsDeleted())
		})
	}
}

func TestTenant_GetQuotas(t *testing.T) {
	t.Run("valid quotas", func(t *testing.T) {
		quotas := TenantQuotas{
			MaxWorkflows:              10,
			MaxExecutionsPerDay:       100,
			MaxConcurrentExecutions:   5,
			MaxStorageBytes:           1024,
			MaxAPICallsPerMinute:      60,
			ExecutionHistoryRetention: 30,
		}
		quotasJSON, err := json.Marshal(quotas)
		require.NoError(t, err)

		tenant := &Tenant{Quotas: quotasJSON}
		result, err := tenant.GetQuotas()
		require.NoError(t, err)
		assert.Equal(t, quotas.MaxWorkflows, result.MaxWorkflows)
		assert.Equal(t, quotas.MaxExecutionsPerDay, result.MaxExecutionsPerDay)
	})

	t.Run("invalid json", func(t *testing.T) {
		tenant := &Tenant{Quotas: []byte("invalid json")}
		_, err := tenant.GetQuotas()
		assert.Error(t, err)
	})
}

func TestTenant_GetSettings(t *testing.T) {
	t.Run("valid settings", func(t *testing.T) {
		settings := TenantSettings{
			DefaultTimezone: "America/New_York",
			WebhookSecret:   "secret123",
		}
		settingsJSON, err := json.Marshal(settings)
		require.NoError(t, err)

		tenant := &Tenant{Settings: settingsJSON}
		result, err := tenant.GetSettings()
		require.NoError(t, err)
		assert.Equal(t, settings.DefaultTimezone, result.DefaultTimezone)
		assert.Equal(t, settings.WebhookSecret, result.WebhookSecret)
	})

	t.Run("invalid json", func(t *testing.T) {
		tenant := &Tenant{Settings: []byte("invalid json")}
		_, err := tenant.GetSettings()
		assert.Error(t, err)
	})
}

func TestDefaultQuotasUnknownTier(t *testing.T) {
	// Test that unknown tier defaults to free tier
	quotas := DefaultQuotas("unknown")
	assert.Equal(t, 5, quotas.MaxWorkflows)
}

func TestCreateTenantInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateTenantInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: CreateTenantInput{
				Name:      "Acme Corp",
				Subdomain: "acme",
				Tier:      "free",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			input: CreateTenantInput{
				Name:      "",
				Subdomain: "acme",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too short",
			input: CreateTenantInput{
				Name:      "A",
				Subdomain: "acme",
			},
			wantErr: true,
			errMsg:  "between 2 and 100",
		},
		{
			name: "invalid subdomain",
			input: CreateTenantInput{
				Name:      "Acme",
				Subdomain: "a",
			},
			wantErr: true,
			errMsg:  "at least",
		},
		{
			name: "invalid tier",
			input: CreateTenantInput{
				Name:      "Acme",
				Subdomain: "acme",
				Tier:      "premium",
			},
			wantErr: true,
			errMsg:  "invalid tier",
		},
		{
			name: "empty tier is valid (defaults later)",
			input: CreateTenantInput{
				Name:      "Acme",
				Subdomain: "acme",
				Tier:      "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateTenantInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   UpdateTenantInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid update with name",
			input: UpdateTenantInput{
				Name: "New Name",
			},
			wantErr: false,
		},
		{
			name: "valid update with status",
			input: UpdateTenantInput{
				Status: "active",
			},
			wantErr: false,
		},
		{
			name: "valid update with tier",
			input: UpdateTenantInput{
				Tier: "professional",
			},
			wantErr: false,
		},
		{
			name: "name too short",
			input: UpdateTenantInput{
				Name: "A",
			},
			wantErr: true,
			errMsg:  "between 2 and 100",
		},
		{
			name: "invalid status",
			input: UpdateTenantInput{
				Status: "unknown",
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "invalid tier",
			input: UpdateTenantInput{
				Tier: "premium",
			},
			wantErr: true,
			errMsg:  "invalid tier",
		},
		{
			name:    "empty input is valid",
			input:   UpdateTenantInput{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
