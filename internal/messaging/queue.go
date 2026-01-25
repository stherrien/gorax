package messaging

import (
	"context"
	"fmt"
	"time"
)

// MessageQueue defines the interface for message queue operations
type MessageQueue interface {
	// Send a message to a queue/topic
	Send(ctx context.Context, destination string, message []byte, attributes map[string]string) error

	// Receive messages from a queue/topic
	Receive(ctx context.Context, source string, maxMessages int, waitTime time.Duration) ([]Message, error)

	// Ack acknowledges successful processing of a message
	Ack(ctx context.Context, message Message) error

	// Nack negatively acknowledges a message (requeue)
	Nack(ctx context.Context, message Message) error

	// GetInfo retrieves information about a queue/topic
	GetInfo(ctx context.Context, name string) (*QueueInfo, error)

	// Close closes the connection
	Close() error
}

// Message represents a message from a queue
type Message struct {
	ID         string            // Unique message identifier
	Body       []byte            // Message body
	Attributes map[string]string // Message attributes/headers
	Receipt    string            // Receipt handle for acknowledgment
	Timestamp  time.Time         // Message timestamp
}

// QueueInfo contains information about a queue/topic
type QueueInfo struct {
	Name             string
	ApproximateCount int
	CreatedAt        time.Time
}

// QueueType represents the type of message queue
type QueueType string

const (
	QueueTypeSQS      QueueType = "sqs"
	QueueTypeKafka    QueueType = "kafka"
	QueueTypeRabbitMQ QueueType = "rabbitmq"
)

// Config represents common configuration for message queues
type Config struct {
	Type       QueueType
	Region     string   // For AWS SQS
	Brokers    []string // For Kafka
	URL        string   // For RabbitMQ
	MaxRetries int
	Timeout    time.Duration
}

// Validate validates the queue configuration
func (c *Config) Validate() error {
	if c.Type == "" {
		return fmt.Errorf("queue type is required")
	}

	switch c.Type {
	case QueueTypeSQS:
		if c.Region == "" {
			return fmt.Errorf("region is required for SQS")
		}
	case QueueTypeKafka:
		if len(c.Brokers) == 0 {
			return fmt.Errorf("brokers are required for Kafka")
		}
	case QueueTypeRabbitMQ:
		if c.URL == "" {
			return fmt.Errorf("URL is required for RabbitMQ")
		}
	default:
		return fmt.Errorf("unsupported queue type: %s", c.Type)
	}

	return nil
}

// NewMessageQueue creates a new message queue client based on configuration
func NewMessageQueue(ctx context.Context, config Config) (MessageQueue, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	switch config.Type {
	case QueueTypeSQS:
		return NewSQSQueue(ctx, config)
	case QueueTypeKafka:
		return NewKafkaQueue(ctx, config)
	case QueueTypeRabbitMQ:
		return NewRabbitMQQueue(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported queue type: %s", config.Type)
	}
}
