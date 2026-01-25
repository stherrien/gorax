package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/integration"
	"github.com/gorax/gorax/internal/integration/integrations"
	inttesting "github.com/gorax/gorax/internal/integration/testing"
)

// TestJiraWorkflow_CreateIssue tests end-to-end workflow with Jira issue creation action.
func TestJiraWorkflow_CreateIssue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create Jira integration and test suite
	jiraInteg := integrations.NewJiraIntegration(nil)
	ts := inttesting.NewTestSuite(t, jiraInteg)
	defer ts.Close()

	// Configure mock Jira server
	mockServer := ts.MockServer()
	mockIssueResponse := map[string]interface{}{
		"id":   "10001",
		"key":  "TEST-123",
		"self": "https://example.atlassian.net/rest/api/3/issue/10001",
	}
	mockServer.OnPost("/rest/api/3/issue", inttesting.JSONResponse(http.StatusCreated, mockIssueResponse))

	// Set up test configuration with mock server URL
	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": mockServer.URL(),
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-api-token",
			},
		},
	}
	ts.WithConfig(config)

	t.Run("CreateIssue_Success", func(t *testing.T) {
		params := integration.JSONMap{
			"action":      "create_issue",
			"project":     "TEST",
			"summary":     "Test Issue from Integration Test",
			"issue_type":  "Task",
			"description": "This is a test issue created by integration tests",
			"priority":    "High",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertSuccess(result, err)
		ts.AssertStatusCode(result, http.StatusCreated)
		ts.AssertRequestCount(1)
		ts.AssertLastRequestMethod("POST")
		ts.AssertLastRequestPath("/rest/api/3/issue")

		// Verify response data
		data, ok := result.Data.(map[string]interface{})
		require.True(t, ok, "result data should be a map")
		assert.Equal(t, "10001", data["id"])
		assert.Equal(t, "TEST-123", data["key"])

		t.Logf("✓ Created issue: %s", data["key"])
	})

	t.Run("CreateIssue_VerifyRequestPayload", func(t *testing.T) {
		mockServer.ClearRequests()

		params := integration.JSONMap{
			"action":      "create_issue",
			"project":     "PROJ",
			"summary":     "Payload Test Issue",
			"issue_type":  "Bug",
			"description": "Testing payload structure",
			"priority":    "Medium",
			"labels":      []string{"test", "integration"},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := ts.Execute(ctx, params)
		require.NoError(t, err)

		// Verify request payload
		lastReq := mockServer.GetLastRequest()
		require.NotNil(t, lastReq)

		var payload map[string]interface{}
		err = json.Unmarshal(lastReq.Body, &payload)
		require.NoError(t, err)

		fields, ok := payload["fields"].(map[string]interface{})
		require.True(t, ok, "payload should have fields")

		// Verify project
		project, ok := fields["project"].(map[string]interface{})
		require.True(t, ok, "fields should have project")
		assert.Equal(t, "PROJ", project["key"])

		// Verify issue type
		issueType, ok := fields["issuetype"].(map[string]interface{})
		require.True(t, ok, "fields should have issuetype")
		assert.Equal(t, "Bug", issueType["name"])

		// Verify summary
		assert.Equal(t, "Payload Test Issue", fields["summary"])

		// Verify priority
		priority, ok := fields["priority"].(map[string]interface{})
		require.True(t, ok, "fields should have priority")
		assert.Equal(t, "Medium", priority["name"])

		t.Logf("✓ Request payload verified")
	})
}

// TestJiraWorkflow_Authentication tests credential usage and authentication.
func TestJiraWorkflow_Authentication(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	jiraInteg := integrations.NewJiraIntegration(nil)
	ts := inttesting.NewTestSuite(t, jiraInteg)
	defer ts.Close()

	mockServer := ts.MockServer()
	mockProjectResponse := map[string]interface{}{
		"id":   "10000",
		"key":  "TEST",
		"name": "Test Project",
	}
	mockServer.OnGet("/rest/api/3/project/TEST", inttesting.JSONResponse(http.StatusOK, mockProjectResponse))

	t.Run("BasicAuth_EmailAPIToken", func(t *testing.T) {
		config := &integration.Config{
			Name:    "test-jira-basic",
			Type:    integration.TypeAPI,
			Enabled: true,
			Settings: integration.JSONMap{
				"base_url": mockServer.URL(),
			},
			Credentials: &integration.Credentials{
				Data: integration.JSONMap{
					"email":     "user@example.com",
					"api_token": "test-api-token",
				},
			},
		}
		ts.WithConfig(config)

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "TEST",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertSuccess(result, err)

		// Verify Authorization header
		lastReq := mockServer.GetLastRequest()
		require.NotNil(t, lastReq)

		authHeader := lastReq.Headers.Get("Authorization")
		assert.True(t, strings.HasPrefix(authHeader, "Basic "), "should use Basic auth")

		t.Logf("✓ Basic auth (email/api_token) verified")
	})

	t.Run("BasicAuth_UsernamePassword", func(t *testing.T) {
		mockServer.ClearRequests()

		config := &integration.Config{
			Name:    "test-jira-userpass",
			Type:    integration.TypeAPI,
			Enabled: true,
			Settings: integration.JSONMap{
				"base_url": mockServer.URL(),
			},
			Credentials: &integration.Credentials{
				Data: integration.JSONMap{
					"username": "admin",
					"password": "secret123",
				},
			},
		}
		ts.WithConfig(config)

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "TEST",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertSuccess(result, err)

		// Verify Authorization header
		lastReq := mockServer.GetLastRequest()
		require.NotNil(t, lastReq)

		authHeader := lastReq.Headers.Get("Authorization")
		assert.True(t, strings.HasPrefix(authHeader, "Basic "), "should use Basic auth")

		t.Logf("✓ Basic auth (username/password) verified")
	})

	t.Run("BearerAuth_Token", func(t *testing.T) {
		mockServer.ClearRequests()

		config := &integration.Config{
			Name:    "test-jira-bearer",
			Type:    integration.TypeAPI,
			Enabled: true,
			Settings: integration.JSONMap{
				"base_url": mockServer.URL(),
			},
			Credentials: &integration.Credentials{
				Data: integration.JSONMap{
					"token": "bearer-token-12345",
				},
			},
		}
		ts.WithConfig(config)

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "TEST",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertSuccess(result, err)

		// Verify Authorization header
		lastReq := mockServer.GetLastRequest()
		require.NotNil(t, lastReq)

		authHeader := lastReq.Headers.Get("Authorization")
		assert.Equal(t, "Bearer bearer-token-12345", authHeader, "should use Bearer auth")

		t.Logf("✓ Bearer auth verified")
	})

	t.Run("Auth_MissingCredentials", func(t *testing.T) {
		config := &integration.Config{
			Name:    "test-jira-no-creds",
			Type:    integration.TypeAPI,
			Enabled: true,
			Settings: integration.JSONMap{
				"base_url": mockServer.URL(),
			},
			Credentials: nil,
		}
		ts.WithConfig(config)

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "TEST",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertError(err)
		assert.False(t, result.Success)
		assert.Equal(t, "AUTH_ERROR", result.ErrorCode)

		t.Logf("✓ Missing credentials handled correctly")
	})

	t.Run("Auth_InvalidCredentials", func(t *testing.T) {
		config := &integration.Config{
			Name:    "test-jira-empty-creds",
			Type:    integration.TypeAPI,
			Enabled: true,
			Settings: integration.JSONMap{
				"base_url": mockServer.URL(),
			},
			Credentials: &integration.Credentials{
				Data: integration.JSONMap{},
			},
		}
		ts.WithConfig(config)

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "TEST",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertError(err)
		assert.False(t, result.Success)
		assert.Equal(t, "AUTH_ERROR", result.ErrorCode)

		t.Logf("✓ Invalid credentials handled correctly")
	})
}

// TestJiraWorkflow_RateLimiting tests rate limit handling.
func TestJiraWorkflow_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	jiraInteg := integrations.NewJiraIntegration(nil)
	ts := inttesting.NewTestSuite(t, jiraInteg)
	defer ts.Close()

	mockServer := ts.MockServer()

	t.Run("RateLimit_429Response", func(t *testing.T) {
		// Configure mock to return 429 initially, then succeed
		callCount := 0
		mockServer.OnRequest("GET", "/rest/api/3/project/TEST", &inttesting.MockResponse{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if callCount == 1 {
					// First call: rate limited
					w.Header().Set("Retry-After", "1")
					w.WriteHeader(http.StatusTooManyRequests)
					_, _ = w.Write([]byte(`{"errorMessages":["Rate limit exceeded"]}`))
					return
				}
				// Subsequent calls: success
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"id":"10000","key":"TEST","name":"Test Project"}`))
			}),
		})

		config := &integration.Config{
			Name:    "test-jira",
			Type:    integration.TypeAPI,
			Enabled: true,
			Settings: integration.JSONMap{
				"base_url": mockServer.URL(),
			},
			Credentials: &integration.Credentials{
				Data: integration.JSONMap{
					"email":     "user@example.com",
					"api_token": "test-token",
				},
			},
		}
		ts.WithConfig(config)

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "TEST",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		// With retry logic, should eventually succeed
		ts.AssertSuccess(result, err)
		assert.GreaterOrEqual(t, callCount, 2, "should have retried after rate limit")

		t.Logf("✓ Rate limiting handled with %d calls", callCount)
	})

	t.Run("RateLimit_PersistentFailure", func(t *testing.T) {
		mockServer.Reset()

		// Configure mock to always return 429
		mockServer.OnGet("/rest/api/3/project/FAIL", &inttesting.MockResponse{
			StatusCode: http.StatusTooManyRequests,
			Headers: map[string]string{
				"Retry-After":  "60",
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"errorMessages": []string{"Rate limit exceeded"},
			},
		})

		config := &integration.Config{
			Name:    "test-jira",
			Type:    integration.TypeAPI,
			Enabled: true,
			Settings: integration.JSONMap{
				"base_url": mockServer.URL(),
			},
			Credentials: &integration.Credentials{
				Data: integration.JSONMap{
					"email":     "user@example.com",
					"api_token": "test-token",
				},
			},
		}
		ts.WithConfig(config)

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "FAIL",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		// Should fail after max retries
		ts.AssertError(err)
		assert.False(t, result.Success)

		// Verify retry attempts occurred
		reqCount := mockServer.GetRequestCount()
		assert.GreaterOrEqual(t, reqCount, 1, "should have made at least one request")

		t.Logf("✓ Persistent rate limit failure handled after %d attempts", reqCount)
	})
}

// TestJiraWorkflow_ErrorHandling tests various error scenarios.
func TestJiraWorkflow_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	jiraInteg := integrations.NewJiraIntegration(nil)
	ts := inttesting.NewTestSuite(t, jiraInteg)
	defer ts.Close()

	mockServer := ts.MockServer()

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": mockServer.URL(),
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}
	ts.WithConfig(config)

	t.Run("Error_JiraValidationError", func(t *testing.T) {
		mockServer.Reset()

		// Jira validation error response
		mockServer.OnPost("/rest/api/3/issue", &inttesting.MockResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"errorMessages": []string{},
				"errors": map[string]string{
					"project":   "project is required",
					"issuetype": "issue type is required",
				},
			},
		})

		params := integration.JSONMap{
			"action":  "create_issue",
			"project": "INVALID",
			"summary": "Test Issue",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertError(err)
		assert.False(t, result.Success)
		assert.Equal(t, "API_ERROR", result.ErrorCode)
		assert.Contains(t, err.Error(), "validation errors")

		t.Logf("✓ Jira validation error handled")
	})

	t.Run("Error_NotFound", func(t *testing.T) {
		mockServer.Reset()

		mockServer.OnGet("/rest/api/3/issue/NONEXISTENT-999", &inttesting.MockResponse{
			StatusCode: http.StatusNotFound,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"errorMessages": []string{"Issue does not exist or you do not have permission to see it."},
			},
		})

		params := integration.JSONMap{
			"action":    "get_issue",
			"issue_key": "NONEXISTENT-999",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertError(err)
		assert.False(t, result.Success)
		assert.Contains(t, err.Error(), "does not exist")

		t.Logf("✓ Not found error handled")
	})

	t.Run("Error_Unauthorized", func(t *testing.T) {
		mockServer.Reset()

		mockServer.OnGet("/rest/api/3/project/TEST", &inttesting.MockResponse{
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"errorMessages": []string{"Invalid credentials"},
			},
		})

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "TEST",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertError(err)
		assert.False(t, result.Success)

		t.Logf("✓ Unauthorized error handled")
	})

	t.Run("Error_ServerError", func(t *testing.T) {
		mockServer.Reset()

		// Server error should trigger retry
		callCount := 0
		mockServer.OnRequest("GET", "/rest/api/3/project/ERROR", &inttesting.MockResponse{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"errorMessages":["Internal server error"]}`))
			}),
		})

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "ERROR",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertError(err)
		assert.False(t, result.Success)
		assert.GreaterOrEqual(t, callCount, 1, "should have attempted request")

		t.Logf("✓ Server error handled with %d attempts", callCount)
	})

	t.Run("Error_MissingRequiredParams", func(t *testing.T) {
		mockServer.Reset()

		// Missing project for create_issue
		params := integration.JSONMap{
			"action":  "create_issue",
			"summary": "Test Issue",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertError(err)
		assert.False(t, result.Success)
		assert.Equal(t, "VALIDATION_ERROR", result.ErrorCode)

		t.Logf("✓ Missing required params error handled")
	})

	t.Run("Error_InvalidAction", func(t *testing.T) {
		params := integration.JSONMap{
			"action": "invalid_action",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertError(err)
		assert.False(t, result.Success)
		assert.Equal(t, "INVALID_ACTION", result.ErrorCode)

		t.Logf("✓ Invalid action error handled")
	})
}

// TestJiraWorkflow_ExecutionLogs tests that execution logs contain Jira response data.
func TestJiraWorkflow_ExecutionLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	jiraInteg := integrations.NewJiraIntegration(nil)
	ts := inttesting.NewTestSuite(t, jiraInteg)
	defer ts.Close()

	mockServer := ts.MockServer()

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": mockServer.URL(),
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}
	ts.WithConfig(config)

	t.Run("Logs_SuccessfulExecution", func(t *testing.T) {
		mockServer.Reset()

		expectedIssue := map[string]interface{}{
			"id":   "10001",
			"key":  "TEST-123",
			"self": "https://example.atlassian.net/rest/api/3/issue/10001",
			"fields": map[string]interface{}{
				"summary": "Test Issue",
				"status": map[string]interface{}{
					"name": "To Do",
				},
			},
		}
		mockServer.OnPost("/rest/api/3/issue", inttesting.JSONResponse(http.StatusCreated, expectedIssue))

		params := integration.JSONMap{
			"action":     "create_issue",
			"project":    "TEST",
			"summary":    "Test Issue",
			"issue_type": "Task",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertSuccess(result, err)

		// Verify result contains Jira response data for logging
		assert.NotNil(t, result.Data, "result should contain data")
		assert.True(t, result.Success, "result should be successful")
		assert.GreaterOrEqual(t, result.Duration, int64(0), "duration should be recorded")
		assert.NotZero(t, result.ExecutedAt, "execution time should be recorded")

		data, ok := result.Data.(map[string]interface{})
		require.True(t, ok, "result data should be a map")
		assert.Equal(t, "10001", data["id"], "should contain issue ID")
		assert.Equal(t, "TEST-123", data["key"], "should contain issue key")

		t.Logf("✓ Execution log data verified: ID=%s, Key=%s, Duration=%dms",
			data["id"], data["key"], result.Duration)
	})

	t.Run("Logs_FailedExecution", func(t *testing.T) {
		mockServer.Reset()

		mockServer.OnPost("/rest/api/3/issue", &inttesting.MockResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"errorMessages": []string{"Project TEST does not exist"},
			},
		})

		params := integration.JSONMap{
			"action":  "create_issue",
			"project": "TEST",
			"summary": "Test Issue",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		// Verify error information is captured for logging
		ts.AssertError(err)
		assert.False(t, result.Success, "result should not be successful")
		assert.NotEmpty(t, result.Error, "error message should be captured")
		assert.NotEmpty(t, result.ErrorCode, "error code should be captured")
		assert.GreaterOrEqual(t, result.Duration, int64(0), "duration should be recorded even on failure")

		t.Logf("✓ Failed execution log data verified: Error=%s, Code=%s, Duration=%dms",
			result.Error, result.ErrorCode, result.Duration)
	})

	t.Run("Logs_SearchResults", func(t *testing.T) {
		mockServer.Reset()

		searchResponse := map[string]interface{}{
			"startAt":    0,
			"maxResults": 50,
			"total":      2,
			"issues": []map[string]interface{}{
				{
					"id":   "10001",
					"key":  "TEST-1",
					"self": "https://example.atlassian.net/rest/api/3/issue/10001",
					"fields": map[string]interface{}{
						"summary": "First Issue",
					},
				},
				{
					"id":   "10002",
					"key":  "TEST-2",
					"self": "https://example.atlassian.net/rest/api/3/issue/10002",
					"fields": map[string]interface{}{
						"summary": "Second Issue",
					},
				},
			},
		}
		mockServer.OnPost("/rest/api/3/search", inttesting.JSONResponse(http.StatusOK, searchResponse))

		params := integration.JSONMap{
			"action": "search_issues",
			"jql":    "project = TEST",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)

		ts.AssertSuccess(result, err)

		// Verify search results for logging
		data, ok := result.Data.(map[string]interface{})
		require.True(t, ok, "result data should be a map")

		total, ok := data["total"].(float64)
		require.True(t, ok, "should have total count")
		assert.Equal(t, float64(2), total, "should return 2 issues")

		issues, ok := data["issues"].([]interface{})
		require.True(t, ok, "should have issues array")
		assert.Len(t, issues, 2, "should have 2 issues")

		t.Logf("✓ Search execution log data verified: Total=%v, Issues=%d, Duration=%dms",
			total, len(issues), result.Duration)
	})
}

