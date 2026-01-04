package storage

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

// AzureBlobStorage implements FileStorage for Azure Blob Storage
type AzureBlobStorage struct {
	client      *azblob.Client
	accountName string
	accountKey  string
}

// NewAzureBlobStorage creates a new Azure Blob Storage client
func NewAzureBlobStorage(accountName, accountKey string) (*AzureBlobStorage, error) {
	if accountName == "" {
		return nil, &ValidationError{Field: "account_name", Message: "account name is required"}
	}
	if accountKey == "" {
		return nil, &ValidationError{Field: "account_key", Message: "account key is required"}
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)

	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credentials: %w", err)
	}

	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %w", err)
	}

	return &AzureBlobStorage{
		client:      client,
		accountName: accountName,
		accountKey:  accountKey,
	}, nil
}

// Upload uploads a file to Azure Blob Storage
func (a *AzureBlobStorage) Upload(ctx context.Context, container, blobName string, data io.Reader, options *UploadOptions) error {
	if err := a.validateContainer(container); err != nil {
		return err
	}
	if err := a.validateBlobName(blobName); err != nil {
		return err
	}
	if err := a.validateUploadOptions(options); err != nil {
		return err
	}

	uploadOptions := &azblob.UploadStreamOptions{}

	if options != nil {
		httpHeaders := &blob.HTTPHeaders{}
		if options.ContentType != "" {
			httpHeaders.BlobContentType = to.Ptr(options.ContentType)
		}
		if options.CacheControl != "" {
			httpHeaders.BlobCacheControl = to.Ptr(options.CacheControl)
		}
		if httpHeaders.BlobContentType != nil || httpHeaders.BlobCacheControl != nil {
			uploadOptions.HTTPHeaders = httpHeaders
		}

		if len(options.Metadata) > 0 {
			metadata := make(map[string]*string)
			for k, v := range options.Metadata {
				val := v
				metadata[k] = &val
			}
			uploadOptions.Metadata = metadata
		}
	}

	_, err := a.client.UploadStream(ctx, container, blobName, data, uploadOptions)
	if err != nil {
		return fmt.Errorf("failed to upload to Azure Blob: %w", err)
	}

	return nil
}

// Download downloads a file from Azure Blob Storage
func (a *AzureBlobStorage) Download(ctx context.Context, container, blobName string) (io.ReadCloser, error) {
	if err := a.validateContainer(container); err != nil {
		return nil, err
	}
	if err := a.validateBlobName(blobName); err != nil {
		return nil, err
	}

	response, err := a.client.DownloadStream(ctx, container, blobName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download from Azure Blob: %w", err)
	}

	return response.Body, nil
}

// List lists files in an Azure Blob container
func (a *AzureBlobStorage) List(ctx context.Context, container, prefix string, options *ListOptions) ([]FileInfo, error) {
	if err := a.validateContainer(container); err != nil {
		return nil, err
	}
	if err := a.validateListOptions(options); err != nil {
		return nil, err
	}

	listOptions := &azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	}

	if options != nil && options.MaxKeys > 0 {
		maxResults := int32(options.MaxKeys)
		listOptions.MaxResults = &maxResults
	}

	var files []FileInfo
	pager := a.client.NewListBlobsFlatPager(container, listOptions)

	maxKeys := MaxListKeys
	if options != nil && options.MaxKeys > 0 {
		maxKeys = options.MaxKeys
	}

	count := 0
	for pager.More() {
		if count >= maxKeys {
			break
		}

		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Azure Blobs: %w", err)
		}

		for _, blob := range page.Segment.BlobItems {
			if count >= maxKeys {
				break
			}

			metadata := make(map[string]string)
			if blob.Metadata != nil {
				for k, v := range blob.Metadata {
					if v != nil {
						metadata[k] = *v
					}
				}
			}

			files = append(files, FileInfo{
				Key:          *blob.Name,
				Size:         *blob.Properties.ContentLength,
				LastModified: *blob.Properties.LastModified,
				ContentType:  getStringValue(blob.Properties.ContentType),
				ETag:         getETagValue(blob.Properties.ETag),
				Metadata:     metadata,
			})
			count++
		}
	}

	return files, nil
}

// Delete deletes a file from Azure Blob Storage
func (a *AzureBlobStorage) Delete(ctx context.Context, container, blobName string) error {
	if err := a.validateContainer(container); err != nil {
		return err
	}
	if err := a.validateBlobName(blobName); err != nil {
		return err
	}

	_, err := a.client.DeleteBlob(ctx, container, blobName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete from Azure Blob: %w", err)
	}

	return nil
}

