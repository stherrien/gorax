package tenant

import (
	"encoding/json"
	"time"
)

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
