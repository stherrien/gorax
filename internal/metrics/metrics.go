package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// Workflow metrics
	WorkflowExecutionsTotal   *prometheus.CounterVec
	WorkflowExecutionDuration *prometheus.HistogramVec

	// Step metrics
	StepExecutionsTotal   *prometheus.CounterVec
	StepExecutionDuration *prometheus.HistogramVec

	// Queue metrics
	QueueDepth *prometheus.GaugeVec

	// Worker metrics
	ActiveWorkers prometheus.Gauge

	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
}

// NewMetrics creates a new Metrics instance with all collectors initialized
func NewMetrics() *Metrics {
	return &Metrics{
		WorkflowExecutionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorax_workflow_executions_total",
				Help: "Total number of workflow executions by status",
			},
			[]string{"tenant_id", "workflow_id", "status"},
		),
		WorkflowExecutionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gorax_workflow_execution_duration_seconds",
				Help:    "Workflow execution duration in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
			},
			[]string{"tenant_id", "workflow_id"},
		),
		StepExecutionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorax_step_executions_total",
				Help: "Total number of step executions by type and status",
			},
			[]string{"tenant_id", "workflow_id", "step_type", "status"},
		),
		StepExecutionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gorax_step_execution_duration_seconds",
				Help:    "Step execution duration in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"tenant_id", "workflow_id", "step_type"},
		),
		QueueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gorax_queue_depth",
				Help: "Current queue depth by queue name",
			},
			[]string{"queue"},
		),
		ActiveWorkers: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "gorax_active_workers",
				Help: "Number of active workers processing jobs",
			},
		),
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorax_http_requests_total",
				Help: "Total number of HTTP requests by method, path, and status",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gorax_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}
}

// Register registers all metrics with the provided registry
func (m *Metrics) Register(registry *prometheus.Registry) error {
	collectors := []prometheus.Collector{
		m.WorkflowExecutionsTotal,
		m.WorkflowExecutionDuration,
		m.StepExecutionsTotal,
		m.StepExecutionDuration,
		m.QueueDepth,
		m.ActiveWorkers,
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return err
		}
	}

	return nil
}

// RecordWorkflowExecution records a workflow execution with status and duration
func (m *Metrics) RecordWorkflowExecution(tenantID, workflowID, status string, durationSeconds float64) {
	m.WorkflowExecutionsTotal.WithLabelValues(tenantID, workflowID, status).Inc()
	m.WorkflowExecutionDuration.WithLabelValues(tenantID, workflowID).Observe(durationSeconds)
}

// RecordStepExecution records a step execution with type, status, and duration
func (m *Metrics) RecordStepExecution(tenantID, workflowID, stepType, status string, durationSeconds float64) {
	m.StepExecutionsTotal.WithLabelValues(tenantID, workflowID, stepType, status).Inc()
	m.StepExecutionDuration.WithLabelValues(tenantID, workflowID, stepType).Observe(durationSeconds)
}

// SetQueueDepth sets the current queue depth for a given queue
func (m *Metrics) SetQueueDepth(queueName string, depth float64) {
	m.QueueDepth.WithLabelValues(queueName).Set(depth)
}

// SetActiveWorkers sets the number of active workers
func (m *Metrics) SetActiveWorkers(count float64) {
	m.ActiveWorkers.Set(count)
}

// RecordHTTPRequest records an HTTP request with method, path, status, and duration
func (m *Metrics) RecordHTTPRequest(method, path, status string, durationSeconds float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(durationSeconds)
}
