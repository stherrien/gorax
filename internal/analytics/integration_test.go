package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_TenantOverviewCalculation tests tenant overview statistics
func TestIntegration_TenantOverviewCalculation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	service := NewService(repo)

	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}

	// Mock repository returns pre-seeded data
	overview, err := service.GetTenantOverview(ctx, tenantID, timeRange)
	require.NoError(t, err)
	require.NotNil(t, overview)

	// Verify counts based on mock data
	assert.Equal(t, 7, overview.TotalExecutions)
	assert.Equal(t, 3, overview.SuccessfulExecutions)
	assert.Equal(t, 1, overview.FailedExecutions)
	assert.Equal(t, 1, overview.CancelledExecutions)
	assert.Equal(t, 1, overview.PendingExecutions)
	assert.Equal(t, 1, overview.RunningExecutions)
	t.Logf("✓ Execution counts: %d total, %d success, %d failed",
		overview.TotalExecutions, overview.SuccessfulExecutions, overview.FailedExecutions)

	// Verify success rate (3 success out of 7 total)
	expectedSuccessRate := 42.86
	assert.InDelta(t, expectedSuccessRate, overview.SuccessRate, 1.0)
	t.Logf("✓ Success rate: %.1f%%", overview.SuccessRate)

	// Verify workflow counts
	assert.Equal(t, 3, overview.TotalWorkflows)
	assert.Equal(t, 2, overview.ActiveWorkflows)
	t.Logf("✓ Total workflows: %d (active: %d)", overview.TotalWorkflows, overview.ActiveWorkflows)
}

// TestIntegration_WorkflowStatistics tests per-workflow statistics
func TestIntegration_WorkflowStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	service := NewService(repo)

	tenantID := "tenant-456"
	workflowID := "wf-test"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}

	stats, err := service.GetWorkflowStats(ctx, tenantID, workflowID, timeRange)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Verify execution counts from mock data
	assert.Equal(t, 6, stats.ExecutionCount)
	assert.Equal(t, 4, stats.SuccessCount)
	assert.Equal(t, 2, stats.FailureCount)
	t.Logf("✓ Counts - Total: %d, Success: %d, Failed: %d",
		stats.ExecutionCount, stats.SuccessCount, stats.FailureCount)

	// Verify success rate (4 out of 6 = 66.67%)
	expectedSuccessRate := 66.67
	assert.InDelta(t, expectedSuccessRate, stats.SuccessRate, 1.0)
	t.Logf("✓ Success rate: %.2f%%", stats.SuccessRate)

	// Verify duration statistics
	assert.Equal(t, int64(300), stats.MinDurationMs)
	assert.Equal(t, int64(2000), stats.MaxDurationMs)
	assert.Equal(t, int64(950), stats.AvgDurationMs)
	t.Logf("✓ Duration - Min: %dms, Max: %dms, Avg: %dms",
		stats.MinDurationMs, stats.MaxDurationMs, stats.AvgDurationMs)
}

// TestIntegration_ExecutionTrends tests time series data
func TestIntegration_ExecutionTrends(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	service := NewService(repo)

	tenantID := "tenant-789"
	now := time.Now()
	twoDaysAgo := now.Add(-48 * time.Hour)

	timeRange := TimeRange{
		StartDate: twoDaysAgo.Truncate(24 * time.Hour),
		EndDate:   now,
	}

	trends, err := service.GetExecutionTrends(ctx, tenantID, timeRange, GranularityDay)
	require.NoError(t, err)
	require.NotNil(t, trends)
	assert.Equal(t, GranularityDay, trends.Granularity)
	t.Logf("✓ Retrieved trends with granularity: %s", trends.Granularity)

	// Verify we have data points
	assert.NotEmpty(t, trends.DataPoints)
	t.Logf("✓ Got %d data points", len(trends.DataPoints))

	// Verify data point structure
	if len(trends.DataPoints) > 0 {
		point := trends.DataPoints[0]
		assert.NotZero(t, point.Timestamp)
		assert.GreaterOrEqual(t, point.ExecutionCount, 0)
		assert.GreaterOrEqual(t, point.SuccessCount, 0)
		assert.GreaterOrEqual(t, point.FailureCount, 0)
		t.Logf("✓ Sample data point - Executions: %d, Success: %d, Failed: %d",
			point.ExecutionCount, point.SuccessCount, point.FailureCount)
	}
}

