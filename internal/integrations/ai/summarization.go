package ai

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/llm"
)

// SummarizationConfig represents configuration for the Summarization action
type SummarizationConfig struct {
	// Provider is the LLM provider name (e.g., "openai", "anthropic", "bedrock")
	Provider string `json:"provider"`

	// Model is the model identifier
	Model string `json:"model"`

	// Text is the text to summarize
	Text string `json:"text"`

	// MaxLength is the target length of the summary in words (approximate)
	MaxLength int `json:"max_length,omitempty"`

	// Format is the summary format: "paragraph" or "bullets"
	Format string `json:"format,omitempty"`

	// Focus is an optional focus area for the summary
	Focus string `json:"focus,omitempty"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls randomness (0-2)
	Temperature *float64 `json:"temperature,omitempty"`
}

// Validate validates the configuration
func (c *SummarizationConfig) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	if c.Text == "" {
		return fmt.Errorf("text is required")
	}
	return nil
}

// SummarizationResult represents the result of a summarization
type SummarizationResult struct {
	// Summary is the generated summary
	Summary string `json:"summary"`

	// WordCount is the approximate word count of the summary
	WordCount int `json:"word_count"`

	// Usage contains token usage information
	Usage llm.TokenUsage `json:"usage"`
}

// SummarizationAction implements the AI Summarization action
type SummarizationAction struct {
	credentialService CredentialService
	providerRegistry  ProviderRegistry
}

// NewSummarizationAction creates a new Summarization action
func NewSummarizationAction(credentialService CredentialService, providerRegistry ProviderRegistry) *SummarizationAction {
	return &SummarizationAction{
		credentialService: credentialService,
		providerRegistry:  providerRegistry,
	}
}

// Execute implements the Action interface
func (a *SummarizationAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(SummarizationConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected SummarizationConfig")
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
	result := &SummarizationResult{
		Summary:   chatResp.Message.Content,
		WordCount: countWords(chatResp.Message.Content),
		Usage:     chatResp.Usage,
	}

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("provider", config.Provider)
	output.WithMetadata("model", chatResp.Model)
	output.WithMetadata("word_count", result.WordCount)
	output.WithMetadata("total_tokens", chatResp.Usage.TotalTokens)

	return output, nil
}

// buildChatRequest builds the summarization prompt
func (a *SummarizationAction) buildChatRequest(config *SummarizationConfig) *llm.ChatRequest {
	// Build system prompt
	systemPrompt := "You are an expert summarizer. Create clear, concise summaries that capture the key points."

	if config.Format == "bullets" {
		systemPrompt += " Format your summary as bullet points."
	}

	if config.Focus != "" {
		systemPrompt += fmt.Sprintf(" Focus specifically on: %s.", config.Focus)
	}

	// Build user prompt
	userPrompt := fmt.Sprintf("Please summarize the following text")
	if config.MaxLength > 0 {
		userPrompt += fmt.Sprintf(" in approximately %d words", config.MaxLength)
	}
	userPrompt += fmt.Sprintf(":\n\n%s", config.Text)

	req := &llm.ChatRequest{
		Model: config.Model,
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	if config.MaxTokens > 0 {
		req.MaxTokens = config.MaxTokens
	}
	if config.Temperature != nil {
		req.Temperature = config.Temperature
	}

	return req
}

// countWords returns an approximate word count
func countWords(text string) int {
	words := 0
	inWord := false
	for _, r := range text {
		if r == ' ' || r == '\n' || r == '\t' {
			if inWord {
				words++
				inWord = false
			}
		} else {
			inWord = true
		}
	}
	if inWord {
		words++
	}
	return words
}
