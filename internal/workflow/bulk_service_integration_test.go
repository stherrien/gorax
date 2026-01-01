package workflow

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_BulkDeleteWorkflows tests bulk deletion of workflows
func TestIntegration_BulkDeleteWorkflows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	webhookService := &mockWebhookService{}
	bulkService := NewBulkService(repo, webhookService, testLogger())

	tenantID := "tenant-123"
	userID := "user-456"

	// Step 1: Create multiple workflows
	workflows := make([]*Workflow, 5)
	for i := 0; i < 5; i++ {
		input := CreateWorkflowInput{
			Name:        "Workflow " + string(rune('A'+i)),
			Description: "Test workflow",
			Definition:  json.RawMessage(`{"nodes":[],"edges":[]}`),
		}

		workflow, err := repo.Create(ctx, tenantID, userID, input)
		require.NoError(t, err)
		workflows[i] = workflow
	}
	t.Logf("✓ Created %d workflows", len(workflows))

	// Step 2: Bulk delete 3 workflows (some valid, some invalid)
	workflowIDs := []string{
		workflows[0].ID,
		workflows[1].ID,
		workflows[2].ID,
		"invalid-id-1",
		"invalid-id-2",
	}

	result := bulkService.BulkDelete(ctx, tenantID, workflowIDs)

	// Step 3: Verify results
	assert.Equal(t, 3, result.SuccessCount, "should successfully delete 3 valid workflows")
	assert.Len(t, result.Failures, 2, "should have 2 failures for invalid IDs")
	t.Logf("✓ Bulk delete: %d successes, %d failures", result.SuccessCount, len(result.Failures))

	// Step 4: Verify failures contain error details
	for _, failure := range result.Failures {
		assert.NotEmpty(t, failure.WorkflowID)
		assert.NotEmpty(t, failure.Error)
		t.Logf("  - Failed to delete %s: %s", failure.WorkflowID, failure.Error)
	}

	// Step 5: Verify deleted workflows are gone
	for i := 0; i < 3; i++ {
		_, err := repo.GetByID(ctx, tenantID, workflows[i].ID)
		assert.Error(t, err, "workflow %s should be deleted", workflows[i].ID)
	}

	// Step 6: Verify remaining workflows still exist
	for i := 3; i < 5; i++ {
		workflow, err := repo.GetByID(ctx, tenantID, workflows[i].ID)
		require.NoError(t, err)
		assert.Equal(t, workflows[i].ID, workflow.ID)
	}
	t.Logf("✓ Verified deleted workflows are gone, remaining workflows still exist")

	// Step 7: Verify webhooks were deleted (called for all IDs, not just successful ones)
	assert.Equal(t, 5, webhookService.deleteCount, "should attempt to delete webhooks for all 5 workflow IDs")
	t.Logf("✓ Webhook deletion called %d times", webhookService.deleteCount)
}

// TestIntegration_BulkEnableDisableWorkflows tests bulk status updates
func TestIntegration_BulkEnableDisableWorkflows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	bulkService := NewBulkService(repo, nil, testLogger())

	tenantID := "tenant-456"
	userID := "user-789"

	// Step 1: Create workflows with different statuses
	workflows := make([]*Workflow, 4)
	for i := 0; i < 4; i++ {
		input := CreateWorkflowInput{
			Name:        "Workflow " + string(rune('A'+i)),
			Description: "Test workflow",
			Definition:  json.RawMessage(`{"nodes":[],"edges":[]}`),
		}

		workflow, err := repo.Create(ctx, tenantID, userID, input)
		require.NoError(t, err)
		workflows[i] = workflow
	}
	t.Logf("✓ Created %d workflows", len(workflows))

	// Step 2: Bulk enable workflows
	workflowIDs := []string{
		workflows[0].ID,
		workflows[1].ID,
		workflows[2].ID,
	}

	enableResult := bulkService.BulkEnable(ctx, tenantID, workflowIDs)

	// Step 3: Verify enable results
	assert.Equal(t, 3, enableResult.SuccessCount)
	assert.Empty(t, enableResult.Failures)
	t.Logf("✓ Bulk enable: %d successes", enableResult.SuccessCount)

	// Step 4: Verify workflows are active
	for i := 0; i < 3; i++ {
		workflow, err := repo.GetByID(ctx, tenantID, workflows[i].ID)
		require.NoError(t, err)
		assert.Equal(t, "active", workflow.Status)
		t.Logf("  - Workflow %s status: %s", workflow.Name, workflow.Status)
	}

	// Step 5: Bulk disable workflows
	disableResult := bulkService.BulkDisable(ctx, tenantID, workflowIDs)

	// Step 6: Verify disable results
	assert.Equal(t, 3, disableResult.SuccessCount)
	assert.Empty(t, disableResult.Failures)
	t.Logf("✓ Bulk disable: %d successes", disableResult.SuccessCount)

	// Step 7: Verify workflows are inactive
	for i := 0; i < 3; i++ {
		workflow, err := repo.GetByID(ctx, tenantID, workflows[i].ID)
		require.NoError(t, err)
		assert.Equal(t, "inactive", workflow.Status)
		t.Logf("  - Workflow %s status: %s", workflow.Name, workflow.Status)
	}

	// Step 8: Verify fourth workflow remains unchanged
	workflow4, err := repo.GetByID(ctx, tenantID, workflows[3].ID)
	require.NoError(t, err)
	assert.Equal(t, "draft", workflow4.Status, "unchanged workflow should still be draft")
	t.Logf("✓ Unaffected workflow remains unchanged")
}

