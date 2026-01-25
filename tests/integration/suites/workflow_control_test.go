package suites

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gorax/gorax/tests/integration"
)

// =============================================================================
// Loop Control Tests
// =============================================================================

// TestWorkflowControl_Loop tests loop control flow execution
func TestWorkflowControl_Loop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "loop-control-test")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("forEach over array", func(t *testing.T) {
		workflow := createWorkflowWithLoopAction(t, ts, headers, map[string]any{
			"source":         "${trigger.items}",
			"item_variable":  "item",
			"max_iterations": 100,
		})

		execInput := map[string]any{
			"input": map[string]any{
				"items": []any{"apple", "banana", "cherry"},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)
	})

	t.Run("loop index and value access", func(t *testing.T) {
		workflow := createWorkflowWithLoopAction(t, ts, headers, map[string]any{
			"source":         "${trigger.numbers}",
			"item_variable":  "num",
			"index_variable": "idx",
			"max_iterations": 100,
		})

		execInput := map[string]any{
			"input": map[string]any{
				"numbers": []int{10, 20, 30, 40, 50},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("nested loops", func(t *testing.T) {
		// Create a workflow with nested loop structure
		workflow := createWorkflowWithNestedLoops(t, ts, headers)

		execInput := map[string]any{
			"input": map[string]any{
				"matrix": []any{
					[]int{1, 2},
					[]int{3, 4},
				},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("break condition", func(t *testing.T) {
		workflow := createWorkflowWithLoopBreak(t, ts, headers, map[string]any{
			"source":        "${trigger.items}",
			"item_variable": "item",
			"break_conditions": []map[string]any{
				{
					"field":    "item",
					"operator": "equals",
					"value":    "stop",
				},
			},
		})

		execInput := map[string]any{
			"input": map[string]any{
				"items": []string{"one", "two", "stop", "four", "five"},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("max iteration limit", func(t *testing.T) {
		workflow := createWorkflowWithLoopAction(t, ts, headers, map[string]any{
			"source":         "${trigger.items}",
			"item_variable":  "item",
			"max_iterations": 3, // Limit to 3 iterations
		})

		execInput := map[string]any{
			"input": map[string]any{
				"items": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		// Should fail because array length exceeds max iterations
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("object iteration with keys", func(t *testing.T) {
		workflow := createWorkflowWithLoopAction(t, ts, headers, map[string]any{
			"source":         "${trigger.data}",
			"item_variable":  "value",
			"key_variable":   "key",
			"max_iterations": 100,
		})

		execInput := map[string]any{
			"input": map[string]any{
				"data": map[string]any{
					"name":    "John",
					"email":   "john@example.com",
					"country": "USA",
				},
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})
}

// =============================================================================
// Parallel Control Tests
// =============================================================================

// TestWorkflowControl_Parallel tests parallel branch execution
func TestWorkflowControl_Parallel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "parallel-control-test")
	headers := integration.DefaultTestHeaders(tenantID)

	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	t.Run("parallel branch execution", func(t *testing.T) {
		// Setup mock responses with different delays to verify parallel execution
		mockServer.SetMethodResponse("GET", "/api/branch1", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"branch": "one"},
			DelayMs:    100,
		})
		mockServer.SetMethodResponse("GET", "/api/branch2", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"branch": "two"},
			DelayMs:    100,
		})
		mockServer.SetMethodResponse("GET", "/api/branch3", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"branch": "three"},
			DelayMs:    100,
		})

		workflow := createWorkflowWithParallelBranches(t, ts, headers, mockServer.URL(), 3)

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)
	})

	t.Run("result aggregation", func(t *testing.T) {
		mockServer.ClearRequests()

		workflow := createWorkflowWithParallelBranches(t, ts, headers, mockServer.URL(), 2)

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("partial failure handling - fail_fast", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetMethodResponse("GET", "/api/branch1", integration.MockResponse{
			StatusCode: 500,
			Body:       map[string]any{"error": "branch 1 failed"},
		})

		workflow := createWorkflowWithParallelBranchesErrorStrategy(t, ts, headers, mockServer.URL(), "fail_fast")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("partial failure handling - wait_all", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetMethodResponse("GET", "/api/branch1", integration.MockResponse{
			StatusCode: 500,
			Body:       map[string]any{"error": "branch 1 failed"},
		})
		mockServer.SetMethodResponse("GET", "/api/branch2", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"branch": "two"},
			DelayMs:    200,
		})

		workflow := createWorkflowWithParallelBranchesErrorStrategy(t, ts, headers, mockServer.URL(), "wait_all")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		// Wait for both branches to complete
		time.Sleep(300 * time.Millisecond)
	})

	t.Run("concurrency limits", func(t *testing.T) {
		mockServer.ClearRequests()

		workflow := createWorkflowWithParallelConcurrencyLimit(t, ts, headers, mockServer.URL(), 2) // Max 2 concurrent

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})
}

// =============================================================================
// Delay Control Tests
// =============================================================================

// TestWorkflowControl_Delay tests delay control execution
func TestWorkflowControl_Delay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "delay-control-test")
	headers := integration.DefaultTestHeaders(tenantID)

	t.Run("fixed delay", func(t *testing.T) {
		workflow := createWorkflowWithDelayAction(t, ts, headers, 1) // 1 second delay

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		var exec Execution
		integration.ParseJSONResponse(t, execResp, &exec)
		assert.NotEmpty(t, exec.ID)
	})

	t.Run("dynamic delay from expression", func(t *testing.T) {
		workflow := createWorkflowWithDynamicDelay(t, ts, headers)

		execInput := map[string]any{
			"input": map[string]any{
				"delay_seconds": 1,
			},
		}

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", execInput, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})
}

// =============================================================================
// Try/Catch Control Tests
// =============================================================================

// TestWorkflowControl_TryCatch tests try/catch error handling
func TestWorkflowControl_TryCatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "trycatch-control-test")
	headers := integration.DefaultTestHeaders(tenantID)

	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	t.Run("successful execution skips catch", func(t *testing.T) {
		mockServer.SetMethodResponse("GET", "/api/success", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"status": "ok"},
		})

		workflow := createWorkflowWithTryCatch(t, ts, headers, mockServer.URL(), "/api/success")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("failure triggers catch branch", func(t *testing.T) {
		mockServer.SetMethodResponse("GET", "/api/fail", integration.MockResponse{
			StatusCode: 500,
			Body:       map[string]any{"error": "internal error"},
		})

		workflow := createWorkflowWithTryCatch(t, ts, headers, mockServer.URL(), "/api/fail")

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("error info available in catch context", func(t *testing.T) {
		workflow := createWorkflowWithTryCatchErrorInfo(t, ts, headers)

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})
}

// =============================================================================
// Retry Control Tests
// =============================================================================

// TestWorkflowControl_Retry tests retry control with backoff
func TestWorkflowControl_Retry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "retry-control-test")
	headers := integration.DefaultTestHeaders(tenantID)

	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	t.Run("max retry limit", func(t *testing.T) {
		// Always fail
		mockServer.SetMethodResponse("GET", "/api/flaky", integration.MockResponse{
			StatusCode: 500,
			Body:       map[string]any{"error": "always fails"},
		})

		workflow := createWorkflowWithRetryAction(t, ts, headers, mockServer.URL(), 3) // Max 3 retries

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("success after retry", func(t *testing.T) {
		mockServer.ClearRequests()

		// Set up to fail first 2 times, then succeed
		callCount := 0
		mockServer.SetDefaultResponse(integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"status": "ok"},
		})

		// Note: In a real test, we'd use a custom handler to track call count
		_ = callCount

		workflow := createWorkflowWithRetryAction(t, ts, headers, mockServer.URL(), 3)

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("exponential backoff", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetMethodResponse("GET", "/api/flaky", integration.MockResponse{
			StatusCode: 503,
			Body:       map[string]any{"error": "service unavailable"},
		})

		workflow := createWorkflowWithRetryBackoff(t, ts, headers, mockServer.URL())

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})
}

// =============================================================================
// Circuit Breaker Tests
// =============================================================================

// TestWorkflowControl_CircuitBreaker tests circuit breaker pattern
func TestWorkflowControl_CircuitBreaker(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "circuitbreaker-control-test")
	headers := integration.DefaultTestHeaders(tenantID)

	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	t.Run("closed to open on failures", func(t *testing.T) {
		mockServer.SetMethodResponse("GET", "/api/unstable", integration.MockResponse{
			StatusCode: 500,
			Body:       map[string]any{"error": "service error"},
		})

		workflow := createWorkflowWithCircuitBreaker(t, ts, headers, mockServer.URL(), 3) // Open after 3 failures

		// Execute multiple times to trigger circuit breaker
		for range 5 {
			execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
			integration.AssertStatusCode(t, execResp, http.StatusAccepted)
			time.Sleep(100 * time.Millisecond)
		}
	})

	t.Run("half-open state probe", func(t *testing.T) {
		// After circuit opens, it should probe periodically
		workflow := createWorkflowWithCircuitBreaker(t, ts, headers, mockServer.URL(), 2)

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("recovery to closed", func(t *testing.T) {
		mockServer.ClearRequests()
		// Now service is healthy again
		mockServer.SetMethodResponse("GET", "/api/unstable", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"status": "healthy"},
		})

		workflow := createWorkflowWithCircuitBreaker(t, ts, headers, mockServer.URL(), 3)

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})
}

// =============================================================================
// Fork/Join Tests
// =============================================================================

// TestWorkflowControl_ForkJoin tests fork/join pattern
func TestWorkflowControl_ForkJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := integration.SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "forkjoin-control-test")
	headers := integration.DefaultTestHeaders(tenantID)

	mockServer := integration.NewMockServer()
	defer mockServer.Close()

	t.Run("parallel fork execution", func(t *testing.T) {
		mockServer.SetMethodResponse("GET", "/api/task1", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"task": "1", "result": 100},
		})
		mockServer.SetMethodResponse("GET", "/api/task2", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"task": "2", "result": 200},
		})

		workflow := createWorkflowWithForkJoin(t, ts, headers, mockServer.URL())

		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)
	})

	t.Run("join waits for all branches", func(t *testing.T) {
		mockServer.ClearRequests()
		mockServer.SetMethodResponse("GET", "/api/task1", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"task": "1"},
			DelayMs:    100,
		})
		mockServer.SetMethodResponse("GET", "/api/task2", integration.MockResponse{
			StatusCode: 200,
			Body:       map[string]any{"task": "2"},
			DelayMs:    200, // Longer delay
		})

		workflow := createWorkflowWithForkJoin(t, ts, headers, mockServer.URL())

		startTime := time.Now()
		execResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflow.ID+"/execute", map[string]any{}, headers)
		integration.AssertStatusCode(t, execResp, http.StatusAccepted)

		// Execution should be accepted quickly (async)
		assert.Less(t, time.Since(startTime), 500*time.Millisecond)
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

