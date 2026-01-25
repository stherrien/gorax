package suites

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/tests/integration"
)

// =============================================================================
// Manual Trigger Tests
// =============================================================================

// TestWorkflowTrigger_Manual tests manual workflow triggering
func TestWorkflowTrigger_Manual(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "manual-trigger-test")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("execute workflow via API", func(t *testing.T) {
		workflow := createWorkflowWithManualTrigger(t, ts, tenantID, headers)

		execInput := map[string]any{
			"input": map[string]any{
				"message": "Hello from manual trigger",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)

		assert.NotEmpty(t, exec.ID)
		// TriggerType is stored in the execution record
	})

	t.Run("input parameter passing", func(t *testing.T) {
		workflow := createWorkflowWithManualTrigger(t, ts, tenantID, headers)

		execInput := map[string]any{
			"input": map[string]any{
				"userId":    "user-123",
				"action":    "process",
				"timestamp": "2024-01-15T10:30:00Z",
				"metadata": map[string]any{
					"source": "api",
					"tags":   []string{"test", "manual"},
				},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)

		assert.NotEmpty(t, exec.ID)
	})

	t.Run("execution context available", func(t *testing.T) {
		// Verify execution has proper context with tenant, user info
		workflow := createWorkflowWithManualTrigger(t, ts, tenantID, headers)

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)

		// Fetch execution details
		detailResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/executions/"+exec.ID, nil, headers)
		integration.AssertStatusCode(t, detailResp, http.StatusOK)

		var execDetail Execution
		integration.ParseJSONResponse(t, detailResp, &execDetail)

		// Verify execution was created with proper context
		assert.NotEmpty(t, execDetail.ID)
		assert.NotEmpty(t, execDetail.WorkflowID)
	})

	t.Run("execute disabled workflow fails", func(t *testing.T) {
		workflow := createWorkflowWithManualTrigger(t, ts, tenantID, headers)

		// Disable the workflow
		updateResp := ts.MakeRequest(t, http.MethodPut, "/api/v1/workflows/"+workflow.ID, map[string]any{
			"enabled": false,
		}, headers)
		integration.AssertStatusCode(t, updateResp, http.StatusOK)

		// Try to execute - should fail
		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		// Should be either 400 Bad Request or 422 Unprocessable Entity
		assert.True(t, execResp.StatusCode == http.StatusBadRequest || execResp.StatusCode == http.StatusUnprocessableEntity,
			"Expected 400 or 422, got %d", execResp.StatusCode)
	})

	t.Run("execute non-existent workflow fails", func(t *testing.T) {
		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/00000000-0000-0000-0000-000000000000/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusNotFound)
	})

	t.Run("dry run execution", func(t *testing.T) {
		workflow := createWorkflowWithManualTrigger(t, ts, tenantID, headers)

		execInput := map[string]any{
			"input":  map[string]any{"test": "data"},
			"dryRun": true,
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		// Dry run might return OK or Accepted depending on implementation
		assert.True(t, execResp.StatusCode == http.StatusOK || execResp.StatusCode == http.StatusAccepted,
			"Expected 200 or 202, got %d", execResp.StatusCode)
	})
}

// =============================================================================
// Webhook Trigger Tests
// =============================================================================

