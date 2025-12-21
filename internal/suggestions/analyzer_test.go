package suggestions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatternAnalyzer_Name(t *testing.T) {
	analyzer := NewPatternAnalyzer(nil)
	assert.Equal(t, "pattern", analyzer.Name())
}

func TestPatternAnalyzer_CanHandle(t *testing.T) {
	analyzer := NewPatternAnalyzer(nil)

	tests := []struct {
		name     string
		errCtx   *ErrorContext
		expected bool
	}{
		{
			name: "can handle with error message",
			errCtx: &ErrorContext{
				ExecutionID:  "exec-123",
				ErrorMessage: "connection refused",
			},
			expected: true,
		},
		{
			name: "can handle with HTTP status",
			errCtx: &ErrorContext{
				ExecutionID: "exec-123",
				HTTPStatus:  500,
			},
			expected: true,
		},
		{
			name: "can handle with both",
			errCtx: &ErrorContext{
				ExecutionID:  "exec-123",
				ErrorMessage: "error",
				HTTPStatus:   500,
			},
			expected: true,
		},
		{
			name: "cannot handle without error info",
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

func TestPatternAnalyzer_Analyze_ConnectionRefused(t *testing.T) {
	analyzer := NewPatternAnalyzer(nil)
	ctx := context.Background()

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "dial tcp 127.0.0.1:8080: connection refused",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	s := suggestions[0]
	assert.Equal(t, ErrorCategoryNetwork, s.Category)
	assert.Equal(t, SuggestionTypeRetry, s.Type)
	assert.Equal(t, ConfidenceHigh, s.Confidence)
	assert.Contains(t, s.Title, "Connection")
	assert.NotNil(t, s.Fix)
	assert.Equal(t, "retry_with_backoff", s.Fix.ActionType)
}

func TestPatternAnalyzer_Analyze_AuthError(t *testing.T) {
	analyzer := NewPatternAnalyzer(nil)
	ctx := context.Background()

	tests := []struct {
		name       string
		httpStatus int
		title      string
	}{
		{
			name:       "401 Unauthorized",
			httpStatus: 401,
			title:      "Authentication Failed",
		},
		{
			name:       "403 Forbidden",
			httpStatus: 403,
			title:      "Access Forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errCtx := &ErrorContext{
				ExecutionID:  "exec-123",
				WorkflowID:   "wf-456",
				NodeID:       "node-789",
				NodeType:     "action:http",
				ErrorMessage: "HTTP error",
				HTTPStatus:   tt.httpStatus,
			}

			suggestions, err := analyzer.Analyze(ctx, errCtx)
			require.NoError(t, err)
			require.NotEmpty(t, suggestions)

			s := suggestions[0]
			assert.Equal(t, ErrorCategoryAuth, s.Category)
			assert.Equal(t, SuggestionTypeCredential, s.Type)
			assert.Equal(t, tt.title, s.Title)
		})
	}
}

func TestPatternAnalyzer_Analyze_RateLimit(t *testing.T) {
	analyzer := NewPatternAnalyzer(nil)
	ctx := context.Background()

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "rate limit exceeded",
		HTTPStatus:   429,
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	s := suggestions[0]
	assert.Equal(t, ErrorCategoryRateLimit, s.Category)
	assert.Equal(t, SuggestionTypeConfigChange, s.Type)
	assert.Contains(t, s.Title, "Rate Limit")
}

func TestPatternAnalyzer_Analyze_Timeout(t *testing.T) {
	analyzer := NewPatternAnalyzer(nil)
	ctx := context.Background()

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "context deadline exceeded",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	s := suggestions[0]
	assert.Equal(t, ErrorCategoryTimeout, s.Category)
	assert.Equal(t, SuggestionTypeConfigChange, s.Type)
	assert.Contains(t, s.Title, "Timeout")
}

func TestPatternAnalyzer_Analyze_NoMatch(t *testing.T) {
	analyzer := NewPatternAnalyzer(nil)
	ctx := context.Background()

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "some random unique error xyz12345",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	assert.Empty(t, suggestions)
}

func TestPatternAnalyzer_Analyze_WithCustomPatterns(t *testing.T) {
	customPatterns := []*BuiltinPattern{
		{
			Name:                  "custom_error",
			Category:              ErrorCategoryConfig,
			MessagePatterns:       []string{"custom_error_pattern"},
			SuggestionType:        SuggestionTypeManual,
			SuggestionTitle:       "Custom Error",
			SuggestionDescription: "A custom error occurred",
			SuggestionConfidence:  ConfidenceLow,
			Priority:              50,
		},
	}

	analyzer := NewPatternAnalyzer(customPatterns)
	ctx := context.Background()

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "got custom_error_pattern in response",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	s := suggestions[0]
	assert.Equal(t, ErrorCategoryConfig, s.Category)
	assert.Equal(t, "Custom Error", s.Title)
}

func TestPatternAnalyzer_Analyze_PopulatesFields(t *testing.T) {
	analyzer := NewPatternAnalyzer(nil)
	ctx := context.Background()

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection refused",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	s := suggestions[0]
	// Check all required fields are populated
	assert.NotEmpty(t, s.ID, "ID should be generated")
	assert.Equal(t, "exec-123", s.ExecutionID)
	assert.Equal(t, "node-789", s.NodeID)
	assert.NotEmpty(t, s.Category)
	assert.NotEmpty(t, s.Type)
	assert.NotEmpty(t, s.Confidence)
	assert.NotEmpty(t, s.Title)
	assert.NotEmpty(t, s.Description)
	assert.Equal(t, SourcePattern, s.Source)
	assert.Equal(t, StatusPending, s.Status)
	assert.NotZero(t, s.CreatedAt)
}

func TestPatternAnalyzer_Analyze_WithTenantID(t *testing.T) {
	analyzer := NewPatternAnalyzerWithTenant(nil, "tenant-abc")
	ctx := context.Background()

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection refused",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	assert.Equal(t, "tenant-abc", suggestions[0].TenantID)
}

func TestPatternAnalyzer_MultipleSuggestions(t *testing.T) {
	// Create patterns that could both match
	customPatterns := []*BuiltinPattern{
		{
			Name:                  "network_generic",
			Category:              ErrorCategoryNetwork,
			MessagePatterns:       []string{"connection"},
			SuggestionType:        SuggestionTypeManual,
			SuggestionTitle:       "Generic Network Error",
			SuggestionDescription: "Network issue",
			SuggestionConfidence:  ConfidenceLow,
			Priority:              50,
		},
		{
			Name:                  "data_generic",
			Category:              ErrorCategoryData,
			MessagePatterns:       []string{"error"},
			SuggestionType:        SuggestionTypeDataFix,
			SuggestionTitle:       "Generic Data Error",
			SuggestionDescription: "Data issue",
			SuggestionConfidence:  ConfidenceLow,
			Priority:              40,
		},
	}

	analyzer := NewPatternAnalyzer(customPatterns)
	ctx := context.Background()

	// Error that matches both patterns
	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection error occurred",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)

	// Should return multiple suggestions (one per category)
	assert.Len(t, suggestions, 2)

	// Higher priority should be first
	assert.Equal(t, ErrorCategoryNetwork, suggestions[0].Category)
	assert.Equal(t, ErrorCategoryData, suggestions[1].Category)
}

func TestPatternAnalyzer_ContextCancellation(t *testing.T) {
	analyzer := NewPatternAnalyzer(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		ErrorMessage: "connection refused",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err) // Pattern matching is fast, shouldn't be affected
	require.NotEmpty(t, suggestions)
}

func TestDatabasePatternAnalyzer_Analyze(t *testing.T) {
	// Create a mock repository with test patterns
	repo := NewMockPatternRepository()
	repo.AddPattern(&ErrorPattern{
		ID:                    "pattern-1",
		Name:                  "db_pattern",
		Category:              ErrorCategoryExternal,
		Patterns:              []string{"database connection failed"},
		SuggestionType:        SuggestionTypeRetry,
		SuggestionTitle:       "Database Connection Failed",
		SuggestionDescription: "Could not connect to database",
		SuggestionConfidence:  ConfidenceHigh,
		Priority:              100,
		IsActive:              true,
	})

	analyzer := NewDatabasePatternAnalyzer(repo, "tenant-123")
	ctx := context.Background()

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		WorkflowID:   "wf-456",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "database connection failed",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	s := suggestions[0]
	assert.Equal(t, ErrorCategoryExternal, s.Category)
	assert.Equal(t, "Database Connection Failed", s.Title)
}

// MockPatternRepository for testing database patterns
type MockPatternRepository struct {
	patterns []*ErrorPattern
	err      error
}

func NewMockPatternRepository() *MockPatternRepository {
	return &MockPatternRepository{
		patterns: make([]*ErrorPattern, 0),
	}
}

func (m *MockPatternRepository) AddPattern(p *ErrorPattern) {
	m.patterns = append(m.patterns, p)
}

func (m *MockPatternRepository) SetError(err error) {
	m.err = err
}

func (m *MockPatternRepository) GetActivePatterns(ctx context.Context, tenantID string) ([]*ErrorPattern, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.patterns, nil
}

func TestDatabasePatternAnalyzer_RepositoryError(t *testing.T) {
	repo := NewMockPatternRepository()
	repo.SetError(assert.AnError)

	analyzer := NewDatabasePatternAnalyzer(repo, "tenant-123")
	ctx := context.Background()

	errCtx := &ErrorContext{
		ExecutionID:  "exec-123",
		ErrorMessage: "some error",
	}

	suggestions, err := analyzer.Analyze(ctx, errCtx)
	require.Error(t, err)
	assert.Nil(t, suggestions)
}
