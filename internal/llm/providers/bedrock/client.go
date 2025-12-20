package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"

	"github.com/gorax/gorax/internal/llm"
)

const (
	providerName          = "bedrock"
	anthropicVersion      = "bedrock-2023-05-31"
	defaultMaxTokens      = 4096
)

// BedrockAPI defines the interface for Bedrock operations (for testing)
type BedrockAPI interface {
	InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}

// Client is an AWS Bedrock API client
type Client struct {
	region    string
	apiClient BedrockAPI
}

// NewClient creates a new Bedrock client
func NewClient(cfg *llm.ProviderConfig) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if cfg.Region == "" {
		return nil, fmt.Errorf("region is required for Bedrock")
	}

	// Build AWS config options
	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(cfg.Region))

	// Use explicit credentials if provided
	if cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "" {
		creds := credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"",
		)
		opts = append(opts, config.WithCredentialsProvider(creds))
	}

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Bedrock runtime client
	bedrockClient := bedrockruntime.NewFromConfig(awsCfg)

	return &Client{
		region:    cfg.Region,
		apiClient: bedrockClient,
	}, nil
}

// Name returns the provider name
func (c *Client) Name() string {
	return providerName
}

// ChatCompletion performs a chat completion request
func (c *Client) ChatCompletion(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if isClaudeModel(req.Model) {
		return c.claudeChatCompletion(ctx, req)
	}
	if isTitanModel(req.Model) {
		return c.titanChatCompletion(ctx, req)
	}

	return nil, &llm.LLMError{
		Provider: providerName,
		Code:     "unsupported_model",
		Message:  fmt.Sprintf("unsupported model: %s", req.Model),
		Cause:    llm.ErrInvalidModel,
	}
}

// claudeChatCompletion handles Claude model requests
func (c *Client) claudeChatCompletion(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	claudeReq := c.buildClaudeRequest(req)

	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	output, err := c.apiClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(req.Model),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        reqBody,
	})
	if err != nil {
		return nil, c.parseAWSError(err)
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(output.Body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return c.convertClaudeResponse(&claudeResp), nil
}

// buildClaudeRequest converts generic request to Claude format
func (c *Client) buildClaudeRequest(req *llm.ChatRequest) *claudeRequest {
	claudeReq := &claudeRequest{
		AnthropicVersion: anthropicVersion,
		MaxTokens:        defaultMaxTokens,
	}

	// Extract system prompt and convert messages
	var messages []claudeMessage
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			claudeReq.System = msg.Content
			continue
		}
		messages = append(messages, claudeMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	claudeReq.Messages = messages

	if req.MaxTokens > 0 {
		claudeReq.MaxTokens = req.MaxTokens
	}
	if req.Temperature != nil {
		claudeReq.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		claudeReq.TopP = *req.TopP
	}
	if len(req.Stop) > 0 {
		claudeReq.StopSequences = req.Stop
	}

	return claudeReq
}

// convertClaudeResponse converts Claude response to generic format
func (c *Client) convertClaudeResponse(resp *claudeResponse) *llm.ChatResponse {
	var content strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			content.WriteString(block.Text)
		}
	}

	return &llm.ChatResponse{
		ID:    resp.ID,
		Model: resp.Model,
		Message: llm.ChatMessage{
			Role:    resp.Role,
			Content: content.String(),
		},
		FinishReason: resp.StopReason,
		Usage: llm.TokenUsage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}

// titanChatCompletion handles Amazon Titan model requests
func (c *Client) titanChatCompletion(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	titanReq := c.buildTitanRequest(req)

	reqBody, err := json.Marshal(titanReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	output, err := c.apiClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(req.Model),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        reqBody,
	})
	if err != nil {
		return nil, c.parseAWSError(err)
	}

	var titanResp titanResponse
	if err := json.Unmarshal(output.Body, &titanResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return c.convertTitanResponse(&titanResp, req.Model), nil
}

// buildTitanRequest converts generic request to Titan format
func (c *Client) buildTitanRequest(req *llm.ChatRequest) *titanRequest {
	// Titan uses a single text input, so we need to format messages
	var inputText strings.Builder
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			inputText.WriteString(fmt.Sprintf("Instructions: %s\n\n", msg.Content))
		case "user":
			inputText.WriteString(fmt.Sprintf("User: %s\n", msg.Content))
		case "assistant":
			inputText.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
		}
	}
	inputText.WriteString("Assistant:")

	titanReq := &titanRequest{
		InputText: inputText.String(),
	}

	// Configure generation settings
	if req.MaxTokens > 0 || req.Temperature != nil || req.TopP != nil || len(req.Stop) > 0 {
		config := titanGenerationConfig{}
		if req.MaxTokens > 0 {
			config.MaxTokenCount = req.MaxTokens
		}
		if req.Temperature != nil {
			config.Temperature = *req.Temperature
		}
		if req.TopP != nil {
			config.TopP = *req.TopP
		}
		if len(req.Stop) > 0 {
			config.StopSequences = req.Stop
		}
		titanReq.TextGenerationConfig = config
	}

	return titanReq
}

