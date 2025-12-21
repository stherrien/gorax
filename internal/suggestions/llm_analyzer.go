package suggestions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gorax/gorax/internal/llm"
)

// LLMAnalyzerConfig holds configuration for the LLM analyzer
type LLMAnalyzerConfig struct {
	// Model is the LLM model to use
	Model string

	// MaxTokens limits the response length
	MaxTokens int

	// Temperature controls randomness (0.0 to 1.0)
	Temperature *float64

	// TenantID is the tenant for suggestions
	TenantID string
}

// DefaultLLMAnalyzerConfig returns default configuration
func DefaultLLMAnalyzerConfig() LLMAnalyzerConfig {
	temp := 0.3
	return LLMAnalyzerConfig{
		Model:       "gpt-4o-mini",
		MaxTokens:   1024,
		Temperature: &temp,
	}
}

// LLMSuggestionResponse is the expected JSON structure from the LLM
type LLMSuggestionResponse struct {
	Suggestions []LLMSuggestion `json:"suggestions"`
}

// LLMSuggestion is a single suggestion from the LLM
type LLMSuggestion struct {
	Category    string         `json:"category"`
	Type        string         `json:"type"`
	Confidence  string         `json:"confidence"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Details     string         `json:"details,omitempty"`
	Fix         *SuggestionFix `json:"fix,omitempty"`
}

// LLMAnalyzer uses an LLM to analyze errors and suggest fixes
type LLMAnalyzer struct {
	provider llm.Provider
	config   LLMAnalyzerConfig
}

// NewLLMAnalyzer creates a new LLM analyzer
func NewLLMAnalyzer(provider llm.Provider, config LLMAnalyzerConfig) *LLMAnalyzer {
	if config.Model == "" {
		config.Model = DefaultLLMAnalyzerConfig().Model
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = DefaultLLMAnalyzerConfig().MaxTokens
	}
	return &LLMAnalyzer{
		provider: provider,
		config:   config,
	}
}

// Name returns the analyzer name
func (a *LLMAnalyzer) Name() string {
	return "llm"
}

// CanHandle returns true if this analyzer can handle the error context
func (a *LLMAnalyzer) CanHandle(errCtx *ErrorContext) bool {
	if errCtx == nil {
		return false
	}
	return errCtx.ErrorMessage != ""
}

// Analyze uses the LLM to analyze the error and generate suggestions
func (a *LLMAnalyzer) Analyze(ctx context.Context, errCtx *ErrorContext) ([]*Suggestion, error) {
	if a.provider == nil {
		return nil, fmt.Errorf("no LLM provider configured")
	}

	if !a.CanHandle(errCtx) {
		return nil, nil
	}

	// Build the prompt
	prompt := a.buildPrompt(errCtx)

	// Create the request
	req := &llm.ChatRequest{
		Model: a.config.Model,
		Messages: []llm.ChatMessage{
			llm.SystemMessage(systemPrompt),
			llm.UserMessage(prompt),
		},
		MaxTokens:   a.config.MaxTokens,
		Temperature: a.config.Temperature,
		ResponseFormat: &llm.ResponseFormat{
			Type: llm.ResponseFormatJSON,
		},
	}

	// Call the LLM
	resp, err := a.provider.ChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM error: %w", err)
	}

	// Parse the response
	suggestions, err := a.parseResponse(resp.Message.Content, errCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return suggestions, nil
}

func (a *LLMAnalyzer) buildPrompt(errCtx *ErrorContext) string {
	var sb strings.Builder

	sb.WriteString("Analyze the following workflow execution error and provide suggestions:\n\n")
	sb.WriteString(fmt.Sprintf("Node Type: %s\n", errCtx.NodeType))
	sb.WriteString(fmt.Sprintf("Error Message: %s\n", errCtx.ErrorMessage))

	if errCtx.ErrorCode != "" {
		sb.WriteString(fmt.Sprintf("Error Code: %s\n", errCtx.ErrorCode))
	}

	if errCtx.HTTPStatus > 0 {
		sb.WriteString(fmt.Sprintf("HTTP Status: %d\n", errCtx.HTTPStatus))
	}

	sb.WriteString(fmt.Sprintf("Retry Count: %d\n", errCtx.RetryCount))

	if len(errCtx.InputData) > 0 {
		inputJSON, _ := json.Marshal(errCtx.InputData)
		sb.WriteString(fmt.Sprintf("\nInput Data:\n%s\n", string(inputJSON)))
	}

	if len(errCtx.NodeConfig) > 0 {
		configJSON, _ := json.Marshal(errCtx.NodeConfig)
		sb.WriteString(fmt.Sprintf("\nNode Config:\n%s\n", string(configJSON)))
	}

	return sb.String()
}

func (a *LLMAnalyzer) parseResponse(content string, errCtx *ErrorContext) ([]*Suggestion, error) {
	// Handle markdown code blocks
	content = stripMarkdownCodeBlock(content)

	var response LLMSuggestionResponse
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	if len(response.Suggestions) == 0 {
		return nil, nil
	}

	suggestions := make([]*Suggestion, 0, len(response.Suggestions))
	for _, llmSugg := range response.Suggestions {
		suggestion := a.convertToSuggestion(llmSugg, errCtx)
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

func (a *LLMAnalyzer) convertToSuggestion(llmSugg LLMSuggestion, errCtx *ErrorContext) *Suggestion {
	suggestion := NewSuggestion(
		a.config.TenantID,
		errCtx.ExecutionID,
		errCtx.NodeID,
		validateCategory(llmSugg.Category),
		validateType(llmSugg.Type),
		validateConfidence(llmSugg.Confidence),
		llmSugg.Title,
		llmSugg.Description,
		SourceLLM,
	)

	if llmSugg.Details != "" {
		suggestion.Details = llmSugg.Details
	}

	if llmSugg.Fix != nil {
		suggestion.Fix = llmSugg.Fix
	}

	return suggestion
}

// stripMarkdownCodeBlock removes markdown code block wrappers if present
func stripMarkdownCodeBlock(content string) string {
	content = strings.TrimSpace(content)

	// Check for ```json wrapper
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	return content
}

