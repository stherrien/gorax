package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Provider wraps the OpenTelemetry tracer provider with additional functionality
type Provider struct {
	tp       *sdktrace.TracerProvider
	config   *TracingConfig
	shutdown sync.Once
	mu       sync.RWMutex
	healthy  bool
}

// globalProvider holds the singleton provider instance
var (
	globalProvider *Provider
	globalMu       sync.RWMutex
)

// InitTracing initializes the OpenTelemetry tracing provider with the given configuration
// Returns a Provider that can be used to access the tracer and a cleanup function
func InitTracing(ctx context.Context, cfg *TracingConfig) (*Provider, func(), error) {
	if cfg == nil {
		cfg = &TracingConfig{Enabled: false}
	}

	// Validate configuration
	if err := cfg.ValidateConfig(); err != nil {
		return nil, nil, fmt.Errorf("invalid tracing configuration: %w", err)
	}

	// If tracing is disabled, return a no-op provider
	if !cfg.Enabled {
		noopTP := trace.NewNoopTracerProvider()
		otel.SetTracerProvider(noopTP)
		return &Provider{
			config:  cfg,
			healthy: true,
		}, func() {}, nil
	}

	// Create resource with service information
	res, err := createResource(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter based on configuration
	exporter, err := createExporter(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create sampler
	sampler := createSampler(cfg.SamplingRate)

	// Create batch span processor with configured options
	bsp := sdktrace.NewBatchSpanProcessor(exporter,
		sdktrace.WithMaxQueueSize(cfg.BatchConfig.MaxQueueSize),
		sdktrace.WithBatchTimeout(time.Duration(cfg.BatchConfig.BatchTimeoutMs)*time.Millisecond),
		sdktrace.WithExportTimeout(time.Duration(cfg.BatchConfig.ExportTimeoutMs)*time.Millisecond),
		sdktrace.WithMaxExportBatchSize(cfg.BatchConfig.MaxExportBatchSize),
	)

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for distributed tracing
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	provider := &Provider{
		tp:      tp,
		config:  cfg,
		healthy: true,
	}

	// Store as global provider
	globalMu.Lock()
	globalProvider = provider
	globalMu.Unlock()

	// Create cleanup function
	cleanup := func() {
		provider.Shutdown(context.Background())
	}

	slog.Info("tracing initialized",
		"service_name", cfg.ServiceName,
		"exporter_type", cfg.ExporterType,
		"endpoint", cfg.ExporterEndpoint,
		"sampling_rate", cfg.SamplingRate,
	)

	return provider, cleanup, nil
}

// createResource creates an OpenTelemetry resource with service information
func createResource(ctx context.Context, cfg *TracingConfig) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(cfg.ServiceName),
		semconv.ServiceVersionKey.String(cfg.ServiceVersion),
	}

	// Add custom resource attributes
	for key, value := range cfg.ResourceAttributes {
		attrs = append(attrs, attribute.String(key, value))
	}

	return resource.New(ctx,
		resource.WithAttributes(attrs...),
		resource.WithProcessRuntimeDescription(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
	)
}

// createExporter creates a trace exporter based on configuration
func createExporter(ctx context.Context, cfg *TracingConfig) (sdktrace.SpanExporter, error) {
	switch cfg.ExporterType {
	case ExporterTypeOTLP, ExporterTypeJaeger:
		// Both OTLP and Jaeger now use OTLP exporter (Jaeger supports OTLP natively)
		return createOTLPExporter(ctx, cfg)
	case ExporterTypeConsole:
		return createConsoleExporter()
	case ExporterTypeNone:
		return &noopExporter{}, nil
	default:
		return createOTLPExporter(ctx, cfg)
	}
}

// createOTLPExporter creates an OTLP gRPC exporter
func createOTLPExporter(ctx context.Context, cfg *TracingConfig) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
	}

	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	}

	// Add headers if configured
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
	}

	return otlptracegrpc.New(ctx, opts...)
}

// createConsoleExporter creates a stdout exporter for debugging
func createConsoleExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithWriter(os.Stdout),
	)
}

// createSampler creates a sampler based on the sampling rate
func createSampler(rate float64) sdktrace.Sampler {
	if rate >= 1.0 {
		return sdktrace.AlwaysSample()
	}
	if rate <= 0.0 {
		return sdktrace.NeverSample()
	}
	return sdktrace.TraceIDRatioBased(rate)
}

// Shutdown gracefully shuts down the tracer provider
func (p *Provider) Shutdown(ctx context.Context) {
	p.shutdown.Do(func() {
		p.mu.Lock()
		p.healthy = false
		p.mu.Unlock()

		if p.tp != nil {
			if err := p.tp.Shutdown(ctx); err != nil {
				slog.Error("failed to shutdown tracer provider", "error", err)
			}
		}

		// Clear global provider
		globalMu.Lock()
		if globalProvider == p {
			globalProvider = nil
		}
		globalMu.Unlock()
	})
}

// Tracer returns a named tracer from the provider
func (p *Provider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	if p.tp != nil {
		return p.tp.Tracer(name, opts...)
	}
	return otel.Tracer(name, opts...)
}

// IsHealthy returns whether the tracing provider is healthy
func (p *Provider) IsHealthy() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.healthy
}

// Config returns the tracing configuration
func (p *Provider) Config() *TracingConfig {
	return p.config
}

// ForceFlush forces a flush of all pending spans
func (p *Provider) ForceFlush(ctx context.Context) error {
	if p.tp != nil {
		return p.tp.ForceFlush(ctx)
	}
	return nil
}

// GetGlobalProvider returns the global tracing provider
func GetGlobalProvider() *Provider {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalProvider
}

// noopExporter is a no-operation span exporter
type noopExporter struct{}

func (e *noopExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (e *noopExporter) Shutdown(ctx context.Context) error {
	return nil
}
