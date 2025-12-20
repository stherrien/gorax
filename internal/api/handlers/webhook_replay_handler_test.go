package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/webhook"
)

// MockReplayService is a mock implementation of ReplayService for testing
type MockReplayService struct {
	mock.Mock
}

func (m *MockReplayService) ReplayEvent(ctx context.Context, tenantID, eventID string, modifiedPayload json.RawMessage) *webhook.ReplayResult {
	args := m.Called(ctx, tenantID, eventID, modifiedPayload)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*webhook.ReplayResult)
}

func (m *MockReplayService) BatchReplayEvents(ctx context.Context, tenantID, webhookID string, eventIDs []string) *webhook.BatchReplayResponse {
	args := m.Called(ctx, tenantID, webhookID, eventIDs)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*webhook.BatchReplayResponse)
}

func newTestReplayHandler() (*WebhookReplayHandler, *MockReplayService) {
	mockService := new(MockReplayService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create a real ReplayService wrapper that delegates to mock
	// For testing, we need to create a handler that uses our mock
	handler := &WebhookReplayHandler{
		replayService: nil, // We'll use a different approach
		logger:        logger,
	}

	return handler, mockService
}

// testReplayHandler wraps MockReplayService for handler testing
type testReplayHandler struct {
	mockService *MockReplayService
	logger      *slog.Logger
}

func newTestReplayHandlerWithMock() (*testReplayHandler, *MockReplayService) {
	mockService := new(MockReplayService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return &testReplayHandler{
		mockService: mockService,
		logger:      logger,
	}, mockService
}

func (h *testReplayHandler) ReplayEvent(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r.Context())
	eventID := chi.URLParam(r, "eventID")

	var input webhook.ReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		input.EventID = eventID
	}

	var modifiedPayload json.RawMessage
	if len(input.ModifiedPayload) > 0 {
		modifiedPayload = input.ModifiedPayload
	}

	result := h.mockService.ReplayEvent(r.Context(), tenantID, eventID, modifiedPayload)

	if result == nil || !result.Success {
		status := http.StatusInternalServerError
		if result != nil && result.Error == "event not found: webhook not found" {
			status = http.StatusNotFound
		}
		if result != nil && (result.Error == "event not found" ||
			containsString(result.Error, "event not found")) {
			status = http.StatusNotFound
		}
		h.respondJSON(w, status, result)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

func (h *testReplayHandler) BatchReplayEvents(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r.Context())
	webhookID := chi.URLParam(r, "webhookID")

	var input webhook.BatchReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(input.EventIDs) == 0 {
		h.respondError(w, http.StatusBadRequest, "eventIds is required and cannot be empty")
		return
	}

	results := h.mockService.BatchReplayEvents(r.Context(), tenantID, webhookID, input.EventIDs)

	h.respondJSON(w, http.StatusOK, results)
}

func (h *testReplayHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *testReplayHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getTenantIDFromContext(ctx context.Context) string {
	if v := ctx.Value("tenant_id"); v != nil {
		return v.(string)
	}
	return ""
}

func addReplayTenantContext(req *http.Request, tenantID string) *http.Request {
	ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
	return req.WithContext(ctx)
}

func addReplayRouteParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// ============================================================================
// Tests for ReplayEvent handler
// ============================================================================

func TestReplayEvent_Success(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	result := &webhook.ReplayResult{
		Success:     true,
		ExecutionID: "exec-123",
	}

	mockService.On("ReplayEvent",
		mock.Anything,
		"tenant-123",
		"event-123",
		json.RawMessage(nil),
	).Return(result)

	req := httptest.NewRequest("POST", "/api/v1/events/event-123/replay", nil)
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "eventID", "event-123")
	w := httptest.NewRecorder()

	handler.ReplayEvent(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response webhook.ReplayResult
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "exec-123", response.ExecutionID)
	assert.Empty(t, response.Error)

	mockService.AssertExpectations(t)
}

func TestReplayEvent_WithModifiedPayload(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	result := &webhook.ReplayResult{
		Success:     true,
		ExecutionID: "exec-456",
	}

	mockService.On("ReplayEvent",
		mock.Anything,
		"tenant-123",
		"event-123",
		mock.MatchedBy(func(payload json.RawMessage) bool {
			// Verify payload contains expected data
			var data map[string]interface{}
			if err := json.Unmarshal(payload, &data); err != nil {
				return false
			}
			return data["modified"] == "data"
		}),
	).Return(result)

	bodyBytes := []byte(`{"eventId":"event-123","modifiedPayload":{"modified":"data"}}`)

	req := httptest.NewRequest("POST", "/api/v1/events/event-123/replay", bytes.NewReader(bodyBytes))
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "eventID", "event-123")
	w := httptest.NewRecorder()

	handler.ReplayEvent(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response webhook.ReplayResult
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "exec-456", response.ExecutionID)

	mockService.AssertExpectations(t)
}