// TestIntegration_BulkExportWorkflows tests bulk export functionality
func TestIntegration_BulkExportWorkflows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	bulkService := NewBulkService(repo, nil, testLogger())

	tenantID := "tenant-export"
	userID := "user-export"

	// Step 1: Create workflows with different definitions
	workflows := make([]*Workflow, 3)
	definitions := []string{
		`{"nodes":[{"id":"1","type":"trigger"}],"edges":[]}`,
		`{"nodes":[{"id":"1","type":"trigger"},{"id":"2","type":"action"}],"edges":[{"source":"1","target":"2"}]}`,
		`{"nodes":[{"id":"1","type":"trigger"},{"id":"2","type":"action"},{"id":"3","type":"action"}],"edges":[]}`,
	}

	for i := 0; i < 3; i++ {
		input := CreateWorkflowInput{
			Name:        "Export Workflow " + string(rune('A'+i)),
			Description: "Workflow for export testing",
			Definition:  json.RawMessage(definitions[i]),
		}

		workflow, err := repo.Create(ctx, tenantID, userID, input)
		require.NoError(t, err)
		workflows[i] = workflow
	}
	t.Logf("✓ Created %d workflows for export", len(workflows))

	// Step 2: Export workflows
	workflowIDs := []string{
		workflows[0].ID,
		workflows[1].ID,
		workflows[2].ID,
		"invalid-id",
	}

	export, result := bulkService.BulkExport(ctx, tenantID, workflowIDs)

	// Step 3: Verify export metadata
	assert.Equal(t, "1.0", export.Version)
	assert.NotZero(t, export.ExportedAt)
	t.Logf("✓ Export version: %s, exported at: %s", export.Version, export.ExportedAt)

	// Step 4: Verify export results
	assert.Equal(t, 3, result.SuccessCount)
	assert.Len(t, result.Failures, 1, "should have 1 failure for invalid ID")
	t.Logf("✓ Export: %d successes, %d failures", result.SuccessCount, len(result.Failures))

	// Step 5: Verify exported workflows
	assert.Len(t, export.Workflows, 3)
	for i, exportItem := range export.Workflows {
		assert.Equal(t, workflows[i].ID, exportItem.ID)
		assert.Equal(t, workflows[i].Name, exportItem.Name)
		assert.Equal(t, workflows[i].Description, exportItem.Description)
		assert.NotEmpty(t, exportItem.Definition)
		assert.Equal(t, workflows[i].Status, exportItem.Status)
		assert.Equal(t, workflows[i].Version, exportItem.Version)

		// Verify definition is valid JSON
		var def map[string]interface{}
		err := json.Unmarshal(exportItem.Definition, &def)
		require.NoError(t, err, "exported definition should be valid JSON")

		t.Logf("  - Exported: %s (%s)", exportItem.Name, exportItem.ID)
	}

	// Step 6: Verify export can be serialized to JSON
	exportJSON, err := json.Marshal(export)
	require.NoError(t, err)
	assert.NotEmpty(t, exportJSON)
	t.Logf("✓ Export serialized to JSON (%d bytes)", len(exportJSON))
}

