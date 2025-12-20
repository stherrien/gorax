package google

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	driveScope = "https://www.googleapis.com/auth/drive"
)

// DriveUploadAction implements the Google Drive Upload action
type DriveUploadAction struct {
	credentialService credential.Service
	baseURL           string
}

// DriveUploadConfig defines the configuration for uploading a file
type DriveUploadConfig struct {
	FileName    string `json:"file_name"`
	Content     string `json:"content"`          // Base64 encoded content
	MimeType    string `json:"mime_type"`        // e.g., "text/plain"
	ParentID    string `json:"parent_id,omitempty"` // Folder ID
	Description string `json:"description,omitempty"`
}

// DriveUploadResult represents the result of uploading a file
type DriveUploadResult struct {
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	WebViewLink string `json:"web_view_link"`
}

// Validate validates the Drive upload configuration
func (c *DriveUploadConfig) Validate() error {
	if c.FileName == "" {
		return fmt.Errorf("file_name is required")
	}
	if c.Content == "" {
		return fmt.Errorf("content is required")
	}
	if c.MimeType == "" {
		return fmt.Errorf("mime_type is required")
	}
	return nil
}

// NewDriveUploadAction creates a new Drive upload action
func NewDriveUploadAction(credentialService credential.Service) *DriveUploadAction {
	return &DriveUploadAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *DriveUploadAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	config, ok := input.Config.(DriveUploadConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected DriveUploadConfig")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	var driveService *drive.Service
	if a.baseURL != "" {
		driveService, err = drive.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		driveService, err = drive.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	// Decode base64 content
	fileContent, err := base64.StdEncoding.DecodeString(config.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content: %w", err)
	}

	// Create file metadata
	file := &drive.File{
		Name:        config.FileName,
		MimeType:    config.MimeType,
		Description: config.Description,
	}

	if config.ParentID != "" {
		file.Parents = []string{config.ParentID}
	}

	// Upload file
	uploadedFile, err := driveService.Files.Create(file).
		Media(bytes.NewReader(fileContent)).
		Fields("id, name, webViewLink").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	result := &DriveUploadResult{
		FileID:      uploadedFile.Id,
		FileName:    uploadedFile.Name,
		WebViewLink: uploadedFile.WebViewLink,
	}

	output := actions.NewActionOutput(result)
	output.WithMetadata("file_id", uploadedFile.Id)
	output.WithMetadata("web_view_link", uploadedFile.WebViewLink)

	return output, nil
}

// DriveDownloadAction implements the Google Drive Download action
type DriveDownloadAction struct {
	credentialService credential.Service
	baseURL           string
}

// DriveDownloadConfig defines the configuration for downloading a file
type DriveDownloadConfig struct {
	FileID string `json:"file_id"`
}

// DriveDownloadResult represents the result of downloading a file
type DriveDownloadResult struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	Content  string `json:"content"` // Base64 encoded content
	MimeType string `json:"mime_type"`
}

// Validate validates the Drive download configuration
func (c *DriveDownloadConfig) Validate() error {
	if c.FileID == "" {
		return fmt.Errorf("file_id is required")
	}
	return nil
}

// NewDriveDownloadAction creates a new Drive download action
func NewDriveDownloadAction(credentialService credential.Service) *DriveDownloadAction {
	return &DriveDownloadAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *DriveDownloadAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	config, ok := input.Config.(DriveDownloadConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected DriveDownloadConfig")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	var driveService *drive.Service
	if a.baseURL != "" {
		driveService, err = drive.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		driveService, err = drive.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	// Get file metadata
	file, err := driveService.Files.Get(config.FileID).Fields("id, name, mimeType").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	// Download file content
	resp, err := driveService.Files.Get(config.FileID).Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	result := &DriveDownloadResult{
		FileID:   file.Id,
		FileName: file.Name,
		Content:  base64.StdEncoding.EncodeToString(content),
		MimeType: file.MimeType,
	}

	output := actions.NewActionOutput(result)
	output.WithMetadata("file_id", file.Id)
	output.WithMetadata("file_name", file.Name)

	return output, nil
}

// DriveListAction implements the Google Drive List Files action
type DriveListAction struct {
	credentialService credential.Service
	baseURL           string
}

// DriveListConfig defines the configuration for listing files
type DriveListConfig struct {
	Query      string `json:"query,omitempty"`       // e.g., "name contains 'report'"
	PageSize   int64  `json:"page_size,omitempty"`   // Default 100
	OrderBy    string `json:"order_by,omitempty"`    // e.g., "createdTime desc"
	FolderID   string `json:"folder_id,omitempty"`   // List files in specific folder
}

// DriveFileInfo represents file metadata
type DriveFileInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MimeType    string `json:"mime_type"`
	WebViewLink string `json:"web_view_link"`
	CreatedTime string `json:"created_time"`
	ModifiedTime string `json:"modified_time"`
}

// DriveListResult represents the result of listing files
type DriveListResult struct {
	Files []DriveFileInfo `json:"files"`
	Count int             `json:"count"`
}

// NewDriveListAction creates a new Drive list action
func NewDriveListAction(credentialService credential.Service) *DriveListAction {
	return &DriveListAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *DriveListAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	config, ok := input.Config.(DriveListConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected DriveListConfig")
	}

	if config.PageSize == 0 {
		config.PageSize = 100
	}

	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	var driveService *drive.Service
	if a.baseURL != "" {
		driveService, err = drive.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		driveService, err = drive.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	// Build query
	query := config.Query
	if config.FolderID != "" {
		if query != "" {
			query = fmt.Sprintf("'%s' in parents and %s", config.FolderID, query)
		} else {
			query = fmt.Sprintf("'%s' in parents", config.FolderID)
		}
	}

	// List files
	listCall := driveService.Files.List().
		PageSize(config.PageSize).
		Fields("files(id, name, mimeType, webViewLink, createdTime, modifiedTime)")

	if query != "" {
		listCall = listCall.Q(query)
	}
	if config.OrderBy != "" {
		listCall = listCall.OrderBy(config.OrderBy)
	}

	fileList, err := listCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]DriveFileInfo, 0, len(fileList.Files))
	for _, f := range fileList.Files {
		files = append(files, DriveFileInfo{
			ID:           f.Id,
			Name:         f.Name,
			MimeType:     f.MimeType,
			WebViewLink:  f.WebViewLink,
			CreatedTime:  f.CreatedTime,
			ModifiedTime: f.ModifiedTime,
		})
	}

	result := &DriveListResult{
		Files: files,
		Count: len(files),
	}

	output := actions.NewActionOutput(result)
	output.WithMetadata("count", len(files))

	return output, nil
}
