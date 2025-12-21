package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/aibuilder"
)

// MockAIBuilderService implements AIBuilderService for testing
type MockAIBuilderService struct {
	mock.Mock
}

func (m *MockAIBuilderService) Generate(ctx context.Context, tenantID, userID string, request *aibuilder.BuildRequest) (*aibuilder.BuildResult, error) {
	args := m.Called(ctx, tenantID, userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aibuilder.BuildResult), args.Error(1)
}

func (m *MockAIBuilderService) Refine(ctx context.Context, tenantID string, request *aibuilder.RefineRequest) (*aibuilder.BuildResult, error) {
	args := m.Called(ctx, tenantID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aibuilder.BuildResult), args.Error(1)
}

func (m *MockAIBuilderService) GetConversation(ctx context.Context, tenantID, conversationID string) (*aibuilder.Conversation, error) {
	args := m.Called(ctx, tenantID, conversationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aibuilder.Conversation), args.Error(1)
}

func (m *MockAIBuilderService) ListConversations(ctx context.Context, tenantID, userID string) ([]*aibuilder.Conversation, error) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aibuilder.Conversation), args.Error(1)
}

func (m *MockAIBuilderService) Apply(ctx context.Context, tenantID, userID string, request *aibuilder.ApplyRequest) (string, error) {
	args := m.Called(ctx, tenantID, userID, request)
	return args.String(0), args.Error(1)
}

func (m *MockAIBuilderService) AbandonConversation(ctx context.Context, tenantID, conversationID string) error {
	args := m.Called(ctx, tenantID, conversationID)
	return args.Error(0)
}

func setupAIBuilderHandler(svc *MockAIBuilderService) (*AIBuilderHandler, *chi.Mux) {
	handler := NewAIBuilderHandler(svc)
	router := chi.NewRouter()

	// Add tenant context middleware for tests
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), tenantIDKey, "tenant-123")
			ctx = context.WithValue(ctx, userIDKey, "user-456")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	router.Route("/api/v1/ai/workflows", func(r chi.Router) {
		r.Post("/generate", handler.Generate)
		r.Post("/refine", handler.Refine)
		r.Get("/conversations", handler.ListConversations)
		r.Get("/conversations/{id}", handler.GetConversation)
		r.Post("/conversations/{id}/apply", handler.Apply)
		r.Post("/conversations/{id}/abandon", handler.Abandon)
	})

	return handler, router
}

