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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/template"
	"github.com/gorax/gorax/internal/tenant"
)

// MockTemplateService is a mock implementation of template service for testing
type MockTemplateService struct {
	mock.Mock
}

func (m *MockTemplateService) CreateTemplate(ctx context.Context, tenantID, userID string, input template.CreateTemplateInput) (*template.Template, error) {
	args := m.Called(ctx, tenantID, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*template.Template), args.Error(1)
}

func (m *MockTemplateService) GetTemplate(ctx context.Context, tenantID, id string) (*template.Template, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*template.Template), args.Error(1)
}

func (m *MockTemplateService) ListTemplates(ctx context.Context, tenantID string, filter template.TemplateFilter) ([]*template.Template, error) {
	args := m.Called(ctx, tenantID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*template.Template), args.Error(1)
}

func (m *MockTemplateService) UpdateTemplate(ctx context.Context, tenantID, id string, input template.UpdateTemplateInput) error {
	args := m.Called(ctx, tenantID, id, input)
	return args.Error(0)
}

func (m *MockTemplateService) DeleteTemplate(ctx context.Context, tenantID, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockTemplateService) CreateFromWorkflow(ctx context.Context, tenantID, userID string, input template.CreateTemplateFromWorkflowInput) (*template.Template, error) {
	args := m.Called(ctx, tenantID, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*template.Template), args.Error(1)
}

func (m *MockTemplateService) InstantiateTemplate(ctx context.Context, tenantID, templateID string, input template.InstantiateTemplateInput) (*template.InstantiateTemplateResult, error) {
	args := m.Called(ctx, tenantID, templateID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*template.InstantiateTemplateResult), args.Error(1)
}

func newTestTemplateHandler() (*TemplateHandler, *MockTemplateService) {
	mockService := new(MockTemplateService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewTemplateHandler(mockService, logger)
	return handler, mockService
}

// addTemplateTestContext adds both tenant and user context to the request for testing
func addTemplateTestContext(req *http.Request, tenantID, userID string) *http.Request {
	t := &tenant.Tenant{
		ID:     tenantID,
		Status: "active",
	}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, t)

	user := &middleware.User{
		ID:       userID,
		TenantID: tenantID,
	}
	ctx = context.WithValue(ctx, middleware.UserContextKey, user)
	return req.WithContext(ctx)
}

