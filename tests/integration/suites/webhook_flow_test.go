package suites

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/tests/integration"
)

// Webhook represents the webhook model for tests
type Webhook struct {
	ID              string  `json:"id"`
	TenantID        string  `json:"tenant_id"`
	WorkflowID      string  `json:"workflow_id"`
	NodeID          string  `json:"node_id"`
	Name            string  `json:"name"`
	Path            string  `json:"path"`
	Secret          string  `json:"secret,omitempty"`
	AuthType        string  `json:"auth_type"`
	Description     string  `json:"description"`
	Priority        int     `json:"priority"`
	Enabled         bool    `json:"enabled"`
	TriggerCount    int     `json:"trigger_count"`
	LastTriggeredAt *string `json:"last_triggered_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// WebhookEvent represents a webhook event log
type WebhookEvent struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenantId"`
	WebhookID      string          `json:"webhookId"`
	ExecutionID    *string         `json:"executionId,omitempty"`
	RequestMethod  string          `json:"requestMethod"`
	RequestHeaders map[string]any  `json:"requestHeaders"`
	RequestBody    json.RawMessage `json:"requestBody"`
	Status         string          `json:"status"`
	ErrorMessage   *string         `json:"errorMessage,omitempty"`
	CreatedAt      string          `json:"createdAt"`
}

