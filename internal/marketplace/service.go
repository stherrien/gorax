package marketplace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
)

// WorkflowService defines the interface for workflow operations
type WorkflowService interface {
	CreateFromTemplate(ctx context.Context, tenantID, userID, templateID, workflowName string, definition json.RawMessage) (string, error)
}

// Service handles marketplace business logic
type Service struct {
	repo            Repository
	workflowService WorkflowService
	logger          *slog.Logger
}

// NewService creates a new marketplace service
func NewService(repo Repository, workflowService WorkflowService, logger *slog.Logger) *Service {
	return &Service{
		repo:            repo,
		workflowService: workflowService,
		logger:          logger,
	}
}

// PublishTemplate publishes a new template to the marketplace
func (s *Service) PublishTemplate(ctx context.Context, userID, userName string, input PublishTemplateInput) (*MarketplaceTemplate, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if err := s.validateDefinition(input.Definition); err != nil {
		return nil, fmt.Errorf("invalid definition: %w", err)
	}

	template := &MarketplaceTemplate{
		Name:             input.Name,
		Description:      input.Description,
		Category:         input.Category,
		Definition:       input.Definition,
		Tags:             input.Tags,
		AuthorID:         userID,
		AuthorName:       userName,
		Version:          input.Version,
		SourceTemplateID: input.SourceTemplateID,
		IsVerified:       false,
		DownloadCount:    0,
		AverageRating:    0,
		TotalRatings:     0,
	}

	if err := s.repo.Publish(ctx, template); err != nil {
		s.logger.Error("failed to publish template",
			"error", err,
			"author_id", userID,
			"name", input.Name)
		return nil, fmt.Errorf("publish template: %w", err)
	}

	s.logger.Info("template published",
		"template_id", template.ID,
		"author_id", userID,
		"name", template.Name)

	return template, nil
}

// GetTemplate retrieves a template by ID
func (s *Service) GetTemplate(ctx context.Context, id string) (*MarketplaceTemplate, error) {
	template, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get template",
			"error", err,
			"template_id", id)
		return nil, fmt.Errorf("get template: %w", err)
	}

	return template, nil
}

// SearchTemplates searches for templates with filters
func (s *Service) SearchTemplates(ctx context.Context, filter SearchFilter) ([]*MarketplaceTemplate, error) {
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	templates, err := s.repo.Search(ctx, filter)
	if err != nil {
		s.logger.Error("failed to search templates",
			"error", err,
			"filter", filter)
		return nil, fmt.Errorf("search templates: %w", err)
	}

	return templates, nil
}

// GetTrending retrieves trending templates
func (s *Service) GetTrending(ctx context.Context, limit int) ([]*MarketplaceTemplate, error) {
	if limit <= 0 {
		limit = 10
	}

	templates, err := s.repo.GetTrending(ctx, 7, limit)
	if err != nil {
		s.logger.Error("failed to get trending templates", "error", err)
		return nil, fmt.Errorf("get trending: %w", err)
	}

	return templates, nil
}

// GetPopular retrieves popular templates
func (s *Service) GetPopular(ctx context.Context, limit int) ([]*MarketplaceTemplate, error) {
	if limit <= 0 {
		limit = 10
	}

	templates, err := s.repo.GetPopular(ctx, limit)
	if err != nil {
		s.logger.Error("failed to get popular templates", "error", err)
		return nil, fmt.Errorf("get popular: %w", err)
	}

	return templates, nil
}

// InstallTemplate installs a template as a workflow in the tenant
func (s *Service) InstallTemplate(ctx context.Context, tenantID, userID, templateID string, input InstallTemplateInput) (*InstallTemplateResult, error) {
	if input.WorkflowName == "" {
		return nil, errors.New("workflow_name is required")
	}

	installation, err := s.repo.GetInstallation(ctx, tenantID, templateID)
	if err == nil && installation != nil {
		return nil, fmt.Errorf("template already installed")
	}

	template, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}

	workflowID, err := s.workflowService.CreateFromTemplate(
		ctx,
		tenantID,
		userID,
		templateID,
		input.WorkflowName,
		template.Definition,
	)
	if err != nil {
		s.logger.Error("failed to create workflow from template",
			"error", err,
			"template_id", templateID,
			"tenant_id", tenantID)
		return nil, fmt.Errorf("create workflow: %w", err)
	}

	if err := s.repo.IncrementDownloadCount(ctx, templateID); err != nil {
		s.logger.Warn("failed to increment download count",
			"error", err,
			"template_id", templateID)
	}

	installation = &TemplateInstallation{
		TemplateID:       templateID,
		TenantID:         tenantID,
		UserID:           userID,
		WorkflowID:       workflowID,
		InstalledVersion: template.Version,
	}

	if err := s.repo.CreateInstallation(ctx, installation); err != nil {
		s.logger.Warn("failed to record installation",
			"error", err,
			"template_id", templateID,
			"tenant_id", tenantID)
	}

	s.logger.Info("template installed",
		"template_id", templateID,
		"workflow_id", workflowID,
		"tenant_id", tenantID)

	return &InstallTemplateResult{
		WorkflowID:   workflowID,
		WorkflowName: input.WorkflowName,
		Definition:   template.Definition,
	}, nil
}

