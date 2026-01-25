package handlers

import (
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
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/oauth"
)

// MockOAuthService is a mock implementation of oauth.OAuthService
type MockOAuthService struct {
	mock.Mock
}

func (m *MockOAuthService) GetProvider(ctx context.Context, providerKey string) (*oauth.OAuthProvider, error) {
	args := m.Called(ctx, providerKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth.OAuthProvider), args.Error(1)
}

func (m *MockOAuthService) ListProviders(ctx context.Context) ([]*oauth.OAuthProvider, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*oauth.OAuthProvider), args.Error(1)
}

func (m *MockOAuthService) Authorize(ctx context.Context, userID, tenantID string, input *oauth.AuthorizeInput) (string, error) {
	args := m.Called(ctx, userID, tenantID, input)
	return args.String(0), args.Error(1)
}

func (m *MockOAuthService) HandleCallback(ctx context.Context, userID, tenantID string, input *oauth.CallbackInput) (*oauth.OAuthConnection, error) {
	args := m.Called(ctx, userID, tenantID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth.OAuthConnection), args.Error(1)
}

func (m *MockOAuthService) GetConnection(ctx context.Context, userID, tenantID, connectionID string) (*oauth.OAuthConnection, error) {
	args := m.Called(ctx, userID, tenantID, connectionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth.OAuthConnection), args.Error(1)
}

func (m *MockOAuthService) ListConnections(ctx context.Context, userID, tenantID string) ([]*oauth.OAuthConnection, error) {
	args := m.Called(ctx, userID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*oauth.OAuthConnection), args.Error(1)
}

func (m *MockOAuthService) RevokeConnection(ctx context.Context, userID, tenantID, connectionID string) error {
	args := m.Called(ctx, userID, tenantID, connectionID)
	return args.Error(0)
}

func (m *MockOAuthService) RefreshToken(ctx context.Context, connectionID string) error {
	args := m.Called(ctx, connectionID)
	return args.Error(0)
}

func (m *MockOAuthService) TestConnection(ctx context.Context, connectionID string) error {
	args := m.Called(ctx, connectionID)
	return args.Error(0)
}

func (m *MockOAuthService) GetAccessToken(ctx context.Context, connectionID string) (string, error) {
	args := m.Called(ctx, connectionID)
	return args.String(0), args.Error(1)
}

// Helper function to create OAuth handler with mock service
func newTestOAuthHandler() (*OAuthHandler, *MockOAuthService) {
	mockService := new(MockOAuthService)
	handler := NewOAuthHandler(mockService)
	return handler, mockService
}

// Helper to add OAuth context (uses string keys like the handler does)
func addOAuthContext(req *http.Request, tenantID, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "user_id", userID)
	return req.WithContext(ctx)
}

