package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Schedule represents the schedule model for tests
type Schedule struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenant_id"`
	WorkflowID     string `json:"workflow_id"`
	Name           string `json:"name"`
	CronExpression string `json:"cron_expression"`
	Timezone       string `json:"timezone"`
	Enabled        bool   `json:"enabled"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ScheduleListResponse represents the list response
type ScheduleListResponse struct {
	Data   []Schedule `json:"data"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
	Total  int        `json:"total"`
}

// ParseCronResponse represents the cron parsing response
type ParseCronResponse struct {
	Valid   bool   `json:"valid"`
	NextRun string `json:"next_run,omitempty"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func TestScheduleAPI_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)

	// Create test tenant and workflow
	tenantID := ts.CreateTestTenant(t, "schedule-test-tenant")
	workflowID := createTestWorkflow(t, ts, tenantID)
	headers := DefaultTestHeaders(tenantID)

	t.Run("creates schedule with valid input", func(t *testing.T) {
		input := map[string]interface{}{
			"name":            "Daily Report",
			"cron_expression": "0 9 * * *",
			"timezone":        "UTC",
			"enabled":         true,
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/schedules", input, headers)
		AssertStatusCode(t, resp, http.StatusCreated)

		var schedule Schedule
		ParseJSONResponse(t, resp, &schedule)

		assert.NotEmpty(t, schedule.ID)
		assert.Equal(t, "Daily Report", schedule.Name)
		assert.Equal(t, "0 9 * * *", schedule.CronExpression)
		assert.Equal(t, "UTC", schedule.Timezone)
		assert.True(t, schedule.Enabled)
	})

	t.Run("creates schedule with different timezone", func(t *testing.T) {
		input := map[string]interface{}{
			"name":            "Weekly Backup",
			"cron_expression": "0 0 * * 0",
			"timezone":        "America/New_York",
			"enabled":         false,
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/schedules", input, headers)
		AssertStatusCode(t, resp, http.StatusCreated)

		var schedule Schedule
		ParseJSONResponse(t, resp, &schedule)

		assert.Equal(t, "America/New_York", schedule.Timezone)
		assert.False(t, schedule.Enabled)
	})

	t.Run("returns error for invalid cron expression", func(t *testing.T) {
		input := map[string]interface{}{
			"name":            "Invalid Schedule",
			"cron_expression": "invalid",
			"timezone":        "UTC",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/schedules", input, headers)
		AssertStatusCode(t, resp, http.StatusBadRequest)
	})

	t.Run("returns error for missing name", func(t *testing.T) {
		input := map[string]interface{}{
			"cron_expression": "0 9 * * *",
			"timezone":        "UTC",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/schedules", input, headers)
		AssertStatusCode(t, resp, http.StatusBadRequest)
	})

	t.Run("returns error for missing cron expression", func(t *testing.T) {
		input := map[string]interface{}{
			"name":     "No Cron",
			"timezone": "UTC",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/schedules", input, headers)
		AssertStatusCode(t, resp, http.StatusBadRequest)
	})
}

func TestScheduleAPI_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)

	// Create test tenant and workflow
	tenantID := ts.CreateTestTenant(t, "schedule-list-tenant")
	workflowID := createTestWorkflow(t, ts, tenantID)
	headers := DefaultTestHeaders(tenantID)

	// Create multiple schedules
	for i := 0; i < 3; i++ {
		input := map[string]interface{}{
			"name":            "Schedule " + string(rune('A'+i)),
			"cron_expression": "0 " + string(rune('0'+i)) + " * * *",
			"timezone":        "UTC",
			"enabled":         true,
		}
		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/schedules", input, headers)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}

	t.Run("lists schedules for workflow", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/workflows/"+workflowID+"/schedules", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var listResp ScheduleListResponse
		ParseJSONResponse(t, resp, &listResp)

		assert.Equal(t, 3, len(listResp.Data))
	})

	t.Run("supports pagination", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/workflows/"+workflowID+"/schedules?limit=2&offset=0", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var listResp ScheduleListResponse
		ParseJSONResponse(t, resp, &listResp)

		assert.Equal(t, 2, len(listResp.Data))
		assert.Equal(t, 2, listResp.Limit)
		assert.Equal(t, 0, listResp.Offset)
	})

	t.Run("lists all schedules for tenant", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/schedules", nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var listResp ScheduleListResponse
		ParseJSONResponse(t, resp, &listResp)

		assert.GreaterOrEqual(t, len(listResp.Data), 3)
	})
}

func TestScheduleAPI_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)

	// Create test tenant, workflow, and schedule
	tenantID := ts.CreateTestTenant(t, "schedule-get-tenant")
	workflowID := createTestWorkflow(t, ts, tenantID)
	headers := DefaultTestHeaders(tenantID)

	input := map[string]interface{}{
		"name":            "Test Schedule",
		"cron_expression": "0 12 * * *",
		"timezone":        "Europe/London",
		"enabled":         true,
	}
	createResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/schedules", input, headers)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created Schedule
	ParseJSONResponse(t, createResp, &created)

	t.Run("gets schedule by ID", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/schedules/"+created.ID, nil, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var schedule Schedule
		ParseJSONResponse(t, resp, &schedule)

		assert.Equal(t, created.ID, schedule.ID)
		assert.Equal(t, "Test Schedule", schedule.Name)
		assert.Equal(t, "0 12 * * *", schedule.CronExpression)
		assert.Equal(t, "Europe/London", schedule.Timezone)
	})

	t.Run("returns 404 for non-existent schedule", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/schedules/00000000-0000-0000-0000-000000000000", nil, headers)
		AssertStatusCode(t, resp, http.StatusNotFound)
	})
}

func TestScheduleAPI_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)

	// Create test tenant, workflow, and schedule
	tenantID := ts.CreateTestTenant(t, "schedule-update-tenant")
	workflowID := createTestWorkflow(t, ts, tenantID)
	headers := DefaultTestHeaders(tenantID)

	input := map[string]interface{}{
		"name":            "Original Schedule",
		"cron_expression": "0 9 * * *",
		"timezone":        "UTC",
		"enabled":         true,
	}
	createResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/schedules", input, headers)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created Schedule
	ParseJSONResponse(t, createResp, &created)

	t.Run("updates schedule name", func(t *testing.T) {
		updateInput := map[string]interface{}{
			"name": "Updated Schedule",
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/schedules/"+created.ID, updateInput, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var updated Schedule
		ParseJSONResponse(t, resp, &updated)

		assert.Equal(t, "Updated Schedule", updated.Name)
		assert.Equal(t, "0 9 * * *", updated.CronExpression) // Unchanged
	})

	t.Run("updates cron expression", func(t *testing.T) {
		updateInput := map[string]interface{}{
			"cron_expression": "0 18 * * *",
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/schedules/"+created.ID, updateInput, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var updated Schedule
		ParseJSONResponse(t, resp, &updated)

		assert.Equal(t, "0 18 * * *", updated.CronExpression)
	})

	t.Run("toggles enabled status", func(t *testing.T) {
		updateInput := map[string]interface{}{
			"enabled": false,
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/schedules/"+created.ID, updateInput, headers)
		AssertStatusCode(t, resp, http.StatusOK)

		var updated Schedule
		ParseJSONResponse(t, resp, &updated)

		assert.False(t, updated.Enabled)
	})

	t.Run("returns error for invalid cron expression update", func(t *testing.T) {
		updateInput := map[string]interface{}{
			"cron_expression": "not-a-cron",
		}

		resp := ts.MakeRequest(t, http.MethodPut, "/api/v1/schedules/"+created.ID, updateInput, headers)
		AssertStatusCode(t, resp, http.StatusBadRequest)
	})
}

func TestScheduleAPI_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)

	// Create test tenant, workflow, and schedule
	tenantID := ts.CreateTestTenant(t, "schedule-delete-tenant")
	workflowID := createTestWorkflow(t, ts, tenantID)
	headers := DefaultTestHeaders(tenantID)

	input := map[string]interface{}{
		"name":            "To Delete",
		"cron_expression": "0 9 * * *",
		"timezone":        "UTC",
		"enabled":         true,
	}
	createResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/"+workflowID+"/schedules", input, headers)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created Schedule
	ParseJSONResponse(t, createResp, &created)

	t.Run("deletes schedule", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/schedules/"+created.ID, nil, headers)
		AssertStatusCode(t, resp, http.StatusNoContent)

		// Verify it's deleted
		getResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/schedules/"+created.ID, nil, headers)
		AssertStatusCode(t, getResp, http.StatusNotFound)
	})

	t.Run("returns 404 for deleting non-existent schedule", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodDelete, "/api/v1/schedules/00000000-0000-0000-0000-000000000000", nil, headers)
		AssertStatusCode(t, resp, http.StatusNotFound)
	})
}

func TestScheduleAPI_ParseCron_Valid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "cron-parse-tenant")
	headers := DefaultTestHeaders(tenantID)

	validCrons := []struct {
		name string
		cron string
	}{
		{"every minute", "* * * * *"},
		{"every hour", "0 * * * *"},
		{"every day at 9am", "0 9 * * *"},
		{"every monday", "0 9 * * 1"},
		{"every 5 minutes", "*/5 * * * *"},
		{"weekdays at 9am", "0 9 * * 1-5"},
		{"first of month", "0 0 1 * *"},
	}

	for _, tc := range validCrons {
		t.Run(tc.name, func(t *testing.T) {
			input := map[string]interface{}{
				"cron_expression": tc.cron,
				"timezone":        "UTC",
			}

			resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/schedules/parse-cron", input, headers)
			AssertStatusCode(t, resp, http.StatusOK)

			var parseResp ParseCronResponse
			ParseJSONResponse(t, resp, &parseResp)

			assert.True(t, parseResp.Valid)
			assert.NotEmpty(t, parseResp.NextRun)
		})
	}
}

func TestScheduleAPI_ParseCron_Invalid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "cron-invalid-tenant")
	headers := DefaultTestHeaders(tenantID)

	invalidCrons := []struct {
		name string
		cron string
	}{
		{"empty string", ""},
		{"random text", "not a cron"},
		{"too few fields", "* * *"},
		{"too many fields", "* * * * * * *"},
		{"invalid minute", "60 * * * *"},
		{"invalid hour", "0 24 * * *"},
		{"invalid day of month", "0 0 32 * *"},
		{"invalid month", "0 0 * 13 *"},
		{"invalid day of week", "0 0 * * 8"},
	}

	for _, tc := range invalidCrons {
		t.Run(tc.name, func(t *testing.T) {
			input := map[string]interface{}{
				"cron_expression": tc.cron,
				"timezone":        "UTC",
			}

			resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/schedules/parse-cron", input, headers)
			AssertStatusCode(t, resp, http.StatusBadRequest)
		})
	}
}

func TestScheduleAPI_ErrorResponses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ts := SetupTestServer(t)
	tenantID := ts.CreateTestTenant(t, "schedule-error-tenant")
	headers := DefaultTestHeaders(tenantID)

	t.Run("returns 401 without auth headers", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/schedules", nil, nil)
		// Depending on auth setup, this might be 401 or 403
		assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusBadRequest)
		resp.Body.Close()
	})

	t.Run("returns 404 for non-existent workflow", func(t *testing.T) {
		input := map[string]interface{}{
			"name":            "Orphan Schedule",
			"cron_expression": "0 9 * * *",
			"timezone":        "UTC",
		}

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/workflows/00000000-0000-0000-0000-000000000000/schedules", input, headers)
		AssertStatusCode(t, resp, http.StatusNotFound)
	})

	t.Run("returns proper error format", func(t *testing.T) {
		resp := ts.MakeRequest(t, http.MethodGet, "/api/v1/schedules/invalid-uuid", nil, headers)
		AssertStatusCode(t, resp, http.StatusBadRequest)

		var errResp ErrorResponse
		ParseJSONResponse(t, resp, &errResp)

		assert.NotEmpty(t, errResp.Error)
		assert.NotEmpty(t, errResp.Code)
	})
}

// createTestWorkflow creates a test workflow and returns its ID
func createTestWorkflow(t *testing.T, ts *TestServer, tenantID string) string {
	t.Helper()

	query := `
		INSERT INTO workflows (id, tenant_id, name, description, version, status, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, 'Test Workflow', 'A test workflow', 1, 'draft', NOW(), NOW())
		RETURNING id
	`

	var workflowID string
	err := ts.DB.QueryRow(query, tenantID).Scan(&workflowID)
	require.NoError(t, err, "failed to create test workflow")

	t.Logf("Created test workflow: %s", workflowID)
	return workflowID
}
