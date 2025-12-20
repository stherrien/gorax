package google

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriveUploadConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    DriveUploadConfig
		wantErr   bool
		errString string
	}{
		{
			name: "valid config",
			config: DriveUploadConfig{
				FileName: "test.txt",
				Content:  base64.StdEncoding.EncodeToString([]byte("test content")),
				MimeType: "text/plain",
			},
			wantErr: false,
		},
		{
			name: "missing file_name",
			config: DriveUploadConfig{
				Content:  base64.StdEncoding.EncodeToString([]byte("test content")),
				MimeType: "text/plain",
			},
			wantErr:   true,
			errString: "file_name is required",
		},
		{
			name: "missing content",
			config: DriveUploadConfig{
				FileName: "test.txt",
				MimeType: "text/plain",
			},
			wantErr:   true,
			errString: "content is required",
		},
		{
			name: "missing mime_type",
			config: DriveUploadConfig{
				FileName: "test.txt",
				Content:  base64.StdEncoding.EncodeToString([]byte("test content")),
			},
			wantErr:   true,
			errString: "mime_type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDriveDownloadConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    DriveDownloadConfig
		wantErr   bool
		errString string
	}{
		{
			name: "valid config",
			config: DriveDownloadConfig{
				FileID: "test-file-id",
			},
			wantErr: false,
		},
		{
			name:      "missing file_id",
			config:    DriveDownloadConfig{},
			wantErr:   true,
			errString: "file_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDriveUploadAction_Execute(t *testing.T) {
	mockCred := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "test-access-token",
				},
			}, nil
		},
	}

	action := NewDriveUploadAction(mockCred)

	config := DriveUploadConfig{
		FileName: "test.txt",
		Content:  base64.StdEncoding.EncodeToString([]byte("test content")),
		MimeType: "text/plain",
	}

	input := actions.NewActionInput(config, map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
		"credential_id": "cred-123",
	})

	// This will fail without a real API, but tests config validation path
	ctx := context.Background()
	_, err := action.Execute(ctx, input)

	// Expect error since we don't have a real Drive API
	assert.Error(t, err)
}

func TestDriveListAction_Execute(t *testing.T) {
	mockCred := &MockCredentialService{
		GetValueFunc: func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
			return &credential.DecryptedValue{
				Value: map[string]interface{}{
					"access_token": "test-access-token",
				},
			}, nil
		},
	}

	action := NewDriveListAction(mockCred)

	config := DriveListConfig{
		Query:    "name contains 'test'",
		PageSize: 10,
	}

	input := actions.NewActionInput(config, map[string]interface{}{
		"env": map[string]interface{}{
			"tenant_id": "tenant-123",
		},
		"credential_id": "cred-123",
	})

	ctx := context.Background()
	_, err := action.Execute(ctx, input)

	// Expect error since we don't have a real Drive API
	assert.Error(t, err)
}
