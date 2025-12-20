package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSNSClient(t *testing.T) {
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
			client, err := NewSNSClient(tt.accessKey, tt.secretKey, tt.region)
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

func TestPublishMessageAction_Name(t *testing.T) {
	action := NewPublishMessageAction("key", "secret", "us-east-1")
	assert.Equal(t, "aws:sns:publish", action.Name())
}

func TestPublishMessageAction_Description(t *testing.T) {
	action := NewPublishMessageAction("key", "secret", "us-east-1")
	assert.NotEmpty(t, action.Description())
}

func TestPublishMessageAction_Validate(t *testing.T) {
	action := NewPublishMessageAction("key", "secret", "us-east-1")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config with topic ARN",
			config: map[string]interface{}{
				"topic_arn": "arn:aws:sns:us-east-1:123456789012:test-topic",
				"message":   "test message",
			},
			wantErr: false,
		},
		{
			name: "valid config with target ARN",
			config: map[string]interface{}{
				"target_arn": "arn:aws:sns:us-east-1:123456789012:endpoint/test",
				"message":    "test message",
			},
			wantErr: false,
		},
		{
			name: "valid config with subject",
			config: map[string]interface{}{
				"topic_arn": "arn:aws:sns:us-east-1:123456789012:test-topic",
				"message":   "test message",
				"subject":   "Test Subject",
			},
			wantErr: false,
		},
		{
			name: "valid config with attributes",
			config: map[string]interface{}{
				"topic_arn": "arn:aws:sns:us-east-1:123456789012:test-topic",
				"message":   "test message",
				"attributes": map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			wantErr: false,
		},
		{
			name: "missing topic and target ARN",
			config: map[string]interface{}{
				"message": "test message",
			},
			wantErr: true,
		},
		{
			name: "missing message",
			config: map[string]interface{}{
				"topic_arn": "arn:aws:sns:us-east-1:123456789012:test-topic",
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

func TestPublishMessageConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  PublishMessageConfig
		wantErr bool
	}{
		{
			name: "valid config with topic ARN",
			config: PublishMessageConfig{
				TopicARN: "arn:aws:sns:us-east-1:123456789012:test-topic",
				Message:  "test message",
			},
			wantErr: false,
		},
		{
			name: "valid config with target ARN",
			config: PublishMessageConfig{
				TargetARN: "arn:aws:sns:us-east-1:123456789012:endpoint/test",
				Message:   "test message",
			},
			wantErr: false,
		},
		{
			name: "valid config with both ARNs (topic takes precedence)",
			config: PublishMessageConfig{
				TopicARN:  "arn:aws:sns:us-east-1:123456789012:test-topic",
				TargetARN: "arn:aws:sns:us-east-1:123456789012:endpoint/test",
				Message:   "test message",
			},
			wantErr: false,
		},
		{
			name: "valid config with attributes",
			config: PublishMessageConfig{
				TopicARN: "arn:aws:sns:us-east-1:123456789012:test-topic",
				Message:  "test message",
				Attributes: map[string]string{
					"key": "value",
				},
			},
			wantErr: false,
		},
		{
			name: "missing both ARNs",
			config: PublishMessageConfig{
				Message: "test message",
			},
			wantErr: true,
		},
		{
			name: "missing message",
			config: PublishMessageConfig{
				TopicARN: "arn:aws:sns:us-east-1:123456789012:test-topic",
			},
			wantErr: true,
		},
		{
			name: "empty message",
			config: PublishMessageConfig{
				TopicARN: "arn:aws:sns:us-east-1:123456789012:test-topic",
				Message:  "",
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

func TestSNSClientValidation(t *testing.T) {
	t.Run("validateSNSConfig with valid params", func(t *testing.T) {
		err := validateSNSConfig("key", "secret", "us-east-1")
		assert.NoError(t, err)
	})

	t.Run("validateSNSConfig with empty access key", func(t *testing.T) {
		err := validateSNSConfig("", "secret", "us-east-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access key")
	})

	t.Run("validateSNSConfig with empty secret key", func(t *testing.T) {
		err := validateSNSConfig("key", "", "us-east-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret key")
	})

	t.Run("validateSNSConfig with empty region", func(t *testing.T) {
		err := validateSNSConfig("key", "secret", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "region")
	})
}

func TestSNSActionExecute_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	action := NewPublishMessageAction("key", "secret", "us-east-1")
	config := map[string]interface{}{
		"topic_arn": "arn:aws:sns:us-east-1:123456789012:test-topic",
		"message":   "test message",
	}

	_, err := action.Execute(ctx, config, nil)
	require.Error(t, err)
}
