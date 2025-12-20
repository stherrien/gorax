package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	// Given: no existing metrics
	// When: creating new metrics
	m := NewMetrics()

	// Then: all metrics should be initialized
	assert.NotNil(t, m)
	assert.NotNil(t, m.WorkflowExecutionsTotal)
	assert.NotNil(t, m.WorkflowExecutionDuration)
	assert.NotNil(t, m.StepExecutionsTotal)
	assert.NotNil(t, m.StepExecutionDuration)
	assert.NotNil(t, m.QueueDepth)
	assert.NotNil(t, m.ActiveWorkers)
	assert.NotNil(t, m.HTTPRequestsTotal)
	assert.NotNil(t, m.HTTPRequestDuration)
}

func TestRegisterMetrics(t *testing.T) {
	// Given: new metrics
	m := NewMetrics()
	registry := prometheus.NewRegistry()

	// When: registering metrics
	err := m.Register(registry)

	// Then: registration should succeed
	assert.NoError(t, err)
}

func TestRegisterMetricsTwice(t *testing.T) {
	// Given: metrics already registered
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: attempting to register again
	err := m.Register(registry)

	// Then: registration should fail
	assert.Error(t, err)
}

func TestRecordWorkflowExecution(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: recording workflow execution
	m.RecordWorkflowExecution("tenant1", "workflow1", "completed", 1.5)

	// Then: metric should be recorded
	metrics, err := registry.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)

	// Find the counter metric
	found := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_workflow_executions_total" {
			found = true
			assert.Equal(t, 1, len(metric.GetMetric()))
		}
	}
	assert.True(t, found, "workflow executions counter should be present")
}

func TestRecordStepExecution(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: recording step execution
	m.RecordStepExecution("tenant1", "workflow1", "http", "completed", 0.5)

	// Then: metric should be recorded
	metrics, err := registry.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metrics)

	// Find the counter metric
	found := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_step_executions_total" {
			found = true
		}
	}
	assert.True(t, found, "step executions counter should be present")
}

func TestSetQueueDepth(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: setting queue depth
	m.SetQueueDepth("default", 42)

	// Then: gauge should be set
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	found := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_queue_depth" {
			found = true
			assert.Equal(t, 1, len(metric.GetMetric()))
			assert.Equal(t, float64(42), metric.GetMetric()[0].GetGauge().GetValue())
		}
	}
	assert.True(t, found, "queue depth gauge should be present")
}

func TestSetActiveWorkers(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: setting active workers
	m.SetActiveWorkers(5)

	// Then: gauge should be set
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	found := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_active_workers" {
			found = true
			assert.Equal(t, 1, len(metric.GetMetric()))
			assert.Equal(t, float64(5), metric.GetMetric()[0].GetGauge().GetValue())
		}
	}
	assert.True(t, found, "active workers gauge should be present")
}

func TestRecordHTTPRequest(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: recording HTTP request
	m.RecordHTTPRequest("GET", "/api/v1/workflows", "200", 0.1)

	// Then: metrics should be recorded
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	foundCounter := false
	foundHistogram := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_http_requests_total" {
			foundCounter = true
		}
		if metric.GetName() == "gorax_http_request_duration_seconds" {
			foundHistogram = true
		}
	}
	assert.True(t, foundCounter, "HTTP requests counter should be present")
	assert.True(t, foundHistogram, "HTTP request duration histogram should be present")
}
