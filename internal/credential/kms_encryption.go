package credential

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

// KMSClientForEncryption defines the interface for AWS KMS operations needed by KMSEncryptionService
type KMSClientForEncryption interface {
	GenerateDataKey(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error)
	Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error)
}

// KMSEncryptionService implements production-grade encryption using AWS KMS for envelope encryption
// This service uses AWS KMS to generate and manage data encryption keys (DEKs)
// The actual credential data is encrypted with AES-256-GCM using the DEK
type KMSEncryptionService struct {
	kmsClient KMSClientForEncryption
	keyID     string
}

// NewKMSEncryptionService creates a new KMS-based encryption service
// kmsClient: AWS KMS client for key operations
// keyID: KMS key ARN or alias (e.g., "alias/gorax-credentials" or full ARN)
func NewKMSEncryptionService(kmsClient KMSClientForEncryption, keyID string) (*KMSEncryptionService, error) {
	if keyID == "" {
		return nil, ErrInvalidKeyID
	}

	if kmsClient == nil {
		return nil, fmt.Errorf("KMS client cannot be nil")
	}

	return &KMSEncryptionService{
		kmsClient: kmsClient,
		keyID:     keyID,
	}, nil
}

// Encrypt encrypts credential data using AWS KMS envelope encryption
// Steps:
// 1. Generate a data encryption key (DEK) via KMS
// 2. Encrypt the credential data with the DEK using AES-256-GCM
// 3. Return the encrypted data and the encrypted DEK
//
// Returns:
// - *EncryptedSecret containing all encrypted components
// - error if encryption fails at any step
func (s *KMSEncryptionService) Encrypt(ctx context.Context, tenantID string, data *CredentialData) (*EncryptedSecret, error) {
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

	// Generate data encryption key via KMS
	dekOutput, err := s.kmsClient.GenerateDataKey(ctx, &kms.GenerateDataKeyInput{
		KeyId:         aws.String(s.keyID),
		NumberOfBytes: aws.Int32(DataKeySize), // 32 bytes for AES-256
	})
	if err != nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to generate data key: %w", err),
		}
	}

	// Validate DEK size
	if len(dekOutput.Plaintext) != DataKeySize {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("invalid data key size: got %d, want %d", len(dekOutput.Plaintext), DataKeySize),
		}
	}

	// Ensure plaintext DEK is cleared after use
	defer ClearKey(dekOutput.Plaintext)

	// Encrypt data with AES-256-GCM using the DEK
	ciphertext, nonce, authTag, err := s.encryptWithAESGCM(plaintext, dekOutput.Plaintext)
	if err != nil {
		return nil, &EncryptionError{
			Op:  "Encrypt",
			Err: fmt.Errorf("failed to encrypt data: %w", err),
		}
	}

	return &EncryptedSecret{
		EncryptedDEK: dekOutput.CiphertextBlob,
		Ciphertext:   ciphertext,
		Nonce:        nonce,
		AuthTag:      authTag,
		KMSKeyID:     s.keyID,
	}, nil
}

// Decrypt decrypts credential data using AWS KMS envelope encryption
// Steps:
// 1. Decrypt the DEK using KMS
// 2. Decrypt the credential data with the DEK using AES-256-GCM
// 3. Deserialize and return the credential data
//
// Parameters:
// - encryptedData: combined format of nonce + ciphertext + authTag
// - encryptedKey: the encrypted DEK from KMS
//
// Returns:
// - *CredentialData containing the decrypted credential
// - error if decryption fails at any step
func (s *KMSEncryptionService) Decrypt(ctx context.Context, encryptedData, encryptedKey []byte) (*CredentialData, error) {
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

	// Parse the combined encrypted data format
	// Format: nonce (12 bytes) + ciphertext + authTag (16 bytes)
	const authTagSize = 16
	minLength := NonceSize + authTagSize + 1 // At least 1 byte of ciphertext

	if len(encryptedData) < minLength {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("encrypted data too short: got %d, need at least %d", len(encryptedData), minLength),
		}
	}

	// Extract components
	nonce := encryptedData[:NonceSize]
	remaining := encryptedData[NonceSize:]
	ciphertext := remaining[:len(remaining)-authTagSize]
	authTag := remaining[len(remaining)-authTagSize:]

	// Validate nonce size
	if len(nonce) != NonceSize {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: ErrInvalidNonce,
		}
	}

	// Decrypt the DEK using KMS
	dekOutput, err := s.kmsClient.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: encryptedKey,
	})
	if err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to decrypt data key: %w", err),
		}
	}

	// Validate decrypted DEK size
	if len(dekOutput.Plaintext) != DataKeySize {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("invalid decrypted key size: got %d, want %d", len(dekOutput.Plaintext), DataKeySize),
		}
	}

	// Ensure plaintext DEK is cleared after use
	defer ClearKey(dekOutput.Plaintext)

	// Decrypt data with AES-256-GCM
	plaintext, err := s.decryptWithAESGCM(ciphertext, nonce, authTag, dekOutput.Plaintext)
	if err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to decrypt data: %w", err),
		}
	}

	// Deserialize credential data
	var credData CredentialData
	if err := json.Unmarshal(plaintext, &credData); err != nil {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("failed to unmarshal credential data: %w", err),
		}
	}

	return &credData, nil
}

// encryptWithAESGCM encrypts plaintext using AES-256-GCM
// Returns: ciphertext (without tag), nonce, authentication tag, error
func (s *KMSEncryptionService) encryptWithAESGCM(plaintext, key []byte) ([]byte, []byte, []byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data (result includes authentication tag at the end)
	// #nosec G407 -- nonce is randomly generated above, not hardcoded
	ciphertextWithTag := gcm.Seal(nil, nonce, plaintext, nil)

	// Split ciphertext and authentication tag
	// GCM auth tag is always the last 16 bytes
	authTagSize := gcm.Overhead()
	if len(ciphertextWithTag) < authTagSize {
		return nil, nil, nil, fmt.Errorf("ciphertext too short")
	}

	ciphertext := ciphertextWithTag[:len(ciphertextWithTag)-authTagSize]
	authTag := ciphertextWithTag[len(ciphertextWithTag)-authTagSize:]

	return ciphertext, nonce, authTag, nil
}

// decryptWithAESGCM decrypts ciphertext using AES-256-GCM
func (s *KMSEncryptionService) decryptWithAESGCM(ciphertext, nonce, authTag, key []byte) ([]byte, error) {
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

	// Recombine ciphertext and auth tag for GCM decryption
	ciphertextWithTag := make([]byte, len(ciphertext)+len(authTag))
	copy(ciphertextWithTag, ciphertext)
	copy(ciphertextWithTag[len(ciphertext):], authTag)

	// Decrypt and verify authentication tag
	plaintext, err := gcm.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt or verify: %w", err)
	}

	return plaintext, nil
}
