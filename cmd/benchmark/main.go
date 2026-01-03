package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gorax/gorax/internal/workflow"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("TEST_DATABASE_URL environment variable must be set")
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Generate benchmark data
	log.Println("Generating benchmark data...")
	log.Println("Creating 100 workflows with 100 executions each (30% failure rate)...")
	generator := workflow.NewBenchmarkDataGenerator(db)
	defer generator.Cleanup(ctx)

	err = generator.GenerateFullDataset(ctx, 100, 100, 0.30) // 100 workflows, 100 executions each, 30% failure rate
	if err != nil {
		log.Fatalf("Failed to generate data: %v", err)
	}

	// Get total execution count
	var totalExec, failedExec int
	db.GetContext(ctx, &totalExec, "SELECT COUNT(*) FROM executions")
	db.GetContext(ctx, &failedExec, "SELECT COUNT(*) FROM executions WHERE status = 'failed'")
	log.Printf("Data generation complete! Created %d total executions (%d failed)\n", totalExec, failedExec)

	// Get tenant ID for queries
	var tenantID string
	err = db.GetContext(ctx, &tenantID, "SELECT id FROM tenants WHERE name = 'Benchmark Tenant' LIMIT 1")
	if err != nil {
		log.Fatalf("Failed to get tenant ID: %v", err)
	}

	// Run benchmarks
	repo := workflow.NewRepository(db)
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now().Add(24 * time.Hour)

	// Warm up
	log.Println("Warming up...")
	_, _ = repo.GetTopFailures(ctx, tenantID, startDate, endDate, 10)
	_, _ = repo.GetTopFailuresOptimized(ctx, tenantID, startDate, endDate, 10)

	// Benchmark GetTopFailures (old)
	log.Println("\nBenchmarking GetTopFailures (correlated subquery)...")
	iterations := 100
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := repo.GetTopFailures(ctx, tenantID, startDate, endDate, 10)
		if err != nil {
			log.Fatalf("GetTopFailures failed: %v", err)
		}
	}
	oldDuration := time.Since(start)
	oldAvg := oldDuration / time.Duration(iterations)
	log.Printf("Average: %v (total: %v for %d iterations)\n", oldAvg, oldDuration, iterations)

	// Benchmark GetTopFailuresOptimized (new)
	log.Println("\nBenchmarking GetTopFailuresOptimized (LATERAL join)...")
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := repo.GetTopFailuresOptimized(ctx, tenantID, startDate, endDate, 10)
		if err != nil {
			log.Fatalf("GetTopFailuresOptimized failed: %v", err)
		}
	}
	newDuration := time.Since(start)
	newAvg := newDuration / time.Duration(iterations)
	log.Printf("Average: %v (total: %v for %d iterations)\n", newAvg, newDuration, iterations)

	// Calculate improvement
	improvement := float64(oldDuration-newDuration) / float64(oldDuration) * 100
	speedup := float64(oldDuration) / float64(newDuration)

	log.Println("\n" + string([]byte{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61}))
	log.Println("RESULTS SUMMARY")
	log.Println(string([]byte{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61}))
	log.Printf("Old Implementation:  %v per query\n", oldAvg)
	log.Printf("New Implementation:  %v per query\n", newAvg)
	log.Printf("Improvement:         %.2f%%\n", improvement)
	log.Printf("Speedup:             %.2fx faster\n", speedup)
	log.Println(string([]byte{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61}))

	// Get EXPLAIN ANALYZE for both queries
	log.Println("\nGenerating EXPLAIN ANALYZE plans...")

	// Old query
	log.Println("\n--- Old Implementation (Correlated Subquery) ---")
	var oldPlan []string
	err = db.SelectContext(ctx, &oldPlan, `
		EXPLAIN ANALYZE
		SELECT
			e.workflow_id,
			w.name as workflow_name,
			COUNT(*) as failure_count,
			MAX(e.completed_at) as last_failed_at,
			(
				SELECT error_message
				FROM executions
				WHERE workflow_id = e.workflow_id
					AND status = 'failed'
					AND error_message IS NOT NULL
				ORDER BY completed_at DESC
				LIMIT 1
			) as error_preview
		FROM executions e
		INNER JOIN workflows w ON e.workflow_id = w.id
		WHERE e.tenant_id = $1
			AND e.created_at >= $2
			AND e.created_at < $3
			AND e.status = 'failed'
		GROUP BY e.workflow_id, w.name
		ORDER BY failure_count DESC
		LIMIT 10
	`, tenantID, startDate, endDate)
	if err != nil {
		log.Printf("Failed to get old query plan: %v", err)
	} else {
		for _, line := range oldPlan {
			log.Println(line)
		}
	}

	// New query
	log.Println("\n--- New Implementation (LATERAL Join) ---")
	var newPlan []string
	err = db.SelectContext(ctx, &newPlan, `
		EXPLAIN ANALYZE
		SELECT
			e.workflow_id,
			w.name as workflow_name,
			COUNT(*) as failure_count,
			MAX(e.completed_at) as last_failed_at,
			latest.error_message as error_preview
		FROM executions e
		INNER JOIN workflows w ON e.workflow_id = w.id
		CROSS JOIN LATERAL (
			SELECT error_message
			FROM executions
			WHERE workflow_id = e.workflow_id
				AND status = 'failed'
				AND error_message IS NOT NULL
			ORDER BY completed_at DESC
			LIMIT 1
		) latest
		WHERE e.tenant_id = $1
			AND e.created_at >= $2
			AND e.created_at < $3
			AND e.status = 'failed'
		GROUP BY e.workflow_id, w.name, latest.error_message
		ORDER BY failure_count DESC
		LIMIT 10
	`, tenantID, startDate, endDate)
	if err != nil {
		log.Printf("Failed to get new query plan: %v", err)
	} else {
		for _, line := range newPlan {
			log.Println(line)
		}
	}

	log.Println("\nBenchmark complete!")
}
