package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"

	"github.com/gorax/gorax/internal/integrations"
)

// SNSClient wraps the AWS SNS client
type SNSClient struct {
	client    *sns.Client
	accessKey string
	secretKey string
	region    string
}

// NewSNSClient creates a new SNS client
func NewSNSClient(accessKey, secretKey, region string) (*SNSClient, error) {
	if err := validateSNSConfig(accessKey, secretKey, region); err != nil {
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

	return &SNSClient{
		client:    sns.NewFromConfig(cfg),
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}, nil
}

// PublishMessageConfig represents configuration for PublishMessage action
type PublishMessageConfig struct {
	TopicARN   string            `json:"topic_arn,omitempty"`
	TargetARN  string            `json:"target_arn,omitempty"`
	Message    string            `json:"message"`
	Subject    string            `json:"subject,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// Validate validates PublishMessageConfig
func (c *PublishMessageConfig) Validate() error {
	if c.TopicARN == "" && c.TargetARN == "" {
		return fmt.Errorf("either topic_arn or target_arn is required")
	}
	if c.Message == "" {
		return fmt.Errorf("message is required")
	}
	return nil
}

// PublishMessageAction implements the aws:sns:publish action
type PublishMessageAction struct {
	accessKey string
	secretKey string
	region    string
}

// NewPublishMessageAction creates a new PublishMessage action
func NewPublishMessageAction(accessKey, secretKey, region string) *PublishMessageAction {
	return &PublishMessageAction{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}
}

// Execute implements the Action interface
func (a *PublishMessageAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var msgConfig PublishMessageConfig
	if err := json.Unmarshal(configJSON, &msgConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := msgConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewSNSClient(a.accessKey, a.secretKey, a.region)
	if err != nil {
		return nil, fmt.Errorf("failed to create SNS client: %w", err)
	}

	publishInput := &sns.PublishInput{
		Message: aws.String(msgConfig.Message),
	}

	// Set topic or target ARN (topic takes precedence)
	if msgConfig.TopicARN != "" {
		publishInput.TopicArn = aws.String(msgConfig.TopicARN)
	} else {
		publishInput.TargetArn = aws.String(msgConfig.TargetARN)
	}

	// Set optional subject
	if msgConfig.Subject != "" {
		publishInput.Subject = aws.String(msgConfig.Subject)
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
		publishInput.MessageAttributes = attributes
	}

	result, err := client.client.Publish(ctx, publishInput)
	if err != nil {
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	return map[string]interface{}{
		"message_id": aws.ToString(result.MessageId),
		"success":    true,
	}, nil
}

// Validate implements the Action interface
func (a *PublishMessageAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var msgConfig PublishMessageConfig
	if err := json.Unmarshal(configJSON, &msgConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return msgConfig.Validate()
}

// Name implements the Action interface
func (a *PublishMessageAction) Name() string {
	return "aws:sns:publish"
}

// Description implements the Action interface
func (a *PublishMessageAction) Description() string {
	return "Publish a message to an SNS topic or endpoint"
}

// validateSNSConfig validates SNS configuration
func validateSNSConfig(accessKey, secretKey, region string) error {
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
var _ integrations.Action = (*PublishMessageAction)(nil)
