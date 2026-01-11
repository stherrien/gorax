package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAudit_EventLogging tests audit event logging functionality
func TestAudit_EventLogging(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tenantID := ts.CreateTestTenant(t, "Audit Test Tenant")
	userID := ts.CreateTestUser(t, tenantID, "audit@example.com", "user")
	headers := DefaultTestHeaders(tenantID)
	headers["X-User-ID"] = userID

	// Perform an action that should generate audit logs (create workflow)
	t.Run("CreateWorkflow_GeneratesAuditLog", func(t *testing.T) {
		workflowPayload := map[string]interface{}{
			"name":        "Audit Test Workflow",
			"description": "Workflow for testing audit logging",
			"definition": map[string]interface{}{
				"nodes": []map[string]interface{}{
					{
						"id":   "1",
						"type": "trigger",
						"data": map[string]interface{}{
							"nodeType": "webhook",
						},
					},
				},
				"edges": []map[string]interface{}{},
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", workflowPayload, headers)
		AssertStatusCode(t, resp, http.StatusCreated)

		var workflow map[string]interface{}
		ParseJSONResponse(t, resp, &workflow)
		workflowID := workflow["id"].(string)

		t.Logf("✓ Workflow created: %s", workflowID)

		// Wait for audit log to be written (async)
		time.Sleep(2 * time.Second)

		// Query audit logs directly from database
		query := `
			SELECT id, event_type, user_id, tenant_id, resource_type, resource_id, action
			FROM audit_logs
			WHERE tenant_id = $1 AND resource_id = $2
			ORDER BY created_at DESC
			LIMIT 1
		`

		var auditLog struct {
			ID           string `db:"id"`
			EventType    string `db:"event_type"`
			UserID       string `db:"user_id"`
			TenantID     string `db:"tenant_id"`
			ResourceType string `db:"resource_type"`
			ResourceID   string `db:"resource_id"`
			Action       string `db:"action"`
		}

		err := ts.DB.Get(&auditLog, query, tenantID, workflowID)
		require.NoError(t, err, "audit log should exist")

		assert.Equal(t, userID, auditLog.UserID)
		assert.Equal(t, tenantID, auditLog.TenantID)
		assert.Equal(t, "workflow", auditLog.ResourceType)
		assert.Equal(t, workflowID, auditLog.ResourceID)
		assert.Equal(t, "create", auditLog.Action)
		t.Logf("✓ Audit log verified: %s", auditLog.ID)
	})
}

// TestAudit_QueryAndFilter tests audit log querying and filtering
func TestAudit_QueryAndFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tenantID := ts.CreateTestTenant(t, "Audit Query Test Tenant")
	user1ID := ts.CreateTestUser(t, tenantID, "user1@example.com", "user")
	user2ID := ts.CreateTestUser(t, tenantID, "user2@example.com", "user")

	// Insert test audit logs
	testLogs := []struct {
		userID       string
		resourceType string
		action       string
		eventType    string
	}{
		{user1ID, "workflow", "create", "workflow.created"},
		{user1ID, "workflow", "update", "workflow.updated"},
		{user2ID, "workflow", "delete", "workflow.deleted"},
		{user2ID, "credential", "create", "credential.created"},
	}

	for i, log := range testLogs {
		query := `
			INSERT INTO audit_logs (id, tenant_id, user_id, event_type, resource_type, resource_id, action, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
		`
		resourceID := "resource-" + string(rune('0'+i))
		_, err := ts.DB.Exec(query, tenantID, log.userID, log.eventType, log.resourceType, resourceID, log.action, time.Now())
		require.NoError(t, err)
	}

	t.Logf("✓ Inserted %d test audit logs", len(testLogs))

	// Test filtering by user
	t.Run("FilterByUser", func(t *testing.T) {
		query := `
			SELECT COUNT(*) FROM audit_logs
			WHERE tenant_id = $1 AND user_id = $2
		`

		var count int
		err := ts.DB.Get(&count, query, tenantID, user1ID)
		require.NoError(t, err)

		assert.Equal(t, 2, count, "should find 2 logs for user1")
		t.Logf("✓ Found %d logs for user1", count)
	})

	// Test filtering by resource type
	t.Run("FilterByResourceType", func(t *testing.T) {
		query := `
			SELECT COUNT(*) FROM audit_logs
			WHERE tenant_id = $1 AND resource_type = $2
		`

		var count int
		err := ts.DB.Get(&count, query, tenantID, "workflow")
		require.NoError(t, err)

		assert.Equal(t, 3, count, "should find 3 workflow logs")
		t.Logf("✓ Found %d workflow logs", count)
	})

	// Test filtering by action
	t.Run("FilterByAction", func(t *testing.T) {
		query := `
			SELECT COUNT(*) FROM audit_logs
			WHERE tenant_id = $1 AND action = $2
		`

		var count int
		err := ts.DB.Get(&count, query, tenantID, "create")
		require.NoError(t, err)

		assert.Equal(t, 2, count, "should find 2 create actions")
		t.Logf("✓ Found %d create actions", count)
	})

	// Test time range filtering
	t.Run("FilterByTimeRange", func(t *testing.T) {
		// Query for logs created in the last hour
		query := `
			SELECT COUNT(*) FROM audit_logs
			WHERE tenant_id = $1 AND created_at > $2
		`

		oneHourAgo := time.Now().Add(-1 * time.Hour)
		var count int
		err := ts.DB.Get(&count, query, tenantID, oneHourAgo)
		require.NoError(t, err)

		assert.Equal(t, len(testLogs), count, "all logs should be within last hour")
		t.Logf("✓ Found %d logs in last hour", count)
	})
}

// TestAudit_Statistics tests audit log statistics aggregation
func TestAudit_Statistics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tenantID := ts.CreateTestTenant(t, "Audit Stats Test Tenant")
	userID := ts.CreateTestUser(t, tenantID, "stats@example.com", "user")

	// Insert various audit logs
	actions := []string{"create", "create", "update", "delete", "create"}
	for i, action := range actions {
		query := `
			INSERT INTO audit_logs (id, tenant_id, user_id, event_type, resource_type, resource_id, action, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
		`
		resourceID := "resource-" + string(rune('0'+i))
		eventType := "workflow." + action + "d"
		_, err := ts.DB.Exec(query, tenantID, userID, eventType, "workflow", resourceID, action, time.Now())
		require.NoError(t, err)
	}

	t.Logf("✓ Inserted test audit logs")

	// Test action statistics
	t.Run("ActionStatistics", func(t *testing.T) {
		query := `
			SELECT action, COUNT(*) as count
			FROM audit_logs
			WHERE tenant_id = $1
			GROUP BY action
			ORDER BY count DESC
		`

		rows, err := ts.DB.Query(query, tenantID)
		require.NoError(t, err)
		defer rows.Close()

		stats := make(map[string]int)
		for rows.Next() {
			var action string
			var count int
			err := rows.Scan(&action, &count)
			require.NoError(t, err)
			stats[action] = count
		}

		assert.Equal(t, 3, stats["create"], "should have 3 create actions")
		assert.Equal(t, 1, stats["update"], "should have 1 update action")
		assert.Equal(t, 1, stats["delete"], "should have 1 delete action")
		t.Logf("✓ Action statistics: %v", stats)
	})

	// Test daily activity
	t.Run("DailyActivity", func(t *testing.T) {
		query := `
			SELECT DATE(created_at) as date, COUNT(*) as count
			FROM audit_logs
			WHERE tenant_id = $1
			GROUP BY DATE(created_at)
			ORDER BY date DESC
		`

		rows, err := ts.DB.Query(query, tenantID)
		require.NoError(t, err)
		defer rows.Close()

		var totalCount int
		for rows.Next() {
			var date time.Time
			var count int
			err := rows.Scan(&date, &count)
			require.NoError(t, err)
			totalCount += count
			t.Logf("  Date: %s, Count: %d", date.Format("2006-01-02"), count)
		}

		assert.Equal(t, len(actions), totalCount, "total count should match")
		t.Logf("✓ Daily activity verified")
	})
}

