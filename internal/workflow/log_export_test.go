package workflow

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogExporter_ExportTXT(t *testing.T) {
	exporter := NewLogExporter()
	now := time.Now()

	execution := &Execution{
		ID:         "exec-123",
		WorkflowID: "wf-456",
		Status:     "completed",
		StartedAt:  &now,
	}

	steps := []*StepExecution{
		{
			ID:          "step-1",
			NodeID:      "node-1",
			NodeType:    "action:http",
			Status:      "completed",
			StartedAt:   &now,
			CompletedAt: ptrTimeLog(now.Add(2 * time.Second)),
			DurationMs:  ptrIntLog(2000),
			InputData:   ptrRawMessageLog(`{"url":"https://api.example.com"}`),
			OutputData:  ptrRawMessageLog(`{"status":200,"body":"OK"}`),
		},
		{
			ID:           "step-2",
			NodeID:       "node-2",
			NodeType:     "action:transform",
			Status:       "failed",
			StartedAt:    &now,
			CompletedAt:  ptrTimeLog(now.Add(1 * time.Second)),
			DurationMs:   ptrIntLog(1000),
			ErrorMessage: ptrStringLog("transformation failed: invalid expression"),
		},
	}

	t.Run("generates valid TXT format", func(t *testing.T) {
		result := exporter.ExportTXT(execution, steps)
		output := string(result)

		assert.Contains(t, output, "Execution ID: exec-123")
		assert.Contains(t, output, "Workflow ID: wf-456")
		assert.Contains(t, output, "Status: completed")
		assert.Contains(t, output, "Step 1: step-1")
		assert.Contains(t, output, "Node ID: node-1")
		assert.Contains(t, output, "Type: action:http")
		assert.Contains(t, output, "Status: completed")
		assert.Contains(t, output, "Duration: 2000ms")
		assert.Contains(t, output, "Step 2: step-2")
		assert.Contains(t, output, "Error: transformation failed: invalid expression")
	})

	t.Run("handles empty steps", func(t *testing.T) {
		result := exporter.ExportTXT(execution, []*StepExecution{})
		output := string(result)

		assert.Contains(t, output, "Execution ID: exec-123")
		assert.Contains(t, output, "No steps executed")
	})

	t.Run("handles nil timestamps", func(t *testing.T) {
		execWithNilTime := &Execution{
			ID:         "exec-999",
			WorkflowID: "wf-999",
			Status:     "pending",
		}
		result := exporter.ExportTXT(execWithNilTime, []*StepExecution{})
		output := string(result)

		assert.Contains(t, output, "Execution ID: exec-999")
		assert.Contains(t, output, "Status: pending")
	})
}

func TestLogExporter_ExportJSON(t *testing.T) {
	exporter := NewLogExporter()
	now := time.Now()

	execution := &Execution{
		ID:         "exec-123",
		WorkflowID: "wf-456",
		Status:     "completed",
		StartedAt:  &now,
	}

	steps := []*StepExecution{
		{
			ID:         "step-1",
			NodeID:     "node-1",
			NodeType:   "action:http",
			Status:     "completed",
			StartedAt:  &now,
			DurationMs: ptrIntLog(2000),
			OutputData: ptrRawMessageLog(`{"status":200}`),
		},
	}

	t.Run("generates valid JSON format", func(t *testing.T) {
		result := exporter.ExportJSON(execution, steps)

		var parsed map[string]interface{}
		err := json.Unmarshal(result, &parsed)
		require.NoError(t, err)

		assert.Equal(t, "exec-123", parsed["execution_id"])
		assert.Equal(t, "wf-456", parsed["workflow_id"])
		assert.Equal(t, "completed", parsed["status"])

		stepsData := parsed["steps"].([]interface{})
		require.Len(t, stepsData, 1)

		step := stepsData[0].(map[string]interface{})
		assert.Equal(t, "step-1", step["step_id"])
		assert.Equal(t, "node-1", step["node_id"])
		assert.Equal(t, "action:http", step["node_type"])
		assert.Equal(t, "completed", step["status"])
		assert.Equal(t, float64(2000), step["duration_ms"])
	})

	t.Run("handles empty steps", func(t *testing.T) {
		result := exporter.ExportJSON(execution, []*StepExecution{})

		var parsed map[string]interface{}
		err := json.Unmarshal(result, &parsed)
		require.NoError(t, err)

		stepsData := parsed["steps"].([]interface{})
		assert.Len(t, stepsData, 0)
	})

	t.Run("includes error messages", func(t *testing.T) {
		stepsWithError := []*StepExecution{
			{
				ID:           "step-err",
				NodeID:       "node-err",
				NodeType:     "action:fail",
				Status:       "failed",
				ErrorMessage: ptrStringLog("connection timeout"),
			},
		}

		result := exporter.ExportJSON(execution, stepsWithError)

		var parsed map[string]interface{}
		err := json.Unmarshal(result, &parsed)
		require.NoError(t, err)

		stepsData := parsed["steps"].([]interface{})
		step := stepsData[0].(map[string]interface{})
		assert.Equal(t, "connection timeout", step["error_message"])
	})
}

