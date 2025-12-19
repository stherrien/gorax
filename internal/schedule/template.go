package schedule

import (
	"errors"
	"time"

	"github.com/lib/pq"
)

// ScheduleTemplate represents a pre-configured schedule pattern
type ScheduleTemplate struct {
	ID             string         `db:"id" json:"id"`
	Name           string         `db:"name" json:"name"`
	Description    string         `db:"description" json:"description"`
	Category       string         `db:"category" json:"category"`
	CronExpression string         `db:"cron_expression" json:"cron_expression"`
	Timezone       string         `db:"timezone" json:"timezone"`
	Tags           pq.StringArray `db:"tags" json:"tags"`
	IsSystem       bool           `db:"is_system" json:"is_system"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
}

// ScheduleTemplateFilter represents filters for listing templates
type ScheduleTemplateFilter struct {
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	IsSystem    *bool    `json:"is_system,omitempty"`
	SearchQuery string   `json:"search_query,omitempty"`
}

// ApplyTemplateInput represents input for applying a template to create a schedule
type ApplyTemplateInput struct {
	WorkflowID string  `json:"workflow_id" validate:"required"`
	Name       *string `json:"name,omitempty"`
	Timezone   *string `json:"timezone,omitempty"`
}

// ScheduleTemplateCategory represents template categories
type ScheduleTemplateCategory string

const (
	CategoryFrequent   ScheduleTemplateCategory = "frequent"
	CategoryDaily      ScheduleTemplateCategory = "daily"
	CategoryWeekly     ScheduleTemplateCategory = "weekly"
	CategoryMonthly    ScheduleTemplateCategory = "monthly"
	CategoryBusiness   ScheduleTemplateCategory = "business"
	CategoryCompliance ScheduleTemplateCategory = "compliance"
	CategorySecurity   ScheduleTemplateCategory = "security"
	CategorySync       ScheduleTemplateCategory = "sync"
	CategoryMonitoring ScheduleTemplateCategory = "monitoring"
)

// Validate validates the template filter
func (f ScheduleTemplateFilter) Validate() error {
	if f.Category != "" && len(f.Category) > 100 {
		return errors.New("category must be 100 characters or less")
	}
	return nil
}

// Validate validates the apply template input
func (i ApplyTemplateInput) Validate() error {
	if i.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}

	if i.Name != nil && len(*i.Name) > 255 {
		return errors.New("name must be 255 characters or less")
	}

	if i.Timezone != nil && len(*i.Timezone) > 100 {
		return errors.New("timezone must be 100 characters or less")
	}

	return nil
}
