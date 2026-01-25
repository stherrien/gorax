package retention

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// PostgresRepository implements the Repository interface for PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewRepository creates a new PostgreSQL repository
func NewRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetRetentionPolicy retrieves the retention policy from tenant settings
func (r *PostgresRepository) GetRetentionPolicy(ctx context.Context, tenantID string) (*RetentionPolicy, error) {
	query := `
		SELECT
			id as tenant_id,
			COALESCE(settings->>'retention_days', '90')::int as retention_days,
			COALESCE((settings->>'retention_enabled')::boolean, true) as retention_enabled
		FROM tenants
		WHERE id = $1 AND status != 'deleted'
	`

	var policy RetentionPolicy
	err := r.db.GetContext(ctx, &policy, query, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get retention policy: %w", err)
	}

	return &policy, nil
}

// DeleteOldExecutions deletes old executions in batches
// This follows the pattern: delete step_executions first, then executions (foreign key order)
func (r *PostgresRepository) DeleteOldExecutions(ctx context.Context, tenantID string, cutoffDate time.Time, batchSize int) (*CleanupResult, error) {
	result := &CleanupResult{
		ExecutionsDeleted:     0,
		StepExecutionsDeleted: 0,
		BatchesProcessed:      0,
	}

	// Process in batches to avoid long-running locks
	for {
		batchResult, err := r.deleteExecutionBatch(ctx, tenantID, cutoffDate, batchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to delete batch: %w", err)
		}

		result.ExecutionsDeleted += batchResult.ExecutionsDeleted
		result.StepExecutionsDeleted += batchResult.StepExecutionsDeleted
		result.BatchesProcessed++

		// Stop if no more records to delete
		if batchResult.ExecutionsDeleted == 0 {
			break
		}

		// Small delay between batches to reduce load
		if batchResult.ExecutionsDeleted == batchSize {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return result, nil
}

// ArchiveAndDeleteOldExecutions archives executions before deleting them
// This follows the pattern: archive execution data, delete step_executions, then delete executions
func (r *PostgresRepository) ArchiveAndDeleteOldExecutions(ctx context.Context, tenantID string, cutoffDate time.Time, batchSize int) (*CleanupResult, error) {
	result := &CleanupResult{
		ExecutionsDeleted:     0,
		StepExecutionsDeleted: 0,
		ExecutionsArchived:    0,
		BatchesProcessed:      0,
	}

	// Process in batches to avoid long-running locks
	for {
		batchResult, err := r.archiveAndDeleteExecutionBatch(ctx, tenantID, cutoffDate, batchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to archive and delete batch: %w", err)
		}

		result.ExecutionsDeleted += batchResult.ExecutionsDeleted
		result.StepExecutionsDeleted += batchResult.StepExecutionsDeleted
		result.ExecutionsArchived += batchResult.ExecutionsArchived
		result.BatchesProcessed++

		// Stop if no more records to process
		if batchResult.ExecutionsDeleted == 0 {
			break
		}

		// Small delay between batches to reduce load
		if batchResult.ExecutionsDeleted == batchSize {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return result, nil
}

// archiveAndDeleteExecutionBatch archives and deletes a single batch of executions
func (r *PostgresRepository) archiveAndDeleteExecutionBatch(ctx context.Context, tenantID string, cutoffDate time.Time, batchSize int) (*CleanupResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// Log rollback error but don't return it
		}
	}()

	// Get executions to archive and delete in this batch
	executionsQuery := `
		SELECT id, tenant_id, workflow_id, status, started_at, completed_at,
		       trigger_type, trigger_data, result, error, created_at, updated_at
		FROM executions
		WHERE tenant_id = $1
		  AND created_at < $2
		  AND status IN ('completed', 'failed')
		ORDER BY created_at ASC
		LIMIT $3
	`

	type executionRow struct {
		ID          string          `db:"id"`
		TenantID    string          `db:"tenant_id"`
		WorkflowID  string          `db:"workflow_id"`
		Status      string          `db:"status"`
		StartedAt   *time.Time      `db:"started_at"`
		CompletedAt *time.Time      `db:"completed_at"`
		TriggerType string          `db:"trigger_type"`
		TriggerData json.RawMessage `db:"trigger_data"`
		Result      json.RawMessage `db:"result"`
		Error       *string         `db:"error"`
		CreatedAt   time.Time       `db:"created_at"`
		UpdatedAt   time.Time       `db:"updated_at"`
	}

	var executions []executionRow
	err = tx.SelectContext(ctx, &executions, executionsQuery, tenantID, cutoffDate, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	// If no executions to process, return
	if len(executions) == 0 {
		return &CleanupResult{
			ExecutionsDeleted:     0,
			StepExecutionsDeleted: 0,
			ExecutionsArchived:    0,
			BatchesProcessed:      0,
		}, nil
	}

	// Collect execution IDs for later operations
	executionIDs := make([]string, len(executions))
	for i, exec := range executions {
		executionIDs[i] = exec.ID
	}

	// Archive each execution
	archiveQuery := `
		INSERT INTO execution_archives (id, tenant_id, workflow_id, execution_data, archived_at, original_created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`

	archivedCount := 0
	for _, exec := range executions {
		// Build execution data JSON including step executions
		executionData := map[string]interface{}{
			"id":           exec.ID,
			"tenant_id":    exec.TenantID,
			"workflow_id":  exec.WorkflowID,
			"status":       exec.Status,
			"started_at":   exec.StartedAt,
			"completed_at": exec.CompletedAt,
			"trigger_type": exec.TriggerType,
			"trigger_data": exec.TriggerData,
			"result":       exec.Result,
			"error":        exec.Error,
			"created_at":   exec.CreatedAt,
			"updated_at":   exec.UpdatedAt,
		}

		// Get step executions for this execution
		stepsQuery := `
			SELECT id, execution_id, node_id, status, started_at, completed_at,
			       input, output, error, created_at
			FROM step_executions
			WHERE execution_id = $1
			ORDER BY created_at ASC
		`
		var steps []map[string]interface{}
		rows, err := tx.QueryxContext(ctx, stepsQuery, exec.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get step executions for %s: %w", exec.ID, err)
		}
		for rows.Next() {
			step := make(map[string]interface{})
			if err := rows.MapScan(step); err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to scan step execution: %w", err)
			}
			steps = append(steps, step)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating step executions: %w", err)
		}

		executionData["step_executions"] = steps

		// Marshal to JSON
		dataJSON, err := json.Marshal(executionData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal execution data: %w", err)
		}

		// Insert archive
		result, err := tx.ExecContext(ctx, archiveQuery, exec.ID, exec.TenantID, exec.WorkflowID, dataJSON, time.Now(), exec.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to archive execution %s: %w", exec.ID, err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			archivedCount++
		}
	}

	// Delete step_executions first (foreign key constraint)
	stepDeleteQuery := `
		DELETE FROM step_executions
		WHERE execution_id = ANY($1)
	`

	stepResult, err := tx.ExecContext(ctx, stepDeleteQuery, executionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to delete step executions: %w", err)
	}

	stepRowsDeleted, err := stepResult.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get step rows deleted: %w", err)
	}

	// Delete executions
	execDeleteQuery := `
		DELETE FROM executions
		WHERE id = ANY($1)
	`

	execResult, err := tx.ExecContext(ctx, execDeleteQuery, executionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to delete executions: %w", err)
	}

	execRowsDeleted, err := execResult.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get execution rows deleted: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &CleanupResult{
		ExecutionsDeleted:     int(execRowsDeleted),
		StepExecutionsDeleted: int(stepRowsDeleted),
		ExecutionsArchived:    archivedCount,
		BatchesProcessed:      0, // Will be incremented by caller
	}, nil
}

// deleteExecutionBatch deletes a single batch of executions
func (r *PostgresRepository) deleteExecutionBatch(ctx context.Context, tenantID string, cutoffDate time.Time, batchSize int) (*CleanupResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get execution IDs to delete in this batch
	executionIDsQuery := `
		SELECT id
		FROM executions
		WHERE tenant_id = $1
		  AND created_at < $2
		  AND status IN ('completed', 'failed') -- Only delete finished executions
		ORDER BY created_at ASC
		LIMIT $3
	`

	var executionIDs []string
	err = tx.SelectContext(ctx, &executionIDs, executionIDsQuery, tenantID, cutoffDate, batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution IDs: %w", err)
	}

	// If no executions to delete, return
	if len(executionIDs) == 0 {
		return &CleanupResult{
			ExecutionsDeleted:     0,
			StepExecutionsDeleted: 0,
			BatchesProcessed:      0,
		}, nil
	}

	// Delete step_executions first (foreign key constraint)
	stepDeleteQuery := `
		DELETE FROM step_executions
		WHERE execution_id = ANY($1)
	`

	stepResult, err := tx.ExecContext(ctx, stepDeleteQuery, executionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to delete step executions: %w", err)
	}

	stepRowsDeleted, err := stepResult.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get step rows deleted: %w", err)
	}

	// Delete executions
	execDeleteQuery := `
		DELETE FROM executions
		WHERE id = ANY($1)
	`

	execResult, err := tx.ExecContext(ctx, execDeleteQuery, executionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to delete executions: %w", err)
	}

	execRowsDeleted, err := execResult.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get execution rows deleted: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &CleanupResult{
		ExecutionsDeleted:     int(execRowsDeleted),
		StepExecutionsDeleted: int(stepRowsDeleted),
		BatchesProcessed:      0, // Will be incremented by caller
	}, nil
}

