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
	assert.NotNil(t, m.WorkflowExecutionsActive)
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
	m.RecordWorkflowExecution("tenant1", "workflow1", "webhook", "completed", 1.5)

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

func TestRecordFormulaEvaluation(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: recording formula evaluation
	m.RecordFormulaEvaluation("success", 0.001)

	// Then: metrics should be recorded
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	foundCounter := false
	foundHistogram := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_formula_evaluations_total" {
			foundCounter = true
		}
		if metric.GetName() == "gorax_formula_evaluation_duration_seconds" {
			foundHistogram = true
		}
	}
	assert.True(t, foundCounter, "formula evaluations counter should be present")
	assert.True(t, foundHistogram, "formula evaluation duration histogram should be present")
}

func TestRecordFormulaCacheHit(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: recording cache hit
	m.RecordFormulaCacheHit()

	// Then: metric should be recorded
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	found := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_formula_cache_hits_total" {
			found = true
		}
	}
	assert.True(t, found, "formula cache hits counter should be present")
}

func TestRecordFormulaCacheMiss(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: recording cache miss
	m.RecordFormulaCacheMiss()

	// Then: metric should be recorded
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	found := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_formula_cache_misses_total" {
			found = true
		}
	}
	assert.True(t, found, "formula cache misses counter should be present")
}

func TestSetDBConnectionPoolStats(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: setting connection pool stats
	m.SetDBConnectionPoolStats("main", 10, 5, 3)

	// Then: gauges should be set
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	foundOpen := false
	foundIdle := false
	foundInUse := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_db_connections_open" {
			foundOpen = true
			assert.Equal(t, 1, len(metric.GetMetric()))
			assert.Equal(t, float64(10), metric.GetMetric()[0].GetGauge().GetValue())
		}
		if metric.GetName() == "gorax_db_connections_idle" {
			foundIdle = true
			assert.Equal(t, 1, len(metric.GetMetric()))
			assert.Equal(t, float64(5), metric.GetMetric()[0].GetGauge().GetValue())
		}
		if metric.GetName() == "gorax_db_connections_in_use" {
			foundInUse = true
			assert.Equal(t, 1, len(metric.GetMetric()))
			assert.Equal(t, float64(3), metric.GetMetric()[0].GetGauge().GetValue())
		}
	}
	assert.True(t, foundOpen, "db connections open gauge should be present")
	assert.True(t, foundIdle, "db connections idle gauge should be present")
	assert.True(t, foundInUse, "db connections in use gauge should be present")
}

func TestRecordDBQuery(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: recording database query
	m.RecordDBQuery("SELECT", "workflows", "success", 0.05)

	// Then: metrics should be recorded
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	foundCounter := false
	foundHistogram := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_db_queries_total" {
			foundCounter = true
		}
		if metric.GetName() == "gorax_db_query_duration_seconds" {
			foundHistogram = true
		}
	}
	assert.True(t, foundCounter, "db queries counter should be present")
	assert.True(t, foundHistogram, "db query duration histogram should be present")
}

func TestIncActiveWorkflowExecutions(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: incrementing active workflow executions
	m.IncActiveWorkflowExecutions("tenant1", "workflow1", "webhook")

	// Then: gauge should be set to 1
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	found := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_workflow_executions_active" {
			found = true
			assert.Equal(t, 1, len(metric.GetMetric()))
			assert.Equal(t, float64(1), metric.GetMetric()[0].GetGauge().GetValue())
		}
	}
	assert.True(t, found, "workflow executions active gauge should be present")
}

func TestDecActiveWorkflowExecutions(t *testing.T) {
	// Given: metrics initialized with active execution
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)
	m.IncActiveWorkflowExecutions("tenant1", "workflow1", "webhook")

	// When: decrementing active workflow executions
	m.DecActiveWorkflowExecutions("tenant1", "workflow1", "webhook")

	// Then: gauge should be back to 0
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	found := false
	for _, metric := range metrics {
		if metric.GetName() == "gorax_workflow_executions_active" {
			found = true
			assert.Equal(t, 1, len(metric.GetMetric()))
			assert.Equal(t, float64(0), metric.GetMetric()[0].GetGauge().GetValue())
		}
	}
	assert.True(t, found, "workflow executions active gauge should be present")
}

func TestActiveWorkflowExecutionsMultipleWorkflows(t *testing.T) {
	// Given: metrics initialized
	m := NewMetrics()
	registry := prometheus.NewRegistry()
	m.Register(registry)

	// When: starting multiple workflow executions
	m.IncActiveWorkflowExecutions("tenant1", "workflow1", "webhook")
	m.IncActiveWorkflowExecutions("tenant1", "workflow2", "schedule")
	m.IncActiveWorkflowExecutions("tenant2", "workflow1", "webhook")

	// Then: should have 3 separate gauge metrics
	metrics, err := registry.Gather()
	assert.NoError(t, err)

	for _, metric := range metrics {
		if metric.GetName() == "gorax_workflow_executions_active" {
			assert.Equal(t, 3, len(metric.GetMetric()), "should have 3 separate metrics for different label combinations")
		}
	}
}
