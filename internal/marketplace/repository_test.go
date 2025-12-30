package marketplace

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	// This would connect to a test database
	// For now, we'll skip integration tests that require DB
	t.Skip("Integration test - requires database")
	return nil
}

func TestPublishTemplate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	definition := json.RawMessage(`{"nodes":[],"edges":[]}`)
	template := &MarketplaceTemplate{
		Name:        "Test Template",
		Description: "Test description",
		Category:    "automation",
		Definition:  definition,
		Tags:        []string{"test", "automation"},
		AuthorID:    "user-1",
		AuthorName:  "Test User",
		Version:     "1.0.0",
	}

	err := repo.Publish(ctx, template)
	require.NoError(t, err)
	assert.NotEmpty(t, template.ID)
	assert.NotZero(t, template.PublishedAt)
}

func TestGetTemplateByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Test getting non-existent template
	_, err := repo.GetByID(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSearchTemplates(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	tests := []struct {
		name   string
		filter SearchFilter
	}{
		{
			name: "search by category",
			filter: SearchFilter{
				Category: "automation",
				Limit:    10,
			},
		},
		{
			name: "search by tags",
			filter: SearchFilter{
				Tags:  []string{"security"},
				Limit: 10,
			},
		},
		{
			name: "search by rating",
			filter: SearchFilter{
				MinRating: func() *float64 { r := 4.0; return &r }(),
				Limit:     10,
			},
		},
		{
			name: "search by query",
			filter: SearchFilter{
				SearchQuery: "webhook",
				Limit:       10,
			},
		},
		{
			name: "sort by popular",
			filter: SearchFilter{
				SortBy: "popular",
				Limit:  10,
			},
		},
		{
			name: "sort by recent",
			filter: SearchFilter{
				SortBy: "recent",
				Limit:  10,
			},
		},
		{
			name: "sort by rating",
			filter: SearchFilter{
				SortBy: "rating",
				Limit:  10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := repo.Search(ctx, tt.filter)
			require.NoError(t, err)
			assert.NotNil(t, templates)
		})
	}
}

func TestGetPopular(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	templates, err := repo.GetPopular(ctx, 10)
	require.NoError(t, err)
	assert.NotNil(t, templates)
	assert.LessOrEqual(t, len(templates), 10)
}

func TestGetTrending(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Get trending in last 7 days
	days := 7
	templates, err := repo.GetTrending(ctx, days, 10)
	require.NoError(t, err)
	assert.NotNil(t, templates)
	assert.LessOrEqual(t, len(templates), 10)
}

func TestGetByAuthor(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	authorID := "user-1"
	templates, err := repo.GetByAuthor(ctx, authorID)
	require.NoError(t, err)
	assert.NotNil(t, templates)
}

func TestIncrementDownloadCount(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	templateID := "template-1"
	err := repo.IncrementDownloadCount(ctx, templateID)
	assert.NoError(t, err)
}

func TestCreateInstallation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	installation := &TemplateInstallation{
		TemplateID:       "template-1",
		TenantID:         "tenant-1",
		UserID:           "user-1",
		WorkflowID:       "workflow-1",
		InstalledVersion: "1.0.0",
	}

	err := repo.CreateInstallation(ctx, installation)
	require.NoError(t, err)
	assert.NotEmpty(t, installation.ID)
	assert.NotZero(t, installation.InstalledAt)
}

func TestGetInstallation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Test getting installation for non-existent template
	_, err := repo.GetInstallation(ctx, "tenant-1", "template-1")
	assert.Error(t, err)
}

func TestCreateReview(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	review := &TemplateReview{
		TemplateID: "template-1",
		TenantID:   "tenant-1",
		UserID:     "user-1",
		UserName:   "Test User",
		Rating:     5,
		Comment:    "Great template!",
	}

	err := repo.CreateReview(ctx, review)
	require.NoError(t, err)
	assert.NotEmpty(t, review.ID)
	assert.NotZero(t, review.CreatedAt)
}

func TestUpdateReview(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	err := repo.UpdateReview(ctx, "tenant-1", "review-1", 4, "Updated comment")
	assert.NoError(t, err)
}

func TestDeleteReview(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	err := repo.DeleteReview(ctx, "tenant-1", "review-1")
	assert.NoError(t, err)
}

func TestGetReviews(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	reviews, err := repo.GetReviews(ctx, "template-1", 10, 0)
	require.NoError(t, err)
	assert.NotNil(t, reviews)
}

func TestGetUserReview(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Test getting review for non-existent user
	_, err := repo.GetUserReview(ctx, "tenant-1", "template-1")
	assert.Error(t, err)
}

func TestUpdateTemplateRating(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	err := repo.UpdateTemplateRating(ctx, "template-1")
	assert.NoError(t, err)
}

func TestSearchFilter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filter  SearchFilter
		wantErr bool
	}{
		{
			name: "valid filter",
			filter: SearchFilter{
				Category: "automation",
				Limit:    10,
			},
			wantErr: false,
		},
		{
			name: "invalid category length",
			filter: SearchFilter{
				Category: string(make([]byte, 101)),
			},
			wantErr: true,
		},
		{
			name: "invalid min rating",
			filter: SearchFilter{
				MinRating: func() *float64 { r := 6.0; return &r }(),
			},
			wantErr: true,
		},
		{
			name: "invalid sort by",
			filter: SearchFilter{
				SortBy: "invalid",
			},
			wantErr: true,
		},
		{
			name: "negative page",
			filter: SearchFilter{
				Page: -1,
			},
			wantErr: true,
		},
		{
			name: "limit too high",
			filter: SearchFilter{
				Limit: 101,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPublishTemplateInput_Validate(t *testing.T) {
	validDef := json.RawMessage(`{"nodes":[],"edges":[]}`)

	tests := []struct {
		name    string
		input   PublishTemplateInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: PublishTemplateInput{
				Name:        "Test Template",
				Description: "A test template description that is long enough",
				Category:    "automation",
				Definition:  validDef,
				Version:     "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			input: PublishTemplateInput{
				Description: "Description",
				Category:    "automation",
				Definition:  validDef,
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "description too short",
			input: PublishTemplateInput{
				Name:        "Test",
				Description: "Short",
				Category:    "automation",
				Definition:  validDef,
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "empty category",
			input: PublishTemplateInput{
				Name:        "Test",
				Description: "Description that is long enough",
				Definition:  validDef,
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "empty definition",
			input: PublishTemplateInput{
				Name:        "Test",
				Description: "Description that is long enough",
				Category:    "automation",
				Version:     "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "empty version",
			input: PublishTemplateInput{
				Name:        "Test",
				Description: "Description that is long enough",
				Category:    "automation",
				Definition:  validDef,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRateTemplateInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   RateTemplateInput
		wantErr bool
	}{
		{
			name: "valid rating",
			input: RateTemplateInput{
				Rating:  5,
				Comment: "Great!",
			},
			wantErr: false,
		},
		{
			name: "rating too low",
			input: RateTemplateInput{
				Rating: 0,
			},
			wantErr: true,
		},
		{
			name: "rating too high",
			input: RateTemplateInput{
				Rating: 6,
			},
			wantErr: true,
		},
		{
			name: "comment too long",
			input: RateTemplateInput{
				Rating:  5,
				Comment: string(make([]byte, 2001)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetCategories(t *testing.T) {
	categories := GetCategories()
	assert.NotEmpty(t, categories)
	assert.Contains(t, categories, "security")
	assert.Contains(t, categories, "automation")
}
