package retention

import (
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when a retention policy is not found
	ErrNotFound = errors.New("retention policy not found")
)

// RetentionPolicy represents the retention policy for a tenant
type RetentionPolicy struct {
	TenantID      string `db:"tenant_id" json:"tenant_id"`
	RetentionDays int    `db:"retention_days" json:"retention_days"`
	Enabled       bool   `db:"retention_enabled" json:"retention_enabled"`
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	ExecutionsDeleted     int `json:"executions_deleted"`
	StepExecutionsDeleted int `json:"step_executions_deleted"`
	BatchesProcessed      int `json:"batches_processed"`
}

// CleanupLog represents an audit log entry for retention cleanup
type CleanupLog struct {
	ID                    string    `db:"id" json:"id"`
	TenantID              string    `db:"tenant_id" json:"tenant_id"`
	ExecutionsDeleted     int       `db:"executions_deleted" json:"executions_deleted"`
	StepExecutionsDeleted int       `db:"step_executions_deleted" json:"step_executions_deleted"`
	RetentionDays         int       `db:"retention_days" json:"retention_days"`
	CutoffDate            time.Time `db:"cutoff_date" json:"cutoff_date"`
	DurationMs            int       `db:"duration_ms" json:"duration_ms"`
	Status                string    `db:"status" json:"status"` // "completed" or "failed"
	ErrorMessage          *string   `db:"error_message" json:"error_message,omitempty"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
}

// Config holds configuration for the retention service
type Config struct {
	DefaultRetentionDays int
	BatchSize            int
	EnableAuditLog       bool
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		DefaultRetentionDays: 90,
		BatchSize:            1000,
		EnableAuditLog:       true,
	}
}