// TestJiraWorkflow_AllActions tests all supported Jira actions.
func TestJiraWorkflow_AllActions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	jiraInteg := integrations.NewJiraIntegration(nil)
	ts := inttesting.NewTestSuite(t, jiraInteg)
	defer ts.Close()

	mockServer := ts.MockServer()

	config := &integration.Config{
		Name:    "test-jira",
		Type:    integration.TypeAPI,
		Enabled: true,
		Settings: integration.JSONMap{
			"base_url": mockServer.URL(),
		},
		Credentials: &integration.Credentials{
			Data: integration.JSONMap{
				"email":     "user@example.com",
				"api_token": "test-token",
			},
		},
	}
	ts.WithConfig(config)

	t.Run("Action_UpdateIssue", func(t *testing.T) {
		mockServer.Reset()
		mockServer.OnPut("/rest/api/3/issue/TEST-123", inttesting.JSONResponse(http.StatusNoContent, nil))

		params := integration.JSONMap{
			"action":      "update_issue",
			"issue_key":   "TEST-123",
			"summary":     "Updated Summary",
			"description": "Updated description",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)
		ts.AssertSuccess(result, err)
		ts.AssertLastRequestMethod("PUT")

		t.Logf("✓ update_issue action verified")
	})

	t.Run("Action_GetIssue", func(t *testing.T) {
		mockServer.Reset()
		mockServer.OnGet("/rest/api/3/issue/TEST-123", inttesting.JSONResponse(http.StatusOK, map[string]interface{}{
			"id":  "10001",
			"key": "TEST-123",
			"fields": map[string]interface{}{
				"summary": "Test Issue",
			},
		}))

		params := integration.JSONMap{
			"action":    "get_issue",
			"issue_key": "TEST-123",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)
		ts.AssertSuccess(result, err)
		ts.AssertLastRequestMethod("GET")

		t.Logf("✓ get_issue action verified")
	})

	t.Run("Action_SearchIssues", func(t *testing.T) {
		mockServer.Reset()
		mockServer.OnPost("/rest/api/3/search", inttesting.JSONResponse(http.StatusOK, map[string]interface{}{
			"total":  1,
			"issues": []map[string]interface{}{{"id": "10001", "key": "TEST-1"}},
		}))

		params := integration.JSONMap{
			"action": "search_issues",
			"jql":    "project = TEST",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)
		ts.AssertSuccess(result, err)
		ts.AssertLastRequestMethod("POST")

		t.Logf("✓ search_issues action verified")
	})

	t.Run("Action_AddComment", func(t *testing.T) {
		mockServer.Reset()
		mockServer.OnPost("/rest/api/3/issue/TEST-123/comment", inttesting.JSONResponse(http.StatusCreated, map[string]interface{}{
			"id":   "10000",
			"body": "Test comment",
		}))

		params := integration.JSONMap{
			"action":    "add_comment",
			"issue_key": "TEST-123",
			"body":      "Test comment",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)
		ts.AssertSuccess(result, err)
		ts.AssertLastRequestMethod("POST")

		t.Logf("✓ add_comment action verified")
	})

	t.Run("Action_GetTransitions", func(t *testing.T) {
		mockServer.Reset()
		mockServer.OnGet("/rest/api/3/issue/TEST-123/transitions", inttesting.JSONResponse(http.StatusOK, map[string]interface{}{
			"transitions": []map[string]interface{}{
				{"id": "21", "name": "In Progress"},
				{"id": "31", "name": "Done"},
			},
		}))

		params := integration.JSONMap{
			"action":    "get_transitions",
			"issue_key": "TEST-123",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)
		ts.AssertSuccess(result, err)
		ts.AssertLastRequestMethod("GET")

		t.Logf("✓ get_transitions action verified")
	})

	t.Run("Action_TransitionIssue", func(t *testing.T) {
		mockServer.Reset()
		mockServer.OnPost("/rest/api/3/issue/TEST-123/transitions", inttesting.JSONResponse(http.StatusNoContent, nil))

		params := integration.JSONMap{
			"action":        "transition_issue",
			"issue_key":     "TEST-123",
			"transition_id": "21",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)
		ts.AssertSuccess(result, err)
		ts.AssertLastRequestMethod("POST")

		t.Logf("✓ transition_issue action verified")
	})

	t.Run("Action_AssignIssue", func(t *testing.T) {
		mockServer.Reset()
		mockServer.OnPut("/rest/api/3/issue/TEST-123/assignee", inttesting.JSONResponse(http.StatusNoContent, nil))

		params := integration.JSONMap{
			"action":     "assign_issue",
			"issue_key":  "TEST-123",
			"account_id": "5b10a2844c20165700ede21g",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)
		ts.AssertSuccess(result, err)
		ts.AssertLastRequestMethod("PUT")

		t.Logf("✓ assign_issue action verified")
	})

	t.Run("Action_GetProject", func(t *testing.T) {
		mockServer.Reset()
		mockServer.OnGet("/rest/api/3/project/TEST", inttesting.JSONResponse(http.StatusOK, map[string]interface{}{
			"id":   "10000",
			"key":  "TEST",
			"name": "Test Project",
		}))

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "TEST",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)
		ts.AssertSuccess(result, err)
		ts.AssertLastRequestMethod("GET")

		t.Logf("✓ get_project action verified")
	})

	t.Run("Action_ListProjects", func(t *testing.T) {
		mockServer.Reset()
		mockServer.OnGet("/rest/api/3/project", inttesting.JSONResponse(http.StatusOK, []map[string]interface{}{
			{"id": "10000", "key": "TEST", "name": "Test Project"},
			{"id": "10001", "key": "DEMO", "name": "Demo Project"},
		}))

		params := integration.JSONMap{
			"action": "list_projects",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ts.Execute(ctx, params)
		ts.AssertSuccess(result, err)
		ts.AssertLastRequestMethod("GET")

		t.Logf("✓ list_projects action verified")
	})
}

