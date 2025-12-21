package suggestions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gorax/gorax/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuggestionService_AnalyzeError(t *testing.T) {
	repo := NewMockRepository()

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository:      repo,
		PatternAnalyzer: NewPatternAnalyzer(nil),
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection refused",
	}

	suggestions, err := service.AnalyzeError(ctx, "tenant-123", errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	// Should have stored in repository
	stored, err := repo.GetByExecutionID(ctx, "tenant-123", "exec-123")
	require.NoError(t, err)
	assert.Len(t, stored, len(suggestions))
}

func TestSuggestionService_AnalyzeError_NoMatch(t *testing.T) {
	repo := NewMockRepository()

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository:      repo,
		PatternAnalyzer: NewPatternAnalyzer(nil),
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "unique random error xyz12345",
	}

	suggestions, err := service.AnalyzeError(ctx, "tenant-123", errCtx)
	require.NoError(t, err)
	assert.Empty(t, suggestions)
}

func TestSuggestionService_GetSuggestions(t *testing.T) {
	repo := NewMockRepository()

	// Pre-populate some suggestions
	s1 := NewSuggestion("tenant-123", "exec-456", "node-1", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "S1", "D1", SourcePattern)
	s2 := NewSuggestion("tenant-123", "exec-456", "node-2", ErrorCategoryAuth, SuggestionTypeCredential, ConfidenceMedium, "S2", "D2", SourceLLM)
	repo.Create(context.Background(), s1)
	repo.Create(context.Background(), s2)

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository: repo,
	})

	ctx := context.Background()
	suggestions, err := service.GetSuggestions(ctx, "tenant-123", "exec-456")
	require.NoError(t, err)
	assert.Len(t, suggestions, 2)
}

func TestSuggestionService_GetSuggestionByID(t *testing.T) {
	repo := NewMockRepository()

	s := NewSuggestion("tenant-123", "exec-456", "node-789", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "Test", "Test description", SourcePattern)
	repo.Create(context.Background(), s)

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository: repo,
	})

	ctx := context.Background()
	retrieved, err := service.GetSuggestionByID(ctx, "tenant-123", s.ID)
	require.NoError(t, err)
	assert.Equal(t, s.ID, retrieved.ID)
	assert.Equal(t, s.Title, retrieved.Title)
}

func TestSuggestionService_GetSuggestionByID_NotFound(t *testing.T) {
	repo := NewMockRepository()

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository: repo,
	})

	ctx := context.Background()
	_, err := service.GetSuggestionByID(ctx, "tenant-123", "nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSuggestionNotFound)
}

func TestSuggestionService_ApplySuggestion(t *testing.T) {
	repo := NewMockRepository()

	s := NewSuggestion("tenant-123", "exec-456", "node-789", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "Test", "Test", SourcePattern)
	repo.Create(context.Background(), s)

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository: repo,
	})

	ctx := context.Background()
	err := service.ApplySuggestion(ctx, "tenant-123", s.ID)
	require.NoError(t, err)

	// Verify status was updated
	retrieved, err := repo.GetByID(ctx, "tenant-123", s.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusApplied, retrieved.Status)
	assert.NotNil(t, retrieved.AppliedAt)
}

func TestSuggestionService_ApplySuggestion_NotFound(t *testing.T) {
	repo := NewMockRepository()

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository: repo,
	})

	ctx := context.Background()
	err := service.ApplySuggestion(ctx, "tenant-123", "nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSuggestionNotFound)
}

func TestSuggestionService_DismissSuggestion(t *testing.T) {
	repo := NewMockRepository()

	s := NewSuggestion("tenant-123", "exec-456", "node-789", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "Test", "Test", SourcePattern)
	repo.Create(context.Background(), s)

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository: repo,
	})

	ctx := context.Background()
	err := service.DismissSuggestion(ctx, "tenant-123", s.ID)
	require.NoError(t, err)

	// Verify status was updated
	retrieved, err := repo.GetByID(ctx, "tenant-123", s.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusDismissed, retrieved.Status)
	assert.NotNil(t, retrieved.DismissedAt)
}

func TestSuggestionService_WithLLMAnalyzer(t *testing.T) {
	repo := NewMockRepository()
	mockProvider := NewMockLLMProvider()

	// Configure mock to return a suggestion
	mockProvider.SetResponse(createMockLLMChatResponse("network", "retry", "high", "LLM Suggestion", "LLM Description"))

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository:      repo,
		PatternAnalyzer: NewPatternAnalyzer(nil),
		LLMAnalyzer: NewLLMAnalyzer(mockProvider, LLMAnalyzerConfig{
			Model:    "test-model",
			TenantID: "tenant-123",
		}),
		UseLLMForUnmatched: true,
	})

	ctx := context.Background()
	// Use an error that won't match patterns but LLM can analyze
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:custom",
		ErrorMessage: "unique custom error xyz12345",
	}

	suggestions, err := service.AnalyzeError(ctx, "tenant-123", errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	// Should have used LLM since patterns didn't match
	foundLLM := false
	for _, s := range suggestions {
		if s.Source == SourceLLM {
			foundLLM = true
			break
		}
	}
	assert.True(t, foundLLM, "should have LLM-generated suggestion")
}

