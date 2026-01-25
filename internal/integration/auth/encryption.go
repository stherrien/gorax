package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gorax/gorax/internal/integration"
)

const (
	// KeySize is the size of the AES-256 key (32 bytes).
	KeySize = 32
	// NonceSize is the size of the GCM nonce (12 bytes).
	NonceSize = 12
)

// CredentialEncryptor handles encryption and decryption of credentials.
type CredentialEncryptor struct {
	masterKey []byte
}

// NewCredentialEncryptor creates a new credential encryptor with the given master key.
func NewCredentialEncryptor(masterKey []byte) (*CredentialEncryptor, error) {
	if len(masterKey) != KeySize {
		return nil, fmt.Errorf("master key must be %d bytes, got %d", KeySize, len(masterKey))
	}

	return &CredentialEncryptor{
		masterKey: masterKey,
	}, nil
}

// Encrypt encrypts the credential data.
func (e *CredentialEncryptor) Encrypt(data integration.JSONMap) (*integration.EncryptedData, error) {
	if data == nil {
		return nil, integration.NewValidationError("data", "data cannot be nil", nil)
	}

	// Serialize data to JSON
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshaling data: %w", err)
	}

	// Generate DEK (Data Encryption Key)
	dek := make([]byte, KeySize)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("generating DEK: %w", err)
	}

	// Encrypt data with DEK
	ciphertext, nonce, err := e.encryptWithKey(dek, plaintext)
	if err != nil {
		return nil, fmt.Errorf("encrypting data: %w", err)
	}

	// Encrypt DEK with master key (envelope encryption)
	encryptedDEK, dekNonce, err := e.encryptWithKey(e.masterKey, dek)
	if err != nil {
		return nil, fmt.Errorf("encrypting DEK: %w", err)
	}

	// Combine DEK nonce with encrypted DEK
	encryptedDEKWithNonce := make([]byte, NonceSize+len(encryptedDEK))
	copy(encryptedDEKWithNonce[:NonceSize], dekNonce)
	copy(encryptedDEKWithNonce[NonceSize:], encryptedDEK)

	return &integration.EncryptedData{
		EncryptedDEK: encryptedDEKWithNonce,
		Ciphertext:   ciphertext,
		Nonce:        nonce,
	}, nil
}

// Decrypt decrypts the credential data.
func (e *CredentialEncryptor) Decrypt(encrypted *integration.EncryptedData) (integration.JSONMap, error) {
	if encrypted == nil {
		return nil, integration.NewValidationError("encrypted", "encrypted data cannot be nil", nil)
	}

	if len(encrypted.EncryptedDEK) < NonceSize+16 { // Minimum: nonce + auth tag
		return nil, integration.NewValidationError("encrypted_dek", "encrypted DEK too short", nil)
	}

	// Extract DEK nonce and encrypted DEK
	dekNonce := encrypted.EncryptedDEK[:NonceSize]
	encryptedDEK := encrypted.EncryptedDEK[NonceSize:]

	// Decrypt DEK with master key
	dek, err := e.decryptWithKey(e.masterKey, encryptedDEK, dekNonce)
	if err != nil {
		return nil, fmt.Errorf("decrypting DEK: %w", err)
	}

	// Decrypt data with DEK
	plaintext, err := e.decryptWithKey(dek, encrypted.Ciphertext, encrypted.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decrypting data: %w", err)
	}

	// Deserialize JSON
	var data integration.JSONMap
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("unmarshaling data: %w", err)
	}

	return data, nil
}

// encryptWithKey encrypts data using AES-256-GCM.
func (e *CredentialEncryptor) encryptWithKey(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// decryptWithKey decrypts data using AES-256-GCM.
func (e *CredentialEncryptor) decryptWithKey(key, ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}

// GenerateMasterKey generates a new random master key.
func GenerateMasterKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generating master key: %w", err)
	}
	return key, nil
}

// EncryptCredentials encrypts credentials and returns an updated copy.
func (e *CredentialEncryptor) EncryptCredentials(creds *integration.Credentials) (*integration.Credentials, error) {
	if creds == nil {
		return nil, integration.NewValidationError("credentials", "credentials cannot be nil", nil)
	}

	if len(creds.Data) == 0 {
		return creds, nil // Nothing to encrypt
	}

	encrypted, err := e.Encrypt(creds.Data)
	if err != nil {
		return nil, err
	}

	// Return a copy with encrypted data
	return &integration.Credentials{
		ID:        creds.ID,
		Type:      creds.Type,
		Name:      creds.Name,
		Data:      nil, // Clear plaintext data
		Encrypted: encrypted,
		ExpiresAt: creds.ExpiresAt,
		RefreshAt: creds.RefreshAt,
		Metadata:  creds.Metadata,
	}, nil
}

// DecryptCredentials decrypts credentials and returns an updated copy.
func (e *CredentialEncryptor) DecryptCredentials(creds *integration.Credentials) (*integration.Credentials, error) {
	if creds == nil {
		return nil, integration.NewValidationError("credentials", "credentials cannot be nil", nil)
	}

	if creds.Encrypted == nil {
		return creds, nil // Nothing to decrypt
	}

	data, err := e.Decrypt(creds.Encrypted)
	if err != nil {
		return nil, err
	}

	// Return a copy with decrypted data
	return &integration.Credentials{
		ID:        creds.ID,
		Type:      creds.Type,
		Name:      creds.Name,
		Data:      data,
		Encrypted: creds.Encrypted, // Keep encrypted data for storage
		ExpiresAt: creds.ExpiresAt,
		RefreshAt: creds.RefreshAt,
		Metadata:  creds.Metadata,
	}, nil
}
