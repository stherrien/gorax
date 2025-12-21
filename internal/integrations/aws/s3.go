package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/gorax/gorax/internal/integrations"
)

// S3Client wraps the AWS S3 client
type S3Client struct {
	client    *s3.Client
	accessKey string
	secretKey string
	region    string
}

// NewS3Client creates a new S3 client
func NewS3Client(accessKey, secretKey, region string) (*S3Client, error) {
	if err := validateS3Config(accessKey, secretKey, region); err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &S3Client{
		client:    s3.NewFromConfig(cfg),
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}, nil
}

// ListBucketsAction implements the aws:s3:list_buckets action
type ListBucketsAction struct {
	accessKey string
	secretKey string
	region    string
}

// NewListBucketsAction creates a new ListBuckets action
func NewListBucketsAction(accessKey, secretKey, region string) *ListBucketsAction {
	return &ListBucketsAction{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}
}

// Execute implements the Action interface
func (a *ListBucketsAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	client, err := NewS3Client(a.accessKey, a.secretKey, a.region)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	result, err := client.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	buckets := make([]map[string]interface{}, 0, len(result.Buckets))
	for _, bucket := range result.Buckets {
		buckets = append(buckets, map[string]interface{}{
			"name":          aws.ToString(bucket.Name),
			"creation_date": bucket.CreationDate,
		})
	}

	return map[string]interface{}{
		"buckets": buckets,
		"count":   len(buckets),
	}, nil
}

// Validate implements the Action interface
func (a *ListBucketsAction) Validate(config map[string]interface{}) error {
	// No configuration needed for listing buckets
	return nil
}

// Name implements the Action interface
func (a *ListBucketsAction) Name() string {
	return "aws:s3:list_buckets"
}

// Description implements the Action interface
func (a *ListBucketsAction) Description() string {
	return "List all S3 buckets in the AWS account"
}

// GetObjectConfig represents configuration for GetObject action
type GetObjectConfig struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

// Validate validates GetObjectConfig
func (c *GetObjectConfig) Validate() error {
	if c.Bucket == "" {
		return fmt.Errorf("bucket is required")
	}
	if c.Key == "" {
		return fmt.Errorf("key is required")
	}
	return nil
}

// GetObjectAction implements the aws:s3:get_object action
type GetObjectAction struct {
	accessKey string
	secretKey string
	region    string
}

// NewGetObjectAction creates a new GetObject action
func NewGetObjectAction(accessKey, secretKey, region string) *GetObjectAction {
	return &GetObjectAction{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}
}

