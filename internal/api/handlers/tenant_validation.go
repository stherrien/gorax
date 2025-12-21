package handlers

import (
	"fmt"
	"regexp"

	"github.com/gorax/gorax/internal/tenant"
)

var (
	// subdomainRegex validates subdomain format (alphanumeric and hyphens, no underscores)
	subdomainRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

	// validStatuses defines valid tenant status values
	validStatuses = map[string]bool{
		"active":    true,
		"suspended": true,
		"inactive":  true,
	}
)

// ValidateCreateTenantInput validates tenant creation input
func ValidateCreateTenantInput(input tenant.CreateTenantInput) error {
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(input.Name) > 100 {
		return fmt.Errorf("name must be less than 100 characters")
	}

	if input.Subdomain == "" {
		return fmt.Errorf("subdomain is required")
	}

	if len(input.Subdomain) < 3 || len(input.Subdomain) > 63 {
		return fmt.Errorf("subdomain must be between 3 and 63 characters")
	}

	if !subdomainRegex.MatchString(input.Subdomain) {
		return fmt.Errorf("subdomain must contain only lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen")
	}

	// Validate tier if provided
	if input.Tier != "" {
		validTiers := map[string]bool{
			"free":         true,
			"professional": true,
			"enterprise":   true,
		}
		if !validTiers[input.Tier] {
			return fmt.Errorf("tier must be one of: free, professional, enterprise")
		}
	}

	return nil
}

// ValidateUpdateTenantInput validates tenant update input
func ValidateUpdateTenantInput(input tenant.UpdateTenantInput) error {
	// Name validation (if provided)
	if input.Name != "" {
		if len(input.Name) > 100 {
			return fmt.Errorf("name must be less than 100 characters")
		}
	}

	// Status validation (if provided)
	if input.Status != "" {
		if !validStatuses[input.Status] {
			return fmt.Errorf("status must be one of: active, suspended, inactive")
		}
	}

	// Tier validation (if provided)
	if input.Tier != "" {
		validTiers := map[string]bool{
			"free":         true,
			"professional": true,
			"enterprise":   true,
		}
		if !validTiers[input.Tier] {
			return fmt.Errorf("tier must be one of: free, professional, enterprise")
		}
	}

	return nil
}
