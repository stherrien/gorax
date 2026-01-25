package storage

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Storage implements FileStorage for AWS S3
type S3Storage struct {
	client *s3.S3
	region string
}

// NewS3Storage creates a new S3 storage client
func NewS3Storage(region, accessKeyID, secretAccessKey string) (*S3Storage, error) {
	if region == "" {
		return nil, &ValidationError{Field: "region", Message: "region is required"}
	}
	if accessKeyID == "" {
		return nil, &ValidationError{Field: "access_key_id", Message: "access key is required"}
	}
	if secretAccessKey == "" {
		return nil, &ValidationError{Field: "secret_access_key", Message: "secret key is required"}
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &S3Storage{
		client: s3.New(sess),
		region: region,
	}, nil
}

// Upload uploads a file to S3
func (s *S3Storage) Upload(ctx context.Context, bucket, key string, data io.Reader, options *UploadOptions) error {
	if err := s.validateBucket(bucket); err != nil {
		return err
	}
	if err := s.validateKey(key); err != nil {
		return err
	}
	if err := s.validateUploadOptions(options); err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   aws.ReadSeekCloser(data),
	}

	if options != nil {
		if options.ContentType != "" {
			input.ContentType = aws.String(options.ContentType)
		}
		if options.CacheControl != "" {
			input.CacheControl = aws.String(options.CacheControl)
		}
		if len(options.Metadata) > 0 {
			metadata := make(map[string]*string)
			for k, v := range options.Metadata {
				metadata[k] = aws.String(v)
			}
			input.Metadata = metadata
		}
		if options.ServerSideEncryption {
			input.ServerSideEncryption = aws.String("AES256")
		}
	}

	_, err := s.client.PutObjectWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// Download downloads a file from S3
func (s *S3Storage) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	if err := s.validateBucket(bucket); err != nil {
		return nil, err
	}
	if err := s.validateKey(key); err != nil {
		return nil, err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return result.Body, nil
}

// List lists files in an S3 bucket
func (s *S3Storage) List(ctx context.Context, bucket, prefix string, options *ListOptions) ([]FileInfo, error) {
	if err := s.validateBucket(bucket); err != nil {
		return nil, err
	}
	if err := s.validateListOptions(options); err != nil {
		return nil, err
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	if options != nil {
		if options.MaxKeys > 0 {
			input.MaxKeys = aws.Int64(int64(options.MaxKeys))
		}
		if options.Delimiter != "" {
			input.Delimiter = aws.String(options.Delimiter)
		}
	}

	var files []FileInfo
	err := s.client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			files = append(files, FileInfo{
				Key:          aws.StringValue(obj.Key),
				Size:         aws.Int64Value(obj.Size),
				LastModified: aws.TimeValue(obj.LastModified),
				ETag:         strings.Trim(aws.StringValue(obj.ETag), "\""),
			})
		}
		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	return files, nil
}

// Delete deletes a file from S3
func (s *S3Storage) Delete(ctx context.Context, bucket, key string) error {
	if err := s.validateBucket(bucket); err != nil {
		return err
	}
	if err := s.validateKey(key); err != nil {
		return err
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObjectWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// GetMetadata retrieves metadata for a file in S3
func (s *S3Storage) GetMetadata(ctx context.Context, bucket, key string) (*FileInfo, error) {
	if err := s.validateBucket(bucket); err != nil {
		return nil, err
	}
	if err := s.validateKey(key); err != nil {
		return nil, err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObjectWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 metadata: %w", err)
	}

	metadata := make(map[string]string)
	for k, v := range result.Metadata {
		metadata[k] = aws.StringValue(v)
	}

	return &FileInfo{
		Key:          key,
		Size:         aws.Int64Value(result.ContentLength),
		LastModified: aws.TimeValue(result.LastModified),
		ContentType:  aws.StringValue(result.ContentType),
		ETag:         strings.Trim(aws.StringValue(result.ETag), "\""),
		Metadata:     metadata,
	}, nil
}

// GetPresignedURL generates a presigned URL for temporary access
func (s *S3Storage) GetPresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	if err := s.validateBucket(bucket); err != nil {
		return "", err
	}
	if err := s.validateKey(key); err != nil {
		return "", err
	}
	if err := s.validatePresignedExpiration(expiration); err != nil {
		return "", err
	}

	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

// Close closes the S3 client connection
func (s *S3Storage) Close() error {
	// S3 client doesn't require explicit closing
	return nil
}

// validateBucket validates an S3 bucket name
func (s *S3Storage) validateBucket(bucket string) error {
	if bucket == "" {
		return &ValidationError{Field: "bucket", Message: "bucket name cannot be empty"}
	}
	if len(bucket) < 3 || len(bucket) > 63 {
		return &ValidationError{Field: "bucket", Message: "bucket name must be between 3 and 63 characters"}
	}

	// S3 bucket naming rules: lowercase, numbers, hyphens, dots
	validBucket := regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*[a-z0-9]$`)
	if !validBucket.MatchString(bucket) {
		return &ValidationError{Field: "bucket", Message: "bucket name contains invalid characters"}
	}

	return nil
}

// validateKey validates an S3 object key
func (s *S3Storage) validateKey(key string) error {
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
func (s *S3Storage) validateUploadOptions(options *UploadOptions) error {
	if options == nil {
		return nil
	}

	if len(options.Metadata) > 100 {
		return &ValidationError{Field: "metadata", Message: "metadata cannot exceed 100 entries"}
	}

	return nil
}

// validateListOptions validates list options
func (s *S3Storage) validateListOptions(options *ListOptions) error {
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
func (s *S3Storage) validatePresignedExpiration(expiration time.Duration) error {
	if expiration <= 0 {
		return &ValidationError{Field: "expiration", Message: "expiration must be positive"}
	}
	if expiration > MaxPresignedExpiration {
		return &ValidationError{Field: "expiration", Message: "expiration cannot exceed 7 days"}
	}

	return nil
}
