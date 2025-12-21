package middleware

import (
	"testing"
)

func TestVerifyTenantAccess(t *testing.T) {
	tests := []struct {
		name           string
		user           *User
		tenantID       string
		expectedResult bool
		description    string
	}{
		{
			name: "user's own tenant - access granted",
			user: &User{
				ID:       "user-1",
				Email:    "user@example.com",
				TenantID: "tenant-1",
			},
			tenantID:       "tenant-1",
			expectedResult: true,
			description:    "user should have access to their own tenant",
		},
		{
			name: "different tenant - access denied",
			user: &User{
				ID:       "user-1",
				Email:    "user@example.com",
				TenantID: "tenant-1",
			},
			tenantID:       "tenant-2",
			expectedResult: false,
			description:    "user should not have access to different tenant",
		},
		{
			name: "admin user with different tenant - access granted",
			user: &User{
				ID:       "admin-1",
				Email:    "admin@example.com",
				TenantID: "tenant-1",
				Traits: map[string]interface{}{
					"role": "admin",
				},
			},
			tenantID:       "tenant-2",
			expectedResult: true,
			description:    "admin should have access to any tenant",
		},
		{
			name: "user with no tenant ID - access denied",
			user: &User{
				ID:       "user-1",
				Email:    "user@example.com",
				TenantID: "",
			},
			tenantID:       "tenant-1",
			expectedResult: false,
			description:    "user with no tenant should not have access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyTenantAccess(tt.user, tt.tenantID)
			if result != tt.expectedResult {
				t.Errorf("%s: expected %v, got %v", tt.description, tt.expectedResult, result)
			}
		})
	}
}

// Note: Full middleware integration tests would require a concrete tenant.Service instance
// For now, we test the core VerifyTenantAccess logic which is the key functionality
