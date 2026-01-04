package marketplace

import (
	"errors"
	"regexp"
	"time"
)

// Category represents a marketplace template category
type Category struct {
	ID            string     `db:"id" json:"id"`
	Name          string     `db:"name" json:"name"`
	Slug          string     `db:"slug" json:"slug"`
	Description   string     `db:"description" json:"description"`
	Icon          string     `db:"icon" json:"icon"`
	ParentID      *string    `db:"parent_id" json:"parent_id,omitempty"`
	DisplayOrder  int        `db:"display_order" json:"display_order"`
	TemplateCount int        `db:"template_count" json:"template_count"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
	Children      []Category `json:"children,omitempty"` // Populated for hierarchical queries
}

// CreateCategoryInput represents input for creating a category
type CreateCategoryInput struct {
	Name         string  `json:"name" validate:"required,min=1,max=255"`
	Slug         string  `json:"slug" validate:"required,min=1,max=255"`
	Description  string  `json:"description,omitempty"`
	Icon         string  `json:"icon,omitempty"`
	ParentID     *string `json:"parent_id,omitempty"`
	DisplayOrder int     `json:"display_order,omitempty"`
}

// UpdateCategoryInput represents input for updating a category
type UpdateCategoryInput struct {
	Name         string  `json:"name,omitempty"`
	Slug         string  `json:"slug,omitempty"`
	Description  string  `json:"description,omitempty"`
	Icon         string  `json:"icon,omitempty"`
	ParentID     *string `json:"parent_id,omitempty"`
	DisplayOrder *int    `json:"display_order,omitempty"`
}

// FeatureTemplateInput represents input for featuring/unfeaturing a template
type FeatureTemplateInput struct {
	IsFeatured bool `json:"is_featured"`
}

// TemplateCategoryAssociation represents junction table entry
type TemplateCategoryAssociation struct {
	ID         string    `db:"id" json:"id"`
	TemplateID string    `db:"template_id" json:"template_id"`
	CategoryID string    `db:"category_id" json:"category_id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// Validate validates the category
func (c Category) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if len(c.Name) > 255 {
		return errors.New("name must be 255 characters or less")
	}
	if c.Slug == "" {
		return errors.New("slug is required")
	}
	if !isValidSlug(c.Slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}
	if len(c.Description) > 1000 {
		return errors.New("description must be 1000 characters or less")
	}
	return nil
}

// Validate validates the create category input
func (i CreateCategoryInput) Validate() error {
	if i.Name == "" {
		return errors.New("name is required")
	}
	if len(i.Name) > 255 {
		return errors.New("name must be 255 characters or less")
	}
	if i.Slug == "" {
		return errors.New("slug is required")
	}
	if !isValidSlug(i.Slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}
	if len(i.Description) > 1000 {
		return errors.New("description must be 1000 characters or less")
	}
	return nil
}

// Validate validates the update category input
func (i UpdateCategoryInput) Validate() error {
	if i.Name != "" && len(i.Name) > 255 {
		return errors.New("name must be 255 characters or less")
	}
	if i.Slug != "" && !isValidSlug(i.Slug) {
		return errors.New("slug must contain only lowercase letters, numbers, and hyphens")
	}
	if len(i.Description) > 1000 {
		return errors.New("description must be 1000 characters or less")
	}
	if i.DisplayOrder != nil && *i.DisplayOrder < 0 {
		return errors.New("display_order must be non-negative")
	}
	return nil
}

// Validate validates the feature template input
func (i FeatureTemplateInput) Validate() error {
	// No validation needed for boolean
	return nil
}

// isValidSlug checks if a slug contains only lowercase letters, numbers, and hyphens
func isValidSlug(slug string) bool {
	match, _ := regexp.MatchString("^[a-z0-9-]+$", slug)
	return match
}
