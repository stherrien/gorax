package marketplace

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_PublishAndSearchTemplate tests the full publish and search workflow
func TestIntegration_PublishAndSearchTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	workflowService := &mockWorkflowService{}
	service := NewService(repo, workflowService, testLogger())

	// Step 1: Publish a template
	definition := json.RawMessage(`{
		"nodes": [
			{"id": "1", "type": "trigger", "data": {"nodeType": "webhook"}},
			{"id": "2", "type": "action", "data": {"nodeType": "http"}}
		],
		"edges": [
			{"id": "e1", "source": "1", "target": "2"}
		]
	}`)

	publishInput := PublishTemplateInput{
		Name:        "Webhook to HTTP Template",
		Description: "A simple template that receives a webhook and calls an HTTP API",
		Category:    string(CategoryIntegration),
		Definition:  definition,
		Tags:        []string{"webhook", "http", "api"},
		Version:     "1.0.0",
	}

	template, err := service.PublishTemplate(ctx, "user-123", "John Doe", publishInput)
	require.NoError(t, err)
	require.NotNil(t, template)
	assert.Equal(t, publishInput.Name, template.Name)
	assert.Equal(t, "user-123", template.AuthorID)
	assert.False(t, template.IsVerified)
	assert.Equal(t, 0, template.DownloadCount)
	t.Logf("✓ Template published with ID: %s", template.ID)

	// Step 2: Search for the template by category
	filter := SearchFilter{
		Category: string(CategoryIntegration),
		Limit:    10,
	}

	templates, err := service.SearchTemplates(ctx, filter)
	require.NoError(t, err)
	assert.NotEmpty(t, templates)

	found := false
	for _, tmpl := range templates {
		if tmpl.ID == template.ID {
			found = true
			assert.Equal(t, template.Name, tmpl.Name)
			break
		}
	}
	assert.True(t, found, "published template should be found in search results")
	t.Logf("✓ Template found in search results")

	// Step 3: Search by tag
	tagFilter := SearchFilter{
		Tags:  []string{"webhook"},
		Limit: 10,
	}

	tagResults, err := service.SearchTemplates(ctx, tagFilter)
	require.NoError(t, err)

	found = false
	for _, tmpl := range tagResults {
		if tmpl.ID == template.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "template should be found by tag")
	t.Logf("✓ Template found by tag search")

	// Step 4: Get template by ID
	retrieved, err := service.GetTemplate(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, template.ID, retrieved.ID)
	assert.Equal(t, template.Name, retrieved.Name)
	assert.Equal(t, template.AuthorID, retrieved.AuthorID)
	t.Logf("✓ Template retrieved by ID")
}

// TestIntegration_InstallTemplateWorkflow tests the complete template installation workflow
func TestIntegration_InstallTemplateWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	workflowService := &mockWorkflowService{
		createdWorkflowID: "workflow-456",
	}
	service := NewService(repo, workflowService, testLogger())

	// Step 1: Publish a template
	definition := json.RawMessage(`{
		"nodes": [
			{"id": "1", "type": "trigger", "data": {"nodeType": "schedule"}},
			{"id": "2", "type": "action", "data": {"nodeType": "http"}}
		],
		"edges": [
			{"id": "e1", "source": "1", "target": "2"}
		]
	}`)

	publishInput := PublishTemplateInput{
		Name:        "Scheduled HTTP Call",
		Description: "Calls an HTTP endpoint on a schedule",
		Category:    string(CategoryAutomation),
		Definition:  definition,
		Tags:        []string{"schedule", "http"},
		Version:     "1.0.0",
	}

	template, err := service.PublishTemplate(ctx, "user-123", "John Doe", publishInput)
	require.NoError(t, err)
	t.Logf("✓ Template published: %s", template.ID)

	// Step 2: Install the template
	tenantID := "tenant-789"
	userID := "user-456"
	installInput := InstallTemplateInput{
		WorkflowName: "My Scheduled Workflow",
	}

	result, err := service.InstallTemplate(ctx, tenantID, userID, template.ID, installInput)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "workflow-456", result.WorkflowID)
	assert.Equal(t, installInput.WorkflowName, result.WorkflowName)
	t.Logf("✓ Template installed as workflow: %s", result.WorkflowID)

	// Step 3: Verify download count incremented
	retrieved, err := service.GetTemplate(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, retrieved.DownloadCount)
	t.Logf("✓ Download count incremented to: %d", retrieved.DownloadCount)

	// Step 4: Try to install again (should fail)
	_, err = service.InstallTemplate(ctx, tenantID, userID, template.ID, installInput)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already installed")
	t.Logf("✓ Duplicate installation prevented")

	// Step 5: Install in different tenant (should succeed)
	differentTenantID := "tenant-999"
	result2, err := service.InstallTemplate(ctx, differentTenantID, userID, template.ID, installInput)
	require.NoError(t, err)
	assert.NotEmpty(t, result2.WorkflowID)
	t.Logf("✓ Template installed in different tenant")

	// Step 6: Verify download count incremented again
	retrieved, err = service.GetTemplate(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, retrieved.DownloadCount)
	t.Logf("✓ Download count incremented to: %d", retrieved.DownloadCount)
}

