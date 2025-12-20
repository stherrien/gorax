package tracing

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupTestTracer(t *testing.T) (*tracetest.InMemoryExporter, func()) {
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

func TestTraceWebhookReceive_Success(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()
	webhookID := "webhook-123"
	tenantID := "tenant-456"
	method := "POST"
	path := "/hooks/webhook-123"

	err := TraceWebhookReceive(ctx, tenantID, webhookID, method, path, func(ctx context.Context) error {
		return nil
	})

	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "webhook.receive", span.Name)
	assert.Equal(t, codes.Ok, span.Status.Code)

	// Check attributes
	attrMap := make(map[string]interface{})
	for _, attr := range span.Attributes {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, tenantID, attrMap["tenant_id"])
	assert.Equal(t, webhookID, attrMap["webhook_id"])
	assert.Equal(t, method, attrMap["http.method"])
	assert.Equal(t, path, attrMap["http.path"])
	assert.Equal(t, "webhook", attrMap["component"])
}

func TestTraceWebhookReceive_Error(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()
	expectedErr := errors.New("webhook processing failed")

	err := TraceWebhookReceive(ctx, "tenant", "webhook", "POST", "/path", func(ctx context.Context) error {
		return expectedErr
	})

	assert.Equal(t, expectedErr, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, codes.Error, span.Status.Code)
	assert.Contains(t, span.Status.Description, "webhook processing failed")
}

func TestTraceWebhookReplay_Success(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()
	tenantID := "tenant-123"
	eventID := "event-456"
	webhookID := "webhook-789"
	executionID := "exec-000"

	result, err := TraceWebhookReplay(ctx, tenantID, webhookID, eventID, func(ctx context.Context) (string, error) {
		return executionID, nil
	})

	require.NoError(t, err)
	assert.Equal(t, executionID, result)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "webhook.replay", span.Name)
	assert.Equal(t, codes.Ok, span.Status.Code)

	attrMap := make(map[string]interface{})
	for _, attr := range span.Attributes {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, tenantID, attrMap["tenant_id"])
	assert.Equal(t, webhookID, attrMap["webhook_id"])
	assert.Equal(t, eventID, attrMap["event_id"])
	assert.Equal(t, "webhook", attrMap["component"])
}

func TestTraceWebhookReplay_Error(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()
	expectedErr := errors.New("replay failed")

	result, err := TraceWebhookReplay(ctx, "tenant", "webhook", "event", func(ctx context.Context) (string, error) {
		return "", expectedErr
	})

	assert.Equal(t, expectedErr, err)
	assert.Empty(t, result)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, codes.Error, span.Status.Code)
}

func TestTraceWebhookBatchReplay_Success(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()
	tenantID := "tenant-123"
	webhookID := "webhook-456"
	eventIDs := []string{"event-1", "event-2", "event-3"}
	successCount := 3
	failureCount := 0

	err := TraceWebhookBatchReplay(ctx, tenantID, webhookID, len(eventIDs), func(ctx context.Context) (int, int, error) {
		return successCount, failureCount, nil
	})

	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "webhook.batch_replay", span.Name)
	assert.Equal(t, codes.Ok, span.Status.Code)

	attrMap := make(map[string]interface{})
	for _, attr := range span.Attributes {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, tenantID, attrMap["tenant_id"])
	assert.Equal(t, webhookID, attrMap["webhook_id"])
	assert.Equal(t, int64(3), attrMap["batch_size"])
	assert.Equal(t, int64(3), attrMap["success_count"])
	assert.Equal(t, int64(0), attrMap["failure_count"])
}

func TestTraceWebhookBatchReplay_PartialFailure(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()
	successCount := 2
	failureCount := 1

	err := TraceWebhookBatchReplay(ctx, "tenant", "webhook", 3, func(ctx context.Context) (int, int, error) {
		return successCount, failureCount, nil
	})

	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	// Should still be OK status even with partial failures
	assert.Equal(t, codes.Ok, span.Status.Code)

	attrMap := make(map[string]interface{})
	for _, attr := range span.Attributes {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, int64(2), attrMap["success_count"])
	assert.Equal(t, int64(1), attrMap["failure_count"])
}

func TestTraceWebhookValidation_Success(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()
	webhookID := "webhook-123"
	authType := "signature"

	err := TraceWebhookValidation(ctx, webhookID, authType, func(ctx context.Context) error {
		return nil
	})

	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "webhook.validate", span.Name)
	assert.Equal(t, codes.Ok, span.Status.Code)

	attrMap := make(map[string]interface{})
	for _, attr := range span.Attributes {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, webhookID, attrMap["webhook_id"])
	assert.Equal(t, authType, attrMap["auth_type"])
	assert.Equal(t, true, attrMap["validation_passed"])
}

func TestTraceWebhookValidation_Failed(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()
	validationErr := errors.New("invalid signature")

	err := TraceWebhookValidation(ctx, "webhook", "signature", func(ctx context.Context) error {
		return validationErr
	})

	assert.Equal(t, validationErr, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, codes.Error, span.Status.Code)

	attrMap := make(map[string]interface{})
	for _, attr := range span.Attributes {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, false, attrMap["validation_passed"])
}

func TestTraceWebhookEventStore_Success(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()
	webhookID := "webhook-123"
	eventID := "event-456"
	payloadSize := 1024

	err := TraceWebhookEventStore(ctx, webhookID, eventID, payloadSize, func(ctx context.Context) error {
		return nil
	})

	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "webhook.event.store", span.Name)
	assert.Equal(t, codes.Ok, span.Status.Code)

	attrMap := make(map[string]interface{})
	for _, attr := range span.Attributes {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, webhookID, attrMap["webhook_id"])
	assert.Equal(t, eventID, attrMap["event_id"])
	assert.Equal(t, int64(1024), attrMap["payload_size_bytes"])
}

func TestAddWebhookAttributes(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()

	// Start a span first
	ctx, span := StartSpan(ctx, "test.span")

	AddWebhookAttributes(ctx, map[string]interface{}{
		"webhook_id":  "webhook-123",
		"tenant_id":   "tenant-456",
		"custom_attr": "custom_value",
		"count":       42,
		"enabled":     true,
		"rate":        0.95,
	})

	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	attrMap := make(map[string]interface{})
	for _, attr := range spans[0].Attributes {
		attrMap[string(attr.Key)] = attr.Value.AsInterface()
	}

	assert.Equal(t, "webhook-123", attrMap["webhook_id"])
	assert.Equal(t, "tenant-456", attrMap["tenant_id"])
	assert.Equal(t, "custom_value", attrMap["custom_attr"])
	assert.Equal(t, int64(42), attrMap["count"])
	assert.Equal(t, true, attrMap["enabled"])
	assert.Equal(t, 0.95, attrMap["rate"])
}

func TestRecordWebhookEvent(t *testing.T) {
	exporter, cleanup := setupTestTracer(t)
	defer cleanup()

	ctx := context.Background()

	// Start a span first
	ctx, span := StartSpan(ctx, "test.span")

	RecordWebhookEvent(ctx, "webhook.received", map[string]interface{}{
		"webhook_id": "webhook-123",
		"status":     "success",
	})

	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	events := spans[0].Events
	require.Len(t, events, 1)
	assert.Equal(t, "webhook.received", events[0].Name)
}
