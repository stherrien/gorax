package schedule

import (
	"context"
	"testing"
	"time"
)

// TestValidateCronExpression tests cron expression validation
func TestValidateCronExpression(t *testing.T) {
	service := NewService(nil, nil)

	tests := []struct {
		name       string
		expression string
		wantErr    bool
	}{
		{
			name:       "valid standard cron",
			expression: "0 */2 * * *",
			wantErr:    false,
		},
		{
			name:       "valid cron with seconds",
			expression: "0 0 */2 * * *",
			wantErr:    false,
		},
		{
			name:       "valid daily at noon",
			expression: "0 12 * * *",
			wantErr:    false,
		},
		{
			name:       "valid every minute",
			expression: "* * * * *",
			wantErr:    false,
		},
		{
			name:       "valid descriptor @daily",
			expression: "@daily",
			wantErr:    false,
		},
		{
			name:       "valid descriptor @hourly",
			expression: "@hourly",
			wantErr:    false,
		},
		{
			name:       "invalid cron expression",
			expression: "invalid",
			wantErr:    true,
		},
		{
			name:       "empty expression",
			expression: "",
			wantErr:    true,
		},
		{
			name:       "too many fields",
			expression: "0 0 0 0 0 0 0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateCronExpression(tt.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCronExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCalculateNextRun tests next run time calculation
func TestCalculateNextRun(t *testing.T) {
	service := NewService(nil, nil)

	tests := []struct {
		name       string
		expression string
		timezone   string
		wantErr    bool
	}{
		{
			name:       "calculate next run UTC",
			expression: "0 12 * * *",
			timezone:   "UTC",
			wantErr:    false,
		},
		{
			name:       "calculate next run EST",
			expression: "0 9 * * *",
			timezone:   "America/New_York",
			wantErr:    false,
		},
		{
			name:       "calculate next run PST",
			expression: "0 0 * * *",
			timezone:   "America/Los_Angeles",
			wantErr:    false,
		},
		{
			name:       "calculate next run with descriptor",
			expression: "@hourly",
			timezone:   "UTC",
			wantErr:    false,
		},
		{
			name:       "invalid timezone falls back to UTC",
			expression: "0 12 * * *",
			timezone:   "Invalid/Timezone",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextRun, err := service.calculateNextRun(tt.expression, tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateNextRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if nextRun.IsZero() {
					t.Error("calculateNextRun() returned zero time")
				}
				if !nextRun.After(time.Now()) {
					t.Error("calculateNextRun() should return future time")
				}
			}
		})
	}
}

// TestParseNextRunTime tests the ParseNextRunTime public method
func TestParseNextRunTime(t *testing.T) {
	service := NewService(nil, nil)

	// Test valid expression
	nextRun, err := service.ParseNextRunTime("0 12 * * *", "UTC")
	if err != nil {
		t.Errorf("ParseNextRunTime() error = %v", err)
	}
	if nextRun.IsZero() {
		t.Error("ParseNextRunTime() returned zero time")
	}
	if !nextRun.After(time.Now()) {
		t.Error("ParseNextRunTime() should return future time")
	}

	// Test invalid expression
	_, err = service.ParseNextRunTime("invalid", "UTC")
	if err == nil {
		t.Error("ParseNextRunTime() should return error for invalid expression")
	}
}

// MockWorkflowService for testing
type MockWorkflowService struct {
	getByIDFunc func(ctx context.Context, tenantID, id string) (interface{}, error)
}

func (m *MockWorkflowService) GetByID(ctx context.Context, tenantID, id string) (interface{}, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, tenantID, id)
	}
	return &struct{}{}, nil
}

// TestCreateWithWorkflowValidation tests schedule creation with workflow validation
func TestCreateWithWorkflowValidation(t *testing.T) {
	// Note: This test is skipped because it requires a real repository
	// In a production environment, you would use integration tests with a test database
	t.Skip("Skipping test that requires database repository")
}
