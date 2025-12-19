package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClient tests client creation
func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		accessToken string
		wantErr     bool
	}{
		{
			name:        "valid token",
			accessToken: "xoxb-test-token",
			wantErr:     false,
		},
		{
			name:        "empty token",
			accessToken: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.accessToken)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// TestClient_SendMessage tests the SendMessage method
func TestClient_SendMessage(t *testing.T) {
	tests := []struct {
		name           string
		request        *SendMessageRequest
		mockResponse   interface{}
		mockStatusCode int
		wantErr        bool
		errorContains  string
	}{
		{
			name: "successful message send",
			request: &SendMessageRequest{
				Channel: "C1234567890",
				Text:    "Hello, world!",
			},
			mockResponse: MessageResponse{
				OK:      true,
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Message: Message{
					Type: "message",
					Text: "Hello, world!",
					TS:   "1503435956.000247",
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "channel not found",
			request: &SendMessageRequest{
				Channel: "C9999999999",
				Text:    "Hello",
			},
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "channel_not_found",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        true,
			errorContains:  "channel not found",
		},
		{
			name: "message too long",
			request: &SendMessageRequest{
				Channel: "C1234567890",
				Text:    string(make([]byte, 50000)), // > 40k chars
			},
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "msg_too_long",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        true,
			errorContains:  "text exceeds",
		},
		{
			name: "invalid auth token",
			request: &SendMessageRequest{
				Channel: "C1234567890",
				Text:    "Hello",
			},
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "invalid_auth",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        true,
			errorContains:  "invalid authentication",
		},
		{
			name: "rate limited",
			request: &SendMessageRequest{
				Channel: "C1234567890",
				Text:    "Hello",
			},
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "rate_limited",
			},
			mockStatusCode: http.StatusTooManyRequests,
			wantErr:        true,
			errorContains:  "rate limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/chat.postMessage", r.URL.Path)
				assert.Equal(t, "Bearer xoxb-test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json; charset=utf-8", r.Header.Get("Content-Type"))

				// Send mock response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create client with mock server URL
			client, err := NewClient("xoxb-test-token")
			require.NoError(t, err)
			client.baseURL = server.URL

			// Execute
			ctx := context.Background()
			resp, err := client.SendMessage(ctx, tt.request)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.OK)
				assert.Equal(t, tt.request.Channel, resp.Channel)
			}
		})
	}
}

// TestClient_UpdateMessage tests the UpdateMessage method
func TestClient_UpdateMessage(t *testing.T) {
	tests := []struct {
		name           string
		request        *UpdateMessageRequest
		mockResponse   interface{}
		mockStatusCode int
		wantErr        bool
	}{
		{
			name: "successful update",
			request: &UpdateMessageRequest{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Text:    "Updated text",
			},
			mockResponse: MessageResponse{
				OK:      true,
				Channel: "C1234567890",
				TS:      "1503435956.000247",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "message not found",
			request: &UpdateMessageRequest{
				Channel: "C1234567890",
				TS:      "9999999999.999999",
				Text:    "Updated",
			},
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "message_not_found",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        true,
		},
		{
			name: "cannot update message",
			request: &UpdateMessageRequest{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Text:    "Updated",
			},
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "cant_update_message",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat.update", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client, _ := NewClient("xoxb-test-token")
			client.baseURL = server.URL

			ctx := context.Background()
			resp, err := client.UpdateMessage(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.True(t, resp.OK)
			}
		})
	}
}

