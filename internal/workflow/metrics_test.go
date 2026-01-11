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

func TestGetTopFailuresOptimized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	tenantID := createTestTenant(t, db)

	tests := []struct {
		name          string
		setupData     func() ([]string, []int)
		limit         int
		wantCount     int
		validateOrder bool
	}{
		{
			name: "basic ordering by failure count",
			setupData: func() ([]string, []int) {
				workflow1ID := createTestWorkflowWithName(t, repo, tenantID, "High Failure Workflow")
				workflow2ID := createTestWorkflowWithName(t, repo, tenantID, "Medium Failure Workflow")
				workflow3ID := createTestWorkflowWithName(t, repo, tenantID, "Low Failure Workflow")

				// Workflow 1: 10 failures
				for i := 0; i < 10; i++ {
					createExecutionWithStatusAndError(t, repo, tenantID, workflow1ID, "failed",
						time.Now().Add(-time.Duration(i)*time.Minute), "Connection timeout")
				}

				// Workflow 2: 5 failures
				for i := 0; i < 5; i++ {
					createExecutionWithStatusAndError(t, repo, tenantID, workflow2ID, "failed",
						time.Now().Add(-time.Duration(i)*time.Minute), "API rate limit exceeded")
				}

				// Workflow 3: 2 failures
				for i := 0; i < 2; i++ {
					createExecutionWithStatusAndError(t, repo, tenantID, workflow3ID, "failed",
						time.Now().Add(-time.Duration(i)*time.Minute), "Invalid input")
				}

				return []string{workflow1ID, workflow2ID, workflow3ID}, []int{10, 5, 2}
			},
			limit:         10,
			wantCount:     3,
			validateOrder: true,
		},
		{
			name: "respects limit parameter",
			setupData: func() ([]string, []int) {
				workflowIDs := make([]string, 5)
				counts := make([]int, 5)
				for i := 0; i < 5; i++ {
					workflowID := createTestWorkflowWithName(t, repo, tenantID, "Workflow "+string(rune('A'+i)))
					workflowIDs[i] = workflowID
					failureCount := 10 - i
					counts[i] = failureCount
					for j := 0; j < failureCount; j++ {
						createExecutionWithStatusAndError(t, repo, tenantID, workflowID, "failed",
							time.Now().Add(-time.Duration(j)*time.Minute), "Test error")
					}
				}
				return workflowIDs, counts
			},
			limit:         3,
			wantCount:     3,
			validateOrder: true,
		},
		{
			name: "includes error preview from latest failure",
			setupData: func() ([]string, []int) {
				workflowID := createTestWorkflowWithName(t, repo, tenantID, "Error Preview Workflow")

				// Create failures with different error messages at different times
				createExecutionWithStatusAndError(t, repo, tenantID, workflowID, "failed",
					time.Now().Add(-10*time.Minute), "Old error message")
				createExecutionWithStatusAndError(t, repo, tenantID, workflowID, "failed",
					time.Now().Add(-5*time.Minute), "Recent error message")
				createExecutionWithStatusAndError(t, repo, tenantID, workflowID, "failed",
					time.Now().Add(-1*time.Minute), "Latest error message")

				return []string{workflowID}, []int{3}
			},
			limit:         10,
			wantCount:     1,
			validateOrder: false,
		},
		{
			name: "handles workflows with no error messages",
			setupData: func() ([]string, []int) {
				workflowID := createTestWorkflowWithName(t, repo, tenantID, "No Error Workflow")

				// Create failures without error messages
				for i := 0; i < 3; i++ {
					createExecutionWithStatus(t, repo, tenantID, workflowID, "failed",
						time.Now().Add(-time.Duration(i)*time.Minute))
				}

				return []string{workflowID}, []int{3}
			},
			limit:         10,
			wantCount:     1,
			validateOrder: false,
		},
		{
			name: "filters by date range",
			setupData: func() ([]string, []int) {
				workflowID := createTestWorkflowWithName(t, repo, tenantID, "Date Range Workflow")

				// Create failures in different time periods
				createExecutionWithStatus(t, repo, tenantID, workflowID, "failed",
					time.Now().Add(-48*time.Hour)) // Outside range
				createExecutionWithStatus(t, repo, tenantID, workflowID, "failed",
					time.Now().Add(-12*time.Hour)) // Within range
				createExecutionWithStatus(t, repo, tenantID, workflowID, "failed",
					time.Now().Add(-6*time.Hour)) // Within range

				return []string{workflowID}, []int{2} // Only 2 in range
			},
			limit:         10,
			wantCount:     1,
			validateOrder: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new tenant for test isolation
			testTenantID := createTestTenant(t, db)

			_, expectedCounts := tt.setupData()

			failures, err := repo.GetTopFailuresOptimized(context.Background(), testTenantID,
				time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour), tt.limit)
			require.NoError(t, err)
			assert.Len(t, failures, tt.wantCount)

			// Validate basic structure
			for _, failure := range failures {
				assert.NotEmpty(t, failure.WorkflowID)
				assert.NotEmpty(t, failure.WorkflowName)
				assert.Greater(t, failure.FailureCount, 0)
				assert.NotNil(t, failure.LastFailedAt)
			}

			// Validate ordering if required
			if tt.validateOrder {
				for i := 0; i < len(failures)-1; i++ {
					assert.GreaterOrEqual(t, failures[i].FailureCount, failures[i+1].FailureCount,
						"Failures should be ordered by count descending")
				}

				// Validate against expected counts
				for i, failure := range failures {
					if i < len(expectedCounts) {
						assert.Equal(t, expectedCounts[i], failure.FailureCount)
					}
				}
			}

			// Test specific to error preview
			if tt.name == "includes error preview from latest failure" {
				assert.NotNil(t, failures[0].ErrorPreview)
				assert.Equal(t, "Latest error message", *failures[0].ErrorPreview)
			}
		})
	}
}

func TestGetTopFailuresOptimized_EmptyResults(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	tenantID := createTestTenant(t, db)

	// No failures created
	failures, err := repo.GetTopFailuresOptimized(context.Background(), tenantID,
		time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour), 10)
	require.NoError(t, err)
	assert.Empty(t, failures)
}

func TestGetTopFailuresOptimized_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	tenant1ID := createTestTenant(t, db)
	tenant2ID := createTestTenant(t, db)

	// Create failures for tenant 1
	workflow1ID := createTestWorkflowWithName(t, repo, tenant1ID, "Tenant 1 Workflow")
	for i := 0; i < 5; i++ {
		createExecutionWithStatus(t, repo, tenant1ID, workflow1ID, "failed", time.Now())
	}

	// Create failures for tenant 2
	workflow2ID := createTestWorkflowWithName(t, repo, tenant2ID, "Tenant 2 Workflow")
	for i := 0; i < 10; i++ {
		createExecutionWithStatus(t, repo, tenant2ID, workflow2ID, "failed", time.Now())
	}

	// Query tenant 1
	failures1, err := repo.GetTopFailuresOptimized(context.Background(), tenant1ID,
		time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour), 10)
	require.NoError(t, err)
	require.Len(t, failures1, 1)
	assert.Equal(t, 5, failures1[0].FailureCount)

	// Query tenant 2
	failures2, err := repo.GetTopFailuresOptimized(context.Background(), tenant2ID,
		time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour), 10)
	require.NoError(t, err)
	require.Len(t, failures2, 1)
	assert.Equal(t, 10, failures2[0].FailureCount)
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
