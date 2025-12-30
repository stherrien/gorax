package workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound = errors.New("workflow not found")
)

// Repository handles workflow database operations
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new workflow repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// setTenantContext sets the tenant ID in the PostgreSQL session for RLS
// Note: This only works within a transaction. For non-transactional queries,
// the setting won't persist. Use set_config with is_local=false for session-level settings.
func (r *Repository) setTenantContext(ctx context.Context, tenantID string) error {
	_, err := r.db.ExecContext(ctx, "SELECT set_config('app.current_tenant_id', $1, false)", tenantID)
	return err
}

// Create inserts a new workflow
func (r *Repository) Create(ctx context.Context, tenantID, createdBy string, input CreateWorkflowInput) (*Workflow, error) {
	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO workflows (id, tenant_id, name, description, definition, status, version, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING *
	`

	var workflow Workflow
	err := r.db.QueryRowxContext(
		ctx, query,
		id, tenantID, input.Name, input.Description, input.Definition, "draft", 1, createdBy, now, now,
	).StructScan(&workflow)

	if err != nil {
		return nil, err
	}

	return &workflow, nil
}

// GetByID retrieves a workflow by ID (tenant-scoped)
func (r *Repository) GetByID(ctx context.Context, tenantID, id string) (*Workflow, error) {
	query := `SELECT * FROM workflows WHERE id = $1 AND tenant_id = $2`

	var workflow Workflow
	err := r.db.GetContext(ctx, &workflow, query, id, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &workflow, nil
}

// Update updates a workflow
func (r *Repository) Update(ctx context.Context, tenantID, id string, input UpdateWorkflowInput) (*Workflow, error) {
	// First get the current workflow to increment version if definition changed
	current, err := r.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	newVersion := current.Version
	if input.Definition != nil {
		newVersion++
	}

	query := `
		UPDATE workflows
		SET name = COALESCE(NULLIF($3, ''), name),
		    description = COALESCE(NULLIF($4, ''), description),
		    definition = COALESCE($5, definition),
		    status = COALESCE(NULLIF($6, ''), status),
		    version = $7,
		    updated_at = $8
		WHERE id = $1 AND tenant_id = $2
		RETURNING *
	`

	var workflow Workflow
	err = r.db.QueryRowxContext(
		ctx, query,
		id, tenantID, input.Name, input.Description, input.Definition, input.Status, newVersion, time.Now(),
	).StructScan(&workflow)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &workflow, nil
}

// Delete deletes a workflow (soft delete by setting status to 'archived')
func (r *Repository) Delete(ctx context.Context, tenantID, id string) error {
	query := `UPDATE workflows SET status = 'archived', updated_at = $3 WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, tenantID, time.Now())
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

// List retrieves all workflows for a tenant with pagination
func (r *Repository) List(ctx context.Context, tenantID string, limit, offset int) ([]*Workflow, error) {
	query := `
		SELECT * FROM workflows
		WHERE tenant_id = $1 AND status != 'archived'
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	var workflows []*Workflow
	err := r.db.SelectContext(ctx, &workflows, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}

	return workflows, nil
}

// Count returns the total number of workflows for a tenant
func (r *Repository) Count(ctx context.Context, tenantID string) (int, error) {
	query := `SELECT COUNT(*) FROM workflows WHERE tenant_id = $1 AND status != 'archived'`

	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CreateExecution creates a new execution record
func (r *Repository) CreateExecution(ctx context.Context, tenantID, workflowID string, workflowVersion int, triggerType string, triggerData []byte) (*Execution, error) {
	id := uuid.New().String()
	now := time.Now()

	// Handle nil or empty trigger data
	var triggerDataParam interface{}
	if len(triggerData) == 0 {
		triggerDataParam = nil
	} else {
		triggerDataParam = triggerData
	}

	query := `
		INSERT INTO executions (id, tenant_id, workflow_id, workflow_version, status, trigger_type, trigger_data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *
	`

	var execution Execution
	err := r.db.QueryRowxContext(
		ctx, query,
		id, tenantID, workflowID, workflowVersion, "pending", triggerType, triggerDataParam, now,
	).StructScan(&execution)

	if err != nil {
		return nil, err
	}

	return &execution, nil
}

// GetExecutionByID retrieves an execution by ID
func (r *Repository) GetExecutionByID(ctx context.Context, tenantID, id string) (*Execution, error) {
	query := `SELECT * FROM executions WHERE id = $1 AND tenant_id = $2`

	var execution Execution
	err := r.db.GetContext(ctx, &execution, query, id, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &execution, nil
}

// UpdateExecutionStatus updates an execution's status
func (r *Repository) UpdateExecutionStatus(ctx context.Context, id string, status ExecutionStatus, outputData []byte, errorMessage *string) error {
	now := time.Now()

	var startedAt, completedAt *time.Time
	if status == ExecutionStatusRunning {
		startedAt = &now
	}
	if status == ExecutionStatusCompleted || status == ExecutionStatusFailed || status == ExecutionStatusCancelled {
		completedAt = &now
	}

	// Handle nil or empty output data
	var outputDataParam interface{}
	if len(outputData) == 0 {
		outputDataParam = nil
	} else {
		outputDataParam = outputData
	}

	query := `
		UPDATE executions
		SET status = $2,
		    output_data = COALESCE($3, output_data),
		    error_message = COALESCE($4, error_message),
		    started_at = COALESCE($5, started_at),
		    completed_at = COALESCE($6, completed_at)
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, status, outputDataParam, errorMessage, startedAt, completedAt)
	return err
}

// ListExecutions retrieves executions for a tenant with pagination
func (r *Repository) ListExecutions(ctx context.Context, tenantID string, workflowID string, limit, offset int) ([]*Execution, error) {
	var query string
	var args []interface{}

	if workflowID != "" {
		query = `
			SELECT * FROM executions
			WHERE tenant_id = $1 AND workflow_id = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{tenantID, workflowID, limit, offset}
	} else {
		query = `
			SELECT * FROM executions
			WHERE tenant_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{tenantID, limit, offset}
	}

	var executions []*Execution
	err := r.db.SelectContext(ctx, &executions, query, args...)
	if err != nil {
		return nil, err
	}

	return executions, nil
}

// CreateStepExecution creates a new step execution record
func (r *Repository) CreateStepExecution(ctx context.Context, executionID, nodeID, nodeType string, inputData []byte) (*StepExecution, error) {
	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO step_executions (id, execution_id, node_id, node_type, status, input_data, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *
	`

	// Handle nil or empty input data
	var inputDataParam interface{}
	if len(inputData) == 0 {
		inputDataParam = nil
	} else {
		inputDataParam = inputData
	}

	var stepExecution StepExecution
	err := r.db.QueryRowxContext(
		ctx, query,
		id, executionID, nodeID, nodeType, "running", inputDataParam, now,
	).StructScan(&stepExecution)

	if err != nil {
		return nil, err
	}

	return &stepExecution, nil
}

// UpdateStepExecution updates a step execution with results
func (r *Repository) UpdateStepExecution(ctx context.Context, id, status string, outputData []byte, errorMessage *string) error {
	now := time.Now()

	// Handle nil or empty output data
	var outputDataParam interface{}
	if len(outputData) == 0 {
		outputDataParam = nil
	} else {
		outputDataParam = outputData
	}

	query := `
		UPDATE step_executions
		SET status = $2,
		    output_data = COALESCE($3, output_data),
		    error_message = COALESCE($4, error_message),
		    completed_at = $5,
		    duration_ms = EXTRACT(EPOCH FROM ($5 - started_at)) * 1000
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, status, outputDataParam, errorMessage, now)
	return err
}

// GetStepExecutionsByExecutionID retrieves all step executions for an execution
func (r *Repository) GetStepExecutionsByExecutionID(ctx context.Context, executionID string) ([]*StepExecution, error) {
	query := `
		SELECT * FROM step_executions
		WHERE execution_id = $1
		ORDER BY started_at ASC
	`

	var stepExecutions []*StepExecution
	err := r.db.SelectContext(ctx, &stepExecutions, query, executionID)
	if err != nil {
		return nil, err
	}

	return stepExecutions, nil
}

// buildExecutionFilterQuery builds the WHERE clause for execution filters
func (r *Repository) buildExecutionFilterQuery(filter ExecutionFilter, args []interface{}, argIndex int) (string, []interface{}) {
	var conditions []string

	if filter.WorkflowID != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("workflow_id = $%d", argIndex))
		args = append(args, filter.WorkflowID)
	}

	if filter.Status != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
	}

	if filter.TriggerType != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("trigger_type = $%d", argIndex))
		args = append(args, filter.TriggerType)
	}

	if filter.StartDate != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filter.StartDate)
	}

	if filter.EndDate != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, filter.EndDate)
	}

	if filter.ErrorSearch != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("error_message ILIKE $%d", argIndex))
		args = append(args, "%"+filter.ErrorSearch+"%")
	}

	if filter.ExecutionIDPrefix != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("id LIKE $%d", argIndex))
		args = append(args, filter.ExecutionIDPrefix+"%")
	}

	if filter.MinDurationMs != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("(EXTRACT(EPOCH FROM (completed_at - started_at)) * 1000) >= $%d", argIndex))
		args = append(args, *filter.MinDurationMs)
	}

	if filter.MaxDurationMs != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("(EXTRACT(EPOCH FROM (completed_at - started_at)) * 1000) <= $%d", argIndex))
		args = append(args, *filter.MaxDurationMs)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + joinConditions(conditions)
	}

	return whereClause, args
}

