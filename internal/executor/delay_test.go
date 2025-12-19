package executor

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/gorax/gorax/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteDelayAction_FixedDuration(t *testing.T) {
	tests := []struct {
		name             string
		duration         string
		expectedDuration time.Duration
		wantErr          bool
		errContains      string
	}{
		{
			name:             "delay 100ms",
			duration:         "100ms",
			expectedDuration: 100 * time.Millisecond,
			wantErr:          false,
		},
		{
			name:             "delay 1 second",
			duration:         "1s",
			expectedDuration: 1 * time.Second,
			wantErr:          false,
		},
		{
			name:             "delay 500ms",
			duration:         "500ms",
			expectedDuration: 500 * time.Millisecond,
			wantErr:          false,
		},
		{
			name:        "invalid duration format",
			duration:    "invalid",
			wantErr:     true,
			errContains: "invalid duration",
		},
		{
			name:        "empty duration",
			duration:    "",
			wantErr:     true,
			errContains: "duration is required",
		},
		{
			name:        "negative duration",
			duration:    "-5s",
			wantErr:     true,
			errContains: "duration must be positive",
		},
		{
			name:             "zero duration",
			duration:         "0s",
			expectedDuration: 0,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test node with delay configuration
			config := workflow.DelayConfig{
				Duration: tt.duration,
			}
			configJSON, err := json.Marshal(config)
			require.NoError(t, err)

			node := workflow.Node{
				ID:   "delay-node-1",
				Type: string(workflow.NodeTypeControlDelay),
				Data: workflow.NodeData{
					Name:   "Test Delay",
					Config: configJSON,
				},
			}

			// Create execution context
			execCtx := &ExecutionContext{
				TenantID:    "test-tenant",
				ExecutionID: "test-execution",
				WorkflowID:  "test-workflow",
				TriggerData: map[string]interface{}{},
				StepOutputs: map[string]interface{}{},
			}

			// Create executor with logger
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			executor := &Executor{
				logger: logger,
			}

			// Execute delay action
			ctx := context.Background()
			start := time.Now()
			output, err := executor.executeDelayAction(ctx, node, execCtx)
			elapsed := time.Since(start)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)

				// Verify the delay was approximately correct (within 50ms tolerance)
				tolerance := 50 * time.Millisecond
				assert.InDelta(t, tt.expectedDuration.Milliseconds(), elapsed.Milliseconds(), float64(tolerance.Milliseconds()))

				// Verify output structure
				result, ok := output.(map[string]interface{})
				require.True(t, ok, "output should be a map")
				assert.Equal(t, tt.duration, result["duration"])
				assert.NotNil(t, result["delayed_ms"])
			}
		})
	}
}

func TestExecuteDelayAction_VariableInterpolation(t *testing.T) {
	tests := []struct {
		name             string
		duration         string
		stepOutputs      map[string]interface{}
		expectedDuration time.Duration
		wantErr          bool
		errContains      string
	}{
		{
			name:     "interpolate from step output",
			duration: "{{steps.previous.delay}}",
			stepOutputs: map[string]interface{}{
				"previous": map[string]interface{}{
					"delay": "200ms",
				},
			},
			expectedDuration: 200 * time.Millisecond,
			wantErr:          false,
		},
		{
			name:     "interpolate with ${} syntax",
			duration: "${steps.config.duration}",
			stepOutputs: map[string]interface{}{
				"config": map[string]interface{}{
					"duration": "150ms",
				},
			},
			expectedDuration: 150 * time.Millisecond,
			wantErr:          false,
		},
		{
			name:     "interpolate missing variable",
			duration: "{{steps.missing.delay}}",
			stepOutputs: map[string]interface{}{
				"other": map[string]interface{}{},
			},
			wantErr:     true,
			errContains: "failed to interpolate duration",
		},
		{
			name:     "interpolate with nested path",
			duration: "{{steps.api.response.delay_seconds}}",
			stepOutputs: map[string]interface{}{
				"api": map[string]interface{}{
					"response": map[string]interface{}{
						"delay_seconds": "300ms",
					},
				},
			},
			expectedDuration: 300 * time.Millisecond,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test node with delay configuration
			config := workflow.DelayConfig{
				Duration: tt.duration,
			}
			configJSON, err := json.Marshal(config)
			require.NoError(t, err)

			node := workflow.Node{
				ID:   "delay-node-1",
				Type: string(workflow.NodeTypeControlDelay),
				Data: workflow.NodeData{
					Name:   "Test Delay",
					Config: configJSON,
				},
			}

			// Create execution context with step outputs
			execCtx := &ExecutionContext{
				TenantID:    "test-tenant",
				ExecutionID: "test-execution",
				WorkflowID:  "test-workflow",
				TriggerData: map[string]interface{}{},
				StepOutputs: tt.stepOutputs,
			}

			// Create executor with logger
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			executor := &Executor{
				logger: logger,
			}

			// Execute delay action
			ctx := context.Background()
			start := time.Now()
			output, err := executor.executeDelayAction(ctx, node, execCtx)
			elapsed := time.Since(start)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)

				// Verify the delay was approximately correct (within 50ms tolerance)
				tolerance := 50 * time.Millisecond
				assert.InDelta(t, tt.expectedDuration.Milliseconds(), elapsed.Milliseconds(), float64(tolerance.Milliseconds()))
			}
		})
	}
}

