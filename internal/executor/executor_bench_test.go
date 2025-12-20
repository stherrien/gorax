package executor

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/gorax/gorax/internal/workflow"
)

// BenchmarkExecuteSimpleWorkflow benchmarks execution of a simple workflow with one step
func BenchmarkExecuteSimpleWorkflow(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := &workflow.Repository{} // Mock repository
	exec := New(repo, logger)

	wf := &workflow.Workflow{
		ID:   "bench-simple",
		Name: "Simple Benchmark Workflow",
		Nodes: []workflow.Node{
			{
				ID:   "step1",
				Type: "transform",
				Config: map[string]interface{}{
					"formula": "trigger.value * 2",
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	triggerData := map[string]interface{}{
		"value": 42,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		execCtx := &ExecutionContext{
			TenantID:    "bench-tenant",
			ExecutionID: "exec-1",
			WorkflowID:  wf.ID,
			TriggerType: "manual",
			TriggerData: triggerData,
			StepOutputs: make(map[string]interface{}),
		}

		_, _ = exec.executeNode(ctx, wf, &wf.Nodes[0], execCtx)
	}
}

// BenchmarkExecuteWorkflowWithRetry benchmarks workflow execution with retry logic
func BenchmarkExecuteWorkflowWithRetry(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := &workflow.Repository{}
	exec := New(repo, logger)

	wf := &workflow.Workflow{
		ID:   "bench-retry",
		Name: "Retry Benchmark Workflow",
		Nodes: []workflow.Node{
			{
				ID:   "step1",
				Type: "transform",
				Config: map[string]interface{}{
					"formula": "trigger.value * 2",
				},
				Retry: &workflow.NodeRetryConfig{
					Enabled:    true,
					MaxRetries: 3,
					Delay:      "100ms",
					Backoff:    "exponential",
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	triggerData := map[string]interface{}{
		"value": 42,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		execCtx := &ExecutionContext{
			TenantID:    "bench-tenant",
			ExecutionID: "exec-1",
			WorkflowID:  wf.ID,
			TriggerType: "manual",
			TriggerData: triggerData,
			StepOutputs: make(map[string]interface{}),
		}

		_, _ = exec.executeNode(ctx, wf, &wf.Nodes[0], execCtx)
	}
}

// BenchmarkExecuteSequentialWorkflow benchmarks sequential workflow execution
func BenchmarkExecuteSequentialWorkflow(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := &workflow.Repository{}
	exec := New(repo, logger)

	wf := &workflow.Workflow{
		ID:   "bench-sequential",
		Name: "Sequential Benchmark Workflow",
		Nodes: []workflow.Node{
			{
				ID:   "step1",
				Type: "transform",
				Config: map[string]interface{}{
					"formula": "trigger.value * 2",
				},
			},
			{
				ID:   "step2",
				Type: "transform",
				Config: map[string]interface{}{
					"formula": "step1 + 10",
				},
			},
			{
				ID:   "step3",
				Type: "transform",
				Config: map[string]interface{}{
					"formula": "step2 * 3",
				},
			},
		},
		Edges: []workflow.Edge{
			{From: "step1", To: "step2"},
			{From: "step2", To: "step3"},
		},
	}

	triggerData := map[string]interface{}{
		"value": 42,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		execCtx := &ExecutionContext{
			TenantID:    "bench-tenant",
			ExecutionID: "exec-1",
			WorkflowID:  wf.ID,
			TriggerType: "manual",
			TriggerData: triggerData,
			StepOutputs: make(map[string]interface{}),
		}

		for _, node := range wf.Nodes {
			result, _ := exec.executeNode(ctx, wf, &node, execCtx)
			if result != nil {
				execCtx.StepOutputs[node.ID] = result
			}
		}
	}
}

// BenchmarkExecuteConditionalWorkflow benchmarks conditional workflow execution
func BenchmarkExecuteConditionalWorkflow(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := &workflow.Repository{}
	exec := New(repo, logger)

	wf := &workflow.Workflow{
		ID:   "bench-conditional",
		Name: "Conditional Benchmark Workflow",
		Nodes: []workflow.Node{
			{
				ID:   "step1",
				Type: "conditional",
				Config: map[string]interface{}{
					"condition": "trigger.value > 50",
					"if": map[string]interface{}{
						"type": "transform",
						"config": map[string]interface{}{
							"formula": "trigger.value * 2",
						},
					},
					"else": map[string]interface{}{
						"type": "transform",
						"config": map[string]interface{}{
							"formula": "trigger.value + 10",
						},
					},
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	triggerData := map[string]interface{}{
		"value": 42,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		execCtx := &ExecutionContext{
			TenantID:    "bench-tenant",
			ExecutionID: "exec-1",
			WorkflowID:  wf.ID,
			TriggerType: "manual",
			TriggerData: triggerData,
			StepOutputs: make(map[string]interface{}),
		}

		_, _ = exec.executeConditionalNode(ctx, wf, &wf.Nodes[0], execCtx)
	}
}

// BenchmarkExecuteLoopWorkflow benchmarks loop workflow execution
func BenchmarkExecuteLoopWorkflow(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := &workflow.Repository{}
	exec := New(repo, logger)

	wf := &workflow.Workflow{
		ID:   "bench-loop",
		Name: "Loop Benchmark Workflow",
		Nodes: []workflow.Node{
			{
				ID:   "loop1",
				Type: "loop",
				Config: map[string]interface{}{
					"items": []interface{}{1, 2, 3, 4, 5},
					"steps": []interface{}{
						map[string]interface{}{
							"id":   "transform",
							"type": "transform",
							"config": map[string]interface{}{
								"formula": "item * 2",
							},
						},
					},
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		execCtx := &ExecutionContext{
			TenantID:    "bench-tenant",
			ExecutionID: "exec-1",
			WorkflowID:  wf.ID,
			TriggerType: "manual",
			TriggerData: make(map[string]interface{}),
			StepOutputs: make(map[string]interface{}),
		}

		_, _ = exec.executeLoopNode(ctx, wf, &wf.Nodes[0], execCtx)
	}
}

// BenchmarkExecuteParallelWorkflow benchmarks parallel workflow execution
func BenchmarkExecuteParallelWorkflow(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := &workflow.Repository{}
	exec := New(repo, logger)

	wf := &workflow.Workflow{
		ID:   "bench-parallel",
		Name: "Parallel Benchmark Workflow",
		Nodes: []workflow.Node{
			{
				ID:   "parallel1",
				Type: "parallel",
				Config: map[string]interface{}{
					"branches": []interface{}{
						map[string]interface{}{
							"id":   "branch1",
							"type": "transform",
							"config": map[string]interface{}{
								"formula": "trigger.value * 2",
							},
						},
						map[string]interface{}{
							"id":   "branch2",
							"type": "transform",
							"config": map[string]interface{}{
								"formula": "trigger.value + 10",
							},
						},
						map[string]interface{}{
							"id":   "branch3",
							"type": "transform",
							"config": map[string]interface{}{
								"formula": "trigger.value / 2",
							},
						},
					},
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	triggerData := map[string]interface{}{
		"value": 42,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		execCtx := &ExecutionContext{
			TenantID:    "bench-tenant",
			ExecutionID: "exec-1",
			WorkflowID:  wf.ID,
			TriggerType: "manual",
			TriggerData: triggerData,
			StepOutputs: make(map[string]interface{}),
		}

		_, _ = exec.executeParallelNode(ctx, wf, &wf.Nodes[0], execCtx)
	}
}

// BenchmarkCircuitBreakerCheck benchmarks circuit breaker state check
func BenchmarkCircuitBreakerCheck(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := DefaultCircuitBreakerConfig()
	registry := NewCircuitBreakerRegistry(config, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.IsOpen("test-service")
	}
}

// BenchmarkContextDataMarshaling benchmarks context data serialization
func BenchmarkContextDataMarshaling(b *testing.B) {
	execCtx := &ExecutionContext{
		TenantID:    "bench-tenant",
		ExecutionID: "exec-1",
		WorkflowID:  "wf-1",
		TriggerType: "manual",
		TriggerData: map[string]interface{}{
			"value":     42,
			"timestamp": "2025-12-20T10:00:00Z",
			"metadata": map[string]interface{}{
				"source": "api",
				"user":   "test@example.com",
			},
		},
		StepOutputs: map[string]interface{}{
			"step1": map[string]interface{}{
				"result": 84,
				"status": "success",
			},
			"step2": map[string]interface{}{
				"result": 94,
				"status": "success",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(execCtx)
	}
}

// BenchmarkContextDataUnmarshaling benchmarks context data deserialization
func BenchmarkContextDataUnmarshaling(b *testing.B) {
	data := []byte(`{
		"TenantID": "bench-tenant",
		"ExecutionID": "exec-1",
		"WorkflowID": "wf-1",
		"TriggerType": "manual",
		"TriggerData": {
			"value": 42,
			"timestamp": "2025-12-20T10:00:00Z",
			"metadata": {
				"source": "api",
				"user": "test@example.com"
			}
		},
		"StepOutputs": {
			"step1": {
				"result": 84,
				"status": "success"
			},
			"step2": {
				"result": 94,
				"status": "success"
			}
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var execCtx ExecutionContext
		_ = json.Unmarshal(data, &execCtx)
	}
}

// BenchmarkRetryStrategyCalculation benchmarks retry delay calculation
func BenchmarkRetryStrategyCalculation(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	config := DefaultRetryConfig()
	strategy := NewRetryStrategy(config, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strategy.GetDelay(i % 5) // Test with attempts 0-4
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation during execution
func BenchmarkMemoryAllocation(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := &workflow.Repository{}
	exec := New(repo, logger)

	wf := &workflow.Workflow{
		ID:   "bench-memory",
		Name: "Memory Benchmark Workflow",
		Nodes: []workflow.Node{
			{
				ID:   "step1",
				Type: "transform",
				Config: map[string]interface{}{
					"formula": "trigger.value * 2",
				},
			},
		},
		Edges: []workflow.Edge{},
	}

	triggerData := map[string]interface{}{
		"value": 42,
	}

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		execCtx := &ExecutionContext{
			TenantID:    "bench-tenant",
			ExecutionID: "exec-1",
			WorkflowID:  wf.ID,
			TriggerType: "manual",
			TriggerData: triggerData,
			StepOutputs: make(map[string]interface{}),
		}

		_, _ = exec.executeNode(ctx, wf, &wf.Nodes[0], execCtx)
	}
}