// TestIntegration_RateAndReviewTemplate tests the rating and review workflow
func TestIntegration_RateAndReviewTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	workflowService := &mockWorkflowService{}
	service := NewService(repo, workflowService, testLogger())

	// Step 1: Publish a template
	definition := json.RawMessage(`{
		"nodes": [{"id": "1", "type": "trigger"}],
		"edges": []
	}`)

	template, err := service.PublishTemplate(ctx, "user-123", "John Doe", PublishTemplateInput{
		Name:        "Test Template",
		Description: "A template for testing reviews",
		Category:    string(CategoryOther),
		Definition:  definition,
		Version:     "1.0.0",
	})
	require.NoError(t, err)
	t.Logf("✓ Template published: %s", template.ID)

	// Step 2: Add first review (5 stars)
	rateInput1 := RateTemplateInput{
		Rating:  5,
		Comment: "Excellent template! Works perfectly.",
	}

	review1, err := service.RateTemplate(ctx, "tenant-1", "user-1", "Alice", template.ID, rateInput1)
	require.NoError(t, err)
	assert.Equal(t, 5, review1.Rating)
	assert.Equal(t, "Alice", review1.UserName)
	t.Logf("✓ First review added: 5 stars")

	// Step 3: Verify average rating
	retrieved, err := service.GetTemplate(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, 5.0, retrieved.AverageRating)
	assert.Equal(t, 1, retrieved.TotalRatings)
	t.Logf("✓ Average rating: %.1f (%d ratings)", retrieved.AverageRating, retrieved.TotalRatings)

	// Step 4: Add second review (3 stars)
	rateInput2 := RateTemplateInput{
		Rating:  3,
		Comment: "Good but needs improvement",
	}

	review2, err := service.RateTemplate(ctx, "tenant-2", "user-2", "Bob", template.ID, rateInput2)
	require.NoError(t, err)
	assert.Equal(t, 3, review2.Rating)
	t.Logf("✓ Second review added: 3 stars")

	// Step 5: Verify updated average rating (5 + 3) / 2 = 4.0
	retrieved, err = service.GetTemplate(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, 4.0, retrieved.AverageRating)
	assert.Equal(t, 2, retrieved.TotalRatings)
	t.Logf("✓ Average rating updated: %.1f (%d ratings)", retrieved.AverageRating, retrieved.TotalRatings)

	// Step 6: Update existing review
	updateInput := RateTemplateInput{
		Rating:  4,
		Comment: "Actually it's pretty good after I figured it out",
	}

	updatedReview, err := service.RateTemplate(ctx, "tenant-2", "user-2", "Bob", template.ID, updateInput)
	require.NoError(t, err)
	assert.Equal(t, 4, updatedReview.Rating)
	assert.Equal(t, updateInput.Comment, updatedReview.Comment)
	t.Logf("✓ Review updated: 3 -> 4 stars")

	// Step 7: Verify average rating updated (5 + 4) / 2 = 4.5
	retrieved, err = service.GetTemplate(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, 4.5, retrieved.AverageRating)
	assert.Equal(t, 2, retrieved.TotalRatings) // Should still be 2
	t.Logf("✓ Average rating after update: %.1f (%d ratings)", retrieved.AverageRating, retrieved.TotalRatings)

	// Step 8: Get all reviews
	reviews, err := service.GetReviews(ctx, template.ID, ReviewSortRecent, 10, 0)
	require.NoError(t, err)
	assert.Len(t, reviews, 2)
	t.Logf("✓ Retrieved %d reviews", len(reviews))

	// Step 9: Delete a review
	err = service.DeleteReview(ctx, "tenant-1", template.ID, review1.ID)
	require.NoError(t, err)
	t.Logf("✓ Review deleted")

	// Step 10: Verify average rating recalculated (only 4-star review remains)
	retrieved, err = service.GetTemplate(ctx, template.ID)
	require.NoError(t, err)
	assert.Equal(t, 4.0, retrieved.AverageRating)
	assert.Equal(t, 1, retrieved.TotalRatings)
	t.Logf("✓ Average rating after deletion: %.1f (%d ratings)", retrieved.AverageRating, retrieved.TotalRatings)
}

