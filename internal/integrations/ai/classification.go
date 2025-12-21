package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/llm"
)

// ClassificationConfig represents configuration for the Classification action
type ClassificationConfig struct {
	// Provider is the LLM provider name (e.g., "openai", "anthropic", "bedrock")
	Provider string `json:"provider"`

	// Model is the model identifier
	Model string `json:"model"`

	// Text is the text to classify
	Text string `json:"text"`

	// Categories is the list of possible categories
	Categories []string `json:"categories"`

	// MultiLabel allows assigning multiple categories
	MultiLabel bool `json:"multi_label,omitempty"`

	// Description provides context about what the categories mean
	Description string `json:"description,omitempty"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls randomness (lower = more deterministic)
	Temperature *float64 `json:"temperature,omitempty"`
}

// Validate validates the configuration
func (c *ClassificationConfig) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	if c.Text == "" {
		return fmt.Errorf("text is required")
	}
	if len(c.Categories) == 0 {
		return fmt.Errorf("categories cannot be empty")
	}
	return nil
}

// ClassificationResult represents the result of a classification
type ClassificationResult struct {
	// Category is the primary assigned category (or categories if multi-label)
	Category string `json:"category"`

	// Categories contains all assigned categories (for multi-label)
	Categories []string `json:"categories,omitempty"`

	// Confidence is the model's confidence level (0-1)
	Confidence float64 `json:"confidence"`

	// Reasoning explains why the category was chosen
	Reasoning string `json:"reasoning"`

	// Usage contains token usage information
	Usage llm.TokenUsage `json:"usage"`
}

// ClassificationAction implements the AI Classification action
type ClassificationAction struct {
	credentialService CredentialService
	providerRegistry  ProviderRegistry
}

// NewClassificationAction creates a new Classification action
func NewClassificationAction(credentialService CredentialService, providerRegistry ProviderRegistry) *ClassificationAction {
	return &ClassificationAction{
		credentialService: credentialService,
		providerRegistry:  providerRegistry,
	}
}

// Execute implements the Action interface
func (a *ClassificationAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(ClassificationConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected ClassificationConfig")
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

	// Parse response
	result, err := a.parseResponse(chatResp.Message.Content, config.MultiLabel)
	if err != nil {
		// Fall back to raw response if parsing fails
		result = &ClassificationResult{
			Category:   chatResp.Message.Content,
			Confidence: 0.5, // Unknown confidence
			Reasoning:  "Unable to parse structured response",
		}
	}
	result.Usage = chatResp.Usage

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("provider", config.Provider)
	output.WithMetadata("model", chatResp.Model)
	output.WithMetadata("category", result.Category)
	output.WithMetadata("confidence", result.Confidence)
	output.WithMetadata("total_tokens", chatResp.Usage.TotalTokens)

	return output, nil
}

// buildChatRequest builds the classification prompt
func (a *ClassificationAction) buildChatRequest(config *ClassificationConfig) *llm.ChatRequest {
	// Build system prompt
	systemPrompt := `You are an expert text classifier. Classify the given text into the provided categories.

Respond with a JSON object in this exact format:
{
  "category": "the primary category",
  "categories": ["category1", "category2"],
  "confidence": 0.95,
  "reasoning": "Brief explanation of why this category was chosen"
}

Rules:
- "category" should be exactly one of the provided categories
- "categories" should only be included if multiple categories apply`

	if config.MultiLabel {
		systemPrompt += `
- You may assign multiple categories if appropriate`
	} else {
		systemPrompt += `
- Assign only ONE category (the most appropriate one)`
	}

	if config.Description != "" {
		systemPrompt += fmt.Sprintf("\n\nContext about the categories: %s", config.Description)
	}

	// Build user prompt
	userPrompt := fmt.Sprintf(`Categories: %s

Text to classify:
%s

Respond with the JSON classification.`, strings.Join(config.Categories, ", "), config.Text)

	req := &llm.ChatRequest{
		Model: config.Model,
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	if config.MaxTokens > 0 {
		req.MaxTokens = config.MaxTokens
	} else {
		req.MaxTokens = 500 // Default limit for classification
	}

	if config.Temperature != nil {
		req.Temperature = config.Temperature
	} else {
		// Lower temperature for more consistent classification
		temp := 0.3
		req.Temperature = &temp
	}

	return req
}

// parseResponse parses the JSON response from the model
func (a *ClassificationAction) parseResponse(content string, multiLabel bool) (*ClassificationResult, error) {
	// Try to extract JSON from the response
	content = strings.TrimSpace(content)

	// Handle markdown code blocks
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	var result ClassificationResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Validate result
	if result.Category == "" && len(result.Categories) > 0 {
		result.Category = result.Categories[0]
	}
	if result.Confidence == 0 {
		result.Confidence = 0.5 // Default confidence
	}

	return &result, nil
}
