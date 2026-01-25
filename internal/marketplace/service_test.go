package marketplace

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Publish(ctx context.Context, template *MarketplaceTemplate) error {
	args := m.Called(ctx, template)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*MarketplaceTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MarketplaceTemplate), args.Error(1)
}

func (m *MockRepository) Search(ctx context.Context, filter SearchFilter) ([]*MarketplaceTemplate, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*MarketplaceTemplate), args.Error(1)
}

func (m *MockRepository) GetPopular(ctx context.Context, limit int) ([]*MarketplaceTemplate, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*MarketplaceTemplate), args.Error(1)
}

func (m *MockRepository) GetTrending(ctx context.Context, days, limit int) ([]*MarketplaceTemplate, error) {
	args := m.Called(ctx, days, limit)
	return args.Get(0).([]*MarketplaceTemplate), args.Error(1)
}

func (m *MockRepository) GetByAuthor(ctx context.Context, authorID string) ([]*MarketplaceTemplate, error) {
	args := m.Called(ctx, authorID)
	return args.Get(0).([]*MarketplaceTemplate), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, id string, input UpdateTemplateInput) error {
	args := m.Called(ctx, id, input)
	return args.Error(0)
}

func (m *MockRepository) IncrementDownloadCount(ctx context.Context, templateID string) error {
	args := m.Called(ctx, templateID)
	return args.Error(0)
}

func (m *MockRepository) CreateInstallation(ctx context.Context, installation *TemplateInstallation) error {
	args := m.Called(ctx, installation)
	return args.Error(0)
}

func (m *MockRepository) GetInstallation(ctx context.Context, tenantID, templateID string) (*TemplateInstallation, error) {
	args := m.Called(ctx, tenantID, templateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TemplateInstallation), args.Error(1)
}

func (m *MockRepository) CreateReview(ctx context.Context, review *TemplateReview) error {
	args := m.Called(ctx, review)
	return args.Error(0)
}

func (m *MockRepository) UpdateReview(ctx context.Context, tenantID, reviewID string, rating int, comment string) error {
	args := m.Called(ctx, tenantID, reviewID, rating, comment)
	return args.Error(0)
}

func (m *MockRepository) DeleteReview(ctx context.Context, tenantID, reviewID string) error {
	args := m.Called(ctx, tenantID, reviewID)
	return args.Error(0)
}

func (m *MockRepository) GetReviews(ctx context.Context, templateID string, sortBy ReviewSortOption, limit, offset int) ([]*TemplateReview, error) {
	args := m.Called(ctx, templateID, sortBy, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*TemplateReview), args.Error(1)
}

func (m *MockRepository) GetUserReview(ctx context.Context, tenantID, templateID string) (*TemplateReview, error) {
	args := m.Called(ctx, tenantID, templateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TemplateReview), args.Error(1)
}

func (m *MockRepository) UpdateTemplateRating(ctx context.Context, templateID string) error {
	args := m.Called(ctx, templateID)
	return args.Error(0)
}

func (m *MockRepository) VoteReviewHelpful(ctx context.Context, vote *ReviewHelpfulVote) error {
	args := m.Called(ctx, vote)
	return args.Error(0)
}

func (m *MockRepository) UnvoteReviewHelpful(ctx context.Context, tenantID, userID, reviewID string) error {
	args := m.Called(ctx, tenantID, userID, reviewID)
	return args.Error(0)
}

