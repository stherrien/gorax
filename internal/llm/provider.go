package llm

import (
	"context"
	"time"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// ChatCompletion performs a chat-style completion
	ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// GenerateEmbeddings generates embeddings for input texts
	GenerateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// CountTokens estimates the token count for the given text and model
	CountTokens(text string, model string) (int, error)

	// ListModels returns available models for this provider
	ListModels(ctx context.Context) ([]Model, error)

	// Name returns the provider name (e.g., "openai", "anthropic", "bedrock")
	Name() string

	// HealthCheck verifies the provider connection is valid
	HealthCheck(ctx context.Context) error
}

// ProviderFactory creates a provider with the given configuration
type ProviderFactory func(config *ProviderConfig) (Provider, error)

// ProviderConfig holds configuration for creating a provider
type ProviderConfig struct {
	// APIKey is the authentication key for the provider
	APIKey string

	// Organization is the organization ID (OpenAI specific)
	Organization string

	// Region is the AWS region (Bedrock specific)
	Region string

	// BaseURL allows overriding the default API endpoint (for proxies or testing)
	BaseURL string

	// MaxRetries is the maximum number of retry attempts for transient errors
	MaxRetries int

	// Timeout is the request timeout duration
	Timeout time.Duration

	// AWSAccessKeyID is for AWS Bedrock authentication
	AWSAccessKeyID string

	// AWSSecretAccessKey is for AWS Bedrock authentication
	AWSSecretAccessKey string
}

// DefaultProviderConfig returns a ProviderConfig with sensible defaults
func DefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		MaxRetries: 3,
		Timeout:    60 * time.Second,
	}
}

// Model represents an available LLM model
type Model struct {
	// ID is the unique identifier used in API calls
	ID string `json:"id"`

	// Name is the human-readable name
	Name string `json:"name"`

	// Provider is the provider name
	Provider string `json:"provider"`

	// MaxTokens is the maximum output tokens
	MaxTokens int `json:"max_tokens"`

	// ContextWindow is the maximum input context size in tokens
	ContextWindow int `json:"context_window"`

	// InputCostPer1M is the cost in USD cents per 1 million input tokens
	InputCostPer1M int `json:"input_cost_per_1m"`

	// OutputCostPer1M is the cost in USD cents per 1 million output tokens
	OutputCostPer1M int `json:"output_cost_per_1m"`

	// Capabilities lists what the model can do
	Capabilities []string `json:"capabilities,omitempty"`
}

// ModelCapability constants
const (
	CapabilityChat       = "chat"
	CapabilityCompletion = "completion"
	CapabilityEmbedding  = "embedding"
	CapabilityVision     = "vision"
	CapabilityFunction   = "function_calling"
	CapabilityJSON       = "json_mode"
)

// HasCapability checks if a model has a specific capability
func (m *Model) HasCapability(capability string) bool {
	for _, cap := range m.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}
