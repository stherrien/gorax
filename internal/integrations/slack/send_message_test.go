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

// MockCredentialService implements credential.Service for testing
type MockCredentialService struct {
	GetValueFunc func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error)
}

func (m *MockCredentialService) GetValue(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
	if m.GetValueFunc != nil {
		return m.GetValueFunc(ctx, tenantID, credentialID, userID)
	}
	return nil, errors.New("not implemented")
}

// Implement other methods to satisfy interface
func (m *MockCredentialService) Create(ctx context.Context, tenantID, userID string, input credential.CreateCredentialInput) (*credential.Credential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) List(ctx context.Context, tenantID string, filter credential.CredentialListFilter, limit, offset int) ([]*credential.Credential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) GetByID(ctx context.Context, tenantID, credentialID string) (*credential.Credential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) Update(ctx context.Context, tenantID, credentialID, userID string, input credential.UpdateCredentialInput) (*credential.Credential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) Delete(ctx context.Context, tenantID, credentialID, userID string) error {
	return errors.New("not implemented")
}

func (m *MockCredentialService) Rotate(ctx context.Context, tenantID, credentialID, userID string, input credential.RotateCredentialInput) (*credential.Credential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) ListVersions(ctx context.Context, tenantID, credentialID string) ([]*credential.CredentialValue, error) {
	return nil, errors.New("not implemented")
}

func (m *MockCredentialService) GetAccessLog(ctx context.Context, tenantID, credentialID string, limit, offset int) ([]*credential.AccessLog, error) {
	return nil, errors.New("not implemented")
}

// TestSendMessageAction_Execute tests the SendMessage action
func TestSendMessageAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         SendMessageConfig
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
			name: "successful message send",
			config: SendMessageConfig{
				Channel: "C1234567890",
				Text:    "Hello, world!",
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
					Text: "Hello, world!",
					TS:   "1503435956.000247",
				},
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				assert.NotNil(t, output)
				assert.NotNil(t, output.Data)

				result, ok := output.Data.(*SendMessageResult)
				assert.True(t, ok, "output.Data should be *SendMessageResult")
				assert.True(t, result.OK)
				assert.Equal(t, "C1234567890", result.Channel)
				assert.Equal(t, "1503435956.000247", result.Timestamp)
				assert.NotNil(t, result.Message)
				assert.Equal(t, "Hello, world!", result.Message.Text)
			},
		},
		{
			name: "successful message send with blocks",
			config: SendMessageConfig{
				Channel: "C1234567890",
				Blocks: []map[string]interface{}{
					{
						"type": "section",
						"text": map[string]interface{}{
							"type": "mrkdwn",
							"text": "*Hello* from blocks!",
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
			config: SendMessageConfig{
				Text: "Hello",
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
			name: "missing text and blocks",
			config: SendMessageConfig{
				Channel: "C1234567890",
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
			config: SendMessageConfig{
				Channel: "C1234567890",
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
			config: SendMessageConfig{
				Channel: "C1234567890",
				Text:    "Hello",
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
			config: SendMessageConfig{
				Channel: "C1234567890",
				Text:    "Hello",
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
			name: "invalid access token in credential",
			config: SendMessageConfig{
				Channel: "C1234567890",
				Text:    "Hello",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					// Missing access_token
					"something_else": "value",
				},
			},
			wantErr:       true,
			errorContains: "access_token not found",
		},
		{
			name: "slack api error - channel not found",
			config: SendMessageConfig{
				Channel: "C9999999999",
				Text:    "Hello",
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
			name: "slack api error - invalid auth",
			config: SendMessageConfig{
				Channel: "C1234567890",
				Text:    "Hello",
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
			name: "successful send with thread",
			config: SendMessageConfig{
				Channel:  "C1234567890",
				Text:     "Reply in thread",
				ThreadTS: "1503435956.000247",
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
				TS:      "1503435999.000248",
			},
			slackStatus: http.StatusOK,
			wantErr:     false,
		},
		{
			name: "successful send with custom username and emoji",
			config: SendMessageConfig{
				Channel:   "C1234567890",
				Text:      "Custom bot message",
				Username:  "Custom Bot",
				IconEmoji: ":robot_face:",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Slack server
			var server *httptest.Server
			if tt.slackResponse != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/chat.postMessage", r.URL.Path)
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
			action := &SendMessageAction{
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

// TestSendMessageAction_Execute_ContextCancellation tests context cancellation
func TestSendMessageAction_Execute_ContextCancellation(t *testing.T) {
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
	action := &SendMessageAction{
		credentialService: mockCred,
		baseURL:           server.URL,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Execute
	config := SendMessageConfig{
		Channel: "C1234567890",
		Text:    "Hello",
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

// TestSendMessageAction_NewSendMessageAction tests action constructor
func TestSendMessageAction_NewSendMessageAction(t *testing.T) {
	mockCred := &MockCredentialService{}

	action := NewSendMessageAction(mockCred)

	require.NotNil(t, action)
	assert.NotNil(t, action.credentialService)
	assert.Equal(t, DefaultBaseURL, action.baseURL)
}

// TestSendMessageResult_JSON tests JSON serialization
func TestSendMessageResult_JSON(t *testing.T) {
	result := &SendMessageResult{
		OK:        true,
		Channel:   "C1234567890",
		Timestamp: "1503435956.000247",
		Message: &Message{
			Type: "message",
			Text: "Hello",
			TS:   "1503435956.000247",
		},
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded SendMessageResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, result.OK, decoded.OK)
	assert.Equal(t, result.Channel, decoded.Channel)
	assert.Equal(t, result.Timestamp, decoded.Timestamp)
}
