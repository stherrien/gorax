package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateKratosSession_Success(t *testing.T) {
	// Create mock Kratos server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/sessions/whoami", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("X-Session-Token"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"active": true,
			"identity": {
				"id": "user-123",
				"traits": {
					"email": "test@example.com",
					"tenant_id": "tenant-456"
				}
			}
		}`))
	}))
	defer server.Close()

	user, err := validateKratosSession(server.URL, "test-token")
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "tenant-456", user.TenantID)
}

func TestValidateKratosSession_InactiveSession(t *testing.T) {
	// Create mock Kratos server returning inactive session
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"active": false,
			"identity": {
				"id": "user-123",
				"traits": {}
			}
		}`))
	}))
	defer server.Close()

	user, err := validateKratosSession(server.URL, "test-token")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInvalidSession, err)
}

func TestValidateKratosSession_Unauthorized(t *testing.T) {
	// Create mock Kratos server returning 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	user, err := validateKratosSession(server.URL, "invalid-token")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInvalidSession, err)
}

func TestValidateKratosSession_Timeout(t *testing.T) {
	// Create slow server that delays response longer than timeout
	// Use a blocking channel to simulate a hanging connection
	blockChan := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until channel is closed or test completes
		<-blockChan
	}))
	defer func() {
		close(blockChan)
		server.Close()
	}()

	start := time.Now()
	user, err := validateKratosSession(server.URL, "test-token")
	duration := time.Since(start)

	// Should fail with timeout error
	require.Error(t, err)
	assert.Nil(t, user)

	// Should timeout in approximately 10 seconds (allow some margin)
	assert.Less(t, duration, 11*time.Second, "Request should timeout before 11 seconds")
	assert.Greater(t, duration, 9*time.Second, "Request should take at least 9 seconds to timeout")

	// Error should indicate timeout or context deadline
	errMsg := err.Error()
	assert.True(t,
		containsAny(errMsg, []string{"timeout", "deadline", "Client.Timeout exceeded"}),
		"Error should indicate timeout, got: %s", errMsg)
}

func TestValidateKratosSession_MalformedJSON(t *testing.T) {
	// Create mock Kratos server returning invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"invalid json`))
	}))
	defer server.Close()

	user, err := validateKratosSession(server.URL, "test-token")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestExtractSessionToken_FromBearer(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")

	token := extractSessionToken(req)
	assert.Equal(t, "test-token-123", token)
}

func TestExtractSessionToken_FromCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "ory_kratos_session",
		Value: "cookie-token-456",
	})

	token := extractSessionToken(req)
	assert.Equal(t, "cookie-token-456", token)
}

func TestExtractSessionToken_BearerPriority(t *testing.T) {
	// When both are present, Bearer token takes priority
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")
	req.AddCookie(&http.Cookie{
		Name:  "ory_kratos_session",
		Value: "cookie-token",
	})

	token := extractSessionToken(req)
	assert.Equal(t, "bearer-token", token)
}

func TestExtractSessionToken_NoToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	token := extractSessionToken(req)
	assert.Empty(t, token)
}

func TestGetUser(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	// No user in context
	user := GetUser(req)
	assert.Nil(t, user)

	// With user in context
	expectedUser := &User{
		ID:       "user-123",
		Email:    "test@example.com",
		TenantID: "tenant-456",
	}
	ctx := req.Context()
	ctx = contextWithValue(ctx, UserContextKey, expectedUser)
	req = req.WithContext(ctx)

	user = GetUser(req)
	assert.Equal(t, expectedUser, user)
}

func TestGetUserID(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	// No user in context
	userID := GetUserID(req)
	assert.Empty(t, userID)

	// With user in context
	expectedUser := &User{ID: "user-789"}
	ctx := req.Context()
	ctx = contextWithValue(ctx, UserContextKey, expectedUser)
	req = req.WithContext(ctx)

	userID = GetUserID(req)
	assert.Equal(t, "user-789", userID)
}

// Helper function to add value to context
func contextWithValue(ctx context.Context, key contextKey, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// Helper function to check if string contains any of the substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
