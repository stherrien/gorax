package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/workflow"
	"github.com/gorax/gorax/internal/workflow/formula"
)

// TestFormulaCacheIntegration tests that the formula cache is properly integrated
// into workflow execution and provides performance benefits
func TestFormulaCacheIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup mock repository
	repo := &mockWorkflowRepository{
		workflows:  make(map[string]*workflow.Workflow),
		executions: make(map[string]*workflow.Execution),
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	tenantID := "test-tenant"

	// Create a workflow with a formula action
	workflowDef := workflow.WorkflowDefinition{
		Nodes: []workflow.Node{
			{
				ID:   "trigger-1",
				Type: string(workflow.NodeTypeTriggerWebhook),
				Data: workflow.NodeData{
					Config: json.RawMessage(`{}`),
				},
			},
			{
				ID:   "formula-1",
				Type: string(workflow.NodeTypeActionFormula),
				Data: workflow.NodeData{
					Config: json.RawMessage(`{
						"expression": "trigger.value * 2 + 10",
						"output_variable": "result"
					}`),
				},
			},
		},
		Edges: []workflow.Edge{
			{
				ID:     "edge-1",
				Source: "trigger-1",
				Target: "formula-1",
			},
		},
	}

	defJSON, err := json.Marshal(workflowDef)
	require.NoError(t, err)

	wf := &workflow.Workflow{
		ID:         "test-workflow-1",
		TenantID:   tenantID,
		Name:       "Formula Cache Test",
		Definition: defJSON,
		Version:    1,
	}
	repo.workflows[wf.ID] = wf

	t.Run("executor with cached evaluator improves performance", func(t *testing.T) {
		ctx := context.Background()

		// Create executor with cached evaluator
		cachedEvaluator := formula.NewCachedEvaluator(1000)
		executor := NewWithCachedEvaluator(repo, logger, nil, cachedEvaluator)

		// Execute the workflow multiple times with the same formula
		iterations := 10
		for i := 0; i < iterations; i++ {
			triggerData := json.RawMessage(`{"value": 42}`)
			execution := &workflow.Execution{
				ID:          string(rune('a' + i)),
				TenantID:    tenantID,
				WorkflowID:  wf.ID,
				Status:      "pending",
				TriggerData: &triggerData,
			}
			repo.executions[execution.ID] = execution

			err = executor.Execute(ctx, execution)
			require.NoError(t, err)

			// Verify execution completed successfully
			assert.Equal(t, "completed", execution.Status)
		}

		// Verify cache was used effectively
		stats := cachedEvaluator.CacheStats()

		// After first execution (miss), subsequent executions should hit cache
		assert.Greater(t, stats.Hits, uint64(iterations-2),
			"Expected cache hits after first compilation")

		// Calculate and verify hit rate
		hitRate := stats.HitRate()
		assert.Greater(t, hitRate, 0.5,
			"Expected hit rate > 50%% (got %.2f%%)", hitRate*100)

		t.Logf("Cache Stats: Hits=%d, Misses=%d, HitRate=%.2f%%",
			stats.Hits, stats.Misses, hitRate*100)
	})

	t.Run("cached evaluator provides correct results", func(t *testing.T) {
		ctx := context.Background()
		cachedEvaluator := formula.NewCachedEvaluator(1000)
		executor := NewWithCachedEvaluator(repo, logger, nil, cachedEvaluator)

		// Test different input values produce correct outputs
		testCases := []struct {
			input    int
			expected float64 // value * 2 + 10
		}{
			{input: 5, expected: 20},
			{input: 10, expected: 30},
			{input: 20, expected: 50},
		}

		for idx, tc := range testCases {
			triggerDataStr := fmt.Sprintf(`{"value": %d}`, tc.input)
			triggerData := json.RawMessage(triggerDataStr)
			execution := &workflow.Execution{
				ID:          string(rune('A' + idx)),
				TenantID:    tenantID,
				WorkflowID:  wf.ID,
				Status:      "pending",
				TriggerData: &triggerData,
			}
			repo.executions[execution.ID] = execution

			err = executor.Execute(ctx, execution)
			require.NoError(t, err)

			// Verify output is correct
			var outputData map[string]interface{}
			if execution.OutputData != nil {
				err = json.Unmarshal(*execution.OutputData, &outputData)
				require.NoError(t, err)
			}

			formulaResult, ok := outputData["formula-1"]
			require.True(t, ok, "Formula output not found")

			resultFloat, ok := formulaResult.(float64)
			require.True(t, ok, "Formula result is not a number")
			assert.Equal(t, tc.expected, resultFloat)
		}
	})

	t.Run("backward compatibility - executor without cache still works", func(t *testing.T) {
		ctx := context.Background()
		// Create executor without cached evaluator (legacy behavior)
		executor := NewWithCachedEvaluator(repo, logger, nil, nil)

		triggerData := json.RawMessage(`{"value": 25}`)
		execution := &workflow.Execution{
			ID:          "legacy-exec",
			TenantID:    tenantID,
			WorkflowID:  wf.ID,
			Status:      "pending",
			TriggerData: &triggerData,
		}
		repo.executions[execution.ID] = execution

		err = executor.Execute(ctx, execution)
		require.NoError(t, err)

		// Verify execution completed successfully
		assert.Equal(t, "completed", execution.Status)

		// Verify output is correct (25 * 2 + 10 = 60)
		var outputData map[string]interface{}
		if execution.OutputData != nil {
			err = json.Unmarshal(*execution.OutputData, &outputData)
			require.NoError(t, err)
		}

		formulaResult := outputData["formula-1"]
		assert.Equal(t, float64(60), formulaResult)
	})
}

// Mock workflow repository for testing
type mockWorkflowRepository struct {
	workflows  map[string]*workflow.Workflow
	executions map[string]*workflow.Execution
}

func (m *mockWorkflowRepository) GetByID(ctx context.Context, tenantID, id string) (*workflow.Workflow, error) {
	if wf, ok := m.workflows[id]; ok {
		return wf, nil
	}
	return nil, fmt.Errorf("workflow not found")
}

func (m *mockWorkflowRepository) UpdateExecutionStatus(ctx context.Context, id string, status string, outputData json.RawMessage, errorMsg *string) error {
	if exec, ok := m.executions[id]; ok {
		exec.Status = status
		exec.OutputData = &outputData
		if errorMsg != nil {
			exec.ErrorMessage = errorMsg
		}
		return nil
	}
	return fmt.Errorf("execution not found")
}

func (m *mockWorkflowRepository) CreateStepExecution(ctx context.Context, executionID, nodeID, nodeType string, inputData []byte) (*workflow.StepExecution, error) {
	rawMsg := json.RawMessage(inputData)
	return &workflow.StepExecution{
		ID:          nodeID + "-step",
		ExecutionID: executionID,
		NodeID:      nodeID,
		NodeType:    nodeType,
		Status:      "pending",
		InputData:   &rawMsg,
	}, nil
}

func (m *mockWorkflowRepository) UpdateStepExecution(ctx context.Context, id, status string, outputData json.RawMessage, errorMsg *string) error {
	// No-op for now
	return nil
}
