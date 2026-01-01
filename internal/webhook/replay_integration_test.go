package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ReplaySingleEvent tests replaying a single webhook event
func TestIntegration_ReplaySingleEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupReplayTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := createTestTenantForReplay(t, db)
	repo := NewRepository(db)
	executor := &testWorkflowExecutor{executionIDs: make(map[string]string)}
	service := NewReplayService(repo, executor, testReplayLogger())

	// Step 1: Create a webhook
	webhook := createTestWebhookForReplay(t, repo, ctx, tenantID, "workflow-123", true)
	t.Logf("✓ Created webhook: %s", webhook.ID)

	// Step 2: Create a webhook event
	event := createTestEvent(t, repo, ctx, tenantID, webhook.ID, EventStatusProcessed)
	t.Logf("✓ Created event: %s", event.ID)

	// Step 3: Replay the event
	result := service.ReplayEvent(ctx, tenantID, event.ID, nil)

	// Step 4: Verify replay succeeded
	require.NotNil(t, result)
	assert.True(t, result.Success, "replay should succeed")
	assert.NotEmpty(t, result.ExecutionID, "execution ID should be returned")
	assert.Empty(t, result.Error, "no error should be returned")
	t.Logf("✓ Event replayed successfully, execution ID: %s", result.ExecutionID)

	// Step 5: Verify executor was called with correct parameters
	execID, exists := executor.executionIDs[webhook.WorkflowID]
	assert.True(t, exists, "workflow should have been executed")
	assert.Equal(t, result.ExecutionID, execID)
	t.Logf("✓ Workflow execution verified")
}

// TestIntegration_ReplayWithModifiedPayload tests replaying with a modified payload
func TestIntegration_ReplayWithModifiedPayload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupReplayTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := createTestTenantForReplay(t, db)
	repo := NewRepository(db)
	executor := &testWorkflowExecutor{
		executionIDs:     make(map[string]string),
		capturedPayloads: make(map[string][]byte),
	}
	service := NewReplayService(repo, executor, testReplayLogger())

	// Step 1: Create webhook and event
	webhook := createTestWebhookForReplay(t, repo, ctx, tenantID, "workflow-456", true)
	originalPayload := json.RawMessage(`{"original": "data"}`)
	event := createTestEventWithPayload(t, repo, ctx, tenantID, webhook.ID, originalPayload)
	t.Logf("✓ Created event with original payload")

	// Step 2: Replay with modified payload
	modifiedPayload := json.RawMessage(`{"modified": "data"}`)
	result := service.ReplayEvent(ctx, tenantID, event.ID, modifiedPayload)

	// Step 3: Verify replay succeeded
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.ExecutionID)
	t.Logf("✓ Event replayed with modified payload")

	// Step 4: Verify modified payload was used
	capturedPayload, exists := executor.capturedPayloads[webhook.WorkflowID]
	assert.True(t, exists, "payload should have been captured")
	assert.Equal(t, string(modifiedPayload), string(capturedPayload))
	t.Logf("✓ Modified payload verified")
}

// TestIntegration_ReplayEventMaxReplayCountExceeded tests replay count enforcement
func TestIntegration_ReplayEventMaxReplayCountExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupReplayTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := createTestTenantForReplay(t, db)
	repo := NewRepository(db)
	executor := &testWorkflowExecutor{executionIDs: make(map[string]string)}
	service := NewReplayService(repo, executor, testReplayLogger())

	// Step 1: Create webhook
	webhook := createTestWebhookForReplay(t, repo, ctx, tenantID, "workflow-789", true)

	// Step 2: Create event with replay count at max
	event := createTestEvent(t, repo, ctx, tenantID, webhook.ID, EventStatusProcessed)

	// Manually set replay count to max in database
	_, err := db.Exec("UPDATE webhook_events SET replay_count = $1 WHERE id = $2", MaxReplayCount, event.ID)
	require.NoError(t, err)
	t.Logf("✓ Set replay count to max (%d)", MaxReplayCount)

	// Step 3: Try to replay
	result := service.ReplayEvent(ctx, tenantID, event.ID, nil)

	// Step 4: Verify replay failed
	require.NotNil(t, result)
	assert.False(t, result.Success, "replay should fail when max count exceeded")
	assert.Empty(t, result.ExecutionID)
	assert.Contains(t, result.Error, "max replay count")
	t.Logf("✓ Replay correctly rejected: %s", result.Error)

	// Step 5: Verify workflow was NOT executed
	_, exists := executor.executionIDs[webhook.WorkflowID]
	assert.False(t, exists, "workflow should not have been executed")
	t.Logf("✓ Workflow execution correctly prevented")
}