// TestIntegration_TopWorkflows tests top workflows ranking
func TestIntegration_TopWorkflows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	service := NewService(repo)

	tenantID := "tenant-999"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}

	topWorkflows, err := service.GetTopWorkflows(ctx, tenantID, timeRange, 3)
	require.NoError(t, err)
	require.NotNil(t, topWorkflows)
	assert.LessOrEqual(t, len(topWorkflows.Workflows), 3)
	t.Logf("✓ Retrieved top %d workflows", len(topWorkflows.Workflows))

	// Verify ordering (by execution count, descending)
	if len(topWorkflows.Workflows) >= 2 {
		for i := 0; i < len(topWorkflows.Workflows)-1; i++ {
			assert.GreaterOrEqual(t,
				topWorkflows.Workflows[i].ExecutionCount,
				topWorkflows.Workflows[i+1].ExecutionCount,
				"workflows should be ordered by execution count")
		}
		t.Logf("✓ Workflows properly ordered by execution count")
	}

	// Verify top workflow
	if len(topWorkflows.Workflows) > 0 {
		topWorkflow := topWorkflows.Workflows[0]
		assert.Equal(t, "wf-C", topWorkflow.WorkflowID)
		assert.Equal(t, 200, topWorkflow.ExecutionCount)
		assert.InDelta(t, 90.0, topWorkflow.SuccessRate, 1.0)
		t.Logf("✓ Top workflow: %s with %d executions (%.1f%% success)",
			topWorkflow.WorkflowName, topWorkflow.ExecutionCount, topWorkflow.SuccessRate)
	}
}

// TestIntegration_ErrorBreakdown tests error analysis
func TestIntegration_ErrorBreakdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	service := NewService(repo)

	tenantID := "tenant-errors"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}

	breakdown, err := service.GetErrorBreakdown(ctx, tenantID, timeRange)
	require.NoError(t, err)
	require.NotNil(t, breakdown)

	// Verify total error count
	totalErrors := 28
	assert.Equal(t, totalErrors, breakdown.TotalErrors)
	t.Logf("✓ Total errors: %d", breakdown.TotalErrors)

	// Verify errors are present
	assert.NotEmpty(t, breakdown.ErrorsByType)
	t.Logf("✓ Found %d unique error types", len(breakdown.ErrorsByType))

	// Verify most common error
	if len(breakdown.ErrorsByType) > 0 {
		topError := breakdown.ErrorsByType[0]
		assert.Equal(t, "Connection timeout", topError.ErrorMessage)
		assert.Equal(t, 13, topError.ErrorCount)
		t.Logf("✓ Most common error: '%s' (%d occurrences)",
			topError.ErrorMessage, topError.ErrorCount)
	}
}

// TestIntegration_NodePerformance tests node-level statistics
func TestIntegration_NodePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	service := NewService(repo)

	tenantID := "tenant-nodes"
	workflowID := "wf-complex"

	performance, err := service.GetNodePerformance(ctx, tenantID, workflowID)
	require.NoError(t, err)
	require.NotNil(t, performance)

	// Verify workflow info
	assert.Equal(t, workflowID, performance.WorkflowID)
	assert.Equal(t, "Complex Workflow", performance.WorkflowName)

	// Verify node count
	assert.Len(t, performance.Nodes, 3)
	t.Logf("✓ Retrieved performance for %d nodes", len(performance.Nodes))

	// Verify specific node stats
	for _, nodeStats := range performance.Nodes {
		switch nodeStats.NodeID {
		case "node-1":
			assert.Equal(t, 3, nodeStats.ExecutionCount)
			assert.Equal(t, 3, nodeStats.SuccessCount)
			assert.Equal(t, 0, nodeStats.FailureCount)
			assert.InDelta(t, 100.0, nodeStats.SuccessRate, 1.0)
			t.Logf("✓ Node-1: %d executions, %.0f%% success",
				nodeStats.ExecutionCount, nodeStats.SuccessRate)

		case "node-2":
			assert.Equal(t, 4, nodeStats.ExecutionCount)
			assert.Equal(t, 3, nodeStats.SuccessCount)
			assert.Equal(t, 1, nodeStats.FailureCount)
			assert.InDelta(t, 75.0, nodeStats.SuccessRate, 1.0)
			t.Logf("✓ Node-2: %d executions, %.0f%% success",
				nodeStats.ExecutionCount, nodeStats.SuccessRate)

		case "node-3":
			assert.Equal(t, 5, nodeStats.ExecutionCount)
			assert.Equal(t, 5, nodeStats.SuccessCount)
			assert.Equal(t, 0, nodeStats.FailureCount)
			assert.InDelta(t, 100.0, nodeStats.SuccessRate, 1.0)
			t.Logf("✓ Node-3: %d executions, %.0f%% success",
				nodeStats.ExecutionCount, nodeStats.SuccessRate)
		}
	}
}

