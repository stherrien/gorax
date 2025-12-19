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

// TestAddReactionAction_Execute tests the AddReaction action
func TestAddReactionAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         AddReactionConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		credError      error
		slackResponse  interface{}
		slackStatus    int
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful reaction add",
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "1503435956.000247",
				Emoji:     "thumbsup",
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
			slackResponse: APIResponse{
				OK: true,
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				assert.NotNil(t, output)
				assert.NotNil(t, output.Data)

				result, ok := output.Data.(*AddReactionResult)
				assert.True(t, ok, "output.Data should be *AddReactionResult")
				assert.True(t, result.OK)
				assert.Equal(t, "C1234567890", result.Channel)
				assert.Equal(t, "1503435956.000247", result.Timestamp)
				assert.Equal(t, "thumbsup", result.Emoji)
			},
		},
		{
			name: "successful reaction with colon prefix/suffix",
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "1503435956.000247",
				Emoji:     ":thumbsup:",
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
			slackResponse: APIResponse{
				OK: true,
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*AddReactionResult)
				assert.True(t, ok)
				assert.True(t, result.OK)
				// Emoji should be stored without colons
				assert.Equal(t, "thumbsup", result.Emoji)
			},
		},
		{
			name: "successful reaction with various emojis",
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "1503435956.000247",
				Emoji:     "white_check_mark",
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
			slackResponse: APIResponse{
				OK: true,
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
		},
		{
			name: "already reacted - treated as success",
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "1503435956.000247",
				Emoji:     "thumbsup",
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
			slackResponse: ErrorResponse{
				OK:    false,
				Error: "already_reacted",
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				result, ok := output.Data.(*AddReactionResult)
				assert.True(t, ok)
				assert.True(t, result.OK)
				// Client handles "already_reacted" transparently as success
				assert.Equal(t, "thumbsup", result.Emoji)
			},
		},
		{
			name: "missing channel",
			config: AddReactionConfig{
				Timestamp: "1503435956.000247",
				Emoji:     "thumbsup",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
			},
			wantErr:       true,
			errorContains: "channel is required",
		},
		{
			name: "missing timestamp",
			config: AddReactionConfig{
				Channel: "C1234567890",
				Emoji:   "thumbsup",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
			},
			wantErr:       true,
			errorContains: "timestamp is required",
		},
		{
			name: "missing emoji",
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "1503435956.000247",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
			},
			wantErr:       true,
			errorContains: "emoji is required",
		},
		{
			name: "missing credential_id",
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "1503435956.000247",
				Emoji:     "thumbsup",
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
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "1503435956.000247",
				Emoji:     "thumbsup",
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
			name: "message not found",
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "9999999999.999999",
				Emoji:     "thumbsup",
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
			slackResponse: ErrorResponse{
				OK:    false,
				Error: "message_not_found",
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "message not found",
		},
		{
			name: "invalid emoji name",
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "1503435956.000247",
				Emoji:     "invalid_emoji_that_does_not_exist",
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
			slackResponse: ErrorResponse{
				OK:    false,
				Error: "invalid_name",
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "invalid arguments",
		},
		{
			name: "channel not found",
			config: AddReactionConfig{
				Channel:   "C9999999999",
				Timestamp: "1503435956.000247",
				Emoji:     "thumbsup",
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
			slackResponse: ErrorResponse{
				OK:    false,
				Error: "channel_not_found",
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "channel not found",
		},
		{
			name: "invalid auth",
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "1503435956.000247",
				Emoji:     "thumbsup",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-invalid-token",
				},
			},
			slackResponse: ErrorResponse{
				OK:    false,
				Error: "invalid_auth",
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "invalid authentication",
		},
		{
			name: "no permission",
			config: AddReactionConfig{
				Channel:   "C1234567890",
				Timestamp: "1503435956.000247",
				Emoji:     "thumbsup",
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
			slackResponse: ErrorResponse{
				OK:    false,
				Error: "missing_scope",
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "missing required OAuth scope",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Slack server
			var server *httptest.Server
			if tt.slackResponse != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/reactions.add", r.URL.Path)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.slackStatus)
					json.NewEncoder(w).Encode(tt.slackResponse)
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
			action := &AddReactionAction{
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

// TestAddReactionAction_Execute_ContextCancellation tests context cancellation
func TestAddReactionAction_Execute_ContextCancellation(t *testing.T) {
	// Create slow mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(APIResponse{OK: true})
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
	action := &AddReactionAction{
		credentialService: mockCred,
		baseURL:           server.URL,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Execute
	config := AddReactionConfig{
		Channel:   "C1234567890",
		Timestamp: "1503435956.000247",
		Emoji:     "thumbsup",
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

// TestAddReactionAction_NewAddReactionAction tests action constructor
func TestAddReactionAction_NewAddReactionAction(t *testing.T) {
	mockCred := &MockCredentialService{}

	action := NewAddReactionAction(mockCred)

	require.NotNil(t, action)
	assert.NotNil(t, action.credentialService)
	assert.Equal(t, DefaultBaseURL, action.baseURL)
}

// TestAddReactionResult_JSON tests JSON serialization
func TestAddReactionResult_JSON(t *testing.T) {
	result := &AddReactionResult{
		OK:        true,
		Channel:   "C1234567890",
		Timestamp: "1503435956.000247",
		Emoji:     "thumbsup",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded AddReactionResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.OK, decoded.OK)
	assert.Equal(t, result.Channel, decoded.Channel)
	assert.Equal(t, result.Timestamp, decoded.Timestamp)
	assert.Equal(t, result.Emoji, decoded.Emoji)
}

// TestAddReactionAction_FromPreviousStepOutput tests using timestamp from previous step
func TestAddReactionAction_FromPreviousStepOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(APIResponse{
			OK: true,
		})
	}))
	defer server.Close()

	mockCred := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "xoxb-test-token",
				},
			}, nil
		},
	}

	action := &AddReactionAction{
		credentialService: mockCred,
		baseURL:           server.URL,
	}

	config := AddReactionConfig{
		Channel:   "C1234567890",
		Timestamp: "1503435956.000247",
		Emoji:     "white_check_mark",
	}

	// Simulate context with previous step output
	input := actions.NewActionInput(config, map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
		"credential_id": "cred-123",
		"steps": map[string]interface{}{
			"send-message": map[string]interface{}{
				"timestamp": "1503435956.000247",
				"channel":   "C1234567890",
			},
		},
	})

	ctx := context.Background()
	output, err := action.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, output)

	result, ok := output.Data.(*AddReactionResult)
	assert.True(t, ok)
	assert.True(t, result.OK)
	assert.Equal(t, "white_check_mark", result.Emoji)
}

// TestAddReactionAction_EmojiNormalization tests emoji name normalization
func TestAddReactionAction_EmojiNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no colons",
			input:    "thumbsup",
			expected: "thumbsup",
		},
		{
			name:     "with colons",
			input:    ":thumbsup:",
			expected: "thumbsup",
		},
		{
			name:     "leading colon only",
			input:    ":thumbsup",
			expected: "thumbsup",
		},
		{
			name:     "trailing colon only",
			input:    "thumbsup:",
			expected: "thumbsup",
		},
		{
			name:     "multiple colons",
			input:    ":::thumbsup:::",
			expected: "thumbsup",
		},
		{
			name:     "complex emoji name",
			input:    ":white_check_mark:",
			expected: "white_check_mark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := normalizeEmoji(tt.input)
			assert.Equal(t, tt.expected, normalized)
		})
	}
}
