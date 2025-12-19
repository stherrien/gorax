package credential

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned when a credential is not found
	ErrNotFound = errors.New("credential not found")

	// ErrInvalidTenantID is returned when tenant ID is empty or invalid
	ErrInvalidTenantID = errors.New("tenant ID cannot be empty")

	// ErrInvalidCredentialID is returned when credential ID is empty or invalid
	ErrInvalidCredentialID = errors.New("credential ID cannot be empty")

	// ErrInvalidCredentialName is returned when credential name is empty or invalid
	ErrInvalidCredentialName = errors.New("credential name cannot be empty")

	// ErrInvalidCredentialType is returned when credential type is invalid
	ErrInvalidCredentialType = errors.New("credential type is invalid")

	// ErrEmptyCredentialData is returned when credential data is empty
	ErrEmptyCredentialData = errors.New("credential data cannot be empty")

	// ErrDuplicateCredential is returned when a credential with the same name already exists
	ErrDuplicateCredential = errors.New("credential with this name already exists")

	// ErrEncryptionFailed is returned when encryption operation fails
	ErrEncryptionFailed = errors.New("encryption failed")

	// ErrDecryptionFailed is returned when decryption operation fails
	ErrDecryptionFailed = errors.New("decryption failed")

	// ErrKMSOperationFailed is returned when KMS operation fails
	ErrKMSOperationFailed = errors.New("KMS operation failed")

	// ErrInvalidEncryptionContext is returned when encryption context is invalid
	ErrInvalidEncryptionContext = errors.New("encryption context is invalid")

	// ErrInvalidKeyID is returned when KMS key ID is empty or invalid
	ErrInvalidKeyID = errors.New("KMS key ID cannot be empty")

	// ErrDataKeyGenerationFailed is returned when DEK generation fails
	ErrDataKeyGenerationFailed = errors.New("data key generation failed")

	// ErrInvalidCiphertext is returned when ciphertext is invalid or corrupted
	ErrInvalidCiphertext = errors.New("invalid or corrupted ciphertext")

	// ErrInvalidNonce is returned when nonce is invalid
	ErrInvalidNonce = errors.New("invalid nonce")
)

// EncryptionError wraps an error with additional context
type EncryptionError struct {
	Op  string // Operation being performed
	Err error  // Underlying error
}

func (e *EncryptionError) Error() string {
	return fmt.Sprintf("encryption error during %s: %v", e.Op, e.Err)
}

func (e *EncryptionError) Unwrap() error {
	return e.Err
}

// DecryptionError wraps an error with additional context
type DecryptionError struct {
	Op  string // Operation being performed
	Err error  // Underlying error
}

func (e *DecryptionError) Error() string {
	return fmt.Sprintf("decryption error during %s: %v", e.Op, e.Err)
}

func (e *DecryptionError) Unwrap() error {
	return e.Err
}

// KMSError wraps an error with additional context
type KMSError struct {
	Op    string // Operation being performed (GenerateDataKey, Decrypt, etc.)
	KeyID string // KMS key ID
	Err   error  // Underlying error
}

func (e *KMSError) Error() string {
	return fmt.Sprintf("KMS error during %s (key: %s): %v", e.Op, e.KeyID, e.Err)
}

func (e *KMSError) Unwrap() error {
	return e.Err
}