// RateTemplate adds or updates a rating for a template
func (s *Service) RateTemplate(ctx context.Context, tenantID, userID, userName, templateID string, input RateTemplateInput) (*TemplateReview, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	existingReview, err := s.repo.GetUserReview(ctx, tenantID, templateID)
	if err == nil && existingReview != nil {
		if err := s.repo.UpdateReview(ctx, tenantID, existingReview.ID, input.Rating, input.Comment); err != nil {
			s.logger.Error("failed to update review",
				"error", err,
				"template_id", templateID,
				"tenant_id", tenantID)
			return nil, fmt.Errorf("update review: %w", err)
		}
	} else {
		review := &TemplateReview{
			TemplateID: templateID,
			TenantID:   tenantID,
			UserID:     userID,
			UserName:   userName,
			Rating:     input.Rating,
			Comment:    input.Comment,
		}

		if err := s.repo.CreateReview(ctx, review); err != nil {
			s.logger.Error("failed to create review",
				"error", err,
				"template_id", templateID,
				"tenant_id", tenantID)
			return nil, fmt.Errorf("create review: %w", err)
		}
	}

	if err := s.repo.UpdateTemplateRating(ctx, templateID); err != nil {
		s.logger.Warn("failed to update template rating",
			"error", err,
			"template_id", templateID)
	}

	review, err := s.repo.GetUserReview(ctx, tenantID, templateID)
	if err != nil {
		return nil, fmt.Errorf("get review: %w", err)
	}

	s.logger.Info("template rated",
		"template_id", templateID,
		"tenant_id", tenantID,
		"rating", input.Rating)

	return review, nil
}

// GetReviews retrieves reviews for a template
func (s *Service) GetReviews(ctx context.Context, templateID string, sortBy ReviewSortOption, limit, offset int) ([]*TemplateReview, error) {
	if limit <= 0 {
		limit = 10
	}

	if sortBy == "" {
		sortBy = ReviewSortRecent
	}

	reviews, err := s.repo.GetReviews(ctx, templateID, sortBy, limit, offset)
	if err != nil {
		s.logger.Error("failed to get reviews",
			"error", err,
			"template_id", templateID)
		return nil, fmt.Errorf("get reviews: %w", err)
	}

	return reviews, nil
}

// DeleteReview deletes a review
func (s *Service) DeleteReview(ctx context.Context, tenantID, templateID, reviewID string) error {
	if err := s.repo.DeleteReview(ctx, tenantID, reviewID); err != nil {
		s.logger.Error("failed to delete review",
			"error", err,
			"review_id", reviewID,
			"tenant_id", tenantID)
		return fmt.Errorf("delete review: %w", err)
	}

	if err := s.repo.UpdateTemplateRating(ctx, templateID); err != nil {
		s.logger.Warn("failed to update template rating after delete",
			"error", err,
			"template_id", templateID)
	}

	s.logger.Info("review deleted",
		"review_id", reviewID,
		"template_id", templateID,
		"tenant_id", tenantID)

	return nil
}

// GetCategories returns all available categories
func (s *Service) GetCategories() []string {
	return GetCategories()
}

// VoteReviewHelpful marks a review as helpful
func (s *Service) VoteReviewHelpful(ctx context.Context, tenantID, userID, reviewID string) error {
	vote := &ReviewHelpfulVote{
		ReviewID: reviewID,
		TenantID: tenantID,
		UserID:   userID,
	}

	if err := s.repo.VoteReviewHelpful(ctx, vote); err != nil {
		s.logger.Error("failed to vote review helpful",
			"error", err,
			"review_id", reviewID,
			"tenant_id", tenantID)
		return fmt.Errorf("vote review helpful: %w", err)
	}

	s.logger.Info("review voted helpful",
		"review_id", reviewID,
		"tenant_id", tenantID)

	return nil
}

