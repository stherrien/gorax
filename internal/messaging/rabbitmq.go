package messaging

import (
	"context"
	"fmt"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// AMQPChannel defines the interface for AMQP channel operations
type AMQPChannel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Ack(tag uint64, multiple bool) error
	Nack(tag uint64, multiple, requeue bool) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Close() error
}

// RabbitMQQueue implements MessageQueue for RabbitMQ
type RabbitMQQueue struct {
	conn    *amqp.Connection
	channel AMQPChannel
	url     string
}

// NewRabbitMQQueue creates a new RabbitMQ queue client
func NewRabbitMQQueue(ctx context.Context, config Config) (*RabbitMQQueue, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("URL is required for RabbitMQ")
	}

	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQQueue{
		conn:    conn,
		channel: ch,
		url:     config.URL,
	}, nil
}

// Send sends a message to a RabbitMQ queue
func (q *RabbitMQQueue) Send(ctx context.Context, destination string, message []byte, attributes map[string]string) error {
	if destination == "" {
		return fmt.Errorf("destination queue is required")
	}

	if len(message) == 0 {
		return fmt.Errorf("message body cannot be empty")
	}

	// Convert attributes to AMQP headers
	headers := make(amqp.Table)
	for key, value := range attributes {
		headers[key] = value
	}

	publishing := amqp.Publishing{
		ContentType:  "application/json",
		Body:         message,
		Headers:      headers,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
	}

	err := q.channel.Publish(
		"",          // exchange (empty for direct routing to queue)
		destination, // routing key (queue name)
		false,       // mandatory
		false,       // immediate
		publishing,
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
	}

	return nil
}

// Receive receives messages from a RabbitMQ queue
func (q *RabbitMQQueue) Receive(ctx context.Context, source string, maxMessages int, waitTime time.Duration) ([]Message, error) {
	if source == "" {
		return nil, fmt.Errorf("source queue is required")
	}

	if maxMessages <= 0 {
		return nil, fmt.Errorf("maxMessages must be greater than 0")
	}

	msgs, err := q.channel.Consume(
		source, // queue
		"",     // consumer tag (empty for auto-generated)
		false,  // auto-ack (we want manual ack)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume from RabbitMQ: %w", err)
	}

	messages := make([]Message, 0, maxMessages)
	timeout := time.After(waitTime)

	for i := 0; i < maxMessages; i++ {
		select {
		case delivery, ok := <-msgs:
			if !ok {
				// Channel closed
				return messages, nil
			}

			// Convert AMQP headers to map
			attributes := make(map[string]string)
			for key, value := range delivery.Headers {
				if strValue, ok := value.(string); ok {
					attributes[key] = strValue
				}
			}

			msg := Message{
				ID:         delivery.MessageId,
				Body:       delivery.Body,
				Attributes: attributes,
				Receipt:    strconv.FormatUint(delivery.DeliveryTag, 10),
				Timestamp:  delivery.Timestamp,
			}
			messages = append(messages, msg)

		case <-timeout:
			// Wait time exceeded, return what we have
			return messages, nil

		case <-ctx.Done():
			return messages, ctx.Err()
		}
	}

	return messages, nil
}

// Ack acknowledges a message
func (q *RabbitMQQueue) Ack(ctx context.Context, message Message) error {
	if message.Receipt == "" {
		return fmt.Errorf("message receipt (delivery tag) is required")
	}

	deliveryTag, err := strconv.ParseUint(message.Receipt, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid delivery tag: %w", err)
	}

	err = q.channel.Ack(deliveryTag, false)
	if err != nil {
		return fmt.Errorf("failed to acknowledge message in RabbitMQ: %w", err)
	}

	return nil
}

// Nack negatively acknowledges a message (requeue)
func (q *RabbitMQQueue) Nack(ctx context.Context, message Message) error {
	if message.Receipt == "" {
		return fmt.Errorf("message receipt (delivery tag) is required")
	}

	deliveryTag, err := strconv.ParseUint(message.Receipt, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid delivery tag: %w", err)
	}

	err = q.channel.Nack(deliveryTag, false, true) // requeue=true
	if err != nil {
		return fmt.Errorf("failed to nack message in RabbitMQ: %w", err)
	}

	return nil
}

// GetInfo retrieves information about a RabbitMQ queue
func (q *RabbitMQQueue) GetInfo(ctx context.Context, name string) (*QueueInfo, error) {
	if name == "" {
		return nil, fmt.Errorf("queue name is required")
	}

	// Passive declare to get queue info without creating it
	queueInfo, err := q.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue info: %w", err)
	}

	info := &QueueInfo{
		Name:             name,
		ApproximateCount: queueInfo.Messages,
		CreatedAt:        time.Now(), // RabbitMQ doesn't expose creation time easily
	}

	return info, nil
}

// Close closes the RabbitMQ connection
func (q *RabbitMQQueue) Close() error {
	var errs []error

	if q.channel != nil {
		if err := q.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("channel close error: %w", err))
		}
	}

	if q.conn != nil {
		if err := q.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("connection close error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close RabbitMQ connections: %v", errs)
	}

	return nil
}
