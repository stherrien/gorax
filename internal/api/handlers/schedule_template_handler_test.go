package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/schedule"
	"github.com/gorax/gorax/internal/tenant"
)

// addTestContext adds tenant and user context to the request for testing
func addTestContext(req *http.Request) *http.Request {
	t := &tenant.Tenant{
		ID:     "tenant-1",
		Status: "active",
	}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, t)

	user := &middleware.User{
		ID:       "user-1",
		TenantID: "tenant-1",
	}
	ctx = context.WithValue(ctx, middleware.UserContextKey, user)
	return req.WithContext(ctx)
}

// MockScheduleTemplateService is a mock implementation of ScheduleTemplateService
type MockScheduleTemplateService struct {
	mock.Mock
}

func (m *MockScheduleTemplateService) ListTemplates(ctx context.Context, filter schedule.ScheduleTemplateFilter) ([]*schedule.ScheduleTemplate, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*schedule.ScheduleTemplate), args.Error(1)
}

func (m *MockScheduleTemplateService) GetTemplate(ctx context.Context, id string) (*schedule.ScheduleTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schedule.ScheduleTemplate), args.Error(1)
}

func (m *MockScheduleTemplateService) ApplyTemplate(ctx context.Context, tenantID, userID, templateID string, input schedule.ApplyTemplateInput) (*schedule.Schedule, error) {
	args := m.Called(ctx, tenantID, userID, templateID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schedule.Schedule), args.Error(1)
}