// TestIntegration_TrendingAndPopularTemplates tests trending and popular queries
func TestIntegration_TrendingAndPopularTemplates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	workflowService := &mockWorkflowService{}
	service := NewService(repo, workflowService, testLogger())

	// Step 1: Publish multiple templates
	templates := make([]*MarketplaceTemplate, 3)
	for i := 0; i < 3; i++ {
		definition := json.RawMessage(`{
			"nodes": [{"id": "1", "type": "trigger"}],
			"edges": []
		}`)

		template, err := service.PublishTemplate(ctx, "user-123", "John Doe", PublishTemplateInput{
			Name:        "Template " + string(rune('A'+i)),
			Description: "Test template for trending/popular tests",
			Category:    string(CategoryAutomation),
			Definition:  definition,
			Version:     "1.0.0",
		})
		require.NoError(t, err)
		templates[i] = template
	}
	t.Logf("✓ Published %d templates", len(templates))

	// Step 2: Simulate downloads (different amounts for each)
	// Template A: 10 downloads
	for i := 0; i < 10; i++ {
		err := repo.IncrementDownloadCount(ctx, templates[0].ID)
		require.NoError(t, err)
	}

	// Template B: 5 downloads
	for i := 0; i < 5; i++ {
		err := repo.IncrementDownloadCount(ctx, templates[1].ID)
		require.NoError(t, err)
	}

	// Template C: 2 downloads
	for i := 0; i < 2; i++ {
		err := repo.IncrementDownloadCount(ctx, templates[2].ID)
		require.NoError(t, err)
	}
	t.Logf("✓ Simulated downloads")

	// Step 3: Get popular templates (should be ordered by download count)
	popular, err := service.GetPopular(ctx, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, popular)

	// Verify ordering (Template A should be first)
	if len(popular) >= 2 {
		assert.GreaterOrEqual(t, popular[0].DownloadCount, popular[1].DownloadCount)
	}
	t.Logf("✓ Retrieved %d popular templates", len(popular))

	// Step 4: Get trending templates (recent downloads)
	trending, err := service.GetTrending(ctx, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, trending)
	t.Logf("✓ Retrieved %d trending templates", len(trending))
}