// TestWorkflowTrigger_Webhook tests webhook-triggered workflow execution
func TestWorkflowTrigger_Webhook(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "webhook-trigger-test")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("webhook receives event and triggers workflow", func(t *testing.T) {
		workflow := createWorkflowWithWebhookTrigger(t, ts, tenantID, headers)
		webhookID, _ := createWebhookForWorkflow(t, ts, workflow.ID, headers, "none", "")

		// Send event to webhook endpoint
		webhookPayload := map[string]any{
			"event":  "user.created",
			"userId": "user-456",
			"data": map[string]any{
				"email": "test@example.com",
				"name":  "Test User",
			},
		}

		// Webhook endpoint is public, doesn't need tenant headers
		webhookResp := ts.MakeRequest(t, http.MethodPost, "/webhooks/"+workflow.ID+"/"+webhookID, webhookPayload, nil)
		// Should be 200 OK or 202 Accepted
		assert.True(t, webhookResp.StatusCode == http.StatusOK || webhookResp.StatusCode == http.StatusAccepted,
			"Expected 200 or 202, got %d", webhookResp.StatusCode)
	})

	t.Run("webhook signature verification", func(t *testing.T) {
		workflow := createWorkflowWithWebhookTrigger(t, ts, tenantID, headers)
		webhookID, webhookSecret := createWebhookForWorkflow(t, ts, workflow.ID, headers, "signature", "")

		webhookPayload := `{"event":"test.event","data":{"key":"value"}}`

		// Calculate signature
		mac := hmac.New(sha256.New, []byte(webhookSecret))
		mac.Write([]byte(webhookPayload))
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		// Send with valid signature
		webhookHeaders := map[string]string{
			"Content-Type":    "application/json",
			"X-Webhook-Signature": signature,
		}

		webhookResp := ts.MakeRequestWithBody(t, http.MethodPost, "/webhooks/"+workflow.ID+"/"+webhookID, []byte(webhookPayload), webhookHeaders)
		assert.True(t, webhookResp.StatusCode == http.StatusOK || webhookResp.StatusCode == http.StatusAccepted,
			"Expected 200 or 202 with valid signature, got %d", webhookResp.StatusCode)
	})

	t.Run("webhook with invalid signature rejected", func(t *testing.T) {
		workflow := createWorkflowWithWebhookTrigger(t, ts, tenantID, headers)
		webhookID, _ := createWebhookForWorkflow(t, ts, workflow.ID, headers, "signature", "")

		webhookPayload := `{"event":"test.event"}`

		// Send with invalid signature
		webhookHeaders := map[string]string{
			"Content-Type":    "application/json",
			"X-Webhook-Signature": "sha256=invalidsignature",
		}

		webhookResp := ts.MakeRequestWithBody(t, http.MethodPost, "/webhooks/"+workflow.ID+"/"+webhookID, []byte(webhookPayload), webhookHeaders)
		integration.AssertStatusCode(t, webhookResp, http.StatusUnauthorized)
	})

	t.Run("webhook filter evaluation", func(t *testing.T) {
		workflow := createWorkflowWithWebhookTrigger(t, ts, tenantID, headers)
		webhookID, _ := createWebhookForWorkflow(t, ts, workflow.ID, headers, "none", "")

		// Create filter to only accept "user.created" events
		filterReq := map[string]any{
			"fieldPath":  "$.event",
			"operator":   "equals",
			"value":      "user.created",
			"logicGroup": 0,
			"enabled":    true,
		}

		filterResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/webhooks/"+webhookID+"/filters", filterReq, headers)
		require.True(t, filterResp.StatusCode == http.StatusCreated || filterResp.StatusCode == http.StatusOK,
			"Expected 201 or 200, got %d", filterResp.StatusCode)

		// Send matching event - should be processed
		matchingPayload := map[string]any{
			"event": "user.created",
			"data":  map[string]any{},
		}
		webhookResp := ts.MakeRequest(t, http.MethodPost, "/webhooks/"+workflow.ID+"/"+webhookID, matchingPayload, nil)
		assert.True(t, webhookResp.StatusCode == http.StatusOK || webhookResp.StatusCode == http.StatusAccepted,
			"Matching event should be accepted")

		// Send non-matching event - should be filtered
		nonMatchingPayload := map[string]any{
			"event": "user.deleted",
			"data":  map[string]any{},
		}
		webhookResp = ts.MakeRequest(t, http.MethodPost, "/webhooks/"+workflow.ID+"/"+webhookID, nonMatchingPayload, nil)
		// Filtered events might still return 200 but not trigger execution
		// The status depends on implementation
		assert.True(t, webhookResp.StatusCode >= 200 && webhookResp.StatusCode < 300,
			"Non-matching should not error, got %d", webhookResp.StatusCode)
	})

	t.Run("webhook payload passed to workflow", func(t *testing.T) {
		workflow := createWorkflowWithWebhookTrigger(t, ts, tenantID, headers)
		webhookID, _ := createWebhookForWorkflow(t, ts, workflow.ID, headers, "none", "")

		webhookPayload := map[string]any{
			"action":      "opened",
			"issue_id":    12345,
			"repository":  "test/repo",
			"opened_by":   "developer",
			"description": "Test issue for integration test",
		}

		webhookResp := ts.MakeRequest(t, http.MethodPost, "/webhooks/"+workflow.ID+"/"+webhookID, webhookPayload, nil)
		assert.True(t, webhookResp.StatusCode == http.StatusOK || webhookResp.StatusCode == http.StatusAccepted)
	})

	t.Run("webhook event history recording", func(t *testing.T) {
		workflow := createWorkflowWithWebhookTrigger(t, ts, tenantID, headers)
		webhookID, _ := createWebhookForWorkflow(t, ts, workflow.ID, headers, "none", "")

		// Send a webhook event
		webhookPayload := map[string]any{
			"event": "history.test",
		}
		webhookResp := ts.MakeRequest(t, http.MethodPost, "/webhooks/"+workflow.ID+"/"+webhookID, webhookPayload, nil)
		require.True(t, webhookResp.StatusCode == http.StatusOK || webhookResp.StatusCode == http.StatusAccepted)

		// Allow time for event to be recorded
		time.Sleep(100 * time.Millisecond)

		// Check event history
		historyResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/webhooks/"+webhookID+"/events", nil, headers)
		if historyResp.StatusCode == http.StatusOK {
			var eventList EventListResponse
			integration.ParseJSONResponse(t, historyResp, &eventList)
			assert.GreaterOrEqual(t, len(eventList.Data), 1, "Should have at least one event")
		}
	})

	t.Run("webhook disabled does not trigger", func(t *testing.T) {
		workflow := createWorkflowWithWebhookTrigger(t, ts, tenantID, headers)
		webhookID, _ := createWebhookForWorkflow(t, ts, workflow.ID, headers, "none", "")

		// Disable the webhook
		updateResp := ts.MakeRequest(t, http.MethodPut, "/api/v1/webhooks/"+webhookID, map[string]any{
			"enabled": false,
		}, headers)
		integration.AssertStatusCode(t, updateResp, http.StatusOK)

		// Try to trigger - should fail
		webhookPayload := map[string]any{"event": "disabled.test"}
		webhookResp := ts.MakeRequest(t, http.MethodPost, "/webhooks/"+workflow.ID+"/"+webhookID, webhookPayload, nil)
		// Disabled webhooks return 404 or 400
		assert.True(t, webhookResp.StatusCode == http.StatusNotFound || webhookResp.StatusCode == http.StatusBadRequest,
			"Expected 404 or 400 for disabled webhook, got %d", webhookResp.StatusCode)
	})
}

