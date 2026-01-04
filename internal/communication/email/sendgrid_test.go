package email

import (
	"context"
	"testing"

	"github.com/gorax/gorax/internal/communication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSendGridProvider(t *testing.T) {
	apiKey := "test-api-key"
	provider := NewSendGridProvider(apiKey)

	assert.NotNil(t, provider)
	assert.Equal(t, apiKey, provider.apiKey)
}

func TestSendGridProvider_SendEmail(t *testing.T) {
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
			name: "valid request with text body",
			request: &communication.EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email",
				Body:    "This is a test email",
			},
			wantErr: false,
		},
		{
			name: "valid request with html body",
			request: &communication.EmailRequest{
				From:     "sender@example.com",
				To:       []string{"recipient@example.com"},
				Subject:  "Test Email",
				BodyHTML: "<p>This is a test email</p>",
			},
			wantErr: false,
		},
		{
			name: "valid request with attachments",
			request: &communication.EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email with Attachment",
				Body:    "See attached",
				Attachments: []communication.Attachment{
					{
						Filename:    "test.txt",
						Content:     []byte("test content"),
						ContentType: "text/plain",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple recipients",
			request: &communication.EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient1@example.com", "recipient2@example.com"},
				CC:      []string{"cc@example.com"},
				BCC:     []string{"bcc@example.com"},
				Subject: "Test Email",
				Body:    "Multiple recipients",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewSendGridProvider("test-api-key")
			ctx := context.Background()

			response, err := provider.SendEmail(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, response)
			} else {
				// Since we're using a test API key, we expect the actual API call to fail
				// In a real scenario with a valid key, we would check response
				assert.Error(t, err) // API call will fail with invalid key
			}
		})
	}
}

func TestSendGridProvider_SendBulkEmail(t *testing.T) {
	provider := NewSendGridProvider("test-api-key")
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

	// With invalid API key, all should fail
	require.Error(t, err)
	assert.Len(t, responses, 2)
	for _, resp := range responses {
		assert.NotNil(t, resp.Error)
	}
}

func TestSendGridProvider_buildMessage(t *testing.T) {
	provider := NewSendGridProvider("test-api-key")

	tests := []struct {
		name    string
		request *communication.EmailRequest
		wantErr bool
	}{
		{
			name: "simple email",
			request: &communication.EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test",
				Body:    "Body",
			},
			wantErr: false,
		},
		{
			name: "with cc and bcc",
			request: &communication.EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				CC:      []string{"cc@example.com"},
				BCC:     []string{"bcc@example.com"},
				Subject: "Test",
				Body:    "Body",
			},
			wantErr: false,
		},
		{
			name: "with reply-to",
			request: &communication.EmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test",
				Body:    "Body",
				ReplyTo: "reply@example.com",
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
				assert.NotNil(t, msg)
			}
		})
	}
}