// TestIntegration_BulkCloneWorkflows tests bulk cloning functionality
func TestIntegration_BulkCloneWorkflows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	webhookService := &mockWebhookService{}
	bulkService := NewBulkService(repo, webhookService, testLogger())

	tenantID := "tenant-clone"
	userID := "user-clone"

	// Step 1: Create source workflows
	sourceWorkflows := make([]*Workflow, 3)
	for i := 0; i < 3; i++ {
		definition := json.RawMessage(`{
			"nodes": [
				{"id":"` + string(rune('1'+i)) + `","type":"trigger:webhook","data":{"nodeType":"webhook"}}
			],
			"edges": []
		}`)

		input := CreateWorkflowInput{
			Name:        "Source Workflow " + string(rune('A'+i)),
			Description: "Original workflow to be cloned",
			Definition:  definition,
		}

		workflow, err := repo.Create(ctx, tenantID, userID, input)
		require.NoError(t, err)
		sourceWorkflows[i] = workflow
	}
	t.Logf("✓ Created %d source workflows", len(sourceWorkflows))

	// Step 2: Bulk clone workflows
	workflowIDs := []string{
		sourceWorkflows[0].ID,
		sourceWorkflows[1].ID,
		"invalid-id",
	}

	clones, result := bulkService.BulkClone(ctx, tenantID, userID, workflowIDs)

	// Step 3: Verify clone results
	assert.Equal(t, 2, result.SuccessCount)
	assert.Len(t, result.Failures, 1, "should have 1 failure for invalid ID")
	t.Logf("✓ Clone: %d successes, %d failures", result.SuccessCount, len(result.Failures))

	// Step 4: Verify cloned workflows
	assert.Len(t, clones, 2)
	for i, clone := range clones {
		// Verify clone has different ID
		assert.NotEqual(t, sourceWorkflows[i].ID, clone.ID, "clone should have different ID")

		// Verify clone has " (Copy)" appended to name
		expectedName := sourceWorkflows[i].Name + " (Copy)"
		assert.Equal(t, expectedName, clone.Name)

		// Verify clone has same description
		assert.Equal(t, sourceWorkflows[i].Description, clone.Description)

		// Verify clone has same definition
		assert.JSONEq(t, string(sourceWorkflows[i].Definition), string(clone.Definition))

		// Verify clone is in draft status
		assert.Equal(t, "draft", clone.Status, "clones should be created as drafts")

		t.Logf("  - Cloned: %s -> %s (ID: %s)", sourceWorkflows[i].Name, clone.Name, clone.ID)
	}

	// Step 5: Verify webhook sync was called
	assert.Equal(t, 2, webhookService.syncCount, "should sync webhooks for 2 clones")
	t.Logf("✓ Webhook sync called %d times", webhookService.syncCount)

	// Step 6: Verify original workflows unchanged
	for _, original := range sourceWorkflows {
		workflow, err := repo.GetByID(ctx, tenantID, original.ID)
		require.NoError(t, err)
		assert.Equal(t, original.Name, workflow.Name)
		assert.Equal(t, original.Status, workflow.Status)
	}
	t.Logf("✓ Original workflows remain unchanged")
}

// TestIntegration_BulkOperationsWithPartialFailures tests error handling
func TestIntegration_BulkOperationsWithPartialFailures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	bulkService := NewBulkService(repo, nil, testLogger())

	tenantID := "tenant-partial"
	userID := "user-partial"

	// Step 1: Create some valid workflows
	validWorkflows := make([]*Workflow, 2)
	for i := 0; i < 2; i++ {
		input := CreateWorkflowInput{
			Name:        "Valid Workflow " + string(rune('A'+i)),
			Description: "Valid workflow",
			Definition:  json.RawMessage(`{"nodes":[],"edges":[]}`),
		}

		workflow, err := repo.Create(ctx, tenantID, userID, input)
		require.NoError(t, err)
		validWorkflows[i] = workflow
	}
	t.Logf("✓ Created %d valid workflows", len(validWorkflows))

	// Step 2: Mix valid and invalid IDs
	mixedIDs := []string{
		validWorkflows[0].ID, // Valid
		"invalid-1",          // Invalid
		validWorkflows[1].ID, // Valid
		"invalid-2",          // Invalid
		"invalid-3",          // Invalid
	}

	// Step 3: Bulk enable with mixed IDs
	result := bulkService.BulkEnable(ctx, tenantID, mixedIDs)

	// Step 4: Verify partial success
	assert.Equal(t, 2, result.SuccessCount, "should succeed for 2 valid IDs")
	assert.Len(t, result.Failures, 3, "should fail for 3 invalid IDs")
	t.Logf("✓ Partial success: %d successes, %d failures", result.SuccessCount, len(result.Failures))

	// Step 5: Verify each failure has details
	invalidIDs := map[string]bool{"invalid-1": true, "invalid-2": true, "invalid-3": true}
	for _, failure := range result.Failures {
		assert.True(t, invalidIDs[failure.WorkflowID], "failure should be for invalid ID")
		assert.NotEmpty(t, failure.Error)
		t.Logf("  - Failed: %s - %s", failure.WorkflowID, failure.Error)
	}

	// Step 6: Verify valid workflows were updated
	for _, workflow := range validWorkflows {
		updated, err := repo.GetByID(ctx, tenantID, workflow.ID)
		require.NoError(t, err)
		assert.Equal(t, "active", updated.Status, "valid workflow should be enabled")
	}
	t.Logf("✓ Valid workflows were updated despite partial failures")
}

