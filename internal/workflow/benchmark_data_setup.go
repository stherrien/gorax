package workflow

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// BenchmarkDataGenerator creates production-like data for benchmarking
type BenchmarkDataGenerator struct {
	db       *sqlx.DB
	tenantID string
}

// NewBenchmarkDataGenerator creates a new data generator
func NewBenchmarkDataGenerator(db *sqlx.DB) *BenchmarkDataGenerator {
	return &BenchmarkDataGenerator{
		db: db,
	}
}

// CreateTestTenant creates a test tenant for benchmarking
func (g *BenchmarkDataGenerator) CreateTestTenant() (string, error) {
	tenantID := uuid.NewString()
	subdomain := "bench-" + uuid.NewString()[:8]
	_, err := g.db.Exec(`INSERT INTO tenants (id, name, subdomain, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		tenantID, "Benchmark Tenant", subdomain, time.Now(), time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to create tenant: %w", err)
	}
	g.tenantID = tenantID
	return tenantID, nil
}

// CreateWorkflows creates N workflows for benchmarking
func (g *BenchmarkDataGenerator) CreateWorkflows(ctx context.Context, count int) ([]string, error) {
	workflowIDs := make([]string, count)

	// Use a transaction for faster inserts
	tx, err := g.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO workflows (id, tenant_id, name, description, definition, status, version, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i := 0; i < count; i++ {
		id := uuid.NewString()
		name := fmt.Sprintf("Benchmark Workflow %d", i)
		definition := `{"nodes":[],"edges":[]}`
		now := time.Now()
		createdBy := uuid.NewString() // Use UUID for created_by

		_, err := stmt.ExecContext(ctx,
			id, g.tenantID, name, "Benchmark workflow", definition, "active", 1, createdBy, now, now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert workflow %d: %w", i, err)
		}
		workflowIDs[i] = id
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return workflowIDs, nil
}

// CreateExecutions creates executions for the given workflows
func (g *BenchmarkDataGenerator) CreateExecutions(ctx context.Context, workflowIDs []string, executionsPerWorkflow int, failureRate float64) error {
	tx, err := g.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO executions (id, tenant_id, workflow_id, workflow_version, trigger_type, status, created_at, started_at, completed_at, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	errorMessages := []string{
		"Database connection timeout",
		"API rate limit exceeded",
		"Network connection refused",
		"Invalid authentication credentials",
		"Resource not found",
		"Internal server error",
		"Request timeout",
		"Service unavailable",
	}

	for _, workflowID := range workflowIDs {
		for j := 0; j < executionsPerWorkflow; j++ {
			id := uuid.NewString()
			createdAt := now.Add(-time.Duration(j) * time.Minute)
			startedAt := createdAt.Add(1 * time.Second)
			completedAt := startedAt.Add(5 * time.Second)

			// Determine status based on failure rate
			var status string
			var errorMsg sql.NullString

			if float64(j%100)/100.0 < failureRate {
				status = "failed"
				errorMsg.String = errorMessages[j%len(errorMessages)]
				errorMsg.Valid = true
			} else {
				status = "completed"
				errorMsg.Valid = false
			}

			_, err := stmt.ExecContext(ctx,
				id, g.tenantID, workflowID, 1, "manual", status, createdAt, startedAt, completedAt, errorMsg,
			)
			if err != nil {
				return fmt.Errorf("failed to insert execution: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Cleanup removes all benchmark data
func (g *BenchmarkDataGenerator) Cleanup(ctx context.Context) error {
	if g.tenantID == "" {
		return nil
	}

	// Delete in reverse order of foreign keys
	queries := []string{
		"DELETE FROM executions WHERE tenant_id = $1",
		"DELETE FROM workflows WHERE tenant_id = $1",
		"DELETE FROM tenants WHERE id = $1",
	}

	for _, query := range queries {
		if _, err := g.db.ExecContext(ctx, query, g.tenantID); err != nil {
			return fmt.Errorf("cleanup failed: %w", err)
		}
	}

	return nil
}

// GenerateFullDataset creates a complete dataset for benchmarking
func (g *BenchmarkDataGenerator) GenerateFullDataset(ctx context.Context, workflowCount, executionsPerWorkflow int, failureRate float64) error {
	// Create tenant
	if _, err := g.CreateTestTenant(); err != nil {
		return err
	}

	// Create workflows
	workflowIDs, err := g.CreateWorkflows(ctx, workflowCount)
	if err != nil {
		return err
	}

	// Create executions
	if err := g.CreateExecutions(ctx, workflowIDs, executionsPerWorkflow, failureRate); err != nil {
		return err
	}

	return nil
}
