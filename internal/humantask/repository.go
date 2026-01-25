package humantask

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository defines the interface for human task data access
type Repository interface {
	Create(ctx context.Context, task *HumanTask) error
	GetByID(ctx context.Context, id uuid.UUID) (*HumanTask, error)
	List(ctx context.Context, filter TaskFilter) ([]*HumanTask, error)
	Update(ctx context.Context, task *HumanTask) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetOverdueTasks(ctx context.Context, tenantID uuid.UUID) ([]*HumanTask, error)
	CountPendingByAssignee(ctx context.Context, tenantID uuid.UUID, assignee string) (int, error)

	// Escalation-related methods
	CreateEscalation(ctx context.Context, escalation *TaskEscalation) error
	GetEscalationsByTaskID(ctx context.Context, taskID uuid.UUID) ([]*TaskEscalation, error)
	GetActiveEscalation(ctx context.Context, taskID uuid.UUID) (*TaskEscalation, error)
	UpdateEscalation(ctx context.Context, escalation *TaskEscalation) error
	CompleteEscalationsByTaskID(ctx context.Context, taskID uuid.UUID, completedBy *uuid.UUID) error
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new human task repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, task *HumanTask) error {
	query := `
		INSERT INTO human_tasks (
			tenant_id, execution_id, step_id, task_type, title, description,
			assignees, status, due_date, config
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowxContext(
		ctx, query,
		task.TenantID,
		task.ExecutionID,
		task.StepID,
		task.TaskType,
		task.Title,
		task.Description,
		task.Assignees,
		task.Status,
		task.DueDate,
		task.Config,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create human task: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*HumanTask, error) {
	var task HumanTask

	query := `
		SELECT id, tenant_id, execution_id, step_id, task_type, title, description,
			assignees, status, due_date, completed_at, completed_by, response_data,
			config, escalation_level, max_escalation_level, last_escalated_at,
			created_at, updated_at
		FROM human_tasks
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &task, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("get human task: %w", err)
	}

	return &task, nil
}

func (r *repository) List(ctx context.Context, filter TaskFilter) ([]*HumanTask, error) {
	var tasks []*HumanTask

	query := `
		SELECT id, tenant_id, execution_id, step_id, task_type, title, description,
			assignees, status, due_date, completed_at, completed_by, response_data,
			config, escalation_level, max_escalation_level, last_escalated_at,
			created_at, updated_at
		FROM human_tasks
		WHERE tenant_id = $1
	`

	args := []interface{}{filter.TenantID}
	argPos := 2

	// Add filters
	if filter.ExecutionID != nil {
		query += fmt.Sprintf(" AND execution_id = $%d", argPos)
		args = append(args, *filter.ExecutionID)
		argPos++
	}

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filter.Status)
		argPos++
	}

	if filter.TaskType != nil {
		query += fmt.Sprintf(" AND task_type = $%d", argPos)
		args = append(args, *filter.TaskType)
		argPos++
	}

	if filter.Assignee != nil {
		query += fmt.Sprintf(" AND assignees @> $%d::jsonb", argPos)
		assigneeJSON, err := json.Marshal([]string{*filter.Assignee})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal assignee filter: %w", err)
		}
		args = append(args, assigneeJSON)
		argPos++
	}

	if filter.DueBefore != nil {
		query += fmt.Sprintf(" AND due_date < $%d", argPos)
		args = append(args, *filter.DueBefore)
		argPos++
	}

	if filter.DueAfter != nil {
		query += fmt.Sprintf(" AND due_date > $%d", argPos)
		args = append(args, *filter.DueAfter)
		argPos++
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	err := r.db.SelectContext(ctx, &tasks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list human tasks: %w", err)
	}

	return tasks, nil
}

