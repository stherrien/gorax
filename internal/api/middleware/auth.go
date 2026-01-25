package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorax/gorax/internal/config"
)

type contextKey string

const (
	// UserContextKey is the context key for the authenticated user
	UserContextKey contextKey = "user"
)

// User represents an authenticated user from Kratos
type User struct {
	ID       string                 `json:"id"`
	Email    string                 `json:"email"`
	Traits   map[string]interface{} `json:"traits"`
	TenantID string                 `json:"tenant_id"`
}

// kratosSession represents the session response from Kratos
type kratosSession struct {
	Active   bool `json:"active"`
	Identity struct {
		ID     string                 `json:"id"`
		Traits map[string]interface{} `json:"traits"`
	} `json:"identity"`
}

// KratosAuth returns middleware that validates sessions with Ory Kratos
func KratosAuth(cfg config.KratosConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract session token from cookie or Authorization header
			sessionToken := extractSessionToken(r)
			if sessionToken == "" {
				http.Error(w, "unauthorized: no session token", http.StatusUnauthorized)
				return
			}

			// Validate session with Kratos
			user, err := validateKratosSession(cfg.PublicURL, sessionToken)
			if err != nil {
				http.Error(w, "unauthorized: invalid session", http.StatusUnauthorized)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractSessionToken(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Try session cookie
	cookie, err := r.Cookie("ory_kratos_session")
	if err == nil {
		return cookie.Value
	}

	return ""
}

func validateKratosSession(kratosURL, sessionToken string) (*User, error) {
	// Create request to Kratos whoami endpoint
	req, err := http.NewRequest("GET", kratosURL+"/sessions/whoami", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Session-Token", sessionToken)

	// Make request with timeout to prevent slowloris/timeout attacks
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidSession
	}

	// Parse response
	var session kratosSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, err
	}

	if !session.Active {
		return nil, ErrInvalidSession
	}

	// Extract email from traits
	email, _ := session.Identity.Traits["email"].(string)

	// Extract tenant_id from traits (set during registration)
	tenantID, _ := session.Identity.Traits["tenant_id"].(string)

	return &User{
		ID:       session.Identity.ID,
		Email:    email,
		Traits:   session.Identity.Traits,
		TenantID: tenantID,
	}, nil
}

// GetUser extracts the user from the request context
func GetUser(r *http.Request) *User {
	user, _ := r.Context().Value(UserContextKey).(*User)
	return user
}

// GetUserID extracts just the user ID from the request context
func GetUserID(r *http.Request) string {
	user := GetUser(r)
	if user != nil {
		return user.ID
	}
	return ""
}

// Custom errors
type AuthError struct {
	Message string
}

func (e AuthError) Error() string {
	return e.Message
}

var ErrInvalidSession = AuthError{Message: "invalid session"}
