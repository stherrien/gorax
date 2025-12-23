package credential

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
)

const (
	// NonceSize is the size of GCM nonce in bytes (96 bits)
	NonceSize = 12
)

// EncryptionService handles envelope encryption for credentials
type EncryptionService struct {
	kmsClient KMSClientInterface
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(kmsClient KMSClientInterface) *EncryptionService {
	return &EncryptionService{
		kmsClient: kmsClient,
	}
}

// Encrypt encrypts credential data using envelope encryption
// Returns the encrypted data (ciphertext + nonce combined) and encrypted data key
func (s *EncryptionService) Encrypt(ctx context.Context, data *CredentialData) ([]byte, []byte, error) {
	if data == nil {
		return nil, nil, &EncryptionError{
			Op:  "Encrypt",
			Err: ErrEmptyCredentialData,
		}
	}

	// Serialize credential data to JSON
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to marshal credential data: %w", err),
		}
	}

	// Generate data encryption key via KMS
	// Note: In a real implementation, we'd pass the KMS key ID and encryption context
	plainKey, encryptedKey, err := s.kmsClient.GenerateDataKey(ctx, "", nil)
	if err != nil {
		return nil, nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to generate data key: %w", err),
		}
	}

	// Ensure key is cleared after use
	defer ClearKey(plainKey)

	// Encrypt data with AES-256-GCM
	encryptedData, err := s.encryptWithAESGCM(plaintext, plainKey)
	if err != nil {
		return nil, nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to encrypt data: %w", err),
		}
	}

	return encryptedData, encryptedKey, nil
}

// Decrypt decrypts credential data using envelope encryption
func (s *EncryptionService) Decrypt(ctx context.Context, encryptedData, encryptedKey []byte) (*CredentialData, error) {
	if len(encryptedData) == 0 {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: ErrInvalidCiphertext,
		}
	}

	if len(encryptedKey) == 0 {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: ErrInvalidCiphertext,
		}
	}

	// Decrypt the data encryption key using KMS
	plainKey, err := s.kmsClient.DecryptDataKey(ctx, encryptedKey, nil)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to decrypt data key: %w", err),
		}
	}

	// Ensure key is cleared after use
	defer ClearKey(plainKey)

	// Decrypt data with AES-256-GCM
	plaintext, err := s.decryptWithAESGCM(encryptedData, plainKey)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to decrypt data: %w", err),
		}
	}

	// Deserialize credential data
	var data CredentialData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to unmarshal credential data: %w", err),
		}
	}

	return &data, nil
}

// encryptWithAESGCM encrypts plaintext using AES-256-GCM
// Returns nonce prepended to ciphertext for convenience
func (s *EncryptionService) encryptWithAESGCM(plaintext, key []byte) ([]byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data (includes authentication tag)
	// #nosec G407 -- nonce is randomly generated via crypto/rand.Reader above (line 140), not hardcoded
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Prepend nonce to ciphertext for easy retrieval during decryption
	result := make([]byte, len(nonce)+len(ciphertext))
	copy(result[:NonceSize], nonce)
	copy(result[NonceSize:], ciphertext)

	return result, nil
}

// decryptWithAESGCM decrypts ciphertext using AES-256-GCM
// Expects nonce to be prepended to ciphertext
func (s *EncryptionService) decryptWithAESGCM(encryptedData, key []byte) ([]byte, error) {
	// Validate minimum length (nonce + at least 1 byte + auth tag)
	if len(encryptedData) < NonceSize+1 {
		return nil, ErrInvalidCiphertext
	}

	// Extract nonce and ciphertext
	nonce := encryptedData[:NonceSize]
	ciphertext := encryptedData[NonceSize:]

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt and verify authentication tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt or verify: %w", err)
	}

	return plaintext, nil
}

