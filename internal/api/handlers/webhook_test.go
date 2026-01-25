package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/webhook"
	"github.com/gorax/gorax/internal/workflow"
)

// MockWebhookWorkflowService is a mock implementation of WebhookWorkflowService for testing
type MockWebhookWorkflowService struct {
	mock.Mock
}

func (m *MockWebhookWorkflowService) Execute(ctx context.Context, tenantID, workflowID, triggerType string, triggerData []byte) (*workflow.Execution, error) {
	args := m.Called(ctx, tenantID, workflowID, triggerType, triggerData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workflow.Execution), args.Error(1)
}

// MockWebhookService is a mock implementation of WebhookService for testing
type MockWebhookService struct {
	mock.Mock
}

func (m *MockWebhookService) GetByWorkflowAndWebhookID(ctx context.Context, workflowID, webhookID string) (*webhook.Webhook, error) {
	args := m.Called(ctx, workflowID, webhookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhook.Webhook), args.Error(1)
}

func (m *MockWebhookService) VerifySignature(payload []byte, signature string, secret string) bool {
	args := m.Called(payload, signature, secret)
	return args.Bool(0)
}

func (m *MockWebhookService) LogEvent(ctx context.Context, event *webhook.WebhookEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func newTestWebhookHandler() (*WebhookHandler, *MockWebhookWorkflowService, *MockWebhookService) {
	mockWorkflowService := new(MockWebhookWorkflowService)
	mockWebhookService := new(MockWebhookService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewWebhookHandler(mockWorkflowService, mockWebhookService, logger)
	return handler, mockWorkflowService, mockWebhookService
}

func addWebhookURLParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// Test fixtures
func createTestWebhookConfig() *webhook.Webhook {
	return &webhook.Webhook{
		ID:         "webhook-123",
		TenantID:   "tenant-123",
		WorkflowID: "workflow-123",
		NodeID:     "node-123",
		Name:       "Test Webhook",
		Path:       "/webhooks/workflow-123/webhook-123",
		Secret:     "test-secret",
		AuthType:   webhook.AuthTypeNone,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func createTestWebhookConfigWithSignature() *webhook.Webhook {
	wh := createTestWebhookConfig()
	wh.AuthType = webhook.AuthTypeSignature
	wh.Secret = "secret-key"
	return wh
}

func createTestExecution() *workflow.Execution {
	return &workflow.Execution{
		ID:         "exec-123",
		TenantID:   "tenant-123",
		WorkflowID: "workflow-123",
		Status:     "pending",
		CreatedAt:  time.Now(),
	}
}

// ============================================================================
// Handle Tests
// ============================================================================

func TestWebhookHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		workflowID     string
		webhookID      string
		body           string
		headers        map[string]string
		setupMock      func(*MockWebhookWorkflowService, *MockWebhookService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "successful webhook execution - no auth",
			workflowID: "workflow-123",
			webhookID:  "webhook-123",
			body:       `{"event": "test", "data": {"foo": "bar"}}`,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				webhookConfig := createTestWebhookConfig()
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "webhook-123").
					Return(webhookConfig, nil)
				mws.On("Execute", mock.Anything, "tenant-123", "workflow-123", "webhook", mock.AnythingOfType("[]uint8")).
					Return(createTestExecution(), nil)
				mwhs.On("LogEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).
					Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, "exec-123", resp["execution_id"])
				assert.Equal(t, "pending", resp["status"])
			},
		},
		{
			name:       "successful webhook execution - with signature auth",
			workflowID: "workflow-123",
			webhookID:  "webhook-123",
			body:       `{"event": "test"}`,
			headers: map[string]string{
				"Content-Type":        "application/json",
				"X-Webhook-Signature": "sha256=valid-signature",
			},
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				webhookConfig := createTestWebhookConfigWithSignature()
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "webhook-123").
					Return(webhookConfig, nil)
				mwhs.On("VerifySignature", []byte(`{"event": "test"}`), "sha256=valid-signature", "secret-key").
					Return(true)
				mws.On("Execute", mock.Anything, "tenant-123", "workflow-123", "webhook", mock.AnythingOfType("[]uint8")).
					Return(createTestExecution(), nil)
				mwhs.On("LogEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).
					Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:       "successful webhook execution - with GitHub style signature",
			workflowID: "workflow-123",
			webhookID:  "webhook-123",
			body:       `{"event": "push"}`,
			headers: map[string]string{
				"Content-Type":          "application/json",
				"X-Hub-Signature-256":   "sha256=github-signature",
			},
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				webhookConfig := createTestWebhookConfigWithSignature()
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "webhook-123").
					Return(webhookConfig, nil)
				mwhs.On("VerifySignature", []byte(`{"event": "push"}`), "sha256=github-signature", "secret-key").
					Return(true)
				mws.On("Execute", mock.Anything, "tenant-123", "workflow-123", "webhook", mock.AnythingOfType("[]uint8")).
					Return(createTestExecution(), nil)
				mwhs.On("LogEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).
					Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:       "webhook not found",
			workflowID: "workflow-123",
			webhookID:  "nonexistent",
			body:       `{"event": "test"}`,
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "nonexistent").
					Return(nil, webhook.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "webhook not found",
		},
		{
			name:       "webhook service error",
			workflowID: "workflow-123",
			webhookID:  "webhook-123",
			body:       `{"event": "test"}`,
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "webhook-123").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to process webhook",
		},
		{
			name:       "invalid signature",
			workflowID: "workflow-123",
			webhookID:  "webhook-123",
			body:       `{"event": "test"}`,
			headers: map[string]string{
				"X-Webhook-Signature": "sha256=invalid-signature",
			},
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				webhookConfig := createTestWebhookConfigWithSignature()
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "webhook-123").
					Return(webhookConfig, nil)
				mwhs.On("VerifySignature", []byte(`{"event": "test"}`), "sha256=invalid-signature", "secret-key").
					Return(false)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid signature",
		},
		{
			name:       "missing signature when required",
			workflowID: "workflow-123",
			webhookID:  "webhook-123",
			body:       `{"event": "test"}`,
			headers:    map[string]string{},
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				webhookConfig := createTestWebhookConfigWithSignature()
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "webhook-123").
					Return(webhookConfig, nil)
				// VerifySignature will be called with empty signature
				mwhs.On("VerifySignature", []byte(`{"event": "test"}`), "", "secret-key").
					Return(false)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid signature",
		},
		{
			name:       "workflow not found",
			workflowID: "workflow-123",
			webhookID:  "webhook-123",
			body:       `{"event": "test"}`,
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				webhookConfig := createTestWebhookConfig()
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "webhook-123").
					Return(webhookConfig, nil)
				mws.On("Execute", mock.Anything, "tenant-123", "workflow-123", "webhook", mock.AnythingOfType("[]uint8")).
					Return(nil, workflow.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "workflow not found",
		},
		{
			name:       "workflow execution error",
			workflowID: "workflow-123",
			webhookID:  "webhook-123",
			body:       `{"event": "test"}`,
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				webhookConfig := createTestWebhookConfig()
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "webhook-123").
					Return(webhookConfig, nil)
				mws.On("Execute", mock.Anything, "tenant-123", "workflow-123", "webhook", mock.AnythingOfType("[]uint8")).
					Return(nil, errors.New("execution failed"))
				// LogEvent should be called for failed execution
				mwhs.On("LogEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).
					Return(nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to execute workflow",
		},
		{
			name:       "log event error does not fail request",
			workflowID: "workflow-123",
			webhookID:  "webhook-123",
			body:       `{"event": "test"}`,
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				webhookConfig := createTestWebhookConfig()
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "webhook-123").
					Return(webhookConfig, nil)
				mws.On("Execute", mock.Anything, "tenant-123", "workflow-123", "webhook", mock.AnythingOfType("[]uint8")).
					Return(createTestExecution(), nil)
				// LogEvent fails but request should still succeed
				mwhs.On("LogEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).
					Return(errors.New("logging failed"))
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:       "webhook with query parameters",
			workflowID: "workflow-123",
			webhookID:  "webhook-123",
			body:       `{"event": "test"}`,
			headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "TestClient/1.0",
			},
			setupMock: func(mws *MockWebhookWorkflowService, mwhs *MockWebhookService) {
				webhookConfig := createTestWebhookConfig()
				mwhs.On("GetByWorkflowAndWebhookID", mock.Anything, "workflow-123", "webhook-123").
					Return(webhookConfig, nil)
				mws.On("Execute", mock.Anything, "tenant-123", "workflow-123", "webhook", mock.AnythingOfType("[]uint8")).
					Return(createTestExecution(), nil)
				mwhs.On("LogEvent", mock.Anything, mock.AnythingOfType("*webhook.WebhookEvent")).
					Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockWorkflowService, mockWebhookService := newTestWebhookHandler()
			tt.setupMock(mockWorkflowService, mockWebhookService)

			req := httptest.NewRequest(http.MethodPost, "/webhooks/"+tt.workflowID+"/"+tt.webhookID, bytes.NewBufferString(tt.body))
			req = addWebhookURLParams(req, map[string]string{
				"workflowID": tt.workflowID,
				"webhookID":  tt.webhookID,
			})

			// Set headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler.Handle(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockWorkflowService.AssertExpectations(t)
			mockWebhookService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestFlattenHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		expected map[string]string
	}{
		{
			name: "multiple headers",
			headers: http.Header{
				"Content-Type":   []string{"application/json"},
				"Authorization":  []string{"Bearer token123"},
				"X-Custom-Header": []string{"value1", "value2"},
			},
			expected: map[string]string{
				"Content-Type":   "application/json",
				"Authorization":  "Bearer token123",
				"X-Custom-Header": "value1", // Only first value
			},
		},
		{
			name:     "empty headers",
			headers:  http.Header{},
			expected: map[string]string{},
		},
		{
			name: "header with empty value",
			headers: http.Header{
				"X-Empty": []string{""},
			},
			expected: map[string]string{
				"X-Empty": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenHeaders(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    map[string][]string
		expected map[string]string
	}{
		{
			name: "multiple query params",
			query: map[string][]string{
				"page":   []string{"1"},
				"limit":  []string{"10"},
				"filter": []string{"active", "pending"},
			},
			expected: map[string]string{
				"page":   "1",
				"limit":  "10",
				"filter": "active", // Only first value
			},
		},
		{
			name:     "empty query",
			query:    map[string][]string{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenQuery(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "non-empty string",
			input: "test value",
		},
		{
			name:  "empty string",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringPtr(tt.input)
			require.NotNil(t, result)
			assert.Equal(t, tt.input, *result)
		})
	}
}
