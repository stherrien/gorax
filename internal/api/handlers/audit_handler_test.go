package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gorax/gorax/internal/audit"
)

// MockAuditService implements AuditService for testing
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) QueryAuditEvents(ctx context.Context, filter audit.QueryFilter) ([]audit.AuditEvent, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]audit.AuditEvent), args.Int(1), args.Error(2)
}

func (m *MockAuditService) GetAuditEvent(ctx context.Context, tenantID, eventID string) (*audit.AuditEvent, error) {
	args := m.Called(ctx, tenantID, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*audit.AuditEvent), args.Error(1)
}

func (m *MockAuditService) GetAuditStats(ctx context.Context, tenantID string, timeRange audit.TimeRange) (*audit.AuditStats, error) {
	args := m.Called(ctx, tenantID, timeRange)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*audit.AuditStats), args.Error(1)
}

func createTestAuditEvent() *audit.AuditEvent {
	return &audit.AuditEvent{
		ID:           "event-123",
		TenantID:     "tenant-123",
		UserID:       "user-123",
		UserEmail:    "user@example.com",
		Category:     audit.CategoryWorkflow,
		EventType:    audit.EventTypeCreate,
		Action:       "workflow.create",
		ResourceType: string(audit.ResourceTypeWorkflow),
		ResourceID:   "workflow-123",
		ResourceName: "Test Workflow",
		IPAddress:    "192.168.1.1",
		Severity:     audit.SeverityInfo,
		Status:       audit.StatusSuccess,
		CreatedAt:    time.Now(),
	}
}

func createTestAuditStats() *audit.AuditStats {
	return &audit.AuditStats{
		TotalEvents: 100,
		EventsByCategory: map[audit.Category]int{
			audit.CategoryWorkflow:       50,
			audit.CategoryAuthentication: 30,
			audit.CategoryDataAccess:     20,
		},
		EventsBySeverity: map[audit.Severity]int{
			audit.SeverityInfo:    80,
			audit.SeverityWarning: 15,
			audit.SeverityError:   5,
		},
		EventsByStatus: map[audit.Status]int{
			audit.StatusSuccess: 95,
			audit.StatusFailure: 5,
		},
		CriticalEvents: 2,
		FailedEvents:   5,
	}
}