// TestIntegration_TimeRangeValidation tests validation of time ranges
func TestIntegration_TimeRangeValidation(t *testing.T) {
	ctx := context.Background()
	repo := setupTestRepository(t)
	service := NewService(repo)

	tests := []struct {
		name      string
		timeRange TimeRange
		expectErr bool
	}{
		{
			name: "valid time range",
			timeRange: TimeRange{
				StartDate: time.Now().Add(-24 * time.Hour),
				EndDate:   time.Now(),
			},
			expectErr: false,
		},
		{
			name: "zero start date",
			timeRange: TimeRange{
				StartDate: time.Time{},
				EndDate:   time.Now(),
			},
			expectErr: true,
		},
		{
			name: "zero end date",
			timeRange: TimeRange{
				StartDate: time.Now().Add(-24 * time.Hour),
				EndDate:   time.Time{},
			},
			expectErr: true,
		},
		{
			name: "end before start",
			timeRange: TimeRange{
				StartDate: time.Now(),
				EndDate:   time.Now().Add(-24 * time.Hour),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetTenantOverview(ctx, "tenant-1", tt.timeRange)
			if tt.expectErr {
				assert.Error(t, err)
				t.Logf("✓ Validation failed as expected: %v", err)
			} else {
				assert.NoError(t, err)
				t.Logf("✓ Validation passed")
			}
		})
	}
}

// TestIntegration_GranularityValidation tests granularity validation
func TestIntegration_GranularityValidation(t *testing.T) {
	ctx := context.Background()
	repo := setupTestRepository(t)
	service := NewService(repo)

	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}

	tests := []struct {
		name        string
		granularity Granularity
		expectErr   bool
	}{
		{name: "hourly", granularity: GranularityHour, expectErr: false},
		{name: "daily", granularity: GranularityDay, expectErr: false},
		{name: "weekly", granularity: GranularityWeek, expectErr: false},
		{name: "monthly", granularity: GranularityMonth, expectErr: false},
		{name: "invalid", granularity: "invalid", expectErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetExecutionTrends(ctx, "tenant-1", timeRange, tt.granularity)
			if tt.expectErr {
				assert.Error(t, err)
				t.Logf("✓ Invalid granularity rejected: %s", tt.granularity)
			} else {
				assert.NoError(t, err)
				t.Logf("✓ Granularity accepted: %s", tt.granularity)
			}
		})
	}
}

// Helper functions
func setupTestRepository(t *testing.T) AnalyticsRepository {
	t.Helper()
	return &mockRepository{}
}

// Mock repository implementing AnalyticsRepository interface
type mockRepository struct{}

func (m *mockRepository) GetWorkflowStats(ctx context.Context, tenantID, workflowID string, timeRange TimeRange) (*WorkflowStats, error) {
	return &WorkflowStats{
		WorkflowID:      workflowID,
		WorkflowName:    "Test Workflow",
		ExecutionCount:  6,
		SuccessCount:    4,
		FailureCount:    2,
		CancelledCount:  0,
		PendingCount:    0,
		RunningCount:    0,
		SuccessRate:     66.67,
		AvgDurationMs:   950,
		MinDurationMs:   300,
		MaxDurationMs:   2000,
		TotalDurationMs: 5700,
		LastExecutedAt:  time.Now(),
	}, nil
}

func (m *mockRepository) GetTenantOverview(ctx context.Context, tenantID string, timeRange TimeRange) (*TenantOverview, error) {
	return &TenantOverview{
		TotalExecutions:      7,
		SuccessfulExecutions: 3,
		FailedExecutions:     1,
		CancelledExecutions:  1,
		PendingExecutions:    1,
		RunningExecutions:    1,
		SuccessRate:          42.86,
		AvgDurationMs:        1320,
		ActiveWorkflows:      2,
		TotalWorkflows:       3,
	}, nil
}

