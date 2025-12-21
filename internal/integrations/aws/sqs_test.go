package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQSClient(t *testing.T) {
	tests := []struct {
		name      string
		accessKey string
		secretKey string
		region    string
		wantErr   bool
	}{
		{
			name:      "valid credentials",
			accessKey: "test-access-key",
			secretKey: "test-secret-key",
			region:    "us-east-1",
			wantErr:   false,
		},
		{
			name:      "missing access key",
			accessKey: "",
			secretKey: "test-secret-key",
			region:    "us-east-1",
			wantErr:   true,
		},
		{
			name:      "missing secret key",
			accessKey: "test-access-key",
			secretKey: "",
			region:    "us-east-1",
			wantErr:   true,
		},
		{
			name:      "missing region",
			accessKey: "test-access-key",
			secretKey: "test-secret-key",
			region:    "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewSQSClient(tt.accessKey, tt.secretKey, tt.region)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestSendMessageAction_Name(t *testing.T) {
	action := NewSendMessageAction("key", "secret", "us-east-1")
	assert.Equal(t, "aws:sqs:send_message", action.Name())
}

func TestSendMessageAction_Description(t *testing.T) {
	action := NewSendMessageAction("key", "secret", "us-east-1")
	assert.NotEmpty(t, action.Description())
}

func TestSendMessageAction_Validate(t *testing.T) {
	action := NewSendMessageAction("key", "secret", "us-east-1")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"queue_url":    "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				"message_body": "test message",
			},
			wantErr: false,
		},
		{
			name: "valid config with delay",
			config: map[string]interface{}{
				"queue_url":        "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				"message_body":     "test message",
				"delay_seconds":    10,
				"message_group_id": "group1",
			},
			wantErr: false,
		},
		{
			name: "valid config with attributes",
			config: map[string]interface{}{
				"queue_url":    "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				"message_body": "test message",
				"attributes": map[string]interface{}{
					"key1": "value1",
				},
			},
			wantErr: false,
		},
		{
			name: "missing queue URL",
			config: map[string]interface{}{
				"message_body": "test message",
			},
			wantErr: true,
		},
		{
			name: "missing message body",
			config: map[string]interface{}{
				"queue_url": "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			config:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := action.Validate(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReceiveMessageAction_Name(t *testing.T) {
	action := NewReceiveMessageAction("key", "secret", "us-east-1")
	assert.Equal(t, "aws:sqs:receive_message", action.Name())
}

func TestReceiveMessageAction_Validate(t *testing.T) {
	action := NewReceiveMessageAction("key", "secret", "us-east-1")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"queue_url": "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			},
			wantErr: false,
		},
		{
			name: "valid config with max messages",
			config: map[string]interface{}{
				"queue_url":          "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				"max_messages":       5,
				"wait_time_seconds":  10,
				"visibility_timeout": 30,
			},
			wantErr: false,
		},
		{
			name:    "missing queue URL",
			config:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := action.Validate(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteMessageAction_Name(t *testing.T) {
	action := NewDeleteMessageAction("key", "secret", "us-east-1")
	assert.Equal(t, "aws:sqs:delete_message", action.Name())
}

func TestDeleteMessageAction_Validate(t *testing.T) {
	action := NewDeleteMessageAction("key", "secret", "us-east-1")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"queue_url":      "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				"receipt_handle": "test-receipt-handle",
			},
			wantErr: false,
		},
		{
			name: "missing queue URL",
			config: map[string]interface{}{
				"receipt_handle": "test-receipt-handle",
			},
			wantErr: true,
		},
		{
			name: "missing receipt handle",
			config: map[string]interface{}{
				"queue_url": "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := action.Validate(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSendMessageConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  SendMessageConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: SendMessageConfig{
				QueueURL:    "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				MessageBody: "test message",
			},
			wantErr: false,
		},
		{
			name: "valid config with FIFO",
			config: SendMessageConfig{
				QueueURL:       "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue.fifo",
				MessageBody:    "test message",
				MessageGroupID: "group1",
			},
			wantErr: false,
		},
		{
			name: "empty queue URL",
			config: SendMessageConfig{
				QueueURL:    "",
				MessageBody: "test message",
			},
			wantErr: true,
		},
		{
			name: "empty message body",
			config: SendMessageConfig{
				QueueURL:    "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				MessageBody: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReceiveMessageConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  ReceiveMessageConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ReceiveMessageConfig{
				QueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
			},
			wantErr: false,
		},
		{
			name: "valid config with options",
			config: ReceiveMessageConfig{
				QueueURL:          "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				MaxMessages:       5,
				WaitTimeSeconds:   10,
				VisibilityTimeout: 30,
			},
			wantErr: false,
		},
		{
			name: "empty queue URL",
			config: ReceiveMessageConfig{
				QueueURL: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteMessageConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  DeleteMessageConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: DeleteMessageConfig{
				QueueURL:      "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				ReceiptHandle: "test-receipt-handle",
			},
			wantErr: false,
		},
		{
			name: "empty queue URL",
			config: DeleteMessageConfig{
				QueueURL:      "",
				ReceiptHandle: "test-receipt-handle",
			},
			wantErr: true,
		},
		{
			name: "empty receipt handle",
			config: DeleteMessageConfig{
				QueueURL:      "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				ReceiptHandle: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSQSClientValidation(t *testing.T) {
	t.Run("validateSQSConfig with valid params", func(t *testing.T) {
		err := validateSQSConfig("key", "secret", "us-east-1")
		assert.NoError(t, err)
	})

	t.Run("validateSQSConfig with empty access key", func(t *testing.T) {
		err := validateSQSConfig("", "secret", "us-east-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access key")
	})

	t.Run("validateSQSConfig with empty secret key", func(t *testing.T) {
		err := validateSQSConfig("key", "", "us-east-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret key")
	})

	t.Run("validateSQSConfig with empty region", func(t *testing.T) {
		err := validateSQSConfig("key", "secret", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "region")
	})
}

func TestSQSActionExecute_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	action := NewSendMessageAction("key", "secret", "us-east-1")
	config := map[string]interface{}{
		"queue_url":    "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		"message_body": "test message",
	}

	_, err := action.Execute(ctx, config, nil)
	require.Error(t, err)
}
