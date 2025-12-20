package credential

import (
	"context"
	"fmt"
)

// SimpleEncryptionAdapter wraps SimpleEncryptionService to implement EncryptionServiceInterface
// This adapter converts between the combined encryptedData format expected by the interface
// and the separate fields expected by SimpleEncryptionService
type SimpleEncryptionAdapter struct {
	service *SimpleEncryptionService
}

// NewSimpleEncryptionAdapter creates a new adapter
func NewSimpleEncryptionAdapter(service *SimpleEncryptionService) *SimpleEncryptionAdapter {
	return &SimpleEncryptionAdapter{service: service}
}

// Decrypt implements EncryptionServiceInterface
// encryptedData format: nonce (12 bytes) + ciphertext + authTag (16 bytes)
// encryptedKey format: nonce (12 bytes) + encrypted DEK
func (a *SimpleEncryptionAdapter) Decrypt(ctx context.Context, encryptedData, encryptedKey []byte) (*CredentialData, error) {
	const authTagSize = 16 // GCM auth tag is always 16 bytes

	// Validate minimum length: nonce + at least 1 byte ciphertext + authTag
	minLength := NonceSize + 1 + authTagSize
	if len(encryptedData) < minLength {
		return nil, &DecryptionError{
			Op:  "Decrypt",
			Err: fmt.Errorf("encrypted data too short: got %d, need at least %d", len(encryptedData), minLength),
		}
	}

	// Extract nonce (first 12 bytes)
	nonce := encryptedData[:NonceSize]

	// Extract ciphertext and authTag from remaining bytes
	remaining := encryptedData[NonceSize:]
	ciphertext := remaining[:len(remaining)-authTagSize]
	authTag := remaining[len(remaining)-authTagSize:]

	// Create EncryptedSecret struct that SimpleEncryptionService expects
	encrypted := &EncryptedSecret{
		EncryptedDEK: encryptedKey,
		Ciphertext:   ciphertext,
		Nonce:        nonce,
		AuthTag:      authTag,
	}

	return a.service.Decrypt(ctx, encrypted)
}