// TestIntegration_BulkCloneWithWebhooks tests webhook handling in clones
func TestIntegration_BulkCloneWithWebhooks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	repo := setupTestRepository(t)
	webhookService := &mockWebhookService{}
	bulkService := NewBulkService(repo, webhookService, testLogger())

	tenantID := "tenant-webhook-clone"
	userID := "user-webhook-clone"

	// Step 1: Create workflow with webhook nodes
	definition := json.RawMessage(`{
		"nodes": [
			{
				"id": "node-1",
				"type": "trigger:webhook",
				"data": {
					"nodeType": "webhook",
					"config": {"auth_type": "signature"}
				}
			},
			{
				"id": "node-2",
				"type": "trigger:webhook",
				"data": {
					"nodeType": "webhook",
					"config": {"auth_type": "none"}
				}
			},
			{
				"id": "node-3",
				"type": "action",
				"data": {"nodeType": "http"}
			}
		],
		"edges": []
	}`)

	input := CreateWorkflowInput{
		Name:        "Webhook Workflow",
		Description: "Workflow with multiple webhook triggers",
		Definition:  definition,
	}

	original, err := repo.Create(ctx, tenantID, userID, input)
	require.NoError(t, err)
	t.Logf("✓ Created workflow with webhook nodes")

	// Step 2: Clone workflow
	clones, result := bulkService.BulkClone(ctx, tenantID, userID, []string{original.ID})

	// Step 3: Verify clone succeeded
	require.Equal(t, 1, result.SuccessCount)
	require.Len(t, clones, 1)
	_ = clones[0] // Verify clone exists
	t.Logf("✓ Workflow cloned successfully")

	// Step 4: Verify webhook sync was called with correct nodes
	assert.Equal(t, 1, webhookService.syncCount)
	assert.Len(t, webhookService.lastWebhookNodes, 2, "should extract 2 webhook nodes")

	// Verify webhook nodes
	nodeIDs := make(map[string]bool)
	for _, node := range webhookService.lastWebhookNodes {
		nodeIDs[node.NodeID] = true
		assert.NotEmpty(t, node.AuthType)
	}
	assert.True(t, nodeIDs["node-1"], "should include node-1")
	assert.True(t, nodeIDs["node-2"], "should include node-2")
	assert.False(t, nodeIDs["node-3"], "should not include non-webhook node-3")
	t.Logf("✓ Webhook sync called with correct %d webhook nodes", len(webhookService.lastWebhookNodes))
}

// Mock webhook service for testing
type mockWebhookService struct {
	deleteCount      int
	syncCount        int
	lastWebhookNodes []WebhookNodeConfig
}

func (m *mockWebhookService) DeleteByWorkflowID(ctx context.Context, workflowID string) error {
	m.deleteCount++
	return nil
}

func (m *mockWebhookService) SyncWorkflowWebhooks(ctx context.Context, tenantID, workflowID string, webhookNodes []WebhookNodeConfig) error {
	m.syncCount++
	m.lastWebhookNodes = webhookNodes
	return nil
}

func (m *mockWebhookService) GetByWorkflowID(ctx context.Context, workflowID string) ([]*WebhookInfo, error) {
	return nil, nil
}