// validateCategory validates and normalizes a category string
func validateCategory(category string) ErrorCategory {
	switch strings.ToLower(category) {
	case "network":
		return ErrorCategoryNetwork
	case "auth":
		return ErrorCategoryAuth
	case "data":
		return ErrorCategoryData
	case "rate_limit":
		return ErrorCategoryRateLimit
	case "timeout":
		return ErrorCategoryTimeout
	case "config":
		return ErrorCategoryConfig
	case "external_service":
		return ErrorCategoryExternal
	default:
		return ErrorCategoryUnknown
	}
}

// validateType validates and normalizes a suggestion type string
func validateType(suggType string) SuggestionType {
	switch strings.ToLower(suggType) {
	case "retry":
		return SuggestionTypeRetry
	case "config_change":
		return SuggestionTypeConfigChange
	case "credential_update":
		return SuggestionTypeCredential
	case "data_fix":
		return SuggestionTypeDataFix
	case "workflow_modification":
		return SuggestionTypeWorkflowFix
	default:
		return SuggestionTypeManual
	}
}

// validateConfidence validates and normalizes a confidence string
func validateConfidence(confidence string) SuggestionConfidence {
	switch strings.ToLower(confidence) {
	case "high":
		return ConfidenceHigh
	case "medium":
		return ConfidenceMedium
	default:
		return ConfidenceLow
	}
}

const systemPrompt = `You are an expert workflow automation error analyzer. Your task is to analyze error messages from workflow executions and provide actionable suggestions to fix them.

For each error, you must respond with a JSON object containing an array of suggestions. Each suggestion should have:
- category: The error category (one of: network, auth, data, rate_limit, timeout, config, external_service, unknown)
- type: The suggestion type (one of: retry, config_change, credential_update, data_fix, workflow_modification, manual_intervention)
- confidence: Your confidence level (one of: high, medium, low)
- title: A short title for the suggestion
- description: A detailed description of what went wrong and how to fix it
- details: Optional additional context
- fix: Optional object with actionable fix data:
  - action_type: The type of fix action
  - config_path: For config changes, the config key to modify
  - old_value: The current value (if known)
  - new_value: The suggested new value
  - retry_config: For retry suggestions, object with max_retries, backoff_ms, backoff_factor

Guidelines:
1. Be specific and actionable in your suggestions
2. Prioritize suggestions by their likelihood of fixing the issue
3. For authentication errors, always suggest credential_update
4. For network errors, consider both retry and config_change options
5. For rate limit errors, suggest adding delays or reducing concurrency
6. If the error is unclear, provide your best analysis with low confidence
7. Return an empty suggestions array if you cannot provide any useful suggestions

Always respond with valid JSON in this format:
{
  "suggestions": [
    {
      "category": "network",
      "type": "retry",
      "confidence": "high",
      "title": "Connection Error",
      "description": "The connection was refused...",
      "fix": {
        "action_type": "retry_with_backoff",
        "retry_config": {
          "max_retries": 3,
          "backoff_ms": 1000,
          "backoff_factor": 2.0
        }
      }
    }
  ]
}`