// TestJiraWorkflow_NetworkTimeout tests handling of network timeouts.
func TestJiraWorkflow_NetworkTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	jiraInteg := integrations.NewJiraIntegration(nil)
	ts := inttesting.NewTestSuite(t, jiraInteg)
	defer ts.Close()

	mockServer := ts.MockServer()

	t.Run("Timeout_SlowResponse", func(t *testing.T) {
		// Configure mock to delay response
		mockServer.OnGet("/rest/api/3/project/SLOW", inttesting.DelayedResponse(
			http.StatusOK,
			map[string]interface{}{
				"id":   "10000",
				"key":  "SLOW",
				"name": "Slow Project",
			},
			2*time.Second, // 2 second delay
		))

		config := &integration.Config{
			Name:    "test-jira",
			Type:    integration.TypeAPI,
			Enabled: true,
			Settings: integration.JSONMap{
				"base_url": mockServer.URL(),
			},
			Credentials: &integration.Credentials{
				Data: integration.JSONMap{
					"email":     "user@example.com",
					"api_token": "test-token",
				},
			},
		}
		ts.WithConfig(config)

		params := integration.JSONMap{
			"action":      "get_project",
			"project_key": "SLOW",
		}

		// Use context with timeout longer than the delay
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		start := time.Now()
		result, err := ts.Execute(ctx, params)
		duration := time.Since(start)

		ts.AssertSuccess(result, err)
		assert.GreaterOrEqual(t, duration, 2*time.Second, "should have taken at least 2s")

		t.Logf("✓ Slow response handled in %v", duration)
	})
}

