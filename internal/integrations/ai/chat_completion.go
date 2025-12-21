package ai

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/llm"
)

// CredentialService defines the credential service interface needed by AI actions
type CredentialService interface {
	GetValue(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error)
}

// ProviderRegistry defines the provider registry interface
type ProviderRegistry interface {
	GetProviderFromCredential(name string, credValue map[string]interface{}) (llm.Provider, error)
}

// ChatCompletionConfig represents configuration for the ChatCompletion action
type ChatCompletionConfig struct {
	// Provider is the LLM provider name (e.g., "openai", "anthropic", "bedrock")
	Provider string `json:"provider"`

	// Model is the model identifier (e.g., "gpt-4o", "claude-3-sonnet-20240229")
	Model string `json:"model"`

	// SystemPrompt is an optional system prompt to prepend to messages
	SystemPrompt string `json:"system_prompt,omitempty"`

	// Messages is the conversation history
	Messages []llm.ChatMessage `json:"messages"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls randomness (0-2)
	Temperature *float64 `json:"temperature,omitempty"`

	// TopP for nucleus sampling (0-1)
	TopP *float64 `json:"top_p,omitempty"`

	// Stop sequences to end generation
	Stop []string `json:"stop,omitempty"`

	// PresencePenalty penalizes new tokens based on presence in text so far
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty penalizes new tokens based on frequency in text so far
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	// User identifier for abuse tracking
	User string `json:"user,omitempty"`
}

// Validate validates the configuration
func (c *ChatCompletionConfig) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	if len(c.Messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}
	return nil
}

// ChatCompletionResult represents the result of a chat completion
type ChatCompletionResult struct {
	// ID is the completion ID from the provider
	ID string `json:"id"`

	// Model is the model that generated the response
	Model string `json:"model"`

	// Role is the message role (usually "assistant")
	Role string `json:"role"`

	// Content is the generated text
	Content string `json:"content"`

	// FinishReason indicates why generation stopped
	FinishReason string `json:"finish_reason"`

	// Usage contains token usage information
	Usage llm.TokenUsage `json:"usage"`
}

// ChatCompletionAction implements the AI ChatCompletion action
type ChatCompletionAction struct {
	credentialService CredentialService
	providerRegistry  ProviderRegistry
}

// NewChatCompletionAction creates a new ChatCompletion action
func NewChatCompletionAction(credentialService CredentialService, providerRegistry ProviderRegistry) *ChatCompletionAction {
	return &ChatCompletionAction{
		credentialService: credentialService,
		providerRegistry:  providerRegistry,
	}
}

// Execute implements the Action interface
func (a *ChatCompletionAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(ChatCompletionConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected ChatCompletionConfig")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Extract tenant_id and credential_id from context
	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	// Retrieve and decrypt credential
	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	// Create provider from credential
	provider, err := a.providerRegistry.GetProviderFromCredential(config.Provider, decryptedCred.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Build request
	chatReq := a.buildChatRequest(&config)

	// Execute chat completion
	chatResp, err := provider.ChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	// Build result
	result := &ChatCompletionResult{
		ID:           chatResp.ID,
		Model:        chatResp.Model,
		Role:         chatResp.Message.Role,
		Content:      chatResp.Message.Content,
		FinishReason: chatResp.FinishReason,
		Usage:        chatResp.Usage,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("provider", config.Provider)
	output.WithMetadata("model", chatResp.Model)
	output.WithMetadata("finish_reason", chatResp.FinishReason)
	output.WithMetadata("prompt_tokens", chatResp.Usage.PromptTokens)
	output.WithMetadata("completion_tokens", chatResp.Usage.CompletionTokens)
	output.WithMetadata("total_tokens", chatResp.Usage.TotalTokens)

	return output, nil
}

// buildChatRequest converts the config to an LLM ChatRequest
func (a *ChatCompletionAction) buildChatRequest(config *ChatCompletionConfig) *llm.ChatRequest {
	req := &llm.ChatRequest{
		Model: config.Model,
	}

	// Add system prompt if provided
	if config.SystemPrompt != "" {
		req.Messages = append(req.Messages, llm.ChatMessage{
			Role:    "system",
			Content: config.SystemPrompt,
		})
	}

	// Add conversation messages
	req.Messages = append(req.Messages, config.Messages...)

	// Set optional parameters
	if config.MaxTokens > 0 {
		req.MaxTokens = config.MaxTokens
	}
	if config.Temperature != nil {
		req.Temperature = config.Temperature
	}
	if config.TopP != nil {
		req.TopP = config.TopP
	}
	if len(config.Stop) > 0 {
		req.Stop = config.Stop
	}
	if config.PresencePenalty != nil {
		req.PresencePenalty = config.PresencePenalty
	}
	if config.FrequencyPenalty != nil {
		req.FrequencyPenalty = config.FrequencyPenalty
	}
	if config.User != "" {
		req.User = config.User
	}

	return req
}

// extractString extracts a string value from a nested map using dot notation
func extractString(data map[string]interface{}, path string) (string, error) {
	keys := parsePath(path)
	current := data

	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key - should be the value
			if val, ok := current[key]; ok {
				if str, ok := val.(string); ok {
					return str, nil
				}
				return "", fmt.Errorf("value at '%s' is not a string", path)
			}
			return "", fmt.Errorf("key '%s' not found in context", path)
		}

		// Intermediate key - should be a map
		if val, ok := current[key]; ok {
			if m, ok := val.(map[string]interface{}); ok {
				current = m
			} else {
				return "", fmt.Errorf("value at '%s' is not a map", key)
			}
		} else {
			return "", fmt.Errorf("key '%s' not found in context", key)
		}
	}

	return "", fmt.Errorf("failed to extract value from path '%s'", path)
}

// parsePath splits a dot-notation path into keys
func parsePath(path string) []string {
	result := []string{}
	current := ""

	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
