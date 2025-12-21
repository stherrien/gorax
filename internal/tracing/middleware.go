package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware returns middleware that traces HTTP requests
func HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Use otelhttp to automatically instrument HTTP handlers
		return otelhttp.NewHandler(next, "http.request",
			otelhttp.WithTracerProvider(otel.GetTracerProvider()),
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return r.Method + " " + r.URL.Path
			}),
			otelhttp.WithSpanOptions(
				trace.WithAttributes(
					semconv.HTTPMethod(""),
					semconv.HTTPTarget(""),
					semconv.HTTPScheme(""),
				),
			),
		)
	}
}

// HTTPClientMiddleware returns an HTTP client transport that traces outgoing requests
func HTTPClientMiddleware(transport http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(transport,
		otelhttp.WithTracerProvider(otel.GetTracerProvider()),
	)
}

// ExtractHTTPTraceContext extracts trace context from HTTP request headers
func ExtractHTTPTraceContext(r *http.Request) context.Context {
	carrier := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			carrier[key] = values[0]
		}
	}
	return ExtractTraceContext(r.Context(), carrier)
}

// InjectHTTPTraceContext injects trace context into HTTP request headers
func InjectHTTPTraceContext(r *http.Request) {
	carrier := make(map[string]string)
	InjectTraceContext(r.Context(), carrier)
	for key, value := range carrier {
		r.Header.Set(key, value)
	}
}

// AddHTTPAttributes adds standard HTTP attributes to a span
func AddHTTPAttributes(span trace.Span, r *http.Request, statusCode int) {
	span.SetAttributes(
		semconv.HTTPMethod(r.Method),
		semconv.HTTPTarget(r.URL.Path),
		semconv.HTTPScheme(r.URL.Scheme),
		semconv.HTTPStatusCode(statusCode),
		attribute.String("http.user_agent", r.UserAgent()),
	)

	if r.URL.RawQuery != "" {
		span.SetAttributes(attribute.String("http.query", r.URL.RawQuery))
	}

	if r.ContentLength > 0 {
		span.SetAttributes(attribute.Int64("http.request.content_length", r.ContentLength))
	}
}