// Helper to add Chi URL params for OAuth
func addOAuthChiURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// Test data helpers
func createTestOAuthProvider() *oauth.OAuthProvider {
	return &oauth.OAuthProvider{
		ID:            "provider-123",
		ProviderKey:   "github",
		Name:          "GitHub",
		Description:   "GitHub OAuth Provider",
		AuthURL:       "https://github.com/login/oauth/authorize",
		TokenURL:      "https://github.com/login/oauth/access_token",
		UserInfoURL:   "https://api.github.com/user",
		DefaultScopes: []string{"user:email", "read:user"},
		ClientID:      "client-id-123",
		Status:        oauth.ProviderStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func createTestOAuthConnection() *oauth.OAuthConnection {
	expiry := time.Now().Add(1 * time.Hour)
	return &oauth.OAuthConnection{
		ID:               "conn-123",
		UserID:           "user-123",
		TenantID:         "tenant-123",
		ProviderKey:      "github",
		ProviderUserID:   "gh-user-456",
		ProviderUsername: "testuser",
		ProviderEmail:    "test@example.com",
		TokenExpiry:      &expiry,
		Scopes:           []string{"user:email", "read:user"},
		Status:           oauth.ConnectionStatusActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// =============================================================================
// ListProviders Tests
// =============================================================================

func TestOAuthHandler_ListProviders(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockOAuthService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "success with providers",
			setupMock: func(m *MockOAuthService) {
				providers := []*oauth.OAuthProvider{
					createTestOAuthProvider(),
					{
						ID:          "provider-456",
						ProviderKey: "google",
						Name:        "Google",
						Status:      oauth.ProviderStatusActive,
					},
				}
				m.On("ListProviders", mock.Anything).Return(providers, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "success empty list",
			setupMock: func(m *MockOAuthService) {
				m.On("ListProviders", mock.Anything).Return([]*oauth.OAuthProvider{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "service error",
			setupMock: func(m *MockOAuthService) {
				m.On("ListProviders", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestOAuthHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/providers", nil)
			rr := httptest.NewRecorder()

			handler.ListProviders(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				var providers []*oauth.OAuthProvider
				err := json.NewDecoder(rr.Body).Decode(&providers)
				require.NoError(t, err)
				assert.Len(t, providers, tt.expectedCount)

				// Verify sensitive data is removed
				for _, p := range providers {
					assert.Nil(t, p.ClientSecretEncrypted)
					assert.Nil(t, p.ClientSecretNonce)
					assert.Nil(t, p.ClientSecretAuthTag)
					assert.Nil(t, p.ClientSecretEncDEK)
					assert.Empty(t, p.ClientSecretKMSKeyID)
				}
			}
			mockService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// Authorize Tests
// =============================================================================

func TestOAuthHandler_Authorize(t *testing.T) {
	tenantID := "tenant-123"
	userID := "user-123"

	tests := []struct {
		name             string
		provider         string
		queryParams      string
		acceptHeader     string
		setupMock        func(*MockOAuthService)
		expectedStatus   int
		expectedLocation string
		expectedJSON     bool
	}{
		{
			name:        "success with redirect",
			provider:    "github",
			queryParams: "?scopes=user:email&redirect_uri=https://app.example.com/callback",
			setupMock: func(m *MockOAuthService) {
				m.On("Authorize", mock.Anything, userID, tenantID, mock.AnythingOfType("*oauth.AuthorizeInput")).
					Return("https://github.com/login/oauth/authorize?client_id=xxx&state=yyy", nil)
			},
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://github.com/login/oauth/authorize?client_id=xxx&state=yyy",
		},
		{
			name:         "success with JSON response",
			provider:     "github",
			queryParams:  "",
			acceptHeader: "application/json",
			setupMock: func(m *MockOAuthService) {
				m.On("Authorize", mock.Anything, userID, tenantID, mock.AnythingOfType("*oauth.AuthorizeInput")).
					Return("https://github.com/login/oauth/authorize?client_id=xxx", nil)
			},
			expectedStatus: http.StatusOK,
			expectedJSON:   true,
		},
		{
			name:        "service error",
			provider:    "invalid",
			queryParams: "",
			setupMock: func(m *MockOAuthService) {
				m.On("Authorize", mock.Anything, userID, tenantID, mock.AnythingOfType("*oauth.AuthorizeInput")).
					Return("", oauth.ErrInvalidProvider)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestOAuthHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize/"+tt.provider+tt.queryParams, nil)
			req = addOAuthContext(req, tenantID, userID)
			req = addOAuthChiURLParam(req, "provider", tt.provider)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			rr := httptest.NewRecorder()

			handler.Authorize(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedLocation != "" {
				assert.Equal(t, tt.expectedLocation, rr.Header().Get("Location"))
			}
			if tt.expectedJSON {
				var resp map[string]string
				err := json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp["authorization_url"])
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestOAuthHandler_Authorize_MissingContext(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing user_id",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "tenant_id", "tenant-123")
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name: "missing tenant_id",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "user_id", "user-123")
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing tenant context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := newTestOAuthHandler()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/authorize/github", nil)
			req = tt.setupContext(req)
			req = addOAuthChiURLParam(req, "provider", "github")
			rr := httptest.NewRecorder()

			handler.Authorize(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedError)
		})
	}
}

// =============================================================================
// Callback Tests
// =============================================================================

func TestOAuthHandler_Callback(t *testing.T) {
	tenantID := "tenant-123"
	userID := "user-123"

	tests := []struct {
		name           string
		provider       string
		queryParams    string
		setupMock      func(*MockOAuthService)
		expectedStatus int
	}{
		{
			name:        "success",
			provider:    "github",
			queryParams: "?code=auth-code-123&state=state-abc",
			setupMock: func(m *MockOAuthService) {
				conn := createTestOAuthConnection()
				m.On("HandleCallback", mock.Anything, userID, tenantID, mock.MatchedBy(func(input *oauth.CallbackInput) bool {
					return input.Code == "auth-code-123" && input.State == "state-abc"
				})).Return(conn, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "callback with error from provider",
			provider:    "github",
			queryParams: "?error=access_denied&state=state-abc",
			setupMock: func(m *MockOAuthService) {
				m.On("HandleCallback", mock.Anything, userID, tenantID, mock.MatchedBy(func(input *oauth.CallbackInput) bool {
					return input.Error == "access_denied"
				})).Return(nil, errors.New("OAuth provider error: access_denied"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid state",
			provider:    "github",
			queryParams: "?code=auth-code&state=invalid-state",
			setupMock: func(m *MockOAuthService) {
				m.On("HandleCallback", mock.Anything, userID, tenantID, mock.Anything).
					Return(nil, oauth.ErrInvalidState)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestOAuthHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/callback/"+tt.provider+tt.queryParams, nil)
			req = addOAuthContext(req, tenantID, userID)
			req = addOAuthChiURLParam(req, "provider", tt.provider)
			rr := httptest.NewRecorder()

			handler.Callback(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				err := json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				assert.True(t, resp["success"].(bool))
				assert.Equal(t, tt.provider, resp["provider"])
				assert.NotNil(t, resp["connection"])

				// Verify sensitive data is removed from connection
				conn := resp["connection"].(map[string]interface{})
				assert.Nil(t, conn["access_token_encrypted"])
				assert.Nil(t, conn["refresh_token_encrypted"])
			}
			mockService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// ListConnections Tests
// =============================================================================

func TestOAuthHandler_ListConnections(t *testing.T) {
	tenantID := "tenant-123"
	userID := "user-123"

	tests := []struct {
		name           string
		setupMock      func(*MockOAuthService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "success with connections",
			setupMock: func(m *MockOAuthService) {
				connections := []*oauth.OAuthConnection{
					createTestOAuthConnection(),
					{
						ID:          "conn-456",
						UserID:      userID,
						TenantID:    tenantID,
						ProviderKey: "google",
						Status:      oauth.ConnectionStatusActive,
					},
				}
				m.On("ListConnections", mock.Anything, userID, tenantID).Return(connections, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "success empty list",
			setupMock: func(m *MockOAuthService) {
				m.On("ListConnections", mock.Anything, userID, tenantID).Return([]*oauth.OAuthConnection{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "service error",
			setupMock: func(m *MockOAuthService) {
				m.On("ListConnections", mock.Anything, userID, tenantID).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestOAuthHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/connections", nil)
			req = addOAuthContext(req, tenantID, userID)
			rr := httptest.NewRecorder()

			handler.ListConnections(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				var connections []*oauth.OAuthConnection
				err := json.NewDecoder(rr.Body).Decode(&connections)
				require.NoError(t, err)
				assert.Len(t, connections, tt.expectedCount)

				// Verify sensitive data is removed
				for _, c := range connections {
					assert.Nil(t, c.AccessTokenEncrypted)
					assert.Nil(t, c.RefreshTokenEncrypted)
					assert.Nil(t, c.RawTokenResponse)
				}
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestOAuthHandler_ListConnections_MissingContext(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing user_id",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "tenant_id", "tenant-123")
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name: "missing tenant_id",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "user_id", "user-123")
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing tenant context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := newTestOAuthHandler()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/connections", nil)
			req = tt.setupContext(req)
			rr := httptest.NewRecorder()

			handler.ListConnections(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedError)
		})
	}
}

// =============================================================================
// GetConnection Tests
// =============================================================================

func TestOAuthHandler_GetConnection(t *testing.T) {
	tenantID := "tenant-123"
	userID := "user-123"
	connectionID := "conn-123"

	tests := []struct {
		name           string
		connectionID   string
		setupMock      func(*MockOAuthService)
		expectedStatus int
	}{
		{
			name:         "success",
			connectionID: connectionID,
			setupMock: func(m *MockOAuthService) {
				conn := createTestOAuthConnection()
				m.On("GetConnection", mock.Anything, userID, tenantID, connectionID).Return(conn, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "connection not found",
			connectionID: "nonexistent",
			setupMock: func(m *MockOAuthService) {
				m.On("GetConnection", mock.Anything, userID, tenantID, "nonexistent").
					Return(nil, oauth.ErrConnectionNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestOAuthHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/connections/"+tt.connectionID, nil)
			req = addOAuthContext(req, tenantID, userID)
			req = addOAuthChiURLParam(req, "id", tt.connectionID)
			rr := httptest.NewRecorder()

			handler.GetConnection(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				var conn oauth.OAuthConnection
				err := json.NewDecoder(rr.Body).Decode(&conn)
				require.NoError(t, err)
				assert.Equal(t, connectionID, conn.ID)
				// Verify sensitive data is removed
				assert.Nil(t, conn.AccessTokenEncrypted)
				assert.Nil(t, conn.RefreshTokenEncrypted)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestOAuthHandler_GetConnection_MissingContext(t *testing.T) {
	handler, _ := newTestOAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/oauth/connections/conn-123", nil)
	req = addOAuthChiURLParam(req, "id", "conn-123")
	// No context added
	rr := httptest.NewRecorder()

	handler.GetConnection(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// =============================================================================
// RevokeConnection Tests
// =============================================================================

func TestOAuthHandler_RevokeConnection(t *testing.T) {
	tenantID := "tenant-123"
	userID := "user-123"
	connectionID := "conn-123"

	tests := []struct {
		name           string
		connectionID   string
		setupMock      func(*MockOAuthService)
		expectedStatus int
	}{
		{
			name:         "success",
			connectionID: connectionID,
			setupMock: func(m *MockOAuthService) {
				m.On("RevokeConnection", mock.Anything, userID, tenantID, connectionID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:         "connection not found",
			connectionID: "nonexistent",
			setupMock: func(m *MockOAuthService) {
				m.On("RevokeConnection", mock.Anything, userID, tenantID, "nonexistent").
					Return(oauth.ErrConnectionNotFound)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:         "unauthorized",
			connectionID: connectionID,
			setupMock: func(m *MockOAuthService) {
				m.On("RevokeConnection", mock.Anything, userID, tenantID, connectionID).
					Return(errors.New("unauthorized"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestOAuthHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/oauth/connections/"+tt.connectionID, nil)
			req = addOAuthContext(req, tenantID, userID)
			req = addOAuthChiURLParam(req, "id", tt.connectionID)
			rr := httptest.NewRecorder()

			handler.RevokeConnection(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestOAuthHandler_RevokeConnection_MissingContext(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing user_id",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "tenant_id", "tenant-123")
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name: "missing tenant_id",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "user_id", "user-123")
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing tenant context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := newTestOAuthHandler()

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/oauth/connections/conn-123", nil)
			req = tt.setupContext(req)
			req = addOAuthChiURLParam(req, "id", "conn-123")
			rr := httptest.NewRecorder()

			handler.RevokeConnection(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedError)
		})
	}
}

// =============================================================================
// TestConnection Tests
// =============================================================================

func TestOAuthHandler_TestConnection(t *testing.T) {
	tenantID := "tenant-123"
	userID := "user-123"
	connectionID := "conn-123"

	tests := []struct {
		name           string
		connectionID   string
		setupMock      func(*MockOAuthService)
		expectedStatus int
		expectedResult bool
	}{
		{
			name:         "success",
			connectionID: connectionID,
			setupMock: func(m *MockOAuthService) {
				m.On("TestConnection", mock.Anything, connectionID).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedResult: true,
		},
		{
			name:         "test failed - token expired",
			connectionID: connectionID,
			setupMock: func(m *MockOAuthService) {
				m.On("TestConnection", mock.Anything, connectionID).Return(oauth.ErrTokenExpired)
			},
			expectedStatus: http.StatusOK,
			expectedResult: false,
		},
		{
			name:         "test failed - connection error",
			connectionID: connectionID,
			setupMock: func(m *MockOAuthService) {
				m.On("TestConnection", mock.Anything, connectionID).Return(errors.New("API request failed"))
			},
			expectedStatus: http.StatusOK,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService := newTestOAuthHandler()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/connections/"+tt.connectionID+"/test", nil)
			req = addOAuthContext(req, tenantID, userID)
			req = addOAuthChiURLParam(req, "id", tt.connectionID)
			rr := httptest.NewRecorder()

			handler.TestConnection(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var resp map[string]interface{}
			err := json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, resp["success"])

			if !tt.expectedResult {
				assert.NotEmpty(t, resp["error"])
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestOAuthHandler_TestConnection_MissingContext(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*http.Request) *http.Request
		expectedStatus int
		expectedError  string
	}{
		{
			name: "missing user_id",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "tenant_id", "tenant-123")
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Unauthorized",
		},
		{
			name: "missing tenant_id",
			setupContext: func(req *http.Request) *http.Request {
				ctx := context.WithValue(req.Context(), "user_id", "user-123")
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing tenant context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _ := newTestOAuthHandler()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/oauth/connections/conn-123/test", nil)
			req = tt.setupContext(req)
			req = addOAuthChiURLParam(req, "id", "conn-123")
			rr := httptest.NewRecorder()

			handler.TestConnection(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedError)
		})
	}
}

// =============================================================================
// NewOAuthHandler Tests
// =============================================================================

func TestNewOAuthHandler(t *testing.T) {
	mockService := new(MockOAuthService)
	handler := NewOAuthHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.service)
}
