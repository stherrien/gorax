package ai

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/llm"
)

// EmbeddingConfig represents configuration for the Embedding action
type EmbeddingConfig struct {
	// Provider is the LLM provider name (e.g., "openai", "bedrock")
	Provider string `json:"provider"`

	// Model is the embedding model identifier (e.g., "text-embedding-3-small")
	Model string `json:"model"`

	// Texts is the list of texts to generate embeddings for
	Texts []string `json:"texts"`

	// User is an optional user identifier for tracking
	User string `json:"user,omitempty"`
}

// Validate validates the configuration
func (c *EmbeddingConfig) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	if len(c.Texts) == 0 {
		return fmt.Errorf("texts cannot be empty")
	}
	return nil
}

// EmbeddingResult represents the result of an embedding generation
type EmbeddingResult struct {
	// Embeddings is the list of embedding vectors
	Embeddings [][]float64 `json:"embeddings"`

	// Dimensions is the dimension of each embedding vector
	Dimensions int `json:"dimensions"`

	// Count is the number of embeddings generated
	Count int `json:"count"`

	// Usage contains token usage information
	Usage llm.TokenUsage `json:"usage"`
}

// EmbeddingAction implements the AI Embedding action
type EmbeddingAction struct {
	credentialService CredentialService
	providerRegistry  ProviderRegistry
}

// NewEmbeddingAction creates a new Embedding action
func NewEmbeddingAction(credentialService CredentialService, providerRegistry ProviderRegistry) *EmbeddingAction {
	return &EmbeddingAction{
		credentialService: credentialService,
		providerRegistry:  providerRegistry,
	}
}

// Execute implements the Action interface
func (a *EmbeddingAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(EmbeddingConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected EmbeddingConfig")
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
	embeddingReq := &llm.EmbeddingRequest{
		Model: config.Model,
		Texts: config.Texts,
		User:  config.User,
	}

	// Execute embedding generation
	embeddingResp, err := provider.GenerateEmbeddings(ctx, embeddingReq)
	if err != nil {
		return nil, err
	}

	// Determine dimensions
	dimensions := 0
	if len(embeddingResp.Embeddings) > 0 {
		dimensions = len(embeddingResp.Embeddings[0])
	}

	// Build result
	result := &EmbeddingResult{
		Embeddings: embeddingResp.Embeddings,
		Dimensions: dimensions,
		Count:      len(embeddingResp.Embeddings),
		Usage:      embeddingResp.Usage,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("provider", config.Provider)
	output.WithMetadata("model", embeddingResp.Model)
	output.WithMetadata("dimensions", dimensions)
	output.WithMetadata("count", result.Count)
	output.WithMetadata("total_tokens", embeddingResp.Usage.TotalTokens)

	return output, nil
}
