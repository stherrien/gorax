package credential

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockKMSClientForEncryption mocks the AWS KMS client for encryption testing
type MockKMSClientForEncryption struct {
	GenerateDataKeyFunc func(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error)
	DecryptFunc         func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error)
}

func (m *MockKMSClientForEncryption) GenerateDataKey(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
	return m.GenerateDataKeyFunc(ctx, params, optFns...)
}

func (m *MockKMSClientForEncryption) Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
	return m.DecryptFunc(ctx, params, optFns...)
}

// TestNewKMSEncryptionService tests KMS encryption service creation
func TestNewKMSEncryptionService(t *testing.T) {
	tests := []struct {
		name       string
		keyID      string
		wantErr    bool
		errMessage string
	}{
		{
			name:    "valid key ID",
			keyID:   "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
			wantErr: false,
		},
		{
			name:    "valid alias",
			keyID:   "alias/gorax-credentials",
			wantErr: false,
		},
		{
			name:       "empty key ID",
			keyID:      "",
			wantErr:    true,
			errMessage: "KMS key ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockKMSClientForEncryption{}
			svc, err := NewKMSEncryptionService(mockClient, tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, svc)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

// TestKMSEncryptionService_Encrypt tests encryption with KMS
func TestKMSEncryptionService_Encrypt(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-123"
	keyID := "alias/test-key"

	tests := []struct {
		name           string
		data           *CredentialData
		setupMock      func(*MockKMSClientForEncryption)
		wantErr        bool
		errContains    string
		validateResult func(*testing.T, *EncryptedSecret)
	}{
		{
			name: "successful encryption",
			data: &CredentialData{
				Value: map[string]interface{}{
					"api_key": "secret-key-123",
				},
			},
			setupMock: func(m *MockKMSClientForEncryption) {
				m.GenerateDataKeyFunc = func(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
					// Validate input
					assert.Equal(t, keyID, *params.KeyId)
					assert.Equal(t, int32(32), *params.NumberOfBytes)

					// Return mock data key
					plainKey := make([]byte, 32)
					for i := range plainKey {
						plainKey[i] = byte(i)
					}
					encryptedKey := []byte("encrypted-dek-blob-from-kms")

					return &kms.GenerateDataKeyOutput{
						Plaintext:      plainKey,
						CiphertextBlob: encryptedKey,
						KeyId:          params.KeyId,
					}, nil
				}
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *EncryptedSecret) {
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.EncryptedDEK)
				assert.NotEmpty(t, result.Ciphertext)
				assert.NotEmpty(t, result.Nonce)
				assert.NotEmpty(t, result.AuthTag)
				assert.Equal(t, keyID, result.KMSKeyID)

				// Validate nonce size
				assert.Equal(t, NonceSize, len(result.Nonce))

				// Validate auth tag size (GCM uses 16 bytes)
				assert.Equal(t, 16, len(result.AuthTag))
			},
		},
		{
			name:        "nil credential data",
			data:        nil,
			setupMock:   func(m *MockKMSClientForEncryption) {},
			wantErr:     true,
			errContains: "credential data cannot be empty",
		},
		{
			name: "empty value map",
			data: &CredentialData{
				Value: map[string]interface{}{},
			},
			setupMock: func(m *MockKMSClientForEncryption) {
				m.GenerateDataKeyFunc = func(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
					plainKey := make([]byte, 32)
					for i := range plainKey {
						plainKey[i] = byte(i)
					}
					return &kms.GenerateDataKeyOutput{
						Plaintext:      plainKey,
						CiphertextBlob: []byte("encrypted-dek"),
						KeyId:          params.KeyId,
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "KMS generate data key failure",
			data: &CredentialData{
				Value: map[string]interface{}{
					"api_key": "test-key",
				},
			},
			setupMock: func(m *MockKMSClientForEncryption) {
				m.GenerateDataKeyFunc = func(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
					return nil, &types.KMSInvalidStateException{
						Message: stringPtr("Key is not available"),
					}
				}
			},
			wantErr:     true,
			errContains: "failed to generate data key",
		},
		{
			name: "invalid data key size from KMS",
			data: &CredentialData{
				Value: map[string]interface{}{
					"api_key": "test-key",
				},
			},
			setupMock: func(m *MockKMSClientForEncryption) {
				m.GenerateDataKeyFunc = func(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
					// Return wrong size key
					return &kms.GenerateDataKeyOutput{
						Plaintext:      make([]byte, 16), // Wrong size
						CiphertextBlob: []byte("encrypted-dek"),
						KeyId:          params.KeyId,
					}, nil
				}
			},
			wantErr:     true,
			errContains: "invalid data key size",
		},
		{
			name: "complex credential data",
			data: &CredentialData{
				Value: map[string]interface{}{
					"client_id":     "test-client",
					"client_secret": "test-secret",
					"access_token":  "test-token",
					"nested": map[string]interface{}{
						"field1": "value1",
						"field2": 123,
					},
				},
			},
			setupMock: func(m *MockKMSClientForEncryption) {
				m.GenerateDataKeyFunc = func(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
					plainKey := make([]byte, 32)
					for i := range plainKey {
						plainKey[i] = byte(i)
					}
					return &kms.GenerateDataKeyOutput{
						Plaintext:      plainKey,
						CiphertextBlob: []byte("encrypted-dek-complex"),
						KeyId:          params.KeyId,
					}, nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockKMSClientForEncryption{}
			tt.setupMock(mockClient)

			svc, err := NewKMSEncryptionService(mockClient, keyID)
			require.NoError(t, err)

			encrypted, err := svc.Encrypt(ctx, tenantID, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, encrypted)
				}
			}
		})
	}
}

// TestKMSEncryptionService_Decrypt tests decryption with KMS
func TestKMSEncryptionService_Decrypt(t *testing.T) {
	ctx := context.Background()
	keyID := "alias/test-key"

	tests := []struct {
		name        string
		setupMock   func(*MockKMSClientForEncryption) (*EncryptedSecret, *CredentialData)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful decryption",
			setupMock: func(m *MockKMSClientForEncryption) (*EncryptedSecret, *CredentialData) {
				// First encrypt some data
				originalData := &CredentialData{
					Value: map[string]interface{}{
						"api_key": "secret-key-123",
					},
				}

				// Use a consistent key for both encrypt and decrypt
				plainKey := make([]byte, 32)
				for i := range plainKey {
					plainKey[i] = byte(i)
				}
				encryptedDEK := []byte("encrypted-dek-blob")

				// Track whether we're in encrypt or decrypt phase
				var encryptedSecret *EncryptedSecret

				// Setup mock for encryption
				m.GenerateDataKeyFunc = func(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
					// Make a copy of the key to avoid it being cleared
					keyCopy := make([]byte, len(plainKey))
					copy(keyCopy, plainKey)
					return &kms.GenerateDataKeyOutput{
						Plaintext:      keyCopy,
						CiphertextBlob: encryptedDEK,
						KeyId:          params.KeyId,
					}, nil
				}

				// Setup mock for decryption - return same key
				m.DecryptFunc = func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
					assert.Equal(t, encryptedDEK, params.CiphertextBlob)
					// Make a copy of the key to avoid it being cleared
					keyCopy := make([]byte, len(plainKey))
					copy(keyCopy, plainKey)
					return &kms.DecryptOutput{
						Plaintext: keyCopy,
						KeyId:     stringPtr(keyID),
					}, nil
				}

				// Encrypt the data
				svc, _ := NewKMSEncryptionService(m, keyID)
				encryptedSecret, _ = svc.Encrypt(ctx, "tenant-123", originalData)

				return encryptedSecret, originalData
			},
			wantErr: false,
		},
		{
			name: "KMS decrypt failure",
			setupMock: func(m *MockKMSClientForEncryption) (*EncryptedSecret, *CredentialData) {
				m.DecryptFunc = func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
					return nil, &types.InvalidCiphertextException{
						Message: stringPtr("Invalid ciphertext"),
					}
				}

				return &EncryptedSecret{
					EncryptedDEK: []byte("invalid-dek"),
					Ciphertext:   []byte("some-ciphertext"),
					Nonce:        make([]byte, NonceSize),
					AuthTag:      make([]byte, 16),
					KMSKeyID:     keyID,
				}, nil
			},
			wantErr:     true,
			errContains: "failed to decrypt data key",
		},
		{
			name: "empty encrypted DEK",
			setupMock: func(m *MockKMSClientForEncryption) (*EncryptedSecret, *CredentialData) {
				return &EncryptedSecret{
					EncryptedDEK: []byte{},
					Ciphertext:   []byte("some-ciphertext"),
					Nonce:        make([]byte, NonceSize),
					AuthTag:      make([]byte, 16),
					KMSKeyID:     keyID,
				}, nil
			},
			wantErr:     true,
			errContains: "invalid or corrupted ciphertext",
		},
		{
			name: "empty ciphertext",
			setupMock: func(m *MockKMSClientForEncryption) (*EncryptedSecret, *CredentialData) {
				return &EncryptedSecret{
					EncryptedDEK: []byte("encrypted-dek"),
					Ciphertext:   []byte{},
					Nonce:        make([]byte, NonceSize),
					AuthTag:      make([]byte, 16),
					KMSKeyID:     keyID,
				}, nil
			},
			wantErr:     true,
			errContains: "encrypted data too short",
		},
		{
			name: "invalid nonce size",
			setupMock: func(m *MockKMSClientForEncryption) (*EncryptedSecret, *CredentialData) {
				// Add mock decrypt that won't be called (will fail before KMS call)
				m.DecryptFunc = func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
					return &kms.DecryptOutput{
						Plaintext: make([]byte, 32),
						KeyId:     stringPtr(keyID),
					}, nil
				}

				return &EncryptedSecret{
					EncryptedDEK: []byte("encrypted-dek"),
					Ciphertext:   []byte("some-ciphertext-data-here-with-enough-bytes"),
					Nonce:        make([]byte, 8), // Wrong size - will cause combined data to be short
					AuthTag:      make([]byte, 16),
					KMSKeyID:     keyID,
				}, nil
			},
			wantErr:     true,
			errContains: "failed to decrypt or verify",
		},
		{
			name: "wrong data key size from KMS",
			setupMock: func(m *MockKMSClientForEncryption) (*EncryptedSecret, *CredentialData) {
				m.DecryptFunc = func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
					return &kms.DecryptOutput{
						Plaintext: make([]byte, 16), // Wrong size
						KeyId:     stringPtr(keyID),
					}, nil
				}

				return &EncryptedSecret{
					EncryptedDEK: []byte("encrypted-dek"),
					Ciphertext:   []byte("some-ciphertext"),
					Nonce:        make([]byte, NonceSize),
					AuthTag:      make([]byte, 16),
					KMSKeyID:     keyID,
				}, nil
			},
			wantErr:     true,
			errContains: "invalid decrypted key size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockKMSClientForEncryption{}
			encrypted, originalData := tt.setupMock(mockClient)

			svc, err := NewKMSEncryptionService(mockClient, keyID)
			require.NoError(t, err)

			// Prepare combined format for Decrypt interface
			var combinedEncryptedData []byte
			if encrypted != nil {
				// Format: nonce + ciphertext + authTag
				combinedEncryptedData = make([]byte, len(encrypted.Nonce)+len(encrypted.Ciphertext)+len(encrypted.AuthTag))
				copy(combinedEncryptedData, encrypted.Nonce)
				copy(combinedEncryptedData[len(encrypted.Nonce):], encrypted.Ciphertext)
				copy(combinedEncryptedData[len(encrypted.Nonce)+len(encrypted.Ciphertext):], encrypted.AuthTag)
			}

			decrypted, err := svc.Decrypt(ctx, combinedEncryptedData, encrypted.EncryptedDEK)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, decrypted)
				if originalData != nil {
					assert.Equal(t, originalData.Value, decrypted.Value)
				}
			}
		})
	}
}

