package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/gorax/gorax/internal/config"
	"github.com/gorax/gorax/internal/user"
)

// MockUserRepository implements user.Repository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	if u.ID == "" {
		u.ID = "user-generated-id"
	}
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByKratosIdentityID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) ListByTenant(ctx context.Context, tenantID string) ([]*user.User, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id string, input user.UpdateUserInput) (*user.User, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Helper to create a mock Kratos server
func newMockKratosServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// Helper to create an auth handler with mock dependencies
func newTestAuthHandler(kratosURL string, mockRepo *MockUserRepository) *AuthHandler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	userService := user.NewService(mockRepo, logger)
	kratosConfig := config.KratosConfig{
		PublicURL: kratosURL,
	}
	return NewAuthHandler(userService, kratosConfig, logger)
}

func TestAuthHandler_InitiateRegistration(t *testing.T) {
	tests := []struct {
		name           string
		kratosHandler  http.HandlerFunc
		expectedStatus int
	}{
		{
			name: "success",
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/self-service/registration/api", r.URL.Path)
				w.Header().Set("X-Flow-Id", "flow-123")
				w.Header().Set("Location", "/registration/ui")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"id":     "flow-123",
					"method": "password",
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "kratos error",
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKratos := newMockKratosServer(tt.kratosHandler)
			defer mockKratos.Close()

			mockRepo := new(MockUserRepository)
			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/auth/registration", nil)
			w := httptest.NewRecorder()

			handler.InitiateRegistration(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		flowID         string
		requestBody    RegisterRequest
		kratosHandler  http.HandlerFunc
		expectedStatus int
	}{
		{
			name:   "success",
			flowID: "flow-123",
			requestBody: RegisterRequest{
				Email:    "test@example.com",
				Password: "SecurePassword123!",
				TenantID: "tenant-123",
			},
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/self-service/registration", r.URL.Path)
				assert.Equal(t, "flow-123", r.URL.Query().Get("flow"))
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"identity": map[string]any{
						"id": "identity-123",
					},
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "missing flow parameter",
			flowID: "",
			requestBody: RegisterRequest{
				Email:    "test@example.com",
				Password: "SecurePassword123!",
			},
			kratosHandler:  nil, // Not called
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockKratos *httptest.Server
			if tt.kratosHandler != nil {
				mockKratos = newMockKratosServer(tt.kratosHandler)
				defer mockKratos.Close()
			} else {
				mockKratos = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Fatal("kratos should not be called")
				}))
				defer mockKratos.Close()
			}

			mockRepo := new(MockUserRepository)
			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			url := "/api/auth/registration"
			if tt.flowID != "" {
				url = fmt.Sprintf("/api/auth/registration?flow=%s", tt.flowID)
			}
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockKratos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("kratos should not be called for invalid JSON")
	}))
	defer mockKratos.Close()

	mockRepo := new(MockUserRepository)
	handler := newTestAuthHandler(mockKratos.URL, mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/registration?flow=test", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_InitiateLogin(t *testing.T) {
	tests := []struct {
		name           string
		kratosHandler  http.HandlerFunc
		expectedStatus int
	}{
		{
			name: "success",
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/self-service/login/api", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"id":     "flow-123",
					"type":   "api",
					"method": "password",
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "kratos returns valid response",
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"id":         "flow-456",
					"expires_at": time.Now().Add(time.Hour).Format(time.RFC3339),
				})
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKratos := newMockKratosServer(tt.kratosHandler)
			defer mockKratos.Close()

			mockRepo := new(MockUserRepository)
			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
			w := httptest.NewRecorder()

			handler.InitiateLogin(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		flowID         string
		requestBody    LoginRequest
		kratosHandler  http.HandlerFunc
		expectedStatus int
		checkCookies   bool
	}{
		{
			name:   "success with session cookie",
			flowID: "flow-123",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "SecurePassword123!",
			},
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/self-service/login", r.URL.Path)
				assert.Equal(t, "flow-123", r.URL.Query().Get("flow"))
				http.SetCookie(w, &http.Cookie{
					Name:  "ory_kratos_session",
					Value: "session-token-123",
				})
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"session": map[string]any{
						"id":       "session-123",
						"active":   true,
						"identity": map[string]any{"id": "identity-123"},
					},
				})
			},
			expectedStatus: http.StatusOK,
			checkCookies:   true,
		},
		{
			name:   "missing flow parameter",
			flowID: "",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "SecurePassword123!",
			},
			kratosHandler:  nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid credentials",
			flowID: "flow-123",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "WrongPassword",
			},
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]any{
					"error": "invalid credentials",
				})
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockKratos *httptest.Server
			if tt.kratosHandler != nil {
				mockKratos = newMockKratosServer(tt.kratosHandler)
				defer mockKratos.Close()
			} else {
				mockKratos = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Fatal("kratos should not be called")
				}))
				defer mockKratos.Close()
			}

			mockRepo := new(MockUserRepository)
			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			url := "/api/auth/login"
			if tt.flowID != "" {
				url = fmt.Sprintf("/api/auth/login?flow=%s", tt.flowID)
			}
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkCookies {
				cookies := w.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "ory_kratos_session" {
						sessionCookie = c
						break
					}
				}
				require.NotNil(t, sessionCookie, "session cookie should be set")
				assert.True(t, sessionCookie.HttpOnly, "session cookie should be HttpOnly")
				assert.True(t, sessionCookie.Secure, "session cookie should be Secure")
				assert.Equal(t, http.SameSiteStrictMode, sessionCookie.SameSite)
			}
		})
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mockKratos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("kratos should not be called for invalid JSON")
	}))
	defer mockKratos.Close()

	mockRepo := new(MockUserRepository)
	handler := newTestAuthHandler(mockKratos.URL, mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login?flow=test", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Logout(t *testing.T) {
	tests := []struct {
		name           string
		sessionToken   string
		tokenSource    string // "header" or "cookie"
		kratosHandler  http.HandlerFunc
		expectedStatus int
		checkCookie    bool
	}{
		{
			name:         "success with header token",
			sessionToken: "session-token-123",
			tokenSource:  "header",
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/self-service/logout/api", r.URL.Path)
				assert.Equal(t, "session-token-123", r.Header.Get("X-Session-Token"))
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusNoContent,
			checkCookie:    true,
		},
		{
			name:         "success with cookie token",
			sessionToken: "session-from-cookie",
			tokenSource:  "cookie",
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusNoContent,
			checkCookie:    true,
		},
		{
			name:           "no session token",
			sessionToken:   "",
			tokenSource:    "",
			kratosHandler:  nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockKratos *httptest.Server
			if tt.kratosHandler != nil {
				mockKratos = newMockKratosServer(tt.kratosHandler)
				defer mockKratos.Close()
			} else {
				mockKratos = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Fatal("kratos should not be called without session")
				}))
				defer mockKratos.Close()
			}

			mockRepo := new(MockUserRepository)
			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
			if tt.sessionToken != "" {
				if tt.tokenSource == "header" {
					req.Header.Set("Authorization", "Bearer "+tt.sessionToken)
				} else if tt.tokenSource == "cookie" {
					req.AddCookie(&http.Cookie{
						Name:  "ory_kratos_session",
						Value: tt.sessionToken,
					})
				}
			}
			w := httptest.NewRecorder()

			handler.Logout(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkCookie {
				cookies := w.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "ory_kratos_session" {
						sessionCookie = c
						break
					}
				}
				require.NotNil(t, sessionCookie, "session cookie should be set to expire")
				assert.Equal(t, -1, sessionCookie.MaxAge, "cookie should be deleted")
			}
		})
	}
}

