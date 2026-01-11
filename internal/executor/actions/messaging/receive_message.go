package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/gorax/gorax/internal/messaging"
)

// ReceiveMessageAction implements the receive_message workflow action
type ReceiveMessageAction struct {
	config            ReceiveMessageConfig
	credentialService CredentialService
	queueFactory      func(ctx context.Context, config messaging.Config) (messaging.MessageQueue, error)
}

// ReceiveMessageConfig represents the configuration for receiving messages
type ReceiveMessageConfig struct {
	QueueType    string `json:"queue_type"`    // sqs, kafka, rabbitmq
	Source       string `json:"source"`        // queue URL, topic name, etc.
	MaxMessages  int    `json:"max_messages"`  // Maximum number of messages to receive
	WaitTime     string `json:"wait_time"`     // Wait time (e.g., "5s", "1m")
	DeleteAfter  bool   `json:"delete_after"`  // Auto-ack after receiving
	CredentialID string `json:"credential_id"` // Queue credential
}

// Validate validates the ReceiveMessageConfig
func (c *ReceiveMessageConfig) Validate() error {
	if c.QueueType == "" {
		return fmt.Errorf("queue_type is required")
	}

	validQueueTypes := []string{"sqs", "kafka", "rabbitmq"}
	isValid := false
	for _, validType := range validQueueTypes {
		if c.QueueType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("unsupported queue_type: %s (must be one of: sqs, kafka, rabbitmq)", c.QueueType)
	}

	if c.Source == "" {
		return fmt.Errorf("source is required")
	}

	if c.MaxMessages <= 0 {
		c.MaxMessages = 10 // Default
	}

	if c.WaitTime == "" {
		c.WaitTime = "5s" // Default
	}

	if c.CredentialID == "" {
		return fmt.Errorf("credential_id is required")
	}

	return nil
}

// NewReceiveMessageAction creates a new ReceiveMessageAction
func NewReceiveMessageAction(config ReceiveMessageConfig, credentialService CredentialService) (*ReceiveMessageAction, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &ReceiveMessageAction{
		config:            config,
		credentialService: credentialService,
		queueFactory:      messaging.NewMessageQueue,
	}, nil
}

// Execute executes the receive message action
func (a *ReceiveMessageAction) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Extract tenant ID from input
	tenantID, ok := input["tenant_id"].(string)
	if !ok {
		return nil, fmt.Errorf("tenant_id is required in input")
	}

	// Parse wait time
	waitTime, err := time.ParseDuration(a.config.WaitTime)
	if err != nil {
		return nil, fmt.Errorf("invalid wait_time format: %w", err)
	}

	// Get credential
	credValue, err := a.credentialService.GetCredentialValue(ctx, tenantID, a.config.CredentialID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	// Create queue config
	queueConfig, err := a.buildQueueConfig(credValue)
	if err != nil {
		return nil, fmt.Errorf("failed to build queue config: %w", err)
	}

	// Create queue client
	queue, err := a.queueFactory(ctx, queueConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue client: %w", err)
	}
	defer queue.Close()

	// Receive messages
	messages, err := queue.Receive(ctx, a.config.Source, a.config.MaxMessages, waitTime)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	// Convert messages to output format
	messagesOutput := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		msgOutput := map[string]interface{}{
			"id":         msg.ID,
			"body":       string(msg.Body),
			"attributes": msg.Attributes,
			"receipt":    msg.Receipt,
			"timestamp":  msg.Timestamp.Format(time.RFC3339),
		}
		messagesOutput = append(messagesOutput, msgOutput)

		// Auto-acknowledge if configured
		if a.config.DeleteAfter {
			if err := queue.Ack(ctx, msg); err != nil {
				// Log error but don't fail the action
				fmt.Printf("Warning: failed to ack message %s: %v\n", msg.ID, err)
			}
		}
	}

	// Return output
	output := map[string]interface{}{
		"success":       true,
		"message_count": len(messages),
		"messages":      messagesOutput,
		"queue_type":    a.config.QueueType,
		"source":        a.config.Source,
		"received_at":   time.Now().UTC().Format(time.RFC3339),
	}

	return output, nil
}

// buildQueueConfig builds the messaging.Config based on the queue type and credential
func (a *ReceiveMessageAction) buildQueueConfig(credValue map[string]interface{}) (messaging.Config, error) {
	config := messaging.Config{
		Type: messaging.QueueType(a.config.QueueType),
	}

	switch a.config.QueueType {
	case "sqs":
		region, ok := credValue["region"].(string)
		if !ok {
			return config, fmt.Errorf("region is required in SQS credential")
		}
		config.Region = region

	case "kafka":
		brokersInterface, ok := credValue["brokers"]
		if !ok {
			return config, fmt.Errorf("brokers are required in Kafka credential")
		}

		// Handle both []interface{} and []string
		brokers := make([]string, 0)
		switch v := brokersInterface.(type) {
		case []interface{}:
			for _, b := range v {
				if broker, ok := b.(string); ok {
					brokers = append(brokers, broker)
				}
			}
		case []string:
			brokers = v
		default:
			return config, fmt.Errorf("brokers must be a string array")
		}

		if len(brokers) == 0 {
			return config, fmt.Errorf("at least one broker is required")
		}
		config.Brokers = brokers

	case "rabbitmq":
		url, ok := credValue["url"].(string)
		if !ok {
			return config, fmt.Errorf("url is required in RabbitMQ credential")
		}
		config.URL = url

	default:
		return config, fmt.Errorf("unsupported queue type: %s", a.config.QueueType)
	}

	return config, nil
}

// GetName returns the action name
func (a *ReceiveMessageAction) GetName() string {
	return "receive_message"
}

// GetType returns the action type
func (a *ReceiveMessageAction) GetType() string {
	return "messaging"
}
