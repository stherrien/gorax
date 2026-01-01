package marketplace

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

// MarketplaceTemplate represents a template in the marketplace
type MarketplaceTemplate struct {
	ID              string          `db:"id" json:"id"`
	Name            string          `db:"name" json:"name"`
	Description     string          `db:"description" json:"description"`
	Category        string          `db:"category" json:"category"`
	Definition      json.RawMessage `db:"definition" json:"definition"`
	Tags            pq.StringArray  `db:"tags" json:"tags"`
	AuthorID        string          `db:"author_id" json:"author_id"`
	AuthorName      string          `db:"author_name" json:"author_name"`
	Version         string          `db:"version" json:"version"`
	DownloadCount   int             `db:"download_count" json:"download_count"`
	AverageRating   float64         `db:"average_rating" json:"average_rating"`
	TotalRatings    int             `db:"total_ratings" json:"total_ratings"`
	IsVerified      bool            `db:"is_verified" json:"is_verified"`
	SourceTenantID  *string         `db:"source_tenant_id" json:"source_tenant_id,omitempty"`
	SourceTemplateID *string        `db:"source_template_id" json:"source_template_id,omitempty"`
	PublishedAt     time.Time       `db:"published_at" json:"published_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

// TemplateVersion represents a version of a marketplace template
type TemplateVersion struct {
	ID          string          `db:"id" json:"id"`
	TemplateID  string          `db:"template_id" json:"template_id"`
	Version     string          `db:"version" json:"version"`
	Definition  json.RawMessage `db:"definition" json:"definition"`
	ChangeNotes string          `db:"change_notes" json:"change_notes"`
	CreatedBy   string          `db:"created_by" json:"created_by"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}

// TemplateReview represents a user review for a marketplace template
type TemplateReview struct {
	ID         string     `db:"id" json:"id"`
	TemplateID string     `db:"template_id" json:"template_id"`
	TenantID   string     `db:"tenant_id" json:"tenant_id"`
	UserID     string     `db:"user_id" json:"user_id"`
	UserName   string     `db:"user_name" json:"user_name"`
	Rating     int        `db:"rating" json:"rating"`
	Comment    string     `db:"comment" json:"comment"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}

// TemplateInstallation tracks template installations
type TemplateInstallation struct {
	ID               string    `db:"id" json:"id"`
	TemplateID       string    `db:"template_id" json:"template_id"`
	TenantID         string    `db:"tenant_id" json:"tenant_id"`
	UserID           string    `db:"user_id" json:"user_id"`
	WorkflowID       string    `db:"workflow_id" json:"workflow_id"`
	InstalledVersion string    `db:"installed_version" json:"installed_version"`
	InstalledAt      time.Time `db:"installed_at" json:"installed_at"`
}

// TemplateCategory represents marketplace template categories
type TemplateCategory string

const (
	CategorySecurity     TemplateCategory = "security"
	CategoryMonitoring   TemplateCategory = "monitoring"
	CategoryIntegration  TemplateCategory = "integration"
	CategoryDataOps      TemplateCategory = "dataops"
	CategoryDevOps       TemplateCategory = "devops"
	CategoryNotification TemplateCategory = "notification"
	CategoryAutomation   TemplateCategory = "automation"
	CategoryAnalytics    TemplateCategory = "analytics"
	CategoryOther        TemplateCategory = "other"
)

// PublishTemplateInput represents input for publishing a template
type PublishTemplateInput struct {
	Name             string          `json:"name" validate:"required,min=1,max=255"`
	Description      string          `json:"description" validate:"required,min=10,max=5000"`
	Category         string          `json:"category" validate:"required,max=100"`
	Definition       json.RawMessage `json:"definition" validate:"required"`
	Tags             []string        `json:"tags"`
	Version          string          `json:"version" validate:"required,semver"`
	SourceTemplateID *string         `json:"source_template_id,omitempty"`
}

// UpdateTemplateInput represents input for updating a marketplace template
type UpdateTemplateInput struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Category    string          `json:"category,omitempty"`
	Definition  json.RawMessage `json:"definition,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Version     string          `json:"version,omitempty"`
}

// InstallTemplateInput represents input for installing a template
type InstallTemplateInput struct {
	WorkflowName string `json:"workflow_name" validate:"required,min=1,max=255"`
}

// RateTemplateInput represents input for rating a template
type RateTemplateInput struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment,omitempty" validate:"max=2000"`
}

// SearchFilter represents search filters for marketplace templates
type SearchFilter struct {
	Category      string   `json:"category,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	SearchQuery   string   `json:"search_query,omitempty"`
	MinRating     *float64 `json:"min_rating,omitempty"`
	IsVerified    *bool    `json:"is_verified,omitempty"`
	SortBy        string   `json:"sort_by,omitempty"` // popular, recent, rating
	Page          int      `json:"page,omitempty"`
	Limit         int      `json:"limit,omitempty"`
}

// InstallTemplateResult represents the result of template installation
type InstallTemplateResult struct {
	WorkflowID   string          `json:"workflow_id"`
	WorkflowName string          `json:"workflow_name"`
	Definition   json.RawMessage `json:"definition"`
}

// Validate validates the publish template input
func (i PublishTemplateInput) Validate() error {
	if i.Name == "" {
		return errors.New("name is required")
	}
	if len(i.Name) > 255 {
		return errors.New("name must be 255 characters or less")
	}
	if i.Description == "" {
		return errors.New("description is required")
	}
	if len(i.Description) < 10 {
		return errors.New("description must be at least 10 characters")
	}
	if len(i.Description) > 5000 {
		return errors.New("description must be 5000 characters or less")
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
	if i.Version == "" {
		return errors.New("version is required")
	}

	var def map[string]interface{}
	if err := json.Unmarshal(i.Definition, &def); err != nil {
		return errors.New("definition must be valid JSON")
	}

	return nil
}

// Validate validates the rate template input
func (i RateTemplateInput) Validate() error {
	if i.Rating < 1 || i.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	if len(i.Comment) > 2000 {
		return errors.New("comment must be 2000 characters or less")
	}
	return nil
}

// Validate validates the search filter
func (f SearchFilter) Validate() error {
	if f.Category != "" && len(f.Category) > 100 {
		return errors.New("category must be 100 characters or less")
	}
	if f.MinRating != nil && (*f.MinRating < 0 || *f.MinRating > 5) {
		return errors.New("min_rating must be between 0 and 5")
	}
	if f.SortBy != "" && f.SortBy != "popular" && f.SortBy != "recent" && f.SortBy != "rating" {
		return errors.New("sort_by must be one of: popular, recent, rating")
	}
	if f.Page < 0 {
		return errors.New("page must be non-negative")
	}
	if f.Limit < 0 {
		return errors.New("limit must be non-negative")
	}
	if f.Limit > 100 {
		return errors.New("limit must be 100 or less")
	}
	return nil
}

// GetCategories returns all available template categories
func GetCategories() []string {
	return []string{
		string(CategorySecurity),
		string(CategoryMonitoring),
		string(CategoryIntegration),
		string(CategoryDataOps),
		string(CategoryDevOps),
		string(CategoryNotification),
		string(CategoryAutomation),
		string(CategoryAnalytics),
		string(CategoryOther),
	}
}
