package tracing

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TraceWorkflowExecution wraps a workflow execution with tracing
func TraceWorkflowExecution(ctx context.Context, tenantID, workflowID, executionID string, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, "workflow.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant_id", tenantID),
		attribute.String("workflow_id", workflowID),
		attribute.String("execution_id", executionID),
		attribute.String("component", "executor"),
	)

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "workflow execution completed")
	return nil
}

// TraceStepExecution wraps a step execution with tracing
func TraceStepExecution(ctx context.Context, tenantID, workflowID, executionID, nodeID, nodeType string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	ctx, span := StartSpan(ctx, "workflow.step.execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant_id", tenantID),
		attribute.String("workflow_id", workflowID),
		attribute.String("execution_id", executionID),
		attribute.String("node_id", nodeID),
		attribute.String("node_type", nodeType),
		attribute.String("component", "executor"),
	)

	output, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Add output size as attribute (for observability)
	if outputJSON, err := json.Marshal(output); err == nil {
		span.SetAttributes(attribute.Int("output_size_bytes", len(outputJSON)))
	}

	span.SetStatus(codes.Ok, "step execution completed")
	return output, nil
}

// TraceHTTPAction wraps an HTTP action with tracing
func TraceHTTPAction(ctx context.Context, method, url string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	ctx, span := StartSpan(ctx, "http.action")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", method),
		attribute.String("http.url", url),
		attribute.String("component", "http_action"),
	)

	output, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "HTTP action completed")
	return output, nil
}

// TraceQueueMessage wraps queue message processing with tracing
func TraceQueueMessage(ctx context.Context, queueName, messageID string, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, "queue.process_message")
	defer span.End()

	span.SetAttributes(
		attribute.String("queue.name", queueName),
		attribute.String("queue.message_id", messageID),
		attribute.String("component", "queue_consumer"),
	)

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "message processed successfully")
	return nil
}

// TraceSubWorkflow wraps a sub-workflow execution with tracing
func TraceSubWorkflow(ctx context.Context, parentWorkflowID, childWorkflowID, executionID string, depth int, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, "workflow.sub_workflow")
	defer span.End()

	span.SetAttributes(
		attribute.String("parent_workflow_id", parentWorkflowID),
		attribute.String("child_workflow_id", childWorkflowID),
		attribute.String("execution_id", executionID),
		attribute.Int("depth", depth),
		attribute.String("component", "executor"),
	)

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "sub-workflow completed")
	return nil
}

// TraceActionWithRetry wraps an action execution with retry tracing
func TraceActionWithRetry(ctx context.Context, actionName string, attempt int, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, "action.retry")
	defer span.End()

	span.SetAttributes(
		attribute.String("action_name", actionName),
		attribute.Int("attempt", attempt),
		attribute.String("component", "retry_strategy"),
	)

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.Bool("retry_needed", true))
		return err
	}

	span.SetStatus(codes.Ok, "action succeeded")
	span.SetAttributes(attribute.Bool("retry_needed", false))
	return nil
}

// AddWorkflowAttributes adds workflow-specific attributes to the current span
func AddWorkflowAttributes(ctx context.Context, attrs map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		SetSpanAttributes(span, attrs)
	}
}

// RecordWorkflowEvent records a workflow event on the current span
func RecordWorkflowEvent(ctx context.Context, eventName string, attrs map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		// Convert attrs to attribute.KeyValue slice
		var kvAttrs []attribute.KeyValue
		for key, value := range attrs {
			switch v := value.(type) {
			case string:
				kvAttrs = append(kvAttrs, attribute.String(key, v))
			case int:
				kvAttrs = append(kvAttrs, attribute.Int(key, v))
			case int64:
				kvAttrs = append(kvAttrs, attribute.Int64(key, v))
			case float64:
				kvAttrs = append(kvAttrs, attribute.Float64(key, v))
			case bool:
				kvAttrs = append(kvAttrs, attribute.Bool(key, v))
			}
		}
		span.AddEvent(eventName, trace.WithAttributes(kvAttrs...))
	}
}

