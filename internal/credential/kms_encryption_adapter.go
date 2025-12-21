package credential

import (
	"context"
)

// KMSEncryptionAdapter wraps KMSEncryptionService to implement EncryptionServiceInterface
// This adapter converts between the combined encryptedData format expected by the interface
// and the separate fields expected by KMSEncryptionService
type KMSEncryptionAdapter struct {
	service *KMSEncryptionService
}

// NewKMSEncryptionAdapter creates a new adapter for KMSEncryptionService
func NewKMSEncryptionAdapter(service *KMSEncryptionService) *KMSEncryptionAdapter {
	return &KMSEncryptionAdapter{service: service}
}

// Encrypt implements EncryptionServiceInterface
// Delegates to KMSEncryptionService and returns the encrypted secret
func (a *KMSEncryptionAdapter) Encrypt(ctx context.Context, tenantID string, data *CredentialData) (*EncryptedSecret, error) {
	return a.service.Encrypt(ctx, tenantID, data)
}

// Decrypt implements EncryptionServiceInterface
// encryptedData format: nonce (12 bytes) + ciphertext + authTag (16 bytes)
// encryptedKey format: encrypted DEK from KMS
func (a *KMSEncryptionAdapter) Decrypt(ctx context.Context, encryptedData, encryptedKey []byte) (*CredentialData, error) {
	return a.service.Decrypt(ctx, encryptedData, encryptedKey)
}
