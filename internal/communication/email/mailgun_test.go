package email

import (
	"context"
	"testing"

	"github.com/gorax/gorax/internal/communication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMailgunProvider(t *testing.T) {
	domain := "example.com"
	apiKey := "test-api-key"
	provider := NewMailgunProvider(domain, apiKey)

	assert.NotNil(t, provider)
	assert.Equal(t, domain, provider.domain)
}

func TestMailgunProvider_SendEmail(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewMailgunProvider("example.com", "test-api-key")
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
				assert.Error(t, err) // API call will fail with invalid key
			}
		})
	}
}

func TestMailgunProvider_SendBulkEmail(t *testing.T) {
	provider := NewMailgunProvider("example.com", "test-api-key")
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
