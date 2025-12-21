package suggestions

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorCategory_String(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		expected string
	}{
		{ErrorCategoryNetwork, "network"},
		{ErrorCategoryAuth, "auth"},
		{ErrorCategoryData, "data"},
		{ErrorCategoryRateLimit, "rate_limit"},
		{ErrorCategoryTimeout, "timeout"},
		{ErrorCategoryConfig, "config"},
		{ErrorCategoryExternal, "external_service"},
		{ErrorCategoryUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.category))
		})
	}
}

func TestSuggestionType_String(t *testing.T) {
	tests := []struct {
		suggType SuggestionType
		expected string
	}{
		{SuggestionTypeRetry, "retry"},
		{SuggestionTypeConfigChange, "config_change"},
		{SuggestionTypeCredential, "credential_update"},
		{SuggestionTypeDataFix, "data_fix"},
		{SuggestionTypeWorkflowFix, "workflow_modification"},
		{SuggestionTypeManual, "manual_intervention"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.suggType))
		})
	}
}

func TestSuggestionConfidence_String(t *testing.T) {
	tests := []struct {
		confidence SuggestionConfidence
		expected   string
	}{
		{ConfidenceHigh, "high"},
		{ConfidenceMedium, "medium"},
		{ConfidenceLow, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.confidence))
		})
	}
}

func TestSuggestion_JSON(t *testing.T) {
	suggestion := &Suggestion{
		ID:          "sugg-123",
		TenantID:    "tenant-456",
		ExecutionID: "exec-789",
		NodeID:      "node-abc",
		Category:    ErrorCategoryNetwork,
		Type:        SuggestionTypeRetry,
		Confidence:  ConfidenceHigh,
		Title:       "Connection Refused",
		Description: "The target service is not accepting connections",
		Details:     "Check if the service is running",
		Fix: &SuggestionFix{
			ActionType: "retry_with_backoff",
			RetryConfig: &RetryConfig{
				MaxRetries:    5,
				BackoffMs:     1000,
				BackoffFactor: 2.0,
			},
		},
		Source:    SourcePattern,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(suggestion)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var decoded Suggestion
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, suggestion.ID, decoded.ID)
	assert.Equal(t, suggestion.Category, decoded.Category)
	assert.Equal(t, suggestion.Type, decoded.Type)
	assert.Equal(t, suggestion.Confidence, decoded.Confidence)
	assert.Equal(t, suggestion.Title, decoded.Title)
	assert.Equal(t, suggestion.Source, decoded.Source)
	assert.NotNil(t, decoded.Fix)
	assert.Equal(t, suggestion.Fix.ActionType, decoded.Fix.ActionType)
	assert.NotNil(t, decoded.Fix.RetryConfig)
	assert.Equal(t, 5, decoded.Fix.RetryConfig.MaxRetries)
}

func TestErrorContext_JSON(t *testing.T) {
	ctx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection refused",
		ErrorCode:    "ECONNREFUSED",
		HTTPStatus:   0,
		RetryCount:   3,
		InputData: map[string]interface{}{
			"url": "https://api.example.com",
		},
		NodeConfig: map[string]interface{}{
			"method":  "POST",
			"timeout": 30,
		},
		Timestamp: time.Now(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(ctx)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var decoded ErrorContext
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, ctx.ExecutionID, decoded.ExecutionID)
	assert.Equal(t, ctx.NodeType, decoded.NodeType)
	assert.Equal(t, ctx.ErrorMessage, decoded.ErrorMessage)
	assert.Equal(t, ctx.RetryCount, decoded.RetryCount)
	assert.NotNil(t, decoded.InputData)
	assert.NotNil(t, decoded.NodeConfig)
}

func TestSuggestionFix_ConfigChange(t *testing.T) {
	fix := &SuggestionFix{
		ActionType: "config_change",
		ConfigPath: "timeout",
		OldValue:   30,
		NewValue:   60,
	}

	data, err := json.Marshal(fix)
	require.NoError(t, err)

	var decoded SuggestionFix
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "config_change", decoded.ActionType)
	assert.Equal(t, "timeout", decoded.ConfigPath)
	assert.Equal(t, float64(30), decoded.OldValue)
	assert.Equal(t, float64(60), decoded.NewValue)
}

func TestSuggestionStatus_Constants(t *testing.T) {
	assert.Equal(t, SuggestionStatus("pending"), StatusPending)
	assert.Equal(t, SuggestionStatus("applied"), StatusApplied)
	assert.Equal(t, SuggestionStatus("dismissed"), StatusDismissed)
}

func TestSuggestionSource_Constants(t *testing.T) {
	assert.Equal(t, SuggestionSource("pattern"), SourcePattern)
	assert.Equal(t, SuggestionSource("llm"), SourceLLM)
}

func TestNewErrorContext(t *testing.T) {
	ctx := NewErrorContext(
		"exec-123",
		"wf-456",
		"node-789",
		"action:http",
		"connection refused",
	)

	assert.Equal(t, "exec-123", ctx.ExecutionID)
	assert.Equal(t, "wf-456", ctx.WorkflowID)
	assert.Equal(t, "node-789", ctx.NodeID)
	assert.Equal(t, "action:http", ctx.NodeType)
	assert.Equal(t, "connection refused", ctx.ErrorMessage)
	assert.NotZero(t, ctx.Timestamp)
}

func TestNewSuggestion(t *testing.T) {
	sugg := NewSuggestion(
		"tenant-123",
		"exec-456",
		"node-789",
		ErrorCategoryAuth,
		SuggestionTypeCredential,
		ConfidenceHigh,
		"Authentication Failed",
		"Invalid credentials",
		SourcePattern,
	)

	assert.NotEmpty(t, sugg.ID)
	assert.Equal(t, "tenant-123", sugg.TenantID)
	assert.Equal(t, "exec-456", sugg.ExecutionID)
	assert.Equal(t, "node-789", sugg.NodeID)
	assert.Equal(t, ErrorCategoryAuth, sugg.Category)
	assert.Equal(t, SuggestionTypeCredential, sugg.Type)
	assert.Equal(t, ConfidenceHigh, sugg.Confidence)
	assert.Equal(t, "Authentication Failed", sugg.Title)
	assert.Equal(t, "Invalid credentials", sugg.Description)
	assert.Equal(t, SourcePattern, sugg.Source)
	assert.Equal(t, StatusPending, sugg.Status)
	assert.NotZero(t, sugg.CreatedAt)
}
