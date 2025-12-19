package queue

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Publisher handles publishing messages to the queue
type Publisher struct {
	sqsClient *SQSClient
	logger    *slog.Logger
	metrics   *PublisherMetrics
}

// PublisherMetrics tracks publisher performance
type PublisherMetrics struct {
	TotalPublished int64
	TotalFailed    int64
	LastPublishAt  time.Time
}

// NewPublisher creates a new queue publisher
func NewPublisher(sqsClient *SQSClient, logger *slog.Logger) *Publisher {
	return &Publisher{
		sqsClient: sqsClient,
		logger:    logger,
		metrics:   &PublisherMetrics{},
	}
}

// PublishExecution publishes a workflow execution message to the queue
func (p *Publisher) PublishExecution(ctx context.Context, msg *ExecutionMessage) error {
	// Validate message
	if err := msg.Validate(); err != nil {
		p.logger.Error("invalid execution message", "error", err)
		return fmt.Errorf("invalid message: %w", err)
	}

	// Marshal message to JSON
	body, err := msg.Marshal()
	if err != nil {
		p.logger.Error("failed to marshal message", "error", err)
		p.metrics.TotalFailed++
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Get message attributes
	attributes := msg.GetMessageAttributes()

	// Send message to SQS
	messageID, err := p.sqsClient.SendMessage(ctx, body, attributes)
	if err != nil {
		p.logger.Error("failed to publish execution message",
			"error", err,
			"execution_id", msg.ExecutionID,
			"workflow_id", msg.WorkflowID,
		)
		p.metrics.TotalFailed++
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Update metrics
	p.metrics.TotalPublished++
	p.metrics.LastPublishAt = time.Now()

	p.logger.Info("execution message published",
		"execution_id", msg.ExecutionID,
		"workflow_id", msg.WorkflowID,
		"tenant_id", msg.TenantID,
		"message_id", *messageID,
	)

	return nil
}

// PublishExecutionBatch publishes multiple workflow execution messages to the queue
func (p *Publisher) PublishExecutionBatch(ctx context.Context, messages []*ExecutionMessage) error {
	if len(messages) == 0 {
		return nil
	}

	if len(messages) > 10 {
		return fmt.Errorf("batch size cannot exceed 10 messages")
	}

	// Validate and prepare batch messages
	batchMessages := make([]BatchMessage, 0, len(messages))
	for _, msg := range messages {
		// Validate message
		if err := msg.Validate(); err != nil {
			p.logger.Error("invalid execution message in batch", "error", err, "execution_id", msg.ExecutionID)
			return fmt.Errorf("invalid message in batch: %w", err)
		}

		// Marshal message to JSON
		body, err := msg.Marshal()
		if err != nil {
			p.logger.Error("failed to marshal message in batch", "error", err, "execution_id", msg.ExecutionID)
			p.metrics.TotalFailed++
			return fmt.Errorf("failed to marshal message in batch: %w", err)
		}

		batchMessages = append(batchMessages, BatchMessage{
			Body:       body,
			Attributes: msg.GetMessageAttributes(),
		})
	}

	// Send batch to SQS
	if err := p.sqsClient.SendMessageBatch(ctx, batchMessages); err != nil {
		p.logger.Error("failed to publish execution batch", "error", err, "count", len(messages))
		p.metrics.TotalFailed += int64(len(messages))
		return fmt.Errorf("failed to publish batch: %w", err)
	}

	// Update metrics
	p.metrics.TotalPublished += int64(len(messages))
	p.metrics.LastPublishAt = time.Now()

	p.logger.Info("execution batch published", "count", len(messages))

	return nil
}

// GetMetrics returns publisher metrics
func (p *Publisher) GetMetrics() PublisherMetrics {
	return *p.metrics
}

// ResetMetrics resets publisher metrics
func (p *Publisher) ResetMetrics() {
	p.metrics.TotalPublished = 0
	p.metrics.TotalFailed = 0
	p.metrics.LastPublishAt = time.Time{}
}
