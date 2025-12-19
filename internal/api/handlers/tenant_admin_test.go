package handlers

import (
	"testing"

	"github.com/gorax/gorax/internal/tenant"
)

// TestTenantQuotasIntegration validates quota structures
func TestTenantQuotasIntegration(t *testing.T) {
	tests := []struct {
		name string
		tier string
		want tenant.TenantQuotas
	}{
		{
			name: "Free tier quotas",
			tier: "free",
			want: tenant.TenantQuotas{
				MaxWorkflows:              5,
				MaxExecutionsPerDay:       100,
				MaxConcurrentExecutions:   2,
				MaxStorageBytes:           100 * 1024 * 1024,
				MaxAPICallsPerMinute:      60,
				ExecutionHistoryRetention: 7,
			},
		},
		{
			name: "Professional tier quotas",
			tier: "professional",
			want: tenant.TenantQuotas{
				MaxWorkflows:              50,
				MaxExecutionsPerDay:       5000,
				MaxConcurrentExecutions:   10,
				MaxStorageBytes:           5 * 1024 * 1024 * 1024,
				MaxAPICallsPerMinute:      300,
				ExecutionHistoryRetention: 30,
			},
		},
		{
			name: "Enterprise tier quotas",
			tier: "enterprise",
			want: tenant.TenantQuotas{
				MaxWorkflows:              -1,
				MaxExecutionsPerDay:       -1,
				MaxConcurrentExecutions:   100,
				MaxStorageBytes:           -1,
				MaxAPICallsPerMinute:      1000,
				ExecutionHistoryRetention: 365,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tenant.DefaultQuotas(tt.tier)

			if got.MaxWorkflows != tt.want.MaxWorkflows {
				t.Errorf("MaxWorkflows = %d, want %d", got.MaxWorkflows, tt.want.MaxWorkflows)
			}
			if got.MaxExecutionsPerDay != tt.want.MaxExecutionsPerDay {
				t.Errorf("MaxExecutionsPerDay = %d, want %d", got.MaxExecutionsPerDay, tt.want.MaxExecutionsPerDay)
			}
			if got.MaxConcurrentExecutions != tt.want.MaxConcurrentExecutions {
				t.Errorf("MaxConcurrentExecutions = %d, want %d", got.MaxConcurrentExecutions, tt.want.MaxConcurrentExecutions)
			}
		})
	}
}

// TestCalculatePercentage validates the percentage calculation helper
func TestCalculatePercentage(t *testing.T) {
	tests := []struct {
		name    string
		current int
		max     int
		want    float64
	}{
		{
			name:    "50% usage",
			current: 50,
			max:     100,
			want:    50.0,
		},
		{
			name:    "Over quota (should cap at 100)",
			current: 150,
			max:     100,
			want:    100.0,
		},
		{
			name:    "Unlimited quota (-1)",
			current: 1000,
			max:     -1,
			want:    0.0,
		},
		{
			name:    "Zero max",
			current: 10,
			max:     0,
			want:    0.0,
		},
		{
			name:    "Zero usage",
			current: 0,
			max:     100,
			want:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculatePercentage(tt.current, tt.max)
			if got != tt.want {
				t.Errorf("calculatePercentage(%d, %d) = %f, want %f", tt.current, tt.max, got, tt.want)
			}
		})
	}
}
