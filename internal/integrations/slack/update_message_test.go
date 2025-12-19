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

// TestUpdateMessageAction_Execute tests the UpdateMessage action
func TestUpdateMessageAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         UpdateMessageConfig
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
			name: "successful message update with text",
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Text:    "Updated message text",
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
			slackResponse: MessageResponse{
				OK:      true,
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Message: Message{
					Type: "message",
					Text: "Updated message text",
					TS:   "1503435956.000247",
				},
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				assert.NotNil(t, output)
				assert.NotNil(t, output.Data)

				result, ok := output.Data.(*UpdateMessageResult)
				assert.True(t, ok, "output.Data should be *UpdateMessageResult")
				assert.True(t, result.OK)
				assert.Equal(t, "C1234567890", result.Channel)
				assert.Equal(t, "1503435956.000247", result.Timestamp)
				assert.NotNil(t, result.Message)
				assert.Equal(t, "Updated message text", result.Message.Text)
			},
		},
		{
			name: "successful message update with blocks",
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Blocks: []map[string]interface{}{
					{
						"type": "section",
						"text": map[string]interface{}{
							"type": "mrkdwn",
							"text": "*Updated* content with blocks",
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
			slackResponse: MessageResponse{
				OK:      true,
				Channel: "C1234567890",
				TS:      "1503435956.000247",
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
		},
		{
			name: "update with both text and blocks",
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Text:    "Fallback text",
				Blocks: []map[string]interface{}{
					{
						"type": "section",
						"text": map[string]interface{}{
							"type": "plain_text",
							"text": "Block content",
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
			slackResponse: MessageResponse{
				OK:      true,
				Channel: "C1234567890",
				TS:      "1503435956.000247",
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
		},
		{
			name: "missing channel",
			config: UpdateMessageConfig{
				TS:   "1503435956.000247",
				Text: "Updated text",
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
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				Text:    "Updated text",
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
			name: "missing text and blocks",
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
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
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Text:    string(make([]byte, 50000)),
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
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Text:    "Updated text",
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
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Text:    "Updated text",
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
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "9999999999.999999",
				Text:    "Updated text",
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
			name: "cannot update message - not owner",
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Text:    "Updated text",
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
				Error: "cant_update_message",
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "unauthorized",
		},
		{
			name: "edit window closed",
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Text:    "Updated text",
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
				Error: "edit_window_closed",
			},
			slackStatus:   http.StatusOK,
			wantErr:       true,
			errorContains: "action restricted",
		},
		{
			name: "channel not found",
			config: UpdateMessageConfig{
				Channel: "C9999999999",
				TS:      "1503435956.000247",
				Text:    "Updated text",
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
			config: UpdateMessageConfig{
				Channel: "C1234567890",
				TS:      "1503435956.000247",
				Text:    "Updated text",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Slack server
			var server *httptest.Server
			if tt.slackResponse != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/chat.update", r.URL.Path)
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
			action := &UpdateMessageAction{
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

// TestUpdateMessageAction_Execute_ContextCancellation tests context cancellation
func TestUpdateMessageAction_Execute_ContextCancellation(t *testing.T) {
	// Create slow mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{OK: true})
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
	action := &UpdateMessageAction{
		credentialService: mockCred,
		baseURL:           server.URL,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Execute
	config := UpdateMessageConfig{
		Channel: "C1234567890",
		TS:      "1503435956.000247",
		Text:    "Updated text",
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

// TestUpdateMessageAction_NewUpdateMessageAction tests action constructor
func TestUpdateMessageAction_NewUpdateMessageAction(t *testing.T) {
	mockCred := &MockCredentialService{}

	action := NewUpdateMessageAction(mockCred)

	require.NotNil(t, action)
	assert.NotNil(t, action.credentialService)
	assert.Equal(t, DefaultBaseURL, action.baseURL)
}

// TestUpdateMessageResult_JSON tests JSON serialization
func TestUpdateMessageResult_JSON(t *testing.T) {
	result := &UpdateMessageResult{
		OK:        true,
		Channel:   "C1234567890",
		Timestamp: "1503435956.000247",
		Message: &Message{
			Type: "message",
			Text: "Updated text",
			TS:   "1503435956.000247",
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded UpdateMessageResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.OK, decoded.OK)
	assert.Equal(t, result.Channel, decoded.Channel)
	assert.Equal(t, result.Timestamp, decoded.Timestamp)
}

// TestUpdateMessageAction_FromPreviousStepOutput tests using timestamp from previous step
func TestUpdateMessageAction_FromPreviousStepOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{
			OK:      true,
			Channel: "C1234567890",
			TS:      "1503435956.000247",
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

	action := &UpdateMessageAction{
		credentialService: mockCred,
		baseURL:           server.URL,
	}

	config := UpdateMessageConfig{
		Channel: "C1234567890",
		TS:      "1503435956.000247",
		Text:    "Updated from workflow",
	}

	// Simulate context with previous step output
	input := actions.NewActionInput(config, map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
		"credential_id": "cred-123",
		"steps": map[string]interface{}{
			"send-initial": map[string]interface{}{
				"timestamp": "1503435956.000247",
				"channel":   "C1234567890",
			},
		},
	})

	ctx := context.Background()
	output, err := action.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, output)

	result, ok := output.Data.(*UpdateMessageResult)
	assert.True(t, ok)
	assert.True(t, result.OK)
}
