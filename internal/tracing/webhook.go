package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TraceWebhookReceive wraps webhook receive handling with tracing
func TraceWebhookReceive(ctx context.Context, tenantID, webhookID, method, path string, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, "webhook.receive")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant_id", tenantID),
		attribute.String("webhook_id", webhookID),
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.String("component", "webhook"),
	)

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "webhook received successfully")
	return nil
}

// TraceWebhookReplay wraps webhook replay with tracing
func TraceWebhookReplay(ctx context.Context, tenantID, webhookID, eventID string, fn func(context.Context) (string, error)) (string, error) {
	ctx, span := StartSpan(ctx, "webhook.replay")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant_id", tenantID),
		attribute.String("webhook_id", webhookID),
		attribute.String("event_id", eventID),
		attribute.String("component", "webhook"),
	)

	executionID, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	span.SetAttributes(attribute.String("execution_id", executionID))
	span.SetStatus(codes.Ok, "webhook replay completed")
	return executionID, nil
}

// TraceWebhookBatchReplay wraps batch webhook replay with tracing
func TraceWebhookBatchReplay(ctx context.Context, tenantID, webhookID string, batchSize int, fn func(context.Context) (int, int, error)) error {
	ctx, span := StartSpan(ctx, "webhook.batch_replay")
	defer span.End()

	span.SetAttributes(
		attribute.String("tenant_id", tenantID),
		attribute.String("webhook_id", webhookID),
		attribute.Int("batch_size", batchSize),
		attribute.String("component", "webhook"),
	)

	successCount, failureCount, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(
		attribute.Int("success_count", successCount),
		attribute.Int("failure_count", failureCount),
	)
	span.SetStatus(codes.Ok, "batch replay completed")
	return nil
}

// TraceWebhookValidation wraps webhook validation with tracing
func TraceWebhookValidation(ctx context.Context, webhookID, authType string, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, "webhook.validate")
	defer span.End()

	span.SetAttributes(
		attribute.String("webhook_id", webhookID),
		attribute.String("auth_type", authType),
		attribute.String("component", "webhook"),
	)

	err := fn(ctx)
	if err != nil {
		span.SetAttributes(attribute.Bool("validation_passed", false))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(attribute.Bool("validation_passed", true))
	span.SetStatus(codes.Ok, "validation passed")
	return nil
}

// TraceWebhookEventStore wraps webhook event storage with tracing
func TraceWebhookEventStore(ctx context.Context, webhookID, eventID string, payloadSize int, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, "webhook.event.store")
	defer span.End()

	span.SetAttributes(
		attribute.String("webhook_id", webhookID),
		attribute.String("event_id", eventID),
		attribute.Int("payload_size_bytes", payloadSize),
		attribute.String("component", "webhook"),
	)

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "event stored successfully")
	return nil
}

// AddWebhookAttributes adds webhook-specific attributes to the current span
func AddWebhookAttributes(ctx context.Context, attrs map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		SetSpanAttributes(span, attrs)
	}
}

// RecordWebhookEvent records a webhook-related event on the current span
func RecordWebhookEvent(ctx context.Context, eventName string, attrs map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
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
