package tracing

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gorax/gorax/internal/config"
)

// NewTracerProvider creates a new OpenTelemetry tracer provider
// Returns the provider, a cleanup function, and any error
func NewTracerProvider(ctx context.Context, cfg *config.ObservabilityConfig) (trace.TracerProvider, func(), error) {
	// If tracing is disabled, return a no-op tracer provider
	if !cfg.TracingEnabled {
		return trace.NewNoopTracerProvider(), func() {}, nil
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.TracingServiceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP trace exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.TracingEndpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create sampler based on sample rate
	var sampler sdktrace.Sampler
	if cfg.TracingSampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.TracingSampleRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.TracingSampleRate)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Cleanup function
	cleanup := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer provider", "error", err)
		}
	}

	return tp, cleanup, nil
}

// InitGlobalTracer initializes the global OpenTelemetry tracer
// Returns a cleanup function that should be called on application shutdown
func InitGlobalTracer(ctx context.Context, cfg *config.ObservabilityConfig) (func(), error) {
	tp, cleanup, err := NewTracerProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator to W3C Trace Context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return cleanup, nil
}

// StartSpan starts a new span with the given name
// Returns a context with the span and the span itself
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer("gorax")
	return tracer.Start(ctx, spanName, opts...)
}

// RecordError records an error on the given span
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
	}
}

// SetSpanAttributes sets multiple attributes on a span
func SetSpanAttributes(span trace.Span, attrs map[string]interface{}) {
	for key, value := range attrs {
		switch v := value.(type) {
		case string:
			span.SetAttributes(attribute.String(key, v))
		case int:
			span.SetAttributes(attribute.Int(key, v))
		case int64:
			span.SetAttributes(attribute.Int64(key, v))
		case float64:
			span.SetAttributes(attribute.Float64(key, v))
		case bool:
			span.SetAttributes(attribute.Bool(key, v))
		default:
			span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}
}

// GetTraceID returns the trace ID from the context
func GetTraceID(ctx context.Context) string {
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	if spanContext.HasTraceID() {
		return spanContext.TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from the context
func GetSpanID(ctx context.Context) string {
	spanContext := trace.SpanFromContext(ctx).SpanContext()
	if spanContext.HasSpanID() {
		return spanContext.SpanID().String()
	}
	return ""
}

// InjectTraceContext injects trace context into a map (e.g., HTTP headers)
func InjectTraceContext(ctx context.Context, carrier map[string]string) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.MapCarrier(carrier))
}

// ExtractTraceContext extracts trace context from a map (e.g., HTTP headers)
func ExtractTraceContext(ctx context.Context, carrier map[string]string) context.Context {
	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, propagation.MapCarrier(carrier))
}
