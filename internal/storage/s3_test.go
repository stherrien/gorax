package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewS3Storage(t *testing.T) {
	tests := []struct {
		name      string
		region    string
		accessKey string
		secretKey string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid configuration",
			region:    "us-east-1",
			accessKey: "test-access-key",
			secretKey: "test-secret-key",
			wantErr:   false,
		},
		{
			name:      "empty region",
			region:    "",
			accessKey: "test-access-key",
			secretKey: "test-secret-key",
			wantErr:   true,
			errMsg:    "region is required",
		},
		{
			name:      "empty access key",
			region:    "us-east-1",
			accessKey: "",
			secretKey: "test-secret-key",
			wantErr:   true,
			errMsg:    "access key is required",
		},
		{
			name:      "empty secret key",
			region:    "us-east-1",
			accessKey: "test-access-key",
			secretKey: "",
			wantErr:   true,
			errMsg:    "secret key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewS3Storage(tt.region, tt.accessKey, tt.secretKey)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, storage)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, storage)
				assert.Equal(t, tt.region, storage.region)
				err = storage.Close()
				assert.NoError(t, err)
			}
		})
	}
}

func TestS3Storage_ValidateBucket(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	tests := []struct {
		name    string
		bucket  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid bucket",
			bucket:  "my-bucket",
			wantErr: false,
		},
		{
			name:    "valid bucket with numbers",
			bucket:  "my-bucket-123",
			wantErr: false,
		},
		{
			name:    "empty bucket",
			bucket:  "",
			wantErr: true,
			errMsg:  "bucket name cannot be empty",
		},
		{
			name:    "bucket too short",
			bucket:  "ab",
			wantErr: true,
			errMsg:  "bucket name must be between 3 and 63 characters",
		},
		{
			name:    "bucket too long",
			bucket:  strings.Repeat("a", 64),
			wantErr: true,
			errMsg:  "bucket name must be between 3 and 63 characters",
		},
		{
			name:    "invalid characters",
			bucket:  "my_bucket",
			wantErr: true,
			errMsg:  "bucket name contains invalid characters",
		},
		{
			name:    "uppercase letters",
			bucket:  "MyBucket",
			wantErr: true,
			errMsg:  "bucket name contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s3.validateBucket(tt.bucket)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestS3Storage_ValidateKey(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	tests := []struct {
		name    string
		key     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid key",
			key:     "path/to/file.txt",
			wantErr: false,
		},
		{
			name:    "valid key with spaces",
			key:     "path/to/my file.txt",
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
			errMsg:  "key cannot be empty",
		},
		{
			name:    "key too long",
			key:     strings.Repeat("a", 1025),
			wantErr: true,
			errMsg:  "key must not exceed 1024 characters",
		},
		{
			name:    "path traversal attempt",
			key:     "../../../etc/passwd",
			wantErr: true,
			errMsg:  "key contains path traversal",
		},
		{
			name:    "starts with slash",
			key:     "/path/to/file.txt",
			wantErr: true,
			errMsg:  "key must not start with /",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s3.validateKey(tt.key)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestS3Storage_ValidateUploadOptions(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	tests := []struct {
		name    string
		options *UploadOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil options",
			options: nil,
			wantErr: false,
		},
		{
			name: "valid options",
			options: &UploadOptions{
				ContentType:  "text/plain",
				CacheControl: "max-age=3600",
				Metadata:     map[string]string{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "too many metadata entries",
			options: &UploadOptions{
				Metadata: func() map[string]string {
					m := make(map[string]string)
					for i := 0; i < 101; i++ {
						m[string(rune(i))] = "value"
					}
					return m
				}(),
			},
			wantErr: true,
			errMsg:  "metadata cannot exceed 100 entries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s3.validateUploadOptions(tt.options)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestS3Storage_ValidatePresignedExpiration(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	tests := []struct {
		name       string
		expiration time.Duration
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid expiration - 1 hour",
			expiration: 1 * time.Hour,
			wantErr:    false,
		},
		{
			name:       "valid expiration - 7 days",
			expiration: 7 * 24 * time.Hour,
			wantErr:    false,
		},
		{
			name:       "zero expiration",
			expiration: 0,
			wantErr:    true,
			errMsg:     "expiration must be positive",
		},
		{
			name:       "negative expiration",
			expiration: -1 * time.Hour,
			wantErr:    true,
			errMsg:     "expiration must be positive",
		},
		{
			name:       "expiration too long",
			expiration: 8 * 24 * time.Hour,
			wantErr:    true,
			errMsg:     "expiration cannot exceed 7 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s3.validatePresignedExpiration(tt.expiration)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestS3Storage_ValidateListOptions(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	tests := []struct {
		name    string
		options *ListOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil options",
			options: nil,
			wantErr: false,
		},
		{
			name: "valid options",
			options: &ListOptions{
				MaxKeys:   100,
				Delimiter: "/",
			},
			wantErr: false,
		},
		{
			name: "zero max keys",
			options: &ListOptions{
				MaxKeys: 0,
			},
			wantErr: false,
		},
		{
			name: "negative max keys",
			options: &ListOptions{
				MaxKeys: -1,
			},
			wantErr: true,
			errMsg:  "max keys must be positive",
		},
		{
			name: "max keys too large",
			options: &ListOptions{
				MaxKeys: 1001,
			},
			wantErr: true,
			errMsg:  "max keys cannot exceed 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s3.validateListOptions(tt.options)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Mock tests for operations (actual integration tests would use LocalStack)
func TestS3Storage_UploadValidation(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	ctx := context.Background()
	content := bytes.NewReader([]byte("test content"))

	tests := []struct {
		name    string
		bucket  string
		key     string
		wantErr bool
	}{
		{
			name:    "valid parameters",
			bucket:  "valid-bucket",
			key:     "path/to/file.txt",
			wantErr: true, // Will fail with AWS error since not mocked
		},
		{
			name:    "invalid bucket",
			bucket:  "",
			key:     "path/to/file.txt",
			wantErr: true,
		},
		{
			name:    "invalid key",
			bucket:  "valid-bucket",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s3.Upload(ctx, tt.bucket, tt.key, content, nil)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestS3Storage_DownloadValidation(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	ctx := context.Background()

	tests := []struct {
		name    string
		bucket  string
		key     string
		wantErr bool
	}{
		{
			name:    "invalid bucket",
			bucket:  "",
			key:     "path/to/file.txt",
			wantErr: true,
		},
		{
			name:    "invalid key",
			bucket:  "valid-bucket",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := s3.Download(ctx, tt.bucket, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, reader)
			}
			if reader != nil {
				io.Copy(io.Discard, reader)
				reader.Close()
			}
		})
	}
}

func TestS3Storage_ListValidation(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	ctx := context.Background()

	tests := []struct {
		name    string
		bucket  string
		prefix  string
		options *ListOptions
		wantErr bool
	}{
		{
			name:    "invalid bucket",
			bucket:  "",
			prefix:  "prefix/",
			wantErr: true,
		},
		{
			name:   "invalid options",
			bucket: "valid-bucket",
			prefix: "prefix/",
			options: &ListOptions{
				MaxKeys: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := s3.List(ctx, tt.bucket, tt.prefix, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, files)
			}
		})
	}
}

func TestS3Storage_DeleteValidation(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	ctx := context.Background()

	tests := []struct {
		name    string
		bucket  string
		key     string
		wantErr bool
	}{
		{
			name:    "invalid bucket",
			bucket:  "",
			key:     "path/to/file.txt",
			wantErr: true,
		},
		{
			name:    "invalid key",
			bucket:  "valid-bucket",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s3.Delete(ctx, tt.bucket, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestS3Storage_GetMetadataValidation(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	ctx := context.Background()

	tests := []struct {
		name    string
		bucket  string
		key     string
		wantErr bool
	}{
		{
			name:    "invalid bucket",
			bucket:  "",
			key:     "path/to/file.txt",
			wantErr: true,
		},
		{
			name:    "invalid key",
			bucket:  "valid-bucket",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := s3.GetMetadata(ctx, tt.bucket, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, info)
			}
		})
	}
}

func TestS3Storage_GetPresignedURLValidation(t *testing.T) {
	s3, err := NewS3Storage("us-east-1", "test-key", "test-secret")
	require.NoError(t, err)
	defer s3.Close()

	ctx := context.Background()

	tests := []struct {
		name       string
		bucket     string
		key        string
		expiration time.Duration
		wantErr    bool
	}{
		{
			name:       "invalid bucket",
			bucket:     "",
			key:        "path/to/file.txt",
			expiration: 1 * time.Hour,
			wantErr:    true,
		},
		{
			name:       "invalid key",
			bucket:     "valid-bucket",
			key:        "",
			expiration: 1 * time.Hour,
			wantErr:    true,
		},
		{
			name:       "invalid expiration",
			bucket:     "valid-bucket",
			key:        "path/to/file.txt",
			expiration: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := s3.GetPresignedURL(ctx, tt.bucket, tt.key, tt.expiration)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, url)
			}
		})
	}
}
