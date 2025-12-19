package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScheduleTemplateFilter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filter  ScheduleTemplateFilter
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid empty filter",
			filter:  ScheduleTemplateFilter{},
			wantErr: false,
		},
		{
			name: "valid filter with category",
			filter: ScheduleTemplateFilter{
				Category: "compliance",
			},
			wantErr: false,
		},
		{
			name: "valid filter with tags",
			filter: ScheduleTemplateFilter{
				Tags: []string{"daily", "security"},
			},
			wantErr: false,
		},
		{
			name: "valid filter with search query",
			filter: ScheduleTemplateFilter{
				SearchQuery: "SOC2",
			},
			wantErr: false,
		},
		{
			name: "valid filter with system templates only",
			filter: ScheduleTemplateFilter{
				IsSystem: boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "valid filter with all fields",
			filter: ScheduleTemplateFilter{
				Category:    "security",
				Tags:        []string{"wiz"},
				SearchQuery: "scan",
				IsSystem:    boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "invalid category too long",
			filter: ScheduleTemplateFilter{
				Category: "this_is_a_very_long_category_name_that_exceeds_the_maximum_allowed_length_of_100_characters_and_should_fail_validation",
			},
			wantErr: true,
			errMsg:  "category must be 100 characters or less",
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

func TestApplyTemplateInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   ApplyTemplateInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with workflow ID only",
			input: ApplyTemplateInput{
				WorkflowID: "550e8400-e29b-41d4-a716-446655440000",
			},
			wantErr: false,
		},
		{
			name: "valid with custom name",
			input: ApplyTemplateInput{
				WorkflowID: "550e8400-e29b-41d4-a716-446655440000",
				Name:       stringPtr("My Custom Schedule"),
			},
			wantErr: false,
		},
		{
			name: "valid with custom timezone",
			input: ApplyTemplateInput{
				WorkflowID: "550e8400-e29b-41d4-a716-446655440000",
				Timezone:   stringPtr("America/New_York"),
			},
			wantErr: false,
		},
		{
			name: "valid with all fields",
			input: ApplyTemplateInput{
				WorkflowID: "550e8400-e29b-41d4-a716-446655440000",
				Name:       stringPtr("Daily Compliance Check"),
				Timezone:   stringPtr("America/Los_Angeles"),
			},
			wantErr: false,
		},
		{
			name: "invalid missing workflow ID",
			input: ApplyTemplateInput{
				Name: stringPtr("Test"),
			},
			wantErr: true,
			errMsg:  "workflow_id is required",
		},
		{
			name: "invalid empty workflow ID",
			input: ApplyTemplateInput{
				WorkflowID: "",
			},
			wantErr: true,
			errMsg:  "workflow_id is required",
		},
		{
			name: "invalid name too long",
			input: ApplyTemplateInput{
				WorkflowID: "550e8400-e29b-41d4-a716-446655440000",
				Name: stringPtr("This is an extremely long schedule name that exceeds the maximum allowed length of 255 characters and should fail validation because it is way too long for a schedule name and nobody should ever need a schedule name this long anyway but we need to test it!"),
			},
			wantErr: true,
			errMsg:  "name must be 255 characters or less",
		},
		{
			name: "invalid timezone too long",
			input: ApplyTemplateInput{
				WorkflowID: "550e8400-e29b-41d4-a716-446655440000",
				Timezone: stringPtr("This_Is_An_Invalid_Timezone_That_Exceeds_The_Maximum_Allowed_Length_Of_100_Characters_And_Should_Fail"),
			},
			wantErr: true,
			errMsg:  "timezone must be 100 characters or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

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

func TestScheduleTemplateCategory_Constants(t *testing.T) {
	assert.Equal(t, ScheduleTemplateCategory("frequent"), CategoryFrequent)
	assert.Equal(t, ScheduleTemplateCategory("daily"), CategoryDaily)
	assert.Equal(t, ScheduleTemplateCategory("weekly"), CategoryWeekly)
	assert.Equal(t, ScheduleTemplateCategory("monthly"), CategoryMonthly)
	assert.Equal(t, ScheduleTemplateCategory("business"), CategoryBusiness)
	assert.Equal(t, ScheduleTemplateCategory("compliance"), CategoryCompliance)
	assert.Equal(t, ScheduleTemplateCategory("security"), CategorySecurity)
	assert.Equal(t, ScheduleTemplateCategory("sync"), CategorySync)
	assert.Equal(t, ScheduleTemplateCategory("monitoring"), CategoryMonitoring)
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}
