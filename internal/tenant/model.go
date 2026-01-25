package tenant

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	// StatusActive indicates the tenant is active and operational
	StatusActive TenantStatus = "active"
	// StatusInactive indicates the tenant is temporarily disabled
	StatusInactive TenantStatus = "inactive"
	// StatusSuspended indicates the tenant is suspended (e.g., billing issues)
	StatusSuspended TenantStatus = "suspended"
	// StatusDeleted indicates the tenant has been soft-deleted
	StatusDeleted TenantStatus = "deleted"
)

// TenantTier represents the pricing tier of a tenant
type TenantTier string

const (
	// TierFree is the free tier with limited features
	TierFree TenantTier = "free"
	// TierProfessional is the professional tier with extended features
	TierProfessional TenantTier = "professional"
	// TierEnterprise is the enterprise tier with unlimited features
	TierEnterprise TenantTier = "enterprise"
)

// IsValidStatus checks if the given status is valid
func IsValidStatus(status string) bool {
	switch TenantStatus(status) {
	case StatusActive, StatusInactive, StatusSuspended, StatusDeleted:
		return true
	default:
		return false
	}
}

// IsValidTier checks if the given tier is valid
func IsValidTier(tier string) bool {
	switch TenantTier(tier) {
	case TierFree, TierProfessional, TierEnterprise:
		return true
	default:
		return false
	}
}

