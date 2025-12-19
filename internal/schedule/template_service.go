package schedule

import (
	"context"
	"fmt"
)

// TemplateService handles business logic for schedule templates
type TemplateService struct {
	templateRepo TemplateRepository
	scheduleRepo Repository
}

// NewTemplateService creates a new template service
func NewTemplateService(templateRepo TemplateRepository, scheduleRepo Repository) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		scheduleRepo: scheduleRepo,
	}
}

// ListTemplates retrieves schedule templates with optional filters
func (s *TemplateService) ListTemplates(ctx context.Context, filter ScheduleTemplateFilter) ([]*ScheduleTemplate, error) {
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	return s.templateRepo.List(ctx, filter)
}

// GetTemplate retrieves a single schedule template by ID
func (s *TemplateService) GetTemplate(ctx context.Context, id string) (*ScheduleTemplate, error) {
	if id == "" {
		return nil, fmt.Errorf("template id is required")
	}

	return s.templateRepo.GetByID(ctx, id)
}

// ApplyTemplate applies a template to create a schedule
func (s *TemplateService) ApplyTemplate(
	ctx context.Context,
	tenantID, userID, templateID string,
	input ApplyTemplateInput,
) (*Schedule, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	return ApplyTemplate(ctx, template, input, s.scheduleRepo, tenantID, userID)
}
