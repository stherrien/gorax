package marketplace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCategory_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cat     Category
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid category",
			cat: Category{
				Name:        "Integration",
				Slug:        "integration",
				Description: "Templates for integrating with external services",
				Icon:        "link",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			cat: Category{
				Slug:        "integration",
				Description: "Test",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too long",
			cat: Category{
				Name:        string(make([]byte, 256)),
				Slug:        "test",
				Description: "Test",
			},
			wantErr: true,
			errMsg:  "name must be 255 characters or less",
		},
		{
			name: "missing slug",
			cat: Category{
				Name:        "Test",
				Description: "Test",
			},
			wantErr: true,
			errMsg:  "slug is required",
		},
		{
			name: "invalid slug format",
			cat: Category{
				Name:        "Test",
				Slug:        "Invalid Slug!",
				Description: "Test",
			},
			wantErr: true,
			errMsg:  "slug must contain only lowercase letters, numbers, and hyphens",
		},
		{
			name: "description too long",
			cat: Category{
				Name:        "Test",
				Slug:        "test",
				Description: string(make([]byte, 1001)),
			},
			wantErr: true,
			errMsg:  "description must be 1000 characters or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cat.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateCategoryInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateCategoryInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: CreateCategoryInput{
				Name:        "DevOps",
				Slug:        "devops",
				Description: "DevOps automation templates",
				Icon:        "server",
			},
			wantErr: false,
		},
		{
			name: "with parent",
			input: CreateCategoryInput{
				Name:        "CI/CD",
				Slug:        "cicd",
				Description: "Continuous integration templates",
				ParentID:    catStringPtr("parent-id"),
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			input: CreateCategoryInput{
				Name: "Test",
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

func TestMarketplaceTemplateWithCategories(t *testing.T) {
	now := time.Now()
	featuredBy := "admin-user"
	template := MarketplaceTemplate{
		ID:          "test-id",
		Name:        "Test Template",
		Description: "Test description",
		IsFeatured:  true,
		FeaturedAt:  &now,
		FeaturedBy:  &featuredBy,
		PublishedAt: now,
		UpdatedAt:   now,
	}

	assert.Equal(t, "test-id", template.ID)
	assert.True(t, template.IsFeatured)
	assert.NotNil(t, template.FeaturedAt)
	assert.Equal(t, "admin-user", *template.FeaturedBy)
}

func TestSearchFilter_Enhanced(t *testing.T) {
	tests := []struct {
		name    string
		filter  EnhancedSearchFilter
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid filter with categories",
			filter: EnhancedSearchFilter{
				CategoryIDs: []string{"cat-1", "cat-2"},
				SortBy:      "popular",
				Page:        1,
				Limit:       20,
			},
			wantErr: false,
		},
		{
			name: "filter by featured",
			filter: EnhancedSearchFilter{
				IsFeatured: catBoolPtr(true),
				SortBy:     "recent",
			},
			wantErr: false,
		},
		{
			name: "invalid sort by",
			filter: EnhancedSearchFilter{
				SortBy: "invalid",
			},
			wantErr: true,
			errMsg:  "sort_by must be one of",
		},
		{
			name: "valid search query",
			filter: EnhancedSearchFilter{
				SearchQuery: "slack integration",
				SortBy:      "relevance",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFeatureTemplateInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   FeatureTemplateInput
		wantErr bool
	}{
		{
			name: "feature template",
			input: FeatureTemplateInput{
				IsFeatured: true,
			},
			wantErr: false,
		},
		{
			name: "unfeature template",
			input: FeatureTemplateInput{
				IsFeatured: false,
			},
			wantErr: false,
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

func catStringPtr(s string) *string {
	return &s
}

func catBoolPtr(b bool) *bool {
	return &b
}