// GetTenantsWithRetention returns all active tenant IDs
func (r *PostgresRepository) GetTenantsWithRetention(ctx context.Context) ([]string, error) {
	query := `
		SELECT id
		FROM tenants
		WHERE status = 'active'
		ORDER BY id
	`

	var tenantIDs []string
	err := r.db.SelectContext(ctx, &tenantIDs, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants: %w", err)
	}

	return tenantIDs, nil
}

// LogCleanup logs a cleanup operation for audit purposes
func (r *PostgresRepository) LogCleanup(ctx context.Context, log *CleanupLog) error {
	query := `
		INSERT INTO retention_cleanup_logs (
			id, tenant_id, executions_deleted, executions_archived, step_executions_deleted,
			retention_days, cutoff_date, duration_ms, status, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		log.ID, log.TenantID, log.ExecutionsDeleted, log.ExecutionsArchived, log.StepExecutionsDeleted,
		log.RetentionDays, log.CutoffDate, log.DurationMs, log.Status, log.ErrorMessage, log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to log cleanup: %w", err)
	}

	return nil
}

// SetRetentionPolicy updates the retention policy in tenant settings
func (r *PostgresRepository) SetRetentionPolicy(ctx context.Context, tenantID string, retentionDays int, enabled bool) error {
	// Get current settings
	var settingsJSON []byte
	err := r.db.GetContext(ctx, &settingsJSON, "SELECT settings FROM tenants WHERE id = $1", tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to get tenant settings: %w", err)
	}

	// Parse existing settings
	var settings map[string]interface{}
	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &settings); err != nil {
			return fmt.Errorf("failed to parse settings: %w", err)
		}
	} else {
		settings = make(map[string]interface{})
	}

	// Update retention settings
	settings["retention_days"] = retentionDays
	settings["retention_enabled"] = enabled

	// Marshal back to JSON
	updatedJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Update tenant
	query := `
		UPDATE tenants
		SET settings = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, updatedJSON)
	if err != nil {
		return fmt.Errorf("failed to update tenant settings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
