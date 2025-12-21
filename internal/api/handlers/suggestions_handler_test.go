package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gorax/gorax/internal/suggestions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSuggestionService implements suggestions.Service for testing
type MockSuggestionService struct {
	suggestions []*suggestions.Suggestion
	err         error
}

func NewMockSuggestionService() *MockSuggestionService {
	return &MockSuggestionService{
		suggestions: make([]*suggestions.Suggestion, 0),
	}
}

func (m *MockSuggestionService) SetSuggestions(s []*suggestions.Suggestion) {
	m.suggestions = s
}

func (m *MockSuggestionService) SetError(err error) {
	m.err = err
}

func (m *MockSuggestionService) AnalyzeError(ctx context.Context, tenantID string, errCtx *suggestions.ErrorContext) ([]*suggestions.Suggestion, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.suggestions, nil
}

func (m *MockSuggestionService) GetSuggestions(ctx context.Context, tenantID, executionID string) ([]*suggestions.Suggestion, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []*suggestions.Suggestion
	for _, s := range m.suggestions {
		if s.TenantID == tenantID && s.ExecutionID == executionID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockSuggestionService) GetSuggestionByID(ctx context.Context, tenantID, suggestionID string) (*suggestions.Suggestion, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, s := range m.suggestions {
		if s.TenantID == tenantID && s.ID == suggestionID {
			return s, nil
		}
	}
	return nil, suggestions.ErrSuggestionNotFound
}

func (m *MockSuggestionService) ApplySuggestion(ctx context.Context, tenantID, suggestionID string) error {
	if m.err != nil {
		return m.err
	}
	for _, s := range m.suggestions {
		if s.TenantID == tenantID && s.ID == suggestionID {
			s.Status = suggestions.StatusApplied
			return nil
		}
	}
	return suggestions.ErrSuggestionNotFound
}

func (m *MockSuggestionService) DismissSuggestion(ctx context.Context, tenantID, suggestionID string) error {
	if m.err != nil {
		return m.err
	}
	for _, s := range m.suggestions {
		if s.TenantID == tenantID && s.ID == suggestionID {
			s.Status = suggestions.StatusDismissed
			return nil
		}
	}
	return suggestions.ErrSuggestionNotFound
}

// Helper to create test context with tenant
func createTestRequestWithTenant(method, path string, body interface{}) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-123")
	return req
}