func (m *MockRepository) HasVotedHelpful(ctx context.Context, tenantID, userID, reviewID string) (bool, error) {
	args := m.Called(ctx, tenantID, userID, reviewID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) CreateReviewReport(ctx context.Context, report *ReviewReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

func (m *MockRepository) GetReviewReports(ctx context.Context, status string, limit, offset int) ([]*ReviewReport, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ReviewReport), args.Error(1)
}

func (m *MockRepository) UpdateReviewReportStatus(ctx context.Context, reportID, status, resolvedBy string, notes *string) error {
	args := m.Called(ctx, reportID, status, resolvedBy, notes)
	return args.Error(0)
}

func (m *MockRepository) HideReview(ctx context.Context, reviewID, reason, hiddenBy string) error {
	args := m.Called(ctx, reviewID, reason, hiddenBy)
	return args.Error(0)
}

func (m *MockRepository) UnhideReview(ctx context.Context, reviewID string) error {
	args := m.Called(ctx, reviewID)
	return args.Error(0)
}

func (m *MockRepository) GetRatingDistribution(ctx context.Context, templateID string) (*RatingDistribution, error) {
	args := m.Called(ctx, templateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RatingDistribution), args.Error(1)
}

// MockWorkflowService is a mock implementation of WorkflowService
type MockWorkflowService struct {
	mock.Mock
}

func (m *MockWorkflowService) CreateFromTemplate(ctx context.Context, tenantID, userID, templateID, workflowName string, definition json.RawMessage) (string, error) {
	args := m.Called(ctx, tenantID, userID, templateID, workflowName, definition)
	return args.String(0), args.Error(1)
}

func TestService_PublishTemplate(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
	input := PublishTemplateInput{
		Name:        "Test Template",
		Description: "This is a test template description that is long enough",
		Category:    "automation",
		Definition:  definition,
		Tags:        []string{"test", "automation"},
		Version:     "1.0.0",
	}

	repo.On("Publish", ctx, mock.AnythingOfType("*marketplace.MarketplaceTemplate")).Return(nil)

	template, err := service.PublishTemplate(ctx, "user-1", "Test User", input)
	require.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, input.Name, template.Name)
	assert.Equal(t, "user-1", template.AuthorID)
	repo.AssertExpectations(t)
}

func TestPublishTemplate_InvalidInput(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	input := PublishTemplateInput{
		Name:        "",
		Description: "Description",
		Category:    "automation",
	}

	_, err := service.PublishTemplate(ctx, "user-1", "Test User", input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input")
}

func TestGetTemplate(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	expectedTemplate := &MarketplaceTemplate{
		ID:          "template-1",
		Name:        "Test Template",
		Description: "Description",
	}

	repo.On("GetByID", ctx, "template-1").Return(expectedTemplate, nil)

	template, err := service.GetTemplate(ctx, "template-1")
	require.NoError(t, err)
	assert.Equal(t, expectedTemplate, template)
	repo.AssertExpectations(t)
}

func TestService_SearchTemplates(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	filter := SearchFilter{
		Category: "automation",
		Limit:    10,
	}

	expectedTemplates := []*MarketplaceTemplate{
		{ID: "1", Name: "Template 1"},
		{ID: "2", Name: "Template 2"},
	}

	repo.On("Search", ctx, filter).Return(expectedTemplates, nil)

	templates, err := service.SearchTemplates(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, expectedTemplates, templates)
	repo.AssertExpectations(t)
}

func TestService_GetTrending(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	expectedTemplates := []*MarketplaceTemplate{
		{ID: "1", Name: "Trending 1"},
	}

	repo.On("GetTrending", ctx, 7, 10).Return(expectedTemplates, nil)

	templates, err := service.GetTrending(ctx, 10)
	require.NoError(t, err)
	assert.Equal(t, expectedTemplates, templates)
	repo.AssertExpectations(t)
}

func TestInstallTemplate(t *testing.T) {
	repo := new(MockRepository)
	workflowService := new(MockWorkflowService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, workflowService, logger)
	ctx := context.Background()

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
	template := &MarketplaceTemplate{
		ID:         "template-1",
		Name:       "Test Template",
		Version:    "1.0.0",
		Definition: definition,
	}

	input := InstallTemplateInput{
		WorkflowName: "My Workflow",
	}

	repo.On("GetInstallation", ctx, "tenant-1", "template-1").Return(nil, errors.New("not found"))
	repo.On("GetByID", ctx, "template-1").Return(template, nil)
	workflowService.On("CreateFromTemplate", ctx, "tenant-1", "user-1", "template-1", "My Workflow", definition).Return("workflow-1", nil)
	repo.On("IncrementDownloadCount", ctx, "template-1").Return(nil)
	repo.On("CreateInstallation", ctx, mock.AnythingOfType("*marketplace.TemplateInstallation")).Return(nil)

	result, err := service.InstallTemplate(ctx, "tenant-1", "user-1", "template-1", input)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "workflow-1", result.WorkflowID)
	assert.Equal(t, "My Workflow", result.WorkflowName)
	repo.AssertExpectations(t)
	workflowService.AssertExpectations(t)
}