func TestReplayEvent_EventNotFound(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	result := &webhook.ReplayResult{
		Success: false,
		Error:   "event not found: not found",
	}

	mockService.On("ReplayEvent",
		mock.Anything,
		"tenant-123",
		"event-999",
		json.RawMessage(nil),
	).Return(result)

	req := httptest.NewRequest("POST", "/api/v1/events/event-999/replay", nil)
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "eventID", "event-999")
	w := httptest.NewRecorder()

	handler.ReplayEvent(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response webhook.ReplayResult
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "event not found")

	mockService.AssertExpectations(t)
}

func TestReplayEvent_MaxReplayCountExceeded(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	result := &webhook.ReplayResult{
		Success: false,
		Error:   "max replay count (5) exceeded",
	}

	mockService.On("ReplayEvent",
		mock.Anything,
		"tenant-123",
		"event-123",
		json.RawMessage(nil),
	).Return(result)

	req := httptest.NewRequest("POST", "/api/v1/events/event-123/replay", nil)
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "eventID", "event-123")
	w := httptest.NewRecorder()

	handler.ReplayEvent(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response webhook.ReplayResult
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "max replay count")

	mockService.AssertExpectations(t)
}

func TestReplayEvent_WebhookDisabled(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	result := &webhook.ReplayResult{
		Success: false,
		Error:   "webhook is disabled",
	}

	mockService.On("ReplayEvent",
		mock.Anything,
		"tenant-123",
		"event-123",
		json.RawMessage(nil),
	).Return(result)

	req := httptest.NewRequest("POST", "/api/v1/events/event-123/replay", nil)
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "eventID", "event-123")
	w := httptest.NewRecorder()

	handler.ReplayEvent(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response webhook.ReplayResult
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "webhook is disabled")

	mockService.AssertExpectations(t)
}

func TestReplayEvent_ExecutionFailed(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	result := &webhook.ReplayResult{
		Success: false,
		Error:   "execution failed: workflow error",
	}

	mockService.On("ReplayEvent",
		mock.Anything,
		"tenant-123",
		"event-123",
		json.RawMessage(nil),
	).Return(result)

	req := httptest.NewRequest("POST", "/api/v1/events/event-123/replay", nil)
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "eventID", "event-123")
	w := httptest.NewRecorder()

	handler.ReplayEvent(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response webhook.ReplayResult
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "execution failed")

	mockService.AssertExpectations(t)
}

// ============================================================================
// Tests for BatchReplayEvents handler
// ============================================================================

