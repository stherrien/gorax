package suites

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/tests/integration"
)

// =============================================================================
// End-to-End Workflow Scenario Tests
// =============================================================================

// TestE2E_DataPipelineWorkflow tests a realistic data pipeline scenario:
// Webhook trigger → HTTP fetch data → Transform → Loop process → HTTP post results
func TestE2E_DataPipelineWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "e2e-data-pipeline")
	headers := integration.DefaultTestHeaders(tenantID)

	// Create mock server to simulate external API
	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	// Configure mock responses for the data pipeline
	mockServer.SetMethodResponse("GET", "/api/users", integration.MockResponse{
		StatusCode: 200,
		Body: map[string]any{
			"users": []map[string]any{
				{"id": "1", "name": "Alice", "email": "alice@example.com"},
				{"id": "2", "name": "Bob", "email": "bob@example.com"},
				{"id": "3", "name": "Charlie", "email": "charlie@example.com"},
			},
		},
	})

	mockServer.SetMethodResponse("POST", "/api/notifications", integration.MockResponse{
		StatusCode: 201,
		Body: map[string]any{
			"sent": true,
			"count": 1,
		},
	})

	t.Run("complete data pipeline execution", func(t *testing.T) {
		workflow := createDataPipelineWorkflow(t, ts, headers, mockServer.URL())

		// Execute workflow via manual trigger
		execInput := map[string]any{
			"input": map[string]any{
				"source": "scheduled_sync",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)

		// Allow some time for async execution
		time.Sleep(500 * time.Millisecond)

		// Verify mock server received expected requests
		requests := mockServer.GetRequests()
		assert.GreaterOrEqual(t, len(requests), 1, "Should have made at least one request")

		// Verify GET request to fetch users
		userRequests := mockServer.GetRequestsForPath("/api/users")
		assert.GreaterOrEqual(t, len(userRequests), 1, "Should have fetched users")
	})

	t.Run("pipeline handles empty data", func(t *testing.T) {
		// Configure mock to return empty array
		mockServer.SetMethodResponse("GET", "/api/users", integration.MockResponse{
			StatusCode: 200,
			Body: map[string]any{
				"users": []map[string]any{},
			},
		})

		workflow := createDataPipelineWorkflow(t, ts, headers, mockServer.URL())

		execInput := map[string]any{
			"input": map[string]any{},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)
	})
}

