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

func TestNewGCSStorage(t *testing.T) {
	tests := []struct {
		name            string
		projectID       string
		credentialsJSON string
		wantErr         bool
		errMsg          string
	}{
		{
			name:            "valid configuration",
			projectID:       "test-project",
			credentialsJSON: `{"type":"service_account","project_id":"test"}`,
			wantErr:         true, // Will fail with actual GCS auth, but validates params
		},
		{
			name:            "empty project ID",
			projectID:       "",
			credentialsJSON: `{"type":"service_account"}`,
			wantErr:         true,
			errMsg:          "project ID is required",
		},
		{
			name:            "empty credentials",
			projectID:       "test-project",
			credentialsJSON: "",
			wantErr:         true,
			errMsg:          "credentials JSON is required",
		},
		{
			name:            "invalid JSON",
			projectID:       "test-project",
			credentialsJSON: "not-json",
			wantErr:         true,
			errMsg:          "invalid credentials JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage, err := NewGCSStorage(ctx, tt.projectID, tt.credentialsJSON)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, storage)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, storage)
				assert.Equal(t, tt.projectID, storage.projectID)
				err = storage.Close()
				assert.NoError(t, err)
			}
		})
	}
}

func TestGCSStorage_ValidateBucket(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS validation tests - cannot create test client")
	}
	defer gcs.Close()

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
			err := gcs.validateBucket(tt.bucket)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGCSStorage_ValidateKey(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS validation tests - cannot create test client")
	}
	defer gcs.Close()

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
			err := gcs.validateKey(tt.key)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGCSStorage_ValidateUploadOptions(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS validation tests - cannot create test client")
	}
	defer gcs.Close()

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
			err := gcs.validateUploadOptions(tt.options)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGCSStorage_ValidatePresignedExpiration(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS validation tests - cannot create test client")
	}
	defer gcs.Close()

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
			err := gcs.validatePresignedExpiration(tt.expiration)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGCSStorage_ValidateListOptions(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS validation tests - cannot create test client")
	}
	defer gcs.Close()

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
			err := gcs.validateListOptions(tt.options)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Mock tests for operations
func TestGCSStorage_UploadValidation(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS upload validation tests - cannot create test client")
	}
	defer gcs.Close()

	content := bytes.NewReader([]byte("test content"))

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
			err := gcs.Upload(ctx, tt.bucket, tt.key, content, nil)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestGCSStorage_DownloadValidation(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS download validation tests - cannot create test client")
	}
	defer gcs.Close()

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
			reader, err := gcs.Download(ctx, tt.bucket, tt.key)
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

func TestGCSStorage_ListValidation(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS list validation tests - cannot create test client")
	}
	defer gcs.Close()

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
			files, err := gcs.List(ctx, tt.bucket, tt.prefix, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, files)
			}
		})
	}
}

func TestGCSStorage_DeleteValidation(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS delete validation tests - cannot create test client")
	}
	defer gcs.Close()

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
			err := gcs.Delete(ctx, tt.bucket, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestGCSStorage_GetMetadataValidation(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS metadata validation tests - cannot create test client")
	}
	defer gcs.Close()

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
			info, err := gcs.GetMetadata(ctx, tt.bucket, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, info)
			}
		})
	}
}

func TestGCSStorage_GetPresignedURLValidation(t *testing.T) {
	ctx := context.Background()
	gcs, err := newGCSStorageForTest(ctx, t)
	if err != nil {
		t.Skip("Skipping GCS presigned URL validation tests - cannot create test client")
	}
	defer gcs.Close()

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
			url, err := gcs.GetPresignedURL(ctx, tt.bucket, tt.key, tt.expiration)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, url)
			}
		})
	}
}

// Helper function to create GCS storage for testing
func newGCSStorageForTest(ctx context.Context, t *testing.T) (*GCSStorage, error) {
	// Use minimal valid JSON that won't authenticate but allows creation
	credJSON := `{"type":"service_account","project_id":"test-project","private_key_id":"test","private_key":"-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----","client_email":"test@test.iam.gserviceaccount.com"}`
	return NewGCSStorage(ctx, "test-project", credJSON)
}
