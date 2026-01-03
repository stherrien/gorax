package marketplace

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// CategoryRepository handles database operations for categories
type CategoryRepository interface {
	Create(ctx context.Context, input CreateCategoryInput) (*Category, error)
	GetByID(ctx context.Context, id string) (*Category, error)
	GetBySlug(ctx context.Context, slug string) (*Category, error)
	List(ctx context.Context, parentID *string) ([]Category, error)
	GetWithChildren(ctx context.Context) ([]Category, error)
	Update(ctx context.Context, id string, input UpdateCategoryInput) (*Category, error)
	Delete(ctx context.Context, id string) error
	GetTemplateCategories(ctx context.Context, templateID string) ([]Category, error)
	AddTemplateCategory(ctx context.Context, templateID, categoryID string) error
	RemoveTemplateCategory(ctx context.Context, templateID, categoryID string) error
	SetTemplateCategories(ctx context.Context, templateID string, categoryIDs []string) error
}

type categoryRepository struct {
	db *sqlx.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *sqlx.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// Create creates a new category
func (r *categoryRepository) Create(ctx context.Context, input CreateCategoryInput) (*Category, error) {
	query := `
		INSERT INTO marketplace_categories (name, slug, description, icon, parent_id, display_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, slug, description, icon, parent_id, display_order, template_count, created_at, updated_at
	`

	var category Category
	err := r.db.GetContext(ctx, &category, query,
		input.Name,
		input.Slug,
		input.Description,
		input.Icon,
		input.ParentID,
		input.DisplayOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &category, nil
}

// GetByID retrieves a category by ID
func (r *categoryRepository) GetByID(ctx context.Context, id string) (*Category, error) {
	query := `
		SELECT id, name, slug, description, icon, parent_id, display_order, template_count, created_at, updated_at
		FROM marketplace_categories
		WHERE id = $1
	`

	var category Category
	err := r.db.GetContext(ctx, &category, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("category not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// GetBySlug retrieves a category by slug
func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*Category, error) {
	query := `
		SELECT id, name, slug, description, icon, parent_id, display_order, template_count, created_at, updated_at
		FROM marketplace_categories
		WHERE slug = $1
	`

	var category Category
	err := r.db.GetContext(ctx, &category, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("category not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// List retrieves all categories with optional parent filter
func (r *categoryRepository) List(ctx context.Context, parentID *string) ([]Category, error) {
	var query string
	var args []interface{}

	if parentID == nil {
		query = `
			SELECT id, name, slug, description, icon, parent_id, display_order, template_count, created_at, updated_at
			FROM marketplace_categories
			WHERE parent_id IS NULL
			ORDER BY display_order, name
		`
	} else {
		query = `
			SELECT id, name, slug, description, icon, parent_id, display_order, template_count, created_at, updated_at
			FROM marketplace_categories
			WHERE parent_id = $1
			ORDER BY display_order, name
		`
		args = append(args, *parentID)
	}

	var categories []Category
	err := r.db.SelectContext(ctx, &categories, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	return categories, nil
}

// GetWithChildren retrieves all categories with their children
func (r *categoryRepository) GetWithChildren(ctx context.Context) ([]Category, error) {
	// Get all categories
	query := `
		SELECT id, name, slug, description, icon, parent_id, display_order, template_count, created_at, updated_at
		FROM marketplace_categories
		ORDER BY display_order, name
	`

	var allCategories []Category
	err := r.db.SelectContext(ctx, &allCategories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	// Build hierarchy
	categoryMap := make(map[string]*Category)
	var rootCategories []Category

	// First pass: create map of all categories
	for i := range allCategories {
		cat := allCategories[i]
		categoryMap[cat.ID] = &cat
	}

	// Second pass: build hierarchy
	for _, cat := range allCategories {
		if cat.ParentID == nil {
			rootCategories = append(rootCategories, cat)
		} else {
			if parent, exists := categoryMap[*cat.ParentID]; exists {
				parent.Children = append(parent.Children, cat)
			}
		}
	}

	return rootCategories, nil
}

// Update updates a category
func (r *categoryRepository) Update(ctx context.Context, id string, input UpdateCategoryInput) (*Category, error) {
	// Build dynamic update query
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if input.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, input.Name)
		argIndex++
	}
	if input.Slug != "" {
		setClauses = append(setClauses, fmt.Sprintf("slug = $%d", argIndex))
		args = append(args, input.Slug)
		argIndex++
	}
	if input.Description != "" {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, input.Description)
		argIndex++
	}
	if input.Icon != "" {
		setClauses = append(setClauses, fmt.Sprintf("icon = $%d", argIndex))
		args = append(args, input.Icon)
		argIndex++
	}
	if input.ParentID != nil {
		setClauses = append(setClauses, fmt.Sprintf("parent_id = $%d", argIndex))
		args = append(args, *input.ParentID)
		argIndex++
	}
	if input.DisplayOrder != nil {
		setClauses = append(setClauses, fmt.Sprintf("display_order = $%d", argIndex))
		args = append(args, *input.DisplayOrder)
		argIndex++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE marketplace_categories
		SET %s
		WHERE id = $%d
		RETURNING id, name, slug, description, icon, parent_id, display_order, template_count, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIndex)

	var category Category
	err := r.db.GetContext(ctx, &category, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("category not found: %w", err)
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &category, nil
}

// Delete deletes a category
func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM marketplace_categories WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

// GetTemplateCategories retrieves all categories for a template
func (r *categoryRepository) GetTemplateCategories(ctx context.Context, templateID string) ([]Category, error) {
	query := `
		SELECT c.id, c.name, c.slug, c.description, c.icon, c.parent_id, c.display_order, c.template_count, c.created_at, c.updated_at
		FROM marketplace_categories c
		INNER JOIN marketplace_template_categories tc ON c.id = tc.category_id
		WHERE tc.template_id = $1
		ORDER BY c.display_order, c.name
	`

	var categories []Category
	err := r.db.SelectContext(ctx, &categories, query, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template categories: %w", err)
	}

	return categories, nil
}

// AddTemplateCategory adds a category to a template
func (r *categoryRepository) AddTemplateCategory(ctx context.Context, templateID, categoryID string) error {
	query := `
		INSERT INTO marketplace_template_categories (template_id, category_id)
		VALUES ($1, $2)
		ON CONFLICT (template_id, category_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, templateID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to add template category: %w", err)
	}

	return nil
}

// RemoveTemplateCategory removes a category from a template
func (r *categoryRepository) RemoveTemplateCategory(ctx context.Context, templateID, categoryID string) error {
	query := `
		DELETE FROM marketplace_template_categories
		WHERE template_id = $1 AND category_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, templateID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to remove template category: %w", err)
	}

	return nil
}

// SetTemplateCategories sets all categories for a template
func (r *categoryRepository) SetTemplateCategories(ctx context.Context, templateID string, categoryIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Remove all existing categories
	deleteQuery := `DELETE FROM marketplace_template_categories WHERE template_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, templateID)
	if err != nil {
		return fmt.Errorf("failed to remove existing categories: %w", err)
	}

	// Add new categories
	if len(categoryIDs) > 0 {
		insertQuery := `
			INSERT INTO marketplace_template_categories (template_id, category_id)
			VALUES ($1, $2)
		`
		for _, categoryID := range categoryIDs {
			_, err = tx.ExecContext(ctx, insertQuery, templateID, categoryID)
			if err != nil {
				return fmt.Errorf("failed to add category: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
