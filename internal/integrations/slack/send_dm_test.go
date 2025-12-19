package slack

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
)

// TestSendDMAction_Execute tests the SendDM action
func TestSendDMAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         SendDMConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		credError      error
		slackResponses map[string]interface{} // Map endpoint to response
		slackStatus    int
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful DM send with user ID",
			config: SendDMConfig{
				User: "U1234567890",
				Text: "Hello via DM!",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-test-token",
				},
			},
			slackResponses: map[string]interface{}{
				"/conversations.open": OpenConversationResponse{
					OK: true,
					Channel: Conversation{
						ID:   "D1234567890",
						IsIM: true,
					},
				},
				"/chat.postMessage": MessageResponse{
					OK:      true,
					Channel: "D1234567890",
					TS:      "1503435956.000247",
					Message: Message{
						Type: "message",
						Text: "Hello via DM!",
						TS:   "1503435956.000247",
					},
				},
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				assert.NotNil(t, output)
				assert.NotNil(t, output.Data)

				result, ok := output.Data.(*SendDMResult)
				assert.True(t, ok, "output.Data should be *SendDMResult")
				assert.True(t, result.OK)
				assert.Equal(t, "D1234567890", result.Channel)
				assert.Equal(t, "U1234567890", result.UserID)
				assert.Equal(t, "1503435956.000247", result.Timestamp)
				assert.NotNil(t, result.Message)
				assert.Equal(t, "Hello via DM!", result.Message.Text)
			},
		},
		{
			name: "successful DM send with email lookup",
			config: SendDMConfig{
				User: "user@example.com",
				Text: "Hello via email!",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-test-token",
				},
			},
			slackResponses: map[string]interface{}{
				"/users.lookupByEmail": UserByEmailResponse{
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
				"/conversations.open": OpenConversationResponse{
					OK: true,
					Channel: Conversation{
						ID:   "D1234567890",
						IsIM: true,
					},
				},
				"/chat.postMessage": MessageResponse{
					OK:      true,
					Channel: "D1234567890",
					TS:      "1503435956.000247",
					Message: Message{
						Type: "message",
						Text: "Hello via email!",
						TS:   "1503435956.000247",
					},
				},
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*SendDMResult)
				assert.True(t, ok)
				assert.True(t, result.OK)
				assert.Equal(t, "D1234567890", result.Channel)
				assert.Equal(t, "U1234567890", result.UserID)
			},
		},
		{
			name: "successful DM with blocks",
			config: SendDMConfig{
				User: "U1234567890",
				Blocks: []map[string]interface{}{
					{
						"type": "section",
						"text": map[string]interface{}{
							"type": "mrkdwn",
							"text": "*Important notification*",
						},
					},
				},
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-test-token",
				},
			},
			slackResponses: map[string]interface{}{
				"/conversations.open": OpenConversationResponse{
					OK: true,
					Channel: Conversation{
						ID:   "D1234567890",
						IsIM: true,
					},
				},
				"/chat.postMessage": MessageResponse{
					OK:      true,
					Channel: "D1234567890",
					TS:      "1503435956.000247",
				},
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
		},
		{
			name: "missing user",
			config: SendDMConfig{
				Text: "Hello",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
			},
			wantErr:       true,
			errorContains: "user is required",
		},
		{
			name: "missing text and blocks",
			config: SendDMConfig{
				User: "U1234567890",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
			},
			wantErr:       true,
			errorContains: "either text or blocks must be provided",
		},
		{
			name: "text too long",
			config: SendDMConfig{
				User: "U1234567890",
				Text: string(make([]byte, 50000)),
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
			},
			wantErr:       true,
			errorContains: "text exceeds",
		},
		{
			name: "missing credential_id",
			config: SendDMConfig{
				User: "U1234567890",
				Text: "Hello",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
			},
			wantErr:       true,
			errorContains: "credential_id is required",
		},
		{
			name: "credential not found",
			config: SendDMConfig{
				User: "U1234567890",
				Text: "Hello",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-not-found",
			},
			credError:     errors.New("credential not found"),
			wantErr:       true,
			errorContains: "credential not found",
		},
		{
			name: "user not found by email",
			config: SendDMConfig{
				User: "nonexistent@example.com",
				Text: "Hello",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-test-token",
				},
			},
			slackResponses: map[string]interface{}{
				"/users.lookupByEmail": ErrorResponse{
					OK:    false,
					Error: "users_not_found",
				},
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "user not found",
		},
		{
			name: "user not found by ID",
			config: SendDMConfig{
				User: "U9999999999",
				Text: "Hello",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-test-token",
				},
			},
			slackResponses: map[string]interface{}{
				"/conversations.open": ErrorResponse{
					OK:    false,
					Error: "user_not_found",
				},
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "user not found",
		},
		{
			name: "failed to open conversation",
			config: SendDMConfig{
				User: "U1234567890",
				Text: "Hello",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-test-token",
				},
			},
			slackResponses: map[string]interface{}{
				"/conversations.open": ErrorResponse{
					OK:    false,
					Error: "invalid_auth",
				},
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "failed to open conversation",
		},
		{
			name: "failed to send message",
			config: SendDMConfig{
				User: "U1234567890",
				Text: "Hello",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-test-token",
				},
			},
			slackResponses: map[string]interface{}{
				"/conversations.open": OpenConversationResponse{
					OK: true,
					Channel: Conversation{
						ID:   "D1234567890",
						IsIM: true,
					},
				},
				"/chat.postMessage": ErrorResponse{
					OK:    false,
					Error: "msg_too_long",
				},
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "text exceeds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Slack server
			var server *httptest.Server
			if tt.slackResponses != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.slackStatus)

					// Route based on endpoint
					if resp, ok := tt.slackResponses[r.URL.Path]; ok {
						json.NewEncoder(w).Encode(resp)
					} else {
						// Default error response
						json.NewEncoder(w).Encode(ErrorResponse{
							OK:    false,
							Error: "not_found",
						})
					}
				}))
				defer server.Close()
			}

			// Create mock credential service
			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					if tt.credError != nil {
						return nil, tt.credError
					}
					return tt.mockCredential, nil
				},
			}

			// Create action
			action := &SendDMAction{
				credentialService: mockCred,
			}
			if server != nil {
				action.baseURL = server.URL
			}

			// Execute
			input := actions.NewActionInput(tt.config, tt.context)
			ctx := context.Background()
			output, err := action.Execute(ctx, input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

// TestSendDMAction_Execute_ContextCancellation tests context cancellation
func TestSendDMAction_Execute_ContextCancellation(t *testing.T) {
	// Create slow mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(OpenConversationResponse{
			OK: true,
			Channel: Conversation{
				ID:   "D1234567890",
				IsIM: true,
			},
		})
	}))
	defer server.Close()

	// Create mock credential service
	mockCred := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-test-token",
				},
			}, nil
		},
	}

	// Create action
	action := &SendDMAction{
		credentialService: mockCred,
		baseURL:           server.URL,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Execute
	config := SendDMConfig{
		User: "U1234567890",
		Text: "Hello",
	}
	input := actions.NewActionInput(config, map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
		"credential_id": "cred-123",
	})

	_, err := action.Execute(ctx, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// TestSendDMAction_NewSendDMAction tests action constructor
func TestSendDMAction_NewSendDMAction(t *testing.T) {
	mockCred := &MockCredentialService{}

	action := NewSendDMAction(mockCred)

	require.NotNil(t, action)
	assert.NotNil(t, action.credentialService)
	assert.Equal(t, DefaultBaseURL, action.baseURL)
}

// TestSendDMResult_JSON tests JSON serialization
func TestSendDMResult_JSON(t *testing.T) {
	result := &SendDMResult{
		OK:        true,
		UserID:    "U1234567890",
		Channel:   "D1234567890",
		Timestamp: "1503435956.000247",
		Message: &Message{
			Type: "message",
			Text: "Hello",
			TS:   "1503435956.000247",
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded SendDMResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.OK, decoded.OK)
	assert.Equal(t, result.UserID, decoded.UserID)
	assert.Equal(t, result.Channel, decoded.Channel)
	assert.Equal(t, result.Timestamp, decoded.Timestamp)
}

// TestSendDMAction_UserIdentification tests user identification logic
func TestSendDMAction_UserIdentification(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		isEmail  bool
		isUserID bool
	}{
		{
			name:     "valid user ID",
			user:     "U1234567890",
			isEmail:  false,
			isUserID: true,
		},
		{
			name:     "valid email",
			user:     "user@example.com",
			isEmail:  true,
			isUserID: false,
		},
		{
			name:     "email with subdomain",
			user:     "user@subdomain.example.com",
			isEmail:  true,
			isUserID: false,
		},
		{
			name:     "user ID with W prefix (bot)",
			user:     "W1234567890",
			isEmail:  false,
			isUserID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmail := isEmail(tt.user)
			assert.Equal(t, tt.isEmail, isEmail, "isEmail check")

			if !isEmail {
				// Should be a valid user ID
				assert.True(t, len(tt.user) > 0, "user ID should not be empty")
			}
		})
	}
}
