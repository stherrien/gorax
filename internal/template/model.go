package template

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

// Template represents a reusable workflow pattern
type Template struct {
	ID          string          `db:"id" json:"id"`
	TenantID    *string         `db:"tenant_id" json:"tenant_id,omitempty"`
	Name        string          `db:"name" json:"name"`
	Description string          `db:"description" json:"description"`
	Category    string          `db:"category" json:"category"`
	Definition  json.RawMessage `db:"definition" json:"definition"`
	Tags        pq.StringArray  `db:"tags" json:"tags"`
	IsPublic    bool            `db:"is_public" json:"is_public"`
	UsageCount  int             `db:"usage_count" json:"usage_count"`
	CreatedBy   string          `db:"created_by" json:"created_by"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

// CreateTemplateInput represents input for creating a template
type CreateTemplateInput struct {
	Name        string          `json:"name" validate:"required,min=1,max=255"`
	Description string          `json:"description"`
	Category    string          `json:"category" validate:"required,max=100"`
	Definition  json.RawMessage `json:"definition" validate:"required"`
	Tags        []string        `json:"tags"`
	IsPublic    bool            `json:"is_public"`
}

// UpdateTemplateInput represents input for updating a template
type UpdateTemplateInput struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Category    string          `json:"category,omitempty"`
	Definition  json.RawMessage `json:"definition,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	IsPublic    *bool           `json:"is_public,omitempty"`
}

// TemplateFilter represents filters for listing templates
type TemplateFilter struct {
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsPublic    *bool    `json:"is_public,omitempty"`
	SearchQuery string   `json:"search_query,omitempty"`
}

// TemplateCategory represents template categories
type TemplateCategory string

const (
	CategorySecurity    TemplateCategory = "security"
	CategoryMonitoring  TemplateCategory = "monitoring"
	CategoryIntegration TemplateCategory = "integration"
	CategoryDataOps     TemplateCategory = "dataops"
	CategoryDevOps      TemplateCategory = "devops"
	CategoryOther       TemplateCategory = "other"
)

// Validate validates the create template input
func (i CreateTemplateInput) Validate() error {
	if i.Name == "" {
		return errors.New("name is required")
	}
	if len(i.Name) > 255 {
		return errors.New("name must be 255 characters or less")
	}
	if i.Category == "" {
		return errors.New("category is required")
	}
	if len(i.Category) > 100 {
		return errors.New("category must be 100 characters or less")
	}
	if len(i.Definition) == 0 {
		return errors.New("definition is required")
	}

	var def map[string]interface{}
	if err := json.Unmarshal(i.Definition, &def); err != nil {
		return errors.New("definition must be valid JSON")
	}

	return nil
}

// Validate validates the template filter
func (f TemplateFilter) Validate() error {
	if f.Category != "" && len(f.Category) > 100 {
		return errors.New("category must be 100 characters or less")
	}
	return nil
}