// Helper functions
func setupTestRepository(t *testing.T) RepositoryInterface {
	t.Helper()
	return &mockRepository{
		workflows: make(map[string]*Workflow),
		nextID:    0,
	}
}

func testLogger() *slog.Logger {
	return slog.Default()
}

// Mock repository implementation
type mockRepository struct {
	workflows map[string]*Workflow
	nextID    int
}

func (m *mockRepository) Create(ctx context.Context, tenantID, userID string, input CreateWorkflowInput) (*Workflow, error) {
	m.nextID++
	workflow := &Workflow{
		ID:          "wf-" + string(rune('A'+m.nextID)),
		TenantID:    tenantID,
		Name:        input.Name,
		Description: input.Description,
		Definition:  input.Definition,
		Status:      "draft",
		Version:     1,
	}
	m.workflows[workflow.ID] = workflow
	return workflow, nil
}

func (m *mockRepository) GetByID(ctx context.Context, tenantID, id string) (*Workflow, error) {
	workflow, ok := m.workflows[id]
	if !ok {
		return nil, ErrWorkflowNotFound
	}
	if workflow.TenantID != tenantID {
		return nil, ErrWorkflowNotFound
	}
	return workflow, nil
}

func (m *mockRepository) Update(ctx context.Context, tenantID, id string, input UpdateWorkflowInput) (*Workflow, error) {
	workflow, err := m.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	if input.Name != "" {
		workflow.Name = input.Name
	}
	if input.Description != "" {
		workflow.Description = input.Description
	}
	if input.Definition != nil {
		workflow.Definition = input.Definition
	}
	if input.Status != "" {
		workflow.Status = input.Status
	}

	workflow.Version++
	m.workflows[id] = workflow
	return workflow, nil
}

func (m *mockRepository) Delete(ctx context.Context, tenantID, id string) error {
	_, err := m.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	delete(m.workflows, id)
	return nil
}

func (m *mockRepository) List(ctx context.Context, tenantID string, limit, offset int) ([]*Workflow, error) {
	var results []*Workflow
	for _, w := range m.workflows {
		if w.TenantID == tenantID {
			results = append(results, w)
		}
	}
	return results, nil
}

func (m *mockRepository) CreateExecution(ctx context.Context, tenantID, workflowID string, workflowVersion int, triggerType string, triggerData []byte) (*Execution, error) {
	return nil, nil
}

func (m *mockRepository) GetExecutionByID(ctx context.Context, tenantID, id string) (*Execution, error) {
	return nil, nil
}

func (m *mockRepository) UpdateExecutionStatus(ctx context.Context, id string, status ExecutionStatus, outputData []byte, errorMessage *string) error {
	return nil
}

func (m *mockRepository) GetStepExecutionsByExecutionID(ctx context.Context, executionID string) ([]*StepExecution, error) {
	return nil, nil
}

func (m *mockRepository) ListExecutions(ctx context.Context, tenantID string, workflowID string, limit, offset int) ([]*Execution, error) {
	return nil, nil
}

func (m *mockRepository) ListExecutionsAdvanced(ctx context.Context, tenantID string, filter ExecutionFilter, cursor string, limit int) (*ExecutionListResult, error) {
	return nil, nil
}

func (m *mockRepository) GetExecutionWithSteps(ctx context.Context, tenantID, executionID string) (*ExecutionWithSteps, error) {
	return nil, nil
}

func (m *mockRepository) CountExecutions(ctx context.Context, tenantID string, filter ExecutionFilter) (int, error) {
	return 0, nil
}

func (m *mockRepository) CreateWorkflowVersion(ctx context.Context, workflowID string, version int, definition json.RawMessage, createdBy string) (*WorkflowVersion, error) {
	return nil, nil
}

func (m *mockRepository) ListWorkflowVersions(ctx context.Context, workflowID string) ([]*WorkflowVersion, error) {
	return nil, nil
}

func (m *mockRepository) GetWorkflowVersion(ctx context.Context, workflowID string, version int) (*WorkflowVersion, error) {
	return nil, nil
}

func (m *mockRepository) RestoreWorkflowVersion(ctx context.Context, tenantID, workflowID string, version int) (*Workflow, error) {
	return nil, nil
}

// Error definitions
var ErrWorkflowNotFound = &workflowError{message: "workflow not found"}

type workflowError struct {
	message string
}

func (e *workflowError) Error() string {
	return e.message
}
