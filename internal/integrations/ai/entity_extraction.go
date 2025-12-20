package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/gorax/gorax/internal/llm"
)

// EntityExtractionConfig represents configuration for the EntityExtraction action
type EntityExtractionConfig struct {
	// Provider is the LLM provider name (e.g., "openai", "anthropic", "bedrock")
	Provider string `json:"provider"`

	// Model is the model identifier
	Model string `json:"model"`

	// Text is the text to extract entities from
	Text string `json:"text"`

	// EntityTypes is the list of entity types to extract
	// Common types: "person", "organization", "location", "date", "email", "phone", "money", "product"
	EntityTypes []string `json:"entity_types"`

	// CustomEntities allows defining custom entity types with descriptions
	CustomEntities map[string]string `json:"custom_entities,omitempty"`

	// MaxTokens limits the response length
	MaxTokens int `json:"max_tokens,omitempty"`

	// Temperature controls randomness
	Temperature *float64 `json:"temperature,omitempty"`
}

// Validate validates the configuration
func (c *EntityExtractionConfig) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if c.Model == "" {
		return fmt.Errorf("model is required")
	}
	if c.Text == "" {
		return fmt.Errorf("text is required")
	}
	if len(c.EntityTypes) == 0 && len(c.CustomEntities) == 0 {
		return fmt.Errorf("entity_types or custom_entities is required")
	}
	return nil
}

// ExtractedEntity represents a single extracted entity
type ExtractedEntity struct {
	// Type is the entity type
	Type string `json:"type"`

	// Value is the extracted text
	Value string `json:"value"`

	// NormalizedValue is the standardized form (e.g., normalized date)
	NormalizedValue string `json:"normalized_value,omitempty"`

	// Confidence is the extraction confidence (0-1)
	Confidence float64 `json:"confidence,omitempty"`

	// Context is the surrounding text that provides context
	Context string `json:"context,omitempty"`
}

// EntityExtractionResult represents the result of entity extraction
type EntityExtractionResult struct {
	// Entities is the list of extracted entities
	Entities []ExtractedEntity `json:"entities"`

	// EntityCount is the number of entities found
	EntityCount int `json:"entity_count"`

	// Usage contains token usage information
	Usage llm.TokenUsage `json:"usage"`
}

// EntityExtractionAction implements the AI EntityExtraction action
type EntityExtractionAction struct {
	credentialService CredentialService
	providerRegistry  ProviderRegistry
}

// NewEntityExtractionAction creates a new EntityExtraction action
func NewEntityExtractionAction(credentialService CredentialService, providerRegistry ProviderRegistry) *EntityExtractionAction {
	return &EntityExtractionAction{
		credentialService: credentialService,
		providerRegistry:  providerRegistry,
	}
}

// Execute implements the Action interface
func (a *EntityExtractionAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	// Parse config
	config, ok := input.Config.(EntityExtractionConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected EntityExtractionConfig")
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
	result, err := a.parseResponse(chatResp.Message.Content)
	if err != nil {
		// Return empty result if parsing fails
		result = &EntityExtractionResult{
			Entities:    []ExtractedEntity{},
			EntityCount: 0,
		}
	}
	result.Usage = chatResp.Usage

	// Create output
	output := actions.NewActionOutput(result)
	output.WithMetadata("provider", config.Provider)
	output.WithMetadata("model", chatResp.Model)
	output.WithMetadata("entity_count", result.EntityCount)
	output.WithMetadata("total_tokens", chatResp.Usage.TotalTokens)

	return output, nil
}

// buildChatRequest builds the entity extraction prompt
func (a *EntityExtractionAction) buildChatRequest(config *EntityExtractionConfig) *llm.ChatRequest {
	// Build entity type descriptions
	var entityDescriptions []string
	for _, et := range config.EntityTypes {
		entityDescriptions = append(entityDescriptions, et)
	}
	for et, desc := range config.CustomEntities {
		entityDescriptions = append(entityDescriptions, fmt.Sprintf("%s (%s)", et, desc))
	}

	// Build system prompt
	systemPrompt := `You are an expert entity extraction system. Extract entities from the given text.

Respond with a JSON object in this exact format:
{
  "entities": [
    {
      "type": "entity_type",
      "value": "extracted text",
      "normalized_value": "standardized form (optional)",
      "confidence": 0.95,
      "context": "surrounding text providing context (optional)"
    }
  ]
}

Rules:
- Only extract entities of the specified types
- Use exact text from the source for "value"
- Include "normalized_value" for dates, phone numbers, and amounts
- Set "confidence" between 0 and 1
- Include "context" only if it helps clarify the entity`

	// Build user prompt
	userPrompt := fmt.Sprintf(`Entity types to extract: %s

Text:
%s

Respond with the JSON containing all extracted entities.`, strings.Join(entityDescriptions, ", "), config.Text)

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
		req.MaxTokens = 1000 // Default for entity extraction
	}

	if config.Temperature != nil {
		req.Temperature = config.Temperature
	} else {
		temp := 0.2 // Low temperature for precise extraction
		req.Temperature = &temp
	}

	return req
}

// parseResponse parses the JSON response from the model
func (a *EntityExtractionAction) parseResponse(content string) (*EntityExtractionResult, error) {
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

	var result EntityExtractionResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result.EntityCount = len(result.Entities)
	return &result, nil
}
