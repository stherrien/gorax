package credential

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServiceImpl_GetValue tests the GetValue method
func TestServiceImpl_GetValue(t *testing.T) {
	tests := []struct {
		name             string
		tenantID         string
		credentialID     string
		userID           string
		repoCredential   *Credential
		repoError        error
		encryptedData    *CredentialData
		encryptionError  error
		wantErr          bool
		errorContains    string
		validateResult   func(t *testing.T, result *DecryptedValue)
	}{
		{
			name:         "successful credential retrieval",
			tenantID:     "tenant-123",
			credentialID: "cred-123",
			userID:       "user-123",
			repoCredential: &Credential{
				ID:           "cred-123",
				TenantID:     "tenant-123",
				Name:         "slack-token",
				Type:         TypeOAuth2,
				EncryptedDEK: []byte("encrypted-key"),
				Ciphertext:   []byte("encrypted-data"),
				Nonce:        []byte("nonce"),
				AuthTag:      []byte("auth-tag"),
				KMSKeyID:     "kms-key-id",
			},
			encryptedData: &CredentialData{
				Value: map[string]interface{}{
					"access_token":  "xoxb-test-token",
					"refresh_token": "xoxe-1-test-refresh",
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *DecryptedValue) {
				assert.NotNil(t, result)
				assert.Equal(t, 1, result.Version)
				assert.Contains(t, result.Value, "access_token")
				assert.Equal(t, "xoxb-test-token", result.Value["access_token"])
			},
		},
		{
			name:           "credential not found",
			tenantID:       "tenant-123",
			credentialID:   "cred-not-found",
			userID:         "user-123",
			repoError:      ErrNotFound,
			wantErr:        true,
			errorContains:  "failed to retrieve credential",
		},
		{
			name:         "decryption failure",
			tenantID:     "tenant-123",
			credentialID: "cred-123",
			userID:       "user-123",
			repoCredential: &Credential{
				ID:           "cred-123",
				TenantID:     "tenant-123",
				Name:         "slack-token",
				EncryptedDEK: []byte("corrupted-key"),
				Ciphertext:   []byte("corrupted-data"),
				Nonce:        []byte("nonce"),
				AuthTag:      []byte("auth-tag"),
				KMSKeyID:     "kms-key-id",
			},
			encryptionError: errors.New("decryption failed"),
			wantErr:         true,
			errorContains:   "decryption failed",
		},
		{
			name:         "missing credential value",
			tenantID:     "tenant-123",
			credentialID: "cred-123",
			userID:       "user-123",
			repoCredential: &Credential{
				ID:           "cred-123",
				TenantID:     "tenant-123",
				Name:         "slack-token",
				EncryptedDEK: nil, // No encrypted data
				Ciphertext:   nil,
			},
			wantErr:       true,
			errorContains: "no credential value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := &MockRepository{
				GetByIDFunc: func(ctx context.Context, tenantID, credentialID string) (*Credential, error) {
					if tt.repoError != nil {
						return nil, tt.repoError
					}
					return tt.repoCredential, nil
				},
				UpdateLastUsedAtFunc: func(ctx context.Context, tenantID, credentialID string) error {
					return nil
				},
				LogAccessFunc: func(ctx context.Context, log *AccessLog) error {
					return nil
				},
			}

			// Create mock encryption service
			mockEncryption := &MockEncryptionService{
				DecryptFunc: func(ctx context.Context, encryptedData, encryptedKey []byte) (*CredentialData, error) {
					if tt.encryptionError != nil {
						return nil, tt.encryptionError
					}
					return tt.encryptedData, nil
				},
			}

			// Create service
			service := NewServiceImpl(mockRepo, mockEncryption)

			// Execute
			ctx := context.Background()
			result, err := service.GetValue(ctx, tt.tenantID, tt.credentialID, tt.userID)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

// TestServiceImpl_GetValue_AccessLogging tests that access is logged
func TestServiceImpl_GetValue_AccessLogging(t *testing.T) {
	var loggedAccess *AccessLog

	mockRepo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, tenantID, credentialID string) (*Credential, error) {
			return &Credential{
				ID:           credentialID,
				TenantID:     tenantID,
				Name:         "test-cred",
				EncryptedDEK: []byte("key"),
				Ciphertext:   []byte("data"),
				Nonce:        []byte("nonce"),
				AuthTag:      []byte("auth-tag"),
				KMSKeyID:     "kms-key-id",
			}, nil
		},
		UpdateLastUsedAtFunc: func(ctx context.Context, tenantID, credentialID string) error {
			return nil
		},
		LogAccessFunc: func(ctx context.Context, log *AccessLog) error {
			loggedAccess = log
			return nil
		},
	}

	mockEncryption := &MockEncryptionService{
		DecryptFunc: func(ctx context.Context, encryptedData, encryptedKey []byte) (*CredentialData, error) {
			return &CredentialData{
				Value: map[string]interface{}{"token": "test"},
			}, nil
		},
	}

	service := NewServiceImpl(mockRepo, mockEncryption)

	_, err := service.GetValue(context.Background(), "tenant-123", "cred-123", "user-123")
	require.NoError(t, err)

	// Verify access was logged
	assert.NotNil(t, loggedAccess)
	assert.Equal(t, "tenant-123", loggedAccess.TenantID)
	assert.Equal(t, "cred-123", loggedAccess.CredentialID)
	assert.Equal(t, "user-123", loggedAccess.AccessedBy)
	assert.Equal(t, AccessTypeRead, loggedAccess.AccessType)
	assert.True(t, loggedAccess.Success)
}

// TestServiceImpl_GetValue_UpdatesAccessTime tests that last accessed time is updated
func TestServiceImpl_GetValue_UpdatesAccessTime(t *testing.T) {
	accessTimeUpdated := false

	mockRepo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, tenantID, credentialID string) (*Credential, error) {
			return &Credential{
				ID:           credentialID,
				TenantID:     tenantID,
				EncryptedDEK: []byte("key"),
				Ciphertext:   []byte("data"),
				Nonce:        []byte("nonce"),
				AuthTag:      []byte("auth-tag"),
				KMSKeyID:     "kms-key-id",
			}, nil
		},
		UpdateLastUsedAtFunc: func(ctx context.Context, tenantID, credentialID string) error {
			accessTimeUpdated = true
			assert.Equal(t, "tenant-123", tenantID)
			assert.Equal(t, "cred-123", credentialID)
			return nil
		},
		LogAccessFunc: func(ctx context.Context, log *AccessLog) error {
			return nil
		},
	}

	mockEncryption := &MockEncryptionService{
		DecryptFunc: func(ctx context.Context, encryptedData, encryptedKey []byte) (*CredentialData, error) {
			return &CredentialData{Value: map[string]interface{}{}}, nil
		},
	}

	service := NewServiceImpl(mockRepo, mockEncryption)

	_, err := service.GetValue(context.Background(), "tenant-123", "cred-123", "user-123")
	require.NoError(t, err)

	assert.True(t, accessTimeUpdated)
}