// WebhookFilter represents a filter for webhook payload evaluation
type WebhookFilter struct {
	ID         string `json:"id"`
	WebhookID  string `json:"webhookId"`
	FieldPath  string `json:"fieldPath"`
	Operator   string `json:"operator"`
	Value      any    `json:"value"`
	LogicGroup int    `json:"logicGroup"`
	Enabled    bool   `json:"enabled"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// FilterResult represents the result of testing filters
type FilterResult struct {
	Passed  bool           `json:"passed"`
	Reason  string         `json:"reason"`
	Details map[string]any `json:"details"`
}

// WebhookListResponse represents the list response
type WebhookListResponse struct {
	Data []Webhook `json:"data"`
}

// FilterListResponse represents filter list response
type FilterListResponse struct {
	Data []WebhookFilter `json:"data"`
}

// EventListResponse represents event list response
type EventListResponse struct {
	Data []WebhookEvent `json:"data"`
}

// TestWebhookFlow_CreateAndConfigure tests webhook creation and configuration
func TestWebhookFlow_CreateAndConfigure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	// Create test tenant and workflow
	tenantID := ts.CreateTestTenant(t, "webhook-test-tenant")
	workflowID := createTestWorkflowForWebhook(t, ts, tenantID)
	headers := integration.DefaultTestHeaders(tenantID)

	var webhookID string

	// Step 1: Create webhook
	t.Run("Step 1: Create webhook with signature auth", func(t *testing.T) {
		createReq := map[string]any{
			"name":        "GitHub Integration",
			"path":        "/github-events",
			"auth_type":   "signature",
			"description": "Receives GitHub webhook events",
			"enabled":     true,
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/webhooks", createReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusCreated)

		var webhook Webhook
		integration.ParseJSONResponse(t, resp, &webhook)

		assert.NotEmpty(t, webhook.ID)
		assert.Equal(t, "GitHub Integration", webhook.Name)
		assert.Equal(t, "signature", webhook.AuthType)
		assert.True(t, webhook.Enabled)
		assert.NotEmpty(t, webhook.Secret, "Signature auth should generate a secret")

		webhookID = webhook.ID
	})

	// Step 2: Get webhook details
	t.Run("Step 2: Get webhook details", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/webhooks/"+webhookID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var webhook Webhook
		integration.ParseJSONResponse(t, resp, &webhook)

		assert.Equal(t, webhookID, webhook.ID)
		assert.Equal(t, "GitHub Integration", webhook.Name)
	})

	// Step 3: List webhooks for workflow
	t.Run("Step 3: List webhooks for workflow", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/workflows/"+workflowID+"/webhooks", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var listResp WebhookListResponse
		integration.ParseJSONResponse(t, resp, &listResp)

		assert.GreaterOrEqual(t, len(listResp.Data), 1)
		found := false
		for _, wh := range listResp.Data {
			if wh.ID == webhookID {
				found = true
				break
			}
		}
		assert.True(t, found, "Webhook should be in list")
	})

	// Step 4: Update webhook
	t.Run("Step 4: Update webhook", func(t *testing.T) {
		updateReq := map[string]any{
			"name":        "GitHub Integration v2",
			"description": "Updated description",
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/webhooks/"+webhookID, updateReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var webhook Webhook
		integration.ParseJSONResponse(t, resp, &webhook)

		assert.Equal(t, "GitHub Integration v2", webhook.Name)
		assert.Equal(t, "Updated description", webhook.Description)
	})

	// Step 5: Disable webhook
	t.Run("Step 5: Disable webhook", func(t *testing.T) {
		updateReq := map[string]any{
			"enabled": false,
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/webhooks/"+webhookID, updateReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var webhook Webhook
		integration.ParseJSONResponse(t, resp, &webhook)

		assert.False(t, webhook.Enabled)
	})

	// Step 6: Re-enable webhook
	t.Run("Step 6: Re-enable webhook", func(t *testing.T) {
		updateReq := map[string]any{
			"enabled": true,
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/webhooks/"+webhookID, updateReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var webhook Webhook
		integration.ParseJSONResponse(t, resp, &webhook)

		assert.True(t, webhook.Enabled)
	})

	// Cleanup
	t.Run("Cleanup: Delete webhook", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/webhooks/"+webhookID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)

		// Verify deletion
		resp = ts.MakeRequest(t, http.MethodGet, "/api/v1/webhooks/"+webhookID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusNotFound)
	})
}

// TestWebhookFlow_FilterConfiguration tests webhook filter management
func TestWebhookFlow_FilterConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	tenantID := ts.CreateTestTenant(t, "webhook-filter-test")
	workflowID := createTestWorkflowForWebhook(t, ts, tenantID)
	headers := integration.DefaultTestHeaders(tenantID)

	// Create webhook first
	var webhookID string
	t.Run("Setup: Create webhook", func(t *testing.T) {
		createReq := map[string]any{
			"name":      "Filtered Webhook",
			"path":      "/filtered-events",
			"auth_type": "none",
			"enabled":   true,
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/webhooks", createReq, headers)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var webhook Webhook
		integration.ParseJSONResponse(t, resp, &webhook)
		webhookID = webhook.ID
	})

	var filterID string

	// Step 1: Create filter
	t.Run("Step 1: Create filter", func(t *testing.T) {
		createReq := map[string]any{
			"fieldPath":  "$.action",
			"operator":   "equals",
			"value":      "opened",
			"logicGroup": 0,
			"enabled":    true,
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/webhooks/"+webhookID+"/filters", createReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusCreated)

		var result struct {
			Data WebhookFilter `json:"data"`
		}
		integration.ParseJSONResponse(t, resp, &result)

		assert.NotEmpty(t, result.Data.ID)
		assert.Equal(t, "$.action", result.Data.FieldPath)
		assert.Equal(t, "equals", result.Data.Operator)
		assert.True(t, result.Data.Enabled)

		filterID = result.Data.ID
	})

	// Step 2: List filters
	t.Run("Step 2: List filters", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/webhooks/"+webhookID+"/filters", nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var result FilterListResponse
		integration.ParseJSONResponse(t, resp, &result)

		assert.GreaterOrEqual(t, len(result.Data), 1)
	})

	// Step 3: Get single filter
	t.Run("Step 3: Get single filter", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/webhooks/"+webhookID+"/filters/"+filterID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var result struct {
			Data WebhookFilter `json:"data"`
		}
		integration.ParseJSONResponse(t, resp, &result)

		assert.Equal(t, filterID, result.Data.ID)
	})

	// Step 4: Update filter
	t.Run("Step 4: Update filter", func(t *testing.T) {
		updateReq := map[string]any{
			"fieldPath":  "$.action",
			"operator":   "in",
			"value":      []string{"opened", "reopened"},
			"logicGroup": 0,
			"enabled":    true,
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/webhooks/"+webhookID+"/filters/"+filterID, updateReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var result struct {
			Data WebhookFilter `json:"data"`
		}
		integration.ParseJSONResponse(t, resp, &result)

		assert.Equal(t, "in", result.Data.Operator)
	})

	// Step 5: Test filters with matching payload
	t.Run("Step 5: Test filters with matching payload", func(t *testing.T) {
		testReq := map[string]any{
			"payload": map[string]any{
				"action": "opened",
				"issue": map[string]any{
					"number": 123,
					"title":  "Test Issue",
				},
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/webhooks/"+webhookID+"/filters/test", testReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var result FilterResult
		integration.ParseJSONResponse(t, resp, &result)

		assert.True(t, result.Passed, "Filter should pass with matching payload")
	})

	// Step 6: Test filters with non-matching payload
	t.Run("Step 6: Test filters with non-matching payload", func(t *testing.T) {
		testReq := map[string]any{
			"payload": map[string]any{
				"action": "closed",
				"issue": map[string]any{
					"number": 123,
				},
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/webhooks/"+webhookID+"/filters/test", testReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var result FilterResult
		integration.ParseJSONResponse(t, resp, &result)

		assert.False(t, result.Passed, "Filter should not pass with non-matching payload")
	})

	// Step 7: Disable filter
	t.Run("Step 7: Disable filter", func(t *testing.T) {
		updateReq := map[string]any{
			"fieldPath":  "$.action",
			"operator":   "in",
			"value":      []string{"opened", "reopened"},
			"logicGroup": 0,
			"enabled":    false,
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/webhooks/"+webhookID+"/filters/"+filterID, updateReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var result struct {
			Data WebhookFilter `json:"data"`
		}
		integration.ParseJSONResponse(t, resp, &result)

		assert.False(t, result.Data.Enabled)
	})

	// Step 8: Delete filter
	t.Run("Step 8: Delete filter", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/webhooks/"+webhookID+"/filters/"+filterID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)

		// Verify deletion
		resp = ts.MakeRequest(t, http.MethodGet, "/api/v1/webhooks/"+webhookID+"/filters/"+filterID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusNotFound)
	})

	// Cleanup
	t.Run("Cleanup: Delete webhook", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/webhooks/"+webhookID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)
	})
}

// TestWebhookFlow_AuthTypes tests different webhook authentication types
func TestWebhookFlow_AuthTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	tenantID := ts.CreateTestTenant(t, "webhook-auth-test")
	workflowID := createTestWorkflowForWebhook(t, ts, tenantID)
	headers := integration.DefaultTestHeaders(tenantID)

	authTypes := []struct {
		name         string
		authType     string
		expectSecret bool
		extraFields  map[string]any
	}{
		{
			name:         "None Auth",
			authType:     "none",
			expectSecret: false,
		},
		{
			name:         "Signature Auth",
			authType:     "signature",
			expectSecret: true,
		},
		{
			name:         "API Key Auth",
			authType:     "api_key",
			expectSecret: true,
		},
	}

	for _, tc := range authTypes {
		t.Run("Create webhook with "+tc.name, func(t *testing.T) {
			createReq := map[string]any{
				"name":      tc.name + " Webhook",
				"path":      "/" + tc.authType + "-webhook",
				"auth_type": tc.authType,
				"enabled":   true,
			}

			// Add extra fields if any
			for k, v := range tc.extraFields {
				createReq[k] = v
			}

			resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/webhooks", createReq, headers)
			integration.AssertStatusCode(t, resp, http.StatusCreated)

			var webhook Webhook
			integration.ParseJSONResponse(t, resp, &webhook)

			assert.Equal(t, tc.authType, webhook.AuthType)
			if tc.expectSecret {
				assert.NotEmpty(t, webhook.Secret, "%s should have a secret", tc.name)
			}

			// Cleanup
			resp = ts.MakeRequest(t, http.MethodDelete, "/api/v1/webhooks/"+webhook.ID, nil, headers)
			require.Equal(t, http.StatusNoContent, resp.StatusCode)
			resp.Body.Close()
		})
	}
}

// TestWebhookFlow_ValidationErrors tests webhook validation
func TestWebhookFlow_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	tenantID := ts.CreateTestTenant(t, "webhook-validation-test")
	workflowID := createTestWorkflowForWebhook(t, ts, tenantID)
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("Missing name", func(t *testing.T) {
		createReq := map[string]any{
			"path":      "/test-webhook",
			"auth_type": "none",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/webhooks", createReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusBadRequest)
	})

	t.Run("Invalid auth type", func(t *testing.T) {
		createReq := map[string]any{
			"name":      "Test Webhook",
			"path":      "/test-webhook",
			"auth_type": "invalid_auth",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/webhooks", createReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusBadRequest)
	})

	t.Run("Non-existent workflow", func(t *testing.T) {
		createReq := map[string]any{
			"name":      "Test Webhook",
			"path":      "/test-webhook",
			"auth_type": "none",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/00000000-0000-0000-0000-000000000000/webhooks", createReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusNotFound)
	})

	t.Run("Non-existent webhook for update", func(t *testing.T) {
		updateReq := map[string]any{
			"name": "Updated Name",
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/webhooks/00000000-0000-0000-0000-000000000000", updateReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusNotFound)
	})
}

// TestWebhookFlow_TenantIsolation tests webhook tenant isolation
func TestWebhookFlow_TenantIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	// Create two tenants
	tenant1ID := ts.CreateTestTenant(t, "webhook-tenant-1")
	tenant2ID := ts.CreateTestTenant(t, "webhook-tenant-2")
	workflow1ID := createTestWorkflowForWebhook(t, ts, tenant1ID)
	headers1 := integration.DefaultTestHeaders(tenant1ID)
	headers2 := integration.DefaultTestHeaders(tenant2ID)

	// Create webhook in tenant 1
	var webhook1ID string
	t.Run("Create webhook in tenant 1", func(t *testing.T) {
		createReq := map[string]any{
			"name":      "Tenant 1 Webhook",
			"path":      "/tenant1-webhook",
			"auth_type": "none",
			"enabled":   true,
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow1ID+"/webhooks", createReq, headers1)
		integration.AssertStatusCode(t, resp, http.StatusCreated)

		var webhook Webhook
		integration.ParseJSONResponse(t, resp, &webhook)
		webhook1ID = webhook.ID
	})

	t.Run("Tenant 1 can access its webhook", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/webhooks/"+webhook1ID, nil, headers1)
		integration.AssertStatusCode(t, resp, http.StatusOK)
	})

	t.Run("Tenant 2 cannot access tenant 1's webhook", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/webhooks/"+webhook1ID, nil, headers2)
		integration.AssertStatusCode(t, resp, http.StatusNotFound)
	})

	t.Run("Tenant 2 cannot update tenant 1's webhook", func(t *testing.T) {
		updateReq := map[string]any{
			"name": "Hijacked Webhook",
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/webhooks/"+webhook1ID, updateReq, headers2)
		integration.AssertStatusCode(t, resp, http.StatusNotFound)
	})

	t.Run("Tenant 2 cannot delete tenant 1's webhook", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/webhooks/"+webhook1ID, nil, headers2)
		integration.AssertStatusCode(t, resp, http.StatusNotFound)
	})

	// Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/webhooks/"+webhook1ID, nil, headers1)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)
	})
}

// TestWebhookFlow_MultipleFiltersLogicGroups tests filter logic groups
func TestWebhookFlow_MultipleFiltersLogicGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	tenantID := ts.CreateTestTenant(t, "webhook-logic-test")
	workflowID := createTestWorkflowForWebhook(t, ts, tenantID)
	headers := integration.DefaultTestHeaders(tenantID)

	// Create webhook
	var webhookID string
	t.Run("Setup: Create webhook", func(t *testing.T) {
		createReq := map[string]any{
			"name":      "Multi-Filter Webhook",
			"path":      "/multi-filter",
			"auth_type": "none",
			"enabled":   true,
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/webhooks", createReq, headers)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var webhook Webhook
		integration.ParseJSONResponse(t, resp, &webhook)
		webhookID = webhook.ID
	})

	// Create multiple filters in same logic group (AND logic)
	t.Run("Create filters in same logic group", func(t *testing.T) {
		// Filter 1: action equals "opened"
		filter1 := map[string]any{
			"fieldPath":  "$.action",
			"operator":   "equals",
			"value":      "opened",
			"logicGroup": 0,
			"enabled":    true,
		}
		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/webhooks/"+webhookID+"/filters", filter1, headers)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()

		// Filter 2: repository.name contains "gorax"
		filter2 := map[string]any{
			"fieldPath":  "$.repository.name",
			"operator":   "contains",
			"value":      "gorax",
			"logicGroup": 0,
			"enabled":    true,
		}
		resp = ts.MakeRequest(t, http.MethodPost, "/api/v1/webhooks/"+webhookID+"/filters", filter2, headers)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	})

	// Test with payload matching both filters
	t.Run("Test payload matching all filters", func(t *testing.T) {
		testReq := map[string]any{
			"payload": map[string]any{
				"action": "opened",
				"repository": map[string]any{
					"name": "gorax-project",
				},
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/webhooks/"+webhookID+"/filters/test", testReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var result FilterResult
		integration.ParseJSONResponse(t, resp, &result)
		assert.True(t, result.Passed, "Should pass when all filters match")
	})

	// Test with payload matching only one filter
	t.Run("Test payload matching only one filter", func(t *testing.T) {
		testReq := map[string]any{
			"payload": map[string]any{
				"action": "opened",
				"repository": map[string]any{
					"name": "other-project",
				},
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/webhooks/"+webhookID+"/filters/test", testReq, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var result FilterResult
		integration.ParseJSONResponse(t, resp, &result)
		// With AND logic (same group), should fail if one doesn't match
		assert.False(t, result.Passed, "Should fail when one filter doesn't match in AND group")
	})

	// Cleanup
	t.Run("Cleanup", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/webhooks/"+webhookID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)
	})
}

// createTestWorkflowForWebhook creates a test workflow and returns its ID
func createTestWorkflowForWebhook(t *testing.T, ts *integration.TestServer, tenantID string) string {
	t.Helper()

	query := `
		INSERT INTO workflows (id, tenant_id, name, description, version, status, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'Webhook Test Workflow', 'A workflow for webhook testing', 1, 'draft', NOW(), NOW())
		RETURNING id
	`

	var workflowID string
	err := ts.DB.QueryRow(query, tenantID).Scan(&workflowID)
	require.NoError(t, err, "failed to create test workflow")

	t.Logf("Created test workflow for webhook: %s", workflowID)
	return workflowID
}