// Tenant represents a tenant in the system
type Tenant struct {
	ID        string          `db:"id" json:"id"`
	Name      string          `db:"name" json:"name"`
	Subdomain string          `db:"subdomain" json:"subdomain"`
	Status    string          `db:"status" json:"status"`
	Tier      string          `db:"tier" json:"tier"`
	Settings  json.RawMessage `db:"settings" json:"settings"`
	Quotas    json.RawMessage `db:"quotas" json:"quotas"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}

// IsActive returns true if the tenant status is active
func (t *Tenant) IsActive() bool {
	return t.Status == string(StatusActive)
}

// IsSuspended returns true if the tenant status is suspended
func (t *Tenant) IsSuspended() bool {
	return t.Status == string(StatusSuspended)
}

// IsDeleted returns true if the tenant status is deleted
func (t *Tenant) IsDeleted() bool {
	return t.Status == string(StatusDeleted)
}

// GetQuotas parses and returns the tenant quotas
func (t *Tenant) GetQuotas() (*TenantQuotas, error) {
	var quotas TenantQuotas
	if err := json.Unmarshal(t.Quotas, &quotas); err != nil {
		return nil, fmt.Errorf("failed to parse tenant quotas: %w", err)
	}
	return &quotas, nil
}

// GetSettings parses and returns the tenant settings
func (t *Tenant) GetSettings() (*TenantSettings, error) {
	var settings TenantSettings
	if err := json.Unmarshal(t.Settings, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse tenant settings: %w", err)
	}
	return &settings, nil
}

// TenantSettings holds tenant-specific settings
type TenantSettings struct {
	DefaultTimezone string `json:"default_timezone"`
	WebhookSecret   string `json:"webhook_secret"`
}

// TenantQuotas holds tenant resource quotas
type TenantQuotas struct {
	MaxWorkflows              int `json:"max_workflows"`
	MaxExecutionsPerDay       int `json:"max_executions_per_day"`
	MaxConcurrentExecutions   int `json:"max_concurrent_executions"`
	MaxStorageBytes           int `json:"max_storage_bytes"`
	MaxAPICallsPerMinute      int `json:"max_api_calls_per_minute"`
	ExecutionHistoryRetention int `json:"execution_history_retention_days"`
}

// DefaultQuotas returns default quotas based on tier
func DefaultQuotas(tier string) TenantQuotas {
	switch tier {
	case "enterprise":
		return TenantQuotas{
			MaxWorkflows:              -1, // unlimited
			MaxExecutionsPerDay:       -1,
			MaxConcurrentExecutions:   100,
			MaxStorageBytes:           -1,
			MaxAPICallsPerMinute:      1000,
			ExecutionHistoryRetention: 365,
		}
	case "professional":
		return TenantQuotas{
			MaxWorkflows:              50,
			MaxExecutionsPerDay:       5000,
			MaxConcurrentExecutions:   10,
			MaxStorageBytes:           5 * 1024 * 1024 * 1024, // 5GB
			MaxAPICallsPerMinute:      300,
			ExecutionHistoryRetention: 30,
		}
	default: // free
		return TenantQuotas{
			MaxWorkflows:              5,
			MaxExecutionsPerDay:       100,
			MaxConcurrentExecutions:   2,
			MaxStorageBytes:           100 * 1024 * 1024, // 100MB
			MaxAPICallsPerMinute:      60,
			ExecutionHistoryRetention: 7,
		}
	}
}

// CreateTenantInput represents input for creating a tenant
type CreateTenantInput struct {
	Name      string `json:"name" validate:"required,min=2,max=100"`
	Subdomain string `json:"subdomain" validate:"required,min=3,max=63,alphanum"`
	Tier      string `json:"tier" validate:"oneof=free professional enterprise"`
}

// UpdateTenantInput represents input for updating a tenant
type UpdateTenantInput struct {
	Name     string          `json:"name,omitempty"`
	Status   string          `json:"status,omitempty"`
	Tier     string          `json:"tier,omitempty"`
	Settings json.RawMessage `json:"settings,omitempty"`
}

// UsageStats represents current usage statistics for a tenant
type UsageStats struct {
	TenantID             string `json:"tenant_id"`
	WorkflowCount        int    `json:"workflow_count"`
	ExecutionsToday      int    `json:"executions_today"`
	ExecutionsThisMonth  int    `json:"executions_this_month"`
	ConcurrentExecutions int    `json:"concurrent_executions"`
	StorageBytes         int64  `json:"storage_bytes"`
}

// Subdomain validation constants
const (
	// MinSubdomainLength is the minimum length for a subdomain
	MinSubdomainLength = 3
	// MaxSubdomainLength is the maximum length for a subdomain
	MaxSubdomainLength = 63
	// DefaultTenantSubdomain is the subdomain for the default tenant in single-tenant mode
	DefaultTenantSubdomain = "default"
	// DefaultTenantName is the name for the default tenant
	DefaultTenantName = "Default Tenant"
)

// subdomainRegex validates subdomain format: lowercase alphanumeric and hyphens, must start/end with alphanumeric
var subdomainRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ValidateSubdomain validates a subdomain format
func ValidateSubdomain(subdomain string) error {
	if len(subdomain) < MinSubdomainLength {
		return fmt.Errorf("subdomain must be at least %d characters", MinSubdomainLength)
	}
	if len(subdomain) > MaxSubdomainLength {
		return fmt.Errorf("subdomain must be at most %d characters", MaxSubdomainLength)
	}
	if !subdomainRegex.MatchString(subdomain) {
		return fmt.Errorf("subdomain must contain only lowercase letters, numbers, and hyphens, and must start and end with a letter or number")
	}
	return nil
}

// Validate validates the CreateTenantInput
func (c *CreateTenantInput) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.Name) < 2 || len(c.Name) > 100 {
		return fmt.Errorf("name must be between 2 and 100 characters")
	}
	if err := ValidateSubdomain(c.Subdomain); err != nil {
		return err
	}
	if c.Tier != "" && !IsValidTier(c.Tier) {
		return fmt.Errorf("invalid tier: must be one of free, professional, enterprise")
	}
	return nil
}

// Validate validates the UpdateTenantInput
func (u *UpdateTenantInput) Validate() error {
	if u.Name != "" && (len(u.Name) < 2 || len(u.Name) > 100) {
		return fmt.Errorf("name must be between 2 and 100 characters")
	}
	if u.Status != "" && !IsValidStatus(u.Status) {
		return fmt.Errorf("invalid status: must be one of active, inactive, suspended")
	}
	if u.Tier != "" && !IsValidTier(u.Tier) {
		return fmt.Errorf("invalid tier: must be one of free, professional, enterprise")
	}
	return nil
}