// TestClient_AddReaction tests the AddReaction method
func TestClient_AddReaction(t *testing.T) {
	tests := []struct {
		name           string
		channel        string
		timestamp      string
		emoji          string
		mockResponse   interface{}
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:      "successful reaction",
			channel:   "C1234567890",
			timestamp: "1503435956.000247",
			emoji:     "thumbsup",
			mockResponse: APIResponse{
				OK: true,
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:      "message not found",
			channel:   "C1234567890",
			timestamp: "9999999999.999999",
			emoji:     "thumbsup",
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "message_not_found",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        true,
		},
		{
			name:      "invalid emoji",
			channel:   "C1234567890",
			timestamp: "1503435956.000247",
			emoji:     "invalid_emoji_name",
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "invalid_name",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        true,
		},
		{
			name:      "already reacted",
			channel:   "C1234567890",
			timestamp: "1503435956.000247",
			emoji:     "thumbsup",
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "already_reacted",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false, // This is not really an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/reactions.add", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client, _ := NewClient("xoxb-test-token")
			client.baseURL = server.URL

			ctx := context.Background()
			err := client.AddReaction(ctx, tt.channel, tt.timestamp, tt.emoji)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestClient_GetUserByEmail tests user lookup by email
func TestClient_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		mockResponse   interface{}
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:  "user found",
			email: "user@example.com",
			mockResponse: UserByEmailResponse{
				OK: true,
				User: User{
					ID:       "U1234567890",
					TeamID:   "T1234567890",
					Name:     "testuser",
					RealName: "Test User",
					Profile: UserProfile{
						Email:       "user@example.com",
						DisplayName: "Test User",
					},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "users_not_found",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/users.lookupByEmail", r.URL.Path)
				assert.Equal(t, tt.email, r.URL.Query().Get("email"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client, _ := NewClient("xoxb-test-token")
			client.baseURL = server.URL

			ctx := context.Background()
			user, err := client.GetUserByEmail(ctx, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Profile.Email)
			}
		})
	}
}

// TestClient_OpenConversation tests opening a DM
func TestClient_OpenConversation(t *testing.T) {
	tests := []struct {
		name           string
		users          []string
		mockResponse   interface{}
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:  "open DM successfully",
			users: []string{"U1234567890"},
			mockResponse: OpenConversationResponse{
				OK: true,
				Channel: Conversation{
					ID:   "D1234567890",
					IsIM: true,
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:  "user not found",
			users: []string{"U9999999999"},
			mockResponse: ErrorResponse{
				OK:    false,
				Error: "user_not_found",
			},
			mockStatusCode: http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/conversations.open", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client, _ := NewClient("xoxb-test-token")
			client.baseURL = server.URL

			ctx := context.Background()
			conv, err := client.OpenConversation(ctx, tt.users)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, conv)
				assert.True(t, conv.IsIM)
			}
		})
	}
}

// TestClient_RateLimitRetry tests rate limit handling with retry
func TestClient_RateLimitRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			// First request: rate limited
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(ErrorResponse{
				OK:    false,
				Error: "rate_limited",
			})
		} else {
			// Second request: success
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(MessageResponse{
				OK:      true,
				Channel: "C1234567890",
				TS:      "1503435956.000247",
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient("xoxb-test-token")
	client.baseURL = server.URL
	client.maxRetries = 3
	client.retryDelay = 100 * time.Millisecond // Short delay for testing

	ctx := context.Background()
	req := &SendMessageRequest{
		Channel: "C1234567890",
		Text:    "Hello",
	}

	start := time.Now()
	resp, err := client.SendMessage(ctx, req)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, attempts) // Should have retried once
	assert.True(t, duration >= 100*time.Millisecond, "Should have waited before retry")
}

// TestClient_ContextCancellation tests that context cancellation is respected
func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{OK: true})
	}))
	defer server.Close()

	client, _ := NewClient("xoxb-test-token")
	client.baseURL = server.URL

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req := &SendMessageRequest{
		Channel: "C1234567890",
		Text:    "Hello",
	}

	_, err := client.SendMessage(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// TestClient_InvalidJSON tests handling of invalid JSON responses
func TestClient_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json {{{"))
	}))
	defer server.Close()

	client, _ := NewClient("xoxb-test-token")
	client.baseURL = server.URL

	ctx := context.Background()
	req := &SendMessageRequest{
		Channel: "C1234567890",
		Text:    "Hello",
	}

	_, err := client.SendMessage(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json")
}
