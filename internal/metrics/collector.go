package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// QueueClient defines the interface for fetching queue metrics
type QueueClient interface {
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}

// Collector periodically collects and updates metrics
type Collector struct {
	metrics     *Metrics
	queueClient QueueClient
	queueURL    string
	logger      *slog.Logger
	stopCh      chan struct{}
}

// NewCollector creates a new metrics collector
func NewCollector(metrics *Metrics, queueClient QueueClient, queueURL string, logger *slog.Logger) *Collector {
	return &Collector{
		metrics:     metrics,
		queueClient: queueClient,
		queueURL:    queueURL,
		logger:      logger,
		stopCh:      make(chan struct{}),
	}
}

// Start begins collecting metrics at regular intervals
func (c *Collector) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Collect initial metrics
	c.collectOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.collectOnce(ctx)
		}
	}
}

// Stop stops the metrics collector
func (c *Collector) Stop() {
	close(c.stopCh)
}

// collectOnce performs a single collection cycle
func (c *Collector) collectOnce(ctx context.Context) {
	// Collect queue depth if queue client is available
	if c.queueClient != nil && c.queueURL != "" {
		c.collectQueueMetrics(ctx)
	}
}

// collectQueueMetrics fetches and updates queue depth metrics
func (c *Collector) collectQueueMetrics(ctx context.Context) {
	input := &sqs.GetQueueAttributesInput{
		QueueUrl:       &c.queueURL,
		AttributeNames: []types.QueueAttributeName{"ApproximateNumberOfMessages", "ApproximateNumberOfMessagesNotVisible"},
	}

	output, err := c.queueClient.GetQueueAttributes(ctx, input)
	if err != nil {
		c.logger.Error("failed to get queue attributes", "error", err)
		return
	}

	if output.Attributes != nil {
		// Get visible messages (ready to be processed)
		if messagesStr, ok := output.Attributes["ApproximateNumberOfMessages"]; ok {
			var messages float64
			if _, err := fmt.Sscanf(messagesStr, "%f", &messages); err == nil {
				c.metrics.SetQueueDepth("default", messages)
			}
		}

		// Optionally track in-flight messages
		if inFlightStr, ok := output.Attributes["ApproximateNumberOfMessagesNotVisible"]; ok {
			var inFlight float64
			if _, err := fmt.Sscanf(inFlightStr, "%f", &inFlight); err == nil {
				// Could add a separate metric for in-flight messages if needed
				c.logger.Debug("queue in-flight messages", "count", inFlight)
			}
		}
	}
}
