package storage

import (
	"context"
	"io"
	"time"
)

// FileStorage defines the interface for cloud storage operations
type FileStorage interface {
	// Upload uploads a file to the storage provider
	Upload(ctx context.Context, bucket, key string, data io.Reader, options *UploadOptions) error

	// Download downloads a file from the storage provider
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)

	// List lists files in a bucket with optional prefix
	List(ctx context.Context, bucket, prefix string, options *ListOptions) ([]FileInfo, error)

	// Delete deletes a file from the storage provider
	Delete(ctx context.Context, bucket, key string) error

	// GetMetadata retrieves metadata for a file
	GetMetadata(ctx context.Context, bucket, key string) (*FileInfo, error)

	// GetPresignedURL generates a presigned URL for temporary access
	GetPresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error)

	// Close closes any open connections
	Close() error
}

// FileInfo represents metadata about a stored file
type FileInfo struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	LastModified time.Time         `json:"last_modified"`
	ContentType  string            `json:"content_type"`
	ETag         string            `json:"etag"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// UploadOptions contains options for uploading files
type UploadOptions struct {
	ContentType          string            `json:"content_type,omitempty"`
	CacheControl         string            `json:"cache_control,omitempty"`
	Metadata             map[string]string `json:"metadata,omitempty"`
	ServerSideEncryption bool              `json:"server_side_encryption,omitempty"`
}

// ListOptions contains options for listing files
type ListOptions struct {
	MaxKeys   int    `json:"max_keys,omitempty"`
	Delimiter string `json:"delimiter,omitempty"`
}

// StorageType represents the type of storage provider
type StorageType string

const (
	StorageTypeS3        StorageType = "s3"
	StorageTypeGCS       StorageType = "gcs"
	StorageTypeAzureBlob StorageType = "azure_blob"
)

// Config contains configuration for storage providers
type Config struct {
	Type StorageType `json:"type"`

	// AWS S3
	AWSRegion          string `json:"aws_region,omitempty"`
	AWSAccessKeyID     string `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey string `json:"aws_secret_access_key,omitempty"`

	// Google Cloud Storage
	GCSProjectID       string `json:"gcs_project_id,omitempty"`
	GCSCredentialsJSON string `json:"gcs_credentials_json,omitempty"`

	// Azure Blob Storage
	AzureAccountName string `json:"azure_account_name,omitempty"`
	AzureAccountKey  string `json:"azure_account_key,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Constants for validation
const (
	MaxFileSizeBytes       = 5 * 1024 * 1024 * 1024 // 5GB
	MaxPresignedExpiration = 7 * 24 * time.Hour     // 7 days
	MaxListKeys            = 1000
)
