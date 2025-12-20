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
			config, created_at, updated_at
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
			config, created_at, updated_at
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
		assigneeJSON, _ := json.Marshal([]string{*filter.Assignee})
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
			config = $6
		WHERE id = $7
	`

	result, err := r.db.ExecContext(
		ctx, query,
		task.Status,
		task.CompletedAt,
		task.CompletedBy,
		task.ResponseData,
		task.DueDate,
		task.Config,
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
			config, created_at, updated_at
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

	assigneeJSON, _ := json.Marshal([]string{assignee})

	err := r.db.GetContext(ctx, &count, query, tenantID, StatusPending, assigneeJSON)
	if err != nil {
		return 0, fmt.Errorf("count pending tasks: %w", err)
	}

	return count, nil
}

// buildFilterQuery builds a SQL query with dynamic filters
func buildFilterQuery(baseQuery string, filter TaskFilter) (string, []interface{}) {
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
		assigneeJSON, _ := json.Marshal([]string{*filter.Assignee})
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

	return query, args
}
