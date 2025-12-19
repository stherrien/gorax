package queue

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Consumer handles consuming messages from the queue
type Consumer struct {
	sqsClient *SQSClient
	logger    *slog.Logger
	handler   MessageHandler
	config    ConsumerConfig
	metrics   *ConsumerMetrics
	mu        sync.RWMutex
	running   bool
}

// ConsumerConfig holds consumer configuration
type ConsumerConfig struct {
	MaxMessages        int32         // Maximum number of messages to receive per poll (1-10)
	WaitTimeSeconds    int32         // Long polling wait time (0-20 seconds)
	VisibilityTimeout  int32         // Message visibility timeout in seconds
	MaxRetries         int           // Maximum number of retries before sending to DLQ
	ProcessTimeout     time.Duration // Maximum time to process a message
	PollInterval       time.Duration // Interval between polls when no messages received
	ConcurrentWorkers  int           // Number of concurrent message processors
	DeleteAfterProcess bool          // Automatically delete message after successful processing
}

// DefaultConsumerConfig returns default consumer configuration
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		MaxMessages:        10,
		WaitTimeSeconds:    20, // Long polling
		VisibilityTimeout:  30,
		MaxRetries:         3,
		ProcessTimeout:     5 * time.Minute,
		PollInterval:       1 * time.Second,
		ConcurrentWorkers:  10,
		DeleteAfterProcess: true,
	}
}

// ConsumerMetrics tracks consumer performance
type ConsumerMetrics struct {
	TotalReceived   int64
	TotalProcessed  int64
	TotalFailed     int64
	TotalDeleted    int64
	LastReceiveAt   time.Time
	LastProcessedAt time.Time
	InFlight        int64
}

// MessageHandler is a function that processes a message
type MessageHandler func(ctx context.Context, msg *ExecutionMessage) error

// NewConsumer creates a new queue consumer
func NewConsumer(sqsClient *SQSClient, handler MessageHandler, config ConsumerConfig, logger *slog.Logger) *Consumer {
	return &Consumer{
		sqsClient: sqsClient,
		logger:    logger,
		handler:   handler,
		config:    config,
		metrics:   &ConsumerMetrics{},
		running:   false,
	}
}