func TestListTemplates(t *testing.T) {
	handler, mockService := newTestTemplateHandler()

	tenantID := "test-tenant-123"
	templates := []*template.Template{
		{
			ID:          "template-1",
			Name:        "Template 1",
			Category:    "security",
			Description: "Test template",
			CreatedAt:   time.Now(),
		},
	}

	mockService.On("ListTemplates", mock.Anything, tenantID, mock.AnythingOfType("template.TemplateFilter")).
		Return(templates, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
	req = addTemplateTestContext(req, tenantID, "user-123")
	w := httptest.NewRecorder()

	handler.ListTemplates(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []*template.Template
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "Template 1", response[0].Name)

	mockService.AssertExpectations(t)
}

func TestListTemplates_WithCategoryFilter(t *testing.T) {
	handler, mockService := newTestTemplateHandler()

	tenantID := "test-tenant-123"
	templates := []*template.Template{
		{
			ID:       "template-1",
			Name:     "Security Template",
			Category: "security",
		},
	}

	expectedFilter := template.TemplateFilter{Category: "security"}
	mockService.On("ListTemplates", mock.Anything, tenantID, expectedFilter).
		Return(templates, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates?category=security", nil)
	req = addTemplateTestContext(req, tenantID, "user-123")
	w := httptest.NewRecorder()

	handler.ListTemplates(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetTemplate(t *testing.T) {
	handler, mockService := newTestTemplateHandler()

	tenantID := "test-tenant-123"
	templateID := "template-123"
	tmpl := &template.Template{
		ID:          templateID,
		Name:        "Test Template",
		Category:    "security",
		Description: "Test description",
	}

	mockService.On("GetTemplate", mock.Anything, tenantID, templateID).
		Return(tmpl, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+templateID, nil)
	req = addTemplateTestContext(req, tenantID, "user-123")
	req = addRouteParam(req, "id", templateID)
	w := httptest.NewRecorder()

	handler.GetTemplate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response template.Template
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Test Template", response.Name)

	mockService.AssertExpectations(t)
}

func TestGetTemplate_NotFound(t *testing.T) {
	handler, mockService := newTestTemplateHandler()

	tenantID := "test-tenant-123"
	templateID := "nonexistent"

	mockService.On("GetTemplate", mock.Anything, tenantID, templateID).
		Return(nil, errors.New("template not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+templateID, nil)
	req = addTemplateTestContext(req, tenantID, "user-123")
	req = addRouteParam(req, "id", templateID)
	w := httptest.NewRecorder()

	handler.GetTemplate(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateTemplate(t *testing.T) {
	handler, mockService := newTestTemplateHandler()

	tenantID := "test-tenant-123"
	userID := "user-123"

	input := template.CreateTemplateInput{
		Name:        "New Template",
		Description: "Test description",
		Category:    "security",
		Definition:  json.RawMessage(`{"nodes":[],"edges":[]}`),
		Tags:        []string{"test"},
	}

	createdTemplate := &template.Template{
		ID:          "template-123",
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		Definition:  input.Definition,
		Tags:        input.Tags,
		CreatedAt:   time.Now(),
	}

	mockService.On("CreateTemplate", mock.Anything, tenantID, userID, input).
		Return(createdTemplate, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(body))
	req = addTemplateTestContext(req, tenantID, "user-123")
	req = req.WithContext(context.WithValue(req.Context(), "userID", userID))
	w := httptest.NewRecorder()

	handler.CreateTemplate(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response template.Template
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "New Template", response.Name)

	mockService.AssertExpectations(t)
}

func TestCreateTemplate_InvalidInput(t *testing.T) {
	handler, _ := newTestTemplateHandler()

	tenantID := "test-tenant-123"
	userID := "user-123"

	invalidInput := map[string]interface{}{
		"name": "",
	}

	body, _ := json.Marshal(invalidInput)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(body))
	req = addTemplateTestContext(req, tenantID, "user-123")
	req = req.WithContext(context.WithValue(req.Context(), "userID", userID))
	w := httptest.NewRecorder()

	handler.CreateTemplate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateTemplate(t *testing.T) {
	handler, mockService := newTestTemplateHandler()

	tenantID := "test-tenant-123"
	templateID := "template-123"

	input := template.UpdateTemplateInput{
		Name:        "Updated Name",
		Description: "Updated description",
	}

	mockService.On("UpdateTemplate", mock.Anything, tenantID, templateID, input).
		Return(nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/"+templateID, bytes.NewReader(body))
	req = addTemplateTestContext(req, tenantID, "user-123")
	req = addRouteParam(req, "id", templateID)
	w := httptest.NewRecorder()

	handler.UpdateTemplate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteTemplate(t *testing.T) {
	handler, mockService := newTestTemplateHandler()

	tenantID := "test-tenant-123"
	templateID := "template-123"

	mockService.On("DeleteTemplate", mock.Anything, tenantID, templateID).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/"+templateID, nil)
	req = addTemplateTestContext(req, tenantID, "user-123")
	req = addRouteParam(req, "id", templateID)
	w := httptest.NewRecorder()

	handler.DeleteTemplate(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateFromWorkflow(t *testing.T) {
	handler, mockService := newTestTemplateHandler()

	tenantID := "test-tenant-123"
	userID := "user-123"
	workflowID := "workflow-123"

	input := template.CreateTemplateFromWorkflowInput{
		WorkflowID: workflowID,
		Name:       "From Workflow",
		Category:   "integration",
		Definition: json.RawMessage(`{"nodes":[],"edges":[]}`),
	}

	createdTemplate := &template.Template{
		ID:       "template-123",
		Name:     input.Name,
		Category: input.Category,
	}

	mockService.On("CreateFromWorkflow", mock.Anything, tenantID, userID, input).
		Return(createdTemplate, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/from-workflow/"+workflowID, bytes.NewReader(body))
	req = addTemplateTestContext(req, tenantID, "user-123")
	req = req.WithContext(context.WithValue(req.Context(), "userID", userID))
	req = addRouteParam(req, "workflowId", workflowID)
	w := httptest.NewRecorder()

	handler.CreateFromWorkflow(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestInstantiateTemplate(t *testing.T) {
	handler, mockService := newTestTemplateHandler()

	tenantID := "test-tenant-123"
	templateID := "template-123"

	input := template.InstantiateTemplateInput{
		WorkflowName: "New Workflow",
	}

	result := &template.InstantiateTemplateResult{
		WorkflowName: "New Workflow",
		Definition:   json.RawMessage(`{"nodes":[],"edges":[]}`),
	}

	mockService.On("InstantiateTemplate", mock.Anything, tenantID, templateID, input).
		Return(result, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID+"/instantiate", bytes.NewReader(body))
	req = addTemplateTestContext(req, tenantID, "user-123")
	req = addRouteParam(req, "id", templateID)
	w := httptest.NewRecorder()

	handler.InstantiateTemplate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response template.InstantiateTemplateResult
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "New Workflow", response.WorkflowName)

	mockService.AssertExpectations(t)
}