// EncryptWithContext encrypts credential data with encryption context
func (s *EncryptionService) EncryptWithContext(ctx context.Context, data *CredentialData, keyID string, encryptionContext map[string]string) ([]byte, []byte, error) {
	if data == nil {
		return nil, nil, &EncryptionError{
			Op:  "EncryptWithContext",
			Err: ErrEmptyCredentialData,
		}
	}

	if keyID == "" {
		return nil, nil, &EncryptionError{
			Op:  "EncryptWithContext",
			Err: ErrInvalidKeyID,
		}
	}

	// Serialize credential data to JSON
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, nil, &EncryptionError{
			Op:  "EncryptWithContext",
			Err: fmt.Errorf("failed to marshal credential data: %w", err),
		}
	}

	// Generate data encryption key via KMS with context
	plainKey, encryptedKey, err := s.kmsClient.GenerateDataKey(ctx, keyID, encryptionContext)
	if err != nil {
		return nil, nil, &EncryptionError{
			Op:  "EncryptWithContext",
			Err: fmt.Errorf("failed to generate data key: %w", err),
		}
	}

	// Ensure key is cleared after use
	defer ClearKey(plainKey)

	// Encrypt data with AES-256-GCM
	encryptedData, err := s.encryptWithAESGCM(plaintext, plainKey)
	if err != nil {
		return nil, nil, &EncryptionError{
			Op:  "EncryptWithContext",
			Err: fmt.Errorf("failed to encrypt data: %w", err),
		}
	}

	return encryptedData, encryptedKey, nil
}

// DecryptWithContext decrypts credential data with encryption context
func (s *EncryptionService) DecryptWithContext(ctx context.Context, encryptedData, encryptedKey []byte, encryptionContext map[string]string) (*CredentialData, error) {
	if len(encryptedData) == 0 {
		return nil, &DecryptionError{
			Op:  "DecryptWithContext",
			Err: ErrInvalidCiphertext,
		}
	}

	if len(encryptedKey) == 0 {
		return nil, &DecryptionError{
			Op:  "DecryptWithContext",
			Err: ErrInvalidCiphertext,
		}
	}

	// Decrypt the data encryption key using KMS with context
	plainKey, err := s.kmsClient.DecryptDataKey(ctx, encryptedKey, encryptionContext)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "DecryptWithContext",
			Err: fmt.Errorf("failed to decrypt data key: %w", err),
		}
	}

	// Ensure key is cleared after use
	defer ClearKey(plainKey)

	// Decrypt data with AES-256-GCM
	plaintext, err := s.decryptWithAESGCM(encryptedData, plainKey)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "DecryptWithContext",
			Err: fmt.Errorf("failed to decrypt data: %w", err),
		}
	}

	// Deserialize credential data
	var data CredentialData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, &DecryptionError{
			Op:  "DecryptWithContext",
			Err: fmt.Errorf("failed to unmarshal credential data: %w", err),
		}
	}

	return &data, nil
}

// SimpleEncryptionService provides testing/development encryption without KMS
// Uses a fixed master key to encrypt DEKs instead of AWS KMS
type SimpleEncryptionService struct {
	masterKey []byte
}

// NewSimpleEncryptionService creates a new simple encryption service for testing
// masterKey must be exactly 32 bytes (256 bits)
func NewSimpleEncryptionService(masterKey []byte) (*SimpleEncryptionService, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("master key must be exactly 32 bytes, got %d", len(masterKey))
	}

	// Make a copy to prevent external modification
	keyCopy := make([]byte, 32)
	copy(keyCopy, masterKey)

	return &SimpleEncryptionService{
		masterKey: keyCopy,
	}, nil
}