// TestIntegration_ReplayDisabledWebhook tests that disabled webhooks cannot be replayed
func TestIntegration_ReplayDisabledWebhook(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupReplayTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := createTestTenantForReplay(t, db)
	repo := NewRepository(db)
	executor := &testWorkflowExecutor{executionIDs: make(map[string]string)}
	service := NewReplayService(repo, executor, testReplayLogger())

	// Step 1: Create DISABLED webhook
	webhook := createTestWebhookForReplay(t, repo, ctx, tenantID, "workflow-disabled", false)
	t.Logf("✓ Created disabled webhook")

	// Step 2: Create event
	event := createTestEvent(t, repo, ctx, tenantID, webhook.ID, EventStatusProcessed)

	// Step 3: Try to replay
	result := service.ReplayEvent(ctx, tenantID, event.ID, nil)

	// Step 4: Verify replay failed
	require.NotNil(t, result)
	assert.False(t, result.Success, "replay should fail for disabled webhook")
	assert.Empty(t, result.ExecutionID)
	assert.Contains(t, result.Error, "disabled")
	t.Logf("✓ Replay correctly rejected: %s", result.Error)

	// Step 5: Verify workflow was NOT executed
	_, exists := executor.executionIDs[webhook.WorkflowID]
	assert.False(t, exists, "workflow should not have been executed")
	t.Logf("✓ Workflow execution correctly prevented")
}

// TestIntegration_BatchReplayEvents tests batch replay functionality
func TestIntegration_BatchReplayEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupReplayTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := createTestTenantForReplay(t, db)
	repo := NewRepository(db)
	executor := &testWorkflowExecutor{executionIDs: make(map[string]string)}
	service := NewReplayService(repo, executor, testReplayLogger())

	// Step 1: Create webhook
	webhook := createTestWebhookForReplay(t, repo, ctx, tenantID, "workflow-batch", true)
	t.Logf("✓ Created webhook")

	// Step 2: Create multiple events
	eventIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		event := createTestEvent(t, repo, ctx, tenantID, webhook.ID, EventStatusProcessed)
		eventIDs[i] = event.ID
	}
	t.Logf("✓ Created %d events", len(eventIDs))

	// Step 3: Batch replay
	response := service.BatchReplayEvents(ctx, tenantID, webhook.ID, eventIDs)

	// Step 4: Verify all replays succeeded
	require.NotNil(t, response)
	assert.Len(t, response.Results, 5, "should have results for all events")

	for i, eventID := range eventIDs {
		result := response.Results[eventID]
		require.NotNil(t, result, "result should exist for event %d", i)
		assert.True(t, result.Success, "replay %d should succeed", i)
		assert.NotEmpty(t, result.ExecutionID, "execution ID should exist for replay %d", i)
		assert.Empty(t, result.Error, "no error for replay %d", i)
	}
	t.Logf("✓ All %d events replayed successfully", len(eventIDs))
}

