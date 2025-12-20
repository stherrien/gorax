package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/eventtypes"
)

// Alias for convenience
type EventType = eventtypes.EventType

// MockEventTypeService is a mock implementation for testing
type MockEventTypeService struct {
	mock.Mock
}

func (m *MockEventTypeService) ListEventTypes(ctx context.Context) ([]EventType, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]EventType), args.Error(1)
}

func newTestEventTypesHandler() (*EventTypesHandler, *MockEventTypeService) {
	mockService := new(MockEventTypeService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewEventTypesHandler(mockService, logger)
	return handler, mockService
}

func TestListEventTypes(t *testing.T) {
	now := time.Now()

	schema1 := json.RawMessage(`{"type":"object","properties":{"method":{"type":"string"}}}`)
	schema2 := json.RawMessage(`{"type":"object","properties":{"execution_id":{"type":"string"}}}`)

	eventTypes := []EventType{
		{
			ID:          "et-1",
			Name:        "webhook.received",
			Description: "Webhook request received",
			Schema:      schema1,
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "et-2",
			Name:        "execution.started",
			Description: "Workflow execution started",
			Schema:      schema2,
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	tests := []struct {
		name           string
		mockReturn     []EventType
		mockError      error
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "success with event types",
			mockReturn:     eventTypes,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.Len(t, data, 2)

				firstEvent := data[0].(map[string]interface{})
				assert.Equal(t, "webhook.received", firstEvent["name"])
				assert.Equal(t, "Webhook request received", firstEvent["description"])
				assert.NotNil(t, firstEvent["schema"])
				assert.Equal(t, float64(1), firstEvent["version"])
			},
		},
		{
			name:           "success with empty list",
			mockReturn:     []EventType{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				data := body["data"].([]interface{})
				assert.Len(t, data, 0)
			},
		},
		{
			name:           "service error",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "failed to list event types")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestEventTypesHandler()

			mockService.On("ListEventTypes", mock.Anything).Return(tt.mockReturn, tt.mockError)

			req := httptest.NewRequest("GET", "/api/v1/event-types", nil)
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

func TestListEventTypesValidSchema(t *testing.T) {
	now := time.Now()

	// Test that JSON schema is properly returned
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"method": {"type": "string"},
			"path": {"type": "string"},
			"timestamp": {"type": "string", "format": "date-time"}
		},
		"required": ["method", "path", "timestamp"]
	}`)

	eventTypes := []EventType{
		{
			ID:          "et-1",
			Name:        "webhook.received",
			Description: "Webhook request received",
			Schema:      schema,
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	handler, mockService := newTestEventTypesHandler()
	mockService.On("ListEventTypes", mock.Anything).Return(eventTypes, nil)

	req := httptest.NewRequest("GET", "/api/v1/event-types", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	firstEvent := data[0].(map[string]interface{})

	// Verify schema structure
	schemaData := firstEvent["schema"].(map[string]interface{})
	assert.Equal(t, "object", schemaData["type"])
	assert.NotNil(t, schemaData["properties"])
	assert.NotNil(t, schemaData["required"])

	// Verify required fields
	required := schemaData["required"].([]interface{})
	assert.Contains(t, required, "method")
	assert.Contains(t, required, "path")
	assert.Contains(t, required, "timestamp")
}

func TestListEventTypesIncludesAllCoreTypes(t *testing.T) {
	// Test that all core event types defined in migration are returned
	coreEventTypes := []string{
		"webhook.received",
		"webhook.processed",
		"webhook.filtered",
		"webhook.failed",
		"execution.started",
		"execution.completed",
		"execution.failed",
		"execution.cancelled",
		"step.started",
		"step.completed",
		"step.failed",
		"step.retrying",
	}

	now := time.Now()
	var eventTypes []EventType
	for i, name := range coreEventTypes {
		eventTypes = append(eventTypes, EventType{
			ID:          string(rune('a' + i)),
			Name:        name,
			Description: "Test event type",
			Schema:      json.RawMessage(`{}`),
			Version:     1,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	handler, mockService := newTestEventTypesHandler()
	mockService.On("ListEventTypes", mock.Anything).Return(eventTypes, nil)

	req := httptest.NewRequest("GET", "/api/v1/event-types", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	assert.Len(t, data, len(coreEventTypes))

	// Verify all core types are present
	returnedNames := make(map[string]bool)
	for _, item := range data {
		eventType := item.(map[string]interface{})
		returnedNames[eventType["name"].(string)] = true
	}

	for _, expectedName := range coreEventTypes {
		assert.True(t, returnedNames[expectedName], "Expected event type %s not found", expectedName)
	}
}

func TestListEventTypesContentType(t *testing.T) {
	handler, mockService := newTestEventTypesHandler()
	mockService.On("ListEventTypes", mock.Anything).Return([]EventType{}, nil)

	req := httptest.NewRequest("GET", "/api/v1/event-types", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}
