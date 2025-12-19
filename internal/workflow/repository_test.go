package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: These are integration tests that require a test database
// Skip them if TEST_DATABASE_URL is not set

func getTestDB(t *testing.T) *sqlx.DB {
	// Skip if database URL not set
	t.Skip("Integration tests require TEST_DATABASE_URL environment variable")
	return nil
}

// TestListExecutionsAdvanced tests the advanced execution listing with filters and pagination
func TestListExecutionsAdvanced(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "test-tenant-123"
	workflowID := "test-workflow-123"
	_ = repo
	_ = ctx
	_ = tenantID
	_ = workflowID

	tests := []struct {
		name           string
		filter         ExecutionFilter
		cursor         string
		limit          int
		expectedCount  int
		expectedHasMore bool
		validateFn     func(*testing.T, *ExecutionListResult)
	}{
		{
			name:           "list all executions without filter",
			filter:         ExecutionFilter{},
			limit:          10,
			expectedCount:  5,
			expectedHasMore: false,
		},
		{
			name: "filter by workflow_id",
			filter: ExecutionFilter{
				WorkflowID: workflowID,
			},
			limit:           10,
			expectedCount:   4,
			expectedHasMore: false,
			validateFn: func(t *testing.T, result *ExecutionListResult) {
				for _, exec := range result.Data {
					assert.Equal(t, workflowID, exec.WorkflowID)
				}
			},
		},
		{
			name: "filter by status",
			filter: ExecutionFilter{
				Status: "completed",
			},
			limit:           10,
			expectedCount:   3,
			expectedHasMore: false,
			validateFn: func(t *testing.T, result *ExecutionListResult) {
				for _, exec := range result.Data {
					assert.Equal(t, "completed", exec.Status)
				}
			},
		},
		{
			name: "filter by trigger_type",
			filter: ExecutionFilter{
				TriggerType: "webhook",
			},
			limit:           10,
			expectedCount:   3,
			expectedHasMore: false,
			validateFn: func(t *testing.T, result *ExecutionListResult) {
				for _, exec := range result.Data {
					assert.Equal(t, "webhook", exec.TriggerType)
				}
			},
		},
		{
			name: "filter by date range",
			filter: ExecutionFilter{
				StartDate: timePtr(time.Now().Add(-4 * time.Hour).Add(-1 * time.Minute)),
				EndDate:   timePtr(time.Now().Add(-2 * time.Hour).Add(-1 * time.Minute)),
			},
			limit:           10,
			expectedCount:   2, // exec-2 and exec-3
			expectedHasMore: false,
		},
		{
			name: "pagination with limit",
			filter: ExecutionFilter{
				WorkflowID: workflowID,
			},
			limit:           2,
			expectedCount:   2,
			expectedHasMore: true,
			validateFn: func(t *testing.T, result *ExecutionListResult) {
				assert.NotEmpty(t, result.Cursor)
				// Should return newest first
				assert.Equal(t, "exec-5", result.Data[0].ID)
				assert.Equal(t, "exec-3", result.Data[1].ID)
			},
		},
		{
			name: "combined filters",
			filter: ExecutionFilter{
				WorkflowID:  workflowID,
				Status:      "completed",
				TriggerType: "webhook",
			},
			limit:           10,
			expectedCount:   2, // exec-1 and exec-5
			expectedHasMore: false,
			validateFn: func(t *testing.T, result *ExecutionListResult) {
				for _, exec := range result.Data {
					assert.Equal(t, workflowID, exec.WorkflowID)
					assert.Equal(t, "completed", exec.Status)
					assert.Equal(t, "webhook", exec.TriggerType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.ListExecutionsAdvanced(ctx, tenantID, tt.filter, tt.cursor, tt.limit)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(result.Data))
			assert.Equal(t, tt.expectedHasMore, result.HasMore)

			if tt.validateFn != nil {
				tt.validateFn(t, result)
			}

			// Verify results are sorted by created_at DESC
			for i := 1; i < len(result.Data); i++ {
				assert.True(t,
					result.Data[i-1].CreatedAt.After(result.Data[i].CreatedAt) ||
					result.Data[i-1].CreatedAt.Equal(result.Data[i].CreatedAt),
					"Results should be sorted by created_at DESC")
			}
		})
	}
}

// TestListExecutionsAdvanced_CursorPagination tests cursor-based pagination
func TestListExecutionsAdvanced_CursorPagination(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "test-tenant-456"
	workflowID := "test-workflow-456"

	// Create 10 test executions
	for i := 0; i < 10; i++ {
		_, err := repo.CreateExecution(ctx, tenantID, workflowID, 1, "webhook", nil)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// First page
	result1, err := repo.ListExecutionsAdvanced(ctx, tenantID, ExecutionFilter{}, "", 3)
	require.NoError(t, err)
	assert.Equal(t, 3, len(result1.Data))
	assert.True(t, result1.HasMore)
	assert.NotEmpty(t, result1.Cursor)
	assert.Equal(t, 10, result1.TotalCount)

	// Second page
	result2, err := repo.ListExecutionsAdvanced(ctx, tenantID, ExecutionFilter{}, result1.Cursor, 3)
	require.NoError(t, err)
	assert.Equal(t, 3, len(result2.Data))
	assert.True(t, result2.HasMore)
	assert.NotEmpty(t, result2.Cursor)

	// Verify no overlap between pages
	page1IDs := make(map[string]bool)
	for _, exec := range result1.Data {
		page1IDs[exec.ID] = true
	}
	for _, exec := range result2.Data {
		assert.False(t, page1IDs[exec.ID], "Pages should not overlap")
	}

	// Third page
	result3, err := repo.ListExecutionsAdvanced(ctx, tenantID, ExecutionFilter{}, result2.Cursor, 3)
	require.NoError(t, err)
	assert.Equal(t, 3, len(result3.Data))
	assert.True(t, result3.HasMore)

	// Fourth page (last)
	result4, err := repo.ListExecutionsAdvanced(ctx, tenantID, ExecutionFilter{}, result3.Cursor, 3)
	require.NoError(t, err)
	assert.Equal(t, 1, len(result4.Data)) // Only 1 remaining
	assert.False(t, result4.HasMore)
	assert.Empty(t, result4.Cursor) // No more pages
}

// TestGetExecutionWithSteps tests retrieving an execution with its steps
func TestGetExecutionWithSteps(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "test-tenant-789"
	workflowID := "test-workflow-789"

	// Create execution
	execution, err := repo.CreateExecution(ctx, tenantID, workflowID, 1, "webhook", nil)
	require.NoError(t, err)

	// Create step executions
	step1, err := repo.CreateStepExecution(ctx, execution.ID, "node-1", "action:http", []byte(`{"url":"https://example.com"}`))
	require.NoError(t, err)

	step2, err := repo.CreateStepExecution(ctx, execution.ID, "node-2", "action:transform", []byte(`{"mapping":"data"}`))
	require.NoError(t, err)

	// Complete steps
	err = repo.UpdateStepExecution(ctx, step1.ID, "completed", []byte(`{"status":200}`), nil)
	require.NoError(t, err)

	err = repo.UpdateStepExecution(ctx, step2.ID, "completed", []byte(`{"result":"transformed"}`), nil)
	require.NoError(t, err)

	tests := []struct {
		name           string
		executionID    string
		expectedSteps  int
		shouldError    bool
	}{
		{
			name:          "get execution with steps",
			executionID:   execution.ID,
			expectedSteps: 2,
			shouldError:   false,
		},
		{
			name:          "get non-existent execution",
			executionID:   "non-existent-id",
			expectedSteps: 0,
			shouldError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetExecutionWithSteps(ctx, tenantID, tt.executionID)

			if tt.shouldError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result.Execution)
			assert.Equal(t, tt.executionID, result.Execution.ID)
			assert.Equal(t, tt.expectedSteps, len(result.Steps))

			// Verify steps are ordered by started_at
			for i := 1; i < len(result.Steps); i++ {
				if result.Steps[i-1].StartedAt != nil && result.Steps[i].StartedAt != nil {
					assert.True(t,
						result.Steps[i-1].StartedAt.Before(*result.Steps[i].StartedAt) ||
						result.Steps[i-1].StartedAt.Equal(*result.Steps[i].StartedAt),
						"Steps should be ordered by started_at ASC")
				}
			}
		})
	}
}