// Start begins consuming messages from the queue
func (c *Consumer) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("consumer already running")
	}
	c.running = true
	c.mu.Unlock()

	c.logger.Info("starting consumer",
		"max_messages", c.config.MaxMessages,
		"wait_time", c.config.WaitTimeSeconds,
		"concurrent_workers", c.config.ConcurrentWorkers,
	)

	// Create worker pool
	messagesChan := make(chan Message, c.config.ConcurrentWorkers*2)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < c.config.ConcurrentWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			c.worker(ctx, workerID, messagesChan)
		}(i)
	}

	// Main polling loop
	go func() {
		defer close(messagesChan)
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("consumer stopping due to context cancellation")
				c.mu.Lock()
				c.running = false
				c.mu.Unlock()
				return
			default:
				// Poll for messages
				messages, err := c.sqsClient.ReceiveMessages(ctx, c.config.MaxMessages, c.config.WaitTimeSeconds)
				if err != nil {
					c.logger.Error("failed to receive messages", "error", err)
					time.Sleep(c.config.PollInterval)
					continue
				}

				if len(messages) == 0 {
					// No messages, wait before next poll
					time.Sleep(c.config.PollInterval)
					continue
				}

				// Update metrics
				c.mu.Lock()
				c.metrics.TotalReceived += int64(len(messages))
				c.metrics.LastReceiveAt = time.Now()
				c.mu.Unlock()

				c.logger.Debug("received messages from queue", "count", len(messages))

				// Send messages to worker pool
				for _, msg := range messages {
					select {
					case messagesChan <- msg:
						c.mu.Lock()
						c.metrics.InFlight++
						c.mu.Unlock()
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	// Wait for all workers to finish
	wg.Wait()

	c.logger.Info("consumer stopped")
	return nil
}

// worker processes messages from the channel
func (c *Consumer) worker(ctx context.Context, workerID int, messagesChan <-chan Message) {
	c.logger.Debug("worker started", "worker_id", workerID)

	for msg := range messagesChan {
		c.processMessage(ctx, msg)

		// Decrement in-flight count
		c.mu.Lock()
		c.metrics.InFlight--
		c.mu.Unlock()
	}

	c.logger.Debug("worker stopped", "worker_id", workerID)
}

// processMessage processes a single message
func (c *Consumer) processMessage(ctx context.Context, msg Message) {
	// Create processing context with timeout
	processCtx, cancel := context.WithTimeout(ctx, c.config.ProcessTimeout)
	defer cancel()

	c.logger.Info("processing message",
		"message_id", msg.ID,
		"receive_count", msg.ApproximateReceiveCount,
	)

	// Unmarshal execution message
	execMsg, err := UnmarshalExecutionMessage(msg.Body)
	if err != nil {
		c.logger.Error("failed to unmarshal message", "error", err, "message_id", msg.ID)
		c.handleFailedMessage(ctx, msg, err)
		return
	}

	// Validate message
	if err := execMsg.Validate(); err != nil {
		c.logger.Error("invalid message", "error", err, "message_id", msg.ID)
		c.handleFailedMessage(ctx, msg, err)
		return
	}

	// Check if message has exceeded max retries
	if msg.ApproximateReceiveCount > c.config.MaxRetries {
		c.logger.Error("message exceeded max retries",
			"message_id", msg.ID,
			"receive_count", msg.ApproximateReceiveCount,
			"max_retries", c.config.MaxRetries,
		)
		// Delete the message to prevent infinite reprocessing
		// It should have already been sent to DLQ by SQS
		c.deleteMessage(ctx, msg.ReceiptHandle)
		return
	}

	// Process the message using handler
	if err := c.handler(processCtx, execMsg); err != nil {
		c.logger.Error("message processing failed",
			"error", err,
			"message_id", msg.ID,
			"execution_id", execMsg.ExecutionID,
		)
		c.handleFailedMessage(ctx, msg, err)
		return
	}

	// Update metrics
	c.mu.Lock()
	c.metrics.TotalProcessed++
	c.metrics.LastProcessedAt = time.Now()
	c.mu.Unlock()

	c.logger.Info("message processed successfully",
		"message_id", msg.ID,
		"execution_id", execMsg.ExecutionID,
	)

	// Delete message from queue if configured
	if c.config.DeleteAfterProcess {
		c.deleteMessage(ctx, msg.ReceiptHandle)
	}
}

// handleFailedMessage handles a failed message
func (c *Consumer) handleFailedMessage(ctx context.Context, msg Message, err error) {
	c.mu.Lock()
	c.metrics.TotalFailed++
	c.mu.Unlock()

	// If the message has not exceeded retries, it will become visible again
	// and be retried automatically by SQS
	// If it has exceeded retries, SQS will send it to the DLQ automatically

	c.logger.Warn("message processing failed, will retry",
		"message_id", msg.ID,
		"receive_count", msg.ApproximateReceiveCount,
		"error", err,
	)

	// Note: We don't delete the message here, so it will become visible again
	// after the visibility timeout expires
}

// deleteMessage deletes a message from the queue
func (c *Consumer) deleteMessage(ctx context.Context, receiptHandle string) {
	if err := c.sqsClient.DeleteMessage(ctx, receiptHandle); err != nil {
		c.logger.Error("failed to delete message", "error", err)
		return
	}

	c.mu.Lock()
	c.metrics.TotalDeleted++
	c.mu.Unlock()
}

// IsRunning returns whether the consumer is running
func (c *Consumer) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// GetMetrics returns consumer metrics
func (c *Consumer) GetMetrics() ConsumerMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return *c.metrics
}

// ResetMetrics resets consumer metrics
func (c *Consumer) ResetMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics.TotalReceived = 0
	c.metrics.TotalProcessed = 0
	c.metrics.TotalFailed = 0
	c.metrics.TotalDeleted = 0
	c.metrics.LastReceiveAt = time.Time{}
	c.metrics.LastProcessedAt = time.Time{}
	// Don't reset InFlight as it represents current state
}
