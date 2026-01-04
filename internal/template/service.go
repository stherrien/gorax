package template

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
)

// Service handles template business logic
type Service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new template service
func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateTemplate creates a new template
func (s *Service) CreateTemplate(ctx context.Context, tenantID, userID string, input CreateTemplateInput) (*Template, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if err := s.validateDefinition(input.Definition); err != nil {
		return nil, fmt.Errorf("invalid definition: %w", err)
	}

	template := &Template{
		TenantID:    &tenantID,
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Definition:  input.Definition,
		Tags:        input.Tags,
		IsPublic:    input.IsPublic,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(ctx, tenantID, template); err != nil {
		s.logger.Error("failed to create template",
			"error", err,
			"tenant_id", tenantID,
			"name", input.Name)
		return nil, fmt.Errorf("create template: %w", err)
	}

	s.logger.Info("template created",
		"template_id", template.ID,
		"tenant_id", tenantID,
		"name", template.Name)

	return template, nil
}

// GetTemplate retrieves a template by ID
func (s *Service) GetTemplate(ctx context.Context, tenantID, id string) (*Template, error) {
	template, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		s.logger.Error("failed to get template",
			"error", err,
			"tenant_id", tenantID,
			"template_id", id)
		return nil, fmt.Errorf("get template: %w", err)
	}

	return template, nil
}

// ListTemplates retrieves templates with optional filters
func (s *Service) ListTemplates(ctx context.Context, tenantID string, filter TemplateFilter) ([]*Template, error) {
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	templates, err := s.repo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Error("failed to list templates",
			"error", err,
			"tenant_id", tenantID)
		return nil, fmt.Errorf("list templates: %w", err)
	}

	return templates, nil
}

// UpdateTemplate updates an existing template
func (s *Service) UpdateTemplate(ctx context.Context, tenantID, id string, input UpdateTemplateInput) error {
	if input.Definition != nil {
		if err := s.validateDefinition(input.Definition); err != nil {
			return fmt.Errorf("invalid definition: %w", err)
		}
	}

	if err := s.repo.Update(ctx, tenantID, id, input); err != nil {
		s.logger.Error("failed to update template",
			"error", err,
			"tenant_id", tenantID,
			"template_id", id)
		return fmt.Errorf("update template: %w", err)
	}

	s.logger.Info("template updated",
		"template_id", id,
		"tenant_id", tenantID)

	return nil
}

// DeleteTemplate deletes a template
func (s *Service) DeleteTemplate(ctx context.Context, tenantID, id string) error {
	if err := s.repo.Delete(ctx, tenantID, id); err != nil {
		s.logger.Error("failed to delete template",
			"error", err,
			"tenant_id", tenantID,
			"template_id", id)
		return fmt.Errorf("delete template: %w", err)
	}

	s.logger.Info("template deleted",
		"template_id", id,
		"tenant_id", tenantID)

	return nil
}

// CreateTemplateFromWorkflowInput represents input for creating a template from a workflow
type CreateTemplateFromWorkflowInput struct {
	WorkflowID  string          `json:"workflow_id" validate:"required"`
	Name        string          `json:"name" validate:"required"`
	Description string          `json:"description"`
	Category    string          `json:"category" validate:"required"`
	Definition  json.RawMessage `json:"definition" validate:"required"`
	Tags        []string        `json:"tags"`
	IsPublic    bool            `json:"is_public"`
}

// CreateFromWorkflow creates a template from an existing workflow
func (s *Service) CreateFromWorkflow(ctx context.Context, tenantID, userID string, input CreateTemplateFromWorkflowInput) (*Template, error) {
	if input.WorkflowID == "" {
		return nil, errors.New("workflow_id is required")
	}

	createInput := CreateTemplateInput{
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Definition:  input.Definition,
		Tags:        input.Tags,
		IsPublic:    input.IsPublic,
	}

	template, err := s.CreateTemplate(ctx, tenantID, userID, createInput)
	if err != nil {
		return nil, fmt.Errorf("create from workflow: %w", err)
	}

	s.logger.Info("template created from workflow",
		"template_id", template.ID,
		"workflow_id", input.WorkflowID,
		"tenant_id", tenantID)

	return template, nil
}

// InstantiateTemplateInput represents input for instantiating a template
type InstantiateTemplateInput struct {
	WorkflowName string `json:"workflow_name" validate:"required"`
}

// InstantiateTemplateResult represents the result of template instantiation
type InstantiateTemplateResult struct {
	WorkflowName string          `json:"workflow_name"`
	Definition   json.RawMessage `json:"definition"`
}

// InstantiateTemplate creates a workflow definition from a template
func (s *Service) InstantiateTemplate(ctx context.Context, tenantID, templateID string, input InstantiateTemplateInput) (*InstantiateTemplateResult, error) {
	if input.WorkflowName == "" {
		return nil, errors.New("workflow_name is required")
	}

	template, err := s.GetTemplate(ctx, tenantID, templateID)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}

	// Increment usage count
	if err := s.repo.IncrementUsageCount(ctx, templateID); err != nil {
		s.logger.Warn("failed to increment usage count",
			"error", err,
			"template_id", templateID)
		// Don't fail the operation if usage count increment fails
	}

	result := &InstantiateTemplateResult{
		WorkflowName: input.WorkflowName,
		Definition:   template.Definition,
	}

	s.logger.Info("template instantiated",
		"template_id", templateID,
		"workflow_name", input.WorkflowName,
		"tenant_id", tenantID)

	return result, nil
}

// validateDefinition validates the workflow definition structure
func (s *Service) validateDefinition(definition json.RawMessage) error {
	if len(definition) == 0 {
		return errors.New("definition is required")
	}

	var def map[string]interface{}
	if err := json.Unmarshal(definition, &def); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if _, ok := def["nodes"]; !ok {
		return errors.New("definition must contain 'nodes' field")
	}

	if _, ok := def["edges"]; !ok {
		return errors.New("definition must contain 'edges' field")
	}

	return nil
}
