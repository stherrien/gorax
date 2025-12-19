package queue

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// SQSClient wraps AWS SQS functionality
type SQSClient struct {
	client   *sqs.Client
	queueURL string
	dlqURL   string
	logger   *slog.Logger
}

// SQSConfig holds configuration for SQS client
type SQSConfig struct {
	QueueURL        string
	DLQueueURL      string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string // For LocalStack or custom endpoints
}

// NewSQSClient creates a new SQS client
func NewSQSClient(ctx context.Context, cfg SQSConfig, logger *slog.Logger) (*SQSClient, error) {
	if cfg.QueueURL == "" {
		return nil, fmt.Errorf("queue URL is required")
	}

	// Build AWS config options
	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(cfg.Region))

	// Use static credentials if provided (for LocalStack)
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	// Load AWS configuration
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create SQS client with custom endpoint if provided (for LocalStack)
	var clientOpts []func(*sqs.Options)
	if cfg.Endpoint != "" {
		clientOpts = append(clientOpts, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
		logger.Info("using custom SQS endpoint", "endpoint", cfg.Endpoint)
	}

	client := sqs.NewFromConfig(awsCfg, clientOpts...)

	logger.Info("SQS client initialized",
		"queue_url", cfg.QueueURL,
		"dlq_url", cfg.DLQueueURL,
		"region", cfg.Region,
	)

	return &SQSClient{
		client:   client,
		queueURL: cfg.QueueURL,
		dlqURL:   cfg.DLQueueURL,
		logger:   logger,
	}, nil
}

// SendMessage sends a message to the SQS queue
func (c *SQSClient) SendMessage(ctx context.Context, messageBody string, attributes map[string]string) (*string, error) {
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(messageBody),
	}

	// Add message attributes if provided
	if len(attributes) > 0 {
		msgAttrs := make(map[string]types.MessageAttributeValue)
		for key, value := range attributes {
			msgAttrs[key] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(value),
			}
		}
		input.MessageAttributes = msgAttrs
	}

	result, err := c.client.SendMessage(ctx, input)
	if err != nil {
		c.logger.Error("failed to send message to SQS", "error", err)
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	c.logger.Debug("message sent to SQS", "message_id", *result.MessageId)
	return result.MessageId, nil
}

// SendMessageBatch sends multiple messages to the SQS queue in a single request
func (c *SQSClient) SendMessageBatch(ctx context.Context, messages []BatchMessage) error {
	if len(messages) == 0 {
		return nil
	}

	if len(messages) > 10 {
		return fmt.Errorf("batch size cannot exceed 10 messages")
	}

	entries := make([]types.SendMessageBatchRequestEntry, 0, len(messages))
	for i, msg := range messages {
		entry := types.SendMessageBatchRequestEntry{
			Id:          aws.String(fmt.Sprintf("msg-%d", i)),
			MessageBody: aws.String(msg.Body),
		}

		// Add message attributes if provided
		if len(msg.Attributes) > 0 {
			msgAttrs := make(map[string]types.MessageAttributeValue)
			for key, value := range msg.Attributes {
				msgAttrs[key] = types.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(value),
				}
			}
			entry.MessageAttributes = msgAttrs
		}

		entries = append(entries, entry)
	}

	result, err := c.client.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(c.queueURL),
		Entries:  entries,
	})
	if err != nil {
		c.logger.Error("failed to send batch messages to SQS", "error", err)
		return fmt.Errorf("failed to send batch messages: %w", err)
	}

	// Log any failed messages
	if len(result.Failed) > 0 {
		c.logger.Warn("some messages failed to send", "count", len(result.Failed))
		for _, failed := range result.Failed {
			c.logger.Error("message send failed",
				"id", *failed.Id,
				"code", *failed.Code,
				"message", *failed.Message,
			)
		}
		return fmt.Errorf("failed to send %d messages", len(result.Failed))
	}

	c.logger.Debug("batch messages sent to SQS", "count", len(result.Successful))
	return nil
}

// ReceiveMessages receives messages from the SQS queue
func (c *SQSClient) ReceiveMessages(ctx context.Context, maxMessages int32, waitTimeSeconds int32) ([]Message, error) {
	if maxMessages <= 0 || maxMessages > 10 {
		maxMessages = 10
	}

	if waitTimeSeconds < 0 || waitTimeSeconds > 20 {
		waitTimeSeconds = 20 // Maximum long polling time
	}

	result, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(c.queueURL),
		MaxNumberOfMessages:   maxMessages,
		WaitTimeSeconds:       waitTimeSeconds,
		MessageAttributeNames: []string{"All"},
		AttributeNames:        []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if err != nil {
		c.logger.Error("failed to receive messages from SQS", "error", err)
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	if len(result.Messages) == 0 {
		return nil, nil
	}

	messages := make([]Message, 0, len(result.Messages))
	for _, msg := range result.Messages {
		attributes := make(map[string]string)
		for key, value := range msg.MessageAttributes {
			if value.StringValue != nil {
				attributes[key] = *value.StringValue
			}
		}

		messages = append(messages, Message{
			ID:              *msg.MessageId,
			Body:            *msg.Body,
			ReceiptHandle:   *msg.ReceiptHandle,
			Attributes:      attributes,
			ApproximateReceiveCount: getApproximateReceiveCount(msg.Attributes),
		})
	}

	c.logger.Debug("received messages from SQS", "count", len(messages))
	return messages, nil
}

// DeleteMessage deletes a message from the queue
func (c *SQSClient) DeleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		c.logger.Error("failed to delete message from SQS", "error", err)
		return fmt.Errorf("failed to delete message: %w", err)
	}

	c.logger.Debug("message deleted from SQS")
	return nil
}