// TestKMSEncryptionService_EncryptDecrypt_RoundTrip tests full round-trip encryption/decryption
func TestKMSEncryptionService_EncryptDecrypt_RoundTrip(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-123"
	keyID := "alias/test-key"

	tests := []struct {
		name string
		data *CredentialData
	}{
		{
			name: "API key credential",
			data: &CredentialData{
				Value: map[string]interface{}{
					"api_key": "super-secret-api-key-12345",
				},
			},
		},
		{
			name: "OAuth2 credential",
			data: &CredentialData{
				Value: map[string]interface{}{
					"client_id":     "my-client-id",
					"client_secret": "my-client-secret",
					"access_token":  "my-access-token",
					"refresh_token": "my-refresh-token",
				},
			},
		},
		{
			name: "Basic auth credential",
			data: &CredentialData{
				Value: map[string]interface{}{
					"username": "admin",
					"password": "super-secure-password",
				},
			},
		},
		{
			name: "Complex nested credential",
			data: &CredentialData{
				Value: map[string]interface{}{
					"api_key": "key-123",
					"config": map[string]interface{}{
						"region":  "us-east-1",
						"timeout": float64(30), // Use float64 as JSON unmarshaling converts numbers to float64
						"nested": map[string]interface{}{
							"deep_value": "deep-secret",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock with consistent key
			// Use a master plainKey that we'll copy from
			masterPlainKey := make([]byte, 32)
			for i := range masterPlainKey {
				masterPlainKey[i] = byte(i)
			}
			encryptedDEK := []byte("encrypted-dek-blob-from-kms")

			mockClient := &MockKMSClientForEncryption{
				GenerateDataKeyFunc: func(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
					// Return a fresh copy to avoid clearKey issues
					keyCopy := make([]byte, len(masterPlainKey))
					copy(keyCopy, masterPlainKey)
					return &kms.GenerateDataKeyOutput{
						Plaintext:      keyCopy,
						CiphertextBlob: encryptedDEK,
						KeyId:          params.KeyId,
					}, nil
				},
				DecryptFunc: func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
					// Return a fresh copy to avoid clearKey issues
					keyCopy := make([]byte, len(masterPlainKey))
					copy(keyCopy, masterPlainKey)
					return &kms.DecryptOutput{
						Plaintext: keyCopy,
						KeyId:     stringPtr(keyID),
					}, nil
				},
			}

			svc, err := NewKMSEncryptionService(mockClient, keyID)
			require.NoError(t, err)

			// Encrypt
			encrypted, err := svc.Encrypt(ctx, tenantID, tt.data)
			require.NoError(t, err)
			require.NotNil(t, encrypted)

			// Verify encrypted fields are populated
			assert.NotEmpty(t, encrypted.EncryptedDEK)
			assert.NotEmpty(t, encrypted.Ciphertext)
			assert.NotEmpty(t, encrypted.Nonce)
			assert.NotEmpty(t, encrypted.AuthTag)
			assert.Equal(t, keyID, encrypted.KMSKeyID)

			// Prepare combined format for decryption
			combinedEncryptedData := make([]byte, len(encrypted.Nonce)+len(encrypted.Ciphertext)+len(encrypted.AuthTag))
			copy(combinedEncryptedData, encrypted.Nonce)
			copy(combinedEncryptedData[len(encrypted.Nonce):], encrypted.Ciphertext)
			copy(combinedEncryptedData[len(encrypted.Nonce)+len(encrypted.Ciphertext):], encrypted.AuthTag)

			// Decrypt
			decrypted, err := svc.Decrypt(ctx, combinedEncryptedData, encrypted.EncryptedDEK)
			require.NoError(t, err)
			require.NotNil(t, decrypted)

			// Verify decrypted data matches original
			assert.Equal(t, tt.data.Value, decrypted.Value)
		})
	}
}

// stringPtr is a helper to get string pointer
func stringPtr(s string) *string {
	return &s
}
