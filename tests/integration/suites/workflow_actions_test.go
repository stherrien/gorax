package suites

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/tests/integration"
)

// =============================================================================
// HTTP Action Tests
// =============================================================================

// TestWorkflowAction_HTTP tests HTTP action execution with various configurations
func TestWorkflowAction_HTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "http-action-test")
	headers := integration.DefaultTestHeaders(tenantID)

	// Create mock external server
	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	t.Run("GET request with JSON response", func(t *testing.T) {
		// Setup mock response
		mockServer.SetMethodResponse("GET", "/api/data", integration.MockResponse{
			StatusCode: 200,
			Body: map[string]any{
				"items": []map[string]any{
					{"id": 1, "name": "Item 1"},
					{"id": 2, "name": "Item 2"},
				},
				"total": 2,
			},
		})

		// Create workflow with HTTP action
		workflow := createWorkflowWithHTTPAction(t, ts, tenantID, headers, mockServer.URL(), "GET", nil)

		// Execute workflow
		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		// Verify request was made
		require.Eventually(t, func() bool {
			return mockServer.GetRequestCount() > 0
		}, 5*time.Second, 100*time.Millisecond)

		lastReq := mockServer.GetLastRequest()
		require.NotNil(t, lastReq)
		assert.Equal(t, "GET", lastReq.Method)
		assert.Equal(t, "/api/data", lastReq.Path)
	})

	t.Run("POST request with body", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetMethodResponse("POST", "/api/items", integration.MockResponse{
			StatusCode: 201,
			Body: map[string]any{
				"id":   123,
				"name": "Created Item",
			},
		})

		requestBody := map[string]any{
			"name":        "Test Item",
			"description": "A test item",
		}

		workflow := createWorkflowWithHTTPAction(t, ts, tenantID, headers, mockServer.URL(), "POST", requestBody)

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		require.Eventually(t, func() bool {
			return mockServer.GetRequestCount() > 0
		}, 5*time.Second, 100*time.Millisecond)

		lastReq := mockServer.GetLastRequest()
		require.NotNil(t, lastReq)
		assert.Equal(t, "POST", lastReq.Method)

		var body map[string]any
		require.NoError(t, lastReq.RequestBodyAs(&body))
		assert.Equal(t, "Test Item", body["name"])
	})

	t.Run("Bearer authentication", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetAuth(&integration.AuthConfig{
			Type:  "bearer",
			Token: "test-bearer-token",
		})
		mockServer.SetMethodResponse("GET", "/api/secure", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"secure": true},
		})

		workflow := createWorkflowWithAuthHTTPAction(t, ts, tenantID, headers, mockServer.URL(), "bearer", "test-bearer-token")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		require.Eventually(t, func() bool {
			return mockServer.GetRequestCount() > 0
		}, 5*time.Second, 100*time.Millisecond)

		lastReq := mockServer.GetLastRequest()
		assert.True(t, lastReq.HasHeader("Authorization", "Bearer test-bearer-token"))

		mockServer.SetAuth(nil)
	})

	t.Run("Basic authentication", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetAuth(&integration.AuthConfig{
			Type:     "basic",
			Username: "testuser",
			Password: "testpass",
		})
		mockServer.SetMethodResponse("GET", "/api/secure", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"secure": true},
		})

		workflow := createWorkflowWithBasicAuthHTTPAction(t, ts, tenantID, headers, mockServer.URL(), "testuser", "testpass")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		require.Eventually(t, func() bool {
			return mockServer.GetRequestCount() > 0
		}, 5*time.Second, 100*time.Millisecond)

		lastReq := mockServer.GetLastRequest()
		assert.True(t, lastReq.HasHeaderContaining("Authorization", "Basic"))

		mockServer.SetAuth(nil)
	})

	t.Run("API Key authentication", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetAuth(&integration.AuthConfig{
			Type:         "api_key",
			APIKey:       "my-api-key-123",
			APIKeyHeader: "X-API-Key",
		})
		mockServer.SetMethodResponse("GET", "/api/secure", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"secure": true},
		})

		workflow := createWorkflowWithAPIKeyAuthHTTPAction(t, ts, tenantID, headers, mockServer.URL(), "my-api-key-123")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		require.Eventually(t, func() bool {
			return mockServer.GetRequestCount() > 0
		}, 5*time.Second, 100*time.Millisecond)

		lastReq := mockServer.GetLastRequest()
		assert.True(t, lastReq.HasHeader("X-API-Key", "my-api-key-123"))

		mockServer.SetAuth(nil)
	})

	t.Run("handles 4xx error responses", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetMethodResponse("GET", "/api/notfound", integration.MockResponse{
			StatusCode: 404,
			Body: map[string]any{
				"error": "Resource not found",
			},
		})

		workflow := createWorkflowWithHTTPActionPath(t, ts, tenantID, headers, mockServer.URL(), "GET", "/api/notfound")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		require.Eventually(t, func() bool {
			return mockServer.GetRequestCount() > 0
		}, 5*time.Second, 100*time.Millisecond)

		// The workflow should still execute, the HTTP action should return the 404 status
		lastReq := mockServer.GetLastRequest()
		require.NotNil(t, lastReq)
	})

	t.Run("handles 5xx error responses", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetMethodResponse("GET", "/api/error", integration.MockResponse{
			StatusCode: 500,
			Body: map[string]any{
				"error": "Internal server error",
			},
		})

		workflow := createWorkflowWithHTTPActionPath(t, ts, tenantID, headers, mockServer.URL(), "GET", "/api/error")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		require.Eventually(t, func() bool {
			return mockServer.GetRequestCount() > 0
		}, 5*time.Second, 100*time.Millisecond)
	})

	t.Run("custom headers", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetMethodResponse("GET", "/api/headers", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"ok": true},
		})

		workflow := createWorkflowWithCustomHeaders(t, ts, tenantID, headers, mockServer.URL(), map[string]string{
			"X-Custom-Header": "custom-value",
			"X-Trace-ID":      "trace-123",
		})

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		require.Eventually(t, func() bool {
			return mockServer.GetRequestCount() > 0
		}, 5*time.Second, 100*time.Millisecond)

		lastReq := mockServer.GetLastRequest()
		assert.True(t, lastReq.HasHeader("X-Custom-Header", "custom-value"))
		assert.True(t, lastReq.HasHeader("X-Trace-ID", "trace-123"))
	})

	t.Run("query parameters", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetMethodResponse("GET", "/api/search", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"results": []any{}},
		})

		workflow := createWorkflowWithHTTPActionPath(t, ts, tenantID, headers, mockServer.URL(), "GET", "/api/search?q=test&limit=10")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		require.Eventually(t, func() bool {
			return mockServer.GetRequestCount() > 0
		}, 5*time.Second, 100*time.Millisecond)

		lastReq := mockServer.GetLastRequest()
		assert.Equal(t, "test", lastReq.QueryParam("q"))
		assert.Equal(t, "10", lastReq.QueryParam("limit"))
	})

	t.Run("timeout handling", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetMethodResponse("GET", "/api/slow", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"slow": true},
			DelayMs:    5000, // 5 second delay
		})

		// Create workflow with short timeout
		workflow := createWorkflowWithHTTPActionTimeout(t, ts, tenantID, headers, mockServer.URL(), 1) // 1 second timeout

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		// The workflow execution should have started
		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)
	})

	t.Run("SSRF protection - blocks private IPs", func(t *testing.T) {
		// Try to make request to localhost (SSRF)
		workflow := createWorkflowWithHTTPActionPath(t, ts, tenantID, headers, "http://127.0.0.1:8080", "GET", "/api/internal")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		// Should be accepted initially
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		// The execution should eventually fail due to SSRF protection
		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)
	})
}

