package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorax/gorax/internal/integration"
)

func TestCredentialEncryptor(t *testing.T) {
	t.Run("encrypt and decrypt", func(t *testing.T) {
		masterKey, err := GenerateMasterKey()
		require.NoError(t, err)
		require.Len(t, masterKey, KeySize)

		encryptor, err := NewCredentialEncryptor(masterKey)
		require.NoError(t, err)

		data := integration.JSONMap{
			"api_key": "secret-key-12345",
			"nested": map[string]any{
				"value": "nested-secret",
			},
		}

		// Encrypt
		encrypted, err := encryptor.Encrypt(data)
		require.NoError(t, err)
		assert.NotNil(t, encrypted)
		assert.NotEmpty(t, encrypted.Ciphertext)
		assert.NotEmpty(t, encrypted.Nonce)
		assert.NotEmpty(t, encrypted.EncryptedDEK)

		// Decrypt
		decrypted, err := encryptor.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, "secret-key-12345", decrypted["api_key"])
	})

	t.Run("encrypt and decrypt credentials", func(t *testing.T) {
		masterKey, err := GenerateMasterKey()
		require.NoError(t, err)

		encryptor, err := NewCredentialEncryptor(masterKey)
		require.NoError(t, err)

		creds := &integration.Credentials{
			ID:   "cred-123",
			Type: integration.CredTypeAPIKey,
			Name: "Test Credential",
			Data: integration.JSONMap{
				"key": "secret-value",
			},
		}

		// Encrypt
		encryptedCreds, err := encryptor.EncryptCredentials(creds)
		require.NoError(t, err)
		assert.Nil(t, encryptedCreds.Data) // Plaintext should be cleared
		assert.NotNil(t, encryptedCreds.Encrypted)

		// Decrypt
		decryptedCreds, err := encryptor.DecryptCredentials(encryptedCreds)
		require.NoError(t, err)
		assert.Equal(t, "secret-value", decryptedCreds.Data["key"])
	})

	t.Run("invalid master key size", func(t *testing.T) {
		_, err := NewCredentialEncryptor([]byte("short-key"))
		assert.Error(t, err)
	})

	t.Run("encrypt nil data", func(t *testing.T) {
		masterKey, _ := GenerateMasterKey()
		encryptor, _ := NewCredentialEncryptor(masterKey)

		_, err := encryptor.Encrypt(nil)
		assert.Error(t, err)
	})

	t.Run("decrypt nil data", func(t *testing.T) {
		masterKey, _ := GenerateMasterKey()
		encryptor, _ := NewCredentialEncryptor(masterKey)

		_, err := encryptor.Decrypt(nil)
		assert.Error(t, err)
	})

	t.Run("decrypt with invalid DEK", func(t *testing.T) {
		masterKey, _ := GenerateMasterKey()
		encryptor, _ := NewCredentialEncryptor(masterKey)

		encrypted := &integration.EncryptedData{
			EncryptedDEK: []byte("too-short"),
			Ciphertext:   []byte("ciphertext"),
			Nonce:        []byte("nonce"),
		}

		_, err := encryptor.Decrypt(encrypted)
		assert.Error(t, err)
	})

	t.Run("encrypt credentials with empty data", func(t *testing.T) {
		masterKey, _ := GenerateMasterKey()
		encryptor, _ := NewCredentialEncryptor(masterKey)

		creds := &integration.Credentials{
			ID:   "cred-123",
			Type: integration.CredTypeAPIKey,
			Data: nil,
		}

		result, err := encryptor.EncryptCredentials(creds)
		require.NoError(t, err)
		assert.Same(t, creds, result) // Should return same credentials unchanged
	})

	t.Run("decrypt credentials without encrypted data", func(t *testing.T) {
		masterKey, _ := GenerateMasterKey()
		encryptor, _ := NewCredentialEncryptor(masterKey)

		creds := &integration.Credentials{
			ID:   "cred-123",
			Type: integration.CredTypeAPIKey,
		}

		result, err := encryptor.DecryptCredentials(creds)
		require.NoError(t, err)
		assert.Same(t, creds, result) // Should return same credentials unchanged
	})

	t.Run("encrypt nil credentials", func(t *testing.T) {
		masterKey, _ := GenerateMasterKey()
		encryptor, _ := NewCredentialEncryptor(masterKey)

		_, err := encryptor.EncryptCredentials(nil)
		assert.Error(t, err)
	})

	t.Run("decrypt nil credentials", func(t *testing.T) {
		masterKey, _ := GenerateMasterKey()
		encryptor, _ := NewCredentialEncryptor(masterKey)

		_, err := encryptor.DecryptCredentials(nil)
		assert.Error(t, err)
	})
}

func TestGenerateMasterKey(t *testing.T) {
	t.Run("generates correct size", func(t *testing.T) {
		key, err := GenerateMasterKey()
		require.NoError(t, err)
		assert.Len(t, key, KeySize)
	})

	t.Run("generates unique keys", func(t *testing.T) {
		key1, _ := GenerateMasterKey()
		key2, _ := GenerateMasterKey()
		assert.NotEqual(t, key1, key2)
	})
}

func TestEncryptionRoundTrip(t *testing.T) {
	// Test that data survives multiple encrypt/decrypt cycles
	masterKey, _ := GenerateMasterKey()
	encryptor, _ := NewCredentialEncryptor(masterKey)

	originalData := integration.JSONMap{
		"string":  "value",
		"number":  float64(42),
		"boolean": true,
		"array":   []any{"a", "b", "c"},
		"nested": map[string]any{
			"deep": "value",
		},
	}

	// Multiple round trips
	data := originalData
	for range 5 {
		encrypted, err := encryptor.Encrypt(data)
		require.NoError(t, err)

		decrypted, err := encryptor.Decrypt(encrypted)
		require.NoError(t, err)

		data = decrypted
	}

	assert.Equal(t, "value", data["string"])
	assert.Equal(t, float64(42), data["number"])
	assert.Equal(t, true, data["boolean"])
}

func TestEncryptionWithDifferentKeys(t *testing.T) {
	// Ensure data encrypted with one key cannot be decrypted with another
	key1, _ := GenerateMasterKey()
	key2, _ := GenerateMasterKey()

	encryptor1, _ := NewCredentialEncryptor(key1)
	encryptor2, _ := NewCredentialEncryptor(key2)

	data := integration.JSONMap{"secret": "value"}

	encrypted, err := encryptor1.Encrypt(data)
	require.NoError(t, err)

	// Trying to decrypt with different key should fail
	_, err = encryptor2.Decrypt(encrypted)
	assert.Error(t, err)
}
