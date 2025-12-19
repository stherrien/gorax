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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/tenant"
)

// MockCredentialService is a mock implementation of credential service for testing
type MockCredentialService struct {
	mock.Mock
}

func (m *MockCredentialService) Create(ctx context.Context, tenantID, userID string, input credential.CreateCredentialInput) (*credential.Credential, error) {
	args := m.Called(ctx, tenantID, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*credential.Credential), args.Error(1)
}

func (m *MockCredentialService) List(ctx context.Context, tenantID string, filter credential.CredentialListFilter, limit, offset int) ([]*credential.Credential, error) {
	args := m.Called(ctx, tenantID, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*credential.Credential), args.Error(1)
}

func (m *MockCredentialService) GetByID(ctx context.Context, tenantID, credentialID string) (*credential.Credential, error) {
	args := m.Called(ctx, tenantID, credentialID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*credential.Credential), args.Error(1)
}

func (m *MockCredentialService) GetValue(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
	args := m.Called(ctx, tenantID, credentialID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*credential.DecryptedValue), args.Error(1)
}

func (m *MockCredentialService) Update(ctx context.Context, tenantID, credentialID, userID string, input credential.UpdateCredentialInput) (*credential.Credential, error) {
	args := m.Called(ctx, tenantID, credentialID, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*credential.Credential), args.Error(1)
}

func (m *MockCredentialService) Delete(ctx context.Context, tenantID, credentialID, userID string) error {
	args := m.Called(ctx, tenantID, credentialID, userID)
	return args.Error(0)
}

func (m *MockCredentialService) Rotate(ctx context.Context, tenantID, credentialID, userID string, input credential.RotateCredentialInput) (*credential.Credential, error) {
	args := m.Called(ctx, tenantID, credentialID, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*credential.Credential), args.Error(1)
}

func (m *MockCredentialService) ListVersions(ctx context.Context, tenantID, credentialID string) ([]*credential.CredentialValue, error) {
	args := m.Called(ctx, tenantID, credentialID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*credential.CredentialValue), args.Error(1)
}

func (m *MockCredentialService) GetAccessLog(ctx context.Context, tenantID, credentialID string, limit, offset int) ([]*credential.AccessLog, error) {
	args := m.Called(ctx, tenantID, credentialID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*credential.AccessLog), args.Error(1)
}