func TestScheduleTemplateHandler_ListTemplates(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockReturn     []*schedule.ScheduleTemplate
		mockError      error
		expectedStatus int
		checkResponse  func(t *testing.T, templates []*schedule.ScheduleTemplate)
	}{
		{
			name:        "list all templates",
			queryParams: "",
			mockReturn: []*schedule.ScheduleTemplate{
				{
					ID:             "template-1",
					Name:           "Daily at 9 AM",
					Description:    "Runs every day at 9:00 AM",
					Category:       "daily",
					CronExpression: "0 9 * * *",
					Timezone:       "UTC",
					IsSystem:       true,
					CreatedAt:      time.Now(),
				},
				{
					ID:             "template-2",
					Name:           "SOC2 Daily Scan",
					Description:    "Daily SOC2 compliance check",
					Category:       "compliance",
					CronExpression: "0 2 * * *",
					Timezone:       "UTC",
					IsSystem:       true,
					CreatedAt:      time.Now(),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, templates []*schedule.ScheduleTemplate) {
				assert.Len(t, templates, 2)
			},
		},
		{
			name:        "filter by category",
			queryParams: "?category=compliance",
			mockReturn: []*schedule.ScheduleTemplate{
				{
					ID:             "template-2",
					Name:           "SOC2 Daily Scan",
					Category:       "compliance",
					CronExpression: "0 2 * * *",
					Timezone:       "UTC",
					IsSystem:       true,
					CreatedAt:      time.Now(),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, templates []*schedule.ScheduleTemplate) {
				assert.Len(t, templates, 1)
				assert.Equal(t, "compliance", templates[0].Category)
			},
		},
		{
			name:           "service error",
			queryParams:    "",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockScheduleTemplateService)
			handler := NewScheduleTemplateHandler(mockService, nil)

			mockService.On("ListTemplates", mock.Anything, mock.Anything).
				Return(tt.mockReturn, tt.mockError)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/schedule-templates"+tt.queryParams, nil)
			req = addTestContext(req)
			w := httptest.NewRecorder()

			handler.ListTemplates(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK && tt.checkResponse != nil {
				var templates []*schedule.ScheduleTemplate
				err := json.NewDecoder(w.Body).Decode(&templates)
				assert.NoError(t, err)
				tt.checkResponse(t, templates)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestScheduleTemplateHandler_GetTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateID     string
		mockReturn     *schedule.ScheduleTemplate
		mockError      error
		expectedStatus int
	}{
		{
			name:       "get existing template",
			templateID: "template-1",
			mockReturn: &schedule.ScheduleTemplate{
				ID:             "template-1",
				Name:           "Daily at 9 AM",
				Description:    "Runs every day at 9:00 AM",
				Category:       "daily",
				CronExpression: "0 9 * * *",
				Timezone:       "UTC",
				IsSystem:       true,
				CreatedAt:      time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "template not found",
			templateID:     "nonexistent",
			mockReturn:     nil,
			mockError:      errors.New("template not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "service error",
			templateID:     "template-1",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockScheduleTemplateService)
			handler := NewScheduleTemplateHandler(mockService, nil)

			mockService.On("GetTemplate", mock.Anything, tt.templateID).
				Return(tt.mockReturn, tt.mockError)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/schedule-templates/"+tt.templateID, nil)
			req = addTestContext(req)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.templateID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler.GetTemplate(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var template schedule.ScheduleTemplate
				err := json.NewDecoder(w.Body).Decode(&template)
				assert.NoError(t, err)
				assert.Equal(t, tt.templateID, template.ID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestScheduleTemplateHandler_ApplyTemplate(t *testing.T) {
	tests := []struct {
		name           string
		templateID     string
		requestBody    schedule.ApplyTemplateInput
		mockReturn     *schedule.Schedule
		mockError      error
		expectedStatus int
	}{
		{
			name:       "apply template successfully",
			templateID: "template-1",
			requestBody: schedule.ApplyTemplateInput{
				WorkflowID: "workflow-1",
			},
			mockReturn: &schedule.Schedule{
				ID:             "schedule-1",
				TenantID:       "tenant-1",
				WorkflowID:     "workflow-1",
				Name:           "Daily at 9 AM",
				CronExpression: "0 9 * * *",
				Timezone:       "UTC",
				Enabled:        true,
				CreatedBy:      "user-1",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:       "apply template with custom name",
			templateID: "template-1",
			requestBody: schedule.ApplyTemplateInput{
				WorkflowID: "workflow-1",
				Name:       stringPtr("My Custom Schedule"),
			},
			mockReturn: &schedule.Schedule{
				ID:             "schedule-1",
				TenantID:       "tenant-1",
				WorkflowID:     "workflow-1",
				Name:           "My Custom Schedule",
				CronExpression: "0 9 * * *",
				Timezone:       "UTC",
				Enabled:        true,
				CreatedBy:      "user-1",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:       "template not found",
			templateID: "nonexistent",
			requestBody: schedule.ApplyTemplateInput{
				WorkflowID: "workflow-1",
			},
			mockReturn:     nil,
			mockError:      errors.New("template not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "workflow not found",
			templateID: "template-1",
			requestBody: schedule.ApplyTemplateInput{
				WorkflowID: "nonexistent",
			},
			mockReturn:     nil,
			mockError:      errors.New("workflow not found"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "schedule already exists",
			templateID: "template-1",
			requestBody: schedule.ApplyTemplateInput{
				WorkflowID: "workflow-1",
			},
			mockReturn:     nil,
			mockError:      errors.New("schedule with name Daily at 9 AM already exists"),
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockScheduleTemplateService)
			handler := NewScheduleTemplateHandler(mockService, nil)

			mockService.On("ApplyTemplate", mock.Anything, "tenant-1", "user-1", tt.templateID, tt.requestBody).
				Return(tt.mockReturn, tt.mockError)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/schedule-templates/"+tt.templateID+"/apply", bytes.NewReader(body))
			req = addTestContext(req)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.templateID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler.ApplyTemplate(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var scheduleResp schedule.Schedule
				err := json.NewDecoder(w.Body).Decode(&scheduleResp)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody.WorkflowID, scheduleResp.WorkflowID)
			}

			mockService.AssertExpectations(t)
		})
	}
}