func TestExecuteDelayAction_ContextCancellation(t *testing.T) {
	// Create test node with 2 second delay
	config := workflow.DelayConfig{
		Duration: "2s",
	}
	configJSON, err := json.Marshal(config)
	require.NoError(t, err)

	node := workflow.Node{
		ID:   "delay-node-1",
		Type: string(workflow.NodeTypeControlDelay),
		Data: workflow.NodeData{
			Name:   "Test Delay",
			Config: configJSON,
		},
	}

	// Create execution context
	execCtx := &ExecutionContext{
		TenantID:    "test-tenant",
		ExecutionID: "test-execution",
		WorkflowID:  "test-workflow",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{},
	}

	// Create executor with logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		logger: logger,
	}

	// Create context with timeout that will cancel after 300ms
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Execute delay action
	start := time.Now()
	output, err := executor.executeDelayAction(ctx, node, execCtx)
	elapsed := time.Since(start)

	// Should return context error
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "context")

	// Should have stopped early (within 400ms, not the full 2s)
	assert.Less(t, elapsed, 500*time.Millisecond)
}

func TestExecuteDelayAction_MissingConfig(t *testing.T) {
	// Create test node without configuration
	node := workflow.Node{
		ID:   "delay-node-1",
		Type: string(workflow.NodeTypeControlDelay),
		Data: workflow.NodeData{
			Name:   "Test Delay",
			Config: json.RawMessage(`{}`),
		},
	}

	// Create execution context
	execCtx := &ExecutionContext{
		TenantID:    "test-tenant",
		ExecutionID: "test-execution",
		WorkflowID:  "test-workflow",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{},
	}

	// Create executor with logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		logger: logger,
	}

	// Execute delay action
	ctx := context.Background()
	output, err := executor.executeDelayAction(ctx, node, execCtx)

	// Should return error about missing duration
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "duration is required")
}

func TestExecuteDelayAction_InvalidJSON(t *testing.T) {
	// Create test node with invalid JSON configuration
	node := workflow.Node{
		ID:   "delay-node-1",
		Type: string(workflow.NodeTypeControlDelay),
		Data: workflow.NodeData{
			Name:   "Test Delay",
			Config: json.RawMessage(`{invalid json`),
		},
	}

	// Create execution context
	execCtx := &ExecutionContext{
		TenantID:    "test-tenant",
		ExecutionID: "test-execution",
		WorkflowID:  "test-workflow",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{},
	}

	// Create executor with logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		logger: logger,
	}

	// Execute delay action
	ctx := context.Background()
	output, err := executor.executeDelayAction(ctx, node, execCtx)

	// Should return parse error
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestExecuteDelayAction_BroadcastEvents(t *testing.T) {
	// This test would require a mock broadcaster
	// For now, we'll test that the function works without a broadcaster
	config := workflow.DelayConfig{
		Duration: "100ms",
	}
	configJSON, err := json.Marshal(config)
	require.NoError(t, err)

	node := workflow.Node{
		ID:   "delay-node-1",
		Type: string(workflow.NodeTypeControlDelay),
		Data: workflow.NodeData{
			Name:   "Test Delay",
			Config: configJSON,
		},
	}

	execCtx := &ExecutionContext{
		TenantID:    "test-tenant",
		ExecutionID: "test-execution",
		WorkflowID:  "test-workflow",
		TriggerData: map[string]interface{}{},
		StepOutputs: map[string]interface{}{},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	executor := &Executor{
		logger:      logger,
		broadcaster: nil, // No broadcaster
	}

	ctx := context.Background()
	output, err := executor.executeDelayAction(ctx, node, execCtx)

	// Should work fine without broadcaster
	assert.NoError(t, err)
	assert.NotNil(t, output)
}
