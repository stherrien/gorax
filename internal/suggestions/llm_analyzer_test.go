package suggestions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/llm"
)

// MockLLMProvider implements llm.Provider for testing
type MockLLMProvider struct {
	response *llm.ChatResponse
	err      error
}

func NewMockLLMProvider() *MockLLMProvider {
	return &MockLLMProvider{}
}

func (m *MockLLMProvider) SetResponse(response *llm.ChatResponse) {
	m.response = response
}

func (m *MockLLMProvider) SetError(err error) {
	m.err = err
}

func (m *MockLLMProvider) ChatCompletion(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}
	// Return empty response if not configured
	return &llm.ChatResponse{
		Message: llm.ChatMessage{
			Role:    "assistant",
			Content: "{}",
		},
	}, nil
}

func (m *MockLLMProvider) GenerateEmbeddings(ctx context.Context, req *llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *MockLLMProvider) CountTokens(text string, model string) (int, error) {
	return len(text) / 4, nil // Rough estimate
}

func (m *MockLLMProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	return []llm.Model{{ID: "test-model"}}, nil
}

func (m *MockLLMProvider) Name() string {
	return "mock"
}

func (m *MockLLMProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func TestLLMAnalyzer_Name(t *testing.T) {
	analyzer := NewLLMAnalyzer(nil, LLMAnalyzerConfig{})
	assert.Equal(t, "llm", analyzer.Name())
}

func TestLLMAnalyzer_CanHandle(t *testing.T) {
	provider := NewMockLLMProvider()
	analyzer := NewLLMAnalyzer(provider, LLMAnalyzerConfig{})

	tests := []struct {
		name     string
		errCtx   *ErrorContext
		expected bool
	}{
		{
			name: "can handle with error message",
			errCtx: &ErrorContext{
				ExecutionID:  "exec-123",
				ErrorMessage: "some error",
			},
			expected: true,
		},
		{
			name: "cannot handle without provider",
			errCtx: &ErrorContext{
				ExecutionID:  "exec-123",
				ErrorMessage: "some error",
			},
			expected: true, // Still can handle, will fail at analysis
		},
		{
			name: "cannot handle without error message",
			errCtx: &ErrorContext{
				ExecutionID: "exec-123",
			},
			expected: false,
		},
		{
			name:     "cannot handle nil context",
			errCtx:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CanHandle(tt.errCtx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLLMAnalyzer_Analyze_Success(t *testing.T) {
	provider := NewMockLLMProvider()

	// Create a mock response with a valid suggestion
	suggestionJSON := LLMSuggestionResponse{
		Suggestions: []LLMSuggestion{
			{
				Category:    "network",
				Type:        "retry",
				Confidence:  "high",
				Title:       "Network Error Detected",
				Description: "The request failed due to a network issue",
				Fix: &SuggestionFix{
					ActionType: "retry_with_backoff",
					RetryConfig: &RetryConfig{
						MaxRetries:    3,
						BackoffMs:     1000,
						BackoffFactor: 2.0,
					},
				},
			},
		},
	}

	responseJSON, _ := json.Marshal(suggestionJSON)
	provider.SetResponse(&llm.ChatResponse{
		Message: llm.ChatMessage{
			Role:    "assistant",
			Content: string(responseJSON),
		},
	})

	analyzer := NewLLMAnalyzer(provider, LLMAnalyzerConfig{
		Model:    "test-model",
		TenantID: "tenant-123",
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "network connection failed",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	require.Len(t, suggestions, 1)

	s := suggestions[0]
	assert.Equal(t, "tenant-123", s.TenantID)
	assert.Equal(t, "exec-123", s.ExecutionID)
	assert.Equal(t, "node-789", s.NodeID)
	assert.Equal(t, ErrorCategoryNetwork, s.Category)
	assert.Equal(t, SuggestionTypeRetry, s.Type)
	assert.Equal(t, ConfidenceHigh, s.Confidence)
	assert.Equal(t, "Network Error Detected", s.Title)
	assert.Equal(t, SourceLLM, s.Source)
	assert.NotNil(t, s.Fix)
}

func TestLLMAnalyzer_Analyze_MultipleSuggestions(t *testing.T) {
	provider := NewMockLLMProvider()

	suggestionJSON := LLMSuggestionResponse{
		Suggestions: []LLMSuggestion{
			{
				Category:    "network",
				Type:        "retry",
				Confidence:  "high",
				Title:       "Retry Request",
				Description: "Network issue, retry recommended",
			},
			{
				Category:    "config",
				Type:        "config_change",
				Confidence:  "medium",
				Title:       "Increase Timeout",
				Description: "Consider increasing timeout",
			},
		},
	}

	responseJSON, _ := json.Marshal(suggestionJSON)
	provider.SetResponse(&llm.ChatResponse{
		Message: llm.ChatMessage{
			Role:    "assistant",
			Content: string(responseJSON),
		},
	})

	analyzer := NewLLMAnalyzer(provider, LLMAnalyzerConfig{
		Model: "test-model",
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		ErrorMessage: "connection timeout",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	assert.Len(t, suggestions, 2)
}

func TestLLMAnalyzer_Analyze_ProviderError(t *testing.T) {
	provider := NewMockLLMProvider()
	provider.SetError(errors.New("API error"))

	analyzer := NewLLMAnalyzer(provider, LLMAnalyzerConfig{
		Model: "test-model",
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		ErrorMessage: "some error",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.Error(t, err)
	assert.Nil(t, suggestions)
	assert.Contains(t, err.Error(), "API error")
}

func TestLLMAnalyzer_Analyze_InvalidJSON(t *testing.T) {
	provider := NewMockLLMProvider()
	provider.SetResponse(&llm.ChatResponse{
		Message: llm.ChatMessage{
			Role:    "assistant",
			Content: "This is not valid JSON",
		},
	})

	analyzer := NewLLMAnalyzer(provider, LLMAnalyzerConfig{
		Model: "test-model",
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		ErrorMessage: "some error",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.Error(t, err)
	assert.Nil(t, suggestions)
}

func TestLLMAnalyzer_Analyze_EmptyResponse(t *testing.T) {
	provider := NewMockLLMProvider()

	suggestionJSON := LLMSuggestionResponse{
		Suggestions: []LLMSuggestion{},
	}

	responseJSON, _ := json.Marshal(suggestionJSON)
	provider.SetResponse(&llm.ChatResponse{
		Message: llm.ChatMessage{
			Role:    "assistant",
			Content: string(responseJSON),
		},
	})

	analyzer := NewLLMAnalyzer(provider, LLMAnalyzerConfig{
		Model: "test-model",
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		ErrorMessage: "some error",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	assert.Empty(t, suggestions)
}

func TestLLMAnalyzer_Analyze_NilProvider(t *testing.T) {
	analyzer := NewLLMAnalyzer(nil, LLMAnalyzerConfig{
		Model: "test-model",
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		ErrorMessage: "some error",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.Error(t, err)
	assert.Nil(t, suggestions)
	assert.Contains(t, err.Error(), "no LLM provider")
}

func TestLLMAnalyzer_Analyze_WithInputData(t *testing.T) {
	provider := NewMockLLMProvider()

	suggestionJSON := LLMSuggestionResponse{
		Suggestions: []LLMSuggestion{
			{
				Category:    "data",
				Type:        "data_fix",
				Confidence:  "medium",
				Title:       "Invalid Data Format",
				Description: "The URL field is malformed",
			},
		},
	}

	responseJSON, _ := json.Marshal(suggestionJSON)
	provider.SetResponse(&llm.ChatResponse{
		Message: llm.ChatMessage{
			Role:    "assistant",
			Content: string(responseJSON),
		},
	})

	analyzer := NewLLMAnalyzer(provider, LLMAnalyzerConfig{
		Model: "test-model",
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		ErrorMessage: "invalid URL format",
		InputData: map[string]interface{}{
			"url": "not-a-valid-url",
		},
		NodeConfig: map[string]interface{}{
			"method": "POST",
		},
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	require.Len(t, suggestions, 1)
	assert.Equal(t, ErrorCategoryData, suggestions[0].Category)
}

func TestLLMAnalyzer_BuildPrompt(t *testing.T) {
	analyzer := NewLLMAnalyzer(nil, LLMAnalyzerConfig{})

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection refused",
		HTTPStatus:   0,
		RetryCount:   2,
		InputData: map[string]interface{}{
			"url": "https://api.example.com",
		},
	}

	prompt := analyzer.buildPrompt(errCtx)

	// Verify prompt contains key information
	assert.Contains(t, prompt, "connection refused")
	assert.Contains(t, prompt, "action:http")
	assert.Contains(t, prompt, "Retry Count: 2")
	assert.Contains(t, prompt, "api.example.com")
}

func TestLLMAnalyzer_ParseResponse_WithMarkdown(t *testing.T) {
	provider := NewMockLLMProvider()

	// Sometimes LLMs wrap JSON in markdown code blocks
	wrappedJSON := "```json\n{\"suggestions\":[{\"category\":\"network\",\"type\":\"retry\",\"confidence\":\"high\",\"title\":\"Test\",\"description\":\"Test description\"}]}\n```"

	provider.SetResponse(&llm.ChatResponse{
		Message: llm.ChatMessage{
			Role:    "assistant",
			Content: wrappedJSON,
		},
	})

	analyzer := NewLLMAnalyzer(provider, LLMAnalyzerConfig{
		Model: "test-model",
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		ErrorMessage: "test error",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	require.Len(t, suggestions, 1)
}

func TestLLMAnalyzer_ValidateCategory(t *testing.T) {
	tests := []struct {
		input    string
		expected ErrorCategory
	}{
		{"network", ErrorCategoryNetwork},
		{"auth", ErrorCategoryAuth},
		{"data", ErrorCategoryData},
		{"rate_limit", ErrorCategoryRateLimit},
		{"timeout", ErrorCategoryTimeout},
		{"config", ErrorCategoryConfig},
		{"external_service", ErrorCategoryExternal},
		{"unknown", ErrorCategoryUnknown},
		{"invalid", ErrorCategoryUnknown},
		{"", ErrorCategoryUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := validateCategory(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLLMAnalyzer_ValidateType(t *testing.T) {
	tests := []struct {
		input    string
		expected SuggestionType
	}{
		{"retry", SuggestionTypeRetry},
		{"config_change", SuggestionTypeConfigChange},
		{"credential_update", SuggestionTypeCredential},
		{"data_fix", SuggestionTypeDataFix},
		{"workflow_modification", SuggestionTypeWorkflowFix},
		{"manual_intervention", SuggestionTypeManual},
		{"invalid", SuggestionTypeManual},
		{"", SuggestionTypeManual},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := validateType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLLMAnalyzer_ValidateConfidence(t *testing.T) {
	tests := []struct {
		input    string
		expected SuggestionConfidence
	}{
		{"high", ConfidenceHigh},
		{"medium", ConfidenceMedium},
		{"low", ConfidenceLow},
		{"invalid", ConfidenceLow},
		{"", ConfidenceLow},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := validateConfidence(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLLMAnalyzerConfig_Defaults(t *testing.T) {
	config := DefaultLLMAnalyzerConfig()

	assert.NotEmpty(t, config.Model)
	assert.Greater(t, config.MaxTokens, 0)
	assert.NotNil(t, config.Temperature)
}
