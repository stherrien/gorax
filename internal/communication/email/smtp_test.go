package email

import (
	"context"
	"testing"

	"github.com/gorax/gorax/internal/communication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSMTPProvider(t *testing.T) {
	provider := NewSMTPProvider("smtp.example.com", 587, "user", "pass", true)

	assert.NotNil(t, provider)
	assert.Equal(t, "smtp.example.com", provider.host)
	assert.Equal(t, 587, provider.port)
	assert.Equal(t, "user", provider.username)
	assert.Equal(t, "pass", provider.password)
	assert.True(t, provider.useTLS)
}

func TestSMTPProvider_SendEmail(t *testing.T) {
	tests := []struct {
		name    string
		request *communication.EmailRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid request",
			request: &communication.EmailRequest{
				Subject: "Test",
			},
			wantErr: true,
			errMsg:  "from address is required",
		},
		{
			name: "valid request",
			request: &communication.EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email",
				Body:    "This is a test email",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewSMTPProvider("smtp.example.com", 587, "user", "pass", true)
			ctx := context.Background()

			response, err := provider.SendEmail(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, response)
			} else {
				// Without a real SMTP server, the connection will fail
				assert.Error(t, err)
			}
		})
	}
}

func TestSMTPProvider_BuildMessage(t *testing.T) {
	provider := NewSMTPProvider("smtp.example.com", 587, "user", "pass", true)

	tests := []struct {
		name    string
		request *communication.EmailRequest
		wantErr bool
	}{
		{
			name: "simple text email",
			request: &communication.EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test",
				Body:    "Body",
			},
			wantErr: false,
		},
		{
			name: "html email",
			request: &communication.EmailRequest{
				From:     "sender@example.com",
				To:       []string{"recipient@example.com"},
				Subject:  "Test",
				BodyHTML: "<p>Body</p>",
			},
			wantErr: false,
		},
		{
			name: "email with attachment",
			request: &communication.EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test",
				Body:    "Body",
				Attachments: []communication.Attachment{
					{
						Filename:    "test.txt",
						Content:     []byte("test"),
						ContentType: "text/plain",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "email with multiple recipients",
			request: &communication.EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient1@example.com", "recipient2@example.com"},
				CC:      []string{"cc@example.com"},
				Subject: "Test",
				Body:    "Body",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := provider.buildMessage(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, msg)
				// Verify message contains key elements
				msgStr := string(msg)
				assert.Contains(t, msgStr, "From:")
				assert.Contains(t, msgStr, "To:")
				assert.Contains(t, msgStr, "Subject:")
			}
		})
	}
}

func TestSMTPProvider_SendBulkEmail(t *testing.T) {
	provider := NewSMTPProvider("smtp.example.com", 587, "user", "pass", true)
	ctx := context.Background()

	requests := []*communication.EmailRequest{
		{
			From:    "sender@example.com",
			To:      []string{"recipient1@example.com"},
			Subject: "Test Email 1",
			Body:    "Body 1",
		},
		{
			From:    "sender@example.com",
			To:      []string{"recipient2@example.com"},
			Subject: "Test Email 2",
			Body:    "Body 2",
		},
	}

	responses, err := provider.SendBulkEmail(ctx, requests)

	// Without a real SMTP server, all should fail
	require.Error(t, err)
	assert.Len(t, responses, 2)
}
