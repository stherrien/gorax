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

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/tenant"
	"github.com/gorax/gorax/internal/webhook"
)

// MockWebhookFilterService is a mock implementation of WebhookFilterService for testing
type MockWebhookFilterService struct {
	mock.Mock
}

func (m *MockWebhookFilterService) ListFilters(ctx context.Context, tenantID, webhookID string) ([]*webhook.WebhookFilter, error) {
	args := m.Called(ctx, tenantID, webhookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*webhook.WebhookFilter), args.Error(1)
}

func (m *MockWebhookFilterService) GetFilter(ctx context.Context, tenantID, webhookID, filterID string) (*webhook.WebhookFilter, error) {
	args := m.Called(ctx, tenantID, webhookID, filterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhook.WebhookFilter), args.Error(1)
}

func (m *MockWebhookFilterService) CreateFilter(ctx context.Context, tenantID, webhookID string, filter *webhook.WebhookFilter) (*webhook.WebhookFilter, error) {
	args := m.Called(ctx, tenantID, webhookID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhook.WebhookFilter), args.Error(1)
}

func (m *MockWebhookFilterService) UpdateFilter(ctx context.Context, tenantID, webhookID, filterID string, filter *webhook.WebhookFilter) (*webhook.WebhookFilter, error) {
	args := m.Called(ctx, tenantID, webhookID, filterID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhook.WebhookFilter), args.Error(1)
}

func (m *MockWebhookFilterService) DeleteFilter(ctx context.Context, tenantID, webhookID, filterID string) error {
	args := m.Called(ctx, tenantID, webhookID, filterID)
	return args.Error(0)
}

func (m *MockWebhookFilterService) TestFilters(ctx context.Context, tenantID, webhookID string, payload map[string]any) (*webhook.FilterResult, error) {
	args := m.Called(ctx, tenantID, webhookID, payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhook.FilterResult), args.Error(1)
}

func newTestWebhookFilterHandler() (*WebhookFilterHandler, *MockWebhookFilterService) {
	mockService := new(MockWebhookFilterService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewWebhookFilterHandler(mockService, logger)
	return handler, mockService
}

func addWebhookFilterContext(req *http.Request, tenantID string) *http.Request {
	t := &tenant.Tenant{
		ID:     tenantID,
		Status: "active",
	}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, t)
	return req.WithContext(ctx)
}

func addWebhookFilterURLParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// Test fixtures
func createTestWebhookFilter() *webhook.WebhookFilter {
	return &webhook.WebhookFilter{
		ID:         "filter-123",
		WebhookID:  "webhook-123",
		FieldPath:  "$.data.status",
		Operator:   webhook.OpEquals,
		Value:      "active",
		LogicGroup: 0,
		Enabled:    true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func createTestFilterResult(passed bool) *webhook.FilterResult {
	return &webhook.FilterResult{
		Passed: passed,
		Reason: "All filters passed",
		Details: map[string]interface{}{
			"filtersEvaluated": 1,
			"filtersPassed":    1,
		},
	}
}

// ============================================================================
// List Handler Tests
// ============================================================================

func TestWebhookFilterHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		setupMock      func(*MockWebhookFilterService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "successful list",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			setupMock: func(m *MockWebhookFilterService) {
				filters := []*webhook.WebhookFilter{createTestWebhookFilter()}
				m.On("ListFilters", mock.Anything, "tenant-123", "webhook-123").Return(filters, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				data := resp["data"].([]interface{})
				assert.Len(t, data, 1)
			},
		},
		{
			name:      "empty list",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			setupMock: func(m *MockWebhookFilterService) {
				m.On("ListFilters", mock.Anything, "tenant-123", "webhook-123").Return([]*webhook.WebhookFilter{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "webhook not found",
			tenantID:  "tenant-123",
			webhookID: "nonexistent",
			setupMock: func(m *MockWebhookFilterService) {
				m.On("ListFilters", mock.Anything, "tenant-123", "nonexistent").Return(nil, webhook.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "webhook not found",
		},
		{
			name:      "service error",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			setupMock: func(m *MockWebhookFilterService) {
				m.On("ListFilters", mock.Anything, "tenant-123", "webhook-123").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to list filters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookFilterHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/"+tt.webhookID+"/filters", nil)
			req = addWebhookFilterContext(req, tt.tenantID)
			req = addWebhookFilterURLParams(req, map[string]string{"id": tt.webhookID})

			rr := httptest.NewRecorder()
			handler.List(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Get Handler Tests
// ============================================================================

func TestWebhookFilterHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		filterID       string
		setupMock      func(*MockWebhookFilterService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful get",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			filterID:  "filter-123",
			setupMock: func(m *MockWebhookFilterService) {
				m.On("GetFilter", mock.Anything, "tenant-123", "webhook-123", "filter-123").
					Return(createTestWebhookFilter(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "filter not found",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			filterID:  "nonexistent",
			setupMock: func(m *MockWebhookFilterService) {
				m.On("GetFilter", mock.Anything, "tenant-123", "webhook-123", "nonexistent").
					Return(nil, webhook.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "filter not found",
		},
		{
			name:      "service error",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			filterID:  "filter-123",
			setupMock: func(m *MockWebhookFilterService) {
				m.On("GetFilter", mock.Anything, "tenant-123", "webhook-123", "filter-123").
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to get filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookFilterHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/"+tt.webhookID+"/filters/"+tt.filterID, nil)
			req = addWebhookFilterContext(req, tt.tenantID)
			req = addWebhookFilterURLParams(req, map[string]string{
				"id":       tt.webhookID,
				"filterID": tt.filterID,
			})

			rr := httptest.NewRecorder()
			handler.Get(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Create Handler Tests
// ============================================================================

func TestWebhookFilterHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		body           interface{}
		setupMock      func(*MockWebhookFilterService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful create",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			body: CreateFilterRequest{
				FieldPath:  "$.data.status",
				Operator:   "equals",
				Value:      "active",
				LogicGroup: 0,
				Enabled:    true,
			},
			setupMock: func(m *MockWebhookFilterService) {
				m.On("CreateFilter", mock.Anything, "tenant-123", "webhook-123", mock.AnythingOfType("*webhook.WebhookFilter")).
					Return(createTestWebhookFilter(), nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			tenantID:       "tenant-123",
			webhookID:      "webhook-123",
			body:           "invalid json",
			setupMock:      func(m *MockWebhookFilterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:      "validation error - missing field path",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			body: CreateFilterRequest{
				FieldPath: "", // Required field missing
				Operator:  "equals",
				Value:     "active",
			},
			setupMock:      func(m *MockWebhookFilterService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "validation error - invalid operator",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			body: CreateFilterRequest{
				FieldPath: "$.data.status",
				Operator:  "invalid_operator",
				Value:     "active",
			},
			setupMock:      func(m *MockWebhookFilterService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "webhook not found",
			tenantID:  "tenant-123",
			webhookID: "nonexistent",
			body: CreateFilterRequest{
				FieldPath:  "$.data.status",
				Operator:   "equals",
				Value:      "active",
				LogicGroup: 0,
				Enabled:    true,
			},
			setupMock: func(m *MockWebhookFilterService) {
				m.On("CreateFilter", mock.Anything, "tenant-123", "nonexistent", mock.AnythingOfType("*webhook.WebhookFilter")).
					Return(nil, webhook.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "webhook not found",
		},
		{
			name:      "service error",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			body: CreateFilterRequest{
				FieldPath:  "$.data.status",
				Operator:   "equals",
				Value:      "active",
				LogicGroup: 0,
				Enabled:    true,
			},
			setupMock: func(m *MockWebhookFilterService) {
				m.On("CreateFilter", mock.Anything, "tenant-123", "webhook-123", mock.AnythingOfType("*webhook.WebhookFilter")).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to create filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookFilterHandler()
			tt.setupMock(mockService)

			var body []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/"+tt.webhookID+"/filters", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = addWebhookFilterContext(req, tt.tenantID)
			req = addWebhookFilterURLParams(req, map[string]string{"id": tt.webhookID})

			rr := httptest.NewRecorder()
			handler.Create(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Update Handler Tests
// ============================================================================

func TestWebhookFilterHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		filterID       string
		body           interface{}
		setupMock      func(*MockWebhookFilterService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful update",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			filterID:  "filter-123",
			body: UpdateFilterRequest{
				FieldPath:  "$.data.new_status",
				Operator:   "contains",
				Value:      "test",
				LogicGroup: 1,
				Enabled:    false,
			},
			setupMock: func(m *MockWebhookFilterService) {
				m.On("UpdateFilter", mock.Anything, "tenant-123", "webhook-123", "filter-123", mock.AnythingOfType("*webhook.WebhookFilter")).
					Return(createTestWebhookFilter(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			tenantID:       "tenant-123",
			webhookID:      "webhook-123",
			filterID:       "filter-123",
			body:           "invalid json",
			setupMock:      func(m *MockWebhookFilterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:      "validation error - missing field path",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			filterID:  "filter-123",
			body: UpdateFilterRequest{
				FieldPath: "",
				Operator:  "equals",
				Value:     "active",
			},
			setupMock:      func(m *MockWebhookFilterService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "filter not found",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			filterID:  "nonexistent",
			body: UpdateFilterRequest{
				FieldPath:  "$.data.status",
				Operator:   "equals",
				Value:      "active",
				LogicGroup: 0,
				Enabled:    true,
			},
			setupMock: func(m *MockWebhookFilterService) {
				m.On("UpdateFilter", mock.Anything, "tenant-123", "webhook-123", "nonexistent", mock.AnythingOfType("*webhook.WebhookFilter")).
					Return(nil, webhook.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "filter not found",
		},
		{
			name:      "service error",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			filterID:  "filter-123",
			body: UpdateFilterRequest{
				FieldPath:  "$.data.status",
				Operator:   "equals",
				Value:      "active",
				LogicGroup: 0,
				Enabled:    true,
			},
			setupMock: func(m *MockWebhookFilterService) {
				m.On("UpdateFilter", mock.Anything, "tenant-123", "webhook-123", "filter-123", mock.AnythingOfType("*webhook.WebhookFilter")).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to update filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookFilterHandler()
			tt.setupMock(mockService)

			var body []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/webhooks/"+tt.webhookID+"/filters/"+tt.filterID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = addWebhookFilterContext(req, tt.tenantID)
			req = addWebhookFilterURLParams(req, map[string]string{
				"id":       tt.webhookID,
				"filterID": tt.filterID,
			})

			rr := httptest.NewRecorder()
			handler.Update(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Delete Handler Tests
// ============================================================================

func TestWebhookFilterHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		filterID       string
		setupMock      func(*MockWebhookFilterService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful delete",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			filterID:  "filter-123",
			setupMock: func(m *MockWebhookFilterService) {
				m.On("DeleteFilter", mock.Anything, "tenant-123", "webhook-123", "filter-123").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:      "filter not found",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			filterID:  "nonexistent",
			setupMock: func(m *MockWebhookFilterService) {
				m.On("DeleteFilter", mock.Anything, "tenant-123", "webhook-123", "nonexistent").
					Return(webhook.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "filter not found",
		},
		{
			name:      "service error",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			filterID:  "filter-123",
			setupMock: func(m *MockWebhookFilterService) {
				m.On("DeleteFilter", mock.Anything, "tenant-123", "webhook-123", "filter-123").
					Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to delete filter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookFilterHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/webhooks/"+tt.webhookID+"/filters/"+tt.filterID, nil)
			req = addWebhookFilterContext(req, tt.tenantID)
			req = addWebhookFilterURLParams(req, map[string]string{
				"id":       tt.webhookID,
				"filterID": tt.filterID,
			})

			rr := httptest.NewRecorder()
			handler.Delete(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test Handler Tests
// ============================================================================

func TestWebhookFilterHandler_Test(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		body           interface{}
		setupMock      func(*MockWebhookFilterService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "successful test - filters pass",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			body: TestFiltersRequest{
				Payload: map[string]any{
					"data": map[string]any{
						"status": "active",
					},
				},
			},
			setupMock: func(m *MockWebhookFilterService) {
				m.On("TestFilters", mock.Anything, "tenant-123", "webhook-123", mock.AnythingOfType("map[string]interface {}")).
					Return(createTestFilterResult(true), nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var resp webhook.FilterResult
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.True(t, resp.Passed)
			},
		},
		{
			name:      "successful test - filters fail",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			body: TestFiltersRequest{
				Payload: map[string]any{
					"data": map[string]any{
						"status": "inactive",
					},
				},
			},
			setupMock: func(m *MockWebhookFilterService) {
				result := &webhook.FilterResult{
					Passed: false,
					Reason: "Filter condition not met",
				}
				m.On("TestFilters", mock.Anything, "tenant-123", "webhook-123", mock.AnythingOfType("map[string]interface {}")).
					Return(result, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var resp webhook.FilterResult
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.False(t, resp.Passed)
			},
		},
		{
			name:           "invalid request body",
			tenantID:       "tenant-123",
			webhookID:      "webhook-123",
			body:           "invalid json",
			setupMock:      func(m *MockWebhookFilterService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:      "validation error - missing payload",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			body: TestFiltersRequest{
				Payload: nil,
			},
			setupMock:      func(m *MockWebhookFilterService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "webhook not found",
			tenantID:  "tenant-123",
			webhookID: "nonexistent",
			body: TestFiltersRequest{
				Payload: map[string]any{"test": "data"},
			},
			setupMock: func(m *MockWebhookFilterService) {
				m.On("TestFilters", mock.Anything, "tenant-123", "nonexistent", mock.AnythingOfType("map[string]interface {}")).
					Return(nil, webhook.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "webhook not found",
		},
		{
			name:      "service error",
			tenantID:  "tenant-123",
			webhookID: "webhook-123",
			body: TestFiltersRequest{
				Payload: map[string]any{"test": "data"},
			},
			setupMock: func(m *MockWebhookFilterService) {
				m.On("TestFilters", mock.Anything, "tenant-123", "webhook-123", mock.AnythingOfType("map[string]interface {}")).
					Return(nil, errors.New("evaluation error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to test filters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestWebhookFilterHandler()
			tt.setupMock(mockService)

			var body []byte
			var err error
			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/"+tt.webhookID+"/filters/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = addWebhookFilterContext(req, tt.tenantID)
			req = addWebhookFilterURLParams(req, map[string]string{"id": tt.webhookID})

			rr := httptest.NewRecorder()
			handler.Test(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
		})
	}
}
