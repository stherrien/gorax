package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestHTTPMetricsMiddleware(t *testing.T) {
	// Given: metrics and middleware
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	handler := HTTPMetricsMiddleware(m)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	// When: making HTTP request
	req := httptest.NewRequest("GET", "/api/v1/workflows", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Then: request should succeed
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())

	// And: metrics should be recorded
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	foundCounter := false
	foundHistogram := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_http_requests_total" {
			foundCounter = true
			assert.Equal(t, 1, len(metric.GetMetric()))
		}
		if metric.GetName() == "gorax_http_request_duration_seconds" {
			foundHistogram = true
			assert.Equal(t, 1, len(metric.GetMetric()))
		}
	}
	assert.True(t, foundCounter, "HTTP requests counter should be present")
	assert.True(t, foundHistogram, "HTTP request duration histogram should be present")
}

func TestHTTPMetricsMiddlewareWithError(t *testing.T) {
	// Given: metrics and middleware
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	handler := HTTPMetricsMiddleware(m)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))

	// When: making HTTP request that returns error
	req := httptest.NewRequest("POST", "/api/v1/workflows", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Then: error response should be returned
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// And: metrics should record the error status
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	found := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_http_requests_total" {
			found = true
			// Verify the status label is "500"
			labels := metric.GetMetric()[0].GetLabel()
			for _, label := range labels {
				if label.GetName() == "status" {
					assert.Equal(t, "500", label.GetValue())
				}
			}
		}
	}
	assert.True(t, found, "HTTP requests counter should be present")
}

func TestNormalizeHTTPPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "static path",
			path:     "/api/v1/workflows",
			expected: "/api/v1/workflows",
		},
		{
			name:     "path with UUID",
			path:     "/api/v1/workflows/123e4567-e89b-12d3-a456-426614174000",
			expected: "/api/v1/workflows/:id",
		},
		{
			name:     "path with multiple UUIDs",
			path:     "/api/v1/workflows/123e4567-e89b-12d3-a456-426614174000/executions/223e4567-e89b-12d3-a456-426614174000",
			expected: "/api/v1/workflows/:id/executions/:id",
		},
		{
			name:     "path with numeric ID",
			path:     "/api/v1/tenants/42/usage",
			expected: "/api/v1/tenants/:id/usage",
		},
		{
			name:     "webhook path",
			path:     "/webhooks/wf-123/hook-456",
			expected: "/webhooks/:id/:id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeHTTPPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseWriter(t *testing.T) {
	// Given: a response writer
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	// When: writing response
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte("test"))

	// Then: status code should be captured
	assert.Equal(t, http.StatusCreated, rw.statusCode)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "test", rec.Body.String())
}

func TestResponseWriterDefaultStatus(t *testing.T) {
	// Given: a response writer that doesn't call WriteHeader
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	// When: only writing body
	rw.Write([]byte("test"))

	// Then: default status should be OK
	assert.Equal(t, http.StatusOK, rw.statusCode)
	assert.Equal(t, "test", rec.Body.String())
}