func TestAIBuilderHandler_Generate(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		result := &aibuilder.BuildResult{
			ConversationID: "conv-123",
			Workflow: &aibuilder.GeneratedWorkflow{
				Name: "Test Workflow",
				Definition: &aibuilder.WorkflowDefinition{
					Nodes: []aibuilder.GeneratedNode{
						{ID: "n1", Type: "trigger:webhook", Name: "Trigger"},
					},
				},
			},
			Explanation: "Here's your workflow",
		}

		svc.On("Generate", mock.Anything, "tenant-123", "user-456", mock.MatchedBy(func(req *aibuilder.BuildRequest) bool {
			return req.Description == "Create a webhook workflow"
		})).Return(result, nil)

		body := `{"description": "Create a webhook workflow"}`
		req := httptest.NewRequest("POST", "/api/v1/ai/workflows/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "conv-123", resp["conversation_id"])
		assert.NotNil(t, resp["workflow"])

		svc.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		req := httptest.NewRequest("POST", "/api/v1/ai/workflows/generate", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		svc.On("Generate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("LLM unavailable"))

		body := `{"description": "Create a webhook workflow"}`
		req := httptest.NewRequest("POST", "/api/v1/ai/workflows/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		svc.AssertExpectations(t)
	})
}

func TestAIBuilderHandler_Refine(t *testing.T) {
	t.Run("successful refinement", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		result := &aibuilder.BuildResult{
			ConversationID: "conv-123",
			Workflow: &aibuilder.GeneratedWorkflow{
				Name: "Refined Workflow",
			},
			Explanation: "Updated workflow",
		}

		svc.On("Refine", mock.Anything, "tenant-123", mock.MatchedBy(func(req *aibuilder.RefineRequest) bool {
			return req.ConversationID == "conv-123" && req.Message == "Add error handling"
		})).Return(result, nil)

		body := `{"conversation_id": "conv-123", "message": "Add error handling"}`
		req := httptest.NewRequest("POST", "/api/v1/ai/workflows/refine", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		svc.AssertExpectations(t)
	})

	t.Run("conversation not found", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		svc.On("Refine", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("conversation not found"))

		body := `{"conversation_id": "conv-123", "message": "Add error handling"}`
		req := httptest.NewRequest("POST", "/api/v1/ai/workflows/refine", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		svc.AssertExpectations(t)
	})
}

func TestAIBuilderHandler_GetConversation(t *testing.T) {
	t.Run("existing conversation", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		conv := &aibuilder.Conversation{
			ID:       "conv-123",
			TenantID: "tenant-123",
			UserID:   "user-456",
			Status:   aibuilder.ConversationStatusActive,
		}

		svc.On("GetConversation", mock.Anything, "tenant-123", "conv-123").Return(conv, nil)

		req := httptest.NewRequest("GET", "/api/v1/ai/workflows/conversations/conv-123", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "conv-123", resp["id"])

		svc.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		svc.On("GetConversation", mock.Anything, "tenant-123", "conv-999").
			Return(nil, errors.New("not found"))

		req := httptest.NewRequest("GET", "/api/v1/ai/workflows/conversations/conv-999", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		svc.AssertExpectations(t)
	})
}

func TestAIBuilderHandler_ListConversations(t *testing.T) {
	svc := &MockAIBuilderService{}
	_, router := setupAIBuilderHandler(svc)

	convs := []*aibuilder.Conversation{
		{ID: "conv-1", TenantID: "tenant-123", UserID: "user-456"},
		{ID: "conv-2", TenantID: "tenant-123", UserID: "user-456"},
	}

	svc.On("ListConversations", mock.Anything, "tenant-123", "user-456").Return(convs, nil)

	req := httptest.NewRequest("GET", "/api/v1/ai/workflows/conversations", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	data := resp["data"].([]interface{})
	assert.Len(t, data, 2)

	svc.AssertExpectations(t)
}

func TestAIBuilderHandler_Apply(t *testing.T) {
	t.Run("successful apply", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		svc.On("Apply", mock.Anything, "tenant-123", "user-456", mock.MatchedBy(func(req *aibuilder.ApplyRequest) bool {
			return req.ConversationID == "conv-123"
		})).Return("workflow-789", nil)

		req := httptest.NewRequest("POST", "/api/v1/ai/workflows/conversations/conv-123/apply", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.Equal(t, "workflow-789", resp["workflow_id"])

		svc.AssertExpectations(t)
	})

	t.Run("no workflow to apply", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		svc.On("Apply", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return("", errors.New("no workflow to apply"))

		req := httptest.NewRequest("POST", "/api/v1/ai/workflows/conversations/conv-123/apply", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		svc.AssertExpectations(t)
	})
}

func TestAIBuilderHandler_Abandon(t *testing.T) {
	t.Run("successful abandon", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		svc.On("AbandonConversation", mock.Anything, "tenant-123", "conv-123").Return(nil)

		req := httptest.NewRequest("POST", "/api/v1/ai/workflows/conversations/conv-123/abandon", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)

		svc.AssertExpectations(t)
	})

	t.Run("conversation not active", func(t *testing.T) {
		svc := &MockAIBuilderService{}
		_, router := setupAIBuilderHandler(svc)

		svc.On("AbandonConversation", mock.Anything, "tenant-123", "conv-123").
			Return(errors.New("conversation is not active"))

		req := httptest.NewRequest("POST", "/api/v1/ai/workflows/conversations/conv-123/abandon", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		svc.AssertExpectations(t)
	})
}

func TestNewAIBuilderHandler(t *testing.T) {
	svc := &MockAIBuilderService{}
	handler := NewAIBuilderHandler(svc)

	require.NotNil(t, handler)
	assert.Equal(t, svc, handler.service)
}