// convertTitanResponse converts Titan response to generic format
func (c *Client) convertTitanResponse(resp *titanResponse, model string) *llm.ChatResponse {
	var outputText string
	var completionTokens int
	var finishReason string

	if len(resp.Results) > 0 {
		outputText = resp.Results[0].OutputText
		completionTokens = resp.Results[0].TokenCount
		finishReason = resp.Results[0].CompletionReason
	}

	return &llm.ChatResponse{
		ID:    fmt.Sprintf("titan-%d", resp.InputTextTokenCount), // Titan doesn't return an ID
		Model: model,
		Message: llm.ChatMessage{
			Role:    "assistant",
			Content: strings.TrimSpace(outputText),
		},
		FinishReason: finishReason,
		Usage: llm.TokenUsage{
			PromptTokens:     resp.InputTextTokenCount,
			CompletionTokens: completionTokens,
			TotalTokens:      resp.InputTextTokenCount + completionTokens,
		},
	}
}

// GenerateEmbeddings generates embeddings for the given texts
func (c *Client) GenerateEmbeddings(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Only Titan embedding models are supported
	if !isTitanEmbeddingModel(req.Model) {
		return nil, &llm.LLMError{
			Provider: providerName,
			Code:     "unsupported_operation",
			Message:  fmt.Sprintf("embeddings not supported for model: %s", req.Model),
			Cause:    llm.ErrUnsupportedOperation,
		}
	}

	embeddings := make([][]float64, len(req.Texts))
	totalTokens := 0

	// Process each text individually (Titan doesn't support batch)
	for i, text := range req.Texts {
		embedReq := titanEmbedRequest{InputText: text}
		reqBody, err := json.Marshal(embedReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		output, err := c.apiClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(req.Model),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
			Body:        reqBody,
		})
		if err != nil {
			return nil, c.parseAWSError(err)
		}

		var embedResp titanEmbedResponse
		if err := json.Unmarshal(output.Body, &embedResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		embeddings[i] = embedResp.Embedding
		totalTokens += embedResp.InputTextTokenCount
	}

	return &llm.EmbeddingResponse{
		Model:      req.Model,
		Embeddings: embeddings,
		Usage: llm.TokenUsage{
			PromptTokens: totalTokens,
			TotalTokens:  totalTokens,
		},
	}, nil
}

// CountTokens estimates the token count for the given text
func (c *Client) CountTokens(text string, model string) (int, error) {
	if text == "" {
		return 0, nil
	}
	// Rough approximation
	return (len(text) + 3) / 4, nil
}

// ListModels returns the available Bedrock models
func (c *Client) ListModels(ctx context.Context) ([]llm.Model, error) {
	return bedrockModels, nil
}

// HealthCheck verifies the API connection
func (c *Client) HealthCheck(ctx context.Context) error {
	req := &llm.ChatRequest{
		Model: "anthropic.claude-3-haiku-20240307-v1:0",
		Messages: []llm.ChatMessage{
			{Role: "user", Content: "hi"},
		},
		MaxTokens: 1,
	}
	_, err := c.ChatCompletion(ctx, req)
	return err
}

// parseAWSError converts AWS errors to LLMError
func (c *Client) parseAWSError(err error) error {
	errMsg := err.Error()

	llmErr := &llm.LLMError{
		Provider: providerName,
		Message:  errMsg,
	}

	// Check for common error patterns
	if strings.Contains(errMsg, "ThrottlingException") || strings.Contains(errMsg, "rate") {
		llmErr.Retryable = true
		llmErr.Cause = llm.ErrRateLimitExceeded
		return llmErr
	}

	if strings.Contains(errMsg, "AccessDeniedException") || strings.Contains(errMsg, "UnauthorizedAccess") {
		llmErr.Cause = llm.ErrInvalidAPIKey
		return llmErr
	}

	if strings.Contains(errMsg, "ValidationException") && strings.Contains(errMsg, "token") {
		llmErr.Cause = llm.ErrContextTooLong
		return llmErr
	}

	if strings.Contains(errMsg, "ServiceUnavailableException") {
		llmErr.Retryable = true
		llmErr.Cause = llm.ErrProviderUnavailable
		return llmErr
	}

	return llmErr
}

// Model type detection helpers
func isClaudeModel(modelID string) bool {
	return strings.HasPrefix(modelID, "anthropic.claude")
}

func isTitanModel(modelID string) bool {
	return strings.HasPrefix(modelID, "amazon.titan")
}

func isTitanEmbeddingModel(modelID string) bool {
	return strings.HasPrefix(modelID, "amazon.titan-embed")
}

// RegisterWithGlobal registers the Bedrock provider with the global registry
func RegisterWithGlobal() error {
	return llm.RegisterProvider(providerName, func(config *llm.ProviderConfig) (llm.Provider, error) {
		return NewClient(config)
	})
}
