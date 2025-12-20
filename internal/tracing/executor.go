package tracing

import (
	"context"
	"encoding/json"

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
