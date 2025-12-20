package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorax/gorax/internal/tenant"
)

func TestValidateCreateTenantInput(t *testing.T) {
	tests := []struct {
		name        string
		input       tenant.CreateTenantInput
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid input",
			input: tenant.CreateTenantInput{
				Name:      "Test Tenant",
				Subdomain: "test-tenant",
				Tier:      "free",
			},
			expectError: false,
		},
		{
			name: "missing name",
			input: tenant.CreateTenantInput{
				Name:      "",
				Subdomain: "test-tenant",
				Tier:      "free",
			},
			expectError: true,
			errorMsg:    "name",
		},
		{
			name: "missing subdomain",
			input: tenant.CreateTenantInput{
				Name:      "Test Tenant",
				Subdomain: "",
				Tier:      "free",
			},
			expectError: true,
			errorMsg:    "subdomain",
		},
		{
			name: "invalid subdomain format",
			input: tenant.CreateTenantInput{
				Name:      "Test Tenant",
				Subdomain: "Test_Tenant!",
				Tier:      "free",
			},
			expectError: true,
			errorMsg:    "subdomain",
		},
		{
			name: "subdomain too short",
			input: tenant.CreateTenantInput{
				Name:      "Test Tenant",
				Subdomain: "ab",
				Tier:      "free",
			},
			expectError: true,
			errorMsg:    "subdomain",
		},
		{
			name: "subdomain too long",
			input: tenant.CreateTenantInput{
				Name:      "Test Tenant",
				Subdomain: "this-is-a-very-long-subdomain-name-that-exceeds-the-maximum-length",
				Tier:      "free",
			},
			expectError: true,
			errorMsg:    "subdomain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateTenantInput(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUpdateTenantInput(t *testing.T) {
	tests := []struct {
		name        string
		input       tenant.UpdateTenantInput
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid update with name",
			input: tenant.UpdateTenantInput{
				Name: "Updated Tenant",
			},
			expectError: false,
		},
		{
			name: "valid update with status",
			input: tenant.UpdateTenantInput{
				Status: "active",
			},
			expectError: false,
		},
		{
			name: "invalid status",
			input: tenant.UpdateTenantInput{
				Status: "invalid-status",
			},
			expectError: true,
			errorMsg:    "status",
		},
		{
			name: "name too long",
			input: tenant.UpdateTenantInput{
				Name: string(make([]byte, 101)), // 101 characters
			},
			expectError: true,
			errorMsg:    "name",
		},
		{
			name: "empty input is valid",
			input: tenant.UpdateTenantInput{
				Name:   "",
				Status: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUpdateTenantInput(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
