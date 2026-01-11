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

func TestNewAzureBlobStorage(t *testing.T) {
	tests := []struct {
		name        string
		accountName string
		accountKey  string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid configuration",
			accountName: "testaccount",
			accountKey:  "dGVzdC1rZXk=", // base64 encoded
			wantErr:     false,
		},
		{
			name:        "empty account name",
			accountName: "",
			accountKey:  "dGVzdC1rZXk=",
			wantErr:     true,
			errMsg:      "account name is required",
		},
		{
			name:        "empty account key",
			accountName: "testaccount",
			accountKey:  "",
			wantErr:     true,
			errMsg:      "account key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewAzureBlobStorage(tt.accountName, tt.accountKey)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, storage)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, storage)
				assert.Equal(t, tt.accountName, storage.accountName)
				err = storage.Close()
				assert.NoError(t, err)
			}
		})
	}
}

func TestAzureBlobStorage_ValidateContainer(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

	tests := []struct {
		name      string
		container string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid container",
			container: "my-container",
			wantErr:   false,
		},
		{
			name:      "valid container with numbers",
			container: "my-container-123",
			wantErr:   false,
		},
		{
			name:      "empty container",
			container: "",
			wantErr:   true,
			errMsg:    "container name cannot be empty",
		},
		{
			name:      "container too short",
			container: "ab",
			wantErr:   true,
			errMsg:    "container name must be between 3 and 63 characters",
		},
		{
			name:      "container too long",
			container: strings.Repeat("a", 64),
			wantErr:   true,
			errMsg:    "container name must be between 3 and 63 characters",
		},
		{
			name:      "invalid characters",
			container: "my_container",
			wantErr:   true,
			errMsg:    "container name contains invalid characters",
		},
		{
			name:      "uppercase letters",
			container: "MyContainer",
			wantErr:   true,
			errMsg:    "container name contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := azure.validateContainer(tt.container)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAzureBlobStorage_ValidateKey(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

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
			errMsg:  "blob name cannot be empty",
		},
		{
			name:    "key too long",
			key:     strings.Repeat("a", 1025),
			wantErr: true,
			errMsg:  "blob name must not exceed 1024 characters",
		},
		{
			name:    "path traversal attempt",
			key:     "../../../etc/passwd",
			wantErr: true,
			errMsg:  "blob name contains path traversal",
		},
		{
			name:    "starts with slash",
			key:     "/path/to/file.txt",
			wantErr: true,
			errMsg:  "blob name must not start with /",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := azure.validateBlobName(tt.key)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAzureBlobStorage_ValidateUploadOptions(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

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
			err := azure.validateUploadOptions(tt.options)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAzureBlobStorage_ValidatePresignedExpiration(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

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
			err := azure.validatePresignedExpiration(tt.expiration)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAzureBlobStorage_ValidateListOptions(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

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
			err := azure.validateListOptions(tt.options)

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
func TestAzureBlobStorage_UploadValidation(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

	ctx := context.Background()
	content := bytes.NewReader([]byte("test content"))

	tests := []struct {
		name      string
		container string
		blob      string
		wantErr   bool
	}{
		{
			name:      "invalid container",
			container: "",
			blob:      "path/to/file.txt",
			wantErr:   true,
		},
		{
			name:      "invalid blob",
			container: "valid-container",
			blob:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := azure.Upload(ctx, tt.container, tt.blob, content, nil)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestAzureBlobStorage_DownloadValidation(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

	ctx := context.Background()

	tests := []struct {
		name      string
		container string
		blob      string
		wantErr   bool
	}{
		{
			name:      "invalid container",
			container: "",
			blob:      "path/to/file.txt",
			wantErr:   true,
		},
		{
			name:      "invalid blob",
			container: "valid-container",
			blob:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := azure.Download(ctx, tt.container, tt.blob)
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

func TestAzureBlobStorage_ListValidation(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

	ctx := context.Background()

	tests := []struct {
		name      string
		container string
		prefix    string
		options   *ListOptions
		wantErr   bool
	}{
		{
			name:      "invalid container",
			container: "",
			prefix:    "prefix/",
			wantErr:   true,
		},
		{
			name:      "invalid options",
			container: "valid-container",
			prefix:    "prefix/",
			options: &ListOptions{
				MaxKeys: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := azure.List(ctx, tt.container, tt.prefix, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, files)
			}
		})
	}
}

func TestAzureBlobStorage_DeleteValidation(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

	ctx := context.Background()

	tests := []struct {
		name      string
		container string
		blob      string
		wantErr   bool
	}{
		{
			name:      "invalid container",
			container: "",
			blob:      "path/to/file.txt",
			wantErr:   true,
		},
		{
			name:      "invalid blob",
			container: "valid-container",
			blob:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := azure.Delete(ctx, tt.container, tt.blob)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestAzureBlobStorage_GetMetadataValidation(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

	ctx := context.Background()

	tests := []struct {
		name      string
		container string
		blob      string
		wantErr   bool
	}{
		{
			name:      "invalid container",
			container: "",
			blob:      "path/to/file.txt",
			wantErr:   true,
		},
		{
			name:      "invalid blob",
			container: "valid-container",
			blob:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := azure.GetMetadata(ctx, tt.container, tt.blob)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, info)
			}
		})
	}
}

func TestAzureBlobStorage_GetPresignedURLValidation(t *testing.T) {
	azure, err := NewAzureBlobStorage("testaccount", "dGVzdC1rZXk=")
	require.NoError(t, err)
	defer azure.Close()

	ctx := context.Background()

	tests := []struct {
		name       string
		container  string
		blob       string
		expiration time.Duration
		wantErr    bool
	}{
		{
			name:       "invalid container",
			container:  "",
			blob:       "path/to/file.txt",
			expiration: 1 * time.Hour,
			wantErr:    true,
		},
		{
			name:       "invalid blob",
			container:  "valid-container",
			blob:       "",
			expiration: 1 * time.Hour,
			wantErr:    true,
		},
		{
			name:       "invalid expiration",
			container:  "valid-container",
			blob:       "path/to/file.txt",
			expiration: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := azure.GetPresignedURL(ctx, tt.container, tt.blob, tt.expiration)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, url)
			}
		})
	}
}
