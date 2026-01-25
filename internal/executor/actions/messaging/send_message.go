package messaging

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gorax/gorax/internal/messaging"
)

// CredentialService defines the interface for retrieving credentials
type CredentialService interface {
	GetCredentialValue(ctx context.Context, tenantID, credentialID string) (map[string]interface{}, error)
}

// SendMessageAction implements the send_message workflow action
type SendMessageAction struct {
	config            SendMessageConfig
	credentialService CredentialService
	queueFactory      func(ctx context.Context, config messaging.Config) (messaging.MessageQueue, error)
}

// SendMessageConfig represents the configuration for sending a message
type SendMessageConfig struct {
	QueueType    string            `json:"queue_type"`    // sqs, kafka, rabbitmq
	Destination  string            `json:"destination"`   // queue URL, topic name, etc.
	Message      string            `json:"message"`       // Message body (supports expressions)
	Attributes   map[string]string `json:"attributes"`    // Message attributes
	CredentialID string            `json:"credential_id"` // Queue credential
}

// Validate validates the SendMessageConfig
func (c *SendMessageConfig) Validate() error {
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

	if c.Destination == "" {
		return fmt.Errorf("destination is required")
	}

	if c.Message == "" {
		return fmt.Errorf("message is required")
	}

	if c.CredentialID == "" {
		return fmt.Errorf("credential_id is required")
	}

	return nil
}

// NewSendMessageAction creates a new SendMessageAction
func NewSendMessageAction(config SendMessageConfig, credentialService CredentialService) (*SendMessageAction, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &SendMessageAction{
		config:            config,
		credentialService: credentialService,
		queueFactory:      messaging.NewMessageQueue,
	}, nil
}

// Execute executes the send message action
func (a *SendMessageAction) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Extract tenant ID from input
	tenantID, ok := input["tenant_id"].(string)
	if !ok {
		return nil, fmt.Errorf("tenant_id is required in input")
	}

	// Evaluate expressions in destination
	destination := a.evaluateExpression(a.config.Destination, input)

	// Evaluate expressions in message
	message := a.evaluateExpression(a.config.Message, input)

	// Evaluate expressions in attributes
	attributes := make(map[string]string)
	for key, value := range a.config.Attributes {
		attributes[key] = a.evaluateExpression(value, input)
	}

	// Get credential
	credValue, err := a.credentialService.GetCredentialValue(ctx, tenantID, a.config.CredentialID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	// Create queue config based on queue type
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

	// Send message
	err = queue.Send(ctx, destination, []byte(message), attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Generate message ID for tracking
	messageID := uuid.New().String()

	// Return success output
	output := map[string]interface{}{
		"success":     true,
		"message_id":  messageID,
		"queue_type":  a.config.QueueType,
		"destination": destination,
		"sent_at":     time.Now().UTC().Format(time.RFC3339),
	}

	return output, nil
}

// buildQueueConfig builds the messaging.Config based on the queue type and credential
func (a *SendMessageAction) buildQueueConfig(credValue map[string]interface{}) (messaging.Config, error) {
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

// evaluateExpression evaluates simple template expressions like {{input.field}}
func (a *SendMessageAction) evaluateExpression(expr string, input map[string]interface{}) string {
	// Simple expression evaluation - replace {{input.field}} with actual values
	result := expr

	// Find all {{...}} patterns
	start := 0
	for {
		startIdx := strings.Index(result[start:], "{{")
		if startIdx == -1 {
			break
		}
		startIdx += start

		endIdx := strings.Index(result[startIdx:], "}}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx

		// Extract the field path
		fieldPath := strings.TrimSpace(result[startIdx+2 : endIdx])

		// Simple path resolution (e.g., "input.user_id")
		value := a.resolvePath(fieldPath, input)

		// Replace the expression with the value
		result = result[:startIdx] + value + result[endIdx+2:]
		start = startIdx + len(value)
	}

	return result
}

// resolvePath resolves a simple dot-notation path like "input.user_id"
func (a *SendMessageAction) resolvePath(path string, data map[string]interface{}) string {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return ""
	}

	// Handle "input.field" notation
	if parts[0] == "input" && len(parts) > 1 {
		if value, ok := data[parts[1]]; ok {
			return fmt.Sprintf("%v", value)
		}
	}

	// Handle direct field access
	if value, ok := data[parts[0]]; ok {
		return fmt.Sprintf("%v", value)
	}

	return ""
}

// GetName returns the action name
func (a *SendMessageAction) GetName() string {
	return "send_message"
}

// GetType returns the action type
func (a *SendMessageAction) GetType() string {
	return "messaging"
}