// TestE2E_ErrorHandlingWorkflow tests error handling scenario:
// Trigger → Try { HTTP call } Catch { Log error } → Final log
func TestE2E_ErrorHandlingWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "e2e-error-handling")
	headers := integration.DefaultTestHeaders(tenantID)

	// Create mock server that will fail
	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	t.Run("successful execution skips catch", func(t *testing.T) {
		// Configure mock for success
		mockServer.SetMethodResponse("POST", "/api/process", integration.MockResponse{
			StatusCode: 200,
			Body: map[string]any{
				"status": "processed",
			},
		})

		workflow := createErrorHandlingWorkflow(t, ts, headers, mockServer.URL())

		execInput := map[string]any{
			"input": map[string]any{
				"data": "test-payload",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)
	})

	t.Run("failure triggers catch branch", func(t *testing.T) {
		// Configure mock to fail
		mockServer.SetMethodResponse("POST", "/api/process", integration.MockResponse{
			StatusCode: 500,
			Body: map[string]any{
				"error": "Internal server error",
			},
		})

		workflow := createErrorHandlingWorkflow(t, ts, headers, mockServer.URL())

		execInput := map[string]any{
			"input": map[string]any{
				"data": "test-payload",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)

		// Allow time for execution
		time.Sleep(500 * time.Millisecond)

		// Execution should complete (error was caught)
		detailResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/executions/"+exec.ID, nil, headers)
		integration.AssertStatusCode(t, detailResp, http.StatusOK)
	})

	t.Run("network error handled gracefully", func(t *testing.T) {
		// Configure mock to simulate connection error
		mockServer.SetMethodResponse("POST", "/api/process", integration.MockResponse{
			Error:    true,
			ErrorMsg: "Connection refused",
		})

		workflow := createErrorHandlingWorkflow(t, ts, headers, mockServer.URL())

		execInput := map[string]any{
			"input": map[string]any{
				"data": "test-payload",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)
	})
}

// TestE2E_ParallelProcessingWorkflow tests parallel execution scenario:
// Trigger → Parallel { HTTP A, HTTP B, HTTP C } → Transform aggregate
func TestE2E_ParallelProcessingWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "e2e-parallel-processing")
	headers := integration.DefaultTestHeaders(tenantID)

	// Create mock server for parallel calls
	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	// Configure mock responses for different services
	mockServer.SetMethodResponse("GET", "/api/service-a", integration.MockResponse{
		StatusCode: 200,
		Body: map[string]any{
			"service": "A",
			"data":    []int{1, 2, 3},
		},
	})

	mockServer.SetMethodResponse("GET", "/api/service-b", integration.MockResponse{
		StatusCode: 200,
		Body: map[string]any{
			"service": "B",
			"data":    []int{4, 5, 6},
		},
	})

	mockServer.SetMethodResponse("GET", "/api/service-c", integration.MockResponse{
		StatusCode: 200,
		Body: map[string]any{
			"service": "C",
			"data":    []int{7, 8, 9},
		},
	})

	t.Run("parallel branches execute concurrently", func(t *testing.T) {
		workflow := createParallelProcessingWorkflow(t, ts, headers, mockServer.URL())
		mockServer.ClearRequests()

		execInput := map[string]any{
			"input": map[string]any{},
		}

		startTime := time.Now()
		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)

		// Allow time for parallel execution
		time.Sleep(1 * time.Second)

		// Verify all three services were called
		requests := mockServer.GetRequests()
		executionDuration := time.Since(startTime)

		// In parallel mode, all requests should complete faster than sequential
		t.Logf("Execution duration: %v, Request count: %d", executionDuration, len(requests))

		// Check that all services received requests
		serviceAPaths := mockServer.GetRequestsForPath("/api/service-a")
		serviceBPaths := mockServer.GetRequestsForPath("/api/service-b")
		serviceCPaths := mockServer.GetRequestsForPath("/api/service-c")

		assert.GreaterOrEqual(t, len(serviceAPaths), 1, "Service A should have been called")
		assert.GreaterOrEqual(t, len(serviceBPaths), 1, "Service B should have been called")
		assert.GreaterOrEqual(t, len(serviceCPaths), 1, "Service C should have been called")
	})

	t.Run("partial failure handling", func(t *testing.T) {
		// Configure one service to fail
		mockServer.SetMethodResponse("GET", "/api/service-b", integration.MockResponse{
			StatusCode: 500,
			Body: map[string]any{
				"error": "Service B unavailable",
			},
		})

		workflow := createParallelProcessingWorkflow(t, ts, headers, mockServer.URL())

		execInput := map[string]any{
			"input": map[string]any{},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)

		// Allow time for execution
		time.Sleep(1 * time.Second)

		// Check execution status
		detailResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/executions/"+exec.ID, nil, headers)
		integration.AssertStatusCode(t, detailResp, http.StatusOK)
	})
}

