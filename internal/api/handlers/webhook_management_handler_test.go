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
)

// MockWebhookManagementService is a mock implementation of webhook service for testing
type MockWebhookManagementService struct {
	mock.Mock
}

func (m *MockWebhookManagementService) List(ctx context.Context, tenantID string, limit, offset int) ([]*webhook.Webhook, int, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*webhook.Webhook), args.Int(1), args.Error(2)
}

func (m *MockWebhookManagementService) GetByID(ctx context.Context, tenantID, webhookID string) (*webhook.Webhook, error) {
	args := m.Called(ctx, tenantID, webhookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhook.Webhook), args.Error(1)
}

func (m *MockWebhookManagementService) CreateWithDetails(ctx context.Context, tenantID, workflowID, name, path, authType, description string, priority int) (*webhook.Webhook, error) {
	args := m.Called(ctx, tenantID, workflowID, name, path, authType, description, priority)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhook.Webhook), args.Error(1)
}

func (m *MockWebhookManagementService) Update(ctx context.Context, tenantID, webhookID, name, authType, description string, priority int, enabled bool) (*webhook.Webhook, error) {
	args := m.Called(ctx, tenantID, webhookID, name, authType, description, priority, enabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhook.Webhook), args.Error(1)
}

func (m *MockWebhookManagementService) DeleteByID(ctx context.Context, tenantID, webhookID string) error {
	args := m.Called(ctx, tenantID, webhookID)
	return args.Error(0)
}

func (m *MockWebhookManagementService) RegenerateSecret(ctx context.Context, tenantID, webhookID string) (*webhook.Webhook, error) {
	args := m.Called(ctx, tenantID, webhookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhook.Webhook), args.Error(1)
}

func (m *MockWebhookManagementService) TestWebhook(ctx context.Context, tenantID, webhookID, method string, headers map[string]string, body json.RawMessage) (*webhook.TestResult, error) {
	args := m.Called(ctx, tenantID, webhookID, method, headers, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhook.TestResult), args.Error(1)
}

func (m *MockWebhookManagementService) GetEventHistory(ctx context.Context, tenantID, webhookID string, limit, offset int) ([]*webhook.Event, int, error) {
	args := m.Called(ctx, tenantID, webhookID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*webhook.Event), args.Int(1), args.Error(2)
}

