package sms

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/communication"
)

func TestNewTwilioProvider(t *testing.T) {
	provider := NewTwilioProvider("test-sid", "test-token")
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.client)
}

func TestTwilioProvider_SendSMS(t *testing.T) {
	tests := []struct {
		name    string
		request *communication.SMSRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid request - missing from",
			request: &communication.SMSRequest{
				To:      "+1234567890",
				Message: "Test message",
			},
			wantErr: true,
			errMsg:  "from number is required",
		},
		{
			name: "invalid request - missing to",
			request: &communication.SMSRequest{
				From:    "+1234567890",
				Message: "Test message",
			},
			wantErr: true,
			errMsg:  "to number is required",
		},
		{
			name: "invalid request - missing message",
			request: &communication.SMSRequest{
				From: "+1234567890",
				To:   "+0987654321",
			},
			wantErr: true,
			errMsg:  "message is required",
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
			provider := NewTwilioProvider("test-sid", "test-token")
			ctx := context.Background()

			response, err := provider.SendSMS(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, response)
			} else {
				// With invalid credentials, API call will fail
				assert.Error(t, err)
			}
		})
	}
}

func TestTwilioProvider_SendBulkSMS(t *testing.T) {
	provider := NewTwilioProvider("test-sid", "test-token")
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

	// With invalid credentials, all should fail
	require.Error(t, err)
	assert.Len(t, responses, 2)
	for _, resp := range responses {
		assert.NotNil(t, resp.Error)
	}
}