func TestAuditHandler_QueryEvents(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockAuditService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful query with no filters",
			queryParams: "",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return f.Limit == 50 && f.Offset == 0
				})).Return(events, 1, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(1), response["total"])
				assert.Equal(t, float64(50), response["limit"])
				assert.Equal(t, float64(0), response["offset"])
			},
		},
		{
			name:        "successful query with tenant filter",
			queryParams: "?tenant_id=tenant-123",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return f.TenantID == "tenant-123"
				})).Return(events, 1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful query with user filters",
			queryParams: "?user_id=user-123&user_email=user@example.com",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return f.UserID == "user-123" && f.UserEmail == "user@example.com"
				})).Return(events, 1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful query with pagination",
			queryParams: "?limit=10&offset=20",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return f.Limit == 10 && f.Offset == 20
				})).Return(events, 100, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(100), response["total"])
				assert.Equal(t, float64(10), response["limit"])
				assert.Equal(t, float64(20), response["offset"])
			},
		},
		{
			name:        "successful query with sorting",
			queryParams: "?sort_by=created_at&sort_direction=desc",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return f.SortBy == "created_at" && f.SortDirection == "desc"
				})).Return(events, 1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful query with category filter",
			queryParams: "?category=workflow&category=authentication",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return len(f.Categories) == 2 &&
						f.Categories[0] == audit.CategoryWorkflow &&
						f.Categories[1] == audit.CategoryAuthentication
				})).Return(events, 2, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful query with event type filter",
			queryParams: "?event_type=create&event_type=update",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return len(f.EventTypes) == 2 &&
						f.EventTypes[0] == audit.EventTypeCreate &&
						f.EventTypes[1] == audit.EventTypeUpdate
				})).Return(events, 2, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful query with severity filter",
			queryParams: "?severity=info&severity=warning",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return len(f.Severities) == 2 &&
						f.Severities[0] == audit.SeverityInfo &&
						f.Severities[1] == audit.SeverityWarning
				})).Return(events, 2, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful query with status filter",
			queryParams: "?status=success&status=failure",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return len(f.Statuses) == 2 &&
						f.Statuses[0] == audit.StatusSuccess &&
						f.Statuses[1] == audit.StatusFailure
				})).Return(events, 2, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful query with date range",
			queryParams: "?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return !f.StartDate.IsZero() && !f.EndDate.IsZero()
				})).Return(events, 10, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid date format is ignored",
			queryParams: "?start_date=invalid&end_date=also-invalid",
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return f.StartDate.IsZero() && f.EndDate.IsZero()
				})).Return(events, 1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service error",
			queryParams: "",
			mockSetup: func(m *MockAuditService) {
				m.On("QueryAuditEvents", mock.Anything, mock.Anything).
					Return(nil, 0, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuditService)
			tt.mockSetup(mockService)

			handler := NewAuditHandler(mockService, slog.New(slog.NewTextHandler(io.Discard, nil)))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit/events"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			handler.QueryEvents(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuditHandler_GetEvent(t *testing.T) {
	tests := []struct {
		name           string
		eventID        string
		tenantID       string
		mockSetup      func(*MockAuditService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:     "successful get event",
			eventID:  "event-123",
			tenantID: "tenant-123",
			mockSetup: func(m *MockAuditService) {
				event := createTestAuditEvent()
				m.On("GetAuditEvent", mock.Anything, "tenant-123", "event-123").Return(event, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var event audit.AuditEvent
				err := json.Unmarshal(rr.Body.Bytes(), &event)
				assert.NoError(t, err)
				assert.Equal(t, "event-123", event.ID)
			},
		},
		{
			name:     "successful get event without tenant filter",
			eventID:  "event-123",
			tenantID: "",
			mockSetup: func(m *MockAuditService) {
				event := createTestAuditEvent()
				m.On("GetAuditEvent", mock.Anything, "", "event-123").Return(event, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing event ID",
			eventID:        "",
			tenantID:       "",
			mockSetup:      func(m *MockAuditService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "event not found",
			eventID:  "nonexistent",
			tenantID: "",
			mockSetup: func(m *MockAuditService) {
				m.On("GetAuditEvent", mock.Anything, "", "nonexistent").
					Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuditService)
			tt.mockSetup(mockService)

			handler := NewAuditHandler(mockService, slog.New(slog.NewTextHandler(io.Discard, nil)))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit/events/"+tt.eventID, nil)
			if tt.tenantID != "" {
				q := req.URL.Query()
				q.Add("tenant_id", tt.tenantID)
				req.URL.RawQuery = q.Encode()
			}

			// Add chi URL param
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.eventID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler.GetEvent(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuditHandler_GetStats(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockAuditService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful get stats with defaults",
			queryParams: "",
			mockSetup: func(m *MockAuditService) {
				stats := createTestAuditStats()
				m.On("GetAuditStats", mock.Anything, "", mock.MatchedBy(func(tr audit.TimeRange) bool {
					// Default time range is last 24 hours
					return !tr.StartDate.IsZero() && !tr.EndDate.IsZero()
				})).Return(stats, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var stats audit.AuditStats
				err := json.Unmarshal(rr.Body.Bytes(), &stats)
				assert.NoError(t, err)
				assert.Equal(t, 100, stats.TotalEvents)
			},
		},
		{
			name:        "successful get stats with tenant filter",
			queryParams: "?tenant_id=tenant-123",
			mockSetup: func(m *MockAuditService) {
				stats := createTestAuditStats()
				m.On("GetAuditStats", mock.Anything, "tenant-123", mock.Anything).Return(stats, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful get stats with custom date range",
			queryParams: "?start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z",
			mockSetup: func(m *MockAuditService) {
				stats := createTestAuditStats()
				m.On("GetAuditStats", mock.Anything, "", mock.MatchedBy(func(tr audit.TimeRange) bool {
					return tr.StartDate.Year() == 2024 && tr.StartDate.Month() == 1 &&
						tr.EndDate.Year() == 2024 && tr.EndDate.Month() == 1
				})).Return(stats, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid start date uses default",
			queryParams: "?start_date=invalid",
			mockSetup: func(m *MockAuditService) {
				stats := createTestAuditStats()
				m.On("GetAuditStats", mock.Anything, "", mock.MatchedBy(func(tr audit.TimeRange) bool {
					// Should default to ~24 hours ago
					return time.Since(tr.StartDate) < 25*time.Hour
				})).Return(stats, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid end date uses default",
			queryParams: "?end_date=invalid",
			mockSetup: func(m *MockAuditService) {
				stats := createTestAuditStats()
				m.On("GetAuditStats", mock.Anything, "", mock.MatchedBy(func(tr audit.TimeRange) bool {
					// Should default to now
					return time.Since(tr.EndDate) < time.Second
				})).Return(stats, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "service error",
			queryParams: "",
			mockSetup: func(m *MockAuditService) {
				m.On("GetAuditStats", mock.Anything, "", mock.Anything).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuditService)
			tt.mockSetup(mockService)

			handler := NewAuditHandler(mockService, slog.New(slog.NewTextHandler(io.Discard, nil)))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit/stats"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			handler.GetStats(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuditHandler_ExportEvents(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*MockAuditService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful JSON export",
			requestBody: `{"format": "json"}`,
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return f.Limit == 10000
				})).Return(events, 1, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				assert.Contains(t, rr.Header().Get("Content-Disposition"), "attachment")
				assert.Contains(t, rr.Header().Get("Content-Disposition"), ".json")

				var events []audit.AuditEvent
				err := json.Unmarshal(rr.Body.Bytes(), &events)
				assert.NoError(t, err)
				assert.Len(t, events, 1)
			},
		},
		{
			name:        "successful CSV export",
			requestBody: `{"format": "csv"}`,
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.Anything).Return(events, 1, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, "text/csv", rr.Header().Get("Content-Type"))
				assert.Contains(t, rr.Header().Get("Content-Disposition"), "attachment")
				assert.Contains(t, rr.Header().Get("Content-Disposition"), ".csv")

				// Check CSV content has headers and data
				body := rr.Body.String()
				assert.Contains(t, body, "ID,Tenant ID,User ID")
				assert.Contains(t, body, "event-123")
			},
		},
		{
			name:        "default format is JSON when not specified",
			requestBody: `{}`,
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.Anything).Return(events, 1, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{
			name:        "export with filters",
			requestBody: `{"tenantId": "tenant-123", "categories": ["workflow"], "eventTypes": ["create"], "severities": ["info"]}`,
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return f.TenantID == "tenant-123" &&
						len(f.Categories) == 1 && f.Categories[0] == audit.CategoryWorkflow &&
						len(f.EventTypes) == 1 && f.EventTypes[0] == audit.EventTypeCreate &&
						len(f.Severities) == 1 && f.Severities[0] == audit.SeverityInfo
				})).Return(events, 1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "export with date range",
			requestBody: `{"startDate": "2024-01-01T00:00:00Z", "endDate": "2024-01-31T23:59:59Z"}`,
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.MatchedBy(func(f audit.QueryFilter) bool {
					return f.StartDate.Year() == 2024 && f.EndDate.Year() == 2024
				})).Return(events, 1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockAuditService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "unsupported export format",
			requestBody: `{"format": "xml"}`,
			mockSetup: func(m *MockAuditService) {
				events := []audit.AuditEvent{*createTestAuditEvent()}
				m.On("QueryAuditEvents", mock.Anything, mock.Anything).Return(events, 1, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error during export",
			requestBody: `{"format": "json"}`,
			mockSetup: func(m *MockAuditService) {
				m.On("QueryAuditEvents", mock.Anything, mock.Anything).
					Return(nil, 0, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAuditService)
			tt.mockSetup(mockService)

			handler := NewAuditHandler(mockService, slog.New(slog.NewTextHandler(io.Discard, nil)))

			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/audit/export", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ExportEvents(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetIntParam(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  string
		key          string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer",
			queryParams:  "?limit=25",
			key:          "limit",
			defaultValue: 50,
			expected:     25,
		},
		{
			name:         "missing parameter uses default",
			queryParams:  "",
			key:          "limit",
			defaultValue: 50,
			expected:     50,
		},
		{
			name:         "invalid integer uses default",
			queryParams:  "?limit=invalid",
			key:          "limit",
			defaultValue: 50,
			expected:     50,
		},
		{
			name:         "empty value uses default",
			queryParams:  "?limit=",
			key:          "limit",
			defaultValue: 50,
			expected:     50,
		},
		{
			name:         "zero value",
			queryParams:  "?offset=0",
			key:          "offset",
			defaultValue: 10,
			expected:     0,
		},
		{
			name:         "negative value",
			queryParams:  "?offset=-5",
			key:          "offset",
			defaultValue: 0,
			expected:     -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test"+tt.queryParams, nil)
			result := getIntParam(req, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuditHandler_ExportCSV_MultipleEvents(t *testing.T) {
	mockService := new(MockAuditService)

	event1 := createTestAuditEvent()
	event1.ID = "event-1"
	event2 := createTestAuditEvent()
	event2.ID = "event-2"
	event2.ErrorMessage = "Some error"

	events := []audit.AuditEvent{*event1, *event2}
	mockService.On("QueryAuditEvents", mock.Anything, mock.Anything).Return(events, 2, nil)

	handler := NewAuditHandler(mockService, slog.New(slog.NewTextHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/audit/export", strings.NewReader(`{"format": "csv"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ExportEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	body := rr.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	assert.Len(t, lines, 3) // Header + 2 events
	assert.Contains(t, lines[1], "event-1")
	assert.Contains(t, lines[2], "event-2")
	assert.Contains(t, lines[2], "Some error")
}

func TestAuditHandler_QueryEvents_EmptyResults(t *testing.T) {
	mockService := new(MockAuditService)
	mockService.On("QueryAuditEvents", mock.Anything, mock.Anything).Return([]audit.AuditEvent{}, 0, nil)

	handler := NewAuditHandler(mockService, slog.New(slog.NewTextHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit/events", nil)
	rr := httptest.NewRecorder()

	handler.QueryEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["total"])
	events := response["events"].([]interface{})
	assert.Len(t, events, 0)
}

func TestAuditHandler_ExportJSON_EmptyResults(t *testing.T) {
	mockService := new(MockAuditService)
	mockService.On("QueryAuditEvents", mock.Anything, mock.Anything).Return([]audit.AuditEvent{}, 0, nil)

	handler := NewAuditHandler(mockService, slog.New(slog.NewTextHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/audit/export", strings.NewReader(`{"format": "json"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ExportEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var events []audit.AuditEvent
	err := json.Unmarshal(rr.Body.Bytes(), &events)
	assert.NoError(t, err)
	assert.Len(t, events, 0)
}

func TestAuditHandler_ExportCSV_EmptyResults(t *testing.T) {
	mockService := new(MockAuditService)
	mockService.On("QueryAuditEvents", mock.Anything, mock.Anything).Return([]audit.AuditEvent{}, 0, nil)

	handler := NewAuditHandler(mockService, slog.New(slog.NewTextHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/audit/export", strings.NewReader(`{"format": "csv"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ExportEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	body := rr.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	assert.Len(t, lines, 1) // Just header, no data
	assert.Contains(t, lines[0], "ID,Tenant ID")
}