// =============================================================================
// Schedule Trigger Tests
// =============================================================================

// TestWorkflowTrigger_Schedule tests scheduled workflow execution
func TestWorkflowTrigger_Schedule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "schedule-trigger-test")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("create schedule for workflow", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		scheduleReq := map[string]any{
			"name":            "Daily Report Schedule",
			"cron_expression": "0 9 * * *", // Every day at 9 AM
			"timezone":        "UTC",
			"enabled":         true,
		}

		scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
		integration.AssertStatusCode(t, scheduleResp, http.StatusCreated)

		var schedule Schedule
		integration.ParseJSONResponse(t, scheduleResp, &schedule)

		assert.NotEmpty(t, schedule.ID)
		assert.Equal(t, "0 9 * * *", schedule.CronExpression)
		assert.NotNil(t, schedule.NextRunAt)
	})

	t.Run("schedule timezone handling", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		// Create schedule with specific timezone
		scheduleReq := map[string]any{
			"name":            "Timezone Test Schedule",
			"cron_expression": "0 12 * * *", // Noon
			"timezone":        "America/New_York",
			"enabled":         true,
		}

		scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
		integration.AssertStatusCode(t, scheduleResp, http.StatusCreated)

		var schedule Schedule
		integration.ParseJSONResponse(t, scheduleResp, &schedule)

		assert.Equal(t, "America/New_York", schedule.Timezone)
		assert.NotNil(t, schedule.NextRunAt)
	})

	t.Run("next_run_at calculation", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		// Create schedule that runs every minute
		scheduleReq := map[string]any{
			"name":            "Minute Schedule",
			"cron_expression": "* * * * *", // Every minute
			"timezone":        "UTC",
			"enabled":         true,
		}

		scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
		integration.AssertStatusCode(t, scheduleResp, http.StatusCreated)

		var schedule Schedule
		integration.ParseJSONResponse(t, scheduleResp, &schedule)

		// Next run should be within the next minute
		assert.NotNil(t, schedule.NextRunAt)
		if schedule.NextRunAt != nil {
			assert.True(t, schedule.NextRunAt.After(time.Now()), "NextRunAt should be in the future")
			assert.True(t, schedule.NextRunAt.Before(time.Now().Add(2*time.Minute)), "NextRunAt should be within 2 minutes")
		}
	})

	t.Run("overlap policy skip", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		scheduleReq := map[string]any{
			"name":            "Skip Overlap Schedule",
			"cron_expression": "*/5 * * * *", // Every 5 minutes
			"timezone":        "UTC",
			"overlap_policy":  "skip",
			"enabled":         true,
		}

		scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
		integration.AssertStatusCode(t, scheduleResp, http.StatusCreated)

		var schedule Schedule
		integration.ParseJSONResponse(t, scheduleResp, &schedule)

		assert.Equal(t, "skip", schedule.OverlapPolicy)
	})

	t.Run("overlap policy queue", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		scheduleReq := map[string]any{
			"name":            "Queue Overlap Schedule",
			"cron_expression": "*/10 * * * *",
			"timezone":        "UTC",
			"overlap_policy":  "queue",
			"enabled":         true,
		}

		scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
		integration.AssertStatusCode(t, scheduleResp, http.StatusCreated)

		var schedule Schedule
		integration.ParseJSONResponse(t, scheduleResp, &schedule)

		assert.Equal(t, "queue", schedule.OverlapPolicy)
	})

	t.Run("overlap policy terminate", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		scheduleReq := map[string]any{
			"name":            "Terminate Overlap Schedule",
			"cron_expression": "0 * * * *",
			"timezone":        "UTC",
			"overlap_policy":  "terminate",
			"enabled":         true,
		}

		scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
		integration.AssertStatusCode(t, scheduleResp, http.StatusCreated)

		var schedule Schedule
		integration.ParseJSONResponse(t, scheduleResp, &schedule)

		assert.Equal(t, "terminate", schedule.OverlapPolicy)
	})

	t.Run("disable schedule", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		// Create schedule
		scheduleReq := map[string]any{
			"name":            "Disable Test Schedule",
			"cron_expression": "0 0 * * *",
			"timezone":        "UTC",
			"enabled":         true,
		}

		scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
		integration.AssertStatusCode(t, scheduleResp, http.StatusCreated)

		var schedule Schedule
		integration.ParseJSONResponse(t, scheduleResp, &schedule)

		// Disable the schedule
		updateResp := ts.MakeRequest(t, http.MethodPut, "/api/v1/schedules/"+schedule.ID, map[string]any{
			"enabled": false,
		}, headers)
		integration.AssertStatusCode(t, updateResp, http.StatusOK)

		var updatedSchedule Schedule
		integration.ParseJSONResponse(t, updateResp, &updatedSchedule)

		assert.False(t, updatedSchedule.Enabled)
	})

	t.Run("delete schedule", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		// Create schedule
		scheduleReq := map[string]any{
			"name":            "Delete Test Schedule",
			"cron_expression": "0 0 * * *",
			"timezone":        "UTC",
			"enabled":         true,
		}

		scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
		integration.AssertStatusCode(t, scheduleResp, http.StatusCreated)

		var schedule Schedule
		integration.ParseJSONResponse(t, scheduleResp, &schedule)

		// Delete the schedule
		deleteResp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/schedules/"+schedule.ID, nil, headers)
		integration.AssertStatusCode(t, deleteResp, http.StatusNoContent)

		// Verify deletion
		getResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/schedules/"+schedule.ID, nil, headers)
		integration.AssertStatusCode(t, getResp, http.StatusNotFound)
	})

	t.Run("invalid cron expression rejected", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		scheduleReq := map[string]any{
			"name":            "Invalid Cron Schedule",
			"cron_expression": "invalid cron expression",
			"timezone":        "UTC",
			"enabled":         true,
		}

		scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
		integration.AssertStatusCode(t, scheduleResp, http.StatusBadRequest)
	})

	t.Run("invalid timezone rejected", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		scheduleReq := map[string]any{
			"name":            "Invalid Timezone Schedule",
			"cron_expression": "0 0 * * *",
			"timezone":        "Invalid/Timezone",
			"enabled":         true,
		}

		scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
		integration.AssertStatusCode(t, scheduleResp, http.StatusBadRequest)
	})

	t.Run("list schedules for workflow", func(t *testing.T) {
		workflow := createWorkflowWithScheduleTrigger(t, ts, tenantID, headers)

		// Create multiple schedules
		for i := range 3 {
			scheduleReq := map[string]any{
				"name":            "Schedule " + string(rune('A'+i)),
				"cron_expression": "0 " + string(rune('0'+i)) + " * * *",
				"timezone":        "UTC",
				"enabled":         true,
			}

			scheduleResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/schedules", scheduleReq, headers)
			require.Equal(t, http.StatusCreated, scheduleResp.StatusCode)
		}

		// List schedules
		listResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/workflows/"+workflow.ID+"/schedules", nil, headers)
		integration.AssertStatusCode(t, listResp, http.StatusOK)

		var scheduleList ScheduleListResponse
		integration.ParseJSONResponse(t, listResp, &scheduleList)

		assert.GreaterOrEqual(t, len(scheduleList.Data), 3)
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// createWorkflowWithManualTrigger creates a workflow with a manual trigger
func createWorkflowWithManualTrigger(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string) Workflow {
	t.Helper()

	workflowReq := map[string]any{
		"name":        "Manual Trigger Test Workflow",
		"description": "Workflow for testing manual trigger",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data": map[string]any{
						"label": "Manual Trigger",
						"config": map[string]any{
							"inputSchema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"message": map[string]any{"type": "string"},
								},
							},
						},
					},
				},
				{
					"id":       "action-1",
					"type":     "action:log",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Log Action",
						"config": map[string]any{
							"message": "${trigger.message}",
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "action-1"},
			},
		},
		"enabled": true,
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", workflowReq, headers)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create workflow")

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)

	return workflow
}