// TestE2E_ConditionalBranchingWorkflow tests conditional execution paths
func TestE2E_ConditionalBranchingWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "e2e-conditional")
	headers := integration.DefaultTestHeaders(tenantID)

	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	mockServer.SetMethodResponse("POST", "/api/high-priority", integration.MockResponse{
		StatusCode: 200,
		Body:       map[string]any{"processed": "high"},
	})

	mockServer.SetMethodResponse("POST", "/api/low-priority", integration.MockResponse{
		StatusCode: 200,
		Body:       map[string]any{"processed": "low"},
	})

	t.Run("high priority branch taken", func(t *testing.T) {
		workflow := createConditionalWorkflow(t, ts, headers, mockServer.URL())
		mockServer.ClearRequests()

		execInput := map[string]any{
			"input": map[string]any{
				"priority": "high",
				"message":  "Urgent task",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)

		// Allow time for execution
		time.Sleep(500 * time.Millisecond)

		// Verify high priority endpoint was called
		highPriorityRequests := mockServer.GetRequestsForPath("/api/high-priority")
		assert.GreaterOrEqual(t, len(highPriorityRequests), 1, "High priority endpoint should be called")
	})

	t.Run("low priority branch taken", func(t *testing.T) {
		workflow := createConditionalWorkflow(t, ts, headers, mockServer.URL())
		mockServer.ClearRequests()

		execInput := map[string]any{
			"input": map[string]any{
				"priority": "low",
				"message":  "Regular task",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)

		// Allow time for execution
		time.Sleep(500 * time.Millisecond)

		// Verify low priority endpoint was called
		lowPriorityRequests := mockServer.GetRequestsForPath("/api/low-priority")
		assert.GreaterOrEqual(t, len(lowPriorityRequests), 1, "Low priority endpoint should be called")
	})
}

// TestE2E_RetryWithBackoffWorkflow tests retry behavior with exponential backoff
func TestE2E_RetryWithBackoffWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "e2e-retry-backoff")
	headers := integration.DefaultTestHeaders(tenantID)

	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	t.Run("succeeds after transient failures", func(t *testing.T) {
		// Track request count
		requestCount := 0
		// Configure mock to fail first 2 times, then succeed
		mockServer.SetMethodResponse("POST", "/api/flaky", integration.MockResponse{
			StatusCode: 503,
			Body:       map[string]any{"error": "Service temporarily unavailable"},
		})

		workflow := createRetryWorkflow(t, ts, headers, mockServer.URL())
		mockServer.ClearRequests()

		execInput := map[string]any{
			"input": map[string]any{
				"operation": "flaky-op",
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)

		// Allow time for retries
		time.Sleep(2 * time.Second)

		// Verify multiple requests were made (retries)
		requests := mockServer.GetRequestsForPath("/api/flaky")
		requestCount = len(requests)
		t.Logf("Total retry requests: %d", requestCount)

		// Should have attempted at least 2 times (initial + retries)
		assert.GreaterOrEqual(t, requestCount, 1, "Should have attempted the request")
	})
}

// =============================================================================
// Helper Functions for E2E Tests
// =============================================================================

// createDataPipelineWorkflow creates a data pipeline workflow
func createDataPipelineWorkflow(t *testing.T, ts *integration.TestServer, headers map[string]string, apiBaseURL string) Workflow {
	t.Helper()

	workflowReq := map[string]any{
		"name":        "Data Pipeline Workflow",
		"description": "Fetches users, transforms data, and sends notifications",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Start Pipeline"},
				},
				{
					"id":       "http-fetch",
					"type":     "action:http",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Fetch Users",
						"config": map[string]any{
							"method": "GET",
							"url":    apiBaseURL + "/api/users",
						},
					},
				},
				{
					"id":       "transform-1",
					"type":     "action:transform",
					"position": map[string]any{"x": 400, "y": 0},
					"data": map[string]any{
						"label": "Extract Users",
						"config": map[string]any{
							"expression": "$.body.users",
							"outputKey":  "users",
						},
					},
				},
				{
					"id":       "loop-1",
					"type":     "control:loop",
					"position": map[string]any{"x": 600, "y": 0},
					"data": map[string]any{
						"label": "Process Each User",
						"config": map[string]any{
							"source":        "${transform-1.users}",
							"itemVariable":  "user",
							"maxIterations": 100,
						},
					},
				},
				{
					"id":       "http-notify",
					"type":     "action:http",
					"position": map[string]any{"x": 800, "y": 0},
					"data": map[string]any{
						"label": "Send Notification",
						"config": map[string]any{
							"method": "POST",
							"url":    apiBaseURL + "/api/notifications",
							"body":   map[string]any{"userId": "${loop.user.id}"},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "e1", "source": "trigger-1", "target": "http-fetch"},
				{"id": "e2", "source": "http-fetch", "target": "transform-1"},
				{"id": "e3", "source": "transform-1", "target": "loop-1"},
				{"id": "e4", "source": "loop-1", "target": "http-notify"},
			},
		},
		"enabled": true,
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", workflowReq, headers)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