// joinConditions joins SQL conditions with AND
func joinConditions(conditions []string) string {
	result := ""
	for i, cond := range conditions {
		if i > 0 {
			result += " AND "
		}
		result += cond
	}
	return result
}

// ListExecutionsAdvanced retrieves executions with advanced filtering and cursor-based pagination
func (r *Repository) ListExecutionsAdvanced(ctx context.Context, tenantID string, filter ExecutionFilter, cursor string, limit int) (*ExecutionListResult, error) {
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	if limit <= 0 {
		limit = 20
	}

	// Decode cursor if provided
	var cursorData PaginationCursor
	if cursor != "" {
		decoded, err := DecodePaginationCursor(cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		cursorData = decoded
	}

	// Build query
	args := []interface{}{tenantID}
	argIndex := 1

	// Add cursor conditions
	cursorCondition := ""
	if cursor != "" {
		argIndex++
		args = append(args, cursorData.CreatedAt)
		argIndex++
		args = append(args, cursorData.ID)
		cursorCondition = fmt.Sprintf(" AND (created_at < $%d OR (created_at = $%d AND id < $%d))", argIndex-1, argIndex-1, argIndex)
	}

	// Add filter conditions
	filterConditions, args := r.buildExecutionFilterQuery(filter, args, argIndex)

	query := fmt.Sprintf(`
		SELECT * FROM executions
		WHERE tenant_id = $1%s%s
		ORDER BY created_at DESC, id DESC
		LIMIT %d
	`, cursorCondition, filterConditions, limit+1) // Fetch one extra to check if there are more

	var executions []*Execution
	err := r.db.SelectContext(ctx, &executions, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list executions: %w", err)
	}

	// Check if there are more results
	hasMore := len(executions) > limit
	if hasMore {
		executions = executions[:limit]
	}

	// Generate cursor for next page
	var nextCursor string
	if hasMore && len(executions) > 0 {
		lastExec := executions[len(executions)-1]
		cursorData := PaginationCursor{
			CreatedAt: lastExec.CreatedAt,
			ID:        lastExec.ID,
		}
		nextCursor = cursorData.Encode()
	}

	// Get total count
	totalCount, err := r.CountExecutions(ctx, tenantID, filter)
	if err != nil {
		return nil, fmt.Errorf("count executions: %w", err)
	}

	return &ExecutionListResult{
		Data:       executions,
		Cursor:     nextCursor,
		HasMore:    hasMore,
		TotalCount: totalCount,
	}, nil
}

// GetExecutionWithSteps retrieves an execution with all its step executions
func (r *Repository) GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*ExecutionWithSteps, error) {
	execution, err := r.GetExecutionByID(ctx, tenantID, executionID)
	if err != nil {
		return nil, fmt.Errorf("get execution: %w", err)
	}

	steps, err := r.GetStepExecutionsByExecutionID(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("get step executions: %w", err)
	}

	return &ExecutionWithSteps{
		Execution: execution,
		Steps:     steps,
	}, nil
}

