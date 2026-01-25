package integration

import (
	"context"
	"maps"
	"sync"
	"time"
)

// Metrics collects metrics for integration operations.
type Metrics struct {
	executionCount      map[string]*ExecutionMetrics
	circuitBreakerStats map[string]*CircuitBreakerMetrics
	mu                  sync.RWMutex
}

// ExecutionMetrics tracks metrics for a specific integration.
type ExecutionMetrics struct {
	Name          string
	TotalCount    int64
	SuccessCount  int64
	ErrorCount    int64
	TotalDuration time.Duration
	LastExecution time.Time
	LastError     time.Time
	LastErrorMsg  string
}

// CircuitBreakerMetrics tracks circuit breaker metrics.
type CircuitBreakerMetrics struct {
	Name            string
	State           string
	OpenCount       int64
	ClosedCount     int64
	HalfOpenCount   int64
	LastStateChange time.Time
}

// NewMetrics creates a new metrics collector.
func NewMetrics() *Metrics {
	return &Metrics{
		executionCount:      make(map[string]*ExecutionMetrics),
		circuitBreakerStats: make(map[string]*CircuitBreakerMetrics),
	}
}

// RecordExecution records an integration execution.
func (m *Metrics) RecordExecution(name string, duration time.Duration, success bool, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics, exists := m.executionCount[name]
	if !exists {
		metrics = &ExecutionMetrics{Name: name}
		m.executionCount[name] = metrics
	}

	metrics.TotalCount++
	metrics.TotalDuration += duration
	metrics.LastExecution = time.Now()

	if success {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
		metrics.LastError = time.Now()
		metrics.LastErrorMsg = errMsg
	}
}

// RecordCircuitBreakerState records a circuit breaker state change.
func (m *Metrics) RecordCircuitBreakerState(name, state string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics, exists := m.circuitBreakerStats[name]
	if !exists {
		metrics = &CircuitBreakerMetrics{Name: name}
		m.circuitBreakerStats[name] = metrics
	}

	metrics.State = state
	metrics.LastStateChange = time.Now()

	switch state {
	case "open":
		metrics.OpenCount++
	case "closed":
		metrics.ClosedCount++
	case "half-open":
		metrics.HalfOpenCount++
	}
}

// GetExecutionMetrics returns metrics for a specific integration.
func (m *Metrics) GetExecutionMetrics(name string) *ExecutionMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.executionCount[name]
}

// GetAllExecutionMetrics returns metrics for all integrations.
func (m *Metrics) GetAllExecutionMetrics() map[string]*ExecutionMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*ExecutionMetrics, len(m.executionCount))
	maps.Copy(result, m.executionCount)
	return result
}

// GetCircuitBreakerMetrics returns circuit breaker metrics.
func (m *Metrics) GetCircuitBreakerMetrics(name string) *CircuitBreakerMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.circuitBreakerStats[name]
}

// Reset clears all metrics.
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executionCount = make(map[string]*ExecutionMetrics)
	m.circuitBreakerStats = make(map[string]*CircuitBreakerMetrics)
}

// AverageDuration returns the average execution duration for an integration.
func (em *ExecutionMetrics) AverageDuration() time.Duration {
	if em.TotalCount == 0 {
		return 0
	}
	return em.TotalDuration / time.Duration(em.TotalCount)
}

// SuccessRate returns the success rate as a percentage.
func (em *ExecutionMetrics) SuccessRate() float64 {
	if em.TotalCount == 0 {
		return 0
	}
	return float64(em.SuccessCount) / float64(em.TotalCount) * 100
}

// ErrorRate returns the error rate as a percentage.
func (em *ExecutionMetrics) ErrorRate() float64 {
	if em.TotalCount == 0 {
		return 0
	}
	return float64(em.ErrorCount) / float64(em.TotalCount) * 100
}

// MetricsCollector wraps an integration to collect metrics.
type MetricsCollector struct {
	integration Integration
	metrics     *Metrics
}

// NewMetricsCollector creates a new metrics-collecting wrapper.
func NewMetricsCollector(integration Integration, metrics *Metrics) *MetricsCollector {
	return &MetricsCollector{
		integration: integration,
		metrics:     metrics,
	}
}

// Name returns the integration name.
func (mc *MetricsCollector) Name() string {
	return mc.integration.Name()
}

// Type returns the integration type.
func (mc *MetricsCollector) Type() IntegrationType {
	return mc.integration.Type()
}

// Execute executes the integration and records metrics.
func (mc *MetricsCollector) Execute(ctx context.Context, config *Config, params JSONMap) (*Result, error) {
	start := time.Now()
	result, err := mc.integration.Execute(ctx, config, params)
	duration := time.Since(start)

	errMsg := ""
	success := err == nil && result != nil && result.Success
	if err != nil {
		errMsg = err.Error()
	} else if result != nil && !result.Success {
		errMsg = result.Error
	}

	mc.metrics.RecordExecution(mc.integration.Name(), duration, success, errMsg)
	return result, err
}

// Validate validates the integration configuration.
func (mc *MetricsCollector) Validate(config *Config) error {
	return mc.integration.Validate(config)
}

// GetSchema returns the integration schema.
func (mc *MetricsCollector) GetSchema() *Schema {
	return mc.integration.GetSchema()
}

// GetMetadata returns the integration metadata.
func (mc *MetricsCollector) GetMetadata() *Metadata {
	return mc.integration.GetMetadata()
}

// Unwrap returns the underlying integration.
func (mc *MetricsCollector) Unwrap() Integration {
	return mc.integration
}

// Global metrics instance.
var globalMetrics *Metrics
var globalMetricsOnce sync.Once

// GlobalMetrics returns the global metrics instance.
func GlobalMetrics() *Metrics {
	globalMetricsOnce.Do(func() {
		globalMetrics = NewMetrics()
	})
	return globalMetrics
}
