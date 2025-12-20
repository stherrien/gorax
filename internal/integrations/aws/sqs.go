package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/gorax/gorax/internal/integrations"
)

// SQSClient wraps the AWS SQS client
type SQSClient struct {
	client    *sqs.Client
	accessKey string
	secretKey string
	region    string
}

// NewSQSClient creates a new SQS client
func NewSQSClient(accessKey, secretKey, region string) (*SQSClient, error) {
	if err := validateSQSConfig(accessKey, secretKey, region); err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &SQSClient{
		client:    sqs.NewFromConfig(cfg),
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}, nil
}

// SendMessageConfig represents configuration for SendMessage action
type SendMessageConfig struct {
	QueueURL         string            `json:"queue_url"`
	MessageBody      string            `json:"message_body"`
	DelaySeconds     int32             `json:"delay_seconds,omitempty"`
	MessageGroupID   string            `json:"message_group_id,omitempty"`
	DeduplicationID  string            `json:"deduplication_id,omitempty"`
	Attributes       map[string]string `json:"attributes,omitempty"`
}

// Validate validates SendMessageConfig
func (c *SendMessageConfig) Validate() error {
	if c.QueueURL == "" {
		return fmt.Errorf("queue_url is required")
	}
	if c.MessageBody == "" {
		return fmt.Errorf("message_body is required")
	}
	return nil
}

// SendMessageAction implements the aws:sqs:send_message action
type SendMessageAction struct {
	accessKey string
	secretKey string
	region    string
}

// NewSendMessageAction creates a new SendMessage action
func NewSendMessageAction(accessKey, secretKey, region string) *SendMessageAction {
	return &SendMessageAction{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}
}

// Execute implements the Action interface
func (a *SendMessageAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var msgConfig SendMessageConfig
	if err := json.Unmarshal(configJSON, &msgConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := msgConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewSQSClient(a.accessKey, a.secretKey, a.region)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQS client: %w", err)
	}

	sendInput := &sqs.SendMessageInput{
		QueueUrl:    aws.String(msgConfig.QueueURL),
		MessageBody: aws.String(msgConfig.MessageBody),
	}

	// Set optional fields
	if msgConfig.DelaySeconds > 0 {
		sendInput.DelaySeconds = msgConfig.DelaySeconds
	}

	if msgConfig.MessageGroupID != "" {
		sendInput.MessageGroupId = aws.String(msgConfig.MessageGroupID)
	}

	if msgConfig.DeduplicationID != "" {
		sendInput.MessageDeduplicationId = aws.String(msgConfig.DeduplicationID)
	}

	// Set message attributes
	if len(msgConfig.Attributes) > 0 {
		attributes := make(map[string]types.MessageAttributeValue)
		for key, value := range msgConfig.Attributes {
			attributes[key] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(value),
			}
		}
		sendInput.MessageAttributes = attributes
	}

	result, err := client.client.SendMessage(ctx, sendInput)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return map[string]interface{}{
		"message_id": aws.ToString(result.MessageId),
		"success":    true,
	}, nil
}