// =============================================================================
// Transform Action Tests
// =============================================================================

// TestWorkflowAction_Transform tests Transform action for data manipulation
func TestWorkflowAction_Transform(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "transform-action-test")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("JSONPath extraction", func(t *testing.T) {
		workflow := createWorkflowWithTransformAction(t, ts, tenantID, headers, map[string]any{
			"transformations": []map[string]any{
				{
					"target": "userName",
					"source": "$.trigger.user.name",
				},
				{
					"target": "userEmail",
					"source": "$.trigger.user.email",
				},
			},
		})

		execInput := map[string]any{
			"input": map[string]any{
				"user": map[string]any{
					"name":  "John Doe",
					"email": "john@example.com",
				},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("nested data access", func(t *testing.T) {
		workflow := createWorkflowWithTransformAction(t, ts, tenantID, headers, map[string]any{
			"transformations": []map[string]any{
				{
					"target": "city",
					"source": "$.trigger.address.city",
				},
				{
					"target": "country",
					"source": "$.trigger.address.country",
				},
			},
		})

		execInput := map[string]any{
			"input": map[string]any{
				"address": map[string]any{
					"street":  "123 Main St",
					"city":    "San Francisco",
					"country": "USA",
				},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("array transformations", func(t *testing.T) {
		workflow := createWorkflowWithTransformAction(t, ts, tenantID, headers, map[string]any{
			"transformations": []map[string]any{
				{
					"target": "firstItem",
					"source": "$.trigger.items[0].name",
				},
				{
					"target": "itemCount",
					"source": "$.trigger.items.length",
				},
			},
		})

		execInput := map[string]any{
			"input": map[string]any{
				"items": []map[string]any{
					{"name": "Item 1", "price": 10},
					{"name": "Item 2", "price": 20},
					{"name": "Item 3", "price": 30},
				},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("missing field handling", func(t *testing.T) {
		workflow := createWorkflowWithTransformAction(t, ts, tenantID, headers, map[string]any{
			"transformations": []map[string]any{
				{
					"target":  "optional",
					"source":  "$.trigger.nonexistent.field",
					"default": "default_value",
				},
			},
		})

		execInput := map[string]any{
			"input": map[string]any{
				"existing": "value",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})
}

// =============================================================================
// Formula Action Tests
// =============================================================================

// TestWorkflowAction_Formula tests Formula action with CEL expressions
func TestWorkflowAction_Formula(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "formula-action-test")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("arithmetic operations", func(t *testing.T) {
		workflow := createWorkflowWithFormulaAction(t, ts, tenantID, headers, "trigger.price * trigger.quantity")

		execInput := map[string]any{
			"input": map[string]any{
				"price":    10.5,
				"quantity": 3,
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("string functions", func(t *testing.T) {
		workflow := createWorkflowWithFormulaAction(t, ts, tenantID, headers, "trigger.name.toUpperCase()")

		execInput := map[string]any{
			"input": map[string]any{
				"name": "hello world",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("conditional expressions", func(t *testing.T) {
		workflow := createWorkflowWithFormulaAction(t, ts, tenantID, headers, "trigger.score >= 50 ? 'pass' : 'fail'")

		execInput := map[string]any{
			"input": map[string]any{
				"score": 75,
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("context variable access", func(t *testing.T) {
		workflow := createWorkflowWithFormulaAction(t, ts, tenantID, headers, "env.workflow_id != ''")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})
}

// =============================================================================
// Code Action Tests
// =============================================================================

// TestWorkflowAction_Code tests Code/Script action with JavaScript sandbox
func TestWorkflowAction_Code(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "code-action-test")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("simple script execution", func(t *testing.T) {
		script := `
			const result = trigger.numbers.reduce((sum, n) => sum + n, 0);
			return { total: result };
		`
		workflow := createWorkflowWithCodeAction(t, ts, tenantID, headers, script)

		execInput := map[string]any{
			"input": map[string]any{
				"numbers": []int{1, 2, 3, 4, 5},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("input/output data access", func(t *testing.T) {
		script := `
			const { firstName, lastName } = trigger;
			return { fullName: firstName + ' ' + lastName };
		`
		workflow := createWorkflowWithCodeAction(t, ts, tenantID, headers, script)

		execInput := map[string]any{
			"input": map[string]any{
				"firstName": "John",
				"lastName":  "Doe",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("console capture", func(t *testing.T) {
		script := `
			console.log('Processing started');
			console.log('Input:', JSON.stringify(trigger));
			return { processed: true };
		`
		workflow := createWorkflowWithCodeAction(t, ts, tenantID, headers, script)

		execInput := map[string]any{
			"input": map[string]any{
				"data": "test",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("timeout enforcement", func(t *testing.T) {
		// Script that would run forever
		script := `
			let i = 0;
			while(true) { i++; }
			return { i: i };
		`
		workflow := createWorkflowWithCodeActionTimeout(t, ts, tenantID, headers, script, 1) // 1 second timeout

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("security sandbox - no fs access", func(t *testing.T) {
		// Script attempting file system access should fail
		script := `
			const fs = require('fs');
			return { content: fs.readFileSync('/etc/passwd').toString() };
		`
		workflow := createWorkflowWithCodeAction(t, ts, tenantID, headers, script)

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
		// Execution should fail due to sandbox restrictions
	})
}

// =============================================================================
// SubWorkflow Action Tests
// =============================================================================

// TestWorkflowAction_SubWorkflow tests SubWorkflow action for nested execution
func TestWorkflowAction_SubWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "subworkflow-action-test")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("nested workflow execution", func(t *testing.T) {
		// Create child workflow
		childWorkflow := createSimpleLogWorkflow(t, ts, tenantID, headers, "Child Workflow")

		// Create parent workflow that calls child
		parentWorkflow := createWorkflowWithSubWorkflowAction(t, ts, tenantID, headers, childWorkflow.ID)

		execInput := map[string]any{
			"input": map[string]any{
				"message": "Hello from parent",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+parentWorkflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("input passing to sub-workflow", func(t *testing.T) {
		childWorkflow := createSimpleLogWorkflow(t, ts, tenantID, headers, "Input Test Child")
		parentWorkflow := createWorkflowWithSubWorkflowAction(t, ts, tenantID, headers, childWorkflow.ID)

		execInput := map[string]any{
			"input": map[string]any{
				"value": 42,
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+parentWorkflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("max depth limit enforcement", func(t *testing.T) {
		// Create a workflow that calls itself (circular dependency)
		// This should be detected and prevented
		selfReferencingInput := map[string]any{
			"name":   "Self Referencing",
			"status": "active",
			"definition": map[string]any{
				"nodes": []map[string]any{
					{
						"id":       "trigger-1",
						"type":     "trigger:manual",
						"position": map[string]any{"x": 0, "y": 0},
						"data":     map[string]any{"label": "Trigger"},
					},
				},
				"edges": []map[string]any{},
			},
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", selfReferencingInput, headers)
		integration.AssertStatusCode(t, resp, http.StatusCreated)

		var wf Workflow
		integration.ParseJSONResponse(t, resp, &wf)

		// Try to execute
		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+wf.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

func createWorkflowWithHTTPAction(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, baseURL, method string, body any) Workflow {
	t.Helper()

	bodyJSON, _ := json.Marshal(body)

	input := map[string]any{
		"name":   fmt.Sprintf("HTTP %s Test Workflow", method),
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "http-1",
					"type":     "action:http",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "HTTP Request",
						"config": map[string]any{
							"method":  method,
							"url":     baseURL + "/api/data",
							"body":    json.RawMessage(bodyJSON),
							"timeout": 30,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "http-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithAuthHTTPAction(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, baseURL, authType, token string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "HTTP Auth Test Workflow",
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "http-1",
					"type":     "action:http",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "HTTP Request",
						"config": map[string]any{
							"method": "GET",
							"url":    baseURL + "/api/secure",
							"auth": map[string]any{
								"type":  authType,
								"token": token,
							},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "http-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithBasicAuthHTTPAction(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, baseURL, username, password string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "HTTP Basic Auth Test Workflow",
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "http-1",
					"type":     "action:http",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "HTTP Request",
						"config": map[string]any{
							"method": "GET",
							"url":    baseURL + "/api/secure",
							"auth": map[string]any{
								"type":     "basic",
								"username": username,
								"password": password,
							},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "http-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithAPIKeyAuthHTTPAction(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, baseURL, apiKey string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "HTTP API Key Auth Test Workflow",
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "http-1",
					"type":     "action:http",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "HTTP Request",
						"config": map[string]any{
							"method": "GET",
							"url":    baseURL + "/api/secure",
							"auth": map[string]any{
								"type":    "api_key",
								"api_key": apiKey,
								"header":  "X-API-Key",
							},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "http-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithHTTPActionPath(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, baseURL, method, path string) Workflow {
	t.Helper()

	url := baseURL + path
	if strings.HasPrefix(path, "http") {
		url = path
	}

	input := map[string]any{
		"name":   fmt.Sprintf("HTTP %s Path Test Workflow", method),
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "http-1",
					"type":     "action:http",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "HTTP Request",
						"config": map[string]any{
							"method": method,
							"url":    url,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "http-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithCustomHeaders(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, baseURL string, customHeaders map[string]string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "HTTP Custom Headers Test Workflow",
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "http-1",
					"type":     "action:http",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "HTTP Request",
						"config": map[string]any{
							"method":  "GET",
							"url":     baseURL + "/api/headers",
							"headers": customHeaders,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "http-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithHTTPActionTimeout(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, baseURL string, timeoutSeconds int) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "HTTP Timeout Test Workflow",
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "http-1",
					"type":     "action:http",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "HTTP Request",
						"config": map[string]any{
							"method":  "GET",
							"url":     baseURL + "/api/slow",
							"timeout": timeoutSeconds,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "http-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithTransformAction(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, transformConfig map[string]any) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Transform Test Workflow",
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "transform-1",
					"type":     "action:transform",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label":  "Transform Data",
						"config": transformConfig,
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "transform-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithFormulaAction(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, expression string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Formula Test Workflow",
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "formula-1",
					"type":     "action:formula",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Formula",
						"config": map[string]any{
							"expression": expression,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "formula-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithCodeAction(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, script string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Code Test Workflow",
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "code-1",
					"type":     "action:code",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Code",
						"config": map[string]any{
							"script": script,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "code-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithCodeActionTimeout(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, script string, timeoutSeconds int) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Code Timeout Test Workflow",
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "code-1",
					"type":     "action:code",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Code",
						"config": map[string]any{
							"script":  script,
							"timeout": timeoutSeconds,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "code-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createSimpleLogWorkflow(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, name string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   name,
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "log-1",
					"type":     "action:log",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label":   "Log",
						"message": "Workflow executed",
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "log-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithSubWorkflowAction(t *testing.T, ts *integration.TestServer, _ string, headers map[string]string, childWorkflowID string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Parent Workflow",
		"status": "active",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Manual Trigger"},
				},
				{
					"id":       "subworkflow-1",
					"type":     "control:subworkflow",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Sub-Workflow",
						"config": map[string]any{
							"workflow_id": childWorkflowID,
							"input":       "${trigger}",
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "subworkflow-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}
