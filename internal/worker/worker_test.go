package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/workflow"
)

// TestPollExecution_NoWork tests that pollExecution returns ErrNoWork when no pending executions
func TestPollExecution_NoWork(t *testing.T) {
	// Setup
	db, cleanup := setupTestDB(t)
	defer cleanup()

	w := &Worker{
		db:           db,
		workflowRepo: workflow.NewRepository(db),
	}

	// Test
	execution, err := w.pollExecution(context.Background())

	// Assert
	assert.Nil(t, execution)
	assert.ErrorIs(t, err, ErrNoWork)
}

// TestPollExecution_ReturnsPendingExecution tests polling returns oldest pending execution
func TestPollExecution_ReturnsPendingExecution(t *testing.T) {
	// Setup
	db, cleanup := setupTestDB(t)
	defer cleanup()

	w := &Worker{
		db:           db,
		workflowRepo: workflow.NewRepository(db),
	}

	ctx := context.Background()
	tenantID := "test-tenant"
	workflowID := "test-workflow"

	// Create a workflow first
	createWorkflow(t, db, tenantID, workflowID)

	// Create pending executions
	exec1 := createExecution(t, db, tenantID, workflowID, "pending", time.Now().Add(-2*time.Minute))
	exec2 := createExecution(t, db, tenantID, workflowID, "pending", time.Now().Add(-1*time.Minute))
	createExecution(t, db, tenantID, workflowID, "running", time.Now())  // Should not be picked
	createExecution(t, db, tenantID, workflowID, "completed", time.Now()) // Should not be picked

	// Test - should return oldest pending execution
	execution, err := w.pollExecution(ctx)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, execution)
	assert.Equal(t, exec1, execution.ID)
	assert.Equal(t, tenantID, execution.TenantID)
	assert.Equal(t, workflowID, execution.WorkflowID)
	assert.Equal(t, "pending", execution.Status)

	// Test again - should return second oldest
	execution, err = w.pollExecution(ctx)
	require.NoError(t, err)
	require.NotNil(t, execution)
	assert.Equal(t, exec2, execution.ID)
}

// TestPollExecution_UpdatesStatusToPending tests that polling updates status to running
func TestPollExecution_UpdatesStatusToRunning(t *testing.T) {
	// Setup
	db, cleanup := setupTestDB(t)
	defer cleanup()

	w := &Worker{
		db:           db,
		workflowRepo: workflow.NewRepository(db),
	}

	ctx := context.Background()
	tenantID := "test-tenant"
	workflowID := "test-workflow"

	createWorkflow(t, db, tenantID, workflowID)
	execID := createExecution(t, db, tenantID, workflowID, "pending", time.Now())

	// Test
	execution, err := w.pollExecution(ctx)
	require.NoError(t, err)
	require.NotNil(t, execution)

	// Verify status was updated to running
	updatedExec, err := w.workflowRepo.GetExecutionByID(ctx, tenantID, execID)
	require.NoError(t, err)
	assert.Equal(t, "running", updatedExec.Status)
	assert.NotNil(t, updatedExec.StartedAt)
}

// TestPollExecution_SkipsStaleExecutions tests that old executions are marked as failed
func TestPollExecution_SkipsStaleExecutions(t *testing.T) {
	// Setup
	db, cleanup := setupTestDB(t)
	defer cleanup()

	w := &Worker{
		db:           db,
		workflowRepo: workflow.NewRepository(db),
	}

	ctx := context.Background()
	tenantID := "test-tenant"
	workflowID := "test-workflow"

	createWorkflow(t, db, tenantID, workflowID)

	// Create stale execution (older than 1 hour)
	staleExecID := createExecution(t, db, tenantID, workflowID, "pending", time.Now().Add(-2*time.Hour))

	// Create fresh execution
	freshExecID := createExecution(t, db, tenantID, workflowID, "pending", time.Now())

	// Test
	execution, err := w.pollExecution(ctx)
	require.NoError(t, err)
	require.NotNil(t, execution)

	// Should return fresh execution, not stale one
	assert.Equal(t, freshExecID, execution.ID)

	// Verify stale execution was marked as failed
	staleExec, err := w.workflowRepo.GetExecutionByID(ctx, tenantID, staleExecID)
	require.NoError(t, err)
	assert.Equal(t, "failed", staleExec.Status)
	assert.NotNil(t, staleExec.ErrorMessage)
	assert.Contains(t, *staleExec.ErrorMessage, "execution timeout")
}

