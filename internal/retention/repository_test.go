package retention

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests require a test database. They will be skipped if TEST_DATABASE_URL is not set.
// To run: TEST_DATABASE_URL="postgres://user:pass@localhost/test_db?sslmode=disable" go test

func setupTestDB(t *testing.T) *sqlx.DB {
	dbURL := getTestDatabaseURL(t)
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sqlx.Connect("postgres", dbURL)
	require.NoError(t, err)

	// Clean up tables
	_, err = db.Exec("DELETE FROM retention_cleanup_logs")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM step_executions")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM executions")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM workflows")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM tenants")
	require.NoError(t, err)

	return db
}

func getTestDatabaseURL(t *testing.T) string {
	// Return empty string if not set - tests will be skipped
	return ""
}

func createTestTenant(t *testing.T, db *sqlx.DB, retentionDays int) string {
	tenantID := uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO tenants (id, name, subdomain, status, tier, settings, quotas, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', 'free',
			jsonb_build_object('retention_days', $4, 'retention_enabled', true),
			'{}', NOW(), NOW())
	`, tenantID, "Test Tenant", "test-"+tenantID[:8], retentionDays)
	require.NoError(t, err)
	return tenantID
}

func createTestWorkflow(t *testing.T, db *sqlx.DB, tenantID string) string {
	workflowID := uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO workflows (id, tenant_id, name, description, definition, status, version, created_by, created_at, updated_at)
		VALUES ($1, $2, 'Test Workflow', 'Test', '{"nodes":[],"edges":[]}', 'active', 1, 'test-user', NOW(), NOW())
	`, workflowID, tenantID)
	require.NoError(t, err)
	return workflowID
}

func createTestExecution(t *testing.T, db *sqlx.DB, tenantID, workflowID string, createdAt time.Time) string {
	executionID := uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO executions (id, tenant_id, workflow_id, workflow_version, status, trigger_type, created_at)
		VALUES ($1, $2, $3, 1, 'completed', 'manual', $4)
	`, executionID, tenantID, workflowID, createdAt)
	require.NoError(t, err)
	return executionID
}

func createTestStepExecution(t *testing.T, db *sqlx.DB, executionID string) string {
	stepID := uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO step_executions (id, execution_id, node_id, node_type, status)
		VALUES ($1, $2, 'node-1', 'http_action', 'completed')
	`, stepID, executionID)
	require.NoError(t, err)
	return stepID
}

func TestRepository_GetRetentionPolicy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("get policy from tenant settings", func(t *testing.T) {
		tenantID := createTestTenant(t, db, 45)

		policy, err := repo.GetRetentionPolicy(ctx, tenantID)
		require.NoError(t, err)
		assert.Equal(t, tenantID, policy.TenantID)
		assert.Equal(t, 45, policy.RetentionDays)
		assert.True(t, policy.Enabled)
	})

	t.Run("tenant not found", func(t *testing.T) {
		policy, err := repo.GetRetentionPolicy(ctx, uuid.New().String())
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, policy)
	})

	t.Run("retention disabled", func(t *testing.T) {
		tenantID := uuid.New().String()
		_, err := db.Exec(`
			INSERT INTO tenants (id, name, subdomain, status, tier, settings, quotas, created_at, updated_at)
			VALUES ($1, $2, $3, 'active', 'free',
				jsonb_build_object('retention_days', 30, 'retention_enabled', false),
				'{}', NOW(), NOW())
		`, tenantID, "Test Tenant", "test-"+tenantID[:8])
		require.NoError(t, err)

		policy, err := repo.GetRetentionPolicy(ctx, tenantID)
		require.NoError(t, err)
		assert.Equal(t, tenantID, policy.TenantID)
		assert.False(t, policy.Enabled)
	})
}