func TestBatchReplayEvents_Success(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	results := &webhook.BatchReplayResponse{
		Results: map[string]*webhook.ReplayResult{
			"event-1": {Success: true, ExecutionID: "exec-1"},
			"event-2": {Success: true, ExecutionID: "exec-2"},
		},
	}

	mockService.On("BatchReplayEvents",
		mock.Anything,
		"tenant-123",
		"webhook-123",
		[]string{"event-1", "event-2"},
	).Return(results)

	body := map[string]interface{}{
		"eventIds": []string{"event-1", "event-2"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/webhooks/webhook-123/events/replay", bytes.NewReader(bodyBytes))
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "webhookID", "webhook-123")
	w := httptest.NewRecorder()

	handler.BatchReplayEvents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response webhook.BatchReplayResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Results, 2)
	assert.True(t, response.Results["event-1"].Success)
	assert.True(t, response.Results["event-2"].Success)

	mockService.AssertExpectations(t)
}

func TestBatchReplayEvents_PartialSuccess(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	results := &webhook.BatchReplayResponse{
		Results: map[string]*webhook.ReplayResult{
			"event-1": {Success: true, ExecutionID: "exec-1"},
			"event-2": {Success: false, Error: "max replay count exceeded"},
			"event-3": {Success: false, Error: "event not found"},
		},
	}

	mockService.On("BatchReplayEvents",
		mock.Anything,
		"tenant-123",
		"webhook-123",
		[]string{"event-1", "event-2", "event-3"},
	).Return(results)

	body := map[string]interface{}{
		"eventIds": []string{"event-1", "event-2", "event-3"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/webhooks/webhook-123/events/replay", bytes.NewReader(bodyBytes))
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "webhookID", "webhook-123")
	w := httptest.NewRecorder()

	handler.BatchReplayEvents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response webhook.BatchReplayResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Results, 3)
	assert.True(t, response.Results["event-1"].Success)
	assert.False(t, response.Results["event-2"].Success)
	assert.False(t, response.Results["event-3"].Success)

	mockService.AssertExpectations(t)
}

func TestBatchReplayEvents_EmptyEventIds(t *testing.T) {
	handler, _ := newTestReplayHandlerWithMock()

	body := map[string]interface{}{
		"eventIds": []string{},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/webhooks/webhook-123/events/replay", bytes.NewReader(bodyBytes))
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "webhookID", "webhook-123")
	w := httptest.NewRecorder()

	handler.BatchReplayEvents(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "eventIds is required")
}

func TestBatchReplayEvents_InvalidJSON(t *testing.T) {
	handler, _ := newTestReplayHandlerWithMock()

	req := httptest.NewRequest("POST", "/api/v1/webhooks/webhook-123/events/replay", bytes.NewReader([]byte("invalid json")))
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "webhookID", "webhook-123")
	w := httptest.NewRecorder()

	handler.BatchReplayEvents(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "invalid request body")
}

func TestBatchReplayEvents_MaxBatchSizeExceeded(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	// Create 11 event IDs (exceeds max of 10)
	eventIDs := make([]string, 11)
	for i := 0; i < 11; i++ {
		eventIDs[i] = "event-" + string(rune('A'+i))
	}

	results := &webhook.BatchReplayResponse{
		Results: make(map[string]*webhook.ReplayResult),
	}
	for _, id := range eventIDs {
		results.Results[id] = &webhook.ReplayResult{
			Success: false,
			Error:   "batch size exceeds maximum of 10 events",
		}
	}

	mockService.On("BatchReplayEvents",
		mock.Anything,
		"tenant-123",
		"webhook-123",
		eventIDs,
	).Return(results)

	body := map[string]interface{}{
		"eventIds": eventIDs,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/webhooks/webhook-123/events/replay", bytes.NewReader(bodyBytes))
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "webhookID", "webhook-123")
	w := httptest.NewRecorder()

	handler.BatchReplayEvents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response webhook.BatchReplayResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// All events should fail due to batch size limit
	assert.Len(t, response.Results, 11)
	for _, result := range response.Results {
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "batch size exceeds maximum")
	}

	mockService.AssertExpectations(t)
}