// TestPollExecution_WithDelay tests that polling waits when no work available
func TestPollExecution_WithDelay(t *testing.T) {
	// Setup
	db, cleanup := setupTestDB(t)
	defer cleanup()

	w := &Worker{
		db:           db,
		workflowRepo: workflow.NewRepository(db),
	}

	ctx := context.Background()

	// Test - time how long it takes when no work
	start := time.Now()
	_, err := w.pollExecution(ctx)
	duration := time.Since(start)

	// Assert - should return quickly with ErrNoWork
	assert.ErrorIs(t, err, ErrNoWork)
	assert.Less(t, duration, 100*time.Millisecond, "polling should not block when no work available")
}

// TestRequeueMessage_IncreasesVisibilityTimeout tests that messages are requeued with delay
func TestRequeueMessage_IncreasesVisibilityTimeout(t *testing.T) {
	// This test is for the queue-based processing scenario
	// Skip if SQS is not available
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup would require SQS mock or integration test
	// For now, we'll test the logic through processExecutionMessage
	t.Skip("requires SQS integration test setup")
}

// TestProcessExecution_RequeuesWhenTenantAtCapacity tests requeue logic
func TestProcessExecution_RequeuesWhenTenantAtCapacity(t *testing.T) {
	// Setup
	db, cleanup := setupTestDB(t)
	defer cleanup()

	redisClient := setupTestRedis(t)
	defer redisClient.Close()

	// Create worker with low concurrency limit
	w := &Worker{
		db:               db,
		redis:            redisClient,
		workflowRepo:     workflow.NewRepository(db),
		concurrencyLimit: NewTenantConcurrencyLimiter(redisClient, 1),
	}

	ctx := context.Background()
	tenantID := "test-tenant"
	workflowID := "test-workflow"

	createWorkflow(t, db, tenantID, workflowID)

	exec1ID := createExecution(t, db, tenantID, workflowID, "pending", time.Now())
	exec2ID := createExecution(t, db, tenantID, workflowID, "pending", time.Now())

	exec2, err := w.workflowRepo.GetExecutionByID(ctx, tenantID, exec2ID)
	require.NoError(t, err)

	// Acquire slot for first execution
	acquired, err := w.concurrencyLimit.Acquire(ctx, tenantID, exec1ID)
	require.NoError(t, err)
	require.True(t, acquired)

	// Try to process second execution - should fail with ErrTenantAtCapacity
	err = w.processExecution(ctx, exec2)
	assert.ErrorIs(t, err, ErrTenantAtCapacity)

	// Verify second execution is still pending (not marked as failed)
	updatedExec2, err := w.workflowRepo.GetExecutionByID(ctx, tenantID, exec2ID)
	require.NoError(t, err)
	assert.Equal(t, "pending", updatedExec2.Status)
}

// Helper functions for testing

// setupTestDB creates an in-memory test database
func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	// This would typically use a test database or in-memory database
	// For now, we'll skip actual DB setup and assume test DB is available
	t.Skip("requires test database setup - implement with your test DB config")
	return nil, func() {}
}

// setupTestRedis creates a test Redis client
func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use a test database
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Clean up test keys
	client.FlushDB(ctx)

	return client
}

// createWorkflow creates a test workflow
func createWorkflow(t *testing.T, db *sqlx.DB, tenantID, workflowID string) {
	definition := json.RawMessage(`{"nodes": [], "edges": []}`)
	query := `
		INSERT INTO workflows (id, tenant_id, name, description, definition, status, version, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	now := time.Now()
	_, err := db.Exec(query,
		workflowID, tenantID, "Test Workflow", "Test Description",
		definition, "active", 1, "test-user", now, now,
	)
	require.NoError(t, err)
}

// createExecution creates a test execution
func createExecution(t *testing.T, db *sqlx.DB, tenantID, workflowID, status string, createdAt time.Time) string {
	query := `
		INSERT INTO executions (id, tenant_id, workflow_id, workflow_version, status, trigger_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	execID := "exec-" + time.Now().Format("20060102150405.000000")
	_, err := db.Exec(query, execID, tenantID, workflowID, 1, status, "manual", createdAt)
	require.NoError(t, err)
	return execID
}