// TestAudit_Retention tests audit log retention and archival
func TestAudit_Retention(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tenantID := ts.CreateTestTenant(t, "Audit Retention Test Tenant")
	userID := ts.CreateTestUser(t, tenantID, "retention@example.com", "user")

	// Insert audit logs with different ages
	testLogs := []struct {
		daysAgo int
		action  string
	}{
		{1, "create"},   // Recent (hot)
		{45, "update"},  // Medium age (hot)
		{100, "delete"}, // Old (should be in warm)
		{400, "create"}, // Very old (should be in cold)
	}

	for i, log := range testLogs {
		query := `
			INSERT INTO audit_logs (id, tenant_id, user_id, event_type, resource_type, resource_id, action, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
		`
		resourceID := "resource-" + string(rune('0'+i))
		eventType := "workflow." + log.action + "d"
		createdAt := time.Now().AddDate(0, 0, -log.daysAgo)
		_, err := ts.DB.Exec(query, tenantID, userID, eventType, "workflow", resourceID, log.action, createdAt)
		require.NoError(t, err)
	}

	t.Logf("✓ Inserted audit logs with different ages")

	// Query logs by age buckets
	t.Run("VerifyAgeBuckets", func(t *testing.T) {
		// Hot retention: last 90 days
		hotQuery := `
			SELECT COUNT(*) FROM audit_logs
			WHERE tenant_id = $1 AND created_at > NOW() - INTERVAL '90 days'
		`

		var hotCount int
		err := ts.DB.Get(&hotCount, hotQuery, tenantID)
		require.NoError(t, err)
		assert.Equal(t, 2, hotCount, "should have 2 logs in hot storage")
		t.Logf("✓ Hot storage: %d logs", hotCount)

		// Warm eligible: 90-365 days old
		warmQuery := `
			SELECT COUNT(*) FROM audit_logs
			WHERE tenant_id = $1
			  AND created_at <= NOW() - INTERVAL '90 days'
			  AND created_at > NOW() - INTERVAL '365 days'
		`

		var warmCount int
		err = ts.DB.Get(&warmCount, warmQuery, tenantID)
		require.NoError(t, err)
		assert.Equal(t, 1, warmCount, "should have 1 log eligible for warm storage")
		t.Logf("✓ Warm eligible: %d logs", warmCount)

		// Cold eligible: older than 365 days
		coldQuery := `
			SELECT COUNT(*) FROM audit_logs
			WHERE tenant_id = $1 AND created_at <= NOW() - INTERVAL '365 days'
		`

		var coldCount int
		err = ts.DB.Get(&coldCount, coldQuery, tenantID)
		require.NoError(t, err)
		assert.Equal(t, 1, coldCount, "should have 1 log eligible for cold storage")
		t.Logf("✓ Cold eligible: %d logs", coldCount)
	})
}