func TestRepository_DeleteOldExecutions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("delete old executions with batch processing", func(t *testing.T) {
		tenantID := createTestTenant(t, db, 30)
		workflowID := createTestWorkflow(t, db, tenantID)

		now := time.Now()
		cutoff := now.Add(-30 * 24 * time.Hour)

		// Create old executions (should be deleted)
		oldExecution1 := createTestExecution(t, db, tenantID, workflowID, cutoff.Add(-1*time.Hour))
		createTestStepExecution(t, db, oldExecution1)
		createTestStepExecution(t, db, oldExecution1)

		oldExecution2 := createTestExecution(t, db, tenantID, workflowID, cutoff.Add(-2*time.Hour))
		createTestStepExecution(t, db, oldExecution2)

		// Create recent executions (should NOT be deleted)
		recentExecution := createTestExecution(t, db, tenantID, workflowID, now.Add(-1*time.Hour))
		createTestStepExecution(t, db, recentExecution)

		result, err := repo.DeleteOldExecutions(ctx, tenantID, cutoff, 1000)
		require.NoError(t, err)
		assert.Equal(t, 2, result.ExecutionsDeleted)
		assert.Equal(t, 3, result.StepExecutionsDeleted)
		assert.Equal(t, 1, result.BatchesProcessed)

		// Verify old executions are deleted
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM executions WHERE id = $1", oldExecution1)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Verify recent execution still exists
		err = db.Get(&count, "SELECT COUNT(*) FROM executions WHERE id = $1", recentExecution)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		tenant1 := createTestTenant(t, db, 30)
		workflow1 := createTestWorkflow(t, db, tenant1)

		tenant2 := createTestTenant(t, db, 30)
		workflow2 := createTestWorkflow(t, db, tenant2)

		cutoff := time.Now().Add(-30 * 24 * time.Hour)
		oldTime := cutoff.Add(-1 * time.Hour)

		// Create old executions for both tenants
		exec1 := createTestExecution(t, db, tenant1, workflow1, oldTime)
		createTestStepExecution(t, db, exec1)

		exec2 := createTestExecution(t, db, tenant2, workflow2, oldTime)
		createTestStepExecution(t, db, exec2)

		// Delete for tenant1 only
		result, err := repo.DeleteOldExecutions(ctx, tenant1, cutoff, 1000)
		require.NoError(t, err)
		assert.Equal(t, 1, result.ExecutionsDeleted)
		assert.Equal(t, 1, result.StepExecutionsDeleted)

		// Verify tenant1 execution deleted
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM executions WHERE id = $1", exec1)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Verify tenant2 execution still exists
		err = db.Get(&count, "SELECT COUNT(*) FROM executions WHERE id = $1", exec2)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("batch size limit", func(t *testing.T) {
		tenantID := createTestTenant(t, db, 30)
		workflowID := createTestWorkflow(t, db, tenantID)

		cutoff := time.Now().Add(-30 * 24 * time.Hour)
		oldTime := cutoff.Add(-1 * time.Hour)

		// Create 5 old executions
		for i := 0; i < 5; i++ {
			execID := createTestExecution(t, db, tenantID, workflowID, oldTime)
			createTestStepExecution(t, db, execID)
		}

		// Delete with batch size of 2
		result, err := repo.DeleteOldExecutions(ctx, tenantID, cutoff, 2)
		require.NoError(t, err)
		assert.Equal(t, 5, result.ExecutionsDeleted)
		assert.Equal(t, 5, result.StepExecutionsDeleted)
		assert.GreaterOrEqual(t, result.BatchesProcessed, 3) // Should process in multiple batches
	})

	t.Run("no executions to delete", func(t *testing.T) {
		tenantID := createTestTenant(t, db, 30)
		workflowID := createTestWorkflow(t, db, tenantID)

		now := time.Now()
		cutoff := now.Add(-30 * 24 * time.Hour)

		// Create only recent executions
		recentExecution := createTestExecution(t, db, tenantID, workflowID, now.Add(-1*time.Hour))
		createTestStepExecution(t, db, recentExecution)

		result, err := repo.DeleteOldExecutions(ctx, tenantID, cutoff, 1000)
		require.NoError(t, err)
		assert.Equal(t, 0, result.ExecutionsDeleted)
		assert.Equal(t, 0, result.StepExecutionsDeleted)
		assert.Equal(t, 0, result.BatchesProcessed)
	})
}

func TestRepository_GetTenantsWithRetention(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("get active tenants only", func(t *testing.T) {
		tenant1 := createTestTenant(t, db, 30)
		tenant2 := createTestTenant(t, db, 60)

		// Create deleted tenant (should not be returned)
		deletedID := uuid.New().String()
		_, err := db.Exec(`
			INSERT INTO tenants (id, name, subdomain, status, tier, settings, quotas, created_at, updated_at)
			VALUES ($1, $2, $3, 'deleted', 'free',
				jsonb_build_object('retention_days', 30, 'retention_enabled', true),
				'{}', NOW(), NOW())
		`, deletedID, "Deleted Tenant", "deleted-"+deletedID[:8])
		require.NoError(t, err)

		tenants, err := repo.GetTenantsWithRetention(ctx)
		require.NoError(t, err)
		assert.Len(t, tenants, 2)
		assert.Contains(t, tenants, tenant1)
		assert.Contains(t, tenants, tenant2)
		assert.NotContains(t, tenants, deletedID)
	})
}

func TestRepository_LogCleanup(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("log cleanup successfully", func(t *testing.T) {
		tenantID := createTestTenant(t, db, 30)

		log := &CleanupLog{
			ID:                    uuid.New().String(),
			TenantID:              tenantID,
			ExecutionsDeleted:     100,
			StepExecutionsDeleted: 300,
			RetentionDays:         30,
			CutoffDate:            time.Now().Add(-30 * 24 * time.Hour),
			DurationMs:            1500,
			Status:                "completed",
			CreatedAt:             time.Now(),
		}

		err := repo.LogCleanup(ctx, log)
		require.NoError(t, err)

		// Verify log was created
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM retention_cleanup_logs WHERE id = $1", log.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify log data
		var storedLog CleanupLog
		err = db.Get(&storedLog, "SELECT * FROM retention_cleanup_logs WHERE id = $1", log.ID)
		require.NoError(t, err)
		assert.Equal(t, log.TenantID, storedLog.TenantID)
		assert.Equal(t, log.ExecutionsDeleted, storedLog.ExecutionsDeleted)
		assert.Equal(t, log.StepExecutionsDeleted, storedLog.StepExecutionsDeleted)
		assert.Equal(t, log.RetentionDays, storedLog.RetentionDays)
		assert.Equal(t, log.Status, storedLog.Status)
	})

	t.Run("log cleanup with error", func(t *testing.T) {
		tenantID := createTestTenant(t, db, 30)

		errorMsg := "database connection timeout"
		log := &CleanupLog{
			ID:                    uuid.New().String(),
			TenantID:              tenantID,
			ExecutionsDeleted:     0,
			StepExecutionsDeleted: 0,
			RetentionDays:         30,
			CutoffDate:            time.Now().Add(-30 * 24 * time.Hour),
			DurationMs:            500,
			Status:                "failed",
			ErrorMessage:          &errorMsg,
			CreatedAt:             time.Now(),
		}

		err := repo.LogCleanup(ctx, log)
		require.NoError(t, err)

		// Verify error message was stored
		var storedLog CleanupLog
		err = db.Get(&storedLog, "SELECT * FROM retention_cleanup_logs WHERE id = $1", log.ID)
		require.NoError(t, err)
		assert.Equal(t, "failed", storedLog.Status)
		assert.NotNil(t, storedLog.ErrorMessage)
		assert.Equal(t, errorMsg, *storedLog.ErrorMessage)
	})
}

func TestRepository_DeleteOldExecutions_ForeignKeyOrder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("deletes step_executions before executions", func(t *testing.T) {
		tenantID := createTestTenant(t, db, 30)
		workflowID := createTestWorkflow(t, db, tenantID)

		cutoff := time.Now().Add(-30 * 24 * time.Hour)
		oldTime := cutoff.Add(-1 * time.Hour)

		// Create execution with multiple steps
		execID := createTestExecution(t, db, tenantID, workflowID, oldTime)
		step1 := createTestStepExecution(t, db, execID)
		step2 := createTestStepExecution(t, db, execID)

		result, err := repo.DeleteOldExecutions(ctx, tenantID, cutoff, 1000)
		require.NoError(t, err)

		// Should delete both steps and execution without foreign key violation
		assert.Equal(t, 1, result.ExecutionsDeleted)
		assert.Equal(t, 2, result.StepExecutionsDeleted)

		// Verify all are deleted
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM step_executions WHERE id IN ($1, $2)", step1, step2)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		err = db.Get(&count, "SELECT COUNT(*) FROM executions WHERE id = $1", execID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