// TestIntegration_TemplateValidation tests definition validation
func TestIntegration_TemplateValidation(t *testing.T) {
	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	workflowService := &mockWorkflowService{}
	service := NewService(repo, workflowService, testLogger())

	tests := []struct {
		name        string
		input       PublishTemplateInput
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid template",
			input: PublishTemplateInput{
				Name:        "Valid Template",
				Description: "This is a valid template description",
				Category:    string(CategoryAutomation),
				Definition: json.RawMessage(`{
					"nodes": [{"id": "1", "type": "trigger"}],
					"edges": []
				}`),
				Version: "1.0.0",
			},
			expectError: false,
		},
		{
			name: "missing nodes field",
			input: PublishTemplateInput{
				Name:        "Invalid Template",
				Description: "Missing nodes field",
				Category:    string(CategoryAutomation),
				Definition:  json.RawMessage(`{"edges": []}`),
				Version:     "1.0.0",
			},
			expectError: true,
			errorMsg:    "nodes",
		},
		{
			name: "missing edges field",
			input: PublishTemplateInput{
				Name:        "Invalid Template",
				Description: "Missing edges field",
				Category:    string(CategoryAutomation),
				Definition:  json.RawMessage(`{"nodes": []}`),
				Version:     "1.0.0",
			},
			expectError: true,
			errorMsg:    "edges",
		},
		{
			name: "invalid JSON",
			input: PublishTemplateInput{
				Name:        "Invalid Template",
				Description: "Invalid JSON definition",
				Category:    string(CategoryAutomation),
				Definition:  json.RawMessage(`{invalid json}`),
				Version:     "1.0.0",
			},
			expectError: true,
			errorMsg:    "JSON",
		},
		{
			name: "empty name",
			input: PublishTemplateInput{
				Name:        "",
				Description: "Valid description",
				Category:    string(CategoryAutomation),
				Definition: json.RawMessage(`{
					"nodes": [{"id": "1", "type": "trigger"}],
					"edges": []
				}`),
				Version: "1.0.0",
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "description too short",
			input: PublishTemplateInput{
				Name:        "Valid Name",
				Description: "Short",
				Category:    string(CategoryAutomation),
				Definition: json.RawMessage(`{
					"nodes": [{"id": "1", "type": "trigger"}],
					"edges": []
				}`),
				Version: "1.0.0",
			},
			expectError: true,
			errorMsg:    "at least 10 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.PublishTemplate(ctx, "user-123", "John Doe", tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				t.Logf("✓ Validation failed as expected: %v", err)
			} else {
				assert.NoError(t, err)
				t.Logf("✓ Validation passed")
			}
		})
	}
}

// Mock implementations for testing
type mockWorkflowService struct {
	createdWorkflowID string
}

func (m *mockWorkflowService) CreateFromTemplate(ctx context.Context, tenantID, userID, templateID, workflowName string, definition json.RawMessage) (string, error) {
	if m.createdWorkflowID != "" {
		return m.createdWorkflowID, nil
	}
	return "mock-workflow-id", nil
}

// Helper functions
func setupTestRepository(t *testing.T) Repository {
	t.Helper()
	return &mockRepository{
		templates:     make(map[string]*MarketplaceTemplate),
		reviews:       make(map[string]*TemplateReview),
		installations: make(map[string]*TemplateInstallation),
	}
}

func testLogger() *slog.Logger {
	return slog.Default()
}

// Mock repository implementation (simplified for testing)
type mockRepository struct {
	templates     map[string]*MarketplaceTemplate
	reviews       map[string]*TemplateReview
	installations map[string]*TemplateInstallation
	nextID        int
}

func (m *mockRepository) Publish(ctx context.Context, template *MarketplaceTemplate) error {
	m.nextID++
	template.ID = "template-" + string(rune('A'+m.nextID))
	template.PublishedAt = time.Now()
	template.UpdatedAt = time.Now()
	m.templates[template.ID] = template
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*MarketplaceTemplate, error) {
	template, ok := m.templates[id]
	if !ok {
		return nil, ErrTemplateNotFound
	}
	return template, nil
}

func (m *mockRepository) Search(ctx context.Context, filter SearchFilter) ([]*MarketplaceTemplate, error) {
	var results []*MarketplaceTemplate
	for _, template := range m.templates {
		if filter.Category != "" && template.Category != filter.Category {
			continue
		}
		if len(filter.Tags) > 0 {
			hasTag := false
			for _, tag := range filter.Tags {
				for _, templateTag := range template.Tags {
					if templateTag == tag {
						hasTag = true
						break
					}
				}
			}
			if !hasTag {
				continue
			}
		}
		results = append(results, template)
	}
	return results, nil
}

func (m *mockRepository) GetTrending(ctx context.Context, days, limit int) ([]*MarketplaceTemplate, error) {
	var results []*MarketplaceTemplate
	for _, template := range m.templates {
		results = append(results, template)
	}
	return results, nil
}

func (m *mockRepository) GetPopular(ctx context.Context, limit int) ([]*MarketplaceTemplate, error) {
	var results []*MarketplaceTemplate
	for _, template := range m.templates {
		results = append(results, template)
	}
	return results, nil
}

func (m *mockRepository) GetByAuthor(ctx context.Context, authorID string) ([]*MarketplaceTemplate, error) {
	var results []*MarketplaceTemplate
	for _, template := range m.templates {
		if template.AuthorID == authorID {
			results = append(results, template)
		}
	}
	return results, nil
}