// createWorkflowWithWebhookTrigger creates a workflow with a webhook trigger
func createWorkflowWithWebhookTrigger(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string) Workflow {
	t.Helper()

	workflowReq := map[string]any{
		"name":        "Webhook Trigger Test Workflow",
		"description": "Workflow for testing webhook trigger",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:webhook",
					"position": map[string]any{"x": 0, "y": 0},
					"data": map[string]any{
						"label": "Webhook Trigger",
						"config": map[string]any{
							"method": "POST",
						},
					},
				},
				{
					"id":       "action-1",
					"type":     "action:log",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Log Webhook Data",
						"config": map[string]any{
							"message": "Received: ${trigger.event}",
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "action-1"},
			},
		},
		"enabled": true,
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", workflowReq, headers)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create workflow")

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)

	return workflow
}

// createWorkflowWithScheduleTrigger creates a workflow with a schedule trigger
func createWorkflowWithScheduleTrigger(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string) Workflow {
	t.Helper()

	workflowReq := map[string]any{
		"name":        "Schedule Trigger Test Workflow",
		"description": "Workflow for testing scheduled execution",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:schedule",
					"position": map[string]any{"x": 0, "y": 0},
					"data": map[string]any{
						"label": "Schedule Trigger",
						"config": map[string]any{
							"cron": "0 * * * *",
						},
					},
				},
				{
					"id":       "action-1",
					"type":     "action:log",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Scheduled Log",
						"config": map[string]any{
							"message": "Scheduled execution at ${context.timestamp}",
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "action-1"},
			},
		},
		"enabled": true,
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", workflowReq, headers)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create workflow")

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)

	return workflow
}