// TestIntegration_BatchReplayExceedsMaxSize tests batch size limit
func TestIntegration_BatchReplayExceedsMaxSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupReplayTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := createTestTenantForReplay(t, db)
	repo := NewRepository(db)
	executor := &testWorkflowExecutor{executionIDs: make(map[string]string)}
	service := NewReplayService(repo, executor, testReplayLogger())

	// Step 1: Create webhook
	webhook := createTestWebhookForReplay(t, repo, ctx, tenantID, "workflow-large-batch", true)

	// Step 2: Create more events than allowed (MaxBatchReplaySize = 10)
	eventIDs := make([]string, MaxBatchReplaySize+5)
	for i := 0; i < len(eventIDs); i++ {
		event := createTestEvent(t, repo, ctx, tenantID, webhook.ID, EventStatusProcessed)
		eventIDs[i] = event.ID
	}
	t.Logf("✓ Created %d events (exceeds limit of %d)", len(eventIDs), MaxBatchReplaySize)

	// Step 3: Try batch replay
	response := service.BatchReplayEvents(ctx, tenantID, webhook.ID, eventIDs)

	// Step 4: Verify all replays failed
	require.NotNil(t, response)
	assert.Len(t, response.Results, len(eventIDs), "should have results for all events")

	for i, eventID := range eventIDs {
		result := response.Results[eventID]
		require.NotNil(t, result, "result should exist for event %d", i)
		assert.False(t, result.Success, "replay %d should fail", i)
		assert.Empty(t, result.ExecutionID)
		assert.Contains(t, result.Error, "batch size exceeds maximum")
	}
	t.Logf("✓ Batch correctly rejected for exceeding size limit")

	// Step 5: Verify no workflows were executed
	assert.Len(t, executor.executionIDs, 0, "no workflows should have been executed")
	t.Logf("✓ No workflows executed")
}

// TestIntegration_ConcurrentReplay tests concurrent replay requests
func TestIntegration_ConcurrentReplay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupReplayTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := createTestTenantForReplay(t, db)
	repo := NewRepository(db)
	executor := &testWorkflowExecutor{
		executionIDs: make(map[string]string),
		mu:           sync.Mutex{},
	}
	service := NewReplayService(repo, executor, testReplayLogger())

	// Step 1: Create webhook
	webhook := createTestWebhookForReplay(t, repo, ctx, tenantID, "workflow-concurrent", true)
	t.Logf("✓ Created webhook")

	// Step 2: Create multiple events
	numEvents := 10
	events := make([]*WebhookEvent, numEvents)
	for i := 0; i < numEvents; i++ {
		events[i] = createTestEvent(t, repo, ctx, tenantID, webhook.ID, EventStatusProcessed)
	}
	t.Logf("✓ Created %d events", numEvents)

	// Step 3: Replay all events concurrently
	var wg sync.WaitGroup
	results := make([]*ReplayResult, numEvents)

	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(index int, eventID string) {
			defer wg.Done()
			results[index] = service.ReplayEvent(ctx, tenantID, eventID, nil)
		}(i, events[i].ID)
	}

	wg.Wait()
	t.Logf("✓ All concurrent replays completed")

	// Step 4: Verify all replays succeeded
	successCount := 0
	for i, result := range results {
		require.NotNil(t, result, "result %d should not be nil", i)
		if result.Success {
			successCount++
			assert.NotEmpty(t, result.ExecutionID, "execution ID should exist for successful replay %d", i)
		}
	}

	assert.Equal(t, numEvents, successCount, "all concurrent replays should succeed")
	t.Logf("✓ All %d concurrent replays succeeded", successCount)

	// Step 5: Verify correct number of executions
	executor.mu.Lock()
	executionCount := len(executor.executionIDs)
	executor.mu.Unlock()

	assert.GreaterOrEqual(t, executionCount, 1, "at least one execution should have occurred")
	t.Logf("✓ Total executions: %d", executionCount)
}

// TestIntegration_ReplayEventNotFound tests replay with non-existent event
func TestIntegration_ReplayEventNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupReplayTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := createTestTenantForReplay(t, db)
	repo := NewRepository(db)
	executor := &testWorkflowExecutor{executionIDs: make(map[string]string)}
	service := NewReplayService(repo, executor, testReplayLogger())

	// Step 1: Try to replay non-existent event
	result := service.ReplayEvent(ctx, tenantID, "non-existent-event-id", nil)

	// Step 2: Verify replay failed
	require.NotNil(t, result)
	assert.False(t, result.Success, "replay should fail for non-existent event")
	assert.Empty(t, result.ExecutionID)
	assert.Contains(t, result.Error, "not found")
	t.Logf("✓ Replay correctly rejected: %s", result.Error)

	// Step 3: Verify no executions occurred
	assert.Len(t, executor.executionIDs, 0, "no workflows should have been executed")
	t.Logf("✓ No workflows executed")
}

