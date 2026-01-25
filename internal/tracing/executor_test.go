package tracing

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupExecutorTestTracer(t *testing.T) (*tracetest.InMemoryExporter, func()) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	// Set global tracer provider
	originalTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)

	cleanup := func() {
		otel.SetTracerProvider(originalTP)
		tp.Shutdown(context.Background())
	}

	return exporter, cleanup
}

func TestTraceWorkflowExecution_Success(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	executed := false
	err := TraceWorkflowExecution(
		context.Background(),
		"tenant-123",
		"workflow-456",
		"execution-789",
		func(ctx context.Context) error {
			executed = true
			// Verify context has valid span
			traceID := GetTraceID(ctx)
			assert.NotEmpty(t, traceID)
			return nil
		},
	)

	assert.NoError(t, err)
	assert.True(t, executed)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, "workflow.execute", spans[0].Name)
	assert.Equal(t, codes.Ok, spans[0].Status.Code)
}

func TestTraceWorkflowExecution_Error(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedErr := errors.New("workflow failed")
	err := TraceWorkflowExecution(
		context.Background(),
		"tenant-123",
		"workflow-456",
		"execution-789",
		func(ctx context.Context) error {
			return expectedErr
		},
	)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
}

func TestTraceStepExecution_Success(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedOutput := map[string]any{"status": "ok"}
	output, err := TraceStepExecution(
		context.Background(),
		"tenant-123",
		"workflow-456",
		"execution-789",
		"node-1",
		"action.http",
		func(ctx context.Context) (any, error) {
			// Verify context has valid span
			traceID := GetTraceID(ctx)
			assert.NotEmpty(t, traceID)
			return expectedOutput, nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, "workflow.step.execute", spans[0].Name)
	assert.Equal(t, codes.Ok, spans[0].Status.Code)
}

func TestTraceStepExecution_Error(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedErr := errors.New("step failed")
	output, err := TraceStepExecution(
		context.Background(),
		"tenant-123",
		"workflow-456",
		"execution-789",
		"node-1",
		"action.http",
		func(ctx context.Context) (any, error) {
			return nil, expectedErr
		},
	)

	assert.Error(t, err)
	assert.Nil(t, output)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
}

func TestTraceHTTPAction_Success(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedOutput := map[string]any{"response": "data"}
	output, err := TraceHTTPAction(
		context.Background(),
		"POST",
		"https://api.example.com/data",
		func(ctx context.Context) (any, error) {
			// Verify context has valid span
			traceID := GetTraceID(ctx)
			assert.NotEmpty(t, traceID)
			return expectedOutput, nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, "http.action", spans[0].Name)
	assert.Equal(t, codes.Ok, spans[0].Status.Code)
}

func TestTraceHTTPAction_Error(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedErr := errors.New("connection refused")
	output, err := TraceHTTPAction(
		context.Background(),
		"GET",
		"https://api.example.com/data",
		func(ctx context.Context) (any, error) {
			return nil, expectedErr
		},
	)

	assert.Error(t, err)
	assert.Nil(t, output)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
}

func TestTraceSubWorkflow_Success(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	executed := false
	err := TraceSubWorkflow(
		context.Background(),
		"parent-workflow-123",
		"child-workflow-456",
		"execution-789",
		2,
		func(ctx context.Context) error {
			executed = true
			// Verify context has valid span
			traceID := GetTraceID(ctx)
			assert.NotEmpty(t, traceID)
			return nil
		},
	)

	assert.NoError(t, err)
	assert.True(t, executed)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, "workflow.sub_workflow", spans[0].Name)
	assert.Equal(t, codes.Ok, spans[0].Status.Code)
}

func TestTraceSubWorkflow_Error(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedErr := errors.New("sub-workflow failed")
	err := TraceSubWorkflow(
		context.Background(),
		"parent-workflow-123",
		"child-workflow-456",
		"execution-789",
		2,
		func(ctx context.Context) error {
			return expectedErr
		},
	)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
}

func TestTraceLoopIteration_Success(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedOutput := map[string]any{"processed": true}
	output, err := TraceLoopIteration(
		context.Background(),
		5,
		"item",
		func(ctx context.Context) (any, error) {
			// Verify context has valid span
			traceID := GetTraceID(ctx)
			assert.NotEmpty(t, traceID)
			return expectedOutput, nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, "workflow.loop.iteration", spans[0].Name)
	assert.Equal(t, codes.Ok, spans[0].Status.Code)
}

func TestTraceLoopIteration_Error(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedErr := errors.New("iteration failed")
	output, err := TraceLoopIteration(
		context.Background(),
		3,
		"item",
		func(ctx context.Context) (any, error) {
			return nil, expectedErr
		},
	)

	assert.Error(t, err)
	assert.Nil(t, output)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
}

func TestTraceParallelBranch_Success(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedOutput := map[string]any{"branch_result": "success"}
	output, err := TraceParallelBranch(
		context.Background(),
		2,
		func(ctx context.Context) (any, error) {
			// Verify context has valid span
			traceID := GetTraceID(ctx)
			assert.NotEmpty(t, traceID)
			return expectedOutput, nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, "workflow.parallel.branch", spans[0].Name)
	assert.Equal(t, codes.Ok, spans[0].Status.Code)
}

func TestTraceParallelBranch_Error(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedErr := errors.New("branch failed")
	output, err := TraceParallelBranch(
		context.Background(),
		1,
		func(ctx context.Context) (any, error) {
			return nil, expectedErr
		},
	)

	assert.Error(t, err)
	assert.Nil(t, output)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
}

func TestTraceNodeExecution_Success(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedOutput := map[string]any{"node_output": "data"}
	output, err := TraceNodeExecution(
		context.Background(),
		"node-abc",
		"action.transform",
		func(ctx context.Context) (any, error) {
			// Verify context has valid span
			traceID := GetTraceID(ctx)
			assert.NotEmpty(t, traceID)
			return expectedOutput, nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, "workflow.node.action.transform", spans[0].Name)
	assert.Equal(t, codes.Ok, spans[0].Status.Code)
}

func TestTraceNodeExecution_Error(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedErr := errors.New("node execution failed")
	output, err := TraceNodeExecution(
		context.Background(),
		"node-xyz",
		"action.code",
		func(ctx context.Context) (any, error) {
			return nil, expectedErr
		},
	)

	assert.Error(t, err)
	assert.Nil(t, output)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
}

func TestTraceRetryAttempt_Success(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	executed := false
	err := TraceRetryAttempt(
		context.Background(),
		"node-retry",
		1,
		3,
		func(ctx context.Context) error {
			executed = true
			// Verify context has valid span
			traceID := GetTraceID(ctx)
			assert.NotEmpty(t, traceID)
			return nil
		},
	)

	assert.NoError(t, err)
	assert.True(t, executed)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, "workflow.retry.attempt", spans[0].Name)
	assert.Equal(t, codes.Ok, spans[0].Status.Code)
}

func TestTraceRetryAttempt_Error(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedErr := errors.New("attempt failed")
	err := TraceRetryAttempt(
		context.Background(),
		"node-retry",
		2,
		3,
		func(ctx context.Context) error {
			return expectedErr
		},
	)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
}

func TestTraceCredentialInjection_Success(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	executed := false
	err := TraceCredentialInjection(
		context.Background(),
		"tenant-123",
		"workflow-456",
		"node-abc",
		3,
		func(ctx context.Context) error {
			executed = true
			// Verify context has valid span
			traceID := GetTraceID(ctx)
			assert.NotEmpty(t, traceID)
			return nil
		},
	)

	assert.NoError(t, err)
	assert.True(t, executed)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, "workflow.credential.inject", spans[0].Name)
	assert.Equal(t, codes.Ok, spans[0].Status.Code)
}

func TestTraceCredentialInjection_Error(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	expectedErr := errors.New("credential injection failed")
	err := TraceCredentialInjection(
		context.Background(),
		"tenant-123",
		"workflow-456",
		"node-abc",
		2,
		func(ctx context.Context) error {
			return expectedErr
		},
	)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
}

func TestRecordErrorWithStackTrace(t *testing.T) {
	_, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	_, span := StartSpan(context.Background(), "test-operation")
	defer span.End()

	testErr := errors.New("test error with stack trace")
	RecordErrorWithStackTrace(span, testErr)

	// Verify no panic and span is still valid
	assert.True(t, span.SpanContext().IsValid())
}

func TestRecordErrorWithStackTrace_NilError(t *testing.T) {
	_, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	_, span := StartSpan(context.Background(), "test-operation")
	defer span.End()

	// Should not panic with nil error
	RecordErrorWithStackTrace(span, nil)

	// Verify span is still valid
	assert.True(t, span.SpanContext().IsValid())
}

func TestCaptureStackTrace(t *testing.T) {
	stackTrace := captureStackTrace(1)

	// Should contain the test function
	assert.Contains(t, stackTrace, "TestCaptureStackTrace")
	// Should contain file and line info
	assert.Contains(t, stackTrace, "executor_test.go")
}

func TestRecordWorkflowEvent(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	ctx, span := StartSpan(context.Background(), "test-operation")

	RecordWorkflowEvent(ctx, "test.event", map[string]any{
		"key_string": "value",
		"key_int":    42,
		"key_bool":   true,
		"key_float":  3.14,
	})

	span.End()

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Len(t, spans[0].Events, 1)
	assert.Equal(t, "test.event", spans[0].Events[0].Name)
}

func TestAddWorkflowAttributes(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	ctx, span := StartSpan(context.Background(), "test-operation")

	AddWorkflowAttributes(ctx, map[string]any{
		"custom_attr": "custom_value",
		"count":       100,
	})

	span.End()

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	attrMap := make(map[string]any)
	for _, attr := range spans[0].Attributes {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, "custom_value", attrMap["custom_attr"])
	assert.Equal(t, int64(100), attrMap["count"])
}

func TestTraceContextPropagation(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	// Create parent workflow execution
	var parentTraceID string
	err := TraceWorkflowExecution(
		context.Background(),
		"tenant-123",
		"workflow-parent",
		"execution-parent",
		func(parentCtx context.Context) error {
			parentTraceID = GetTraceID(parentCtx)

			// Create child step execution
			_, err := TraceStepExecution(
				parentCtx,
				"tenant-123",
				"workflow-parent",
				"execution-parent",
				"node-1",
				"action.http",
				func(stepCtx context.Context) (any, error) {
					// Verify child span has same trace ID
					childTraceID := GetTraceID(stepCtx)
					assert.Equal(t, parentTraceID, childTraceID)
					return nil, nil
				},
			)
			return err
		},
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, parentTraceID)

	// Should have 2 spans with same trace ID
	spans := exporter.GetSpans()
	assert.Len(t, spans, 2)
	assert.Equal(t, spans[0].SpanContext.TraceID(), spans[1].SpanContext.TraceID())
}

func TestNestedTracingSpans(t *testing.T) {
	exporter, cleanup := setupExecutorTestTracer(t)
	defer cleanup()

	var traceIDs []string
	err := TraceWorkflowExecution(
		context.Background(),
		"tenant-123",
		"workflow-456",
		"execution-789",
		func(workflowCtx context.Context) error {
			traceIDs = append(traceIDs, GetTraceID(workflowCtx))

			// Nested step execution
			_, err := TraceStepExecution(
				workflowCtx,
				"tenant-123",
				"workflow-456",
				"execution-789",
				"node-1",
				"action.http",
				func(stepCtx context.Context) (any, error) {
					traceIDs = append(traceIDs, GetTraceID(stepCtx))

					// Nested HTTP action
					return TraceHTTPAction(
						stepCtx,
						"GET",
						"https://api.example.com",
						func(httpCtx context.Context) (any, error) {
							traceIDs = append(traceIDs, GetTraceID(httpCtx))
							return map[string]any{"status": "ok"}, nil
						},
					)
				},
			)
			return err
		},
	)

	assert.NoError(t, err)
	assert.Len(t, traceIDs, 3)

	// All trace IDs should be the same (same trace)
	for _, traceID := range traceIDs {
		assert.Equal(t, traceIDs[0], traceID)
	}

	// Should have 3 spans all with same trace ID
	spans := exporter.GetSpans()
	assert.Len(t, spans, 3)
	for _, span := range spans {
		assert.Equal(t, spans[0].SpanContext.TraceID(), span.SpanContext.TraceID())
	}
}
