package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of analytics repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetWorkflowStats(ctx context.Context, tenantID, workflowID string, timeRange TimeRange) (*WorkflowStats, error) {
	args := m.Called(ctx, tenantID, workflowID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*WorkflowStats), args.Error(1)
}

func (m *MockRepository) GetTenantOverview(ctx context.Context, tenantID string, timeRange TimeRange) (*TenantOverview, error) {
	args := m.Called(ctx, tenantID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TenantOverview), args.Error(1)
}

func (m *MockRepository) GetExecutionTrends(ctx context.Context, tenantID string, timeRange TimeRange, granularity Granularity) (*ExecutionTrends, error) {
	args := m.Called(ctx, tenantID, timeRange, granularity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExecutionTrends), args.Error(1)
}

func (m *MockRepository) GetTopWorkflows(ctx context.Context, tenantID string, timeRange TimeRange, limit int) (*TopWorkflows, error) {
	args := m.Called(ctx, tenantID, timeRange, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TopWorkflows), args.Error(1)
}

func (m *MockRepository) GetErrorBreakdown(ctx context.Context, tenantID string, timeRange TimeRange) (*ErrorBreakdown, error) {
	args := m.Called(ctx, tenantID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ErrorBreakdown), args.Error(1)
}

func (m *MockRepository) GetNodePerformance(ctx context.Context, tenantID, workflowID string) (*NodePerformance, error) {
	args := m.Called(ctx, tenantID, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*NodePerformance), args.Error(1)
}

func TestServiceGetWorkflowStats_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	workflowID := "workflow-1"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
	}

	expectedStats := &WorkflowStats{
		WorkflowID:     "workflow-1",
		WorkflowName:   "Test Workflow",
		ExecutionCount: 100,
		SuccessCount:   90,
		FailureCount:   10,
		SuccessRate:    0.90,
		AvgDurationMs:  1500,
	}

	mockRepo.On("GetWorkflowStats", ctx, tenantID, workflowID, timeRange).
		Return(expectedStats, nil)

	stats, err := service.GetWorkflowStats(ctx, tenantID, workflowID, timeRange)

	require.NoError(t, err)
	assert.Equal(t, expectedStats, stats)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetWorkflowStats_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	workflowID := "workflow-1"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}

	mockRepo.On("GetWorkflowStats", ctx, tenantID, workflowID, timeRange).
		Return(nil, errors.New("database error"))

	stats, err := service.GetWorkflowStats(ctx, tenantID, workflowID, timeRange)

	assert.Error(t, err)
	assert.Nil(t, stats)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetTenantOverview_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
	}

	expectedOverview := &TenantOverview{
		TotalExecutions:      500,
		SuccessfulExecutions: 450,
		FailedExecutions:     50,
		SuccessRate:          0.90,
		AvgDurationMs:        2000,
		ActiveWorkflows:      10,
		TotalWorkflows:       15,
	}

	mockRepo.On("GetTenantOverview", ctx, tenantID, timeRange).
		Return(expectedOverview, nil)

	overview, err := service.GetTenantOverview(ctx, tenantID, timeRange)

	require.NoError(t, err)
	assert.Equal(t, expectedOverview, overview)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetExecutionTrends_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC),
	}
	granularity := GranularityDay

	expectedTrends := &ExecutionTrends{
		Granularity: GranularityDay,
		StartDate:   timeRange.StartDate,
		EndDate:     timeRange.EndDate,
		DataPoints: []TimeSeriesPoint{
			{
				Timestamp:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				ExecutionCount: 50,
				SuccessCount:   45,
				FailureCount:   5,
				SuccessRate:    0.90,
				AvgDurationMs:  1500,
			},
		},
	}

	mockRepo.On("GetExecutionTrends", ctx, tenantID, timeRange, granularity).
		Return(expectedTrends, nil)

	trends, err := service.GetExecutionTrends(ctx, tenantID, timeRange, granularity)

	require.NoError(t, err)
	assert.Equal(t, expectedTrends, trends)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetExecutionTrends_InvalidGranularity(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}
	granularity := Granularity("invalid")

	trends, err := service.GetExecutionTrends(ctx, tenantID, timeRange, granularity)

	assert.Error(t, err)
	assert.Nil(t, trends)
	assert.Contains(t, err.Error(), "invalid granularity")
}