// UnvoteReviewHelpful removes a helpful vote from a review
func (s *Service) UnvoteReviewHelpful(ctx context.Context, tenantID, userID, reviewID string) error {
	if err := s.repo.UnvoteReviewHelpful(ctx, tenantID, userID, reviewID); err != nil {
		s.logger.Error("failed to unvote review helpful",
			"error", err,
			"review_id", reviewID,
			"tenant_id", tenantID)
		return fmt.Errorf("unvote review helpful: %w", err)
	}

	s.logger.Info("review unvoted helpful",
		"review_id", reviewID,
		"tenant_id", tenantID)

	return nil
}

// HasVotedHelpful checks if a user has voted a review as helpful
func (s *Service) HasVotedHelpful(ctx context.Context, tenantID, userID, reviewID string) (bool, error) {
	return s.repo.HasVotedHelpful(ctx, tenantID, userID, reviewID)
}

// ReportReview reports a review for moderation
func (s *Service) ReportReview(ctx context.Context, tenantID, userID, reviewID string, input ReportReviewInput) error {
	if err := input.Validate(); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	report := &ReviewReport{
		ReviewID:         reviewID,
		ReporterTenantID: tenantID,
		ReporterUserID:   userID,
		Reason:           input.Reason,
		Details:          input.Details,
	}

	if err := s.repo.CreateReviewReport(ctx, report); err != nil {
		s.logger.Error("failed to report review",
			"error", err,
			"review_id", reviewID,
			"tenant_id", tenantID)
		return fmt.Errorf("report review: %w", err)
	}

	s.logger.Info("review reported",
		"review_id", reviewID,
		"tenant_id", tenantID,
		"reason", input.Reason)

	return nil
}

// GetReviewReports retrieves review reports (admin only)
func (s *Service) GetReviewReports(ctx context.Context, status string, limit, offset int) ([]*ReviewReport, error) {
	if limit <= 0 {
		limit = 20
	}

	reports, err := s.repo.GetReviewReports(ctx, status, limit, offset)
	if err != nil {
		s.logger.Error("failed to get review reports", "error", err)
		return nil, fmt.Errorf("get review reports: %w", err)
	}

	return reports, nil
}

// ResolveReviewReport resolves a review report (admin only)
func (s *Service) ResolveReviewReport(ctx context.Context, reportID, status, resolvedBy string, notes *string) error {
	validStatuses := map[string]bool{
		string(ReportStatusReviewed):  true,
		string(ReportStatusActioned):  true,
		string(ReportStatusDismissed): true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	if err := s.repo.UpdateReviewReportStatus(ctx, reportID, status, resolvedBy, notes); err != nil {
		s.logger.Error("failed to resolve report",
			"error", err,
			"report_id", reportID)
		return fmt.Errorf("resolve report: %w", err)
	}

	s.logger.Info("review report resolved",
		"report_id", reportID,
		"status", status,
		"resolved_by", resolvedBy)

	return nil
}

// HideReview hides a review (admin/moderator only)
func (s *Service) HideReview(ctx context.Context, reviewID, reason, hiddenBy string) error {
	if err := s.repo.HideReview(ctx, reviewID, reason, hiddenBy); err != nil {
		s.logger.Error("failed to hide review",
			"error", err,
			"review_id", reviewID)
		return fmt.Errorf("hide review: %w", err)
	}

	s.logger.Info("review hidden",
		"review_id", reviewID,
		"hidden_by", hiddenBy,
		"reason", reason)

	return nil
}

// UnhideReview unhides a review (admin/moderator only)
func (s *Service) UnhideReview(ctx context.Context, reviewID string) error {
	if err := s.repo.UnhideReview(ctx, reviewID); err != nil {
		s.logger.Error("failed to unhide review",
			"error", err,
			"review_id", reviewID)
		return fmt.Errorf("unhide review: %w", err)
	}

	s.logger.Info("review unhidden", "review_id", reviewID)

	return nil
}

// GetRatingDistribution retrieves the rating distribution for a template
func (s *Service) GetRatingDistribution(ctx context.Context, templateID string) (*RatingDistribution, error) {
	dist, err := s.repo.GetRatingDistribution(ctx, templateID)
	if err != nil {
		s.logger.Error("failed to get rating distribution",
			"error", err,
			"template_id", templateID)
		return nil, fmt.Errorf("get rating distribution: %w", err)
	}

	return dist, nil
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
