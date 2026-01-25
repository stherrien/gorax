package template

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Repository defines template data access operations
type Repository interface {
	Create(ctx context.Context, tenantID string, template *Template) error
	GetByID(ctx context.Context, tenantID, id string) (*Template, error)
	List(ctx context.Context, tenantID string, filter TemplateFilter) ([]*Template, error)
	Update(ctx context.Context, tenantID, id string, input UpdateTemplateInput) error
	Delete(ctx context.Context, tenantID, id string) error
	IncrementUsageCount(ctx context.Context, id string) error
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewRepository creates a new template repository
func NewRepository(db *sqlx.DB) Repository {
	return &PostgresRepository{db: db}
}

// Create creates a new template
func (r *PostgresRepository) Create(ctx context.Context, tenantID string, template *Template) error {
	query := `
		INSERT INTO workflow_templates (
			tenant_id, name, description, category, definition,
			tags, is_public, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		template.TenantID,
		template.Name,
		template.Description,
		template.Category,
		template.Definition,
		pq.Array(template.Tags),
		template.IsPublic,
		template.CreatedBy,
	).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("template with name %s already exists", template.Name)
		}
		return fmt.Errorf("create template: %w", err)
	}

	return nil
}

// GetByID retrieves a template by ID
func (r *PostgresRepository) GetByID(ctx context.Context, tenantID, id string) (*Template, error) {
	query := `
		SELECT id, tenant_id, name, description, category, definition,
			   tags, is_public, usage_count, created_by, created_at, updated_at
		FROM workflow_templates
		WHERE id = $1 AND (tenant_id = $2 OR is_public = true)
	`

	var template Template
	err := r.db.GetContext(ctx, &template, query, id, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("get template: %w", err)
	}

	return &template, nil
}

// List retrieves templates with optional filters
func (r *PostgresRepository) List(ctx context.Context, tenantID string, filter TemplateFilter) ([]*Template, error) {
	query := `
		SELECT id, tenant_id, name, description, category, definition,
			   tags, is_public, usage_count, created_by, created_at, updated_at
		FROM workflow_templates
		WHERE (tenant_id = $1 OR is_public = true)
	`

	args := []interface{}{tenantID}
	argCount := 1

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

	if filter.IsPublic != nil {
		argCount++
		query += fmt.Sprintf(" AND is_public = $%d", argCount)
		args = append(args, *filter.IsPublic)
	}

	if filter.SearchQuery != "" {
		argCount++
		searchPattern := "%" + filter.SearchQuery + "%"
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		args = append(args, searchPattern)
	}

	query += " ORDER BY created_at DESC"

	var templates []*Template
	err := r.db.SelectContext(ctx, &templates, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}

	return templates, nil
}

// Update updates an existing template
func (r *PostgresRepository) Update(ctx context.Context, tenantID, id string, input UpdateTemplateInput) error {
	updates := []string{}
	args := []interface{}{}
	argCount := 0

	if input.Name != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("name = $%d", argCount))
		args = append(args, input.Name)
	}

	if input.Description != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("description = $%d", argCount))
		args = append(args, input.Description)
	}

	if input.Category != "" {
		argCount++
		updates = append(updates, fmt.Sprintf("category = $%d", argCount))
		args = append(args, input.Category)
	}

	if input.Definition != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("definition = $%d", argCount))
		args = append(args, input.Definition)
	}

	if input.Tags != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("tags = $%d", argCount))
		args = append(args, pq.Array(input.Tags))
	}

	if input.IsPublic != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("is_public = $%d", argCount))
		args = append(args, *input.IsPublic)
	}

	if len(updates) == 0 {
		return nil
	}

	argCount++
	args = append(args, id)
	argCount++
	args = append(args, tenantID)

	query := fmt.Sprintf(`
		UPDATE workflow_templates
		SET %s, updated_at = NOW()
		WHERE id = $%d AND tenant_id = $%d
	`, strings.Join(updates, ", "), argCount-1, argCount)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

// Delete deletes a template
func (r *PostgresRepository) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM workflow_templates WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

// IncrementUsageCount increments the usage count for a template
func (r *PostgresRepository) IncrementUsageCount(ctx context.Context, id string) error {
	query := `
		UPDATE workflow_templates
		SET usage_count = usage_count + 1
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("increment usage count: %w", err)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
