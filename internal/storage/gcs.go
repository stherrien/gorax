package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCSStorage implements FileStorage for Google Cloud Storage
type GCSStorage struct {
	client    *storage.Client
	projectID string
}

// NewGCSStorage creates a new GCS storage client
func NewGCSStorage(ctx context.Context, projectID, credentialsJSON string) (*GCSStorage, error) {
	if projectID == "" {
		return nil, &ValidationError{Field: "project_id", Message: "project ID is required"}
	}
	if credentialsJSON == "" {
		return nil, &ValidationError{Field: "credentials_json", Message: "credentials JSON is required"}
	}

	// Validate JSON format
	var creds map[string]interface{}
	if err := json.Unmarshal([]byte(credentialsJSON), &creds); err != nil {
		return nil, &ValidationError{Field: "credentials_json", Message: "invalid credentials JSON"}
	}

	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(credentialsJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStorage{
		client:    client,
		projectID: projectID,
	}, nil
}

// Upload uploads a file to GCS
func (g *GCSStorage) Upload(ctx context.Context, bucket, key string, data io.Reader, options *UploadOptions) error {
	if err := g.validateBucket(bucket); err != nil {
		return err
	}
	if err := g.validateKey(key); err != nil {
		return err
	}
	if err := g.validateUploadOptions(options); err != nil {
		return err
	}

	obj := g.client.Bucket(bucket).Object(key)
	writer := obj.NewWriter(ctx)

	if options != nil {
		if options.ContentType != "" {
			writer.ContentType = options.ContentType
		}
		if options.CacheControl != "" {
			writer.CacheControl = options.CacheControl
		}
		if len(options.Metadata) > 0 {
			writer.Metadata = options.Metadata
		}
		// GCS doesn't have a simple boolean for encryption; it uses customer-managed keys
		// For server-side encryption, GCS encrypts by default
	}

	if _, err := io.Copy(writer, data); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	return nil
}

// Download downloads a file from GCS
func (g *GCSStorage) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	if err := g.validateBucket(bucket); err != nil {
		return nil, err
	}
	if err := g.validateKey(key); err != nil {
		return nil, err
	}

	obj := g.client.Bucket(bucket).Object(key)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to download from GCS: %w", err)
	}

	return reader, nil
}

// List lists files in a GCS bucket
func (g *GCSStorage) List(ctx context.Context, bucket, prefix string, options *ListOptions) ([]FileInfo, error) {
	if err := g.validateBucket(bucket); err != nil {
		return nil, err
	}
	if err := g.validateListOptions(options); err != nil {
		return nil, err
	}

	query := &storage.Query{
		Prefix: prefix,
	}

	if options != nil {
		if options.Delimiter != "" {
			query.Delimiter = options.Delimiter
		}
	}

	var files []FileInfo
	it := g.client.Bucket(bucket).Objects(ctx, query)

	count := 0
	maxKeys := MaxListKeys
	if options != nil && options.MaxKeys > 0 {
		maxKeys = options.MaxKeys
	}

	for {
		if count >= maxKeys {
			break
		}

		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list GCS objects: %w", err)
		}

		files = append(files, FileInfo{
			Key:          attrs.Name,
			Size:         attrs.Size,
			LastModified: attrs.Updated,
			ContentType:  attrs.ContentType,
			ETag:         attrs.Etag,
			Metadata:     attrs.Metadata,
		})
		count++
	}

	return files, nil
}

// Delete deletes a file from GCS
func (g *GCSStorage) Delete(ctx context.Context, bucket, key string) error {
	if err := g.validateBucket(bucket); err != nil {
		return err
	}
	if err := g.validateKey(key); err != nil {
		return err
	}

	obj := g.client.Bucket(bucket).Object(key)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete from GCS: %w", err)
	}

	return nil
}

// GetMetadata retrieves metadata for a file in GCS
func (g *GCSStorage) GetMetadata(ctx context.Context, bucket, key string) (*FileInfo, error) {
	if err := g.validateBucket(bucket); err != nil {
		return nil, err
	}
	if err := g.validateKey(key); err != nil {
		return nil, err
	}

	obj := g.client.Bucket(bucket).Object(key)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCS metadata: %w", err)
	}

	return &FileInfo{
		Key:          attrs.Name,
		Size:         attrs.Size,
		LastModified: attrs.Updated,
		ContentType:  attrs.ContentType,
		ETag:         attrs.Etag,
		Metadata:     attrs.Metadata,
	}, nil
}

// GetPresignedURL generates a presigned URL for temporary access
func (g *GCSStorage) GetPresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	if err := g.validateBucket(bucket); err != nil {
		return "", err
	}
	if err := g.validateKey(key); err != nil {
		return "", err
	}
	if err := g.validatePresignedExpiration(expiration); err != nil {
		return "", err
	}

	// Note: SignedURL requires service account credentials with proper setup
	// For production, this would need the actual credentials
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	url, err := storage.SignedURL(bucket, key, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

// Close closes the GCS client connection
func (g *GCSStorage) Close() error {
	return g.client.Close()
}

// validateBucket validates a GCS bucket name
func (g *GCSStorage) validateBucket(bucket string) error {
	if bucket == "" {
		return &ValidationError{Field: "bucket", Message: "bucket name cannot be empty"}
	}
	if len(bucket) < 3 || len(bucket) > 63 {
		return &ValidationError{Field: "bucket", Message: "bucket name must be between 3 and 63 characters"}
	}

	// GCS bucket naming rules: lowercase, numbers, hyphens, dots
	validBucket := regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*[a-z0-9]$`)
	if !validBucket.MatchString(bucket) {
		return &ValidationError{Field: "bucket", Message: "bucket name contains invalid characters"}
	}

	return nil
}

// validateKey validates a GCS object key
func (g *GCSStorage) validateKey(key string) error {
	if key == "" {
		return &ValidationError{Field: "key", Message: "key cannot be empty"}
	}
	if len(key) > 1024 {
		return &ValidationError{Field: "key", Message: "key must not exceed 1024 characters"}
	}
	if strings.Contains(key, "..") {
		return &ValidationError{Field: "key", Message: "key contains path traversal"}
	}
	if strings.HasPrefix(key, "/") {
		return &ValidationError{Field: "key", Message: "key must not start with /"}
	}

	return nil
}

// validateUploadOptions validates upload options
func (g *GCSStorage) validateUploadOptions(options *UploadOptions) error {
	if options == nil {
		return nil
	}

	if len(options.Metadata) > 100 {
		return &ValidationError{Field: "metadata", Message: "metadata cannot exceed 100 entries"}
	}

	return nil
}

// validateListOptions validates list options
func (g *GCSStorage) validateListOptions(options *ListOptions) error {
	if options == nil {
		return nil
	}

	if options.MaxKeys < 0 {
		return &ValidationError{Field: "max_keys", Message: "max keys must be positive"}
	}
	if options.MaxKeys > MaxListKeys {
		return &ValidationError{Field: "max_keys", Message: fmt.Sprintf("max keys cannot exceed %d", MaxListKeys)}
	}

	return nil
}

// validatePresignedExpiration validates presigned URL expiration
func (g *GCSStorage) validatePresignedExpiration(expiration time.Duration) error {
	if expiration <= 0 {
		return &ValidationError{Field: "expiration", Message: "expiration must be positive"}
	}
	if expiration > MaxPresignedExpiration {
		return &ValidationError{Field: "expiration", Message: "expiration cannot exceed 7 days"}
	}

	return nil
}
