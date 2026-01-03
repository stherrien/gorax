package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaWriter defines the interface for Kafka writer operations
type KafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// KafkaReader defines the interface for Kafka reader operations
type KafkaReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// KafkaQueue implements MessageQueue for Apache Kafka
type KafkaQueue struct {
	writer        KafkaWriter
	reader        KafkaReader
	brokers       []string
	consumerGroup string
	pendingMsgs   map[string]kafka.Message
	mu            sync.RWMutex
}

// NewKafkaQueue creates a new Apache Kafka queue client
func NewKafkaQueue(ctx context.Context, config Config) (*KafkaQueue, error) {
	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("brokers are required for Kafka")
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	return &KafkaQueue{
		writer:      writer,
		brokers:     config.Brokers,
		pendingMsgs: make(map[string]kafka.Message),
	}, nil
}

// Send sends a message to a Kafka topic
func (q *KafkaQueue) Send(ctx context.Context, destination string, message []byte, attributes map[string]string) error {
	if destination == "" {
		return fmt.Errorf("destination topic is required")
	}

	if len(message) == 0 {
		return fmt.Errorf("message body cannot be empty")
	}

	// Convert attributes to Kafka headers
	headers := make([]kafka.Header, 0, len(attributes))
	for key, value := range attributes {
		headers = append(headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	kafkaMsg := kafka.Message{
		Topic:   destination,
		Value:   message,
		Headers: headers,
		Time:    time.Now(),
	}

	err := q.writer.WriteMessages(ctx, kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	return nil
}

// Receive receives messages from a Kafka topic
func (q *KafkaQueue) Receive(ctx context.Context, source string, maxMessages int, waitTime time.Duration) ([]Message, error) {
	if source == "" {
		return nil, fmt.Errorf("source topic is required")
	}

	if maxMessages <= 0 {
		return nil, fmt.Errorf("maxMessages must be greater than 0")
	}

	// Create reader if not exists
	if q.reader == nil {
		readerConfig := kafka.ReaderConfig{
			Brokers:        q.brokers,
			Topic:          source,
			GroupID:        q.consumerGroup,
			MinBytes:       1,
			MaxBytes:       10e6, // 10MB
			CommitInterval: time.Second,
		}
		q.reader = kafka.NewReader(readerConfig)
	}

	messages := make([]Message, 0, maxMessages)
	fetchCtx, cancel := context.WithTimeout(ctx, waitTime)
	defer cancel()

	for i := 0; i < maxMessages; i++ {
		kafkaMsg, err := q.reader.FetchMessage(fetchCtx)
		if err != nil {
			if err == context.DeadlineExceeded {
				break // No more messages available
			}
			return nil, fmt.Errorf("failed to fetch message from Kafka: %w", err)
		}

		// Convert Kafka headers to map
		attributes := make(map[string]string)
		for _, header := range kafkaMsg.Headers {
			attributes[header.Key] = string(header.Value)
		}

		msgID := fmt.Sprintf("%s-%d-%d", kafkaMsg.Topic, kafkaMsg.Partition, kafkaMsg.Offset)
		receipt := fmt.Sprintf("kafka-%d", time.Now().UnixNano())

		msg := Message{
			ID:         msgID,
			Body:       kafkaMsg.Value,
			Attributes: attributes,
			Receipt:    receipt,
			Timestamp:  kafkaMsg.Time,
		}

		// Store pending message for acknowledgment
		q.mu.Lock()
		q.pendingMsgs[receipt] = kafkaMsg
		q.mu.Unlock()

		messages = append(messages, msg)
	}

	return messages, nil
}

// Ack acknowledges a message by committing the offset
func (q *KafkaQueue) Ack(ctx context.Context, message Message) error {
	if message.Receipt == "" {
		return fmt.Errorf("message receipt is required")
	}

	q.mu.Lock()
	kafkaMsg, ok := q.pendingMsgs[message.Receipt]
	if !ok {
		q.mu.Unlock()
		return fmt.Errorf("message not found in pending messages")
	}
	delete(q.pendingMsgs, message.Receipt)
	q.mu.Unlock()

	if q.reader == nil {
		return fmt.Errorf("reader not initialized")
	}

	err := q.reader.CommitMessages(ctx, kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to commit message in Kafka: %w", err)
	}

	return nil
}

// Nack is not supported in Kafka (commit-based system)
// We simply don't commit the message, and it will be redelivered
func (q *KafkaQueue) Nack(ctx context.Context, message Message) error {
	// Remove from pending messages without committing
	q.mu.Lock()
	delete(q.pendingMsgs, message.Receipt)
	q.mu.Unlock()

	return nil
}

// GetInfo retrieves information about a Kafka topic
func (q *KafkaQueue) GetInfo(ctx context.Context, name string) (*QueueInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("topic name is required")
	}

	// Kafka doesn't provide easy access to topic metadata without admin client
	// Return basic info
	info := &QueueInfo{
		Name:             name,
		ApproximateCount: 0, // Would need admin client to get accurate count
		CreatedAt:        time.Now(),
	}

	return info, nil
}

// Close closes the Kafka connections
func (q *KafkaQueue) Close() error {
	var errs []error

	if q.writer != nil {
		if err := q.writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("writer close error: %w", err))
		}
	}

	if q.reader != nil {
		if err := q.reader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("reader close error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close Kafka connections: %v", errs)
	}

	return nil
}

// SetConsumerGroup sets the consumer group for the Kafka reader
func (q *KafkaQueue) SetConsumerGroup(group string) {
	q.consumerGroup = group
}
