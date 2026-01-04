package communication

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/gorax/gorax/internal/credential"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCredentialService is a mock for credential.Service
type MockCredentialService struct {
	mock.Mock
}

func (m *MockCredentialService) GetDecryptedValue(ctx context.Context, id string) (*credential.DecryptedValue, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*credential.DecryptedValue), args.Error(1)
}

func TestSendEmailAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SendEmailConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: SendEmailConfig{
				Provider:     "sendgrid",
				From:         "sender@example.com",
				To:           []string{"recipient@example.com"},
				Subject:      "Test",
				Body:         "Body",
				CredentialID: "cred-123",
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			config: SendEmailConfig{
				From:         "sender@example.com",
				To:           []string{"recipient@example.com"},
				Subject:      "Test",
				Body:         "Body",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "missing from",
			config: SendEmailConfig{
				Provider:     "sendgrid",
				To:           []string{"recipient@example.com"},
				Subject:      "Test",
				Body:         "Body",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "from address is required",
		},
		{
			name: "missing recipients",
			config: SendEmailConfig{
				Provider:     "sendgrid",
				From:         "sender@example.com",
				Subject:      "Test",
				Body:         "Body",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "at least one recipient is required",
		},
		{
			name: "missing subject",
			config: SendEmailConfig{
				Provider:     "sendgrid",
				From:         "sender@example.com",
				To:           []string{"recipient@example.com"},
				Body:         "Body",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "subject is required",
		},
		{
			name: "missing body",
			config: SendEmailConfig{
				Provider:     "sendgrid",
				From:         "sender@example.com",
				To:           []string{"recipient@example.com"},
				Subject:      "Test",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "email body is required",
		},
		{
			name: "missing credential ID",
			config: SendEmailConfig{
				Provider: "sendgrid",
				From:     "sender@example.com",
				To:       []string{"recipient@example.com"},
				Subject:  "Test",
				Body:     "Body",
			},
			wantErr: true,
			errMsg:  "credential_id is required",
		},
		{
			name: "smtp without config",
			config: SendEmailConfig{
				Provider:     "smtp",
				From:         "sender@example.com",
				To:           []string{"recipient@example.com"},
				Subject:      "Test",
				Body:         "Body",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "smtp_config is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewSendEmailAction(tt.config, nil)
			err := action.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSendEmailAction_BuildRequest(t *testing.T) {
	tests := []struct {
		name    string
		config  SendEmailConfig
		wantErr bool
	}{
		{
			name: "simple email",
			config: SendEmailConfig{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test",
				Body:    "Body",
			},
			wantErr: false,
		},
		{
			name: "email with attachments",
			config: SendEmailConfig{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test",
				Body:    "Body",
				Attachments: []AttachmentConfig{
					{
						Filename:    "test.txt",
						Content:     base64.StdEncoding.EncodeToString([]byte("test content")),
						ContentType: "text/plain",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "email with invalid attachment",
			config: SendEmailConfig{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test",
				Body:    "Body",
				Attachments: []AttachmentConfig{
					{
						Filename:    "test.txt",
						Content:     "invalid-base64!!!",
						ContentType: "text/plain",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewSendEmailAction(tt.config, nil)
			request, err := action.buildRequest()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, request)
				assert.Equal(t, tt.config.From, request.From)
				assert.Equal(t, tt.config.To, request.To)
			}
		})
	}
}

func TestSendEmailAction_Name(t *testing.T) {
	action := NewSendEmailAction(SendEmailConfig{}, nil)
	assert.Equal(t, "send_email", action.Name())
}
