package schedule

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound = errors.New("schedule not found")
)

// Repository handles schedule database operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new schedule repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new schedule
func (r *Repository) Create(ctx context.Context, tenantID, workflowID, createdBy string, input CreateScheduleInput) (*Schedule, error) {
	id := uuid.New().String()
	now := time.Now()

	// Default timezone to UTC if not provided
	timezone := input.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	query := `
		INSERT INTO schedules (id, tenant_id, workflow_id, name, cron_expression, timezone, enabled, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING *
	`

	var schedule Schedule
	err := r.db.QueryRowxContext(
		ctx, query,
		id, tenantID, workflowID, input.Name, input.CronExpression, timezone, input.Enabled, createdBy, now, now,
	).StructScan(&schedule)

	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

// GetByID retrieves a schedule by ID (tenant-scoped)
func (r *Repository) GetByID(ctx context.Context, tenantID, id string) (*Schedule, error) {
	query := `SELECT * FROM schedules WHERE id = $1 AND tenant_id = $2`

	var schedule Schedule
	err := r.db.GetContext(ctx, &schedule, query, id, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &schedule, nil
}

// GetByIDWithoutTenant retrieves a schedule by ID without tenant check (for scheduler)
func (r *Repository) GetByIDWithoutTenant(ctx context.Context, id string) (*Schedule, error) {
	query := `SELECT * FROM schedules WHERE id = $1`

	var schedule Schedule
	err := r.db.GetContext(ctx, &schedule, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &schedule, nil
}

// Update updates a schedule
func (r *Repository) Update(ctx context.Context, tenantID, id string, input UpdateScheduleInput) (*Schedule, error) {
	query := `
		UPDATE schedules
		SET name = COALESCE($3, name),
		    cron_expression = COALESCE($4, cron_expression),
		    timezone = COALESCE($5, timezone),
		    enabled = COALESCE($6, enabled),
		    updated_at = $7
		WHERE id = $1 AND tenant_id = $2
		RETURNING *
	`

	var schedule Schedule
	err := r.db.QueryRowxContext(
		ctx, query,
		id, tenantID, input.Name, input.CronExpression, input.Timezone, input.Enabled, time.Now(),
	).StructScan(&schedule)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &schedule, nil
}

// Delete deletes a schedule
func (r *Repository) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM schedules WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, tenantID)
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

// List retrieves all schedules for a workflow
func (r *Repository) List(ctx context.Context, tenantID, workflowID string, limit, offset int) ([]*Schedule, error) {
	query := `
		SELECT * FROM schedules
		WHERE tenant_id = $1 AND workflow_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	var schedules []*Schedule
	err := r.db.SelectContext(ctx, &schedules, query, tenantID, workflowID, limit, offset)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

// ListAll retrieves all schedules for a tenant
func (r *Repository) ListAll(ctx context.Context, tenantID string, limit, offset int) ([]*ScheduleWithWorkflow, error) {
	query := `
		SELECT
			s.*,
			w.name as workflow_name,
			w.status as workflow_status
		FROM schedules s
		JOIN workflows w ON s.workflow_id = w.id
		WHERE s.tenant_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var schedules []*ScheduleWithWorkflow
	err := r.db.SelectContext(ctx, &schedules, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

// GetDueSchedules retrieves schedules that are due for execution
func (r *Repository) GetDueSchedules(ctx context.Context, beforeTime time.Time) ([]*Schedule, error) {
	query := `
		SELECT * FROM schedules
		WHERE enabled = true
		AND (next_run_at IS NULL OR next_run_at <= $1)
		ORDER BY next_run_at ASC NULLS FIRST
		LIMIT 100
	`

	var schedules []*Schedule
	err := r.db.SelectContext(ctx, &schedules, query, beforeTime)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

// UpdateNextRunTime updates the next run time for a schedule
func (r *Repository) UpdateNextRunTime(ctx context.Context, id string, nextRunAt time.Time) error {
	query := `
		UPDATE schedules
		SET next_run_at = $2,
		    updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, nextRunAt, time.Now())
	return err
}

// UpdateLastRun updates the last run information for a schedule
func (r *Repository) UpdateLastRun(ctx context.Context, id string, lastRunAt time.Time, executionID string, nextRunAt time.Time) error {
	query := `
		UPDATE schedules
		SET last_run_at = $2,
		    last_execution_id = $3,
		    next_run_at = $4,
		    updated_at = $5
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, lastRunAt, executionID, nextRunAt, time.Now())
	return err
}

// Count returns the total number of schedules for a workflow
func (r *Repository) Count(ctx context.Context, tenantID, workflowID string) (int, error) {
	query := `SELECT COUNT(*) FROM schedules WHERE tenant_id = $1 AND workflow_id = $2`

	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID, workflowID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CountAll returns the total number of schedules for a tenant
func (r *Repository) CountAll(ctx context.Context, tenantID string) (int, error) {
	query := `SELECT COUNT(*) FROM schedules WHERE tenant_id = $1`

	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID)
	if err != nil {
		return 0, err
	}

	return count, nil
}
