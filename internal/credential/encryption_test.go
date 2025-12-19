package credential

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSimpleEncryptionService_EncryptDecrypt tests the simple encryption service
func TestSimpleEncryptionService_EncryptDecrypt(t *testing.T) {
	// Generate a random 32-byte master key for testing
	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)

	// Create encryption service
	svc, err := NewSimpleEncryptionService(masterKey)
	require.NoError(t, err)
	require.NotNil(t, svc)

	ctx := context.Background()
	tenantID := "test-tenant-123"

	tests := []struct {
		name    string
		data    *CredentialData
		wantErr bool
	}{
		{
			name: "encrypt API key credential",
			data: &CredentialData{
				Value: map[string]interface{}{
					"api_key": "test-api-key-12345",
				},
			},
			wantErr: false,
		},
		{
			name: "encrypt OAuth2 credential",
			data: &CredentialData{
				Value: map[string]interface{}{
					"client_id":     "test-client-id",
					"client_secret": "test-client-secret",
					"access_token":  "test-access-token",
					"refresh_token": "test-refresh-token",
				},
			},
			wantErr: false,
		},
		{
			name: "encrypt basic auth credential",
			data: &CredentialData{
				Value: map[string]interface{}{
					"username": "testuser",
					"password": "testpass123",
				},
			},
			wantErr: false,
		},
		{
			name: "encrypt custom credential",
			data: &CredentialData{
				Value: map[string]interface{}{
					"api_key": "custom-key",
					"secret":  "custom-secret",
					"region":  "us-east-1",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil credential data",
			data:    nil,
			wantErr: true,
		},
		{
			name: "empty value map",
			data: &CredentialData{
				Value: map[string]interface{}{},
			},
			wantErr: false, // Empty map is valid, just encrypts to "{}"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := svc.Encrypt(ctx, tenantID, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, encrypted)

			// Verify encrypted fields are populated
			assert.NotEmpty(t, encrypted.EncryptedDEK, "EncryptedDEK should not be empty")
			assert.NotEmpty(t, encrypted.Ciphertext, "Ciphertext should not be empty")
			assert.NotEmpty(t, encrypted.Nonce, "Nonce should not be empty")
			assert.NotEmpty(t, encrypted.AuthTag, "AuthTag should not be empty")
			assert.Equal(t, "simple-encryption", encrypted.KMSKeyID)

			// Verify nonce is correct size (12 bytes for AES-GCM)
			assert.Equal(t, NonceSize, len(encrypted.Nonce), "Nonce should be 12 bytes")

			// Decrypt
			decrypted, err := svc.Decrypt(ctx, encrypted)
			require.NoError(t, err)
			require.NotNil(t, decrypted)

			// Compare original and decrypted data
			expectedJSON, err := json.Marshal(tt.data)
			require.NoError(t, err)

			decryptedJSON, err := json.Marshal(decrypted)
			require.NoError(t, err)

			assert.JSONEq(t, string(expectedJSON), string(decryptedJSON))
		})
	}
}

// TestSimpleEncryptionService_DecryptErrors tests decryption error cases
func TestSimpleEncryptionService_DecryptErrors(t *testing.T) {
	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)

	svc, err := NewSimpleEncryptionService(masterKey)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name      string
		encrypted *EncryptedSecret
		wantErr   bool
	}{
		{
			name:      "nil encrypted secret",
			encrypted: nil,
			wantErr:   true,
		},
		{
			name: "empty ciphertext",
			encrypted: &EncryptedSecret{
				EncryptedDEK: make([]byte, 48),
				Ciphertext:   []byte{},
				Nonce:        make([]byte, 12),
				AuthTag:      make([]byte, 16),
				KMSKeyID:     "simple-encryption",
			},
			wantErr: true,
		},
		{
			name: "invalid nonce size",
			encrypted: &EncryptedSecret{
				EncryptedDEK: make([]byte, 48),
				Ciphertext:   make([]byte, 32),
				Nonce:        make([]byte, 8), // Wrong size
				AuthTag:      make([]byte, 16),
				KMSKeyID:     "simple-encryption",
			},
			wantErr: true,
		},
		{
			name: "corrupted ciphertext",
			encrypted: &EncryptedSecret{
				EncryptedDEK: make([]byte, 48),
				Ciphertext:   []byte("corrupted data"),
				Nonce:        make([]byte, 12),
				AuthTag:      make([]byte, 16),
				KMSKeyID:     "simple-encryption",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Decrypt(ctx, tt.encrypted)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSimpleEncryptionService_DifferentKeys tests that different keys produce different results
func TestSimpleEncryptionService_DifferentKeys(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	_, err := rand.Read(key1)
	require.NoError(t, err)
	_, err = rand.Read(key2)
	require.NoError(t, err)

	svc1, err := NewSimpleEncryptionService(key1)
	require.NoError(t, err)
	svc2, err := NewSimpleEncryptionService(key2)
	require.NoError(t, err)

	ctx := context.Background()
	tenantID := "test-tenant"
	data := &CredentialData{
		Value: map[string]interface{}{
			"secret": "my-secret-value",
		},
	}

	// Encrypt with first service
	encrypted1, err := svc1.Encrypt(ctx, tenantID, data)
	require.NoError(t, err)

	// Encrypt same data with second service
	encrypted2, err := svc2.Encrypt(ctx, tenantID, data)
	require.NoError(t, err)

	// Ciphertexts should be different
	assert.NotEqual(t, encrypted1.Ciphertext, encrypted2.Ciphertext)

	// Encrypted DEKs should be different
	assert.NotEqual(t, encrypted1.EncryptedDEK, encrypted2.EncryptedDEK)

	// But both should decrypt correctly with their respective services
	decrypted1, err := svc1.Decrypt(ctx, encrypted1)
	require.NoError(t, err)

	decrypted2, err := svc2.Decrypt(ctx, encrypted2)
	require.NoError(t, err)

	// Both should match original data
	expectedJSON, _ := json.Marshal(data)
	decrypted1JSON, _ := json.Marshal(decrypted1)
	decrypted2JSON, _ := json.Marshal(decrypted2)

	assert.JSONEq(t, string(expectedJSON), string(decrypted1JSON))
	assert.JSONEq(t, string(expectedJSON), string(decrypted2JSON))
}

// TestSimpleEncryptionService_InvalidMasterKey tests invalid master key scenarios
func TestSimpleEncryptionService_InvalidMasterKey(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr bool
	}{
		{
			name:    "valid 32-byte key",
			key:     make([]byte, 32),
			wantErr: false,
		},
		{
			name:    "nil key",
			key:     nil,
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     []byte{},
			wantErr: true,
		},
		{
			name:    "too short key (16 bytes)",
			key:     make([]byte, 16),
			wantErr: true,
		},
		{
			name:    "too long key (64 bytes)",
			key:     make([]byte, 64),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr && len(tt.key) == 32 {
				// Fill with random data for valid case
				rand.Read(tt.key)
			}

			svc, err := NewSimpleEncryptionService(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, svc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}