// RecordErrorWithStackTrace records an error on the span with a stack trace
func RecordErrorWithStackTrace(span trace.Span, err error) {
	if err == nil || !span.SpanContext().IsValid() {
		return
	}

	// Capture stack trace
	stackTrace := captureStackTrace(3) // Skip 3 frames: runtime.Callers, captureStackTrace, RecordErrorWithStackTrace

	// Record the error with stack trace
	span.RecordError(err, trace.WithStackTrace(true))
	span.SetAttributes(
		attribute.String("error.message", err.Error()),
		attribute.String("error.stack_trace", stackTrace),
	)
	span.SetStatus(codes.Error, err.Error())
}

// captureStackTrace captures a stack trace, skipping the specified number of frames
func captureStackTrace(skip int) string {
	const maxFrames = 32
	pcs := make([]uintptr, maxFrames)
	n := runtime.Callers(skip, pcs)
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var sb strings.Builder

	for {
		frame, more := frames.Next()
		// Skip runtime and standard library frames
		if !strings.Contains(frame.File, "runtime/") {
			sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		}
		if !more {
			break
		}
	}

	return sb.String()
}

// TraceLoopIteration wraps a loop iteration with tracing
func TraceLoopIteration(ctx context.Context, iterationIndex int, itemVariable string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	ctx, span := StartSpan(ctx, "workflow.loop.iteration",
		trace.WithAttributes(
			attribute.Int("loop.iteration.index", iterationIndex),
			attribute.String("loop.item_variable", itemVariable),
			attribute.String("component", "executor"),
		),
	)
	defer span.End()

	output, err := fn(ctx)
	if err != nil {
		RecordErrorWithStackTrace(span, err)
		return nil, err
	}

	span.SetStatus(codes.Ok, "iteration completed")
	return output, nil
}

// TraceParallelBranch wraps a parallel branch execution with tracing
func TraceParallelBranch(ctx context.Context, branchIndex int, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	ctx, span := StartSpan(ctx, "workflow.parallel.branch",
		trace.WithAttributes(
			attribute.Int("parallel.branch.index", branchIndex),
			attribute.String("component", "executor"),
		),
	)
	defer span.End()

	output, err := fn(ctx)
	if err != nil {
		RecordErrorWithStackTrace(span, err)
		return nil, err
	}

	span.SetStatus(codes.Ok, "branch completed")
	return output, nil
}

// TraceNodeExecution wraps a single node execution with tracing
func TraceNodeExecution(ctx context.Context, nodeID, nodeType string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	ctx, span := StartSpan(ctx, fmt.Sprintf("workflow.node.%s", nodeType),
		trace.WithAttributes(
			attribute.String("node.id", nodeID),
			attribute.String("node.type", nodeType),
			attribute.String("component", "executor"),
		),
	)
	defer span.End()

	output, err := fn(ctx)
	if err != nil {
		RecordErrorWithStackTrace(span, err)
		return nil, err
	}

	// Add output size as attribute
	if outputJSON, jsonErr := json.Marshal(output); jsonErr == nil {
		span.SetAttributes(attribute.Int("node.output_size_bytes", len(outputJSON)))
	}

	span.SetStatus(codes.Ok, "node execution completed")
	return output, nil
}

// TraceRetryAttempt wraps a retry attempt with tracing
func TraceRetryAttempt(ctx context.Context, nodeID string, attempt, maxRetries int, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, "workflow.retry.attempt",
		trace.WithAttributes(
			attribute.String("node.id", nodeID),
			attribute.Int("retry.attempt", attempt),
			attribute.Int("retry.max_retries", maxRetries),
			attribute.String("component", "retry_strategy"),
		),
	)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		span.SetAttributes(attribute.Bool("retry.will_retry", attempt < maxRetries))
		RecordErrorWithStackTrace(span, err)
		return err
	}

	span.SetStatus(codes.Ok, "attempt succeeded")
	return nil
}

// TraceCredentialInjection wraps credential injection with tracing
func TraceCredentialInjection(ctx context.Context, tenantID, workflowID, nodeID string, credentialCount int, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, "workflow.credential.inject",
		trace.WithAttributes(
			attribute.String("tenant_id", tenantID),
			attribute.String("workflow_id", workflowID),
			attribute.String("node_id", nodeID),
			attribute.Int("credential.count", credentialCount),
			attribute.String("component", "credential_injector"),
		),
	)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		RecordErrorWithStackTrace(span, err)
		return err
	}

	span.SetStatus(codes.Ok, "credentials injected")
	return nil
}
