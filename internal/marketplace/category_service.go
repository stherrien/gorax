package marketplace

import (
	"context"
	"errors"
	"fmt"
)

// CategoryService handles business logic for categories
type CategoryService interface {
	CreateCategory(ctx context.Context, input CreateCategoryInput) (*Category, error)
	GetCategory(ctx context.Context, id string) (*Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*Category, error)
	ListCategories(ctx context.Context, parentID *string) ([]Category, error)
	GetCategoriesWithHierarchy(ctx context.Context) ([]Category, error)
	UpdateCategory(ctx context.Context, id string, input UpdateCategoryInput) (*Category, error)
	DeleteCategory(ctx context.Context, id string) error
	GetTemplateCategories(ctx context.Context, templateID string) ([]Category, error)
	SetTemplateCategories(ctx context.Context, templateID string, categoryIDs []string) error
}

type categoryService struct {
	repo CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(repo CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

// CreateCategory creates a new category
func (s *categoryService) CreateCategory(ctx context.Context, input CreateCategoryInput) (*Category, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Check if slug already exists
	existing, err := s.repo.GetBySlug(ctx, input.Slug)
	if err == nil && existing != nil {
		return nil, errors.New("category with this slug already exists")
	}

	// Verify parent exists if specified
	if input.ParentID != nil {
		parent, err := s.repo.GetByID(ctx, *input.ParentID)
		if err != nil || parent == nil {
			return nil, errors.New("parent category not found")
		}

		// Prevent deep nesting (max 2 levels)
		if parent.ParentID != nil {
			return nil, errors.New("categories can only be nested 2 levels deep")
		}
	}

	category, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// GetCategory retrieves a category by ID
func (s *categoryService) GetCategory(ctx context.Context, id string) (*Category, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

// GetCategoryBySlug retrieves a category by slug
func (s *categoryService) GetCategoryBySlug(ctx context.Context, slug string) (*Category, error) {
	category, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

// ListCategories retrieves all categories
func (s *categoryService) ListCategories(ctx context.Context, parentID *string) ([]Category, error) {
	categories, err := s.repo.List(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	return categories, nil
}

// GetCategoriesWithHierarchy retrieves all categories with children
func (s *categoryService) GetCategoriesWithHierarchy(ctx context.Context) ([]Category, error) {
	categories, err := s.repo.GetWithChildren(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories with hierarchy: %w", err)
	}

	return categories, nil
}

// UpdateCategory updates a category
func (s *categoryService) UpdateCategory(ctx context.Context, id string, input UpdateCategoryInput) (*Category, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Verify category exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	// Check slug uniqueness if changing
	if input.Slug != "" && input.Slug != existing.Slug {
		slugExists, err := s.repo.GetBySlug(ctx, input.Slug)
		if err == nil && slugExists != nil {
			return nil, errors.New("category with this slug already exists")
		}
	}

	// Verify parent exists if changing
	if input.ParentID != nil {
		if *input.ParentID == id {
			return nil, errors.New("category cannot be its own parent")
		}

		parent, err := s.repo.GetByID(ctx, *input.ParentID)
		if err != nil || parent == nil {
			return nil, errors.New("parent category not found")
		}

		// Prevent deep nesting
		if parent.ParentID != nil {
			return nil, errors.New("categories can only be nested 2 levels deep")
		}

		// Prevent circular references
		if err := s.checkCircularReference(ctx, id, *input.ParentID); err != nil {
			return nil, err
		}
	}

	category, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

// DeleteCategory deletes a category
func (s *categoryService) DeleteCategory(ctx context.Context, id string) error {
	// Check if category exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Check if category has children
	children, err := s.repo.List(ctx, &id)
	if err != nil {
		return fmt.Errorf("failed to check for child categories: %w", err)
	}

	if len(children) > 0 {
		return errors.New("cannot delete category with child categories")
	}

	// Note: ON DELETE CASCADE in DB will remove template associations
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// GetTemplateCategories retrieves all categories for a template
func (s *categoryService) GetTemplateCategories(ctx context.Context, templateID string) ([]Category, error) {
	categories, err := s.repo.GetTemplateCategories(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template categories: %w", err)
	}

	return categories, nil
}

// SetTemplateCategories sets all categories for a template
func (s *categoryService) SetTemplateCategories(ctx context.Context, templateID string, categoryIDs []string) error {
	// Verify all categories exist
	for _, categoryID := range categoryIDs {
		_, err := s.repo.GetByID(ctx, categoryID)
		if err != nil {
			return fmt.Errorf("category %s not found: %w", categoryID, err)
		}
	}

	// Limit to maximum 5 categories per template
	if len(categoryIDs) > 5 {
		return errors.New("templates can have a maximum of 5 categories")
	}

	err := s.repo.SetTemplateCategories(ctx, templateID, categoryIDs)
	if err != nil {
		return fmt.Errorf("failed to set template categories: %w", err)
	}

	return nil
}

// checkCircularReference checks for circular parent-child references
func (s *categoryService) checkCircularReference(ctx context.Context, categoryID, newParentID string) error {
	current := newParentID
	visited := make(map[string]bool)

	for current != "" {
		if current == categoryID {
			return errors.New("circular reference detected")
		}

		if visited[current] {
			return errors.New("circular reference detected in category hierarchy")
		}

		visited[current] = true

		category, err := s.repo.GetByID(ctx, current)
		if err != nil {
			break
		}

		if category.ParentID == nil {
			break
		}

		current = *category.ParentID
	}

	return nil
}
