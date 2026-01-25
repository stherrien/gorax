package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// Workflow metrics
	WorkflowExecutionsTotal   *prometheus.CounterVec
	WorkflowExecutionDuration *prometheus.HistogramVec
	WorkflowExecutionsActive  *prometheus.GaugeVec

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

	// Formula evaluation metrics
	FormulaEvaluationsTotal   *prometheus.CounterVec
	FormulaEvaluationDuration *prometheus.HistogramVec
	FormulaCacheHitsTotal     *prometheus.CounterVec
	FormulaCacheMissesTotal   *prometheus.CounterVec

	// Database metrics
	DBConnectionsOpen  *prometheus.GaugeVec
	DBConnectionsIdle  *prometheus.GaugeVec
	DBConnectionsInUse *prometheus.GaugeVec
	DBQueryDuration    *prometheus.HistogramVec
	DBQueriesTotal     *prometheus.CounterVec
}

// NewMetrics creates a new Metrics instance with all collectors initialized
func NewMetrics() *Metrics {
	return &Metrics{
		WorkflowExecutionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorax_workflow_executions_total",
				Help: "Total number of workflow executions by status and trigger type",
			},
			[]string{"tenant_id", "workflow_id", "trigger_type", "status"},
		),
		WorkflowExecutionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gorax_workflow_execution_duration_seconds",
				Help:    "Workflow execution duration in seconds by trigger type",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
			},
			[]string{"tenant_id", "workflow_id", "trigger_type"},
		),
		WorkflowExecutionsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gorax_workflow_executions_active",
				Help: "Number of currently active workflow executions",
			},
			[]string{"tenant_id", "workflow_id", "trigger_type"},
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
		FormulaEvaluationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorax_formula_evaluations_total",
				Help: "Total number of formula evaluations by status",
			},
			[]string{"status"},
		),
		FormulaEvaluationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gorax_formula_evaluation_duration_seconds",
				Help:    "Formula evaluation duration in seconds",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
			},
			[]string{},
		),
		FormulaCacheHitsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorax_formula_cache_hits_total",
				Help: "Total number of formula cache hits",
			},
			[]string{},
		),
		FormulaCacheMissesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorax_formula_cache_misses_total",
				Help: "Total number of formula cache misses",
			},
			[]string{},
		),
		DBConnectionsOpen: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gorax_db_connections_open",
				Help: "Number of open database connections",
			},
			[]string{"pool"},
		),
		DBConnectionsIdle: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gorax_db_connections_idle",
				Help: "Number of idle database connections",
			},
			[]string{"pool"},
		),
		DBConnectionsInUse: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gorax_db_connections_in_use",
				Help: "Number of database connections in use",
			},
			[]string{"pool"},
		),
		DBQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gorax_db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
			},
			[]string{"operation", "table"},
		),
		DBQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gorax_db_queries_total",
				Help: "Total number of database queries by operation and status",
			},
			[]string{"operation", "table", "status"},
		),
	}
}

// Register registers all metrics with the provided registry
func (m *Metrics) Register(registry *prometheus.Registry) error {
	collectors := []prometheus.Collector{
		m.WorkflowExecutionsTotal,
		m.WorkflowExecutionDuration,
		m.WorkflowExecutionsActive,
		m.StepExecutionsTotal,
		m.StepExecutionDuration,
		m.QueueDepth,
		m.ActiveWorkers,
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.FormulaEvaluationsTotal,
		m.FormulaEvaluationDuration,
		m.FormulaCacheHitsTotal,
		m.FormulaCacheMissesTotal,
		m.DBConnectionsOpen,
		m.DBConnectionsIdle,
		m.DBConnectionsInUse,
		m.DBQueryDuration,
		m.DBQueriesTotal,
	}

	for _, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return err
		}
	}

	return nil
}

// RecordWorkflowExecution records a workflow execution with status and duration
func (m *Metrics) RecordWorkflowExecution(tenantID, workflowID, triggerType, status string, durationSeconds float64) {
	m.WorkflowExecutionsTotal.WithLabelValues(tenantID, workflowID, triggerType, status).Inc()
	m.WorkflowExecutionDuration.WithLabelValues(tenantID, workflowID, triggerType).Observe(durationSeconds)
}

// IncActiveWorkflowExecutions increments the active workflow executions gauge
func (m *Metrics) IncActiveWorkflowExecutions(tenantID, workflowID, triggerType string) {
	m.WorkflowExecutionsActive.WithLabelValues(tenantID, workflowID, triggerType).Inc()
}

// DecActiveWorkflowExecutions decrements the active workflow executions gauge
func (m *Metrics) DecActiveWorkflowExecutions(tenantID, workflowID, triggerType string) {
	m.WorkflowExecutionsActive.WithLabelValues(tenantID, workflowID, triggerType).Dec()
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

// RecordFormulaEvaluation records a formula evaluation with status and duration
func (m *Metrics) RecordFormulaEvaluation(status string, durationSeconds float64) {
	m.FormulaEvaluationsTotal.WithLabelValues(status).Inc()
	m.FormulaEvaluationDuration.WithLabelValues().Observe(durationSeconds)
}

// RecordFormulaCacheHit records a cache hit
func (m *Metrics) RecordFormulaCacheHit() {
	m.FormulaCacheHitsTotal.WithLabelValues().Inc()
}

// RecordFormulaCacheMiss records a cache miss
func (m *Metrics) RecordFormulaCacheMiss() {
	m.FormulaCacheMissesTotal.WithLabelValues().Inc()
}

// SetDBConnectionPoolStats sets database connection pool statistics
func (m *Metrics) SetDBConnectionPoolStats(poolName string, open, idle, inUse int) {
	m.DBConnectionsOpen.WithLabelValues(poolName).Set(float64(open))
	m.DBConnectionsIdle.WithLabelValues(poolName).Set(float64(idle))
	m.DBConnectionsInUse.WithLabelValues(poolName).Set(float64(inUse))
}

// RecordDBQuery records a database query with operation, table, status, and duration
func (m *Metrics) RecordDBQuery(operation, table, status string, durationSeconds float64) {
	m.DBQueriesTotal.WithLabelValues(operation, table, status).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(durationSeconds)
}