func TestServiceGetTopWorkflows_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
	}
	limit := 10

	expectedWorkflows := &TopWorkflows{
		Workflows: []TopWorkflow{
			{
				WorkflowID:     "workflow-1",
				WorkflowName:   "Top Workflow",
				ExecutionCount: 500,
				SuccessRate:    0.95,
				AvgDurationMs:  1000,
			},
		},
		Total: 1,
	}

	mockRepo.On("GetTopWorkflows", ctx, tenantID, timeRange, limit).
		Return(expectedWorkflows, nil)

	workflows, err := service.GetTopWorkflows(ctx, tenantID, timeRange, limit)

	require.NoError(t, err)
	assert.Equal(t, expectedWorkflows, workflows)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetTopWorkflows_DefaultLimit(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}
	limit := 0

	expectedWorkflows := &TopWorkflows{
		Workflows: []TopWorkflow{},
		Total:     0,
	}

	mockRepo.On("GetTopWorkflows", ctx, tenantID, timeRange, 10).
		Return(expectedWorkflows, nil)

	workflows, err := service.GetTopWorkflows(ctx, tenantID, timeRange, limit)

	require.NoError(t, err)
	assert.Equal(t, expectedWorkflows, workflows)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetErrorBreakdown_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
	}

	expectedBreakdown := &ErrorBreakdown{
		TotalErrors: 40,
		ErrorsByType: []ErrorInfo{
			{
				ErrorMessage: "Connection timeout",
				ErrorCount:   20,
				WorkflowID:   "workflow-1",
				WorkflowName: "Test Workflow",
				Percentage:   50.0,
			},
		},
	}

	mockRepo.On("GetErrorBreakdown", ctx, tenantID, timeRange).
		Return(expectedBreakdown, nil)

	breakdown, err := service.GetErrorBreakdown(ctx, tenantID, timeRange)

	require.NoError(t, err)
	assert.Equal(t, expectedBreakdown, breakdown)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetNodePerformance_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	workflowID := "workflow-1"

	expectedPerformance := &NodePerformance{
		WorkflowID:   "workflow-1",
		WorkflowName: "Test Workflow",
		Nodes: []NodeStats{
			{
				NodeID:         "node-1",
				NodeType:       "action:http",
				ExecutionCount: 100,
				SuccessCount:   95,
				FailureCount:   5,
				SuccessRate:    0.95,
				AvgDurationMs:  500,
			},
		},
	}

	mockRepo.On("GetNodePerformance", ctx, tenantID, workflowID).
		Return(expectedPerformance, nil)

	performance, err := service.GetNodePerformance(ctx, tenantID, workflowID)

	require.NoError(t, err)
	assert.Equal(t, expectedPerformance, performance)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetNodePerformance_WorkflowNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	workflowID := "nonexistent"

	mockRepo.On("GetNodePerformance", ctx, tenantID, workflowID).
		Return(nil, ErrNotFound)

	performance, err := service.GetNodePerformance(ctx, tenantID, workflowID)

	assert.Error(t, err)
	assert.Nil(t, performance)
	mockRepo.AssertExpectations(t)
}

// Edge Case Tests

func TestServiceGetWorkflowStats_InvalidTimeRange_ZeroStartDate(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	workflowID := "workflow-1"
	timeRange := TimeRange{
		StartDate: time.Time{}, // Zero value
		EndDate:   time.Now(),
	}

	stats, err := service.GetWorkflowStats(ctx, tenantID, workflowID, timeRange)

	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "start date and end date are required")
}

func TestServiceGetWorkflowStats_InvalidTimeRange_ZeroEndDate(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	workflowID := "workflow-1"
	timeRange := TimeRange{
		StartDate: time.Now(),
		EndDate:   time.Time{}, // Zero value
	}

	stats, err := service.GetWorkflowStats(ctx, tenantID, workflowID, timeRange)

	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "start date and end date are required")
}

func TestServiceGetWorkflowStats_InvalidTimeRange_EndBeforeStart(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	workflowID := "workflow-1"
	timeRange := TimeRange{
		StartDate: time.Now(),
		EndDate:   time.Now().Add(-24 * time.Hour), // End before start
	}

	stats, err := service.GetWorkflowStats(ctx, tenantID, workflowID, timeRange)

	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "end date must be after start date")
}

func TestServiceGetTenantOverview_InvalidTimeRange(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Time{},
		EndDate:   time.Time{},
	}

	overview, err := service.GetTenantOverview(ctx, tenantID, timeRange)

	assert.Error(t, err)
	assert.Nil(t, overview)
	assert.Contains(t, err.Error(), "start date and end date are required")
}

func TestServiceGetTenantOverview_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}

	mockRepo.On("GetTenantOverview", ctx, tenantID, timeRange).
		Return(nil, errors.New("database error"))

	overview, err := service.GetTenantOverview(ctx, tenantID, timeRange)

	assert.Error(t, err)
	assert.Nil(t, overview)
	assert.Contains(t, err.Error(), "get tenant overview")
	mockRepo.AssertExpectations(t)
}

func TestServiceGetExecutionTrends_InvalidTimeRange(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Time{},
		EndDate:   time.Now(),
	}
	granularity := GranularityDay

	trends, err := service.GetExecutionTrends(ctx, tenantID, timeRange, granularity)

	assert.Error(t, err)
	assert.Nil(t, trends)
	assert.Contains(t, err.Error(), "start date and end date are required")
}