func TestSuggestionService_PatternTakesPrecedence(t *testing.T) {
	repo := NewMockRepository()
	mockProvider := NewMockLLMProvider()

	// Configure mock to return a different suggestion
	mockProvider.SetResponse(createMockLLMChatResponse("data", "data_fix", "low", "LLM Suggestion", "LLM Description"))

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository:      repo,
		PatternAnalyzer: NewPatternAnalyzer(nil),
		LLMAnalyzer: NewLLMAnalyzer(mockProvider, LLMAnalyzerConfig{
			Model: "test-model",
		}),
		UseLLMForUnmatched: true,
	})

	ctx := context.Background()
	// Use an error that matches patterns
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection refused",
	}

	suggestions, err := service.AnalyzeError(ctx, "tenant-123", errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	// First suggestion should be from pattern (higher confidence)
	assert.Equal(t, SourcePattern, suggestions[0].Source)
}

func TestSuggestionService_RepositoryError(t *testing.T) {
	repo := NewMockRepository()
	repo.SetError(errors.New("database error"))

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository:      repo,
		PatternAnalyzer: NewPatternAnalyzer(nil),
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		ErrorMessage: "connection refused",
	}

	// Analysis should still work, but storage fails
	suggestions, err := service.AnalyzeError(ctx, "tenant-123", errCtx)
	// Returns suggestions even if storage fails
	require.NotEmpty(t, suggestions)
	// But should have logged the error (we can't verify logging easily)
	_ = err // Error is logged but not returned to not block analysis
}

func TestSuggestionService_DeleteSuggestion(t *testing.T) {
	repo := NewMockRepository()

	s := NewSuggestion("tenant-123", "exec-456", "node-789", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "Test", "Test", SourcePattern)
	repo.Create(context.Background(), s)

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository: repo,
	})

	ctx := context.Background()
	err := service.DeleteSuggestion(ctx, "tenant-123", s.ID)
	require.NoError(t, err)

	// Should be gone
	_, err = repo.GetByID(ctx, "tenant-123", s.ID)
	assert.ErrorIs(t, err, ErrSuggestionNotFound)
}

func TestSuggestionService_DeleteByExecutionID(t *testing.T) {
	repo := NewMockRepository()

	// Create multiple suggestions
	s1 := NewSuggestion("tenant-123", "exec-456", "node-1", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "S1", "D1", SourcePattern)
	s2 := NewSuggestion("tenant-123", "exec-456", "node-2", ErrorCategoryAuth, SuggestionTypeCredential, ConfidenceMedium, "S2", "D2", SourcePattern)
	s3 := NewSuggestion("tenant-123", "exec-other", "node-3", ErrorCategoryData, SuggestionTypeDataFix, ConfidenceLow, "S3", "D3", SourcePattern)
	repo.Create(context.Background(), s1)
	repo.Create(context.Background(), s2)
	repo.Create(context.Background(), s3)

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository: repo,
	})

	ctx := context.Background()
	err := service.DeleteByExecutionID(ctx, "tenant-123", "exec-456")
	require.NoError(t, err)

	// s1 and s2 should be gone
	suggestions, err := repo.GetByExecutionID(ctx, "tenant-123", "exec-456")
	require.NoError(t, err)
	assert.Empty(t, suggestions)

	// s3 should still exist
	retrieved, err := repo.GetByID(ctx, "tenant-123", s3.ID)
	require.NoError(t, err)
	assert.Equal(t, s3.ID, retrieved.ID)
}

func TestSuggestionService_ReanalyzeError(t *testing.T) {
	repo := NewMockRepository()

	// Pre-populate with old suggestions
	s := NewSuggestion("tenant-123", "exec-456", "node-789", ErrorCategoryData, SuggestionTypeDataFix, ConfidenceLow, "Old", "Old suggestion", SourcePattern)
	repo.Create(context.Background(), s)

	service := NewSuggestionService(SuggestionServiceConfig{
		Repository:      repo,
		PatternAnalyzer: NewPatternAnalyzer(nil),
	})

	ctx := context.Background()
	errCtx := &ErrorContext{
		ExecutionID:  "exec-456",
		WorkflowID:   "wf-789",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection refused",
	}

	// Reanalyze should delete old and create new
	suggestions, err := service.ReanalyzeError(ctx, "tenant-123", errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	// Old suggestion should be gone, new one should exist
	all, err := repo.GetByExecutionID(ctx, "tenant-123", "exec-456")
	require.NoError(t, err)

	// Should have new network error suggestion, not the old data fix
	for _, sugg := range all {
		assert.NotEqual(t, s.ID, sugg.ID, "old suggestion should be deleted")
		assert.Equal(t, ErrorCategoryNetwork, sugg.Category)
	}
}

// Helper to create mock LLM chat response
func createMockLLMChatResponse(category, suggType, confidence, title, description string) *llm.ChatResponse {
	response := LLMSuggestionResponse{
		Suggestions: []LLMSuggestion{
			{
				Category:    category,
				Type:        suggType,
				Confidence:  confidence,
				Title:       title,
				Description: description,
			},
		},
	}
	responseJSON, _ := json.Marshal(response)
	return &llm.ChatResponse{
		Message: llm.ChatMessage{
			Role:    "assistant",
			Content: string(responseJSON),
		},
	}
}