func (m *mockRepository) GetExecutionTrends(ctx context.Context, tenantID string, timeRange TimeRange, granularity Granularity) (*ExecutionTrends, error) {
	now := time.Now()
	return &ExecutionTrends{
		Granularity: granularity,
		StartDate:   timeRange.StartDate,
		EndDate:     timeRange.EndDate,
		DataPoints: []TimeSeriesPoint{
			{
				Timestamp:      now.Add(-48 * time.Hour),
				ExecutionCount: 3,
				SuccessCount:   2,
				FailureCount:   1,
				SuccessRate:    66.67,
				AvgDurationMs:  1000,
			},
			{
				Timestamp:      now.Add(-24 * time.Hour),
				ExecutionCount: 3,
				SuccessCount:   1,
				FailureCount:   2,
				SuccessRate:    33.33,
				AvgDurationMs:  1233,
			},
			{
				Timestamp:      now,
				ExecutionCount: 4,
				SuccessCount:   3,
				FailureCount:   1,
				SuccessRate:    75.0,
				AvgDurationMs:  1125,
			},
		},
	}, nil
}

func (m *mockRepository) GetTopWorkflows(ctx context.Context, tenantID string, timeRange TimeRange, limit int) (*TopWorkflows, error) {
	return &TopWorkflows{
		Workflows: []TopWorkflow{
			{WorkflowID: "wf-C", WorkflowName: "Workflow C", ExecutionCount: 200, SuccessRate: 90.0, AvgDurationMs: 1500, LastExecutedAt: time.Now()},
			{WorkflowID: "wf-A", WorkflowName: "Workflow A", ExecutionCount: 100, SuccessRate: 95.0, AvgDurationMs: 1200, LastExecutedAt: time.Now()},
			{WorkflowID: "wf-E", WorkflowName: "Workflow E", ExecutionCount: 75, SuccessRate: 93.0, AvgDurationMs: 1100, LastExecutedAt: time.Now()},
		},
		Total: 5,
	}, nil
}

func (m *mockRepository) GetErrorBreakdown(ctx context.Context, tenantID string, timeRange TimeRange) (*ErrorBreakdown, error) {
	return &ErrorBreakdown{
		TotalErrors: 28,
		ErrorsByType: []ErrorInfo{
			{ErrorMessage: "Connection timeout", ErrorCount: 13, Percentage: 46.43, WorkflowID: "wf-1", WorkflowName: "API Workflow", LastOccurrence: time.Now()},
			{ErrorMessage: "Database connection failed", ErrorCount: 8, Percentage: 28.57, WorkflowID: "wf-2", WorkflowName: "Data Pipeline", LastOccurrence: time.Now()},
			{ErrorMessage: "Invalid response format", ErrorCount: 5, Percentage: 17.86, WorkflowID: "wf-1", WorkflowName: "API Workflow", LastOccurrence: time.Now()},
			{ErrorMessage: "SMTP authentication failed", ErrorCount: 2, Percentage: 7.14, WorkflowID: "wf-3", WorkflowName: "Email Sender", LastOccurrence: time.Now()},
		},
	}, nil
}

func (m *mockRepository) GetNodePerformance(ctx context.Context, tenantID, workflowID string) (*NodePerformance, error) {
	return &NodePerformance{
		WorkflowID:   workflowID,
		WorkflowName: "Complex Workflow",
		Nodes: []NodeStats{
			{NodeID: "node-1", NodeType: "trigger", ExecutionCount: 3, SuccessCount: 3, FailureCount: 0, SuccessRate: 100.0, AvgDurationMs: 55, MinDurationMs: 50, MaxDurationMs: 60},
			{NodeID: "node-2", NodeType: "http", ExecutionCount: 4, SuccessCount: 3, FailureCount: 1, SuccessRate: 75.0, AvgDurationMs: 205, MinDurationMs: 150, MaxDurationMs: 250},
			{NodeID: "node-3", NodeType: "transform", ExecutionCount: 5, SuccessCount: 5, FailureCount: 0, SuccessRate: 100.0, AvgDurationMs: 110, MinDurationMs: 100, MaxDurationMs: 120},
		},
	}, nil
}