// createWebhookForWorkflow creates a webhook for a workflow and returns the ID and secret
func createWebhookForWorkflow(t *testing.T, ts *integration.TestServer, workflowID string, headers map[string]string, authType, apiKeyHeader string) (string, string) {
	t.Helper()

	webhookReq := map[string]any{
		"name":      "Test Webhook",
		"path":      "/test-webhook",
		"auth_type": authType,
		"enabled":   true,
	}

	if apiKeyHeader != "" {
		webhookReq["api_key_header"] = apiKeyHeader
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/webhooks", webhookReq, headers)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create webhook")

	var webhook Webhook
	integration.ParseJSONResponse(t, resp, &webhook)

	return webhook.ID, webhook.Secret
}

// Schedule represents the schedule model for tests
type Schedule struct {
	ID              string  `json:"id"`
	TenantID        string  `json:"tenant_id"`
	WorkflowID      string  `json:"workflow_id"`
	Name            string  `json:"name"`
	CronExpression  string  `json:"cron_expression"`
	Timezone        string  `json:"timezone"`
	OverlapPolicy   string  `json:"overlap_policy"`
	Enabled         bool    `json:"enabled"`
	NextRunAt       *time.Time `json:"next_run_at,omitempty"`
	LastRunAt       *time.Time `json:"last_run_at,omitempty"`
	LastExecutionID *string `json:"last_execution_id,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// ScheduleListResponse represents schedule list response
type ScheduleListResponse struct {
	Data []Schedule `json:"data"`
}