// TestNewServiceImpl tests service constructor
func TestNewServiceImpl(t *testing.T) {
	mockRepo := &MockRepository{}
	mockEncryption := &MockEncryptionService{}

	service := NewServiceImpl(mockRepo, mockEncryption)

	require.NotNil(t, service)
	assert.NotNil(t, service.(*ServiceImpl).repo)
	assert.NotNil(t, service.(*ServiceImpl).encryption)
}

// Mock implementations for testing

type MockRepository struct {
	GetByIDFunc          func(ctx context.Context, tenantID, credentialID string) (*Credential, error)
	UpdateLastUsedAtFunc func(ctx context.Context, tenantID, credentialID string) error
	LogAccessFunc        func(ctx context.Context, log *AccessLog) error
}

func (m *MockRepository) Create(ctx context.Context, tenantID, createdBy string, cred *Credential) (*Credential, error) {
	return nil, nil
}

func (m *MockRepository) GetByID(ctx context.Context, tenantID, credentialID string) (*Credential, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, tenantID, credentialID)
	}
	return nil, nil
}

func (m *MockRepository) GetByName(ctx context.Context, tenantID, name string) (*Credential, error) {
	return nil, nil
}

func (m *MockRepository) List(ctx context.Context, tenantID string, filter CredentialListFilter) ([]*Credential, error) {
	return nil, nil
}

func (m *MockRepository) Update(ctx context.Context, tenantID, id string, input *UpdateCredentialInput) (*Credential, error) {
	return nil, nil
}

func (m *MockRepository) UpdateLastUsedAt(ctx context.Context, tenantID, credentialID string) error {
	if m.UpdateLastUsedAtFunc != nil {
		return m.UpdateLastUsedAtFunc(ctx, tenantID, credentialID)
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, tenantID, credentialID string) error {
	return nil
}

func (m *MockRepository) LogAccess(ctx context.Context, log *AccessLog) error {
	if m.LogAccessFunc != nil {
		return m.LogAccessFunc(ctx, log)
	}
	return nil
}

func (m *MockRepository) GetAccessLogs(ctx context.Context, credentialID string, limit, offset int) ([]*AccessLog, error) {
	return nil, nil
}

type MockEncryptionService struct {
	EncryptFunc func(ctx context.Context, tenantID string, data *CredentialData) (*EncryptedSecret, error)
	DecryptFunc func(ctx context.Context, encryptedData, encryptedKey []byte) (*CredentialData, error)
}

func (m *MockEncryptionService) Encrypt(ctx context.Context, tenantID string, data *CredentialData) (*EncryptedSecret, error) {
	if m.EncryptFunc != nil {
		return m.EncryptFunc(ctx, tenantID, data)
	}
	return nil, nil
}

func (m *MockEncryptionService) Decrypt(ctx context.Context, encryptedData, encryptedKey []byte) (*CredentialData, error) {
	if m.DecryptFunc != nil {
		return m.DecryptFunc(ctx, encryptedData, encryptedKey)
	}
	return nil, nil
}