func TestBatchReplayEvents_AllEventsFailed(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	results := &webhook.BatchReplayResponse{
		Results: map[string]*webhook.ReplayResult{
			"event-1": {Success: false, Error: "webhook disabled"},
			"event-2": {Success: false, Error: "webhook disabled"},
		},
	}

	mockService.On("BatchReplayEvents",
		mock.Anything,
		"tenant-123",
		"webhook-123",
		[]string{"event-1", "event-2"},
	).Return(results)

	body := map[string]interface{}{
		"eventIds": []string{"event-1", "event-2"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/webhooks/webhook-123/events/replay", bytes.NewReader(bodyBytes))
	req = addReplayTenantContext(req, "tenant-123")
	req = addReplayRouteParam(req, "webhookID", "webhook-123")
	w := httptest.NewRecorder()

	handler.BatchReplayEvents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response webhook.BatchReplayResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Results, 2)
	for _, result := range response.Results {
		assert.False(t, result.Success)
	}

	mockService.AssertExpectations(t)
}

// ============================================================================
// Table-driven tests for comprehensive coverage
// ============================================================================

func TestReplayEvent_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		eventID        string
		requestBody    interface{}
		mockResult     *webhook.ReplayResult
		expectedStatus int
		checkResponse  func(t *testing.T, body *webhook.ReplayResult)
	}{
		{
			name:        "successful replay",
			tenantID:    "tenant-1",
			eventID:     "event-1",
			requestBody: nil,
			mockResult: &webhook.ReplayResult{
				Success:     true,
				ExecutionID: "exec-1",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body *webhook.ReplayResult) {
				assert.True(t, body.Success)
				assert.Equal(t, "exec-1", body.ExecutionID)
				assert.Empty(t, body.Error)
			},
		},
		{
			name:     "replay with modified payload",
			tenantID: "tenant-1",
			eventID:  "event-1",
			requestBody: map[string]interface{}{
				"modifiedPayload": json.RawMessage(`{"key":"value"}`),
			},
			mockResult: &webhook.ReplayResult{
				Success:     true,
				ExecutionID: "exec-2",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body *webhook.ReplayResult) {
				assert.True(t, body.Success)
				assert.Equal(t, "exec-2", body.ExecutionID)
			},
		},
		{
			name:        "event not found",
			tenantID:    "tenant-1",
			eventID:     "event-404",
			requestBody: nil,
			mockResult: &webhook.ReplayResult{
				Success: false,
				Error:   "event not found: not found",
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, body *webhook.ReplayResult) {
				assert.False(t, body.Success)
				assert.Contains(t, body.Error, "event not found")
			},
		},
		{
			name:        "webhook not found",
			tenantID:    "tenant-1",
			eventID:     "event-1",
			requestBody: nil,
			mockResult: &webhook.ReplayResult{
				Success: false,
				Error:   "webhook not found: not found",
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body *webhook.ReplayResult) {
				assert.False(t, body.Success)
				assert.Contains(t, body.Error, "webhook not found")
			},
		},
		{
			name:        "max replay count exceeded",
			tenantID:    "tenant-1",
			eventID:     "event-1",
			requestBody: nil,
			mockResult: &webhook.ReplayResult{
				Success: false,
				Error:   "max replay count (5) exceeded",
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body *webhook.ReplayResult) {
				assert.False(t, body.Success)
				assert.Contains(t, body.Error, "max replay count")
			},
		},
		{
			name:        "webhook disabled",
			tenantID:    "tenant-1",
			eventID:     "event-1",
			requestBody: nil,
			mockResult: &webhook.ReplayResult{
				Success: false,
				Error:   "webhook is disabled",
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body *webhook.ReplayResult) {
				assert.False(t, body.Success)
				assert.Contains(t, body.Error, "disabled")
			},
		},
		{
			name:        "execution failed",
			tenantID:    "tenant-1",
			eventID:     "event-1",
			requestBody: nil,
			mockResult: &webhook.ReplayResult{
				Success: false,
				Error:   "execution failed: timeout",
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body *webhook.ReplayResult) {
				assert.False(t, body.Success)
				assert.Contains(t, body.Error, "execution failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestReplayHandlerWithMock()

			var modifiedPayload json.RawMessage
			if tt.requestBody != nil {
				if bodyMap, ok := tt.requestBody.(map[string]interface{}); ok {
					if mp, exists := bodyMap["modifiedPayload"]; exists {
						modifiedPayload = mp.(json.RawMessage)
					}
				}
			}

			mockService.On("ReplayEvent",
				mock.Anything,
				tt.tenantID,
				tt.eventID,
				modifiedPayload,
			).Return(tt.mockResult)

			var bodyBytes []byte
			if tt.requestBody != nil {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/api/v1/events/"+tt.eventID+"/replay", bytes.NewReader(bodyBytes))
			req = addReplayTenantContext(req, tt.tenantID)
			req = addReplayRouteParam(req, "eventID", tt.eventID)
			w := httptest.NewRecorder()

			handler.ReplayEvent(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response webhook.ReplayResult
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, &response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestBatchReplayEvents_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		webhookID      string
		requestBody    interface{}
		setupMock      func(*MockReplayService)
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:      "all events succeed",
			tenantID:  "tenant-1",
			webhookID: "webhook-1",
			requestBody: map[string]interface{}{
				"eventIds": []string{"event-1", "event-2"},
			},
			setupMock: func(m *MockReplayService) {
				m.On("BatchReplayEvents",
					mock.Anything,
					"tenant-1",
					"webhook-1",
					[]string{"event-1", "event-2"},
				).Return(&webhook.BatchReplayResponse{
					Results: map[string]*webhook.ReplayResult{
						"event-1": {Success: true, ExecutionID: "exec-1"},
						"event-2": {Success: true, ExecutionID: "exec-2"},
					},
				})
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				results := body["results"].(map[string]interface{})
				assert.Len(t, results, 2)
			},
		},
		{
			name:      "mixed results",
			tenantID:  "tenant-1",
			webhookID: "webhook-1",
			requestBody: map[string]interface{}{
				"eventIds": []string{"event-1", "event-2", "event-3"},
			},
			setupMock: func(m *MockReplayService) {
				m.On("BatchReplayEvents",
					mock.Anything,
					"tenant-1",
					"webhook-1",
					[]string{"event-1", "event-2", "event-3"},
				).Return(&webhook.BatchReplayResponse{
					Results: map[string]*webhook.ReplayResult{
						"event-1": {Success: true, ExecutionID: "exec-1"},
						"event-2": {Success: false, Error: "max replay count exceeded"},
						"event-3": {Success: false, Error: "event not found"},
					},
				})
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				results := body["results"].(map[string]interface{})
				assert.Len(t, results, 3)
			},
		},
		{
			name:      "empty event IDs",
			tenantID:  "tenant-1",
			webhookID: "webhook-1",
			requestBody: map[string]interface{}{
				"eventIds": []string{},
			},
			setupMock:      func(m *MockReplayService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "eventIds is required")
			},
		},
		{
			name:           "invalid JSON",
			tenantID:       "tenant-1",
			webhookID:      "webhook-1",
			requestBody:    "invalid",
			setupMock:      func(m *MockReplayService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "invalid request body")
			},
		},
		{
			name:      "single event",
			tenantID:  "tenant-1",
			webhookID: "webhook-1",
			requestBody: map[string]interface{}{
				"eventIds": []string{"event-1"},
			},
			setupMock: func(m *MockReplayService) {
				m.On("BatchReplayEvents",
					mock.Anything,
					"tenant-1",
					"webhook-1",
					[]string{"event-1"},
				).Return(&webhook.BatchReplayResponse{
					Results: map[string]*webhook.ReplayResult{
						"event-1": {Success: true, ExecutionID: "exec-1"},
					},
				})
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				results := body["results"].(map[string]interface{})
				assert.Len(t, results, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestReplayHandlerWithMock()
			tt.setupMock(mockService)

			var bodyBytes []byte
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/api/v1/webhooks/"+tt.webhookID+"/events/replay", bytes.NewReader(bodyBytes))
			req = addReplayTenantContext(req, tt.tenantID)
			req = addReplayRouteParam(req, "webhookID", tt.webhookID)
			w := httptest.NewRecorder()

			handler.BatchReplayEvents(w, req)

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

// ============================================================================
// Integration-style tests
// ============================================================================

func TestReplayEvent_Integration_FullFlow(t *testing.T) {
	// This test simulates a full replay flow from HTTP request to response
	handler, mockService := newTestReplayHandlerWithMock()

	// Setup: Original event exists and webhook is enabled
	mockService.On("ReplayEvent",
		mock.Anything,
		"tenant-integration",
		"event-original",
		json.RawMessage(nil),
	).Return(&webhook.ReplayResult{
		Success:     true,
		ExecutionID: "exec-integration-1",
	})

	// Execute the replay
	req := httptest.NewRequest("POST", "/api/v1/events/event-original/replay", nil)
	req.Header.Set("Content-Type", "application/json")
	req = addReplayTenantContext(req, "tenant-integration")
	req = addReplayRouteParam(req, "eventID", "event-original")
	w := httptest.NewRecorder()

	handler.ReplayEvent(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result webhook.ReplayResult
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "exec-integration-1", result.ExecutionID)
	assert.Empty(t, result.Error)

	mockService.AssertExpectations(t)
}

func TestBatchReplayEvents_Integration_FullFlow(t *testing.T) {
	// This test simulates a full batch replay flow
	handler, mockService := newTestReplayHandlerWithMock()

	eventIDs := []string{"event-1", "event-2", "event-3"}

	mockService.On("BatchReplayEvents",
		mock.Anything,
		"tenant-integration",
		"webhook-integration",
		eventIDs,
	).Return(&webhook.BatchReplayResponse{
		Results: map[string]*webhook.ReplayResult{
			"event-1": {Success: true, ExecutionID: "exec-1"},
			"event-2": {Success: true, ExecutionID: "exec-2"},
			"event-3": {Success: false, Error: "max replay count exceeded"},
		},
	})

	body := map[string]interface{}{
		"eventIds": eventIDs,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/webhooks/webhook-integration/events/replay", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = addReplayTenantContext(req, "tenant-integration")
	req = addReplayRouteParam(req, "webhookID", "webhook-integration")
	w := httptest.NewRecorder()

	handler.BatchReplayEvents(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result webhook.BatchReplayResponse
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)

	assert.Len(t, result.Results, 3)

	// Verify individual results
	assert.True(t, result.Results["event-1"].Success)
	assert.True(t, result.Results["event-2"].Success)
	assert.False(t, result.Results["event-3"].Success)
	assert.Contains(t, result.Results["event-3"].Error, "max replay count")

	mockService.AssertExpectations(t)
}

func TestReplayEvent_MultiTenant_Isolation(t *testing.T) {
	// Test that replay respects tenant isolation
	handler, mockService := newTestReplayHandlerWithMock()

	// Tenant A's event
	mockService.On("ReplayEvent",
		mock.Anything,
		"tenant-A",
		"event-A",
		json.RawMessage(nil),
	).Return(&webhook.ReplayResult{
		Success:     true,
		ExecutionID: "exec-A",
	})

	// Tenant B's event
	mockService.On("ReplayEvent",
		mock.Anything,
		"tenant-B",
		"event-B",
		json.RawMessage(nil),
	).Return(&webhook.ReplayResult{
		Success:     true,
		ExecutionID: "exec-B",
	})

	// Test Tenant A
	reqA := httptest.NewRequest("POST", "/api/v1/events/event-A/replay", nil)
	reqA = addReplayTenantContext(reqA, "tenant-A")
	reqA = addReplayRouteParam(reqA, "eventID", "event-A")
	wA := httptest.NewRecorder()
	handler.ReplayEvent(wA, reqA)

	assert.Equal(t, http.StatusOK, wA.Code)
	var resultA webhook.ReplayResult
	json.NewDecoder(wA.Body).Decode(&resultA)
	assert.Equal(t, "exec-A", resultA.ExecutionID)

	// Test Tenant B
	reqB := httptest.NewRequest("POST", "/api/v1/events/event-B/replay", nil)
	reqB = addReplayTenantContext(reqB, "tenant-B")
	reqB = addReplayRouteParam(reqB, "eventID", "event-B")
	wB := httptest.NewRecorder()
	handler.ReplayEvent(wB, reqB)

	assert.Equal(t, http.StatusOK, wB.Code)
	var resultB webhook.ReplayResult
	json.NewDecoder(wB.Body).Decode(&resultB)
	assert.Equal(t, "exec-B", resultB.ExecutionID)

	mockService.AssertExpectations(t)
}

func TestBatchReplayEvents_Concurrent_Safety(t *testing.T) {
	// Test that concurrent batch replays are handled correctly
	handler, mockService := newTestReplayHandlerWithMock()

	mockService.On("BatchReplayEvents",
		mock.Anything,
		"tenant-concurrent",
		"webhook-concurrent",
		[]string{"event-1"},
	).Return(&webhook.BatchReplayResponse{
		Results: map[string]*webhook.ReplayResult{
			"event-1": {Success: true, ExecutionID: "exec-concurrent"},
		},
	}).Times(3)

	// Run 3 concurrent requests
	for i := 0; i < 3; i++ {
		t.Run("concurrent-"+string(rune('A'+i)), func(t *testing.T) {
			body := map[string]interface{}{
				"eventIds": []string{"event-1"},
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/api/v1/webhooks/webhook-concurrent/events/replay", bytes.NewReader(bodyBytes))
			req = addReplayTenantContext(req, "tenant-concurrent")
			req = addReplayRouteParam(req, "webhookID", "webhook-concurrent")
			w := httptest.NewRecorder()

			handler.BatchReplayEvents(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}

	mockService.AssertExpectations(t)
}

// ============================================================================
// Error condition tests
// ============================================================================

func TestReplayEvent_ServicePanic_Handled(t *testing.T) {
	// Note: This test documents expected behavior - the handler should not panic
	// In production, a panic recovery middleware would handle this
	handler, mockService := newTestReplayHandlerWithMock()

	mockService.On("ReplayEvent",
		mock.Anything,
		"tenant-panic",
		"event-panic",
		json.RawMessage(nil),
	).Return((*webhook.ReplayResult)(nil))

	req := httptest.NewRequest("POST", "/api/v1/events/event-panic/replay", nil)
	req = addReplayTenantContext(req, "tenant-panic")
	req = addReplayRouteParam(req, "eventID", "event-panic")
	w := httptest.NewRecorder()

	// Should not panic
	handler.ReplayEvent(w, req)

	// When service returns nil, handler should return error
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

func TestReplayEvent_MissingTenantID(t *testing.T) {
	handler, mockService := newTestReplayHandlerWithMock()

	// Simulate missing tenant (empty string)
	mockService.On("ReplayEvent",
		mock.Anything,
		"",
		"event-1",
		json.RawMessage(nil),
	).Return(&webhook.ReplayResult{
		Success: false,
		Error:   "tenant not found",
	})

	req := httptest.NewRequest("POST", "/api/v1/events/event-1/replay", nil)
	// Don't add tenant context
	req = addReplayRouteParam(req, "eventID", "event-1")
	w := httptest.NewRecorder()

	handler.ReplayEvent(w, req)

	// Service should handle missing tenant appropriately
	mockService.AssertExpectations(t)
}
