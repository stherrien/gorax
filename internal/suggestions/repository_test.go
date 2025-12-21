package suggestions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository implements Repository for testing
type MockRepository struct {
	suggestions map[string]*Suggestion
	err         error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		suggestions: make(map[string]*Suggestion),
	}
}

func (m *MockRepository) SetError(err error) {
	m.err = err
}

func (m *MockRepository) Create(ctx context.Context, suggestion *Suggestion) error {
	if m.err != nil {
		return m.err
	}
	m.suggestions[suggestion.ID] = suggestion
	return nil
}

func (m *MockRepository) CreateBatch(ctx context.Context, suggestions []*Suggestion) error {
	if m.err != nil {
		return m.err
	}
	for _, s := range suggestions {
		m.suggestions[s.ID] = s
	}
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, tenantID, id string) (*Suggestion, error) {
	if m.err != nil {
		return nil, m.err
	}
	s, ok := m.suggestions[id]
	if !ok || s.TenantID != tenantID {
		return nil, ErrSuggestionNotFound
	}
	return s, nil
}

func (m *MockRepository) GetByExecutionID(ctx context.Context, tenantID, executionID string) ([]*Suggestion, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []*Suggestion
	for _, s := range m.suggestions {
		if s.TenantID == tenantID && s.ExecutionID == executionID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockRepository) UpdateStatus(ctx context.Context, tenantID, id string, status SuggestionStatus) error {
	if m.err != nil {
		return m.err
	}
	s, ok := m.suggestions[id]
	if !ok || s.TenantID != tenantID {
		return ErrSuggestionNotFound
	}
	s.Status = status
	now := time.Now()
	if status == StatusApplied {
		s.AppliedAt = &now
	} else if status == StatusDismissed {
		s.DismissedAt = &now
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, tenantID, id string) error {
	if m.err != nil {
		return m.err
	}
	s, ok := m.suggestions[id]
	if !ok || s.TenantID != tenantID {
		return ErrSuggestionNotFound
	}
	delete(m.suggestions, id)
	return nil
}

func (m *MockRepository) DeleteByExecutionID(ctx context.Context, tenantID, executionID string) error {
	if m.err != nil {
		return m.err
	}
	for id, s := range m.suggestions {
		if s.TenantID == tenantID && s.ExecutionID == executionID {
			delete(m.suggestions, id)
		}
	}
	return nil
}

func TestMockRepository_Create(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	suggestion := NewSuggestion(
		"tenant-123",
		"exec-456",
		"node-789",
		ErrorCategoryNetwork,
		SuggestionTypeRetry,
		ConfidenceHigh,
		"Test Suggestion",
		"Test Description",
		SourcePattern,
	)

	err := repo.Create(ctx, suggestion)
	require.NoError(t, err)

	// Verify it was stored
	retrieved, err := repo.GetByID(ctx, "tenant-123", suggestion.ID)
	require.NoError(t, err)
	assert.Equal(t, suggestion.ID, retrieved.ID)
	assert.Equal(t, suggestion.Title, retrieved.Title)
}

func TestMockRepository_GetByID_NotFound(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "tenant-123", "nonexistent")
	assert.ErrorIs(t, err, ErrSuggestionNotFound)
}

func TestMockRepository_GetByID_WrongTenant(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	suggestion := NewSuggestion(
		"tenant-123",
		"exec-456",
		"node-789",
		ErrorCategoryNetwork,
		SuggestionTypeRetry,
		ConfidenceHigh,
		"Test",
		"Test",
		SourcePattern,
	)
	err := repo.Create(ctx, suggestion)
	require.NoError(t, err)

	// Try to get with wrong tenant
	_, err = repo.GetByID(ctx, "wrong-tenant", suggestion.ID)
	assert.ErrorIs(t, err, ErrSuggestionNotFound)
}

func TestMockRepository_GetByExecutionID(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Create multiple suggestions
	s1 := NewSuggestion("tenant-123", "exec-456", "node-1", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "S1", "D1", SourcePattern)
	s2 := NewSuggestion("tenant-123", "exec-456", "node-2", ErrorCategoryAuth, SuggestionTypeCredential, ConfidenceMedium, "S2", "D2", SourceLLM)
	s3 := NewSuggestion("tenant-123", "exec-other", "node-3", ErrorCategoryData, SuggestionTypeDataFix, ConfidenceLow, "S3", "D3", SourcePattern)

	require.NoError(t, repo.Create(ctx, s1))
	require.NoError(t, repo.Create(ctx, s2))
	require.NoError(t, repo.Create(ctx, s3))

	// Get suggestions for exec-456
	suggestions, err := repo.GetByExecutionID(ctx, "tenant-123", "exec-456")
	require.NoError(t, err)
	assert.Len(t, suggestions, 2)
}

func TestMockRepository_CreateBatch(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	suggestions := []*Suggestion{
		NewSuggestion("tenant-123", "exec-456", "node-1", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "S1", "D1", SourcePattern),
		NewSuggestion("tenant-123", "exec-456", "node-2", ErrorCategoryAuth, SuggestionTypeCredential, ConfidenceMedium, "S2", "D2", SourceLLM),
	}

	err := repo.CreateBatch(ctx, suggestions)
	require.NoError(t, err)

	// Verify both were stored
	retrieved, err := repo.GetByExecutionID(ctx, "tenant-123", "exec-456")
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestMockRepository_UpdateStatus(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	suggestion := NewSuggestion("tenant-123", "exec-456", "node-789", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "Test", "Test", SourcePattern)
	require.NoError(t, repo.Create(ctx, suggestion))

	// Update status to applied
	err := repo.UpdateStatus(ctx, "tenant-123", suggestion.ID, StatusApplied)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetByID(ctx, "tenant-123", suggestion.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusApplied, retrieved.Status)
	assert.NotNil(t, retrieved.AppliedAt)
}

func TestMockRepository_UpdateStatus_Dismissed(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	suggestion := NewSuggestion("tenant-123", "exec-456", "node-789", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "Test", "Test", SourcePattern)
	require.NoError(t, repo.Create(ctx, suggestion))

	// Update status to dismissed
	err := repo.UpdateStatus(ctx, "tenant-123", suggestion.ID, StatusDismissed)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetByID(ctx, "tenant-123", suggestion.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusDismissed, retrieved.Status)
	assert.NotNil(t, retrieved.DismissedAt)
}

func TestMockRepository_Delete(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	suggestion := NewSuggestion("tenant-123", "exec-456", "node-789", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "Test", "Test", SourcePattern)
	require.NoError(t, repo.Create(ctx, suggestion))

	// Delete
	err := repo.Delete(ctx, "tenant-123", suggestion.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = repo.GetByID(ctx, "tenant-123", suggestion.ID)
	assert.ErrorIs(t, err, ErrSuggestionNotFound)
}

func TestMockRepository_DeleteByExecutionID(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Create suggestions for different executions
	s1 := NewSuggestion("tenant-123", "exec-456", "node-1", ErrorCategoryNetwork, SuggestionTypeRetry, ConfidenceHigh, "S1", "D1", SourcePattern)
	s2 := NewSuggestion("tenant-123", "exec-456", "node-2", ErrorCategoryAuth, SuggestionTypeCredential, ConfidenceMedium, "S2", "D2", SourceLLM)
	s3 := NewSuggestion("tenant-123", "exec-other", "node-3", ErrorCategoryData, SuggestionTypeDataFix, ConfidenceLow, "S3", "D3", SourcePattern)

	require.NoError(t, repo.Create(ctx, s1))
	require.NoError(t, repo.Create(ctx, s2))
	require.NoError(t, repo.Create(ctx, s3))

	// Delete by execution ID
	err := repo.DeleteByExecutionID(ctx, "tenant-123", "exec-456")
	require.NoError(t, err)

	// Verify exec-456 suggestions are gone
	suggestions, err := repo.GetByExecutionID(ctx, "tenant-123", "exec-456")
	require.NoError(t, err)
	assert.Len(t, suggestions, 0)

	// Verify other execution's suggestion is still there
	retrieved, err := repo.GetByID(ctx, "tenant-123", s3.ID)
	require.NoError(t, err)
	assert.Equal(t, s3.ID, retrieved.ID)
}
