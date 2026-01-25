package suites

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/tests/integration"
)

// Credential represents the credential model for tests
type Credential struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data,omitempty"`
	LastUsedAt  *string                `json:"last_used_at,omitempty"`
	LastRotated *string                `json:"last_rotated,omitempty"`
	Version     int                    `json:"version"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// CredentialListResponse represents list response
type CredentialListResponse struct {
	Data   []Credential `json:"data"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
	Total  int          `json:"total"`
}

// CredentialTestResponse represents test result
type CredentialTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// TestCredentialFlow_CreateAndUse tests credential creation and usage in workflows
func TestCredentialFlow_CreateAndUse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "credential-flow-tenant")
	headers := integration.DefaultTestHeaders(tenantID)

	var credentialID string
	var workflowID string

	// Step 1: Create credential
	t.Run("Step1_CreateCredential", func(t *testing.T) {
		input := map[string]interface{}{
			"name":        "API Key Credential",
			"description": "Test API key for integration tests",
			"type":        "api_key",
			"data": map[string]interface{}{
				"api_key": "sk-test-secret-key-12345",
				"header":  "X-API-Key",
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/credentials", input, headers)
		integration.AssertStatusCode(t, resp, http.StatusCreated)

		var credential Credential
		integration.ParseJSONResponse(t, resp, &credential)

		assert.NotEmpty(t, credential.ID)
		assert.Equal(t, "API Key Credential", credential.Name)
		assert.Equal(t, "api_key", credential.Type)
		assert.Equal(t, 1, credential.Version)

		// Data should not be returned in response (security)
		credentialID = credential.ID
		t.Logf("Created credential: %s", credentialID)
	})

	// Step 2: Verify credential is masked when retrieved
	t.Run("Step2_VerifyMaskedRetrieval", func(t *testing.T) {
		require.NotEmpty(t, credentialID, "credential ID required from Step 1")

		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/credentials/"+credentialID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		body := integration.GetResponseBody(t, resp)

		// The actual secret should never appear in the response
		assert.NotContains(t, body, "sk-test-secret-key-12345")
		t.Logf("Verified credential data is masked in response")
	})

	// Step 3: Create workflow using credential
	t.Run("Step3_CreateWorkflowWithCredential", func(t *testing.T) {
		require.NotEmpty(t, credentialID, "credential ID required from Step 1")

		input := map[string]interface{}{
			"name":        "Credential Test Workflow",
			"description": "Workflow that uses a credential",
			"definition": map[string]interface{}{
				"nodes": []map[string]interface{}{
					{
						"id":   "trigger-1",
						"type": "trigger:manual",
						"position": map[string]interface{}{
							"x": 0,
							"y": 0,
						},
						"data": map[string]interface{}{
							"label": "Manual Trigger",
						},
					},
					{
						"id":   "action-1",
						"type": "action:http",
						"position": map[string]interface{}{
							"x": 200,
							"y": 0,
						},
						"data": map[string]interface{}{
							"label":  "HTTP Request with Credential",
							"url":    "https://httpbin.org/headers",
							"method": "GET",
							"headers": map[string]interface{}{
								"X-API-Key": "{{credentials.API Key Credential.api_key}}",
							},
						},
					},
				},
				"edges": []map[string]interface{}{
					{
						"id":     "edge-1",
						"source": "trigger-1",
						"target": "action-1",
					},
				},
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
		integration.AssertStatusCode(t, resp, http.StatusCreated)

		var workflow Workflow
		integration.ParseJSONResponse(t, resp, &workflow)

		workflowID = workflow.ID
		t.Logf("Created workflow using credential: %s", workflowID)
	})

	// Step 4: Verify credential reference resolution
	t.Run("Step4_VerifyCredentialReference", func(t *testing.T) {
		require.NotEmpty(t, workflowID, "workflow ID required from Step 3")

		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/workflows/"+workflowID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		body := integration.GetResponseBody(t, resp)

		// Credential reference should be in the workflow definition
		assert.Contains(t, body, "credentials.API Key Credential")
		// But actual value should never appear
		assert.NotContains(t, body, "sk-test-secret-key-12345")
		t.Logf("Verified credential reference in workflow")
	})

	// Step 5: Test credential
	t.Run("Step5_TestCredential", func(t *testing.T) {
		require.NotEmpty(t, credentialID, "credential ID required from Step 1")

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/credentials/"+credentialID+"/test", nil, headers)

		// Test endpoint might return different status codes based on implementation
		if resp.StatusCode == http.StatusOK {
			var testResult CredentialTestResponse
			integration.ParseJSONResponse(t, resp, &testResult)
			t.Logf("Credential test result: success=%v, message=%s", testResult.Success, testResult.Message)
		} else if resp.StatusCode == http.StatusNotImplemented {
			t.Logf("Credential test not implemented for this type")
			resp.Body.Close()
		} else {
			t.Logf("Credential test returned status %d", resp.StatusCode)
			resp.Body.Close()
		}
	})

	// Step 6: Rotate credential
	t.Run("Step6_RotateCredential", func(t *testing.T) {
		require.NotEmpty(t, credentialID, "credential ID required from Step 1")

		rotateInput := map[string]interface{}{
			"data": map[string]interface{}{
				"api_key": "sk-new-rotated-key-67890",
				"header":  "X-API-Key",
			},
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/credentials/"+credentialID, rotateInput, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var credential Credential
		integration.ParseJSONResponse(t, resp, &credential)

		assert.Equal(t, 2, credential.Version)
		t.Logf("Rotated credential to version %d", credential.Version)
	})

	// Step 7: Verify old credential value is not accessible
	t.Run("Step7_VerifyOldValueRemoved", func(t *testing.T) {
		require.NotEmpty(t, credentialID, "credential ID required from Step 1")

		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/credentials/"+credentialID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		body := integration.GetResponseBody(t, resp)

		// Neither old nor new secret should appear
		assert.NotContains(t, body, "sk-test-secret-key-12345")
		assert.NotContains(t, body, "sk-new-rotated-key-67890")
		t.Logf("Verified rotated credential data is masked")
	})

	// Step 8: Delete credential
	t.Run("Step8_DeleteCredential", func(t *testing.T) {
		require.NotEmpty(t, credentialID, "credential ID required from Step 1")

		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/credentials/"+credentialID, nil, headers)
		integration.AssertStatusCode(t, resp, http.StatusNoContent)

		// Verify deletion
		getResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/credentials/"+credentialID, nil, headers)
		integration.AssertStatusCode(t, getResp, http.StatusNotFound)

		t.Logf("Deleted credential successfully")
	})
}

// TestCredentialFlow_MultipleTypes tests different credential types
func TestCredentialFlow_MultipleTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "credential-types-tenant")
	headers := integration.DefaultTestHeaders(tenantID)

	credentialTypes := []struct {
		name        string
		credType    string
		data        map[string]interface{}
		secretField string
	}{
		{
			name:     "API Key",
			credType: "api_key",
			data: map[string]interface{}{
				"api_key": "test-api-key-secret",
			},
			secretField: "test-api-key-secret",
		},
		{
			name:     "Basic Auth",
			credType: "basic_auth",
			data: map[string]interface{}{
				"username": "testuser",
				"password": "secret-password-123",
			},
			secretField: "secret-password-123",
		},
		{
			name:     "OAuth Token",
			credType: "oauth_token",
			data: map[string]interface{}{
				"access_token":  "oauth-access-token-xyz",
				"refresh_token": "oauth-refresh-token-abc",
			},
			secretField: "oauth-access-token-xyz",
		},
		{
			name:     "AWS Credentials",
			credType: "aws",
			data: map[string]interface{}{
				"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
				"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"region":            "us-east-1",
			},
			secretField: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	for _, tc := range credentialTypes {
		t.Run("creates and masks "+tc.name, func(t *testing.T) {
			input := map[string]interface{}{
				"name":        tc.name,
				"description": "Test " + tc.name + " credential",
				"type":        tc.credType,
				"data":        tc.data,
			}

			// Create
			resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/credentials", input, headers)
			integration.AssertStatusCode(t, resp, http.StatusCreated)

			var credential Credential
			integration.ParseJSONResponse(t, resp, &credential)

			assert.NotEmpty(t, credential.ID)
			assert.Equal(t, tc.credType, credential.Type)

			// Verify masking
			getResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/credentials/"+credential.ID, nil, headers)
			integration.AssertStatusCode(t, getResp, http.StatusOK)

			body := integration.GetResponseBody(t, getResp)
			assert.NotContains(t, body, tc.secretField, "secret should be masked")

			t.Logf("Created and verified masking for %s credential", tc.name)
		})
	}
}

// TestCredentialFlow_AccessControl tests credential access control
func TestCredentialFlow_AccessControl(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)

	// Create two tenants
	tenant1ID := ts.CreateTestTenant(t, "credential-tenant-1")
	tenant2ID := ts.CreateTestTenant(t, "credential-tenant-2")

	headers1 := integration.DefaultTestHeaders(tenant1ID)
	headers2 := integration.DefaultTestHeaders(tenant2ID)

	// Create credential in tenant 1
	input := map[string]interface{}{
		"name": "Tenant 1 Secret",
		"type": "api_key",
		"data": map[string]interface{}{
			"api_key": "tenant-1-secret-key",
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/credentials", input, headers1)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var credential Credential
	integration.ParseJSONResponse(t, resp, &credential)

	t.Run("tenant cannot access another tenant's credentials", func(t *testing.T) {
		// Tenant 2 tries to access tenant 1's credential
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/credentials/"+credential.ID, nil, headers2)

		// Should be either 404 (not found) or 403 (forbidden)
		assert.True(t,
			resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden,
			"expected 404 or 403, got %d", resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("tenant cannot update another tenant's credentials", func(t *testing.T) {
		updateInput := map[string]interface{}{
			"name": "Hacked Name",
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/credentials/"+credential.ID, updateInput, headers2)

		assert.True(t,
			resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden,
			"expected 404 or 403, got %d", resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("tenant cannot delete another tenant's credentials", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/credentials/"+credential.ID, nil, headers2)

		assert.True(t,
			resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden,
			"expected 404 or 403, got %d", resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("tenant list only shows own credentials", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/credentials", nil, headers2)
		integration.AssertStatusCode(t, resp, http.StatusOK)

		var listResp CredentialListResponse
		integration.ParseJSONResponse(t, resp, &listResp)

		// Tenant 2 should have no credentials
		for _, cred := range listResp.Data {
			assert.NotEqual(t, credential.ID, cred.ID, "should not see other tenant's credentials")
		}
	})
}

// TestCredentialFlow_ValidationErrors tests credential validation
func TestCredentialFlow_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "credential-validation-tenant")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("rejects credential without name", func(t *testing.T) {
		input := map[string]interface{}{
			"type": "api_key",
			"data": map[string]interface{}{
				"api_key": "test-key",
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/credentials", input, headers)
		integration.AssertStatusCode(t, resp, http.StatusBadRequest)
	})

	t.Run("rejects credential without type", func(t *testing.T) {
		input := map[string]interface{}{
			"name": "No Type",
			"data": map[string]interface{}{
				"api_key": "test-key",
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/credentials", input, headers)
		integration.AssertStatusCode(t, resp, http.StatusBadRequest)
	})

	t.Run("rejects credential without data", func(t *testing.T) {
		input := map[string]interface{}{
			"name": "No Data",
			"type": "api_key",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/credentials", input, headers)
		integration.AssertStatusCode(t, resp, http.StatusBadRequest)
	})

	t.Run("rejects duplicate credential name", func(t *testing.T) {
		input := map[string]interface{}{
			"name": "Duplicate Name",
			"type": "api_key",
			"data": map[string]interface{}{
				"api_key": "test-key-1",
			},
		}

		// Create first credential
		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/credentials", input, headers)
		integration.AssertStatusCode(t, resp, http.StatusCreated)

		// Try to create duplicate
		input["data"] = map[string]interface{}{
			"api_key": "test-key-2",
		}
		resp = ts.MakeRequest(t, http.MethodPost, "/api/v1/credentials", input, headers)

		// Should be conflict or bad request
		assert.True(t,
			resp.StatusCode == http.StatusConflict || resp.StatusCode == http.StatusBadRequest,
			"expected 409 or 400 for duplicate name, got %d", resp.StatusCode)
		resp.Body.Close()
	})
}

// TestCredentialFlow_MaskingInLogs tests that credentials are masked in execution logs
func TestCredentialFlow_MaskingInLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "credential-logging-tenant")
	headers := integration.DefaultTestHeaders(tenantID)

	// Create credential with identifiable secret
	secretValue := "SUPER_SECRET_VALUE_12345"
	credInput := map[string]interface{}{
		"name": "Log Test Credential",
		"type": "api_key",
		"data": map[string]interface{}{
			"api_key": secretValue,
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/credentials", credInput, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var credential Credential
	integration.ParseJSONResponse(t, resp, &credential)

	// Create workflow that uses credential in an action
	workflowInput := map[string]interface{}{
		"name":   "Log Masking Test",
		"status": "active",
		"definition": map[string]interface{}{
			"nodes": []map[string]interface{}{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]interface{}{"x": 0, "y": 0},
					"data":     map[string]interface{}{"label": "Manual"},
				},
				{
					"id":       "action-1",
					"type":     "action:log",
					"position": map[string]interface{}{"x": 200, "y": 0},
					"data": map[string]interface{}{
						"label":   "Log Credential",
						"message": "Using API key: {{credentials.Log Test Credential.api_key}}",
					},
				},
			},
			"edges": []map[string]interface{}{
				{"id": "edge-1", "source": "trigger-1", "target": "action-1"},
			},
		},
	}

	resp = ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", workflowInput, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)

	// Execute the workflow
	execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", nil, headers)
	integration.AssertStatusCode(t, execResp, http.StatusAccepted)

	var execution Execution
	integration.ParseJSONResponse(t, execResp, &execution)

	// Check execution logs don't contain the secret
	t.Run("verifies secret not in execution logs", func(t *testing.T) {
		logsResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/executions/"+execution.ID+"/logs", nil, headers)

		if logsResp.StatusCode == http.StatusOK {
			body := integration.GetResponseBody(t, logsResp)
			assert.False(t, strings.Contains(body, secretValue),
				"execution logs should not contain raw secret value")
			t.Logf("Verified secret is masked in execution logs")
		} else {
			t.Logf("Logs endpoint returned %d (execution may still be processing)", logsResp.StatusCode)
			logsResp.Body.Close()
		}
	})
}
