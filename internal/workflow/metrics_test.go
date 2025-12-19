package workflow

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping integration test - set TEST_DATABASE_URL environment variable to run")
		return nil
	}

	db, err := sqlx.Connect("postgres", dbURL)
	require.NoError(t, err)
	return db
}

func createTestTenant(t *testing.T, db *sqlx.DB) string {
	tenantID := uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO tenants (id, name, subdomain, status, tier)
		VALUES ($1, $2, $3, $4, $5)
	`, tenantID, "Test Tenant", "test-"+tenantID[:8], "active", "free")
	require.NoError(t, err)
	return tenantID
}

func createTestWorkflow(t *testing.T, repo *Repository, tenantID string) string {
	return createTestWorkflowWithName(t, repo, tenantID, "Test Workflow")
}

func TestGetExecutionTrends(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	tenantID := createTestTenant(t, db)

	// Create test workflow
	workflowID := createTestWorkflow(t, repo, tenantID)

	// Create executions across different days
	now := time.Now().Truncate(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	// Today: 5 success, 2 failed
	createExecutionWithStatus(t, repo, tenantID, workflowID, "completed", now)
	createExecutionWithStatus(t, repo, tenantID, workflowID, "completed", now.Add(1*time.Hour))
	createExecutionWithStatus(t, repo, tenantID, workflowID, "completed", now.Add(2*time.Hour))
	createExecutionWithStatus(t, repo, tenantID, workflowID, "completed", now.Add(3*time.Hour))
	createExecutionWithStatus(t, repo, tenantID, workflowID, "completed", now.Add(4*time.Hour))
	createExecutionWithStatus(t, repo, tenantID, workflowID, "failed", now.Add(5*time.Hour))
	createExecutionWithStatus(t, repo, tenantID, workflowID, "failed", now.Add(6*time.Hour))

	// Yesterday: 3 success, 1 failed
	createExecutionWithStatus(t, repo, tenantID, workflowID, "completed", yesterday)
	createExecutionWithStatus(t, repo, tenantID, workflowID, "completed", yesterday.Add(2*time.Hour))
	createExecutionWithStatus(t, repo, tenantID, workflowID, "completed", yesterday.Add(4*time.Hour))
	createExecutionWithStatus(t, repo, tenantID, workflowID, "failed", yesterday.Add(6*time.Hour))

	// Two days ago: 2 success
	createExecutionWithStatus(t, repo, tenantID, workflowID, "completed", twoDaysAgo)
	createExecutionWithStatus(t, repo, tenantID, workflowID, "completed", twoDaysAgo.Add(3*time.Hour))

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		groupBy   string
		wantCount int
	}{
		{
			name:      "daily trend for 3 days",
			startDate: twoDaysAgo,
			endDate:   now.Add(24 * time.Hour),
			groupBy:   "day",
			wantCount: 3,
		},
		{
			name:      "hourly trend for today",
			startDate: now,
			endDate:   now.Add(24 * time.Hour),
			groupBy:   "hour",
			wantCount: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trends, err := repo.GetExecutionTrends(context.Background(), tenantID, tt.startDate, tt.endDate, tt.groupBy)
			require.NoError(t, err)
			assert.Len(t, trends, tt.wantCount)

			// Verify structure
			for _, trend := range trends {
				assert.NotEmpty(t, trend.Date)
				assert.GreaterOrEqual(t, trend.Count, 0)
				assert.GreaterOrEqual(t, trend.Success, 0)
				assert.GreaterOrEqual(t, trend.Failed, 0)
				assert.Equal(t, trend.Count, trend.Success+trend.Failed)
			}
		})
	}
}

func TestGetDurationStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	tenantID := createTestTenant(t, db)

	// Create two workflows
	workflow1ID := createTestWorkflow(t, repo, tenantID)
	workflow2ID := createTestWorkflowWithName(t, repo, tenantID, "Workflow 2")

	// Create executions with different durations for workflow 1
	// Durations: 100ms, 200ms, 300ms, 400ms, 500ms, 1000ms, 2000ms, 5000ms, 10000ms
	durations := []int{100, 200, 300, 400, 500, 1000, 2000, 5000, 10000}
	for _, d := range durations {
		createExecutionWithDuration(t, repo, tenantID, workflow1ID, d)
	}

	// Create executions for workflow 2
	createExecutionWithDuration(t, repo, tenantID, workflow2ID, 150)
	createExecutionWithDuration(t, repo, tenantID, workflow2ID, 250)

	stats, err := repo.GetDurationStats(context.Background(), tenantID, time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	require.Len(t, stats, 2)

	// Find workflow 1 stats
	var w1Stats *DurationStats
	for i := range stats {
		if stats[i].WorkflowID == workflow1ID {
			w1Stats = &stats[i]
			break
		}
	}
	require.NotNil(t, w1Stats)

	assert.Equal(t, workflow1ID, w1Stats.WorkflowID)
	assert.NotEmpty(t, w1Stats.WorkflowName)
	assert.Equal(t, 9, w1Stats.TotalRuns)
	assert.Greater(t, w1Stats.AvgDuration, 0.0)
	assert.Greater(t, w1Stats.P50Duration, 0.0)
	assert.Greater(t, w1Stats.P90Duration, 0.0)
	assert.Greater(t, w1Stats.P99Duration, 0.0)

	// Verify percentile ordering
	assert.LessOrEqual(t, w1Stats.P50Duration, w1Stats.P90Duration)
	assert.LessOrEqual(t, w1Stats.P90Duration, w1Stats.P99Duration)
}

func TestGetTopFailures(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	tenantID := createTestTenant(t, db)

	// Create three workflows with different failure rates
	workflow1ID := createTestWorkflowWithName(t, repo, tenantID, "High Failure Workflow")
	workflow2ID := createTestWorkflowWithName(t, repo, tenantID, "Medium Failure Workflow")
	workflow3ID := createTestWorkflowWithName(t, repo, tenantID, "Low Failure Workflow")

	// Workflow 1: 10 failures, 5 success
	for i := 0; i < 10; i++ {
		createExecutionWithStatus(t, repo, tenantID, workflow1ID, "failed", time.Now())
	}
	for i := 0; i < 5; i++ {
		createExecutionWithStatus(t, repo, tenantID, workflow1ID, "completed", time.Now())
	}

	// Workflow 2: 5 failures, 10 success
	for i := 0; i < 5; i++ {
		createExecutionWithStatus(t, repo, tenantID, workflow2ID, "failed", time.Now())
	}
	for i := 0; i < 10; i++ {
		createExecutionWithStatus(t, repo, tenantID, workflow2ID, "completed", time.Now())
	}

	// Workflow 3: 2 failures, 20 success
	for i := 0; i < 2; i++ {
		createExecutionWithStatus(t, repo, tenantID, workflow3ID, "failed", time.Now())
	}
	for i := 0; i < 20; i++ {
		createExecutionWithStatus(t, repo, tenantID, workflow3ID, "completed", time.Now())
	}

	failures, err := repo.GetTopFailures(context.Background(), tenantID, time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour), 10)
	require.NoError(t, err)
	require.Len(t, failures, 3)

	// Should be ordered by failure count (descending)
	assert.Equal(t, workflow1ID, failures[0].WorkflowID)
	assert.Equal(t, "High Failure Workflow", failures[0].WorkflowName)
	assert.Equal(t, 10, failures[0].FailureCount)
	assert.NotNil(t, failures[0].LastFailedAt)

	assert.Equal(t, workflow2ID, failures[1].WorkflowID)
	assert.Equal(t, 5, failures[1].FailureCount)

	assert.Equal(t, workflow3ID, failures[2].WorkflowID)
	assert.Equal(t, 2, failures[2].FailureCount)
}

func TestGetTriggerTypeBreakdown(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	tenantID := createTestTenant(t, db)

	workflowID := createTestWorkflow(t, repo, tenantID)

	// Create executions with different trigger types
	for i := 0; i < 15; i++ {
		createExecutionWithTriggerType(t, repo, tenantID, workflowID, "webhook")
	}
	for i := 0; i < 10; i++ {
		createExecutionWithTriggerType(t, repo, tenantID, workflowID, "schedule")
	}
	for i := 0; i < 5; i++ {
		createExecutionWithTriggerType(t, repo, tenantID, workflowID, "manual")
	}

	breakdown, err := repo.GetTriggerTypeBreakdown(context.Background(), tenantID, time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	require.Len(t, breakdown, 3)

	// Verify counts
	totalCount := 0
	for _, item := range breakdown {
		totalCount += item.Count
		assert.NotEmpty(t, item.TriggerType)
		assert.Greater(t, item.Count, 0)
		assert.Greater(t, item.Percentage, 0.0)
	}
	assert.Equal(t, 30, totalCount)

	// Verify percentages sum to ~100%
	totalPercentage := 0.0
	for _, item := range breakdown {
		totalPercentage += item.Percentage
	}
	assert.InDelta(t, 100.0, totalPercentage, 0.1)
}

// Helper functions

func createExecutionWithStatus(t *testing.T, repo *Repository, tenantID, workflowID, status string, createdAt time.Time) string {
	exec, err := repo.CreateExecution(context.Background(), tenantID, workflowID, 1, "manual", nil)
	require.NoError(t, err)

	// Update created_at and status
	_, err = repo.db.ExecContext(context.Background(),
		"UPDATE executions SET status = $1, created_at = $2, started_at = $2, completed_at = $3 WHERE id = $4",
		status, createdAt, createdAt.Add(1*time.Second), exec.ID)
	require.NoError(t, err)

	return exec.ID
}

func createExecutionWithDuration(t *testing.T, repo *Repository, tenantID, workflowID string, durationMs int) string {
	exec, err := repo.CreateExecution(context.Background(), tenantID, workflowID, 1, "manual", nil)
	require.NoError(t, err)

	startedAt := time.Now().Add(-1 * time.Hour)
	completedAt := startedAt.Add(time.Duration(durationMs) * time.Millisecond)

	_, err = repo.db.ExecContext(context.Background(),
		"UPDATE executions SET status = $1, started_at = $2, completed_at = $3 WHERE id = $4",
		"completed", startedAt, completedAt, exec.ID)
	require.NoError(t, err)

	return exec.ID
}

func createExecutionWithTriggerType(t *testing.T, repo *Repository, tenantID, workflowID, triggerType string) string {
	exec, err := repo.CreateExecution(context.Background(), tenantID, workflowID, 1, triggerType, nil)
	require.NoError(t, err)
	return exec.ID
}

func createTestWorkflowWithName(t *testing.T, repo *Repository, tenantID, name string) string {
	definition := json.RawMessage(`{
		"nodes": [{"id": "1", "type": "trigger:webhook", "position": {"x": 0, "y": 0}, "data": {"name": "Start"}}],
		"edges": []
	}`)

	workflow, err := repo.Create(context.Background(), tenantID, "test-user", CreateWorkflowInput{
		Name:        name,
		Description: "Test workflow",
		Definition:  definition,
	})
	require.NoError(t, err)
	return workflow.ID
}
