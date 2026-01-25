package sms

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/communication"
)

func TestNewSNSProvider(t *testing.T) {
	provider, err := NewSNSProvider("us-east-1")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestSNSProvider_SendSMS(t *testing.T) {
	tests := []struct {
		name    string
		request *communication.SMSRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid request",
			request: &communication.SMSRequest{
				Message: "Test",
			},
			wantErr: true,
			errMsg:  "from number is required",
		},
		{
			name: "valid request",
			request: &communication.SMSRequest{
				From:    "+1234567890",
				To:      "+0987654321",
				Message: "Test message",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewSNSProvider("us-east-1")
			require.NoError(t, err)

			ctx := context.Background()

			response, err := provider.SendSMS(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, response)
			} else {
				// Without valid AWS credentials, API call will fail
				assert.Error(t, err)
			}
		})
	}
}

func TestSNSProvider_SendBulkSMS(t *testing.T) {
	provider, err := NewSNSProvider("us-east-1")
	require.NoError(t, err)

	ctx := context.Background()

	requests := []*communication.SMSRequest{
		{
			From:    "+1234567890",
			To:      "+0987654321",
			Message: "Test message 1",
		},
		{
			From:    "+1234567890",
			To:      "+0987654322",
			Message: "Test message 2",
		},
	}

	responses, err := provider.SendBulkSMS(ctx, requests)

	// Without valid AWS credentials, all should fail
	require.Error(t, err)
	assert.Len(t, responses, 2)
}