// Execute implements the Action interface
func (a *GetObjectAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var objConfig GetObjectConfig
	if err := json.Unmarshal(configJSON, &objConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := objConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewS3Client(a.accessKey, a.secretKey, a.region)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	result, err := client.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(objConfig.Bucket),
		Key:    aws.String(objConfig.Key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return map[string]interface{}{
		"body":           string(body),
		"content_type":   aws.ToString(result.ContentType),
		"content_length": result.ContentLength,
		"last_modified":  result.LastModified,
		"etag":           aws.ToString(result.ETag),
	}, nil
}

// Validate implements the Action interface
func (a *GetObjectAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var objConfig GetObjectConfig
	if err := json.Unmarshal(configJSON, &objConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return objConfig.Validate()
}

// Name implements the Action interface
func (a *GetObjectAction) Name() string {
	return "aws:s3:get_object"
}

// Description implements the Action interface
func (a *GetObjectAction) Description() string {
	return "Get an object from an S3 bucket"
}

// PutObjectConfig represents configuration for PutObject action
type PutObjectConfig struct {
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	Body        string `json:"body"`
	ContentType string `json:"content_type,omitempty"`
}

// Validate validates PutObjectConfig
func (c *PutObjectConfig) Validate() error {
	if c.Bucket == "" {
		return fmt.Errorf("bucket is required")
	}
	if c.Key == "" {
		return fmt.Errorf("key is required")
	}
	if c.Body == "" {
		return fmt.Errorf("body is required")
	}
	return nil
}

// PutObjectAction implements the aws:s3:put_object action
type PutObjectAction struct {
	accessKey string
	secretKey string
	region    string
}

// NewPutObjectAction creates a new PutObject action
func NewPutObjectAction(accessKey, secretKey, region string) *PutObjectAction {
	return &PutObjectAction{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}
}

// Execute implements the Action interface
func (a *PutObjectAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var objConfig PutObjectConfig
	if err := json.Unmarshal(configJSON, &objConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := objConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewS3Client(a.accessKey, a.secretKey, a.region)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	putInput := &s3.PutObjectInput{
		Bucket: aws.String(objConfig.Bucket),
		Key:    aws.String(objConfig.Key),
		Body:   strings.NewReader(objConfig.Body),
	}

	if objConfig.ContentType != "" {
		putInput.ContentType = aws.String(objConfig.ContentType)
	}

	result, err := client.client.PutObject(ctx, putInput)
	if err != nil {
		return nil, fmt.Errorf("failed to put object: %w", err)
	}

	return map[string]interface{}{
		"etag":    aws.ToString(result.ETag),
		"success": true,
	}, nil
}

// Validate implements the Action interface
func (a *PutObjectAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var objConfig PutObjectConfig
	if err := json.Unmarshal(configJSON, &objConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return objConfig.Validate()
}

// Name implements the Action interface
func (a *PutObjectAction) Name() string {
	return "aws:s3:put_object"
}

// Description implements the Action interface
func (a *PutObjectAction) Description() string {
	return "Upload an object to an S3 bucket"
}

// DeleteObjectConfig represents configuration for DeleteObject action
type DeleteObjectConfig struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

// Validate validates DeleteObjectConfig
func (c *DeleteObjectConfig) Validate() error {
	if c.Bucket == "" {
		return fmt.Errorf("bucket is required")
	}
	if c.Key == "" {
		return fmt.Errorf("key is required")
	}
	return nil
}

// DeleteObjectAction implements the aws:s3:delete_object action
type DeleteObjectAction struct {
	accessKey string
	secretKey string
	region    string
}

// NewDeleteObjectAction creates a new DeleteObject action
func NewDeleteObjectAction(accessKey, secretKey, region string) *DeleteObjectAction {
	return &DeleteObjectAction{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}
}

// Execute implements the Action interface
func (a *DeleteObjectAction) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var objConfig DeleteObjectConfig
	if err := json.Unmarshal(configJSON, &objConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := objConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	client, err := NewS3Client(a.accessKey, a.secretKey, a.region)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	_, err = client.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(objConfig.Bucket),
		Key:    aws.String(objConfig.Key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete object: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"bucket":  objConfig.Bucket,
		"key":     objConfig.Key,
	}, nil
}

// Validate implements the Action interface
func (a *DeleteObjectAction) Validate(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var objConfig DeleteObjectConfig
	if err := json.Unmarshal(configJSON, &objConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return objConfig.Validate()
}

// Name implements the Action interface
func (a *DeleteObjectAction) Name() string {
	return "aws:s3:delete_object"
}

// Description implements the Action interface
func (a *DeleteObjectAction) Description() string {
	return "Delete an object from an S3 bucket"
}

// validateS3Config validates S3 configuration
func validateS3Config(accessKey, secretKey, region string) error {
	if accessKey == "" {
		return fmt.Errorf("access key is required")
	}
	if secretKey == "" {
		return fmt.Errorf("secret key is required")
	}
	if region == "" {
		return fmt.Errorf("region is required")
	}
	return nil
}

// Ensure all actions implement the Action interface
var (
	_ integrations.Action = (*ListBucketsAction)(nil)
	_ integrations.Action = (*GetObjectAction)(nil)
	_ integrations.Action = (*PutObjectAction)(nil)
	_ integrations.Action = (*DeleteObjectAction)(nil)
)