// GetMetadata retrieves metadata for a file in Azure Blob Storage
func (a *AzureBlobStorage) GetMetadata(ctx context.Context, container, blobName string) (*FileInfo, error) {
	if err := a.validateContainer(container); err != nil {
		return nil, err
	}
	if err := a.validateBlobName(blobName); err != nil {
		return nil, err
	}

	response, err := a.client.ServiceClient().NewContainerClient(container).NewBlobClient(blobName).GetProperties(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure Blob metadata: %w", err)
	}

	metadata := make(map[string]string)
	if response.Metadata != nil {
		for k, v := range response.Metadata {
			if v != nil {
				metadata[k] = *v
			}
		}
	}

	return &FileInfo{
		Key:          blobName,
		Size:         *response.ContentLength,
		LastModified: *response.LastModified,
		ContentType:  getStringValue(response.ContentType),
		ETag:         getETagValue(response.ETag),
		Metadata:     metadata,
	}, nil
}

// GetPresignedURL generates a presigned URL for temporary access
func (a *AzureBlobStorage) GetPresignedURL(ctx context.Context, container, blobName string, expiration time.Duration) (string, error) {
	if err := a.validateContainer(container); err != nil {
		return "", err
	}
	if err := a.validateBlobName(blobName); err != nil {
		return "", err
	}
	if err := a.validatePresignedExpiration(expiration); err != nil {
		return "", err
	}

	// Create SAS query parameters
	startTime := time.Now()
	expiryTime := startTime.Add(expiration)

	sasQueryParams, err := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     startTime,
		ExpiryTime:    expiryTime,
		Permissions:   to.Ptr(sas.BlobPermissions{Read: true}).String(),
		ContainerName: container,
		BlobName:      blobName,
	}.SignWithSharedKey(a.getSharedKeyCredential())

	if err != nil {
		return "", fmt.Errorf("failed to generate SAS token: %w", err)
	}

	// Construct URL
	sasURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s",
		a.accountName, container, blobName, sasQueryParams.Encode())

	return sasURL, nil
}

// Close closes the Azure Blob Storage client
func (a *AzureBlobStorage) Close() error {
	// Azure SDK client doesn't require explicit closing
	return nil
}

// validateContainer validates an Azure container name
func (a *AzureBlobStorage) validateContainer(container string) error {
	if container == "" {
		return &ValidationError{Field: "container", Message: "container name cannot be empty"}
	}
	if len(container) < 3 || len(container) > 63 {
		return &ValidationError{Field: "container", Message: "container name must be between 3 and 63 characters"}
	}

	// Azure container naming rules: lowercase, numbers, hyphens
	validContainer := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if !validContainer.MatchString(container) {
		return &ValidationError{Field: "container", Message: "container name contains invalid characters"}
	}

	return nil
}

// validateBlobName validates an Azure blob name
func (a *AzureBlobStorage) validateBlobName(blobName string) error {
	if blobName == "" {
		return &ValidationError{Field: "blob_name", Message: "blob name cannot be empty"}
	}
	if len(blobName) > 1024 {
		return &ValidationError{Field: "blob_name", Message: "blob name must not exceed 1024 characters"}
	}
	if strings.Contains(blobName, "..") {
		return &ValidationError{Field: "blob_name", Message: "blob name contains path traversal"}
	}
	if strings.HasPrefix(blobName, "/") {
		return &ValidationError{Field: "blob_name", Message: "blob name must not start with /"}
	}

	return nil
}

// validateUploadOptions validates upload options
func (a *AzureBlobStorage) validateUploadOptions(options *UploadOptions) error {
	if options == nil {
		return nil
	}

	if len(options.Metadata) > 100 {
		return &ValidationError{Field: "metadata", Message: "metadata cannot exceed 100 entries"}
	}

	return nil
}

// validateListOptions validates list options
func (a *AzureBlobStorage) validateListOptions(options *ListOptions) error {
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
func (a *AzureBlobStorage) validatePresignedExpiration(expiration time.Duration) error {
	if expiration <= 0 {
		return &ValidationError{Field: "expiration", Message: "expiration must be positive"}
	}
	if expiration > MaxPresignedExpiration {
		return &ValidationError{Field: "expiration", Message: "expiration cannot exceed 7 days"}
	}

	return nil
}

// getSharedKeyCredential returns a shared key credential for SAS generation
func (a *AzureBlobStorage) getSharedKeyCredential() *azblob.SharedKeyCredential {
	cred, _ := azblob.NewSharedKeyCredential(a.accountName, a.accountKey)
	return cred
}

// getStringValue safely extracts string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// getETagValue safely extracts ETag value
func getETagValue(etag interface{}) string {
	if etag == nil {
		return ""
	}

	// Handle different ETag types
	switch v := etag.(type) {
	case *string:
		if v == nil {
			return ""
		}
		return strings.Trim(*v, "\"")
	case string:
		return strings.Trim(v, "\"")
	default:
		// For azcore.ETag type, convert to string
		etagStr := fmt.Sprintf("%v", etag)
		return strings.Trim(etagStr, "\"")
	}
}

// Ensure unused imports are handled
var (
	_ = streaming.NopCloser
	_ = service.Client{}
	_ = container.Client{}
)
