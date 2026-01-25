package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// BenchmarkGetTopFailures_Current benchmarks the current implementation with correlated subquery
func BenchmarkGetTopFailures_Current(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Skipping benchmark - set TEST_DATABASE_URL environment variable to run")
		return
	}
	defer db.Close()

	repo := NewRepository(db)
	tenantID := createTestTenant(&testing.T{}, db)

	// Create test data: 100 workflows with varying failure counts
	workflowIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		workflowID := createTestWorkflowWithName(&testing.T{}, repo, tenantID, "Workflow "+uuid.NewString()[:8])
		workflowIDs[i] = workflowID

		// Create 50-100 failed executions per workflow
		failureCount := 50 + (i % 50)
		for j := 0; j < failureCount; j++ {
			createExecutionWithStatusAndError(&testing.T{}, repo, tenantID, workflowID, "failed",
				time.Now().Add(-time.Duration(j)*time.Minute), "Database connection timeout")
		}

		// Add some successful executions too
		for j := 0; j < 20; j++ {
			createExecutionWithStatus(&testing.T{}, repo, tenantID, workflowID, "completed",
				time.Now().Add(-time.Duration(j)*time.Minute))
		}
	}

	ctx := context.Background()
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now().Add(24 * time.Hour)

	// Reset timer to exclude setup time
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.GetTopFailures(ctx, tenantID, startDate, endDate, 10)
		if err != nil {
			b.Fatalf("GetTopFailures failed: %v", err)
		}
	}
}

// BenchmarkGetTopFailures_Optimized benchmarks the optimized implementation with LATERAL join
func BenchmarkGetTopFailures_Optimized(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Skipping benchmark - set TEST_DATABASE_URL environment variable to run")
		return
	}
	defer db.Close()

	repo := NewRepository(db)
	tenantID := createTestTenant(&testing.T{}, db)

	// Create test data: 100 workflows with varying failure counts
	workflowIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		workflowID := createTestWorkflowWithName(&testing.T{}, repo, tenantID, "Workflow "+uuid.NewString()[:8])
		workflowIDs[i] = workflowID

		// Create 50-100 failed executions per workflow
		failureCount := 50 + (i % 50)
		for j := 0; j < failureCount; j++ {
			createExecutionWithStatusAndError(&testing.T{}, repo, tenantID, workflowID, "failed",
				time.Now().Add(-time.Duration(j)*time.Minute), "Database connection timeout")
		}

		// Add some successful executions too
		for j := 0; j < 20; j++ {
			createExecutionWithStatus(&testing.T{}, repo, tenantID, workflowID, "completed",
				time.Now().Add(-time.Duration(j)*time.Minute))
		}
	}

	ctx := context.Background()
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now().Add(24 * time.Hour)

	// Reset timer to exclude setup time
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.GetTopFailuresOptimized(ctx, tenantID, startDate, endDate, 10)
		if err != nil {
			b.Fatalf("GetTopFailuresOptimized failed: %v", err)
		}
	}
}

// Helper to create execution with error message
func createExecutionWithStatusAndError(t *testing.T, repo *Repository, tenantID, workflowID, status string, createdAt time.Time, errorMsg string) string {
	exec, err := repo.CreateExecution(context.Background(), tenantID, workflowID, 1, "manual", nil)
	if err != nil {
		t.Fatalf("CreateExecution failed: %v", err)
		return ""
	}

	// Update with error message
	_, err = repo.db.ExecContext(context.Background(),
		"UPDATE executions SET status = $1, created_at = $2, started_at = $2, completed_at = $3, error_message = $4 WHERE id = $5",
		status, createdAt, createdAt.Add(1*time.Second), errorMsg, exec.ID)
	if err != nil {
		t.Fatalf("Update execution failed: %v", err)
		return ""
	}

	return exec.ID
}
