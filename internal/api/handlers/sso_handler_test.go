package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/sso"
)

// MockSSOService is a mock implementation of sso.Service
type MockSSOService struct {
	mock.Mock
}

func (m *MockSSOService) CreateProvider(ctx context.Context, tenantID uuid.UUID, req *sso.CreateProviderRequest, createdBy uuid.UUID) (*sso.Provider, error) {
	args := m.Called(ctx, tenantID, req, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sso.Provider), args.Error(1)
}

func (m *MockSSOService) GetProvider(ctx context.Context, tenantID, providerID uuid.UUID) (*sso.Provider, error) {
	args := m.Called(ctx, tenantID, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sso.Provider), args.Error(1)
}

func (m *MockSSOService) ListProviders(ctx context.Context, tenantID uuid.UUID) ([]*sso.Provider, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*sso.Provider), args.Error(1)
}

func (m *MockSSOService) UpdateProvider(ctx context.Context, tenantID, providerID uuid.UUID, req *sso.UpdateProviderRequest, updatedBy uuid.UUID) (*sso.Provider, error) {
	args := m.Called(ctx, tenantID, providerID, req, updatedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sso.Provider), args.Error(1)
}

func (m *MockSSOService) DeleteProvider(ctx context.Context, tenantID, providerID uuid.UUID) error {
	args := m.Called(ctx, tenantID, providerID)
	return args.Error(0)
}

func (m *MockSSOService) GetProviderByDomain(ctx context.Context, emailDomain string) (*sso.Provider, error) {
	args := m.Called(ctx, emailDomain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sso.Provider), args.Error(1)
}

func (m *MockSSOService) InitiateLogin(ctx context.Context, providerID uuid.UUID, relayState string) (string, error) {
	args := m.Called(ctx, providerID, relayState)
	return args.String(0), args.Error(1)
}

func (m *MockSSOService) HandleCallback(ctx context.Context, providerID uuid.UUID, r *http.Request) (*sso.AuthenticationResponse, error) {
	args := m.Called(ctx, providerID, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sso.AuthenticationResponse), args.Error(1)
}

func (m *MockSSOService) GetMetadata(ctx context.Context, providerID uuid.UUID) (string, error) {
	args := m.Called(ctx, providerID)
	return args.String(0), args.Error(1)
}

func (m *MockSSOService) ValidateProvider(ctx context.Context, provider *sso.Provider) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

// Helper function to create SSO handler with mock service
func newTestSSOHandler() (*SSOHandler, *MockSSOService) {
	mockService := new(MockSSOService)
	handler := NewSSOHandler(mockService)
	return handler, mockService
}

// Helper to add SSO context (uses string keys like the handler does)
func addSSOContext(req *http.Request, tenantID, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "user_id", userID)
	return req.WithContext(ctx)
}