func newTestWebhookManagementHandler() (*WebhookManagementHandler, *MockWebhookManagementService) {
	mockService := new(MockWebhookManagementService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewWebhookManagementHandler(mockService, logger)
	return handler, mockService
}

func addRouteParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestListWebhooks(t *testing.T) {
	now := time.Now()
	webhooks := []*webhook.Webhook{
		{
			ID:           "webhook-1",
			TenantID:     "tenant-123",
			WorkflowID:   "workflow-1",
			Name:         "Test Webhook",
			Path:         "/webhooks/workflow-1/webhook-1",
			AuthType:     "signature",
			Enabled:      true,
			Priority:     1,
			TriggerCount: 5,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	tests := []struct {
		name           string
		tenantID       string
		queryParams    string
		expectedLimit  int
		expectedOffset int
		mockReturn     []*webhook.Webhook
		mockTotal      int
		mockError      error
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "success with default pagination",
			tenantID:       "tenant-123",
			queryParams:    "",
			expectedLimit:  20,
			expectedOffset: 0,
			mockReturn:     webhooks,
			mockTotal:      1,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, float64(1), body["total"])
				assert.Equal(t, float64(20), body["limit"])
				assert.Equal(t, float64(0), body["offset"])
				assert.NotNil(t, body["data"])
			},
		},
		{
			name:           "success with custom pagination",
			tenantID:       "tenant-123",
			queryParams:    "?limit=10&offset=5",
			expectedLimit:  10,
			expectedOffset: 5,
			mockReturn:     webhooks,
			mockTotal:      15,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, float64(15), body["total"])
				assert.Equal(t, float64(10), body["limit"])
				assert.Equal(t, float64(5), body["offset"])
			},
		},
		{
			name:           "service error",
			tenantID:       "tenant-123",
			queryParams:    "",
			expectedLimit:  20,
			expectedOffset: 0,
			mockReturn:     nil,
			mockTotal:      0,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "failed to list webhooks")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookManagementHandler()

			mockService.On("List",
				mock.Anything,
				tt.tenantID,
				tt.expectedLimit,
				tt.expectedOffset,
			).Return(tt.mockReturn, tt.mockTotal, tt.mockError)

			req := httptest.NewRequest("GET", "/api/v1/webhooks"+tt.queryParams, nil)
			req = addTenantContext(req, tt.tenantID)
			w := httptest.NewRecorder()

			handler.List(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetWebhook(t *testing.T) {
	now := time.Now()
	testWebhook := &webhook.Webhook{
		ID:         "webhook-1",
		TenantID:   "tenant-123",
		WorkflowID: "workflow-1",
		Name:       "Test Webhook",
		Path:       "/webhooks/workflow-1/webhook-1",
		AuthType:   "signature",
		Enabled:    true,
		Priority:   1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		mockReturn     *webhook.Webhook
		mockError      error
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "success",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			mockReturn:     testWebhook,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].(map[string]interface{})
				assert.Equal(t, "webhook-1", data["id"])
				assert.Equal(t, "Test Webhook", data["name"])
			},
		},
		{
			name:           "not found",
			tenantID:       "tenant-123",
			webhookID:      "webhook-999",
			mockReturn:     nil,
			mockError:      webhook.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "webhook not found")
			},
		},
		{
			name:           "service error",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "failed to get webhook")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookManagementHandler()

			mockService.On("GetByID",
				mock.Anything,
				tt.tenantID,
				tt.webhookID,
			).Return(tt.mockReturn, tt.mockError)

			req := httptest.NewRequest("GET", "/api/v1/webhooks/"+tt.webhookID, nil)
			req = addTenantContext(req, tt.tenantID)
			req = addRouteParam(req, "id", tt.webhookID)
			w := httptest.NewRecorder()

			handler.Get(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestCreateWebhook(t *testing.T) {
	now := time.Now()
	expectedWebhook := &webhook.Webhook{
		ID:         "webhook-1",
		TenantID:   "tenant-123",
		WorkflowID: "workflow-1",
		Name:       "New Webhook",
		Path:       "/webhooks/workflow-1/webhook-1",
		AuthType:   "signature",
		Enabled:    true,
		Priority:   1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tests := []struct {
		name           string
		tenantID       string
		requestBody    interface{}
		setupMock      func(mockService *MockWebhookManagementService)
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:     "success",
			tenantID: "tenant-123",
			requestBody: map[string]interface{}{
				"name":        "New Webhook",
				"workflowId":  "workflow-1",
				"path":        "/custom/path",
				"authType":    "signature",
				"description": "Test webhook",
				"priority":    1,
			},
			setupMock: func(mockService *MockWebhookManagementService) {
				mockService.On("CreateWithDetails",
					mock.Anything,
					"tenant-123",
					"workflow-1",
					"New Webhook",
					"/custom/path",
					"signature",
					"Test webhook",
					1,
				).Return(expectedWebhook, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].(map[string]interface{})
				assert.Equal(t, "webhook-1", data["id"])
				assert.Equal(t, "New Webhook", data["name"])
			},
		},
		{
			name:     "missing required fields",
			tenantID: "tenant-123",
			requestBody: map[string]interface{}{
				"name": "",
			},
			setupMock:      func(mockService *MockWebhookManagementService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "error")
			},
		},
		{
			name:     "invalid auth type",
			tenantID: "tenant-123",
			requestBody: map[string]interface{}{
				"name":       "Test",
				"workflowId": "workflow-1",
				"path":       "/test",
				"authType":   "invalid_auth",
				"priority":   1,
			},
			setupMock:      func(mockService *MockWebhookManagementService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "error")
			},
		},
		{
			name:     "invalid priority",
			tenantID: "tenant-123",
			requestBody: map[string]interface{}{
				"name":       "Test",
				"workflowId": "workflow-1",
				"path":       "/test",
				"authType":   "signature",
				"priority":   10,
			},
			setupMock:      func(mockService *MockWebhookManagementService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "error")
			},
		},
		{
			name:           "invalid json body",
			tenantID:       "tenant-123",
			requestBody:    "invalid json",
			setupMock:      func(mockService *MockWebhookManagementService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "invalid request body")
			},
		},
		{
			name:     "service error",
			tenantID: "tenant-123",
			requestBody: map[string]interface{}{
				"name":        "New Webhook",
				"workflowId":  "workflow-1",
				"path":        "/custom/path",
				"authType":    "signature",
				"description": "Test webhook",
				"priority":    1,
			},
			setupMock: func(mockService *MockWebhookManagementService) {
				mockService.On("CreateWithDetails",
					mock.Anything,
					"tenant-123",
					"workflow-1",
					"New Webhook",
					"/custom/path",
					"signature",
					"Test webhook",
					1,
				).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "failed to create webhook")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookManagementHandler()
			tt.setupMock(mockService)

			var bodyBytes []byte
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/api/v1/webhooks", bytes.NewReader(bodyBytes))
			req = addTenantContext(req, tt.tenantID)
			w := httptest.NewRecorder()

			handler.Create(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateWebhook(t *testing.T) {
	now := time.Now()
	updatedWebhook := &webhook.Webhook{
		ID:         "webhook-1",
		TenantID:   "tenant-123",
		WorkflowID: "workflow-1",
		Name:       "Updated Webhook",
		Path:       "/webhooks/workflow-1/webhook-1",
		AuthType:   "api_key",
		Enabled:    false,
		Priority:   2,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		requestBody    interface{}
		setupMock      func(mockService *MockWebhookManagementService)
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:      "success",
			tenantID:  "tenant-123",
			webhookID: "webhook-1",
			requestBody: map[string]interface{}{
				"name":        "Updated Webhook",
				"authType":    "api_key",
				"description": "Updated description",
				"priority":    2,
				"enabled":     false,
			},
			setupMock: func(mockService *MockWebhookManagementService) {
				mockService.On("Update",
					mock.Anything,
					"tenant-123",
					"webhook-1",
					"Updated Webhook",
					"api_key",
					"Updated description",
					2,
					false,
				).Return(updatedWebhook, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].(map[string]interface{})
				assert.Equal(t, "Updated Webhook", data["name"])
				assert.Equal(t, "api_key", data["auth_type"])
			},
		},
		{
			name:      "not found",
			tenantID:  "tenant-123",
			webhookID: "webhook-999",
			requestBody: map[string]interface{}{
				"name":     "Updated Webhook",
				"priority": 1,
			},
			setupMock: func(mockService *MockWebhookManagementService) {
				mockService.On("Update",
					mock.Anything,
					"tenant-123",
					"webhook-999",
					"Updated Webhook",
					"",
					"",
					1,
					false,
				).Return(nil, webhook.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "webhook not found")
			},
		},
		{
			name:           "invalid json body",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			requestBody:    "invalid json",
			setupMock:      func(mockService *MockWebhookManagementService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "invalid request body")
			},
		},
		{
			name:      "invalid auth type",
			tenantID:  "tenant-123",
			webhookID: "webhook-1",
			requestBody: map[string]interface{}{
				"authType": "invalid_auth",
				"priority": 1,
			},
			setupMock:      func(mockService *MockWebhookManagementService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "error")
			},
		},
		{
			name:      "service error",
			tenantID:  "tenant-123",
			webhookID: "webhook-1",
			requestBody: map[string]interface{}{
				"name":     "Updated",
				"priority": 1,
			},
			setupMock: func(mockService *MockWebhookManagementService) {
				mockService.On("Update",
					mock.Anything,
					"tenant-123",
					"webhook-1",
					"Updated",
					"",
					"",
					1,
					false,
				).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "failed to update webhook")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookManagementHandler()
			tt.setupMock(mockService)

			var bodyBytes []byte
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest("PUT", "/api/v1/webhooks/"+tt.webhookID, bytes.NewReader(bodyBytes))
			req = addTenantContext(req, tt.tenantID)
			req = addRouteParam(req, "id", tt.webhookID)
			w := httptest.NewRecorder()

			handler.Update(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteWebhook(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		mockError      error
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "success",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
			checkResponse: func(t *testing.T, body []byte) {
				assert.Empty(t, body)
			},
		},
		{
			name:           "not found",
			tenantID:       "tenant-123",
			webhookID:      "webhook-999",
			mockError:      webhook.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				json.Unmarshal(body, &response)
				assert.Contains(t, response["error"], "webhook not found")
			},
		},
		{
			name:           "service error",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				json.Unmarshal(body, &response)
				assert.Contains(t, response["error"], "failed to delete webhook")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookManagementHandler()

			mockService.On("DeleteByID",
				mock.Anything,
				tt.tenantID,
				tt.webhookID,
			).Return(tt.mockError)

			req := httptest.NewRequest("DELETE", "/api/v1/webhooks/"+tt.webhookID, nil)
			req = addTenantContext(req, tt.tenantID)
			req = addRouteParam(req, "id", tt.webhookID)
			w := httptest.NewRecorder()

			handler.Delete(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestRegenerateSecret(t *testing.T) {
	now := time.Now()
	updatedWebhook := &webhook.Webhook{
		ID:         "webhook-1",
		TenantID:   "tenant-123",
		WorkflowID: "workflow-1",
		Name:       "Test Webhook",
		Path:       "/webhooks/workflow-1/webhook-1",
		Secret:     "new-secret-value",
		AuthType:   "signature",
		Enabled:    true,
		Priority:   1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		mockReturn     *webhook.Webhook
		mockError      error
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "success",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			mockReturn:     updatedWebhook,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].(map[string]interface{})
				assert.Equal(t, "new-secret-value", data["secret"])
				assert.Equal(t, "webhook-1", data["id"])
			},
		},
		{
			name:           "not found",
			tenantID:       "tenant-123",
			webhookID:      "webhook-999",
			mockReturn:     nil,
			mockError:      webhook.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "webhook not found")
			},
		},
		{
			name:           "service error",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "failed to regenerate secret")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookManagementHandler()

			mockService.On("RegenerateSecret",
				mock.Anything,
				tt.tenantID,
				tt.webhookID,
			).Return(tt.mockReturn, tt.mockError)

			req := httptest.NewRequest("POST", "/api/v1/webhooks/"+tt.webhookID+"/regenerate-secret", nil)
			req = addTenantContext(req, tt.tenantID)
			req = addRouteParam(req, "id", tt.webhookID)
			w := httptest.NewRecorder()

			handler.RegenerateSecret(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestTestWebhook(t *testing.T) {
	testResult := &webhook.TestResult{
		Success:      true,
		StatusCode:   200,
		ResponseTime: 150,
		ExecutionID:  "exec-1",
	}

	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		requestBody    interface{}
		setupMock      func(mockService *MockWebhookManagementService)
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:      "success",
			tenantID:  "tenant-123",
			webhookID: "webhook-1",
			requestBody: map[string]interface{}{
				"method":  "POST",
				"headers": map[string]string{"Content-Type": "application/json"},
				"body":    json.RawMessage(`{"test":"data"}`),
			},
			setupMock: func(mockService *MockWebhookManagementService) {
				mockService.On("TestWebhook",
					mock.Anything,
					"tenant-123",
					"webhook-1",
					"POST",
					map[string]string{"Content-Type": "application/json"},
					json.RawMessage(`{"test":"data"}`),
				).Return(testResult, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, true, body["success"])
				assert.Equal(t, float64(200), body["statusCode"])
				assert.Equal(t, float64(150), body["responseTimeMs"])
			},
		},
		{
			name:      "not found",
			tenantID:  "tenant-123",
			webhookID: "webhook-999",
			requestBody: map[string]interface{}{
				"method": "POST",
			},
			setupMock: func(mockService *MockWebhookManagementService) {
				mockService.On("TestWebhook",
					mock.Anything,
					"tenant-123",
					"webhook-999",
					"POST",
					mock.Anything,
					mock.Anything,
				).Return(nil, webhook.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "webhook not found")
			},
		},
		{
			name:           "invalid json body",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			requestBody:    "invalid json",
			setupMock:      func(mockService *MockWebhookManagementService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "invalid request body")
			},
		},
		{
			name:      "missing required method field",
			tenantID:  "tenant-123",
			webhookID: "webhook-1",
			requestBody: map[string]interface{}{
				"headers": map[string]string{"Content-Type": "application/json"},
			},
			setupMock:      func(mockService *MockWebhookManagementService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "error")
			},
		},
		{
			name:      "service error",
			tenantID:  "tenant-123",
			webhookID: "webhook-1",
			requestBody: map[string]interface{}{
				"method": "POST",
			},
			setupMock: func(mockService *MockWebhookManagementService) {
				mockService.On("TestWebhook",
					mock.Anything,
					"tenant-123",
					"webhook-1",
					"POST",
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("execution error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "failed to test webhook")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookManagementHandler()
			tt.setupMock(mockService)

			var bodyBytes []byte
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/api/v1/webhooks/"+tt.webhookID+"/test", bytes.NewReader(bodyBytes))
			req = addTenantContext(req, tt.tenantID)
			req = addRouteParam(req, "id", tt.webhookID)
			w := httptest.NewRecorder()

			handler.TestWebhook(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetEventHistory(t *testing.T) {
	now := time.Now()
	events := []*webhook.Event{
		{
			ID:           "event-1",
			WebhookID:    "webhook-1",
			ExecutionID:  "exec-1",
			Status:       "success",
			StatusCode:   200,
			ResponseTime: 120,
			CreatedAt:    now,
		},
	}

	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		queryParams    string
		expectedLimit  int
		expectedOffset int
		mockReturn     []*webhook.Event
		mockTotal      int
		mockError      error
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "success with default pagination",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			queryParams:    "",
			expectedLimit:  20,
			expectedOffset: 0,
			mockReturn:     events,
			mockTotal:      1,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, float64(1), body["total"])
				assert.Equal(t, float64(20), body["limit"])
				assert.Equal(t, float64(0), body["offset"])
				assert.NotNil(t, body["data"])
			},
		},
		{
			name:           "success with custom pagination",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			queryParams:    "?limit=10&offset=5",
			expectedLimit:  10,
			expectedOffset: 5,
			mockReturn:     events,
			mockTotal:      15,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, float64(15), body["total"])
				assert.Equal(t, float64(10), body["limit"])
				assert.Equal(t, float64(5), body["offset"])
			},
		},
		{
			name:           "webhook not found",
			tenantID:       "tenant-123",
			webhookID:      "webhook-999",
			queryParams:    "",
			expectedLimit:  20,
			expectedOffset: 0,
			mockReturn:     nil,
			mockTotal:      0,
			mockError:      webhook.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "webhook not found")
			},
		},
		{
			name:           "service error",
			tenantID:       "tenant-123",
			webhookID:      "webhook-1",
			queryParams:    "",
			expectedLimit:  20,
			expectedOffset: 0,
			mockReturn:     nil,
			mockTotal:      0,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "failed to get event history")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookManagementHandler()

			mockService.On("GetEventHistory",
				mock.Anything,
				tt.tenantID,
				tt.webhookID,
				tt.expectedLimit,
				tt.expectedOffset,
			).Return(tt.mockReturn, tt.mockTotal, tt.mockError)

			req := httptest.NewRequest("GET", "/api/v1/webhooks/"+tt.webhookID+"/events"+tt.queryParams, nil)
			req = addTenantContext(req, tt.tenantID)
			req = addRouteParam(req, "id", tt.webhookID)
			w := httptest.NewRecorder()

			handler.GetEventHistory(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}
