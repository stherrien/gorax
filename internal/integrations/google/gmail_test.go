package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGmailSendAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         GmailSendConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		credError      error
		gmailResponse  interface{}
		gmailStatus    int
		wantErr        bool
		errorContains  string
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful email send",
			config: GmailSendConfig{
				To:      "recipient@example.com",
				Subject: "Test Email",
				Body:    "This is a test email",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "test-access-token",
				},
			},
			gmailResponse: map[string]interface{}{
				"id":       "msg123",
				"threadId": "thread123",
			},
			gmailStatus: http.StatusOK,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				assert.NotNil(t, output)
				result, ok := output.Data.(*GmailSendResult)
				assert.True(t, ok)
				assert.Equal(t, "msg123", result.MessageID)
				assert.Equal(t, "thread123", result.ThreadID)
			},
		},
		{
			name: "email with CC and BCC",
			config: GmailSendConfig{
				To:      "recipient@example.com",
				Cc:      "cc@example.com",
				Bcc:     "bcc@example.com",
				Subject: "Test Email",
				Body:    "This is a test email",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "test-access-token",
				},
			},
			gmailResponse: map[string]interface{}{
				"id":       "msg456",
				"threadId": "thread456",
			},
			gmailStatus: http.StatusOK,
			wantErr:     false,
		},
		{
			name: "missing recipient",
			config: GmailSendConfig{
				Subject: "Test Email",
				Body:    "This is a test email",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "to is required",
		},
		{
			name: "missing subject",
			config: GmailSendConfig{
				To:   "recipient@example.com",
				Body: "This is a test email",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "subject is required",
		},
		{
			name: "missing body",
			config: GmailSendConfig{
				To:      "recipient@example.com",
				Subject: "Test Email",
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			wantErr:       true,
			errorContains: "body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Gmail server
			var server *httptest.Server
			if tt.gmailResponse != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.gmailStatus)
					json.NewEncoder(w).Encode(tt.gmailResponse)
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
			action := NewGmailSendAction(mockCred)
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

func TestGmailReadAction_Execute(t *testing.T) {
	tests := []struct {
		name           string
		config         GmailReadConfig
		context        map[string]interface{}
		mockCredential *credential.DecryptedValue
		gmailResponse  interface{}
		gmailStatus    int
		wantErr        bool
		validate       func(t *testing.T, output *actions.ActionOutput)
	}{
		{
			name: "successful email read",
			config: GmailReadConfig{
				Query:      "from:test@example.com",
				MaxResults: 10,
			},
			context: map[string]interface{}{
				"env": map[string]interface{}{
					"tenant_id": "tenant-123",
				},
				"credential_id": "cred-123",
			},
			mockCredential: &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "test-access-token",
				},
			},
			gmailResponse: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"id":       "msg1",
						"threadId": "thread1",
					},
				},
			},
			gmailStatus: http.StatusOK,
			wantErr:     false,
			validate: func(t *testing.T, output *actions.ActionOutput) {
				assert.NotNil(t, output)
				result, ok := output.Data.(*GmailReadResult)
				assert.True(t, ok)
				assert.Len(t, result.Messages, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.gmailResponse != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.gmailStatus)
					json.NewEncoder(w).Encode(tt.gmailResponse)
				}))
				defer server.Close()
			}

			mockCred := &MockCredentialService{
				GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
					return tt.mockCredential, nil
				},
			}

			action := NewGmailReadAction(mockCred)
			if server != nil {
				action.baseURL = server.URL
			}

			input := actions.NewActionInput(tt.config, tt.context)
			ctx := context.Background()
			output, err := action.Execute(ctx, input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

func TestGmailSendConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    GmailSendConfig
		wantErr   bool
		errString string
	}{
		{
			name: "valid config",
			config: GmailSendConfig{
				To:      "test@example.com",
				Subject: "Test",
				Body:    "Test body",
			},
			wantErr: false,
		},
		{
			name: "missing to",
			config: GmailSendConfig{
				Subject: "Test",
				Body:    "Test body",
			},
			wantErr:   true,
			errString: "to is required",
		},
		{
			name: "missing subject",
			config: GmailSendConfig{
				To:   "test@example.com",
				Body: "Test body",
			},
			wantErr:   true,
			errString: "subject is required",
		},
		{
			name: "missing body",
			config: GmailSendConfig{
				To:      "test@example.com",
				Subject: "Test",
			},
			wantErr:   true,
			errString: "body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