func TestSuggestionsHandler_List(t *testing.T) {
	service := NewMockSuggestionService()
	service.SetSuggestions([]*suggestions.Suggestion{
		suggestions.NewSuggestion("tenant-123", "exec-456", "node-1", suggestions.ErrorCategoryNetwork, suggestions.SuggestionTypeRetry, suggestions.ConfidenceHigh, "S1", "D1", suggestions.SourcePattern),
		suggestions.NewSuggestion("tenant-123", "exec-456", "node-2", suggestions.ErrorCategoryAuth, suggestions.SuggestionTypeCredential, suggestions.ConfidenceMedium, "S2", "D2", suggestions.SourceLLM),
	})

	handler := NewSuggestionsHandler(service, nil)

	r := chi.NewRouter()
	r.Get("/executions/{executionID}/suggestions", withTenantMiddleware(handler.List))

	req := createTestRequestWithTenant("GET", "/executions/exec-456/suggestions", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
}

func TestSuggestionsHandler_Get(t *testing.T) {
	service := NewMockSuggestionService()
	s := suggestions.NewSuggestion("tenant-123", "exec-456", "node-789", suggestions.ErrorCategoryNetwork, suggestions.SuggestionTypeRetry, suggestions.ConfidenceHigh, "Test", "Test desc", suggestions.SourcePattern)
	service.SetSuggestions([]*suggestions.Suggestion{s})

	handler := NewSuggestionsHandler(service, nil)

	r := chi.NewRouter()
	r.Get("/suggestions/{suggestionID}", withTenantMiddleware(handler.Get))

	req := createTestRequestWithTenant("GET", "/suggestions/"+s.ID, nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, s.ID, data["id"])
}

func TestSuggestionsHandler_Get_NotFound(t *testing.T) {
	service := NewMockSuggestionService()
	handler := NewSuggestionsHandler(service, nil)

	r := chi.NewRouter()
	r.Get("/suggestions/{suggestionID}", withTenantMiddleware(handler.Get))

	req := createTestRequestWithTenant("GET", "/suggestions/nonexistent", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSuggestionsHandler_Analyze(t *testing.T) {
	service := NewMockSuggestionService()
	service.SetSuggestions([]*suggestions.Suggestion{
		suggestions.NewSuggestion("tenant-123", "exec-456", "node-789", suggestions.ErrorCategoryNetwork, suggestions.SuggestionTypeRetry, suggestions.ConfidenceHigh, "Connection Error", "Connection refused", suggestions.SourcePattern),
	})

	handler := NewSuggestionsHandler(service, nil)

	r := chi.NewRouter()
	r.Post("/executions/{executionID}/analyze", withTenantMiddleware(handler.Analyze))

	body := AnalyzeRequest{
		WorkflowID:   "wf-123",
		NodeID:       "node-789",
		NodeType:     "action:http",
		ErrorMessage: "connection refused",
	}

	req := createTestRequestWithTenant("POST", "/executions/exec-456/analyze", body)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, data)
}

func TestSuggestionsHandler_Analyze_InvalidBody(t *testing.T) {
	service := NewMockSuggestionService()
	handler := NewSuggestionsHandler(service, nil)

	r := chi.NewRouter()
	r.Post("/executions/{executionID}/analyze", withTenantMiddleware(handler.Analyze))

	req := httptest.NewRequest("POST", "/executions/exec-456/analyze", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-123")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSuggestionsHandler_Apply(t *testing.T) {
	service := NewMockSuggestionService()
	s := suggestions.NewSuggestion("tenant-123", "exec-456", "node-789", suggestions.ErrorCategoryNetwork, suggestions.SuggestionTypeRetry, suggestions.ConfidenceHigh, "Test", "Test", suggestions.SourcePattern)
	service.SetSuggestions([]*suggestions.Suggestion{s})

	handler := NewSuggestionsHandler(service, nil)

	r := chi.NewRouter()
	r.Post("/suggestions/{suggestionID}/apply", withTenantMiddleware(handler.Apply))

	req := createTestRequestWithTenant("POST", "/suggestions/"+s.ID+"/apply", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify suggestion was marked as applied
	assert.Equal(t, suggestions.StatusApplied, s.Status)
}

func TestSuggestionsHandler_Apply_NotFound(t *testing.T) {
	service := NewMockSuggestionService()
	handler := NewSuggestionsHandler(service, nil)

	r := chi.NewRouter()
	r.Post("/suggestions/{suggestionID}/apply", withTenantMiddleware(handler.Apply))

	req := createTestRequestWithTenant("POST", "/suggestions/nonexistent/apply", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestSuggestionsHandler_Dismiss(t *testing.T) {
	service := NewMockSuggestionService()
	s := suggestions.NewSuggestion("tenant-123", "exec-456", "node-789", suggestions.ErrorCategoryNetwork, suggestions.SuggestionTypeRetry, suggestions.ConfidenceHigh, "Test", "Test", suggestions.SourcePattern)
	service.SetSuggestions([]*suggestions.Suggestion{s})

	handler := NewSuggestionsHandler(service, nil)

	r := chi.NewRouter()
	r.Post("/suggestions/{suggestionID}/dismiss", withTenantMiddleware(handler.Dismiss))

	req := createTestRequestWithTenant("POST", "/suggestions/"+s.ID+"/dismiss", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify suggestion was marked as dismissed
	assert.Equal(t, suggestions.StatusDismissed, s.Status)
}

func TestSuggestionsHandler_Dismiss_NotFound(t *testing.T) {
	service := NewMockSuggestionService()
	handler := NewSuggestionsHandler(service, nil)

	r := chi.NewRouter()
	r.Post("/suggestions/{suggestionID}/dismiss", withTenantMiddleware(handler.Dismiss))

	req := createTestRequestWithTenant("POST", "/suggestions/nonexistent/dismiss", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// Simple middleware to inject tenant ID for testing
func withTenantMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		ctx := context.WithValue(r.Context(), suggestionsTenantIDKey, tenantID)
		next(w, r.WithContext(ctx))
	}
}
