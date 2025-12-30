package analytics

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, mock
}

func TestGetWorkflowStats_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "workflow-1"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
	}

	rows := sqlmock.NewRows([]string{
		"workflow_id", "workflow_name", "execution_count", "success_count",
		"failure_count", "cancelled_count", "pending_count", "running_count",
		"success_rate", "avg_duration_ms", "min_duration_ms", "max_duration_ms",
		"total_duration_ms", "last_executed_at",
	}).AddRow(
		"workflow-1", "Test Workflow", 100, 90, 8, 2, 0, 0,
		0.90, 1500, 100, 5000, 150000,
		time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC),
	)

	mock.ExpectQuery("SELECT (.+) FROM executions").
		WithArgs(tenantID, workflowID, timeRange.StartDate, timeRange.EndDate).
		WillReturnRows(rows)

	stats, err := repo.GetWorkflowStats(ctx, tenantID, workflowID, timeRange)

	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, "workflow-1", stats.WorkflowID)
	assert.Equal(t, "Test Workflow", stats.WorkflowName)
	assert.Equal(t, 100, stats.ExecutionCount)
	assert.Equal(t, 90, stats.SuccessCount)
	assert.Equal(t, 8, stats.FailureCount)
	assert.Equal(t, 0.90, stats.SuccessRate)
	assert.Equal(t, int64(1500), stats.AvgDurationMs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWorkflowStats_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "nonexistent"
	timeRange := TimeRange{
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now(),
	}

	mock.ExpectQuery("SELECT (.+) FROM executions").
		WithArgs(tenantID, workflowID, timeRange.StartDate, timeRange.EndDate).
		WillReturnError(sql.ErrNoRows)

	stats, err := repo.GetWorkflowStats(ctx, tenantID, workflowID, timeRange)

	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTenantOverview_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
	}

	rows := sqlmock.NewRows([]string{
		"total_executions", "successful_executions", "failed_executions",
		"cancelled_executions", "pending_executions", "running_executions",
		"success_rate", "avg_duration_ms",
	}).AddRow(500, 450, 40, 10, 0, 0, 0.90, 2000)

	workflowRows := sqlmock.NewRows([]string{"active_workflows", "total_workflows"}).
		AddRow(10, 15)

	mock.ExpectQuery("SELECT (.+) FROM executions").
		WithArgs(tenantID, timeRange.StartDate, timeRange.EndDate).
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT (.+) FROM workflows").
		WithArgs(tenantID).
		WillReturnRows(workflowRows)

	overview, err := repo.GetTenantOverview(ctx, tenantID, timeRange)

	require.NoError(t, err)
	assert.NotNil(t, overview)
	assert.Equal(t, 500, overview.TotalExecutions)
	assert.Equal(t, 450, overview.SuccessfulExecutions)
	assert.Equal(t, 40, overview.FailedExecutions)
	assert.Equal(t, 0.90, overview.SuccessRate)
	assert.Equal(t, int64(2000), overview.AvgDurationMs)
	assert.Equal(t, 10, overview.ActiveWorkflows)
	assert.Equal(t, 15, overview.TotalWorkflows)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetExecutionTrends_HourlyGranularity(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC),
	}
	granularity := GranularityHour

	rows := sqlmock.NewRows([]string{
		"timestamp", "execution_count", "success_count", "failure_count",
		"success_rate", "avg_duration_ms",
	}).
		AddRow(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 10, 9, 1, 0.90, 1500).
		AddRow(time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC), 15, 14, 1, 0.93, 1600)

	mock.ExpectQuery("SELECT (.+) FROM executions").
		WithArgs(tenantID, timeRange.StartDate, timeRange.EndDate).
		WillReturnRows(rows)

	trends, err := repo.GetExecutionTrends(ctx, tenantID, timeRange, granularity)

	require.NoError(t, err)
	assert.NotNil(t, trends)
	assert.Equal(t, GranularityHour, trends.Granularity)
	assert.Len(t, trends.DataPoints, 2)
	assert.Equal(t, 10, trends.DataPoints[0].ExecutionCount)
	assert.Equal(t, 9, trends.DataPoints[0].SuccessCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetExecutionTrends_DailyGranularity(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 7, 23, 59, 59, 0, time.UTC),
	}
	granularity := GranularityDay

	rows := sqlmock.NewRows([]string{
		"timestamp", "execution_count", "success_count", "failure_count",
		"success_rate", "avg_duration_ms",
	}).
		AddRow(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 50, 45, 5, 0.90, 1500).
		AddRow(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), 60, 55, 5, 0.92, 1600)

	mock.ExpectQuery("SELECT (.+) FROM executions").
		WithArgs(tenantID, timeRange.StartDate, timeRange.EndDate).
		WillReturnRows(rows)

	trends, err := repo.GetExecutionTrends(ctx, tenantID, timeRange, granularity)

	require.NoError(t, err)
	assert.NotNil(t, trends)
	assert.Equal(t, GranularityDay, trends.Granularity)
	assert.Len(t, trends.DataPoints, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTopWorkflows_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
	}
	limit := 10

	rows := sqlmock.NewRows([]string{
		"workflow_id", "workflow_name", "execution_count", "success_rate",
		"avg_duration_ms", "last_executed_at",
	}).
		AddRow("workflow-1", "Top Workflow", 500, 0.95, 1000,
			time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC)).
		AddRow("workflow-2", "Second Workflow", 300, 0.90, 1500,
			time.Date(2024, 1, 31, 11, 0, 0, 0, time.UTC))

	countRows := sqlmock.NewRows([]string{"total"}).AddRow(2)

	mock.ExpectQuery("SELECT (.+) FROM executions").
		WithArgs(tenantID, timeRange.StartDate, timeRange.EndDate, limit).
		WillReturnRows(rows)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(tenantID, timeRange.StartDate, timeRange.EndDate).
		WillReturnRows(countRows)

	topWorkflows, err := repo.GetTopWorkflows(ctx, tenantID, timeRange, limit)

	require.NoError(t, err)
	assert.NotNil(t, topWorkflows)
	assert.Len(t, topWorkflows.Workflows, 2)
	assert.Equal(t, 2, topWorkflows.Total)
	assert.Equal(t, "Top Workflow", topWorkflows.Workflows[0].WorkflowName)
	assert.Equal(t, 500, topWorkflows.Workflows[0].ExecutionCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetErrorBreakdown_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"
	timeRange := TimeRange{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
	}

	errorRows := sqlmock.NewRows([]string{
		"error_message", "error_count", "workflow_id", "workflow_name",
		"last_occurrence", "percentage",
	}).
		AddRow("Connection timeout", 20, "workflow-1", "Test Workflow",
			time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC), 50.0).
		AddRow("Invalid input", 15, "workflow-2", "Another Workflow",
			time.Date(2024, 1, 31, 11, 0, 0, 0, time.UTC), 37.5)

	countRows := sqlmock.NewRows([]string{"total_errors"}).AddRow(40)

	mock.ExpectQuery("SELECT (.+) FROM executions").
		WithArgs(tenantID, timeRange.StartDate, timeRange.EndDate).
		WillReturnRows(errorRows)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs(tenantID, timeRange.StartDate, timeRange.EndDate).
		WillReturnRows(countRows)

	breakdown, err := repo.GetErrorBreakdown(ctx, tenantID, timeRange)

	require.NoError(t, err)
	assert.NotNil(t, breakdown)
	assert.Equal(t, 40, breakdown.TotalErrors)
	assert.Len(t, breakdown.ErrorsByType, 2)
	assert.Equal(t, "Connection timeout", breakdown.ErrorsByType[0].ErrorMessage)
	assert.Equal(t, 20, breakdown.ErrorsByType[0].ErrorCount)
	assert.Equal(t, 50.0, breakdown.ErrorsByType[0].Percentage)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNodePerformance_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "workflow-1"

	workflowRows := sqlmock.NewRows([]string{"name"}).AddRow("Test Workflow")

	nodeRows := sqlmock.NewRows([]string{
		"node_id", "node_type", "execution_count", "success_count",
		"failure_count", "success_rate", "avg_duration_ms",
		"min_duration_ms", "max_duration_ms",
	}).
		AddRow("node-1", "action:http", 100, 95, 5, 0.95, 500, 100, 2000).
		AddRow("node-2", "action:transform", 100, 98, 2, 0.98, 50, 10, 200)

	mock.ExpectQuery("SELECT name FROM workflows").
		WithArgs(workflowID, tenantID).
		WillReturnRows(workflowRows)

	mock.ExpectQuery("SELECT (.+) FROM step_executions").
		WithArgs(workflowID, tenantID).
		WillReturnRows(nodeRows)

	performance, err := repo.GetNodePerformance(ctx, tenantID, workflowID)

	require.NoError(t, err)
	assert.NotNil(t, performance)
	assert.Equal(t, "workflow-1", performance.WorkflowID)
	assert.Equal(t, "Test Workflow", performance.WorkflowName)
	assert.Len(t, performance.Nodes, 2)
	assert.Equal(t, "node-1", performance.Nodes[0].NodeID)
	assert.Equal(t, 100, performance.Nodes[0].ExecutionCount)
	assert.Equal(t, 0.95, performance.Nodes[0].SuccessRate)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNodePerformance_WorkflowNotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "tenant-123"
	workflowID := "nonexistent"

	mock.ExpectQuery("SELECT name FROM workflows").
		WithArgs(workflowID, tenantID).
		WillReturnError(sql.ErrNoRows)

	performance, err := repo.GetNodePerformance(ctx, tenantID, workflowID)

	assert.Error(t, err)
	assert.Nil(t, performance)
	assert.NoError(t, mock.ExpectationsWereMet())
}