func TestAuthHandler_GetCurrentUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		authUser       *middleware.User
		mockUser       *user.User
		mockError      error
		expectedStatus int
	}{
		{
			name: "success - user in database",
			authUser: &middleware.User{
				ID:       "kratos-identity-123",
				Email:    "test@example.com",
				TenantID: "tenant-123",
			},
			mockUser: &user.User{
				ID:               "user-123",
				TenantID:         "tenant-123",
				KratosIdentityID: "kratos-identity-123",
				Email:            "test@example.com",
				Role:             "admin",
				CreatedAt:        now,
				UpdatedAt:        now,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "user not in database - returns auth user",
			authUser: &middleware.User{
				ID:       "kratos-identity-456",
				Email:    "new@example.com",
				TenantID: "tenant-123",
			},
			mockUser:       nil,
			mockError:      fmt.Errorf("user not found"),
			expectedStatus: http.StatusOK, // Still returns OK with auth user info
		},
		{
			name:           "no authenticated user",
			authUser:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKratos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("kratos should not be called for GetCurrentUser")
			}))
			defer mockKratos.Close()

			mockRepo := new(MockUserRepository)
			if tt.authUser != nil {
				mockRepo.On("GetByKratosIdentityID", mock.Anything, tt.authUser.ID).Return(tt.mockUser, tt.mockError)
			}
			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			if tt.authUser != nil {
				ctx := context.WithValue(req.Context(), middleware.UserContextKey, tt.authUser)
				req = req.WithContext(ctx)
			}
			w := httptest.NewRecorder()

			handler.GetCurrentUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_RequestPasswordReset(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    PasswordResetRequest
		kratosHandler  http.HandlerFunc
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: PasswordResetRequest{
				Email: "test@example.com",
			},
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/self-service/recovery/api", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"id":     "recovery-flow-123",
					"type":   "api",
					"method": "code",
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "kratos error - body still decoded",
			requestBody: PasswordResetRequest{
				Email: "test@example.com",
			},
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				// Note: Handler doesn't propagate Kratos status code, it returns 200 if body decodes
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]any{
					"error": "service unavailable",
				})
			},
			expectedStatus: http.StatusOK, // Handler returns 200 as long as body decodes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKratos := newMockKratosServer(tt.kratosHandler)
			defer mockKratos.Close()

			mockRepo := new(MockUserRepository)
			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/password-reset", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.RequestPasswordReset(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_RequestPasswordReset_InvalidJSON(t *testing.T) {
	mockKratos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("kratos should not be called for invalid JSON")
	}))
	defer mockKratos.Close()

	mockRepo := new(MockUserRepository)
	handler := newTestAuthHandler(mockKratos.URL, mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/password-reset", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	handler.RequestPasswordReset(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_ConfirmPasswordReset(t *testing.T) {
	tests := []struct {
		name           string
		flowID         string
		requestBody    PasswordResetConfirm
		kratosHandler  http.HandlerFunc
		expectedStatus int
	}{
		{
			name:   "success",
			flowID: "recovery-flow-123",
			requestBody: PasswordResetConfirm{
				Code:     "123456",
				Password: "NewSecurePassword123!",
			},
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/self-service/recovery", r.URL.Path)
				assert.Equal(t, "recovery-flow-123", r.URL.Query().Get("flow"))
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"status": "success",
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "missing flow parameter",
			flowID: "",
			requestBody: PasswordResetConfirm{
				Code:     "123456",
				Password: "NewSecurePassword123!",
			},
			kratosHandler:  nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid code",
			flowID: "recovery-flow-123",
			requestBody: PasswordResetConfirm{
				Code:     "invalid-code",
				Password: "NewSecurePassword123!",
			},
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]any{
					"error": "invalid code",
				})
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockKratos *httptest.Server
			if tt.kratosHandler != nil {
				mockKratos = newMockKratosServer(tt.kratosHandler)
				defer mockKratos.Close()
			} else {
				mockKratos = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Fatal("kratos should not be called")
				}))
				defer mockKratos.Close()
			}

			mockRepo := new(MockUserRepository)
			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			url := "/api/auth/password-reset/confirm"
			if tt.flowID != "" {
				url = fmt.Sprintf("/api/auth/password-reset/confirm?flow=%s", tt.flowID)
			}
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.ConfirmPasswordReset(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_ConfirmPasswordReset_InvalidJSON(t *testing.T) {
	mockKratos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("kratos should not be called for invalid JSON")
	}))
	defer mockKratos.Close()

	mockRepo := new(MockUserRepository)
	handler := newTestAuthHandler(mockKratos.URL, mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/password-reset/confirm?flow=test", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	handler.ConfirmPasswordReset(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RequestEmailVerification(t *testing.T) {
	tests := []struct {
		name           string
		kratosHandler  http.HandlerFunc
		expectedStatus int
	}{
		{
			name: "success",
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/self-service/verification/api", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"id":     "verification-flow-123",
					"type":   "api",
					"method": "code",
				})
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKratos := newMockKratosServer(tt.kratosHandler)
			defer mockKratos.Close()

			mockRepo := new(MockUserRepository)
			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/email-verification", nil)
			w := httptest.NewRecorder()

			handler.RequestEmailVerification(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_ConfirmEmailVerification(t *testing.T) {
	tests := []struct {
		name           string
		flowID         string
		requestBody    VerificationConfirm
		kratosHandler  http.HandlerFunc
		expectedStatus int
	}{
		{
			name:   "success",
			flowID: "verification-flow-123",
			requestBody: VerificationConfirm{
				Code: "123456",
			},
			kratosHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/self-service/verification", r.URL.Path)
				assert.Equal(t, "verification-flow-123", r.URL.Query().Get("flow"))
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"status": "verified",
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "missing flow parameter",
			flowID: "",
			requestBody: VerificationConfirm{
				Code: "123456",
			},
			kratosHandler:  nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockKratos *httptest.Server
			if tt.kratosHandler != nil {
				mockKratos = newMockKratosServer(tt.kratosHandler)
				defer mockKratos.Close()
			} else {
				mockKratos = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Fatal("kratos should not be called")
				}))
				defer mockKratos.Close()
			}

			mockRepo := new(MockUserRepository)
			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			url := "/api/auth/email-verification/confirm"
			if tt.flowID != "" {
				url = fmt.Sprintf("/api/auth/email-verification/confirm?flow=%s", tt.flowID)
			}
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.ConfirmEmailVerification(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_ConfirmEmailVerification_InvalidJSON(t *testing.T) {
	mockKratos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("kratos should not be called for invalid JSON")
	}))
	defer mockKratos.Close()

	mockRepo := new(MockUserRepository)
	handler := newTestAuthHandler(mockKratos.URL, mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/email-verification/confirm?flow=test", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	handler.ConfirmEmailVerification(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_KratosWebhook(t *testing.T) {
	// Store original env var and restore after test
	originalSecret := os.Getenv("KRATOS_WEBHOOK_SECRET")
	defer os.Setenv("KRATOS_WEBHOOK_SECRET", originalSecret)

	tests := []struct {
		name           string
		secretEnv      string
		requestSecret  string
		webhook        user.KratosIdentityWebhook
		mockError      error
		expectedStatus int
	}{
		{
			name:          "success",
			secretEnv:     "webhook-secret-123",
			requestSecret: "webhook-secret-123",
			webhook: user.KratosIdentityWebhook{
				IdentityID: "identity-123",
				Email:      "test@example.com",
				TenantID:   "tenant-123",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing secret env var",
			secretEnv:      "",
			requestSecret:  "some-secret",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:          "invalid secret",
			secretEnv:     "correct-secret",
			requestSecret: "wrong-secret",
			webhook: user.KratosIdentityWebhook{
				IdentityID: "identity-123",
				Email:      "test@example.com",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:          "sync error",
			secretEnv:     "webhook-secret-123",
			requestSecret: "webhook-secret-123",
			webhook: user.KratosIdentityWebhook{
				IdentityID: "identity-123",
				Email:      "test@example.com",
				TenantID:   "tenant-123",
			},
			mockError:      fmt.Errorf("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("KRATOS_WEBHOOK_SECRET", tt.secretEnv)

			mockKratos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("kratos should not be called for webhook")
			}))
			defer mockKratos.Close()

			mockRepo := new(MockUserRepository)

			// Set up mock for sync
			if tt.secretEnv != "" && tt.requestSecret == tt.secretEnv {
				// Mock GetByKratosIdentityID to return not found (new user)
				mockRepo.On("GetByKratosIdentityID", mock.Anything, tt.webhook.IdentityID).Return(nil, fmt.Errorf("not found"))
				if tt.mockError == nil {
					mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
				} else {
					mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(tt.mockError)
				}
			}

			handler := newTestAuthHandler(mockKratos.URL, mockRepo)

			body, _ := json.Marshal(tt.webhook)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/webhook/kratos", bytes.NewReader(body))
			req.Header.Set("X-Webhook-Secret", tt.requestSecret)
			w := httptest.NewRecorder()

			handler.KratosWebhook(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_KratosWebhook_InvalidJSON(t *testing.T) {
	originalSecret := os.Getenv("KRATOS_WEBHOOK_SECRET")
	defer os.Setenv("KRATOS_WEBHOOK_SECRET", originalSecret)
	os.Setenv("KRATOS_WEBHOOK_SECRET", "valid-secret")

	mockKratos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("kratos should not be called")
	}))
	defer mockKratos.Close()

	mockRepo := new(MockUserRepository)
	handler := newTestAuthHandler(mockKratos.URL, mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/webhook/kratos", bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("X-Webhook-Secret", "valid-secret")
	w := httptest.NewRecorder()

	handler.KratosWebhook(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExtractSessionToken(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func(req *http.Request)
		expectedToken string
	}{
		{
			name: "from bearer header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer token-from-header")
			},
			expectedToken: "token-from-header",
		},
		{
			name: "from cookie",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "ory_kratos_session",
					Value: "token-from-cookie",
				})
			},
			expectedToken: "token-from-cookie",
		},
		{
			name: "header takes precedence over cookie",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer token-from-header")
				req.AddCookie(&http.Cookie{
					Name:  "ory_kratos_session",
					Value: "token-from-cookie",
				})
			},
			expectedToken: "token-from-header",
		},
		{
			name:          "no token",
			setupRequest:  func(req *http.Request) {},
			expectedToken: "",
		},
		{
			name: "invalid bearer format",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "NotBearer token")
			},
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			tt.setupRequest(req)

			token := extractSessionToken(req)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}