// TestAudit_Performance tests audit logging performance
func TestAudit_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tenantID := ts.CreateTestTenant(t, "Audit Performance Test Tenant")
	userID := ts.CreateTestUser(t, tenantID, "perf@example.com", "user")

	// Test bulk insertion performance
	t.Run("BulkInsertionPerformance", func(t *testing.T) {
		numLogs := 100
		start := time.Now()

		// Insert logs in batches
		batchSize := 10
		for i := 0; i < numLogs; i += batchSize {
			tx, err := ts.DB.Begin()
			require.NoError(t, err)

			for j := 0; j < batchSize && i+j < numLogs; j++ {
				query := `
					INSERT INTO audit_logs (id, tenant_id, user_id, event_type, resource_type, resource_id, action, created_at)
					VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
				`
				resourceID := "perf-test-" + string(rune('0'+i+j))
				_, err := tx.Exec(query, tenantID, userID, "workflow.created", "workflow", resourceID, "create", time.Now())
				require.NoError(t, err)
			}

			err = tx.Commit()
			require.NoError(t, err)
		}

		duration := time.Since(start)
		logsPerSecond := float64(numLogs) / duration.Seconds()

		t.Logf("✓ Inserted %d logs in %v (%.2f logs/sec)", numLogs, duration, logsPerSecond)
		assert.Less(t, duration, 5*time.Second, "bulk insertion should complete within 5 seconds")
		assert.Greater(t, logsPerSecond, 20.0, "should insert at least 20 logs/sec")
	})

	// Test query performance
	t.Run("QueryPerformance", func(t *testing.T) {
		// Query with various filters
		start := time.Now()

		query := `
			SELECT id, event_type, user_id, resource_type, action, created_at
			FROM audit_logs
			WHERE tenant_id = $1
			  AND created_at > $2
			ORDER BY created_at DESC
			LIMIT 50
		`

		rows, err := ts.DB.Query(query, tenantID, time.Now().Add(-24*time.Hour))
		require.NoError(t, err)
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}

		duration := time.Since(start)
		t.Logf("✓ Queried %d logs in %v", count, duration)
		assert.Less(t, duration, 100*time.Millisecond, "query should complete within 100ms")
	})
}