// Validate implements the Action interface
func (a *SendMessageAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var msgConfig SendMessageConfig
	if err := json.Unmarshal(configJSON, &msgConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return msgConfig.Validate()
}

// Name implements the Action interface
func (a *SendMessageAction) Name() string {
	return "aws:sqs:send_message"
}

// Description implements the Action interface
func (a *SendMessageAction) Description() string {
	return "Send a message to an SQS queue"
}

// ReceiveMessageConfig represents configuration for ReceiveMessage action
type ReceiveMessageConfig struct {
	QueueURL          string `json:"queue_url"`
	MaxMessages       int32  `json:"max_messages,omitempty"`
	WaitTimeSeconds   int32  `json:"wait_time_seconds,omitempty"`
	VisibilityTimeout int32  `json:"visibility_timeout,omitempty"`
}

// Validate validates ReceiveMessageConfig
func (c *ReceiveMessageConfig) Validate() error {
	if c.QueueURL == "" {
		return fmt.Errorf("queue_url is required")
	}
	return nil
}

// ReceiveMessageAction implements the aws:sqs:receive_message action
type ReceiveMessageAction struct {
	accessKey string
	secretKey string
	region    string
}

// NewReceiveMessageAction creates a new ReceiveMessage action
func NewReceiveMessageAction(accessKey, secretKey, region string) *ReceiveMessageAction {
	return &ReceiveMessageAction{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}
}

// Execute implements the Action interface
func (a *ReceiveMessageAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var msgConfig ReceiveMessageConfig
	if err := json.Unmarshal(configJSON, &msgConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := msgConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewSQSClient(a.accessKey, a.secretKey, a.region)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQS client: %w", err)
	}

	receiveInput := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(msgConfig.QueueURL),
		MessageAttributeNames: []string{"All"},
	}

	// Set optional fields with defaults
	if msgConfig.MaxMessages > 0 {
		receiveInput.MaxNumberOfMessages = msgConfig.MaxMessages
	} else {
		receiveInput.MaxNumberOfMessages = 1
	}

	if msgConfig.WaitTimeSeconds > 0 {
		receiveInput.WaitTimeSeconds = msgConfig.WaitTimeSeconds
	}

	if msgConfig.VisibilityTimeout > 0 {
		receiveInput.VisibilityTimeout = msgConfig.VisibilityTimeout
	}

	result, err := client.client.ReceiveMessage(ctx, receiveInput)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}

	messages := make([]map[string]interface{}, 0, len(result.Messages))
	for _, msg := range result.Messages {
		msgData := map[string]interface{}{
			"message_id":     aws.ToString(msg.MessageId),
			"body":           aws.ToString(msg.Body),
			"receipt_handle": aws.ToString(msg.ReceiptHandle),
		}

		// Add attributes if present
		if len(msg.MessageAttributes) > 0 {
			attrs := make(map[string]string)
			for key, value := range msg.MessageAttributes {
				attrs[key] = aws.ToString(value.StringValue)
			}
			msgData["attributes"] = attrs
		}

		messages = append(messages, msgData)
	}

	return map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	}, nil
}

// Validate implements the Action interface
func (a *ReceiveMessageAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var msgConfig ReceiveMessageConfig
	if err := json.Unmarshal(configJSON, &msgConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return msgConfig.Validate()
}

// Name implements the Action interface
func (a *ReceiveMessageAction) Name() string {
	return "aws:sqs:receive_message"
}

// Description implements the Action interface
func (a *ReceiveMessageAction) Description() string {
	return "Receive messages from an SQS queue"
}

// DeleteMessageConfig represents configuration for DeleteMessage action
type DeleteMessageConfig struct {
	QueueURL      string `json:"queue_url"`
	ReceiptHandle string `json:"receipt_handle"`
}

// Validate validates DeleteMessageConfig
func (c *DeleteMessageConfig) Validate() error {
	if c.QueueURL == "" {
		return fmt.Errorf("queue_url is required")
	}
	if c.ReceiptHandle == "" {
		return fmt.Errorf("receipt_handle is required")
	}
	return nil
}

// DeleteMessageAction implements the aws:sqs:delete_message action
type DeleteMessageAction struct {
	accessKey string
	secretKey string
	region    string
}

// NewDeleteMessageAction creates a new DeleteMessage action
func NewDeleteMessageAction(accessKey, secretKey, region string) *DeleteMessageAction {
	return &DeleteMessageAction{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}
}

// Execute implements the Action interface
func (a *DeleteMessageAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var msgConfig DeleteMessageConfig
	if err := json.Unmarshal(configJSON, &msgConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := msgConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewSQSClient(a.accessKey, a.secretKey, a.region)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQS client: %w", err)
	}

	_, err = client.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(msgConfig.QueueURL),
		ReceiptHandle: aws.String(msgConfig.ReceiptHandle),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// Validate implements the Action interface
func (a *DeleteMessageAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var msgConfig DeleteMessageConfig
	if err := json.Unmarshal(configJSON, &msgConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return msgConfig.Validate()
}

// Name implements the Action interface
func (a *DeleteMessageAction) Name() string {
	return "aws:sqs:delete_message"
}

// Description implements the Action interface
func (a *DeleteMessageAction) Description() string {
	return "Delete a message from an SQS queue"
}

// validateSQSConfig validates SQS configuration
func validateSQSConfig(accessKey, secretKey, region string) error {
	if accessKey == "" {
		return fmt.Errorf("access key is required")
	}
	if secretKey == "" {
		return fmt.Errorf("secret key is required")
	}
	if region == "" {
		return fmt.Errorf("region is required")
	}
	return nil
}

// Ensure all actions implement the Action interface
var (
	_ integrations.Action = (*SendMessageAction)(nil)
	_ integrations.Action = (*ReceiveMessageAction)(nil)
	_ integrations.Action = (*DeleteMessageAction)(nil)
)