func (m *mockRepository) Update(ctx context.Context, id string, input UpdateTemplateInput) error {
	template, ok := m.templates[id]
	if !ok {
		return ErrTemplateNotFound
	}

	if input.Name != "" {
		template.Name = input.Name
	}
	if input.Description != "" {
		template.Description = input.Description
	}
	if input.Category != "" {
		template.Category = input.Category
	}
	if input.Definition != nil {
		template.Definition = input.Definition
	}
	if input.Tags != nil {
		template.Tags = input.Tags
	}
	if input.Version != "" {
		template.Version = input.Version
	}

	template.UpdatedAt = time.Now()
	return nil
}

func (m *mockRepository) IncrementDownloadCount(ctx context.Context, templateID string) error {
	template, ok := m.templates[templateID]
	if !ok {
		return ErrTemplateNotFound
	}
	template.DownloadCount++
	return nil
}

func (m *mockRepository) CreateInstallation(ctx context.Context, installation *TemplateInstallation) error {
	m.nextID++
	installation.ID = "install-" + string(rune('A'+m.nextID))
	installation.InstalledAt = time.Now()
	key := installation.TenantID + ":" + installation.TemplateID
	m.installations[key] = installation
	return nil
}

func (m *mockRepository) GetInstallation(ctx context.Context, tenantID, templateID string) (*TemplateInstallation, error) {
	key := tenantID + ":" + templateID
	installation, ok := m.installations[key]
	if !ok {
		return nil, nil
	}
	return installation, nil
}

func (m *mockRepository) CreateReview(ctx context.Context, review *TemplateReview) error {
	m.nextID++
	review.ID = "review-" + string(rune('A'+m.nextID))
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()
	key := review.TenantID + ":" + review.TemplateID
	m.reviews[key] = review
	return nil
}

func (m *mockRepository) GetUserReview(ctx context.Context, tenantID, templateID string) (*TemplateReview, error) {
	key := tenantID + ":" + templateID
	review, ok := m.reviews[key]
	if !ok {
		return nil, nil
	}
	return review, nil
}

func (m *mockRepository) UpdateReview(ctx context.Context, tenantID, reviewID string, rating int, comment string) error {
	for key, review := range m.reviews {
		if review.ID == reviewID && review.TenantID == tenantID {
			review.Rating = rating
			review.Comment = comment
			review.UpdatedAt = time.Now()
			m.reviews[key] = review
			return nil
		}
	}
	return ErrReviewNotFound
}

func (m *mockRepository) GetReviews(ctx context.Context, templateID string, sortBy ReviewSortOption, limit, offset int) ([]*TemplateReview, error) {
	var results []*TemplateReview
	for _, review := range m.reviews {
		if review.TemplateID == templateID {
			results = append(results, review)
		}
	}
	return results, nil
}

func (m *mockRepository) CreateReviewReport(ctx context.Context, report *ReviewReport) error {
	// Mock implementation - just accept the report
	return nil
}

func (m *mockRepository) DeleteReview(ctx context.Context, tenantID, reviewID string) error {
	for key, review := range m.reviews {
		if review.ID == reviewID && review.TenantID == tenantID {
			delete(m.reviews, key)
			return nil
		}
	}
	return ErrReviewNotFound
}

func (m *mockRepository) UpdateTemplateRating(ctx context.Context, templateID string) error {
	template, ok := m.templates[templateID]
	if !ok {
		return ErrTemplateNotFound
	}

	// Calculate average rating from reviews
	var totalRating int
	var count int
	for _, review := range m.reviews {
		if review.TemplateID == templateID {
			totalRating += review.Rating
			count++
		}
	}

	if count > 0 {
		template.AverageRating = float64(totalRating) / float64(count)
		template.TotalRatings = count
	} else {
		template.AverageRating = 0
		template.TotalRatings = 0
	}

	return nil
}

// Error definitions
var (
	ErrTemplateNotFound = &templateError{message: "template not found"}
	ErrReviewNotFound   = &templateError{message: "review not found"}
)

type templateError struct {
	message string
}

func (e *templateError) Error() string {
	return e.message
}
