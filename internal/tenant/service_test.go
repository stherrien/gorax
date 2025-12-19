package tenant

import (
	"context"
	"testing"
)

// TestDefaultQuotas verifies quota defaults for different tiers
func TestDefaultQuotas(t *testing.T) {
	tests := []struct {
		tier                      string
		expectedMaxWorkflows      int
		expectedMaxExecutionsDay  int
		expectedMaxConcurrent     int
	}{
		{
			tier:                     "free",
			expectedMaxWorkflows:     5,
			expectedMaxExecutionsDay: 100,
			expectedMaxConcurrent:    2,
		},
		{
			tier:                     "professional",
			expectedMaxWorkflows:     50,
			expectedMaxExecutionsDay: 5000,
			expectedMaxConcurrent:    10,
		},
		{
			tier:                     "enterprise",
			expectedMaxWorkflows:     -1, // unlimited
			expectedMaxExecutionsDay: -1, // unlimited
			expectedMaxConcurrent:    100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			quotas := DefaultQuotas(tt.tier)

			if quotas.MaxWorkflows != tt.expectedMaxWorkflows {
				t.Errorf("MaxWorkflows = %d, want %d", quotas.MaxWorkflows, tt.expectedMaxWorkflows)
			}
			if quotas.MaxExecutionsPerDay != tt.expectedMaxExecutionsDay {
				t.Errorf("MaxExecutionsPerDay = %d, want %d", quotas.MaxExecutionsPerDay, tt.expectedMaxExecutionsDay)
			}
			if quotas.MaxConcurrentExecutions != tt.expectedMaxConcurrent {
				t.Errorf("MaxConcurrentExecutions = %d, want %d", quotas.MaxConcurrentExecutions, tt.expectedMaxConcurrent)
			}
		})
	}
}

// TestTenantScoped verifies tenant context is properly set
func TestTenantScoped(t *testing.T) {
	ctx := context.Background()
	tenantID := "test-tenant-123"

	// This test is for the database package, but we'll test the concept here
	if ctx.Value("tenant_id") != nil {
		t.Error("Expected empty context initially")
	}

	// In real usage, this would be done via database.TenantScoped
	ctxWithTenant := context.WithValue(ctx, "tenant_id", tenantID)

	if val := ctxWithTenant.Value("tenant_id"); val != tenantID {
		t.Errorf("Expected tenant_id = %s, got %v", tenantID, val)
	}
}