// DeleteMessageBatch deletes multiple messages from the queue
func (c *SQSClient) DeleteMessageBatch(ctx context.Context, receiptHandles []string) error {
	if len(receiptHandles) == 0 {
		return nil
	}

	if len(receiptHandles) > 10 {
		return fmt.Errorf("batch size cannot exceed 10 messages")
	}

	entries := make([]types.DeleteMessageBatchRequestEntry, 0, len(receiptHandles))
	for i, handle := range receiptHandles {
		entries = append(entries, types.DeleteMessageBatchRequestEntry{
			Id:            aws.String(fmt.Sprintf("msg-%d", i)),
			ReceiptHandle: aws.String(handle),
		})
	}

	result, err := c.client.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
		QueueUrl: aws.String(c.queueURL),
		Entries:  entries,
	})
	if err != nil {
		c.logger.Error("failed to delete batch messages from SQS", "error", err)
		return fmt.Errorf("failed to delete batch messages: %w", err)
	}

	// Log any failed deletions
	if len(result.Failed) > 0 {
		c.logger.Warn("some messages failed to delete", "count", len(result.Failed))
		for _, failed := range result.Failed {
			c.logger.Error("message deletion failed",
				"id", *failed.Id,
				"code", *failed.Code,
				"message", *failed.Message,
			)
		}
		return fmt.Errorf("failed to delete %d messages", len(result.Failed))
	}

	c.logger.Debug("batch messages deleted from SQS", "count", len(result.Successful))
	return nil
}

// ChangeMessageVisibility changes the visibility timeout of a message
func (c *SQSClient) ChangeMessageVisibility(ctx context.Context, receiptHandle string, visibilityTimeout int32) error {
	_, err := c.client.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(c.queueURL),
		ReceiptHandle:     aws.String(receiptHandle),
		VisibilityTimeout: visibilityTimeout,
	})
	if err != nil {
		c.logger.Error("failed to change message visibility", "error", err)
		return fmt.Errorf("failed to change message visibility: %w", err)
	}

	c.logger.Debug("message visibility changed", "timeout", visibilityTimeout)
	return nil
}

// GetQueueAttributes retrieves queue attributes
func (c *SQSClient) GetQueueAttributes(ctx context.Context) (*QueueAttributes, error) {
	result, err := c.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(c.queueURL),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
			types.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
			types.QueueAttributeNameApproximateNumberOfMessagesDelayed,
		},
	})
	if err != nil {
		c.logger.Error("failed to get queue attributes", "error", err)
		return nil, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	attrs := &QueueAttributes{}
	if val, ok := result.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)]; ok {
		fmt.Sscanf(val, "%d", &attrs.ApproximateNumberOfMessages)
	}
	if val, ok := result.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessagesNotVisible)]; ok {
		fmt.Sscanf(val, "%d", &attrs.ApproximateNumberOfMessagesNotVisible)
	}
	if val, ok := result.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessagesDelayed)]; ok {
		fmt.Sscanf(val, "%d", &attrs.ApproximateNumberOfMessagesDelayed)
	}

	return attrs, nil
}

// GetDLQAttributes retrieves dead-letter queue attributes
func (c *SQSClient) GetDLQAttributes(ctx context.Context) (*QueueAttributes, error) {
	if c.dlqURL == "" {
		return nil, fmt.Errorf("dead-letter queue URL not configured")
	}

	result, err := c.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(c.dlqURL),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
		},
	})
	if err != nil {
		c.logger.Error("failed to get DLQ attributes", "error", err)
		return nil, fmt.Errorf("failed to get DLQ attributes: %w", err)
	}

	attrs := &QueueAttributes{}
	if val, ok := result.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)]; ok {
		fmt.Sscanf(val, "%d", &attrs.ApproximateNumberOfMessages)
	}

	return attrs, nil
}

// Message represents an SQS message
type Message struct {
	ID                      string
	Body                    string
	ReceiptHandle           string
	Attributes              map[string]string
	ApproximateReceiveCount int
}

// BatchMessage represents a message for batch sending
type BatchMessage struct {
	Body       string
	Attributes map[string]string
}

// QueueAttributes represents queue metrics
type QueueAttributes struct {
	ApproximateNumberOfMessages            int
	ApproximateNumberOfMessagesNotVisible  int
	ApproximateNumberOfMessagesDelayed     int
}

// Helper function to extract approximate receive count from message attributes
func getApproximateReceiveCount(attrs map[string]string) int {
	if val, ok := attrs[string(types.MessageSystemAttributeNameApproximateReceiveCount)]; ok {
		var count int
		fmt.Sscanf(val, "%d", &count)
		return count
	}
	return 0
}