func newTestCredentialHandler() (*CredentialHandler, *MockCredentialService) {
	mockService := new(MockCredentialService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := NewCredentialHandler(mockService, logger)
	return handler, mockService
}

// addUserContext adds both tenant and user context to the request for testing
func addUserContext(req *http.Request, tenantID, userID string) *http.Request {
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

// TestCreate_Success tests successful credential creation
func TestCreate_Success(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	now := time.Now()
	input := credential.CreateCredentialInput{
		Name:        "My API Key",
		Description: "Test API key",
		Type:        credential.TypeAPIKey,
		Value: map[string]interface{}{
			"api_key": "secret-key-123",
		},
		Metadata: map[string]interface{}{
			"env": "production",
		},
	}

	expectedCred := &credential.Credential{
		ID:          "cred-123",
		TenantID:    "tenant-123",
		Name:        "My API Key",
		Description: "Test API key",
		Type:        credential.TypeAPIKey,
		Status:      credential.StatusActive,
		CreatedBy:   "user-123",
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata: map[string]interface{}{
			"env": "production",
		},
	}

	mockService.On("Create", mock.Anything, "tenant-123", "user-123", input).Return(expectedCred, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials", bytes.NewReader(body))
	req = addUserContext(req, "tenant-123", "user-123")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "cred-123", data["id"])
	assert.Equal(t, "My API Key", data["name"])
	assert.NotContains(t, data, "value") // Should not return value

	mockService.AssertExpectations(t)
}

// TestCreate_ValidationError tests credential creation with invalid input
func TestCreate_ValidationError(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	input := credential.CreateCredentialInput{
		Name: "", // Invalid: empty name
		Type: credential.TypeAPIKey,
	}

	mockService.On("Create", mock.Anything, "tenant-123", "user-123", input).
		Return(nil, &credential.ValidationError{Message: "name is required"})

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials", bytes.NewReader(body))
	req = addUserContext(req, "tenant-123", "user-123")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "name is required")

	mockService.AssertExpectations(t)
}

// TestCreate_InvalidJSON tests credential creation with invalid JSON
func TestCreate_InvalidJSON(t *testing.T) {
	handler, _ := newTestCredentialHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials", bytes.NewReader([]byte("invalid json")))
	req = addUserContext(req, "tenant-123", "user-123")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestList_Success tests successful credential listing
func TestList_Success(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	now := time.Now()
	credentials := []*credential.Credential{
		{
			ID:          "cred-1",
			TenantID:    "tenant-123",
			Name:        "API Key 1",
			Type:        credential.TypeAPIKey,
			Status:      credential.StatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "cred-2",
			TenantID:    "tenant-123",
			Name:        "OAuth Token",
			Type:        credential.TypeOAuth2,
			Status:      credential.StatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	mockService.On("List", mock.Anything, "tenant-123", credential.CredentialListFilter{}, 0, 0).
		Return(credentials, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials", nil)
	req = addUserContext(req, "tenant-123", "user-123")
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	mockService.AssertExpectations(t)
}

// TestList_WithFilters tests credential listing with filters
func TestList_WithFilters(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	mockService.On("List", mock.Anything, "tenant-123",
		credential.CredentialListFilter{Type: credential.TypeAPIKey}, 10, 0).
		Return([]*credential.Credential{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials?type=api_key&limit=10", nil)
	req = addUserContext(req, "tenant-123", "user-123")
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestGet_Success tests successful credential retrieval
func TestGet_Success(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	now := time.Now()
	expectedCred := &credential.Credential{
		ID:          "cred-123",
		TenantID:    "tenant-123",
		Name:        "My API Key",
		Type:        credential.TypeAPIKey,
		Status:      credential.StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockService.On("GetByID", mock.Anything, "tenant-123", "cred-123").Return(expectedCred, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials/cred-123", nil)
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "cred-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "cred-123", data["id"])

	mockService.AssertExpectations(t)
}

// TestGet_NotFound tests credential not found
func TestGet_NotFound(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	mockService.On("GetByID", mock.Anything, "tenant-123", "non-existent").
		Return(nil, credential.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials/non-existent", nil)
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "non-existent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestGetValue_Success tests successful credential value retrieval
func TestGetValue_Success(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	now := time.Now()
	expectedValue := &credential.DecryptedValue{
		Version: 1,
		Value: map[string]interface{}{
			"api_key": "secret-key-123",
		},
		CreatedAt: now,
	}

	mockService.On("GetValue", mock.Anything, "tenant-123", "cred-123", "user-123").
		Return(expectedValue, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials/cred-123/value", nil)
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "cred-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetValue(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "value")
	assert.Equal(t, float64(1), data["version"])

	mockService.AssertExpectations(t)
}

// TestGetValue_Unauthorized tests unauthorized access to credential value
func TestGetValue_Unauthorized(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	mockService.On("GetValue", mock.Anything, "tenant-123", "cred-123", "user-123").
		Return(nil, credential.ErrUnauthorized)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials/cred-123/value", nil)
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "cred-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetValue(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockService.AssertExpectations(t)
}

// TestUpdate_Success tests successful credential update
func TestUpdate_Success(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	now := time.Now()
	newName := "Updated Name"
	newStatus := credential.StatusInactive
	input := credential.UpdateCredentialInput{
		Name:   &newName,
		Status: &newStatus,
	}

	updatedCred := &credential.Credential{
		ID:        "cred-123",
		TenantID:  "tenant-123",
		Name:      "Updated Name",
		Type:      credential.TypeAPIKey,
		Status:    credential.StatusInactive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService.On("Update", mock.Anything, "tenant-123", "cred-123", "user-123", input).
		Return(updatedCred, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/credentials/cred-123", bytes.NewReader(body))
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "cred-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "Updated Name", data["name"])
	assert.Equal(t, "inactive", data["status"])

	mockService.AssertExpectations(t)
}

// TestUpdate_ValidationError tests update with invalid input
func TestUpdate_ValidationError(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	invalidStatus := credential.CredentialStatus("invalid")
	input := credential.UpdateCredentialInput{
		Status: &invalidStatus,
	}

	mockService.On("Update", mock.Anything, "tenant-123", "cred-123", "user-123", input).
		Return(nil, &credential.ValidationError{Message: "invalid status"})

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/credentials/cred-123", bytes.NewReader(body))
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "cred-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

// TestDelete_Success tests successful credential deletion
func TestDelete_Success(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	mockService.On("Delete", mock.Anything, "tenant-123", "cred-123", "user-123").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/credentials/cred-123", nil)
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "cred-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

// TestDelete_NotFound tests deletion of non-existent credential
func TestDelete_NotFound(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	mockService.On("Delete", mock.Anything, "tenant-123", "non-existent", "user-123").
		Return(credential.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/credentials/non-existent", nil)
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "non-existent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestRotate_Success tests successful credential rotation
func TestRotate_Success(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	now := time.Now()
	input := credential.RotateCredentialInput{
		Value: map[string]interface{}{
			"api_key": "new-secret-key-456",
		},
	}

	rotatedCred := &credential.Credential{
		ID:        "cred-123",
		TenantID:  "tenant-123",
		Name:      "My API Key",
		Type:      credential.TypeAPIKey,
		Status:    credential.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockService.On("Rotate", mock.Anything, "tenant-123", "cred-123", "user-123", input).
		Return(rotatedCred, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials/cred-123/rotate", bytes.NewReader(body))
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "cred-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Rotate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	mockService.AssertExpectations(t)
}

// TestRotate_ValidationError tests rotation with invalid input
func TestRotate_ValidationError(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	input := credential.RotateCredentialInput{
		Value: nil, // Invalid
	}

	mockService.On("Rotate", mock.Anything, "tenant-123", "cred-123", "user-123", input).
		Return(nil, &credential.ValidationError{Message: "value is required"})

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials/cred-123/rotate", bytes.NewReader(body))
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "cred-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Rotate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

// TestListVersions_Success tests successful version listing
func TestListVersions_Success(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	now := time.Now()
	versions := []*credential.CredentialValue{
		{
			ID:           "ver-2",
			CredentialID: "cred-123",
			Version:      2,
			CreatedAt:    now,
			CreatedBy:    "user-123",
			IsActive:     true,
		},
		{
			ID:           "ver-1",
			CredentialID: "cred-123",
			Version:      1,
			CreatedAt:    now.Add(-24 * time.Hour),
			CreatedBy:    "user-123",
			IsActive:     false,
		},
	}

	mockService.On("ListVersions", mock.Anything, "tenant-123", "cred-123").
		Return(versions, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials/cred-123/versions", nil)
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "cred-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.ListVersions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	mockService.AssertExpectations(t)
}

// TestGetAccessLog_Success tests successful access log retrieval
func TestGetAccessLog_Success(t *testing.T) {
	handler, mockService := newTestCredentialHandler()

	now := time.Now()
	logs := []*credential.AccessLog{
		{
			ID:           "log-1",
			CredentialID: "cred-123",
			TenantID:     "tenant-123",
			AccessedBy:   "user-123",
			AccessType:   "read",
			AccessedAt:   now,
			Success:      true,
		},
		{
			ID:           "log-2",
			CredentialID: "cred-123",
			TenantID:     "tenant-123",
			AccessedBy:   "user-456",
			AccessType:   "rotate",
			AccessedAt:   now.Add(-1 * time.Hour),
			Success:      true,
		},
	}

	mockService.On("GetAccessLog", mock.Anything, "tenant-123", "cred-123", 0, 0).
		Return(logs, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/credentials/cred-123/access-log", nil)
	req = addUserContext(req, "tenant-123", "user-123")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("credentialID", "cred-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetAccessLog(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	mockService.AssertExpectations(t)
}

// TestMissingTenantOrUser tests handlers when tenant or user context is missing
func TestMissingTenantOrUser(t *testing.T) {
	handler, _ := newTestCredentialHandler()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		url     string
		method  string
	}{
		{
			name:    "Create without context",
			handler: handler.Create,
			url:     "/api/v1/credentials",
			method:  http.MethodPost,
		},
		{
			name:    "List without context",
			handler: handler.List,
			url:     "/api/v1/credentials",
			method:  http.MethodGet,
		},
		{
			name:    "GetValue without context",
			handler: handler.GetValue,
			url:     "/api/v1/credentials/cred-123/value",
			method:  http.MethodGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	}
}