// CountExecutions returns the total count of executions matching the filter
func (r *Repository) CountExecutions(ctx context.Context, tenantID string, filter ExecutionFilter) (int, error) {
	if err := filter.Validate(); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}

	args := []interface{}{tenantID}
	filterConditions, args := r.buildExecutionFilterQuery(filter, args, 1)

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM executions
		WHERE tenant_id = $1%s
	`, filterConditions)

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("count executions: %w", err)
	}

	return count, nil
}

// CreateWorkflowVersion creates a new workflow version record
func (r *Repository) CreateWorkflowVersion(ctx context.Context, workflowID string, version int, definition json.RawMessage, createdBy string) (*WorkflowVersion, error) {
	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO workflow_versions (id, workflow_id, version, definition, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *
	`

	var workflowVersion WorkflowVersion
	err := r.db.QueryRowxContext(
		ctx, query,
		id, workflowID, version, definition, createdBy, now,
	).StructScan(&workflowVersion)

	if err != nil {
		return nil, fmt.Errorf("create workflow version: %w", err)
	}

	return &workflowVersion, nil
}

// ListWorkflowVersions retrieves all versions for a workflow
func (r *Repository) ListWorkflowVersions(ctx context.Context, workflowID string) ([]*WorkflowVersion, error) {
	query := `
		SELECT * FROM workflow_versions
		WHERE workflow_id = $1
		ORDER BY version DESC
	`

	var versions []*WorkflowVersion
	err := r.db.SelectContext(ctx, &versions, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("list workflow versions: %w", err)
	}

	return versions, nil
}

// GetWorkflowVersion retrieves a specific version of a workflow
func (r *Repository) GetWorkflowVersion(ctx context.Context, workflowID string, version int) (*WorkflowVersion, error) {
	query := `
		SELECT * FROM workflow_versions
		WHERE workflow_id = $1 AND version = $2
	`

	var workflowVersion WorkflowVersion
	err := r.db.GetContext(ctx, &workflowVersion, query, workflowID, version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get workflow version: %w", err)
	}

	return &workflowVersion, nil
}

// RestoreWorkflowVersion restores a workflow to a previous version
func (r *Repository) RestoreWorkflowVersion(ctx context.Context, tenantID, workflowID string, version int) (*Workflow, error) {
	// Get the version to restore
	versionData, err := r.GetWorkflowVersion(ctx, workflowID, version)
	if err != nil {
		return nil, err
	}

	// Update the workflow with the version's definition (increments version automatically)
	updatedWorkflow, err := r.Update(ctx, tenantID, workflowID, UpdateWorkflowInput{
		Definition: versionData.Definition,
	})
	if err != nil {
		return nil, fmt.Errorf("restore workflow version: %w", err)
	}

	return updatedWorkflow, nil
}