// Encrypt encrypts credential data using envelope encryption with a fixed master key
func (s *SimpleEncryptionService) Encrypt(ctx context.Context, tenantID string, data *CredentialData) (*EncryptedSecret, error) {
	if data == nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: ErrEmptyCredentialData,
		}
	}

	// Serialize credential data to JSON
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to marshal credential data: %w", err),
		}
	}

	// Generate random DEK (32 bytes for AES-256)
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to generate DEK: %w", err),
		}
	}
	defer ClearKey(dek)

	// Encrypt DEK with master key
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to create cipher for DEK encryption: %w", err),
		}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to create GCM for DEK encryption: %w", err),
		}
	}

	dekNonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, dekNonce); err != nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to generate DEK nonce: %w", err),
		}
	}

	// Encrypt DEK (includes auth tag)
	// #nosec G407 -- dekNonce is randomly generated via crypto/rand.Reader above (line 354), not hardcoded
	encryptedDEK := gcm.Seal(nil, dekNonce, dek, nil)

	// Prepend nonce to encrypted DEK
	encryptedDEKWithNonce := make([]byte, len(dekNonce)+len(encryptedDEK))
	copy(encryptedDEKWithNonce, dekNonce)
	copy(encryptedDEKWithNonce[len(dekNonce):], encryptedDEK)

	// Encrypt credential data with DEK
	dataBlock, err := aes.NewCipher(dek)
	if err != nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to create cipher for data encryption: %w", err),
		}
	}

	dataGCM, err := cipher.NewGCM(dataBlock)
	if err != nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to create GCM for data encryption: %w", err),
		}
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to generate nonce: %w", err),
		}
	}

	// Encrypt and get ciphertext with auth tag
	// #nosec G407 -- nonce is randomly generated via crypto/rand.Reader above (line 387), not hardcoded
	ciphertextWithTag := dataGCM.Seal(nil, nonce, plaintext, nil)

	// Split ciphertext and auth tag
	// GCM auth tag is always the last 16 bytes
	authTagSize := dataGCM.Overhead()
	if len(ciphertextWithTag) < authTagSize {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("ciphertext too short"),
		}
	}

	ciphertext := ciphertextWithTag[:len(ciphertextWithTag)-authTagSize]
	authTag := ciphertextWithTag[len(ciphertextWithTag)-authTagSize:]

	return &EncryptedSecret{
		EncryptedDEK: encryptedDEKWithNonce,
		Ciphertext:   ciphertext,
		Nonce:        nonce,
		AuthTag:      authTag,
		KMSKeyID:     "simple-encryption", // Identifier for non-KMS encryption
	}, nil
}

// Decrypt decrypts credential data using envelope encryption with a fixed master key
func (s *SimpleEncryptionService) Decrypt(ctx context.Context, encrypted *EncryptedSecret) (*CredentialData, error) {
	if encrypted == nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: ErrInvalidCiphertext,
		}
	}

	// Validate encrypted data
	if len(encrypted.EncryptedDEK) < NonceSize+1 {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("encrypted DEK too short"),
		}
	}

	if len(encrypted.Nonce) != NonceSize {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: ErrInvalidNonce,
		}
	}

	// Extract nonce and encrypted DEK
	dekNonce := encrypted.EncryptedDEK[:NonceSize]
	encryptedDEK := encrypted.EncryptedDEK[NonceSize:]

	// Decrypt DEK with master key
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to create cipher for DEK decryption: %w", err),
		}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to create GCM for DEK decryption: %w", err),
		}
	}

	dek, err := gcm.Open(nil, dekNonce, encryptedDEK, nil)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to decrypt DEK: %w", err),
		}
	}
	defer ClearKey(dek)

	// Validate DEK size
	if len(dek) != 32 {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("invalid DEK size: got %d, want 32", len(dek)),
		}
	}

	// Decrypt credential data with DEK
	dataBlock, err := aes.NewCipher(dek)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to create cipher for data decryption: %w", err),
		}
	}

	dataGCM, err := cipher.NewGCM(dataBlock)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to create GCM for data decryption: %w", err),
		}
	}

	// Combine ciphertext and auth tag for GCM
	ciphertextWithTag := make([]byte, len(encrypted.Ciphertext)+len(encrypted.AuthTag))
	copy(ciphertextWithTag, encrypted.Ciphertext)
	copy(ciphertextWithTag[len(encrypted.Ciphertext):], encrypted.AuthTag)

	// Decrypt and verify
	plaintext, err := dataGCM.Open(nil, encrypted.Nonce, ciphertextWithTag, nil)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to decrypt data: %w", err),
		}
	}

	// Deserialize credential data
	var data CredentialData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to unmarshal credential data: %w", err),
		}
	}

	return &data, nil
}