// Helper to add Chi URL params
func addChiURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// Test data helpers
func createTestProvider() *sso.Provider {
	tenantID := uuid.New()
	providerID := uuid.New()
	createdBy := uuid.New()
	return &sso.Provider{
		ID:         providerID,
		TenantID:   tenantID,
		Name:       "Test SAML Provider",
		Type:       sso.ProviderTypeSAML,
		Enabled:    true,
		EnforceSSO: false,
		Config:     json.RawMessage(`{"entity_id": "test-entity"}`),
		Domains:    []string{"example.com"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		CreatedBy:  &createdBy,
		UpdatedBy:  &createdBy,
	}
}

// =============================================================================
// CreateProvider Tests
// =============================================================================

func TestSSOHandler_CreateProvider(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockSSOService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "success",
			requestBody: sso.CreateProviderRequest{
				Name:       "Test Provider",
				Type:       sso.ProviderTypeSAML,
				Enabled:    true,
				EnforceSSO: false,
				Config:     json.RawMessage(`{"entity_id": "test"}`),
				Domains:    []string{"example.com"},
			},
			setupMock: func(m *MockSSOService) {
				provider := createTestProvider()
				m.On("CreateProvider", mock.Anything, tenantID, mock.AnythingOfType("*sso.CreateProviderRequest"), userID).
					Return(provider, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: "invalid json",
			setupMock: func(m *MockSSOService) {
				// No mock needed - will fail before service call
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "service error",
			requestBody: sso.CreateProviderRequest{
				Name:       "Test Provider",
				Type:       sso.ProviderTypeSAML,
				Enabled:    true,
				Config:     json.RawMessage(`{}`),
				Domains:    []string{"example.com"},
			},
			setupMock: func(m *MockSSOService) {
				m.On("CreateProvider", mock.Anything, tenantID, mock.AnythingOfType("*sso.CreateProviderRequest"), userID).
					Return(nil, errors.New("validation failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to create provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestSSOHandler()
			tt.setupMock(mockService)

			var body []byte
			switch v := tt.requestBody.(type) {
			case string:
				body = []byte(v)
			default:
				var err error
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/sso/providers", bytes.NewReader(body))
			req = addSSOContext(req, tenantID.String(), userID.String())
			rr := httptest.NewRecorder()

			handler.CreateProvider(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestSSOHandler_CreateProvider_MissingContext(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing tenant_id",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "user_id", uuid.New().String())
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name: "missing user_id",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "tenant_id", uuid.New().String())
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name: "invalid tenant_id format",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "tenant_id", "not-a-uuid")
				ctx = context.WithValue(ctx, "user_id", uuid.New().String())
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := newTestSSOHandler()

			body, _ := json.Marshal(sso.CreateProviderRequest{
				Name:    "Test",
				Type:    sso.ProviderTypeSAML,
				Config:  json.RawMessage(`{}`),
				Domains: []string{"example.com"},
			})

			req := httptest.NewRequest(http.MethodPost, "/api/v1/sso/providers", bytes.NewReader(body))
			req = tt.setupContext(req)
			rr := httptest.NewRecorder()

			handler.CreateProvider(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedError)
		})
	}
}

// =============================================================================
// ListProviders Tests
// =============================================================================

func TestSSOHandler_ListProviders(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name           string
		setupMock      func(*MockSSOService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "success with providers",
			setupMock: func(m *MockSSOService) {
				providers := []*sso.Provider{
					createTestProvider(),
					createTestProvider(),
				}
				m.On("ListProviders", mock.Anything, tenantID).Return(providers, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "success empty list",
			setupMock: func(m *MockSSOService) {
				m.On("ListProviders", mock.Anything, tenantID).Return([]*sso.Provider{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "service error",
			setupMock: func(m *MockSSOService) {
				m.On("ListProviders", mock.Anything, tenantID).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestSSOHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/sso/providers", nil)
			req = addSSOContext(req, tenantID.String(), uuid.New().String())
			rr := httptest.NewRecorder()

			handler.ListProviders(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				var providers []*sso.Provider
				err := json.NewDecoder(rr.Body).Decode(&providers)
				require.NoError(t, err)
				assert.Len(t, providers, tt.expectedCount)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestSSOHandler_ListProviders_MissingTenantID(t *testing.T) {
	handler, _ := newTestSSOHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sso/providers", nil)
	rr := httptest.NewRecorder()

	handler.ListProviders(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// =============================================================================
// GetProvider Tests
// =============================================================================

func TestSSOHandler_GetProvider(t *testing.T) {
	tenantID := uuid.New()
	providerID := uuid.New()

	tests := []struct {
		name           string
		providerID     string
		setupMock      func(*MockSSOService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:       "success",
			providerID: providerID.String(),
			setupMock: func(m *MockSSOService) {
				provider := createTestProvider()
				provider.ID = providerID
				m.On("GetProvider", mock.Anything, tenantID, providerID).Return(provider, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid provider ID",
			providerID:     "not-a-uuid",
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid provider ID",
		},
		{
			name:       "provider not found",
			providerID: providerID.String(),
			setupMock: func(m *MockSSOService) {
				m.On("GetProvider", mock.Anything, tenantID, providerID).Return(nil, errors.New("provider not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Failed to get provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestSSOHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/sso/providers/"+tt.providerID, nil)
			req = addSSOContext(req, tenantID.String(), uuid.New().String())
			req = addChiURLParam(req, "id", tt.providerID)
			rr := httptest.NewRecorder()

			handler.GetProvider(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// UpdateProvider Tests
// =============================================================================

func TestSSOHandler_UpdateProvider(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	providerID := uuid.New()

	tests := []struct {
		name           string
		providerID     string
		requestBody    interface{}
		setupMock      func(*MockSSOService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:       "success",
			providerID: providerID.String(),
			requestBody: sso.UpdateProviderRequest{
				Name: stringPtr("Updated Provider"),
			},
			setupMock: func(m *MockSSOService) {
				provider := createTestProvider()
				provider.ID = providerID
				provider.Name = "Updated Provider"
				m.On("UpdateProvider", mock.Anything, tenantID, providerID, mock.AnythingOfType("*sso.UpdateProviderRequest"), userID).
					Return(provider, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid provider ID",
			providerID:     "invalid",
			requestBody:    sso.UpdateProviderRequest{},
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid provider ID",
		},
		{
			name:           "invalid request body",
			providerID:     providerID.String(),
			requestBody:    "invalid json",
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:       "service error",
			providerID: providerID.String(),
			requestBody: sso.UpdateProviderRequest{
				Name: stringPtr("Updated"),
			},
			setupMock: func(m *MockSSOService) {
				m.On("UpdateProvider", mock.Anything, tenantID, providerID, mock.AnythingOfType("*sso.UpdateProviderRequest"), userID).
					Return(nil, errors.New("update failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to update provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestSSOHandler()
			tt.setupMock(mockService)

			var body []byte
			switch v := tt.requestBody.(type) {
			case string:
				body = []byte(v)
			default:
				var err error
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/sso/providers/"+tt.providerID, bytes.NewReader(body))
			req = addSSOContext(req, tenantID.String(), userID.String())
			req = addChiURLParam(req, "id", tt.providerID)
			rr := httptest.NewRecorder()

			handler.UpdateProvider(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// DeleteProvider Tests
// =============================================================================

func TestSSOHandler_DeleteProvider(t *testing.T) {
	tenantID := uuid.New()
	providerID := uuid.New()

	tests := []struct {
		name           string
		providerID     string
		setupMock      func(*MockSSOService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:       "success",
			providerID: providerID.String(),
			setupMock: func(m *MockSSOService) {
				m.On("DeleteProvider", mock.Anything, tenantID, providerID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid provider ID",
			providerID:     "invalid",
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid provider ID",
		},
		{
			name:       "service error",
			providerID: providerID.String(),
			setupMock: func(m *MockSSOService) {
				m.On("DeleteProvider", mock.Anything, tenantID, providerID).Return(errors.New("delete failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to delete provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestSSOHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/sso/providers/"+tt.providerID, nil)
			req = addSSOContext(req, tenantID.String(), uuid.New().String())
			req = addChiURLParam(req, "id", tt.providerID)
			rr := httptest.NewRecorder()

			handler.DeleteProvider(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// InitiateLogin Tests
// =============================================================================

func TestSSOHandler_InitiateLogin(t *testing.T) {
	providerID := uuid.New()

	tests := []struct {
		name             string
		providerID       string
		relayState       string
		setupMock        func(*MockSSOService)
		expectedStatus   int
		expectedLocation string
		expectedError    string
	}{
		{
			name:       "success",
			providerID: providerID.String(),
			relayState: "https://app.example.com/callback",
			setupMock: func(m *MockSSOService) {
				m.On("InitiateLogin", mock.Anything, providerID, "https://app.example.com/callback").
					Return("https://idp.example.com/sso?SAMLRequest=...", nil)
			},
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://idp.example.com/sso?SAMLRequest=...",
		},
		{
			name:       "success without relay state",
			providerID: providerID.String(),
			relayState: "",
			setupMock: func(m *MockSSOService) {
				m.On("InitiateLogin", mock.Anything, providerID, "").
					Return("https://idp.example.com/sso", nil)
			},
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://idp.example.com/sso",
		},
		{
			name:           "invalid provider ID",
			providerID:     "invalid",
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid provider ID",
		},
		{
			name:       "service error - provider disabled",
			providerID: providerID.String(),
			setupMock: func(m *MockSSOService) {
				m.On("InitiateLogin", mock.Anything, providerID, "").
					Return("", errors.New("SSO provider is disabled"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to initiate login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestSSOHandler()
			tt.setupMock(mockService)

			url := "/api/v1/sso/login/" + tt.providerID
			if tt.relayState != "" {
				url += "?relay_state=" + tt.relayState
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = addChiURLParam(req, "id", tt.providerID)
			rr := httptest.NewRecorder()

			handler.InitiateLogin(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedLocation != "" {
				assert.Equal(t, tt.expectedLocation, rr.Header().Get("Location"))
			}
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// HandleCallback Tests
// =============================================================================

func TestSSOHandler_HandleCallback(t *testing.T) {
	providerID := uuid.New()

	tests := []struct {
		name           string
		providerID     string
		method         string
		setupMock      func(*MockSSOService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:       "success POST",
			providerID: providerID.String(),
			method:     http.MethodPost,
			setupMock: func(m *MockSSOService) {
				authResp := &sso.AuthenticationResponse{
					UserAttributes: sso.UserAttributes{
						ExternalID: "ext-123",
						Email:      "user@example.com",
						FirstName:  "John",
						LastName:   "Doe",
					},
					SessionToken: "session-token-123",
					ExpiresAt:    time.Now().Add(24 * time.Hour),
				}
				m.On("HandleCallback", mock.Anything, providerID, mock.AnythingOfType("*http.Request")).
					Return(authResp, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "success GET (OIDC redirect)",
			providerID: providerID.String(),
			method:     http.MethodGet,
			setupMock: func(m *MockSSOService) {
				authResp := &sso.AuthenticationResponse{
					UserAttributes: sso.UserAttributes{
						ExternalID: "ext-456",
						Email:      "user@example.com",
					},
					SessionToken: "session-token-456",
					ExpiresAt:    time.Now().Add(24 * time.Hour),
				}
				m.On("HandleCallback", mock.Anything, providerID, mock.AnythingOfType("*http.Request")).
					Return(authResp, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid provider ID",
			providerID:     "invalid",
			method:         http.MethodPost,
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid provider ID",
		},
		{
			name:       "authentication failed",
			providerID: providerID.String(),
			method:     http.MethodPost,
			setupMock: func(m *MockSSOService) {
				m.On("HandleCallback", mock.Anything, providerID, mock.AnythingOfType("*http.Request")).
					Return(nil, errors.New("invalid SAML response"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "SSO authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestSSOHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(tt.method, "/api/v1/sso/callback/"+tt.providerID, nil)
			req = addChiURLParam(req, "id", tt.providerID)
			rr := httptest.NewRecorder()

			handler.HandleCallback(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
			if tt.expectedStatus == http.StatusOK {
				var resp sso.AuthenticationResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp.SessionToken)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// HandleSAMLAssertion Tests
// =============================================================================

func TestSSOHandler_HandleSAMLAssertion(t *testing.T) {
	providerID := uuid.New()

	tests := []struct {
		name           string
		formData       url.Values
		setupMock      func(*MockSSOService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "success",
			formData: url.Values{
				"RelayState":   {providerID.String()},
				"SAMLResponse": {"base64-encoded-saml-response"},
			},
			setupMock: func(m *MockSSOService) {
				authResp := &sso.AuthenticationResponse{
					UserAttributes: sso.UserAttributes{
						ExternalID: "ext-789",
						Email:      "saml-user@example.com",
					},
					SessionToken: "saml-session-token",
					ExpiresAt:    time.Now().Add(24 * time.Hour),
				}
				m.On("HandleCallback", mock.Anything, providerID, mock.AnythingOfType("*http.Request")).
					Return(authResp, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing relay state",
			formData: url.Values{
				"SAMLResponse": {"base64-encoded-saml-response"},
			},
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Relay state required",
		},
		{
			name: "invalid relay state (not UUID)",
			formData: url.Values{
				"RelayState":   {"not-a-uuid"},
				"SAMLResponse": {"base64-encoded-saml-response"},
			},
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid relay state",
		},
		{
			name: "authentication failed",
			formData: url.Values{
				"RelayState":   {providerID.String()},
				"SAMLResponse": {"invalid-response"},
			},
			setupMock: func(m *MockSSOService) {
				m.On("HandleCallback", mock.Anything, providerID, mock.AnythingOfType("*http.Request")).
					Return(nil, errors.New("SAML signature validation failed"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "SAML authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestSSOHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/sso/acs", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler.HandleSAMLAssertion(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// GetMetadata Tests
// =============================================================================

func TestSSOHandler_GetMetadata(t *testing.T) {
	providerID := uuid.New()

	tests := []struct {
		name           string
		providerID     string
		setupMock      func(*MockSSOService)
		expectedStatus int
		contentType    string
		expectedError  string
	}{
		{
			name:       "success",
			providerID: providerID.String(),
			setupMock: func(m *MockSSOService) {
				metadata := `<?xml version="1.0"?>
<EntityDescriptor entityID="https://sp.example.com">
  <SPSSODescriptor>
    <AssertionConsumerService Location="https://sp.example.com/acs"/>
  </SPSSODescriptor>
</EntityDescriptor>`
				m.On("GetMetadata", mock.Anything, providerID).Return(metadata, nil)
			},
			expectedStatus: http.StatusOK,
			contentType:    "application/xml",
		},
		{
			name:           "invalid provider ID",
			providerID:     "invalid",
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid provider ID",
		},
		{
			name:       "service error",
			providerID: providerID.String(),
			setupMock: func(m *MockSSOService) {
				m.On("GetMetadata", mock.Anything, providerID).Return("", errors.New("failed to generate metadata"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to get metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestSSOHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/sso/metadata/"+tt.providerID, nil)
			req = addChiURLParam(req, "id", tt.providerID)
			rr := httptest.NewRecorder()

			handler.GetMetadata(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.contentType != "" {
				assert.Equal(t, tt.contentType, rr.Header().Get("Content-Type"))
			}
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
			mockService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// DiscoverProvider Tests
// =============================================================================

func TestSSOHandler_DiscoverProvider(t *testing.T) {
	providerID := uuid.New()

	tests := []struct {
		name           string
		email          string
		setupMock      func(*MockSSOService)
		expectedStatus int
		ssoAvailable   bool
		expectedError  string
	}{
		{
			name:  "success - SSO available",
			email: "user@example.com",
			setupMock: func(m *MockSSOService) {
				provider := createTestProvider()
				provider.ID = providerID
				provider.Name = "Example Corp SSO"
				provider.Type = sso.ProviderTypeSAML
				provider.EnforceSSO = true
				m.On("GetProviderByDomain", mock.Anything, "example.com").Return(provider, nil)
			},
			expectedStatus: http.StatusOK,
			ssoAvailable:   true,
		},
		{
			name:  "success - SSO not available",
			email: "user@noproviderdomain.com",
			setupMock: func(m *MockSSOService) {
				m.On("GetProviderByDomain", mock.Anything, "noproviderdomain.com").
					Return(nil, errors.New("provider not found"))
			},
			expectedStatus: http.StatusNotFound,
			ssoAvailable:   false,
		},
		{
			name:           "missing email parameter",
			email:          "",
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email parameter required",
		},
		{
			name:           "invalid email format - no @",
			email:          "invalidemail",
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid email format",
		},
		{
			name:           "invalid email format - multiple @",
			email:          "user@@example.com",
			setupMock:      func(m *MockSSOService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestSSOHandler()
			tt.setupMock(mockService)

			url := "/api/v1/sso/discover"
			if tt.email != "" {
				url += "?email=" + tt.email
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			handler.DiscoverProvider(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusNotFound {
				var resp map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, tt.ssoAvailable, resp["sso_available"])
				if tt.ssoAvailable {
					assert.NotEmpty(t, resp["provider_id"])
					assert.NotEmpty(t, resp["provider_name"])
					assert.NotEmpty(t, resp["provider_type"])
				}
			}
			mockService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// RegisterRoutes Tests
// =============================================================================

func TestSSOHandler_RegisterRoutes(t *testing.T) {
	handler, _ := newTestSSOHandler()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	// Verify routes are registered
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/sso/providers"},
		{http.MethodGet, "/sso/providers"},
		{http.MethodGet, "/sso/providers/{id}"},
		{http.MethodPut, "/sso/providers/{id}"},
		{http.MethodDelete, "/sso/providers/{id}"},
		{http.MethodGet, "/sso/login/{id}"},
		{http.MethodPost, "/sso/callback/{id}"},
		{http.MethodGet, "/sso/callback/{id}"},
		{http.MethodGet, "/sso/metadata/{id}"},
		{http.MethodPost, "/sso/acs"},
		{http.MethodGet, "/sso/discover"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			// Walk routes to verify they exist
			found := false
			walkErr := chi.Walk(r, func(method string, routePattern string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
				if method == route.method && routePattern == route.path {
					found = true
				}
				return nil
			})
			require.NoError(t, walkErr)
			assert.True(t, found, "Route %s %s not found", route.method, route.path)
		})
	}
}

// =============================================================================
// Context Helper Functions Tests
// =============================================================================

func TestGetTenantIDFromContext(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expectError bool
	}{
		{
			name: "valid tenant ID",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), "tenant_id", uuid.New().String())
			},
			expectError: false,
		},
		{
			name: "missing tenant ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectError: true,
		},
		{
			name: "invalid tenant ID format",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), "tenant_id", "not-a-uuid")
			},
			expectError: true,
		},
		{
			name: "wrong type in context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), "tenant_id", 12345)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			tenantID, err := getTenantIDFromContext(ctx)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, uuid.Nil, tenantID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tenantID)
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expectError bool
	}{
		{
			name: "valid user ID",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), "user_id", uuid.New().String())
			},
			expectError: false,
		},
		{
			name: "missing user ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectError: true,
		},
		{
			name: "invalid user ID format",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), "user_id", "not-a-uuid")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			userID, err := getUserIDFromContext(ctx)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, uuid.Nil, userID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, userID)
			}
		})
	}
}

// stringPtr is defined in webhook.go
