package marketplace

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

// MarketplaceTemplate represents a template in the marketplace
type MarketplaceTemplate struct {
	ID               string          `db:"id" json:"id"`
	Name             string          `db:"name" json:"name"`
	Description      string          `db:"description" json:"description"`
	Category         string          `db:"category" json:"category"`
	Definition       json.RawMessage `db:"definition" json:"definition"`
	Tags             pq.StringArray  `db:"tags" json:"tags"`
	AuthorID         string          `db:"author_id" json:"author_id"`
	AuthorName       string          `db:"author_name" json:"author_name"`
	Version          string          `db:"version" json:"version"`
	DownloadCount    int             `db:"download_count" json:"download_count"`
	AverageRating    float64         `db:"average_rating" json:"average_rating"`
	TotalRatings     int             `db:"total_ratings" json:"total_ratings"`
	Rating1Count     int             `db:"rating_1_count" json:"rating_1_count"`
	Rating2Count     int             `db:"rating_2_count" json:"rating_2_count"`
	Rating3Count     int             `db:"rating_3_count" json:"rating_3_count"`
	Rating4Count     int             `db:"rating_4_count" json:"rating_4_count"`
	Rating5Count     int             `db:"rating_5_count" json:"rating_5_count"`
	IsVerified       bool            `db:"is_verified" json:"is_verified"`
	IsFeatured       bool            `db:"is_featured" json:"is_featured"`
	FeaturedAt       *time.Time      `db:"featured_at" json:"featured_at,omitempty"`
	FeaturedBy       *string         `db:"featured_by" json:"featured_by,omitempty"`
	SourceTenantID   *string         `db:"source_tenant_id" json:"source_tenant_id,omitempty"`
	SourceTemplateID *string         `db:"source_template_id" json:"source_template_id,omitempty"`
	PublishedAt      time.Time       `db:"published_at" json:"published_at"`
	UpdatedAt        time.Time       `db:"updated_at" json:"updated_at"`
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
	ID           string     `db:"id" json:"id"`
	TemplateID   string     `db:"template_id" json:"template_id"`
	TenantID     string     `db:"tenant_id" json:"tenant_id"`
	UserID       string     `db:"user_id" json:"user_id"`
	UserName     string     `db:"user_name" json:"user_name"`
	Rating       int        `db:"rating" json:"rating"`
	Comment      string     `db:"comment" json:"comment"`
	HelpfulCount int        `db:"helpful_count" json:"helpful_count"`
	IsHidden     bool       `db:"is_hidden" json:"is_hidden"`
	HiddenReason *string    `db:"hidden_reason" json:"hidden_reason,omitempty"`
	HiddenAt     *time.Time `db:"hidden_at" json:"hidden_at,omitempty"`
	HiddenBy     *string    `db:"hidden_by" json:"hidden_by,omitempty"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
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

// ReviewHelpfulVote tracks helpful votes for reviews
type ReviewHelpfulVote struct {
	ID       string    `db:"id" json:"id"`
	ReviewID string    `db:"review_id" json:"review_id"`
	TenantID string    `db:"tenant_id" json:"tenant_id"`
	UserID   string    `db:"user_id" json:"user_id"`
	VotedAt  time.Time `db:"voted_at" json:"voted_at"`
}

// ReviewReport represents a report of an inappropriate review
type ReviewReport struct {
	ID               string     `db:"id" json:"id"`
	ReviewID         string     `db:"review_id" json:"review_id"`
	ReporterTenantID string     `db:"reporter_tenant_id" json:"reporter_tenant_id"`
	ReporterUserID   string     `db:"reporter_user_id" json:"reporter_user_id"`
	Reason           string     `db:"reason" json:"reason"`
	Details          string     `db:"details" json:"details"`
	Status           string     `db:"status" json:"status"`
	ResolvedAt       *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
	ResolvedBy       *string    `db:"resolved_by" json:"resolved_by,omitempty"`
	ResolutionNotes  *string    `db:"resolution_notes" json:"resolution_notes,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
}

// RatingDistribution represents the distribution of ratings for a template
type RatingDistribution struct {
	Rating1Count    int     `json:"rating_1_count"`
	Rating2Count    int     `json:"rating_2_count"`
	Rating3Count    int     `json:"rating_3_count"`
	Rating4Count    int     `json:"rating_4_count"`
	Rating5Count    int     `json:"rating_5_count"`
	TotalRatings    int     `json:"total_ratings"`
	AverageRating   float64 `json:"average_rating"`
	Rating1Percent  float64 `json:"rating_1_percent"`
	Rating2Percent  float64 `json:"rating_2_percent"`
	Rating3Percent  float64 `json:"rating_3_percent"`
	Rating4Percent  float64 `json:"rating_4_percent"`
	Rating5Percent  float64 `json:"rating_5_percent"`
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

// ReportReviewInput represents input for reporting a review
type ReportReviewInput struct {
	Reason  string `json:"reason" validate:"required,oneof=spam inappropriate offensive misleading other"`
	Details string `json:"details,omitempty" validate:"max=1000"`
}

// ReviewSortOption represents sorting options for reviews
type ReviewSortOption string

const (
	ReviewSortRecent   ReviewSortOption = "recent"
	ReviewSortHelpful  ReviewSortOption = "helpful"
	ReviewSortRatingH  ReviewSortOption = "rating_high"
	ReviewSortRatingL  ReviewSortOption = "rating_low"
)

// ReviewReportReason represents the reason for reporting a review
type ReviewReportReason string

const (
	ReportReasonSpam         ReviewReportReason = "spam"
	ReportReasonInappropriate ReviewReportReason = "inappropriate"
	ReportReasonOffensive     ReviewReportReason = "offensive"
	ReportReasonMisleading    ReviewReportReason = "misleading"
	ReportReasonOther         ReviewReportReason = "other"
)

// ReviewReportStatus represents the status of a review report
type ReviewReportStatus string

const (
	ReportStatusPending  ReviewReportStatus = "pending"
	ReportStatusReviewed ReviewReportStatus = "reviewed"
	ReportStatusActioned ReviewReportStatus = "actioned"
	ReportStatusDismissed ReviewReportStatus = "dismissed"
)

// SearchFilter represents search filters for marketplace templates
type SearchFilter struct {
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	SearchQuery string   `json:"search_query,omitempty"`
	MinRating   *float64 `json:"min_rating,omitempty"`
	IsVerified  *bool    `json:"is_verified,omitempty"`
	SortBy      string   `json:"sort_by,omitempty"` // popular, recent, rating
	Page        int      `json:"page,omitempty"`
	Limit       int      `json:"limit,omitempty"`
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

// Validate validates the report review input
func (i ReportReviewInput) Validate() error {
	if i.Reason == "" {
		return errors.New("reason is required")
	}

	validReasons := map[string]bool{
		string(ReportReasonSpam):         true,
		string(ReportReasonInappropriate): true,
		string(ReportReasonOffensive):     true,
		string(ReportReasonMisleading):    true,
		string(ReportReasonOther):         true,
	}

	if !validReasons[i.Reason] {
		return errors.New("reason must be one of: spam, inappropriate, offensive, misleading, other")
	}

	if len(i.Details) > 1000 {
		return errors.New("details must be 1000 characters or less")
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