// TestJiraWorkflow_ConfigurationValidation tests configuration validation.
func TestJiraWorkflow_ConfigurationValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	jiraInteg := integrations.NewJiraIntegration(nil)

	t.Run("Validate_MissingBaseURL", func(t *testing.T) {
		config := &integration.Config{
			Name:     "test-jira",
			Type:     integration.TypeAPI,
			Enabled:  true,
			Settings: integration.JSONMap{},
			Credentials: &integration.Credentials{
				Data: integration.JSONMap{
					"email":     "user@example.com",
					"api_token": "test-token",
				},
			},
		}

		err := jiraInteg.Validate(config)
		assert.Error(t, err, "should error on missing base_url")
		assert.Contains(t, err.Error(), "base_url")

		t.Logf("✓ Missing base_url validation")
	})

	t.Run("Validate_MissingCredentials", func(t *testing.T) {
		config := &integration.Config{
			Name:    "test-jira",
			Type:    integration.TypeAPI,
			Enabled: true,
			Settings: integration.JSONMap{
				"base_url": "https://example.atlassian.net",
			},
			Credentials: nil,
		}

		err := jiraInteg.Validate(config)
		assert.Error(t, err, "should error on missing credentials")

		t.Logf("✓ Missing credentials validation")
	})

	t.Run("Validate_ValidConfig", func(t *testing.T) {
		config := &integration.Config{
			Name:    "test-jira",
			Type:    integration.TypeAPI,
			Enabled: true,
			Settings: integration.JSONMap{
				"base_url": "https://example.atlassian.net",
			},
			Credentials: &integration.Credentials{
				Data: integration.JSONMap{
					"email":     "user@example.com",
					"api_token": "test-token",
				},
			},
		}

		err := jiraInteg.Validate(config)
		assert.NoError(t, err, "should not error on valid config")

		t.Logf("✓ Valid config validation")
	})

	t.Run("Validate_NilConfig", func(t *testing.T) {
		err := jiraInteg.Validate(nil)
		assert.Error(t, err, "should error on nil config")

		t.Logf("✓ Nil config validation")
	})
}
