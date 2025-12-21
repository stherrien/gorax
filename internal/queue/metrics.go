package queue

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// MetricsCollector collects queue metrics
type MetricsCollector struct {
	sqsClient *SQSClient
	logger    *slog.Logger
	metrics   *QueueMetrics
	mu        sync.RWMutex
	stopChan  chan struct{}
	interval  time.Duration
}

// QueueMetrics represents queue metrics
type QueueMetrics struct {
	// Main queue metrics
	ApproximateNumberOfMessages           int
	ApproximateNumberOfMessagesNotVisible int
	ApproximateNumberOfMessagesDelayed    int

	// Dead-letter queue metrics
	DLQApproximateNumberOfMessages int

	// Calculated metrics
	TotalMessages int // Messages + NotVisible + Delayed
	QueueDepth    int // Same as TotalMessages for backward compatibility

	// Timestamp
	LastUpdated time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(sqsClient *SQSClient, interval time.Duration, logger *slog.Logger) *MetricsCollector {
	if interval <= 0 {
		interval = 30 * time.Second // Default to 30 seconds
	}

	return &MetricsCollector{
		sqsClient: sqsClient,
		logger:    logger,
		metrics:   &QueueMetrics{},
		stopChan:  make(chan struct{}),
		interval:  interval,
	}
}

// Start begins collecting metrics
func (c *MetricsCollector) Start(ctx context.Context) error {
	c.logger.Info("starting metrics collector", "interval", c.interval)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Collect initial metrics
	if err := c.collectMetrics(ctx); err != nil {
		c.logger.Error("failed to collect initial metrics", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("metrics collector stopping")
			return ctx.Err()
		case <-c.stopChan:
			c.logger.Info("metrics collector stopped")
			return nil
		case <-ticker.C:
			if err := c.collectMetrics(ctx); err != nil {
				c.logger.Error("failed to collect metrics", "error", err)
			}
		}
	}
}

// Stop stops the metrics collector
func (c *MetricsCollector) Stop() {
	close(c.stopChan)
}

// collectMetrics collects metrics from SQS
func (c *MetricsCollector) collectMetrics(ctx context.Context) error {
	// Get main queue attributes
	attrs, err := c.sqsClient.GetQueueAttributes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get queue attributes: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.ApproximateNumberOfMessages = attrs.ApproximateNumberOfMessages
	c.metrics.ApproximateNumberOfMessagesNotVisible = attrs.ApproximateNumberOfMessagesNotVisible
	c.metrics.ApproximateNumberOfMessagesDelayed = attrs.ApproximateNumberOfMessagesDelayed

	// Calculate total messages
	c.metrics.TotalMessages = attrs.ApproximateNumberOfMessages +
		attrs.ApproximateNumberOfMessagesNotVisible +
		attrs.ApproximateNumberOfMessagesDelayed
	c.metrics.QueueDepth = c.metrics.TotalMessages

	// Get DLQ attributes if configured
	dlqAttrs, err := c.sqsClient.GetDLQAttributes(ctx)
	if err == nil {
		c.metrics.DLQApproximateNumberOfMessages = dlqAttrs.ApproximateNumberOfMessages
	} else {
		// DLQ might not be configured, don't log as error
		c.logger.Debug("DLQ metrics not available", "error", err)
	}

	c.metrics.LastUpdated = time.Now()

	c.logger.Debug("metrics collected",
		"visible", c.metrics.ApproximateNumberOfMessages,
		"not_visible", c.metrics.ApproximateNumberOfMessagesNotVisible,
		"delayed", c.metrics.ApproximateNumberOfMessagesDelayed,
		"total", c.metrics.TotalMessages,
		"dlq", c.metrics.DLQApproximateNumberOfMessages,
	)

	return nil
}

// GetMetrics returns the current metrics
func (c *MetricsCollector) GetMetrics() QueueMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return *c.metrics
}

// GetQueueDepth returns the current queue depth (total messages)
func (c *MetricsCollector) GetQueueDepth() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metrics.QueueDepth
}

// GetDLQDepth returns the dead-letter queue depth
func (c *MetricsCollector) GetDLQDepth() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metrics.DLQApproximateNumberOfMessages
}

// IsHealthy returns true if the queue is healthy
// A queue is considered unhealthy if:
// - DLQ has messages (indicating failures)
// - Queue depth is extremely high (configurable threshold)
func (c *MetricsCollector) IsHealthy(maxQueueDepth int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check DLQ - any messages indicate issues
	if c.metrics.DLQApproximateNumberOfMessages > 0 {
		return false
	}

	// Check queue depth if threshold is configured
	if maxQueueDepth > 0 && c.metrics.QueueDepth > maxQueueDepth {
		return false
	}

	return true
}

// GetHealthStatus returns a detailed health status
func (c *MetricsCollector) GetHealthStatus(maxQueueDepth int) HealthStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := HealthStatus{
		Healthy:     true,
		QueueDepth:  c.metrics.QueueDepth,
		DLQDepth:    c.metrics.DLQApproximateNumberOfMessages,
		LastChecked: c.metrics.LastUpdated,
	}

	// Check DLQ
	if c.metrics.DLQApproximateNumberOfMessages > 0 {
		status.Healthy = false
		status.Issues = append(status.Issues, fmt.Sprintf("DLQ contains %d messages", c.metrics.DLQApproximateNumberOfMessages))
	}

	// Check queue depth
	if maxQueueDepth > 0 && c.metrics.QueueDepth > maxQueueDepth {
		status.Healthy = false
		status.Issues = append(status.Issues, fmt.Sprintf("Queue depth %d exceeds threshold %d", c.metrics.QueueDepth, maxQueueDepth))
	}

	// Check if metrics are stale (haven't been updated in 5 minutes)
	if time.Since(c.metrics.LastUpdated) > 5*time.Minute {
		status.Healthy = false
		status.Issues = append(status.Issues, "Metrics are stale")
	}

	return status
}

// HealthStatus represents the health status of the queue
type HealthStatus struct {
	Healthy     bool
	QueueDepth  int
	DLQDepth    int
	LastChecked time.Time
	Issues      []string
}