// TestIntegration_BatchReplayMixedSuccess tests batch replay with partial failures
func TestIntegration_BatchReplayMixedSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupReplayTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := createTestTenantForReplay(t, db)
	repo := NewRepository(db)
	executor := &testWorkflowExecutor{executionIDs: make(map[string]string)}
	service := NewReplayService(repo, executor, testReplayLogger())

	// Step 1: Create webhook
	webhook := createTestWebhookForReplay(t, repo, ctx, tenantID, "workflow-mixed", true)

	// Step 2: Create mix of valid and invalid events
	eventIDs := make([]string, 5)

	// Valid events (0-2)
	for i := 0; i < 3; i++ {
		event := createTestEvent(t, repo, ctx, tenantID, webhook.ID, EventStatusProcessed)
		eventIDs[i] = event.ID
	}

	// Event at max replay count (3)
	eventMaxReplay := createTestEvent(t, repo, ctx, tenantID, webhook.ID, EventStatusProcessed)
	eventIDs[3] = eventMaxReplay.ID
	_, err := db.Exec("UPDATE webhook_events SET replay_count = $1 WHERE id = $2", MaxReplayCount, eventMaxReplay.ID)
	require.NoError(t, err)

	// Non-existent event (4)
	eventIDs[4] = "non-existent-event-id"

	t.Logf("✓ Created 3 valid events, 1 at max replay count, 1 non-existent")

	// Step 3: Batch replay
	response := service.BatchReplayEvents(ctx, tenantID, webhook.ID, eventIDs)

	// Step 4: Verify mixed results
	require.NotNil(t, response)
	assert.Len(t, response.Results, 5)

	// First 3 should succeed
	for i := 0; i < 3; i++ {
		result := response.Results[eventIDs[i]]
		assert.True(t, result.Success, "replay %d should succeed", i)
		assert.NotEmpty(t, result.ExecutionID)
	}

	// Event at max replay count should fail
	result3 := response.Results[eventIDs[3]]
	assert.False(t, result3.Success, "replay at max count should fail")
	assert.Contains(t, result3.Error, "max replay count")

	// Non-existent event should fail
	result4 := response.Results[eventIDs[4]]
	assert.False(t, result4.Success, "non-existent event replay should fail")
	assert.Contains(t, result4.Error, "not found")

	t.Logf("✓ Mixed batch results verified: 3 success, 2 failures")
}

// TestIntegration_ReplayCreatesNewEventRecord tests that replay creates new event records
func TestIntegration_ReplayCreatesNewEventRecord(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupReplayTestDB(t)
	defer db.Close()

	ctx := context.Background()
	tenantID := createTestTenantForReplay(t, db)
	repo := NewRepository(db)
	executor := &testWorkflowExecutor{executionIDs: make(map[string]string)}
	service := NewReplayService(repo, executor, testReplayLogger())

	// Step 1: Create webhook and event
	webhook := createTestWebhookForReplay(t, repo, ctx, tenantID, "workflow-new-record", true)
	originalEvent := createTestEvent(t, repo, ctx, tenantID, webhook.ID, EventStatusProcessed)
	t.Logf("✓ Created original event: %s", originalEvent.ID)

	// Step 2: Count events before replay
	var countBefore int
	err := db.Get(&countBefore, "SELECT COUNT(*) FROM webhook_events WHERE webhook_id = $1", webhook.ID)
	require.NoError(t, err)
	t.Logf("✓ Events before replay: %d", countBefore)

	// Step 3: Replay the event
	result := service.ReplayEvent(ctx, tenantID, originalEvent.ID, nil)
	require.True(t, result.Success)
	t.Logf("✓ Event replayed successfully")

	// Step 4: Count events after replay
	var countAfter int
	err = db.Get(&countAfter, "SELECT COUNT(*) FROM webhook_events WHERE webhook_id = $1", webhook.ID)
	require.NoError(t, err)
	t.Logf("✓ Events after replay: %d", countAfter)

	// Step 5: Verify a new event record was created
	assert.Equal(t, countBefore+1, countAfter, "replay should create a new event record")

	// Step 6: Verify new event has correct attributes
	var newEvent WebhookEvent
	err = db.Get(&newEvent, `
		SELECT * FROM webhook_events
		WHERE webhook_id = $1 AND source_event_id = $2
		ORDER BY created_at DESC LIMIT 1
	`, webhook.ID, originalEvent.ID)
	require.NoError(t, err)

	assert.Equal(t, webhook.ID, newEvent.WebhookID)
	assert.Equal(t, originalEvent.ReplayCount+1, newEvent.ReplayCount)
	assert.NotNil(t, newEvent.SourceEventID)
	assert.Equal(t, originalEvent.ID, *newEvent.SourceEventID)
	t.Logf("✓ New event record verified: replay_count=%d, source_event_id=%s", newEvent.ReplayCount, *newEvent.SourceEventID)
}

