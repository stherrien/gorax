package credential

import (
	"context"
	"fmt"
	"time"
)

// ServiceRepositoryInterface defines the repository operations needed by the service
type ServiceRepositoryInterface interface {
	GetByID(ctx context.Context, tenantID, id string) (*Credential, error)
	UpdateLastUsedAt(ctx context.Context, tenantID, id string) error
	LogAccess(ctx context.Context, log *AccessLog) error
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	repo       ServiceRepositoryInterface
	encryption EncryptionServiceInterface
}

// NewServiceImpl creates a new credential service implementation
func NewServiceImpl(repo ServiceRepositoryInterface, encryption EncryptionServiceInterface) Service {
	return &ServiceImpl{
		repo:       repo,
		encryption: encryption,
	}
}

// GetValue retrieves and decrypts a credential value
func (s *ServiceImpl) GetValue(ctx context.Context, tenantID, credentialID, userID string) (*DecryptedValue, error) {
	// Retrieve credential from repository
	cred, err := s.repo.GetByID(ctx, tenantID, credentialID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	// Check if credential has encrypted data
	if len(cred.Ciphertext) == 0 || len(cred.EncryptedDEK) == 0 {
		return nil, fmt.Errorf("no credential value found for credential %s", credentialID)
	}

	// Prepare encrypted secret for decryption
	encryptedSecret := &EncryptedSecret{
		EncryptedDEK: cred.EncryptedDEK,
		Ciphertext:   cred.Ciphertext,
		Nonce:        cred.Nonce,
		AuthTag:      cred.AuthTag,
		KMSKeyID:     cred.KMSKeyID,
	}

	// Decrypt the credential value
	decryptedData, err := s.encryption.Decrypt(ctx, encryptedSecret.Ciphertext, encryptedSecret.EncryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Update last used time (synchronous for now - could be made async)
	if err := s.repo.UpdateLastUsedAt(ctx, tenantID, credentialID); err != nil {
		// Log error but don't fail the request
		// In production, this should use a proper logger
	}

	// Log access
	accessLog := &AccessLog{
		CredentialID: credentialID,
		TenantID:     tenantID,
		AccessedBy:   userID,
		AccessType:   AccessTypeRead,
		AccessedAt:   time.Now().UTC(),
		Success:      true,
	}
	if err := s.repo.LogAccess(ctx, accessLog); err != nil {
		// Log error but don't fail the request
		// In production, this should use a proper logger
	}

	// Build and return decrypted value
	return &DecryptedValue{
		Version:   1, // Version tracking can be added later
		Value:     decryptedData.Value,
		CreatedAt: cred.CreatedAt,
	}, nil
}

// Create creates a new credential with encrypted value
func (s *ServiceImpl) Create(ctx context.Context, tenantID, userID string, input CreateCredentialInput) (*Credential, error) {
	return nil, fmt.Errorf("not implemented")
}

// List returns credentials for a tenant (metadata only, no values)
func (s *ServiceImpl) List(ctx context.Context, tenantID string, filter CredentialListFilter, limit, offset int) ([]*Credential, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetByID returns credential metadata by ID
func (s *ServiceImpl) GetByID(ctx context.Context, tenantID, credentialID string) (*Credential, error) {
	return nil, fmt.Errorf("not implemented")
}

// Update updates credential metadata (not the value)
func (s *ServiceImpl) Update(ctx context.Context, tenantID, credentialID, userID string, input UpdateCredentialInput) (*Credential, error) {
	return nil, fmt.Errorf("not implemented")
}

// Delete soft-deletes a credential
func (s *ServiceImpl) Delete(ctx context.Context, tenantID, credentialID, userID string) error {
	return fmt.Errorf("not implemented")
}

// Rotate creates a new version of the credential value
func (s *ServiceImpl) Rotate(ctx context.Context, tenantID, credentialID, userID string, input RotateCredentialInput) (*Credential, error) {
	return nil, fmt.Errorf("not implemented")
}

// ListVersions returns all versions of a credential
func (s *ServiceImpl) ListVersions(ctx context.Context, tenantID, credentialID string) ([]*CredentialValue, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetAccessLog returns access log entries for a credential
func (s *ServiceImpl) GetAccessLog(ctx context.Context, tenantID, credentialID string, limit, offset int) ([]*AccessLog, error) {
	return nil, fmt.Errorf("not implemented")
}
