package tenant

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound = errors.New("tenant not found")
)

// Repository handles tenant database operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new tenant repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new tenant
func (r *Repository) Create(ctx context.Context, input CreateTenantInput) (*Tenant, error) {
	id := uuid.New().String()
	now := time.Now()

	// Set default quotas based on tier
	quotas := DefaultQuotas(input.Tier)
	quotasJSON, err := json.Marshal(quotas)
	if err != nil {
		return nil, err
	}

	// Default settings
	settings := TenantSettings{
		DefaultTimezone: "UTC",
		WebhookSecret:   uuid.New().String(),
	}
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO tenants (id, name, subdomain, status, tier, settings, quotas, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, subdomain, status, tier, settings, quotas, created_at, updated_at
	`

	var tenant Tenant
	err = r.db.QueryRowxContext(
		ctx, query,
		id, input.Name, input.Subdomain, "active", input.Tier, settingsJSON, quotasJSON, now, now,
	).StructScan(&tenant)

	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// GetByID retrieves a tenant by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Tenant, error) {
	query := `SELECT * FROM tenants WHERE id = $1`

	var tenant Tenant
	err := r.db.GetContext(ctx, &tenant, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// GetBySubdomain retrieves a tenant by subdomain
func (r *Repository) GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	query := `SELECT * FROM tenants WHERE subdomain = $1`

	var tenant Tenant
	err := r.db.GetContext(ctx, &tenant, query, subdomain)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// Update updates a tenant
func (r *Repository) Update(ctx context.Context, id string, input UpdateTenantInput) (*Tenant, error) {
	query := `
		UPDATE tenants
		SET name = COALESCE(NULLIF($2, ''), name),
		    status = COALESCE(NULLIF($3, ''), status),
		    tier = COALESCE(NULLIF($4, ''), tier),
		    settings = COALESCE($5, settings),
		    updated_at = $6
		WHERE id = $1
		RETURNING *
	`

	var tenant Tenant
	err := r.db.QueryRowxContext(
		ctx, query,
		id, input.Name, input.Status, input.Tier, input.Settings, time.Now(),
	).StructScan(&tenant)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// Delete deletes a tenant (soft delete by setting status to 'deleted')
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `UPDATE tenants SET status = 'deleted', updated_at = $2 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// List retrieves all tenants with pagination
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*Tenant, error) {
	query := `
		SELECT * FROM tenants
		WHERE status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var tenants []*Tenant
	err := r.db.SelectContext(ctx, &tenants, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return tenants, nil
}

// Count returns the total number of active tenants
func (r *Repository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM tenants WHERE status != 'deleted'`

	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// UpdateQuotas updates tenant quotas
func (r *Repository) UpdateQuotas(ctx context.Context, id string, quotas TenantQuotas) (*Tenant, error) {
	quotasJSON, err := json.Marshal(quotas)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE tenants
		SET quotas = $2,
		    updated_at = $3
		WHERE id = $1
		RETURNING *
	`

	var tenant Tenant
	err = r.db.QueryRowxContext(ctx, query, id, quotasJSON, time.Now()).StructScan(&tenant)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// GetWorkflowCount returns the count of active workflows for a tenant
func (r *Repository) GetWorkflowCount(ctx context.Context, tenantID string) (int, error) {
	query := `SELECT COUNT(*) FROM workflows WHERE tenant_id = $1 AND status != 'archived'`

	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetExecutionStats returns execution statistics for a tenant
func (r *Repository) GetExecutionStats(ctx context.Context, tenantID string) (*UsageStats, error) {
	today := time.Now().Truncate(24 * time.Hour)
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)

	query := `
		SELECT
			$1 as tenant_id,
			(SELECT COUNT(*) FROM workflows WHERE tenant_id = $1 AND status != 'archived') as workflow_count,
			(SELECT COUNT(*) FROM executions WHERE tenant_id = $1 AND created_at >= $2) as executions_today,
			(SELECT COUNT(*) FROM executions WHERE tenant_id = $1 AND created_at >= $3) as executions_this_month,
			(SELECT COUNT(*) FROM executions WHERE tenant_id = $1 AND status = 'running') as concurrent_executions,
			0 as storage_bytes
	`

	var stats UsageStats
	err := r.db.GetContext(ctx, &stats, query, tenantID, today, monthStart)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetConcurrentExecutions returns the count of currently running executions
func (r *Repository) GetConcurrentExecutions(ctx context.Context, tenantID string) (int, error) {
	query := `SELECT COUNT(*) FROM executions WHERE tenant_id = $1 AND status = 'running'`

	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID)
	if err != nil {
		return 0, err
	}

	return count, nil
}