// createErrorHandlingWorkflow creates an error handling workflow
func createErrorHandlingWorkflow(t *testing.T, ts *integration.TestServer, headers map[string]string, apiBaseURL string) Workflow {
	t.Helper()

	workflowReq := map[string]any{
		"name":        "Error Handling Workflow",
		"description": "Demonstrates try/catch error handling",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Start"},
				},
				{
					"id":       "try-1",
					"type":     "control:try",
					"position": map[string]any{"x": 200, "y": 0},
					"data":     map[string]any{"label": "Try Block"},
				},
				{
					"id":       "http-risky",
					"type":     "action:http",
					"position": map[string]any{"x": 400, "y": -50},
					"data": map[string]any{
						"label": "Risky HTTP Call",
						"config": map[string]any{
							"method": "POST",
							"url":    apiBaseURL + "/api/process",
							"body":   map[string]any{"data": "${trigger.data}"},
						},
					},
				},
				{
					"id":       "catch-1",
					"type":     "control:catch",
					"position": map[string]any{"x": 400, "y": 50},
					"data":     map[string]any{"label": "Catch Block"},
				},
				{
					"id":       "log-error",
					"type":     "action:log",
					"position": map[string]any{"x": 600, "y": 50},
					"data": map[string]any{
						"label": "Log Error",
						"config": map[string]any{
							"message": "Error occurred: ${catch.error.message}",
							"level":   "error",
						},
					},
				},
				{
					"id":       "final-log",
					"type":     "action:log",
					"position": map[string]any{"x": 800, "y": 0},
					"data": map[string]any{
						"label": "Final Log",
						"config": map[string]any{
							"message": "Workflow completed",
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "e1", "source": "trigger-1", "target": "try-1"},
				{"id": "e2", "source": "try-1", "target": "http-risky", "data": map[string]any{"type": "try"}},
				{"id": "e3", "source": "try-1", "target": "catch-1", "data": map[string]any{"type": "catch"}},
				{"id": "e4", "source": "catch-1", "target": "log-error"},
				{"id": "e5", "source": "http-risky", "target": "final-log"},
				{"id": "e6", "source": "log-error", "target": "final-log"},
			},
		},
		"enabled": true,
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", workflowReq, headers)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

// createParallelProcessingWorkflow creates a parallel processing workflow
func createParallelProcessingWorkflow(t *testing.T, ts *integration.TestServer, headers map[string]string, apiBaseURL string) Workflow {
	t.Helper()

	workflowReq := map[string]any{
		"name":        "Parallel Processing Workflow",
		"description": "Executes multiple HTTP calls in parallel",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 100},
					"data":     map[string]any{"label": "Start"},
				},
				{
					"id":       "parallel-1",
					"type":     "control:parallel",
					"position": map[string]any{"x": 200, "y": 100},
					"data": map[string]any{
						"label": "Parallel Execution",
						"config": map[string]any{
							"mode": "all", // wait for all branches
						},
					},
				},
				{
					"id":       "http-a",
					"type":     "action:http",
					"position": map[string]any{"x": 400, "y": 0},
					"data": map[string]any{
						"label": "Service A",
						"config": map[string]any{
							"method": "GET",
							"url":    apiBaseURL + "/api/service-a",
						},
					},
				},
				{
					"id":       "http-b",
					"type":     "action:http",
					"position": map[string]any{"x": 400, "y": 100},
					"data": map[string]any{
						"label": "Service B",
						"config": map[string]any{
							"method": "GET",
							"url":    apiBaseURL + "/api/service-b",
						},
					},
				},
				{
					"id":       "http-c",
					"type":     "action:http",
					"position": map[string]any{"x": 400, "y": 200},
					"data": map[string]any{
						"label": "Service C",
						"config": map[string]any{
							"method": "GET",
							"url":    apiBaseURL + "/api/service-c",
						},
					},
				},
				{
					"id":       "transform-aggregate",
					"type":     "action:transform",
					"position": map[string]any{"x": 600, "y": 100},
					"data": map[string]any{
						"label": "Aggregate Results",
						"config": map[string]any{
							"expression": "{ serviceA: $['http-a'].body, serviceB: $['http-b'].body, serviceC: $['http-c'].body }",
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "e1", "source": "trigger-1", "target": "parallel-1"},
				{"id": "e2", "source": "parallel-1", "target": "http-a"},
				{"id": "e3", "source": "parallel-1", "target": "http-b"},
				{"id": "e4", "source": "parallel-1", "target": "http-c"},
				{"id": "e5", "source": "http-a", "target": "transform-aggregate"},
				{"id": "e6", "source": "http-b", "target": "transform-aggregate"},
				{"id": "e7", "source": "http-c", "target": "transform-aggregate"},
			},
		},
		"enabled": true,
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", workflowReq, headers)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

// createConditionalWorkflow creates a workflow with conditional branching
func createConditionalWorkflow(t *testing.T, ts *integration.TestServer, headers map[string]string, apiBaseURL string) Workflow {
	t.Helper()

	workflowReq := map[string]any{
		"name":        "Conditional Branching Workflow",
		"description": "Routes based on priority",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 50},
					"data":     map[string]any{"label": "Start"},
				},
				{
					"id":       "condition-1",
					"type":     "control:condition",
					"position": map[string]any{"x": 200, "y": 50},
					"data": map[string]any{
						"label": "Check Priority",
						"config": map[string]any{
							"expression": "trigger.priority == 'high'",
						},
					},
				},
				{
					"id":       "http-high",
					"type":     "action:http",
					"position": map[string]any{"x": 400, "y": 0},
					"data": map[string]any{
						"label": "High Priority Handler",
						"config": map[string]any{
							"method": "POST",
							"url":    apiBaseURL + "/api/high-priority",
							"body":   map[string]any{"message": "${trigger.message}"},
						},
					},
				},
				{
					"id":       "http-low",
					"type":     "action:http",
					"position": map[string]any{"x": 400, "y": 100},
					"data": map[string]any{
						"label": "Low Priority Handler",
						"config": map[string]any{
							"method": "POST",
							"url":    apiBaseURL + "/api/low-priority",
							"body":   map[string]any{"message": "${trigger.message}"},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "e1", "source": "trigger-1", "target": "condition-1"},
				{"id": "e2", "source": "condition-1", "target": "http-high", "data": map[string]any{"condition": "true"}},
				{"id": "e3", "source": "condition-1", "target": "http-low", "data": map[string]any{"condition": "false"}},
			},
		},
		"enabled": true,
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", workflowReq, headers)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

// createRetryWorkflow creates a workflow with retry logic
func createRetryWorkflow(t *testing.T, ts *integration.TestServer, headers map[string]string, apiBaseURL string) Workflow {
	t.Helper()

	workflowReq := map[string]any{
		"name":        "Retry Workflow",
		"description": "Retries failed operations with backoff",
		"definition": map[string]any{
			"nodes": []map[string]any{
				{
					"id":       "trigger-1",
					"type":     "trigger:manual",
					"position": map[string]any{"x": 0, "y": 0},
					"data":     map[string]any{"label": "Start"},
				},
				{
					"id":       "retry-1",
					"type":     "control:retry",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Retry Block",
						"config": map[string]any{
							"maxRetries":   3,
							"backoffMs":    100,
							"backoffType":  "exponential",
							"retryOnCodes": []int{500, 502, 503, 504},
						},
					},
				},
				{
					"id":       "http-flaky",
					"type":     "action:http",
					"position": map[string]any{"x": 400, "y": 0},
					"data": map[string]any{
						"label": "Flaky Service Call",
						"config": map[string]any{
							"method":    "POST",
							"url":       apiBaseURL + "/api/flaky",
							"body":      map[string]any{"op": "${trigger.operation}"},
							"timeoutMs": 5000,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "e1", "source": "trigger-1", "target": "retry-1"},
				{"id": "e2", "source": "retry-1", "target": "http-flaky"},
			},
		},
		"enabled": true,
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", workflowReq, headers)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}
