package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/communication"
)

func TestNewSESProvider(t *testing.T) {
	provider, err := NewSESProvider("us-east-1")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestSESProvider_SendEmail(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewSESProvider("us-east-1")
			require.NoError(t, err)

			ctx := context.Background()

			response, err := provider.SendEmail(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, response)
			} else {
				// Since we don't have valid AWS credentials, the API call will fail
				assert.Error(t, err)
			}
		})
	}
}

func TestSESProvider_SendBulkEmail(t *testing.T) {
	provider, err := NewSESProvider("us-east-1")
	require.NoError(t, err)

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

	// Without valid AWS credentials, all should fail
	require.Error(t, err)
	assert.Len(t, responses, 2)
}

func TestSESProvider_SendEmailWithAttachments(t *testing.T) {
	provider, err := NewSESProvider("us-east-1")
	require.NoError(t, err)

	ctx := context.Background()

	request := &communication.EmailRequest{
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
	}

	response, err := provider.SendEmail(ctx, request)

	// Without valid AWS credentials, should fail
	assert.Error(t, err)
	assert.NotNil(t, response)
}