func TestLogExporter_ExportCSV(t *testing.T) {
	exporter := NewLogExporter()
	now := time.Now()

	execution := &Execution{
		ID:         "exec-123",
		WorkflowID: "wf-456",
		Status:     "completed",
		StartedAt:  &now,
	}

	steps := []*StepExecution{
		{
			ID:         "step-1",
			NodeID:     "node-1",
			NodeType:   "action:http",
			Status:     "completed",
			StartedAt:  &now,
			DurationMs: ptrIntLog(2000),
		},
		{
			ID:           "step-2",
			NodeID:       "node-2",
			NodeType:     "action:transform",
			Status:       "failed",
			ErrorMessage: ptrStringLog("parse error"),
		},
	}

	t.Run("generates valid CSV format", func(t *testing.T) {
		result := exporter.ExportCSV(execution, steps)
		output := string(result)

		reader := csv.NewReader(strings.NewReader(output))
		records, err := reader.ReadAll()
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(records), 3) // header + 2 rows

		// Check header
		header := records[0]
		assert.Contains(t, header, "step_id")
		assert.Contains(t, header, "node_id")
		assert.Contains(t, header, "node_type")
		assert.Contains(t, header, "status")
		assert.Contains(t, header, "duration_ms")
		assert.Contains(t, header, "error_message")

		// Check first data row
		row1 := records[1]
		assert.Contains(t, row1, "step-1")
		assert.Contains(t, row1, "node-1")
		assert.Contains(t, row1, "action:http")
		assert.Contains(t, row1, "completed")
		assert.Contains(t, row1, "2000")

		// Check second data row
		row2 := records[2]
		assert.Contains(t, row2, "step-2")
		assert.Contains(t, row2, "failed")
		assert.Contains(t, row2, "parse error")
	})

	t.Run("handles empty steps", func(t *testing.T) {
		result := exporter.ExportCSV(execution, []*StepExecution{})
		output := string(result)

		reader := csv.NewReader(strings.NewReader(output))
		records, err := reader.ReadAll()
		require.NoError(t, err)
		require.Len(t, records, 1) // header only
	})

	t.Run("escapes CSV special characters", func(t *testing.T) {
		stepsWithSpecialChars := []*StepExecution{
			{
				ID:           "step-1",
				NodeID:       "node-1",
				NodeType:     "action:test",
				Status:       "failed",
				ErrorMessage: ptrStringLog(`Error: "quoted", with, commas`),
			},
		}

		result := exporter.ExportCSV(execution, stepsWithSpecialChars)
		output := string(result)

		reader := csv.NewReader(strings.NewReader(output))
		records, err := reader.ReadAll()
		require.NoError(t, err)
		require.Len(t, records, 2) // header + 1 row

		// CSV should properly escape quotes and commas
		assert.Contains(t, records[1], `Error: "quoted", with, commas`)
	})
}

func TestLogExporter_FormatSizes(t *testing.T) {
	exporter := NewLogExporter()
	now := time.Now()

	execution := &Execution{
		ID:         "exec-large",
		WorkflowID: "wf-large",
		Status:     "completed",
		StartedAt:  &now,
	}

	// Create large dataset
	steps := make([]*StepExecution, 100)
	for i := 0; i < 100; i++ {
		steps[i] = &StepExecution{
			ID:         fmt.Sprintf("step-%d", i),
			NodeID:     fmt.Sprintf("node-%d", i),
			NodeType:   "action:http",
			Status:     "completed",
			StartedAt:  &now,
			DurationMs: ptrIntLog(100 + i*10),
			OutputData: ptrRawMessageLog(fmt.Sprintf(`{"data":"some output data for step %d"}`, i)),
		}
	}

	t.Run("TXT format handles large datasets", func(t *testing.T) {
		result := exporter.ExportTXT(execution, steps)
		assert.Greater(t, len(result), 1000)
		assert.Contains(t, string(result), "step-0")
	})

	t.Run("JSON format handles large datasets", func(t *testing.T) {
		result := exporter.ExportJSON(execution, steps)
		assert.Greater(t, len(result), 1000)

		var parsed map[string]interface{}
		err := json.Unmarshal(result, &parsed)
		require.NoError(t, err)
	})

	t.Run("CSV format handles large datasets", func(t *testing.T) {
		result := exporter.ExportCSV(execution, steps)
		assert.Greater(t, len(result), 1000)

		reader := csv.NewReader(strings.NewReader(string(result)))
		records, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Len(t, records, 101) // header + 100 rows
	})
}

// Helper functions for tests
func ptrTimeLog(t time.Time) *time.Time {
	return &t
}

func ptrIntLog(i int) *int {
	return &i
}

func ptrStringLog(s string) *string {
	return &s
}

func ptrRawMessageLog(s string) *json.RawMessage {
	raw := json.RawMessage(s)
	return &raw
}