func TestInstallTemplate_AlreadyInstalled(t *testing.T) {
	repo := new(MockRepository)
	workflowService := new(MockWorkflowService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, workflowService, logger)
	ctx := context.Background()

	installation := &TemplateInstallation{
		ID:         "install-1",
		TemplateID: "template-1",
	}

	input := InstallTemplateInput{
		WorkflowName: "My Workflow",
	}

	repo.On("GetInstallation", ctx, "tenant-1", "template-1").Return(installation, nil)

	_, err := service.InstallTemplate(ctx, "tenant-1", "user-1", "template-1", input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
	repo.AssertExpectations(t)
}

func TestRateTemplate(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	input := RateTemplateInput{
		Rating:  5,
		Comment: "Great template!",
	}

	// First call - check if review exists (returns not found)
	repo.On("GetUserReview", ctx, "tenant-1", "template-1").Return(nil, errors.New("review not found")).Once()
	repo.On("CreateReview", ctx, mock.AnythingOfType("*marketplace.TemplateReview")).Return(nil)
	repo.On("UpdateTemplateRating", ctx, "template-1").Return(nil)
	// Second call - return the created review
	repo.On("GetUserReview", ctx, "tenant-1", "template-1").Return(&TemplateReview{
		ID:         "review-1",
		TemplateID: "template-1",
		TenantID:   "tenant-1",
		UserID:     "user-1",
		UserName:   "Test User",
		Rating:     5,
		Comment:    "Great template!",
	}, nil).Once()

	review, err := service.RateTemplate(ctx, "tenant-1", "user-1", "Test User", "template-1", input)
	require.NoError(t, err)
	assert.NotNil(t, review)
	assert.Equal(t, 5, review.Rating)
	repo.AssertExpectations(t)
}

func TestRateTemplate_UpdateExisting(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	existingReview := &TemplateReview{
		ID:         "review-1",
		TemplateID: "template-1",
		Rating:     3,
	}

	input := RateTemplateInput{
		Rating:  5,
		Comment: "Updated review!",
	}

	// First call - check if review exists (returns existing)
	repo.On("GetUserReview", ctx, "tenant-1", "template-1").Return(existingReview, nil).Once()
	repo.On("UpdateReview", ctx, "tenant-1", "review-1", 5, "Updated review!").Return(nil)
	repo.On("UpdateTemplateRating", ctx, "template-1").Return(nil)
	// Second call - return the updated review
	repo.On("GetUserReview", ctx, "tenant-1", "template-1").Return(&TemplateReview{
		ID:      "review-1",
		Rating:  5,
		Comment: "Updated review!",
	}, nil).Once()

	review, err := service.RateTemplate(ctx, "tenant-1", "user-1", "Test User", "template-1", input)
	require.NoError(t, err)
	assert.NotNil(t, review)
	assert.Equal(t, 5, review.Rating)
	repo.AssertExpectations(t)
}

func TestService_GetReviews(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	expectedReviews := []*TemplateReview{
		{ID: "1", Rating: 5},
		{ID: "2", Rating: 4},
	}

	repo.On("GetReviews", ctx, "template-1", ReviewSortRecent, 10, 0).Return(expectedReviews, nil)

	reviews, err := service.GetReviews(ctx, "template-1", ReviewSortRecent, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, expectedReviews, reviews)
	repo.AssertExpectations(t)
}

func TestService_DeleteReview(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	repo.On("DeleteReview", ctx, "tenant-1", "review-1").Return(nil)
	repo.On("UpdateTemplateRating", ctx, "template-1").Return(nil)

	err := service.DeleteReview(ctx, "tenant-1", "template-1", "review-1")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_GetCategories(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)

	categories := service.GetCategories()
	assert.NotEmpty(t, categories)
	assert.Contains(t, categories, "security")
	assert.Contains(t, categories, "automation")
}

// ==================== Error Scenario Tests ====================

func TestService_PublishTemplate_RepositoryError(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	input := PublishTemplateInput{
		Name:        "Test Template",
		Description: "A valid description with more than ten characters",
		Category:    "security",
		Definition:  json.RawMessage(`{"nodes":[],"edges":[]}`),
		Version:     "1.0.0",
	}

	repo.On("Publish", ctx, mock.AnythingOfType("*marketplace.MarketplaceTemplate")).Return(errors.New("database connection lost"))

	template, err := service.PublishTemplate(ctx, "user-1", "User One", input)
	assert.Error(t, err)
	assert.Nil(t, template)
	assert.Contains(t, err.Error(), "publish template")
	repo.AssertExpectations(t)
}

func TestService_SearchTemplates_RepositoryError(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	filter := SearchFilter{Limit: 10}
	repo.On("Search", ctx, filter).Return(([]*MarketplaceTemplate)(nil), errors.New("search index unavailable"))

	templates, err := service.SearchTemplates(ctx, filter)
	assert.Error(t, err)
	assert.Nil(t, templates)
	assert.Contains(t, err.Error(), "search templates")
	repo.AssertExpectations(t)
}

func TestService_GetTrending_RepositoryError(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	repo.On("GetTrending", ctx, 7, 10).Return(([]*MarketplaceTemplate)(nil), errors.New("database timeout"))

	templates, err := service.GetTrending(ctx, 10)
	assert.Error(t, err)
	assert.Nil(t, templates)
	assert.Contains(t, err.Error(), "get trending")
	repo.AssertExpectations(t)
}

func TestService_GetPopular_RepositoryError(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	repo.On("GetPopular", ctx, 10).Return(([]*MarketplaceTemplate)(nil), errors.New("replica not available"))

	templates, err := service.GetPopular(ctx, 10)
	assert.Error(t, err)
	assert.Nil(t, templates)
	assert.Contains(t, err.Error(), "get popular")
	repo.AssertExpectations(t)
}

func TestService_GetReviews_RepositoryError(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	repo.On("GetReviews", ctx, "template-1", ReviewSortRecent, 10, 0).Return(([]*TemplateReview)(nil), errors.New("query timeout"))

	reviews, err := service.GetReviews(ctx, "template-1", ReviewSortRecent, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, reviews)
	assert.Contains(t, err.Error(), "get reviews")
	repo.AssertExpectations(t)
}

func TestService_DeleteReview_RepositoryError(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	repo.On("DeleteReview", ctx, "tenant-1", "review-1").Return(errors.New("constraint violation"))

	err := service.DeleteReview(ctx, "tenant-1", "template-1", "review-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete review")
	repo.AssertExpectations(t)
}

func TestService_ValidateDefinition_EdgeCases(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	tests := []struct {
		name       string
		definition json.RawMessage
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "empty definition",
			definition: json.RawMessage(``),
			wantErr:    true,
			errMsg:     "definition is required", // From input.Validate()
		},
		{
			name:       "invalid JSON",
			definition: json.RawMessage(`{invalid json}`),
			wantErr:    true,
			errMsg:     "definition must be valid JSON", // From input.Validate()
		},
		{
			name:       "missing nodes field",
			definition: json.RawMessage(`{"edges":[]}`),
			wantErr:    true,
			errMsg:     "definition must contain 'nodes' field", // From validateDefinition()
		},
		{
			name:       "missing edges field",
			definition: json.RawMessage(`{"nodes":[]}`),
			wantErr:    true,
			errMsg:     "definition must contain 'edges' field", // From validateDefinition()
		},
		{
			name:       "valid minimal definition",
			definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := PublishTemplateInput{
				Name:        "Test Template",
				Description: "A valid description with more than ten characters",
				Category:    "security",
				Definition:  tt.definition,
				Version:     "1.0.0",
			}

			if !tt.wantErr {
				repo.On("Publish", ctx, mock.AnythingOfType("*marketplace.MarketplaceTemplate")).Return(nil).Once()
			}

			_, err := service.PublishTemplate(ctx, "user-1", "User", input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_InstallTemplate_WorkflowServiceError(t *testing.T) {
	repo := new(MockRepository)
	workflowSvc := new(MockWorkflowService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, workflowSvc, logger)
	ctx := context.Background()

	template := &MarketplaceTemplate{
		ID:         "template-1",
		Name:       "Test Template",
		Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
		Version:    "1.0.0",
	}

	repo.On("GetInstallation", ctx, "tenant-1", "template-1").Return(nil, errors.New("not found"))
	repo.On("GetByID", ctx, "template-1").Return(template, nil)
	workflowSvc.On("CreateFromTemplate", ctx, "tenant-1", "user-1", "template-1", "My Workflow", template.Definition).Return("", errors.New("quota exceeded"))

	result, err := service.InstallTemplate(ctx, "tenant-1", "user-1", "template-1", InstallTemplateInput{WorkflowName: "My Workflow"})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "create workflow")
	repo.AssertExpectations(t)
	workflowSvc.AssertExpectations(t)
}

func TestService_RateTemplate_CreateReviewError(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	// No existing review
	repo.On("GetUserReview", ctx, "tenant-1", "template-1").Return(nil, errors.New("not found")).Once()
	// Create review fails
	repo.On("CreateReview", ctx, mock.AnythingOfType("*marketplace.TemplateReview")).Return(errors.New("duplicate key"))

	input := RateTemplateInput{Rating: 5, Comment: "Great!"}
	review, err := service.RateTemplate(ctx, "tenant-1", "user-1", "User One", "template-1", input)
	assert.Error(t, err)
	assert.Nil(t, review)
	assert.Contains(t, err.Error(), "create review")
	repo.AssertExpectations(t)
}

func TestService_RateTemplate_InvalidRating(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	tests := []struct {
		name    string
		rating  int
		wantErr bool
	}{
		{"rating too low", 0, true},
		{"rating negative", -1, true},
		{"rating too high", 6, true},
		{"rating valid min", 1, false},
		{"rating valid max", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				repo.On("GetUserReview", ctx, "tenant-1", "template-1").Return(nil, errors.New("not found")).Once()
				repo.On("CreateReview", ctx, mock.AnythingOfType("*marketplace.TemplateReview")).Return(nil).Once()
				repo.On("UpdateTemplateRating", ctx, "template-1").Return(nil).Once()
				repo.On("GetUserReview", ctx, "tenant-1", "template-1").Return(&TemplateReview{ID: "1"}, nil).Once()
			}

			input := RateTemplateInput{Rating: tt.rating}
			_, err := service.RateTemplate(ctx, "tenant-1", "user-1", "User", "template-1", input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_SearchTemplates_DefaultLimit(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	// When limit is 0, it should default to 20
	repo.On("Search", ctx, mock.MatchedBy(func(f SearchFilter) bool {
		return f.Limit == 20
	})).Return([]*MarketplaceTemplate{}, nil)

	_, err := service.SearchTemplates(ctx, SearchFilter{Limit: 0})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_GetTrending_DefaultLimit(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	// When limit is 0 or negative, it should default to 10
	repo.On("GetTrending", ctx, 7, 10).Return([]*MarketplaceTemplate{}, nil)

	_, err := service.GetTrending(ctx, 0)
	assert.NoError(t, err)

	_, err = service.GetTrending(ctx, -1)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_GetReviews_DefaultLimit(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	// When limit is 0 or negative, it should default to 10
	repo.On("GetReviews", ctx, "template-1", ReviewSortRecent, 10, 0).Return([]*TemplateReview{}, nil)

	_, err := service.GetReviews(ctx, "template-1", ReviewSortRecent, 0, 0)
	assert.NoError(t, err)

	_, err = service.GetReviews(ctx, "template-1", ReviewSortRecent, -5, 0)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// ==================== Enhanced Review Feature Tests ====================

func TestVoteReviewHelpful(t *testing.T) {
	tests := []struct {
		name      string
		reviewID  string
		tenantID  string
		userID    string
		mockError error
		wantError bool
	}{
		{
			name:      "successful vote",
			reviewID:  "review-1",
			tenantID:  "tenant-1",
			userID:    "user-1",
			mockError: nil,
			wantError: false,
		},
		{
			name:      "duplicate vote",
			reviewID:  "review-1",
			tenantID:  "tenant-1",
			userID:    "user-1",
			mockError: errors.New("duplicate vote"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			service := NewService(repo, nil, logger)

			repo.On("VoteReviewHelpful", mock.Anything, mock.AnythingOfType("*marketplace.ReviewHelpfulVote")).
				Return(tt.mockError)

			err := service.VoteReviewHelpful(context.Background(), tt.tenantID, tt.userID, tt.reviewID)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestUnvoteReviewHelpful(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	repo.On("UnvoteReviewHelpful", ctx, "tenant-1", "user-1", "review-1").Return(nil)

	err := service.UnvoteReviewHelpful(ctx, "tenant-1", "user-1", "review-1")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestReportReview(t *testing.T) {
	tests := []struct {
		name      string
		input     ReportReviewInput
		mockError error
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid spam report",
			input: ReportReviewInput{
				Reason:  "spam",
				Details: "This is spam content",
			},
			wantError: false,
		},
		{
			name: "valid offensive report",
			input: ReportReviewInput{
				Reason:  "offensive",
				Details: "Offensive language",
			},
			wantError: false,
		},
		{
			name: "invalid reason",
			input: ReportReviewInput{
				Reason:  "invalid",
				Details: "Details",
			},
			wantError: true,
			errorMsg:  "invalid input",
		},
		{
			name: "empty reason",
			input: ReportReviewInput{
				Reason:  "",
				Details: "Details",
			},
			wantError: true,
			errorMsg:  "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			service := NewService(repo, nil, logger)

			if tt.input.Validate() == nil {
				repo.On("CreateReviewReport", mock.Anything, mock.AnythingOfType("*marketplace.ReviewReport")).
					Return(tt.mockError)
			}

			err := service.ReportReview(context.Background(), "tenant-1", "user-1", "review-1", tt.input)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestGetReviewsWithSorting(t *testing.T) {
	tests := []struct {
		name   string
		sortBy ReviewSortOption
	}{
		{"sort by recent", ReviewSortRecent},
		{"sort by helpful", ReviewSortHelpful},
		{"sort by rating high", ReviewSortRatingH},
		{"sort by rating low", ReviewSortRatingL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			service := NewService(repo, nil, logger)
			ctx := context.Background()

			expectedReviews := []*TemplateReview{
				{ID: "review-1", Rating: 5},
				{ID: "review-2", Rating: 4},
			}

			repo.On("GetReviews", ctx, "template-1", tt.sortBy, 10, 0).Return(expectedReviews, nil)

			reviews, err := service.GetReviews(ctx, "template-1", tt.sortBy, 10, 0)
			assert.NoError(t, err)
			assert.Equal(t, expectedReviews, reviews)

			repo.AssertExpectations(t)
		})
	}
}

func TestHideReview(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	repo.On("HideReview", ctx, "review-1", "spam", "admin-1").Return(nil)

	err := service.HideReview(ctx, "review-1", "spam", "admin-1")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetRatingDistribution(t *testing.T) {
	repo := new(MockRepository)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewService(repo, nil, logger)
	ctx := context.Background()

	expectedDist := &RatingDistribution{
		Rating1Count:   5,
		Rating2Count:   10,
		Rating3Count:   15,
		Rating4Count:   20,
		Rating5Count:   50,
		TotalRatings:   100,
		AverageRating:  4.0,
		Rating1Percent: 5.0,
		Rating2Percent: 10.0,
		Rating3Percent: 15.0,
		Rating4Percent: 20.0,
		Rating5Percent: 50.0,
	}

	repo.On("GetRatingDistribution", ctx, "template-1").Return(expectedDist, nil)

	dist, err := service.GetRatingDistribution(ctx, "template-1")
	require.NoError(t, err)
	assert.Equal(t, expectedDist, dist)
	repo.AssertExpectations(t)
}

func TestResolveReviewReport(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		wantError bool
		errorMsg  string
	}{
		{"valid reviewed status", string(ReportStatusReviewed), false, ""},
		{"valid actioned status", string(ReportStatusActioned), false, ""},
		{"valid dismissed status", string(ReportStatusDismissed), false, ""},
		{"invalid status", "invalid", true, "invalid status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			service := NewService(repo, nil, logger)
			ctx := context.Background()

			if !tt.wantError {
				repo.On("UpdateReviewReportStatus", ctx, "report-1", tt.status, "admin-1", (*string)(nil)).Return(nil)
			}

			err := service.ResolveReviewReport(ctx, "report-1", tt.status, "admin-1", nil)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestReportReviewInput_Validate(t *testing.T) {
	tests := []struct {
		name      string
		input     ReportReviewInput
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid spam",
			input:     ReportReviewInput{Reason: "spam", Details: "This is spam"},
			wantError: false,
		},
		{
			name:      "valid inappropriate",
			input:     ReportReviewInput{Reason: "inappropriate", Details: "Inappropriate content"},
			wantError: false,
		},
		{
			name:      "empty reason",
			input:     ReportReviewInput{Reason: "", Details: "Details"},
			wantError: true,
			errorMsg:  "reason is required",
		},
		{
			name:      "invalid reason",
			input:     ReportReviewInput{Reason: "invalid", Details: "Details"},
			wantError: true,
			errorMsg:  "reason must be one of",
		},
		{
			name:      "details too long",
			input:     ReportReviewInput{Reason: "spam", Details: string(make([]byte, 1001))},
			wantError: true,
			errorMsg:  "details must be 1000 characters or less",
		},
		{
			name:      "valid without details",
			input:     ReportReviewInput{Reason: "spam", Details: ""},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
