package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/gorax/gorax/internal/integrations"
)

// LambdaClient wraps the AWS Lambda client
type LambdaClient struct {
	client    *lambda.Client
	accessKey string
	secretKey string
	region    string
}

// NewLambdaClient creates a new Lambda client
func NewLambdaClient(accessKey, secretKey, region string) (*LambdaClient, error) {
	if err := validateLambdaConfig(accessKey, secretKey, region); err != nil {
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

	return &LambdaClient{
		client:    lambda.NewFromConfig(cfg),
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}, nil
}

// InvokeFunctionConfig represents configuration for InvokeFunction action
type InvokeFunctionConfig struct {
	FunctionName   string                 `json:"function_name"`
	Payload        map[string]interface{} `json:"payload"`
	InvocationType string                 `json:"invocation_type,omitempty"` // RequestResponse, Event, DryRun
	Qualifier      string                 `json:"qualifier,omitempty"`       // Version or alias
}

// Validate validates InvokeFunctionConfig
func (c *InvokeFunctionConfig) Validate() error {
	if c.FunctionName == "" {
		return fmt.Errorf("function_name is required")
	}
	if c.Payload == nil {
		return fmt.Errorf("payload is required")
	}

	// Validate invocation type if provided
	if c.InvocationType != "" {
		validTypes := map[string]bool{
			"RequestResponse": true,
			"Event":           true,
			"DryRun":          true,
		}
		if !validTypes[c.InvocationType] {
			return fmt.Errorf("invocation_type must be one of: RequestResponse, Event, DryRun")
		}
	}

	return nil
}

// InvokeFunctionAction implements the aws:lambda:invoke action
type InvokeFunctionAction struct {
	accessKey string
	secretKey string
	region    string
}

// NewInvokeFunctionAction creates a new InvokeFunction action
func NewInvokeFunctionAction(accessKey, secretKey, region string) *InvokeFunctionAction {
	return &InvokeFunctionAction{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}
}

// Execute implements the Action interface
func (a *InvokeFunctionAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var funcConfig InvokeFunctionConfig
	if err := json.Unmarshal(configJSON, &funcConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := funcConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewLambdaClient(a.accessKey, a.secretKey, a.region)
	if err != nil {
		return nil, fmt.Errorf("failed to create Lambda client: %w", err)
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(funcConfig.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Build invoke input
	invokeInput := &lambda.InvokeInput{
		FunctionName: aws.String(funcConfig.FunctionName),
		Payload:      payloadBytes,
	}

	// Set invocation type (default to RequestResponse for synchronous)
	invocationType := funcConfig.InvocationType
	if invocationType == "" {
		invocationType = "RequestResponse"
	}
	invokeInput.InvocationType = types.InvocationType(invocationType)

	// Set qualifier if provided
	if funcConfig.Qualifier != "" {
		invokeInput.Qualifier = aws.String(funcConfig.Qualifier)
	}

	// Invoke the function
	result, err := client.client.Invoke(ctx, invokeInput)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke function: %w", err)
	}

	output := map[string]interface{}{
		"status_code": result.StatusCode,
		"success":     result.FunctionError == nil,
	}

	// For async invocations, there's no response payload
	if invocationType == "Event" {
		output["invocation_type"] = "async"
		return output, nil
	}

	// For sync invocations, parse the response
	if result.FunctionError != nil {
		output["error"] = aws.ToString(result.FunctionError)
	}

	// Parse response payload
	if len(result.Payload) > 0 {
		var responsePayload interface{}
		if err := json.Unmarshal(result.Payload, &responsePayload); err != nil {
			// If not valid JSON, return as string
			output["payload"] = string(result.Payload)
		} else {
			output["payload"] = responsePayload
		}
	}

	// Add execution metadata if available
	if result.ExecutedVersion != nil {
		output["executed_version"] = aws.ToString(result.ExecutedVersion)
	}

	if result.LogResult != nil {
		output["log_result"] = aws.ToString(result.LogResult)
	}

	return output, nil
}

// Validate implements the Action interface
func (a *InvokeFunctionAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var funcConfig InvokeFunctionConfig
	if err := json.Unmarshal(configJSON, &funcConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return funcConfig.Validate()
}

// Name implements the Action interface
func (a *InvokeFunctionAction) Name() string {
	return "aws:lambda:invoke"
}

// Description implements the Action interface
func (a *InvokeFunctionAction) Description() string {
	return "Invoke an AWS Lambda function synchronously or asynchronously"
}

// validateLambdaConfig validates Lambda configuration
func validateLambdaConfig(accessKey, secretKey, region string) error {
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
var _ integrations.Action = (*InvokeFunctionAction)(nil)
