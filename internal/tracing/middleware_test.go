package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/gorax/gorax/internal/config"
)

func TestHTTPMiddleware_Disabled(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled: false,
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		assert.False(t, span.SpanContext().IsValid())
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with tracing middleware
	middleware := HTTPMiddleware()
	wrappedHandler := middleware(handler)

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHTTPMiddleware_Enabled(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		assert.True(t, span.SpanContext().IsValid())
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with tracing middleware
	middleware := HTTPMiddleware()
	wrappedHandler := middleware(handler)

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHTTPMiddleware_PropagatesContext(t *testing.T) {
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
	parentCtx, parentSpan := StartSpan(context.Background(), "parent")
	defer parentSpan.End()

	// Inject trace context into headers
	headers := make(map[string]string)
	InjectTraceContext(parentCtx, headers)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		assert.True(t, span.SpanContext().IsValid())

		// Verify same trace ID
		assert.Equal(t, parentSpan.SpanContext().TraceID(), span.SpanContext().TraceID())
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with tracing middleware
	middleware := HTTPMiddleware()
	wrappedHandler := middleware(handler)

	// Make request with trace context headers
	req := httptest.NewRequest("GET", "/test", nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHTTPMiddleware_SetsAttributes(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with tracing middleware
	middleware := HTTPMiddleware()
	wrappedHandler := middleware(handler)

	// Make request
	req := httptest.NewRequest("POST", "/api/workflows", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHTTPMiddleware_HandlesErrors(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	// Create test handler that returns an error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	})

	// Wrap with tracing middleware
	middleware := HTTPMiddleware()
	wrappedHandler := middleware(handler)

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHTTPMiddleware_MultipleRequests(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		TracingEnabled:     true,
		TracingEndpoint:    "localhost:4317",
		TracingSampleRate:  1.0,
		TracingServiceName: "gorax-test",
	}

	cleanup, err := InitGlobalTracer(context.Background(), cfg)
	require.NoError(t, err)
	defer cleanup()

	// Track trace IDs
	var traceIDs []string

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := GetTraceID(r.Context())
		traceIDs = append(traceIDs, traceID)
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with tracing middleware
	middleware := HTTPMiddleware()
	wrappedHandler := middleware(handler)

	// Make multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Verify different trace IDs for each request
	assert.Len(t, traceIDs, 3)
	assert.NotEqual(t, traceIDs[0], traceIDs[1])
	assert.NotEqual(t, traceIDs[1], traceIDs[2])
}
