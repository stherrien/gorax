package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewS3Client(t *testing.T) {
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
			client, err := NewS3Client(tt.accessKey, tt.secretKey, tt.region)
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

func TestListBucketsAction_Name(t *testing.T) {
	action := NewListBucketsAction("key", "secret", "us-east-1")
	assert.Equal(t, "aws:s3:list_buckets", action.Name())
}

func TestListBucketsAction_Description(t *testing.T) {
	action := NewListBucketsAction("key", "secret", "us-east-1")
	assert.NotEmpty(t, action.Description())
}

func TestListBucketsAction_Validate(t *testing.T) {
	action := NewListBucketsAction("key", "secret", "us-east-1")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "empty config is valid",
			config:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "nil config is valid",
			config:  nil,
			wantErr: false,
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

func TestGetObjectAction_Name(t *testing.T) {
	action := NewGetObjectAction("key", "secret", "us-east-1")
	assert.Equal(t, "aws:s3:get_object", action.Name())
}

func TestGetObjectAction_Validate(t *testing.T) {
	action := NewGetObjectAction("key", "secret", "us-east-1")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"bucket": "test-bucket",
				"key":    "test-key",
			},
			wantErr: false,
		},
		{
			name: "missing bucket",
			config: map[string]interface{}{
				"key": "test-key",
			},
			wantErr: true,
		},
		{
			name: "missing key",
			config: map[string]interface{}{
				"bucket": "test-bucket",
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

func TestPutObjectAction_Name(t *testing.T) {
	action := NewPutObjectAction("key", "secret", "us-east-1")
	assert.Equal(t, "aws:s3:put_object", action.Name())
}

func TestPutObjectAction_Validate(t *testing.T) {
	action := NewPutObjectAction("key", "secret", "us-east-1")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"bucket": "test-bucket",
				"key":    "test-key",
				"body":   "test content",
			},
			wantErr: false,
		},
		{
			name: "valid config with content_type",
			config: map[string]interface{}{
				"bucket":       "test-bucket",
				"key":          "test-key",
				"body":         "test content",
				"content_type": "text/plain",
			},
			wantErr: false,
		},
		{
			name: "missing bucket",
			config: map[string]interface{}{
				"key":  "test-key",
				"body": "test content",
			},
			wantErr: true,
		},
		{
			name: "missing key",
			config: map[string]interface{}{
				"bucket": "test-bucket",
				"body":   "test content",
			},
			wantErr: true,
		},
		{
			name: "missing body",
			config: map[string]interface{}{
				"bucket": "test-bucket",
				"key":    "test-key",
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

func TestDeleteObjectAction_Name(t *testing.T) {
	action := NewDeleteObjectAction("key", "secret", "us-east-1")
	assert.Equal(t, "aws:s3:delete_object", action.Name())
}

func TestDeleteObjectAction_Validate(t *testing.T) {
	action := NewDeleteObjectAction("key", "secret", "us-east-1")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"bucket": "test-bucket",
				"key":    "test-key",
			},
			wantErr: false,
		},
		{
			name: "missing bucket",
			config: map[string]interface{}{
				"key": "test-key",
			},
			wantErr: true,
		},
		{
			name: "missing key",
			config: map[string]interface{}{
				"bucket": "test-bucket",
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

func TestS3ConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  GetObjectConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: GetObjectConfig{
				Bucket: "test-bucket",
				Key:    "test-key",
			},
			wantErr: false,
		},
		{
			name: "empty bucket",
			config: GetObjectConfig{
				Bucket: "",
				Key:    "test-key",
			},
			wantErr: true,
		},
		{
			name: "empty key",
			config: GetObjectConfig{
				Bucket: "test-bucket",
				Key:    "",
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

func TestPutObjectConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  PutObjectConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: PutObjectConfig{
				Bucket: "test-bucket",
				Key:    "test-key",
				Body:   "content",
			},
			wantErr: false,
		},
		{
			name: "valid config with content type",
			config: PutObjectConfig{
				Bucket:      "test-bucket",
				Key:         "test-key",
				Body:        "content",
				ContentType: "text/plain",
			},
			wantErr: false,
		},
		{
			name: "empty bucket",
			config: PutObjectConfig{
				Bucket: "",
				Key:    "test-key",
				Body:   "content",
			},
			wantErr: true,
		},
		{
			name: "empty key",
			config: PutObjectConfig{
				Bucket: "test-bucket",
				Key:    "",
				Body:   "content",
			},
			wantErr: true,
		},
		{
			name: "empty body",
			config: PutObjectConfig{
				Bucket: "test-bucket",
				Key:    "test-key",
				Body:   "",
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

func TestS3ClientValidation(t *testing.T) {
	t.Run("validateS3Config with valid params", func(t *testing.T) {
		err := validateS3Config("key", "secret", "us-east-1")
		assert.NoError(t, err)
	})

	t.Run("validateS3Config with empty access key", func(t *testing.T) {
		err := validateS3Config("", "secret", "us-east-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access key")
	})

	t.Run("validateS3Config with empty secret key", func(t *testing.T) {
		err := validateS3Config("key", "", "us-east-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret key")
	})

	t.Run("validateS3Config with empty region", func(t *testing.T) {
		err := validateS3Config("key", "secret", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "region")
	})
}

func TestS3ActionExecute_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	action := NewGetObjectAction("key", "secret", "us-east-1")
	config := map[string]interface{}{
		"bucket": "test-bucket",
		"key":    "test-key",
	}

	_, err := action.Execute(ctx, config, nil)
	require.Error(t, err)
}