// =============================================================================
// Helper Functions
// =============================================================================

func setupReplayTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sqlx.Connect("postgres", dbURL)
	require.NoError(t, err)

	// Clean up tables
	_, err = db.Exec("DELETE FROM webhook_events")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM webhook_filters")
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM webhooks")
	require.NoError(t, err)

	return db
}

func createTestTenantForReplay(t *testing.T, db *sqlx.DB) string {
	t.Helper()

	tenantID := uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO tenants (id, name, subdomain, status, tier, settings, quotas, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', 'free', '{}', '{}', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, tenantID, "Test Tenant "+tenantID[:8], "test-"+tenantID[:8])
	require.NoError(t, err)
	return tenantID
}

func createTestWebhookForReplay(t *testing.T, repo *Repository, ctx context.Context, tenantID, workflowID string, enabled bool) *Webhook {
	t.Helper()

	nodeID := uuid.New().String()
	secret := "test-secret-" + uuid.New().String()[:8]

	created, err := repo.Create(ctx, tenantID, workflowID, nodeID, secret, AuthTypeSignature)
	require.NoError(t, err)

	// If we need to disable it, update it
	if !enabled {
		err = repo.UpdateEnabled(ctx, created.ID, enabled)
		require.NoError(t, err)
		created.Enabled = enabled
	}

	return created
}

func createTestEvent(t *testing.T, repo *Repository, ctx context.Context, tenantID, webhookID string, status WebhookEventStatus) *WebhookEvent {
	t.Helper()

	payload := json.RawMessage(`{"test": "data", "timestamp": "2024-01-01T00:00:00Z"}`)
	return createTestEventWithPayload(t, repo, ctx, tenantID, webhookID, payload)
}

func createTestEventWithPayload(t *testing.T, repo *Repository, ctx context.Context, tenantID, webhookID string, payload json.RawMessage) *WebhookEvent {
	t.Helper()

	event := &WebhookEvent{
		TenantID:       tenantID,
		WebhookID:      webhookID,
		RequestMethod:  "POST",
		RequestHeaders: map[string]string{"Content-Type": "application/json"},
		RequestBody:    payload,
		Status:         EventStatusProcessed,
		ReplayCount:    0,
	}

	err := repo.CreateEvent(ctx, event)
	require.NoError(t, err)
	return event
}

func testReplayLogger() *slog.Logger {
	return slog.Default()
}

// Test workflow executor that records executions
type testWorkflowExecutor struct {
	executionIDs     map[string]string // workflowID -> executionID
	capturedPayloads map[string][]byte // workflowID -> payload
	mu               sync.Mutex
}

func (e *testWorkflowExecutor) Execute(ctx context.Context, tenantID, workflowID, triggerType string, triggerData []byte) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	executionID := fmt.Sprintf("exec-%s-%d", workflowID[:8], time.Now().UnixNano())
	e.executionIDs[workflowID] = executionID

	if e.capturedPayloads != nil {
		e.capturedPayloads[workflowID] = triggerData
	}

	return executionID, nil
}
