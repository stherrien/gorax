package tracing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/gorax/gorax/internal/config"
)

func TestNewTracerProvider_Disabled(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled: false,
	}

	tp, cleanup, err := NewTracerProvider(context.Background(), cfg)

	require.NoError(t, err)
	assert.NotNil(t, tp)
	assert.NotNil(t, cleanup)

	// Verify tracer is NoOp
	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	assert.False(t, span.SpanContext().IsValid())

	cleanup()
}

func TestNewTracerProvider_Enabled(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	tp, cleanup, err := NewTracerProvider(context.Background(), cfg)

	require.NoError(t, err)
	assert.NotNil(t, tp)
	assert.NotNil(t, cleanup)

	// Verify tracer is valid
	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	assert.True(t, span.SpanContext().IsValid())
	span.End()

	cleanup()
}

func TestNewTracerProvider_WithSampling(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate float64
	}{
		{"full_sampling", 1.0},
		{"half_sampling", 0.5},
		{"no_sampling", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ObservabilityConfig{
				TracingEnabled:     true,
				TracingEndpoint:    "localhost:4317",
				TracingSampleRate:  tt.sampleRate,
				TracingServiceName: "gorax-test",
			}

			tp, cleanup, err := NewTracerProvider(context.Background(), cfg)
			require.NoError(t, err)
			defer cleanup()

			assert.NotNil(t, tp)
		})
	}
}

func TestInitGlobalTracer(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	// Verify global tracer is set
	tracer := otel.Tracer("test")
	assert.NotNil(t, tracer)

	ctx, span := tracer.Start(context.Background(), "test-span")
	assert.True(t, span.SpanContext().IsValid())
	assert.NotNil(t, ctx)
	span.End()
}

func TestStartSpan(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test-operation")

	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	assert.True(t, span.SpanContext().IsValid())

	span.End()
}

func TestStartSpan_WithParent(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	// Create parent span
	parentCtx, parentSpan := StartSpan(context.Background(), "parent-operation")
	parentSpanContext := parentSpan.SpanContext()

	// Create child span
	_, childSpan := StartSpan(parentCtx, "child-operation")
	childSpanContext := childSpan.SpanContext()

	// Verify parent-child relationship
	assert.True(t, childSpanContext.IsValid())
	assert.Equal(t, parentSpanContext.TraceID(), childSpanContext.TraceID())
	assert.NotEqual(t, parentSpanContext.SpanID(), childSpanContext.SpanID())

	childSpan.End()
	parentSpan.End()
}

func TestRecordError(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	_, span := StartSpan(context.Background(), "test-operation")
	defer span.End()

	testErr := assert.AnError
	RecordError(span, testErr)

	// Verify span recorded error
	// Note: We can't directly assert on span internals, but we verify no panic
}

func TestSetSpanAttributes(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	_, span := StartSpan(context.Background(), "test-operation")
	defer span.End()

	attrs := map[string]interface{}{
		"string_attr": "value",
		"int_attr":    42,
		"bool_attr":   true,
		"float_attr":  3.14,
	}

	SetSpanAttributes(span, attrs)

	// Verify no panic
}

func TestGetTraceID(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	ctx, span := StartSpan(context.Background(), "test-operation")
	defer span.End()

	traceID := GetTraceID(ctx)
	assert.NotEmpty(t, traceID)
	assert.Len(t, traceID, 32) // Trace ID should be 32 hex characters
}

func TestGetSpanID(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	ctx, span := StartSpan(context.Background(), "test-operation")
	defer span.End()

	spanID := GetSpanID(ctx)
	assert.NotEmpty(t, spanID)
	assert.Len(t, spanID, 16) // Span ID should be 16 hex characters
}

func TestExtractTraceContext(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	ctx, span := StartSpan(context.Background(), "test-operation")
	defer span.End()

	headers := map[string]string{
		"content-type": "application/json",
	}

	InjectTraceContext(ctx, headers)

	// Should have traceparent header
	assert.NotEmpty(t, headers["traceparent"])
}

func TestSpanFromContext_NoSpan(t *testing.T) {
	span := trace.SpanFromContext(context.Background())
	assert.NotNil(t, span)
	assert.False(t, span.SpanContext().IsValid())
}

func TestSpanFromContext_WithSpan(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	ctx, originalSpan := StartSpan(context.Background(), "test-operation")
	defer originalSpan.End()

	retrievedSpan := trace.SpanFromContext(ctx)
	assert.NotNil(t, retrievedSpan)
	assert.True(t, retrievedSpan.SpanContext().IsValid())
	assert.Equal(t, originalSpan.SpanContext().SpanID(), retrievedSpan.SpanContext().SpanID())
}