func TestServiceGetExecutionTrends_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}
	granularity := GranularityDay

	mockRepo.On("GetExecutionTrends", ctx, tenantID, timeRange, granularity).
		Return(nil, errors.New("database error"))

	trends, err := service.GetExecutionTrends(ctx, tenantID, timeRange, granularity)

	assert.Error(t, err)
	assert.Nil(t, trends)
	assert.Contains(t, err.Error(), "get execution trends")
	mockRepo.AssertExpectations(t)
}

func TestServiceGetTopWorkflows_NegativeLimit(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}
	limit := -5 // Negative limit should default to 10

	expectedWorkflows := &TopWorkflows{
		Workflows: []TopWorkflow{},
		Total:     0,
	}

	mockRepo.On("GetTopWorkflows", ctx, tenantID, timeRange, 10).
		Return(expectedWorkflows, nil)

	workflows, err := service.GetTopWorkflows(ctx, tenantID, timeRange, limit)

	require.NoError(t, err)
	assert.Equal(t, expectedWorkflows, workflows)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetTopWorkflows_ExcessiveLimit(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}
	limit := 500 // Excessive limit should be capped at 100

	expectedWorkflows := &TopWorkflows{
		Workflows: []TopWorkflow{},
		Total:     0,
	}

	mockRepo.On("GetTopWorkflows", ctx, tenantID, timeRange, 100).
		Return(expectedWorkflows, nil)

	workflows, err := service.GetTopWorkflows(ctx, tenantID, timeRange, limit)

	require.NoError(t, err)
	assert.Equal(t, expectedWorkflows, workflows)
	mockRepo.AssertExpectations(t)
}

func TestServiceGetTopWorkflows_InvalidTimeRange(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Now(),
		EndDate:   time.Now().Add(-24 * time.Hour), // End before start
	}
	limit := 10

	workflows, err := service.GetTopWorkflows(ctx, tenantID, timeRange, limit)

	assert.Error(t, err)
	assert.Nil(t, workflows)
	assert.Contains(t, err.Error(), "end date must be after start date")
}

func TestServiceGetTopWorkflows_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}
	limit := 10

	mockRepo.On("GetTopWorkflows", ctx, tenantID, timeRange, limit).
		Return(nil, errors.New("database error"))

	workflows, err := service.GetTopWorkflows(ctx, tenantID, timeRange, limit)

	assert.Error(t, err)
	assert.Nil(t, workflows)
	assert.Contains(t, err.Error(), "get top workflows")
	mockRepo.AssertExpectations(t)
}

func TestServiceGetErrorBreakdown_InvalidTimeRange(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Time{},
		EndDate:   time.Time{},
	}

	breakdown, err := service.GetErrorBreakdown(ctx, tenantID, timeRange)

	assert.Error(t, err)
	assert.Nil(t, breakdown)
	assert.Contains(t, err.Error(), "start date and end date are required")
}

func TestServiceGetErrorBreakdown_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}

	mockRepo.On("GetErrorBreakdown", ctx, tenantID, timeRange).
		Return(nil, errors.New("database error"))

	breakdown, err := service.GetErrorBreakdown(ctx, tenantID, timeRange)

	assert.Error(t, err)
	assert.Nil(t, breakdown)
	assert.Contains(t, err.Error(), "get error breakdown")
	mockRepo.AssertExpectations(t)
}

func TestServiceGetNodePerformance_RepositoryError(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	tenantID := "tenant-123"
	workflowID := "workflow-1"

	mockRepo.On("GetNodePerformance", ctx, tenantID, workflowID).
		Return(nil, errors.New("database error"))

	performance, err := service.GetNodePerformance(ctx, tenantID, workflowID)

	assert.Error(t, err)
	assert.Nil(t, performance)
	assert.Contains(t, err.Error(), "get node performance")
	mockRepo.AssertExpectations(t)
}

func TestValidateGranularity_AllValidValues(t *testing.T) {
	tests := []struct {
		name        string
		granularity Granularity
		wantErr     bool
	}{
		{"hour", GranularityHour, false},
		{"day", GranularityDay, false},
		{"week", GranularityWeek, false},
		{"month", GranularityMonth, false},
		{"invalid", Granularity("invalid"), true},
		{"empty", Granularity(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGranularity(tt.granularity)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTimeRange_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name      string
		timeRange TimeRange
		wantErr   bool
		errMsg    string
	}{
		{
			name: "same start and end time (valid edge case)",
			timeRange: TimeRange{
				StartDate: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "end date 1 nanosecond after start (valid)",
			timeRange: TimeRange{
				StartDate: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 1, 12, 0, 0, 1, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "end date 1 nanosecond before start (invalid)",
			timeRange: TimeRange{
				StartDate: time.Date(2024, 1, 1, 12, 0, 0, 1, time.UTC),
				EndDate:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			wantErr: true,
			errMsg:  "end date must be after start date",
		},
		{
			name: "very large time range (1000 years)",
			timeRange: TimeRange{
				StartDate: time.Date(1024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTimeRange(tt.timeRange)
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