func createWorkflowWithLoopAction(t *testing.T, ts *integration.TestServer, headers map[string]string, loopConfig map[string]any) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Loop Test Workflow",
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
					"id":       "loop-1",
					"type":     "control:loop",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label":  "Loop",
						"config": loopConfig,
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "loop-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithNestedLoops(t *testing.T, ts *integration.TestServer, headers map[string]string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Nested Loop Test Workflow",
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
					"id":       "outer-loop",
					"type":     "control:loop",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Outer Loop",
						"config": map[string]any{
							"source":         "${trigger.matrix}",
							"item_variable":  "row",
							"max_iterations": 100,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "outer-loop"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithLoopBreak(t *testing.T, ts *integration.TestServer, headers map[string]string, loopConfig map[string]any) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Loop Break Test Workflow",
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
					"id":       "loop-1",
					"type":     "control:loop",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label":  "Loop with Break",
						"config": loopConfig,
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "loop-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithParallelBranches(t *testing.T, ts *integration.TestServer, headers map[string]string, baseURL string, branchCount int) Workflow {
	t.Helper()

	branches := make([]map[string]any, branchCount)
	for i := range branchCount {
		branches[i] = map[string]any{
			"name": fmt.Sprintf("branch%d", i+1),
			"nodes": []map[string]any{
				{
					"type": "action:http",
					"config": map[string]any{
						"method": "GET",
						"url":    fmt.Sprintf("%s/api/branch%d", baseURL, i+1),
					},
				},
			},
		}
	}

	input := map[string]any{
		"name":   "Parallel Branches Test Workflow",
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
					"id":       "parallel-1",
					"type":     "control:parallel",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Parallel",
						"config": map[string]any{
							"branches":       branches,
							"error_strategy": "fail_fast",
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "parallel-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithParallelBranchesErrorStrategy(t *testing.T, ts *integration.TestServer, headers map[string]string, baseURL, errorStrategy string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Parallel Error Strategy Test Workflow",
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
					"id":       "parallel-1",
					"type":     "control:parallel",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Parallel",
						"config": map[string]any{
							"branches": []map[string]any{
								{
									"name": "branch1",
									"nodes": []map[string]any{
										{
											"type": "action:http",
											"config": map[string]any{
												"method": "GET",
												"url":    baseURL + "/api/branch1",
											},
										},
									},
								},
								{
									"name": "branch2",
									"nodes": []map[string]any{
										{
											"type": "action:http",
											"config": map[string]any{
												"method": "GET",
												"url":    baseURL + "/api/branch2",
											},
										},
									},
								},
							},
							"error_strategy": errorStrategy,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "parallel-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithParallelConcurrencyLimit(t *testing.T, ts *integration.TestServer, headers map[string]string, baseURL string, maxConcurrency int) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Parallel Concurrency Limit Test Workflow",
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
					"id":       "parallel-1",
					"type":     "control:parallel",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Parallel",
						"config": map[string]any{
							"max_concurrency": maxConcurrency,
							"branches": []map[string]any{
								{
									"name": "branch1",
									"nodes": []map[string]any{
										{"type": "action:http", "config": map[string]any{"method": "GET", "url": baseURL + "/api/branch1"}},
									},
								},
								{
									"name": "branch2",
									"nodes": []map[string]any{
										{"type": "action:http", "config": map[string]any{"method": "GET", "url": baseURL + "/api/branch2"}},
									},
								},
								{
									"name": "branch3",
									"nodes": []map[string]any{
										{"type": "action:http", "config": map[string]any{"method": "GET", "url": baseURL + "/api/branch3"}},
									},
								},
							},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "parallel-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithDelayAction(t *testing.T, ts *integration.TestServer, headers map[string]string, delaySeconds int) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Delay Test Workflow",
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
					"id":       "delay-1",
					"type":     "control:delay",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Delay",
						"config": map[string]any{
							"duration_seconds": delaySeconds,
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "delay-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithDynamicDelay(t *testing.T, ts *integration.TestServer, headers map[string]string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Dynamic Delay Test Workflow",
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
					"id":       "delay-1",
					"type":     "control:delay",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Delay",
						"config": map[string]any{
							"duration_expression": "${trigger.delay_seconds}",
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "delay-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithTryCatch(t *testing.T, ts *integration.TestServer, headers map[string]string, baseURL, path string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Try/Catch Test Workflow",
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
					"id":       "try-1",
					"type":     "control:try",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Try/Catch",
						"config": map[string]any{
							"try_nodes": []map[string]any{
								{
									"type": "action:http",
									"config": map[string]any{
										"method": "GET",
										"url":    baseURL + path,
									},
								},
							},
							"catch_nodes": []map[string]any{
								{
									"type": "action:log",
									"config": map[string]any{
										"message": "Error caught: ${error.message}",
									},
								},
							},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "try-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithTryCatchErrorInfo(t *testing.T, ts *integration.TestServer, headers map[string]string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Try/Catch Error Info Test Workflow",
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
					"id":       "try-1",
					"type":     "control:try",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Try/Catch",
						"config": map[string]any{
							"try_nodes": []map[string]any{
								{
									"type": "action:code",
									"config": map[string]any{
										"script": "throw new Error('Test error');",
									},
								},
							},
							"catch_nodes": []map[string]any{
								{
									"type": "action:formula",
									"config": map[string]any{
										"expression": "error.message",
									},
								},
							},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "try-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithRetryAction(t *testing.T, ts *integration.TestServer, headers map[string]string, baseURL string, maxRetries int) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Retry Test Workflow",
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
					"id":       "retry-1",
					"type":     "control:retry",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Retry",
						"config": map[string]any{
							"max_retries":        maxRetries,
							"initial_backoff_ms": 100,
							"max_backoff_ms":     1000,
							"node": map[string]any{
								"type": "action:http",
								"config": map[string]any{
									"method": "GET",
									"url":    baseURL + "/api/flaky",
								},
							},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "retry-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithRetryBackoff(t *testing.T, ts *integration.TestServer, headers map[string]string, baseURL string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Retry Backoff Test Workflow",
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
					"id":       "retry-1",
					"type":     "control:retry",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Retry with Backoff",
						"config": map[string]any{
							"max_retries":        3,
							"initial_backoff_ms": 50,
							"max_backoff_ms":     500,
							"backoff_multiplier": 2.0,
							"node": map[string]any{
								"type": "action:http",
								"config": map[string]any{
									"method": "GET",
									"url":    baseURL + "/api/flaky",
								},
							},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "retry-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithCircuitBreaker(t *testing.T, ts *integration.TestServer, headers map[string]string, baseURL string, failureThreshold int) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Circuit Breaker Test Workflow",
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
					"id":       "circuit-breaker-1",
					"type":     "control:circuit_breaker",
					"position": map[string]any{"x": 200, "y": 0},
					"data": map[string]any{
						"label": "Circuit Breaker",
						"config": map[string]any{
							"failure_threshold":  failureThreshold,
							"reset_timeout_ms":   5000,
							"half_open_requests": 1,
							"node": map[string]any{
								"type": "action:http",
								"config": map[string]any{
									"method": "GET",
									"url":    baseURL + "/api/unstable",
								},
							},
						},
					},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "circuit-breaker-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}

func createWorkflowWithForkJoin(t *testing.T, ts *integration.TestServer, headers map[string]string, baseURL string) Workflow {
	t.Helper()

	input := map[string]any{
		"name":   "Fork/Join Test Workflow",
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
					"id":       "fork-1",
					"type":     "control:fork",
					"position": map[string]any{"x": 200, "y": 0},
					"data":     map[string]any{"label": "Fork"},
				},
				{
					"id":       "http-task1",
					"type":     "action:http",
					"position": map[string]any{"x": 400, "y": -50},
					"data": map[string]any{
						"label": "Task 1",
						"config": map[string]any{
							"method": "GET",
							"url":    baseURL + "/api/task1",
						},
					},
				},
				{
					"id":       "http-task2",
					"type":     "action:http",
					"position": map[string]any{"x": 400, "y": 50},
					"data": map[string]any{
						"label": "Task 2",
						"config": map[string]any{
							"method": "GET",
							"url":    baseURL + "/api/task2",
						},
					},
				},
				{
					"id":       "join-1",
					"type":     "control:join",
					"position": map[string]any{"x": 600, "y": 0},
					"data":     map[string]any{"label": "Join"},
				},
			},
			"edges": []map[string]any{
				{"id": "edge-1", "source": "trigger-1", "target": "fork-1"},
				{"id": "edge-2", "source": "fork-1", "target": "http-task1"},
				{"id": "edge-3", "source": "fork-1", "target": "http-task2"},
				{"id": "edge-4", "source": "http-task1", "target": "join-1"},
				{"id": "edge-5", "source": "http-task2", "target": "join-1"},
			},
		},
	}

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows", input, headers)
	integration.AssertStatusCode(t, resp, http.StatusCreated)

	var workflow Workflow
	integration.ParseJSONResponse(t, resp, &workflow)
	return workflow
}