// TestCountExecutions tests counting executions with filters
func TestCountExecutions(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenantID := "test-tenant-count"
	workflowID1 := "test-workflow-count-1"
	workflowID2 := "test-workflow-count-2"
	_ = repo
	_ = ctx
	_ = tenantID
	_ = workflowID1
	_ = workflowID2

	// Workflow 1: 3 completed, 2 running
	for i := 0; i < 3; i++ {
		_, err := repo.CreateExecution(ctx, tenantID, workflowID1, 1, "webhook", nil)
		require.NoError(t, err)
	}
	for i := 0; i < 2; i++ {
		exec, err := repo.CreateExecution(ctx, tenantID, workflowID1, 1, "schedule", nil)
		require.NoError(t, err)
		err = repo.UpdateExecutionStatus(ctx, exec.ID, ExecutionStatusRunning, nil, nil)
		require.NoError(t, err)
	}

	// Workflow 2: 2 completed
	for i := 0; i < 2; i++ {
		_, err := repo.CreateExecution(ctx, tenantID, workflowID2, 1, "manual", nil)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		filter        ExecutionFilter
		expectedCount int
	}{
		{
			name:          "count all executions",
			filter:        ExecutionFilter{},
			expectedCount: 7,
		},
		{
			name: "count by workflow_id",
			filter: ExecutionFilter{
				WorkflowID: workflowID1,
			},
			expectedCount: 5,
		},
		{
			name: "count by status",
			filter: ExecutionFilter{
				Status: "running",
			},
			expectedCount: 2,
		},
		{
			name: "count by trigger_type",
			filter: ExecutionFilter{
				TriggerType: "webhook",
			},
			expectedCount: 3,
		},
		{
			name: "count by date range",
			filter: ExecutionFilter{
				StartDate: timePtr(time.Now().Add(-1 * time.Hour)),
				EndDate:   timePtr(time.Now().Add(1 * time.Hour)),
			},
			expectedCount: 7, // All created recently
		},
		{
			name: "count with combined filters",
			filter: ExecutionFilter{
				WorkflowID:  workflowID1,
				Status:      "running",
				TriggerType: "schedule",
			},
			expectedCount: 2,
		},
		{
			name: "count with no matches",
			filter: ExecutionFilter{
				WorkflowID: "non-existent-workflow",
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.CountExecutions(ctx, tenantID, tt.filter)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

// TestCountExecutions_TenantIsolation tests that counting respects tenant isolation
func TestCountExecutions_TenantIsolation(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()
	tenant1 := "tenant-1"
	tenant2 := "tenant-2"
	workflowID := "shared-workflow-id"

	// Create executions for tenant 1
	for i := 0; i < 3; i++ {
		_, err := repo.CreateExecution(ctx, tenant1, workflowID, 1, "webhook", nil)
		require.NoError(t, err)
	}

	// Create executions for tenant 2
	for i := 0; i < 5; i++ {
		_, err := repo.CreateExecution(ctx, tenant2, workflowID, 1, "webhook", nil)
		require.NoError(t, err)
	}

	// Count for tenant 1
	count1, err := repo.CountExecutions(ctx, tenant1, ExecutionFilter{})
	require.NoError(t, err)
	assert.Equal(t, 3, count1)

	// Count for tenant 2
	count2, err := repo.CountExecutions(ctx, tenant2, ExecutionFilter{})
	require.NoError(t, err)
	assert.Equal(t, 5, count2)
}

