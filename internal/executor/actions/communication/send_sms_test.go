package communication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendSMSAction_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SendSMSConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: SendSMSConfig{
				Provider:     "twilio",
				From:         "+1234567890",
				To:           "+0987654321",
				Message:      "Test message",
				CredentialID: "cred-123",
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			config: SendSMSConfig{
				From:         "+1234567890",
				To:           "+0987654321",
				Message:      "Test message",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "missing from",
			config: SendSMSConfig{
				Provider:     "twilio",
				To:           "+0987654321",
				Message:      "Test message",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "from number is required",
		},
		{
			name: "missing to",
			config: SendSMSConfig{
				Provider:     "twilio",
				From:         "+1234567890",
				Message:      "Test message",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "to number is required",
		},
		{
			name: "missing message",
			config: SendSMSConfig{
				Provider:     "twilio",
				From:         "+1234567890",
				To:           "+0987654321",
				CredentialID: "cred-123",
			},
			wantErr: true,
			errMsg:  "message is required",
		},
		{
			name: "missing credential ID",
			config: SendSMSConfig{
				Provider: "twilio",
				From:     "+1234567890",
				To:       "+0987654321",
				Message:  "Test message",
			},
			wantErr: true,
			errMsg:  "credential_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewSendSMSAction(tt.config, nil)
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

func TestSendSMSAction_Name(t *testing.T) {
	action := NewSendSMSAction(SendSMSConfig{}, nil)
	assert.Equal(t, "send_sms", action.Name())
}
