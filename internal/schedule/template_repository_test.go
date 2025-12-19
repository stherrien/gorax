package schedule

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTemplateRepository is a mock implementation for testing
type MockTemplateRepository struct {
	templates []*ScheduleTemplate
}

func (m *MockTemplateRepository) List(ctx context.Context, filter ScheduleTemplateFilter) ([]*ScheduleTemplate, error) {
	var result []*ScheduleTemplate

	for _, tmpl := range m.templates {
		// Apply category filter
		if filter.Category != "" && tmpl.Category != filter.Category {
			continue
		}

		// Apply tags filter
		if len(filter.Tags) > 0 {
			hasTag := false
			for _, filterTag := range filter.Tags {
				for _, tag := range tmpl.Tags {
					if tag == filterTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Apply system filter
		if filter.IsSystem != nil && tmpl.IsSystem != *filter.IsSystem {
			continue
		}

		// Apply search query
		if filter.SearchQuery != "" {
			if !containsIgnoreCase(tmpl.Name, filter.SearchQuery) &&
				!containsIgnoreCase(tmpl.Description, filter.SearchQuery) {
				continue
			}
		}

		result = append(result, tmpl)
	}

	return result, nil
}

func (m *MockTemplateRepository) GetByID(ctx context.Context, id string) (*ScheduleTemplate, error) {
	for _, tmpl := range m.templates {
		if tmpl.ID == id {
			return tmpl, nil
		}
	}
	return nil, errTemplateNotFoundTest{}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && containsHelper(toLower(s), toLower(substr))
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// errTemplateNotFoundTest is returned when template is not found (test-only error)
type errTemplateNotFoundTest struct{}

func (e errTemplateNotFoundTest) Error() string {
	return "template not found"
}

func createTestTemplates() []*ScheduleTemplate {
	return []*ScheduleTemplate{
		{
			ID:             "tmpl-1",
			Name:           "SOC2 Daily Scan",
			Description:    "Daily SOC2 compliance scan",
			Category:       "compliance",
			CronExpression: "0 2 * * *",
			Timezone:       "UTC",
			Tags:           []string{"compliance", "soc2", "daily"},
			IsSystem:       true,
		},
		{
			ID:             "tmpl-2",
			Name:           "Hourly Security Check",
			Description:    "Hourly security vulnerability scan",
			Category:       "security",
			CronExpression: "0 * * * *",
			Timezone:       "UTC",
			Tags:           []string{"security", "hourly"},
			IsSystem:       true,
		},
		{
			ID:             "tmpl-3",
			Name:           "Weekly Report",
			Description:    "Weekly compliance report generation",
			Category:       "compliance",
			CronExpression: "0 9 * * 1",
			Timezone:       "America/New_York",
			Tags:           []string{"compliance", "weekly", "report"},
			IsSystem:       true,
		},
		{
			ID:             "tmpl-4",
			Name:           "Every 5 Minutes",
			Description:    "High frequency monitoring",
			Category:       "frequent",
			CronExpression: "*/5 * * * *",
			Timezone:       "UTC",
			Tags:           []string{"frequent", "monitoring"},
			IsSystem:       true,
		},
		{
			ID:             "tmpl-5",
			Name:           "Wiz Daily Sync",
			Description:    "Daily Wiz vulnerability sync",
			Category:       "security",
			CronExpression: "0 3 * * *",
			Timezone:       "UTC",
			Tags:           []string{"security", "wiz", "daily"},
			IsSystem:       true,
		},
		{
			ID:             "tmpl-6",
			Name:           "Custom User Template",
			Description:    "User-created template",
			Category:       "custom",
			CronExpression: "0 12 * * *",
			Timezone:       "Europe/London",
			Tags:           []string{"custom"},
			IsSystem:       false,
		},
	}
}

func TestScheduleTemplateRepository_List(t *testing.T) {
	repo := &MockTemplateRepository{templates: createTestTemplates()}
	ctx := context.Background()

	tests := []struct {
		name          string
		filter        ScheduleTemplateFilter
		expectMinimum int
		checkCategory string
	}{
		{
			name:          "list all templates",
			filter:        ScheduleTemplateFilter{},
			expectMinimum: 6,
		},
		{
			name: "filter by compliance category",
			filter: ScheduleTemplateFilter{
				Category: "compliance",
			},
			expectMinimum: 2,
			checkCategory: "compliance",
		},
		{
			name: "filter by security category",
			filter: ScheduleTemplateFilter{
				Category: "security",
			},
			expectMinimum: 2,
			checkCategory: "security",
		},
		{
			name: "filter by frequent category",
			filter: ScheduleTemplateFilter{
				Category: "frequent",
			},
			expectMinimum: 1,
			checkCategory: "frequent",
		},
		{
			name: "filter by tags",
			filter: ScheduleTemplateFilter{
				Tags: []string{"daily"},
			},
			expectMinimum: 2,
		},
		{
			name: "filter by system templates",
			filter: ScheduleTemplateFilter{
				IsSystem: boolPtr(true),
			},
			expectMinimum: 5,
		},
		{
			name: "filter by non-system templates",
			filter: ScheduleTemplateFilter{
				IsSystem: boolPtr(false),
			},
			expectMinimum: 1,
		},
		{
			name: "search query for SOC2",
			filter: ScheduleTemplateFilter{
				SearchQuery: "SOC2",
			},
			expectMinimum: 1,
		},
		{
			name: "search query case insensitive",
			filter: ScheduleTemplateFilter{
				SearchQuery: "wiz",
			},
			expectMinimum: 1,
		},
		{
			name: "combined filters",
			filter: ScheduleTemplateFilter{
				Category: "security",
				Tags:     []string{"hourly"},
			},
			expectMinimum: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templates, err := repo.List(ctx, tt.filter)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(templates), tt.expectMinimum,
				"expected at least %d templates, got %d", tt.expectMinimum, len(templates))

			if tt.checkCategory != "" {
				for _, tmpl := range templates {
					assert.Equal(t, tt.checkCategory, tmpl.Category,
						"expected category %s, got %s", tt.checkCategory, tmpl.Category)
				}
			}

			// Verify all templates have required fields
			for _, tmpl := range templates {
				assert.NotEmpty(t, tmpl.ID)
				assert.NotEmpty(t, tmpl.Name)
				assert.NotEmpty(t, tmpl.Category)
				assert.NotEmpty(t, tmpl.CronExpression)
				assert.NotEmpty(t, tmpl.Timezone)
			}
		})
	}
}

func TestScheduleTemplateRepository_GetByID(t *testing.T) {
	repo := &MockTemplateRepository{templates: createTestTemplates()}
	ctx := context.Background()

	tests := []struct {
		name      string
		id        string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "get existing template",
			id:        "tmpl-1",
			wantError: false,
		},
		{
			name:      "get non-existent template",
			id:        "00000000-0000-0000-0000-000000000000",
			wantError: true,
			errorMsg:  "template not found",
		},
		{
			name:      "get with empty ID",
			id:        "",
			wantError: true,
			errorMsg:  "template not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := repo.GetByID(ctx, tt.id)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, template)
			} else {
				require.NoError(t, err)
				require.NotNil(t, template)
				assert.Equal(t, tt.id, template.ID)
				assert.NotEmpty(t, template.Name)
				assert.NotEmpty(t, template.Category)
				assert.NotEmpty(t, template.CronExpression)
				assert.NotEmpty(t, template.Timezone)
			}
		})
	}
}

func TestScheduleTemplateRepository_GetByID_VerifyTemplateContent(t *testing.T) {
	repo := &MockTemplateRepository{templates: createTestTemplates()}
	ctx := context.Background()

	// Get the SOC2 template
	retrieved, err := repo.GetByID(ctx, "tmpl-1")
	require.NoError(t, err)
	assert.Equal(t, "SOC2 Daily Scan", retrieved.Name)
	assert.Equal(t, "compliance", retrieved.Category)
	assert.Equal(t, "0 2 * * *", retrieved.CronExpression)
	assert.Equal(t, "UTC", retrieved.Timezone)
	assert.True(t, retrieved.IsSystem)
	assert.Contains(t, []string(retrieved.Tags), "compliance")
	assert.Contains(t, []string(retrieved.Tags), "soc2")
}

// Helper boolPtr is defined in template_test.go