func (r *repository) Update(ctx context.Context, task *HumanTask) error {
	query := `
		UPDATE human_tasks
		SET status = $1,
			completed_at = $2,
			completed_by = $3,
			response_data = $4,
			due_date = $5,
			config = $6,
			assignees = $7,
			escalation_level = $8,
			max_escalation_level = $9,
			last_escalated_at = $10
		WHERE id = $11
	`

	result, err := r.db.ExecContext(
		ctx, query,
		task.Status,
		task.CompletedAt,
		task.CompletedBy,
		task.ResponseData,
		task.DueDate,
		task.Config,
		task.Assignees,
		task.EscalationLevel,
		task.MaxEscalationLevel,
		task.LastEscalatedAt,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("update human task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if rows == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM human_tasks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete human task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if rows == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *repository) GetOverdueTasks(ctx context.Context, tenantID uuid.UUID) ([]*HumanTask, error) {
	var tasks []*HumanTask

	query := `
		SELECT id, tenant_id, execution_id, step_id, task_type, title, description,
			assignees, status, due_date, completed_at, completed_by, response_data,
			config, escalation_level, max_escalation_level, last_escalated_at,
			created_at, updated_at
		FROM human_tasks
		WHERE tenant_id = $1
			AND status = $2
			AND due_date < NOW()
		ORDER BY due_date ASC
	`

	err := r.db.SelectContext(ctx, &tasks, query, tenantID, StatusPending)
	if err != nil {
		return nil, fmt.Errorf("get overdue tasks: %w", err)
	}

	return tasks, nil
}

func (r *repository) CountPendingByAssignee(ctx context.Context, tenantID uuid.UUID, assignee string) (int, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM human_tasks
		WHERE tenant_id = $1
			AND status = $2
			AND assignees @> $3::jsonb
	`

	assigneeJSON, err := json.Marshal([]string{assignee})
	if err != nil {
		return 0, fmt.Errorf("failed to marshal assignee: %w", err)
	}

	err = r.db.GetContext(ctx, &count, query, tenantID, StatusPending, assigneeJSON)
	if err != nil {
		return 0, fmt.Errorf("count pending tasks: %w", err)
	}

	return count, nil
}

// buildFilterQuery builds a SQL query with dynamic filters
func buildFilterQuery(baseQuery string, filter TaskFilter) (string, []interface{}, error) {
	var conditions []string
	var args []interface{}
	argPos := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argPos))
	args = append(args, filter.TenantID)
	argPos++

	if filter.ExecutionID != nil {
		conditions = append(conditions, fmt.Sprintf("execution_id = $%d", argPos))
		args = append(args, *filter.ExecutionID)
		argPos++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *filter.Status)
		argPos++
	}

	if filter.TaskType != nil {
		conditions = append(conditions, fmt.Sprintf("task_type = $%d", argPos))
		args = append(args, *filter.TaskType)
		argPos++
	}

	if filter.Assignee != nil {
		conditions = append(conditions, fmt.Sprintf("assignees @> $%d::jsonb", argPos))
		assigneeJSON, err := json.Marshal([]string{*filter.Assignee})
		if err != nil {
			return "", nil, fmt.Errorf("failed to marshal assignee filter: %w", err)
		}
		args = append(args, assigneeJSON)
		argPos++
	}

	if filter.DueBefore != nil {
		conditions = append(conditions, fmt.Sprintf("due_date < $%d", argPos))
		args = append(args, *filter.DueBefore)
		argPos++
	}

	if filter.DueAfter != nil {
		conditions = append(conditions, fmt.Sprintf("due_date > $%d", argPos))
		args = append(args, *filter.DueAfter)
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args, nil
}

// Escalation repository methods

func (r *repository) CreateEscalation(ctx context.Context, escalation *TaskEscalation) error {
	query := `
		INSERT INTO task_escalations (
			task_id, escalation_level, escalated_from, escalated_to,
			escalation_reason, timeout_minutes, auto_action_taken, status, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id, escalated_at, created_at
	`

	err := r.db.QueryRowxContext(
		ctx, query,
		escalation.TaskID,
		escalation.EscalationLevel,
		escalation.EscalatedFrom,
		escalation.EscalatedTo,
		escalation.EscalationReason,
		escalation.TimeoutMinutes,
		escalation.AutoActionTaken,
		escalation.Status,
		escalation.Metadata,
	).Scan(&escalation.ID, &escalation.EscalatedAt, &escalation.CreatedAt)

	if err != nil {
		return fmt.Errorf("create task escalation: %w", err)
	}

	return nil
}

func (r *repository) GetEscalationsByTaskID(ctx context.Context, taskID uuid.UUID) ([]*TaskEscalation, error) {
	var escalations []*TaskEscalation

	query := `
		SELECT id, task_id, escalation_level, escalated_at, escalated_from, escalated_to,
			escalation_reason, timeout_minutes, auto_action_taken, status,
			completed_at, completed_by, metadata, created_at
		FROM task_escalations
		WHERE task_id = $1
		ORDER BY escalation_level ASC
	`

	err := r.db.SelectContext(ctx, &escalations, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("get escalations by task: %w", err)
	}

	return escalations, nil
}

func (r *repository) GetActiveEscalation(ctx context.Context, taskID uuid.UUID) (*TaskEscalation, error) {
	var escalation TaskEscalation

	query := `
		SELECT id, task_id, escalation_level, escalated_at, escalated_from, escalated_to,
			escalation_reason, timeout_minutes, auto_action_taken, status,
			completed_at, completed_by, metadata, created_at
		FROM task_escalations
		WHERE task_id = $1 AND status = $2
		ORDER BY escalation_level DESC
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &escalation, query, taskID, EscalationStatusActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get active escalation: %w", err)
	}

	return &escalation, nil
}

func (r *repository) UpdateEscalation(ctx context.Context, escalation *TaskEscalation) error {
	query := `
		UPDATE task_escalations
		SET status = $1,
			completed_at = $2,
			completed_by = $3,
			auto_action_taken = $4,
			metadata = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(
		ctx, query,
		escalation.Status,
		escalation.CompletedAt,
		escalation.CompletedBy,
		escalation.AutoActionTaken,
		escalation.Metadata,
		escalation.ID,
	)
	if err != nil {
		return fmt.Errorf("update escalation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("escalation not found")
	}

	return nil
}

func (r *repository) CompleteEscalationsByTaskID(ctx context.Context, taskID uuid.UUID, completedBy *uuid.UUID) error {
	query := `
		UPDATE task_escalations
		SET status = $1,
			completed_at = NOW(),
			completed_by = $2
		WHERE task_id = $3 AND status = $4
	`

	_, err := r.db.ExecContext(ctx, query, EscalationStatusCompleted, completedBy, taskID, EscalationStatusActive)
	if err != nil {
		return fmt.Errorf("complete escalations: %w", err)
	}

	return nil
}
