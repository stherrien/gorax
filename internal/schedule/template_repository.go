package schedule

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// TemplateRepository defines schedule template data access operations
type TemplateRepository interface {
	List(ctx context.Context, filter ScheduleTemplateFilter) ([]*ScheduleTemplate, error)
	GetByID(ctx context.Context, id string) (*ScheduleTemplate, error)
}

// PostgresTemplateRepository implements TemplateRepository using PostgreSQL
type PostgresTemplateRepository struct {
	db *sqlx.DB
}

// NewTemplateRepository creates a new schedule template repository
func NewTemplateRepository(db *sqlx.DB) TemplateRepository {
	return &PostgresTemplateRepository{db: db}
}

// List retrieves schedule templates with optional filters
func (r *PostgresTemplateRepository) List(ctx context.Context, filter ScheduleTemplateFilter) ([]*ScheduleTemplate, error) {
	query := `
		SELECT id, name, description, category, cron_expression,
			   timezone, tags, is_system, created_at
		FROM schedule_templates
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 0

	if filter.Category != "" {
		argCount++
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, filter.Category)
	}

	if len(filter.Tags) > 0 {
		argCount++
		query += fmt.Sprintf(" AND tags && $%d", argCount)
		args = append(args, pq.Array(filter.Tags))
	}

	if filter.IsSystem != nil {
		argCount++
		query += fmt.Sprintf(" AND is_system = $%d", argCount)
		args = append(args, *filter.IsSystem)
	}

	if filter.SearchQuery != "" {
		argCount++
		searchPattern := "%" + filter.SearchQuery + "%"
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		args = append(args, searchPattern)
	}

	query += " ORDER BY category, name"

	var templates []*ScheduleTemplate
	err := r.db.SelectContext(ctx, &templates, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list schedule templates: %w", err)
	}

	return templates, nil
}

// GetByID retrieves a schedule template by ID
func (r *PostgresTemplateRepository) GetByID(ctx context.Context, id string) (*ScheduleTemplate, error) {
	query := `
		SELECT id, name, description, category, cron_expression,
			   timezone, tags, is_system, created_at
		FROM schedule_templates
		WHERE id = $1
	`

	var template ScheduleTemplate
	err := r.db.GetContext(ctx, &template, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("get schedule template: %w", err)
	}

	return &template, nil
}

// ApplyTemplate creates a schedule from a template
func ApplyTemplate(
	ctx context.Context,
	template *ScheduleTemplate,
	input ApplyTemplateInput,
	scheduleRepo Repository,
	tenantID, userID string,
) (*Schedule, error) {
	scheduleName := template.Name
	if input.Name != nil && *input.Name != "" {
		scheduleName = *input.Name
	}

	timezone := template.Timezone
	if input.Timezone != nil && *input.Timezone != "" {
		timezone = *input.Timezone
	}

	createInput := CreateScheduleInput{
		Name:           scheduleName,
		CronExpression: template.CronExpression,
		Timezone:       timezone,
		Enabled:        true,
	}

	schedule, err := scheduleRepo.Create(ctx, tenantID, input.WorkflowID, userID, createInput)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("schedule with name %s already exists for this workflow", scheduleName)
		}
		return nil, fmt.Errorf("create schedule from template: %w", err)
	}

	return schedule, nil
}
