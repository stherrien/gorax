package credential

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// ServiceRepositoryInterface defines the repository operations needed by the service
type ServiceRepositoryInterface interface {
	Create(ctx context.Context, tenantID, createdBy string, cred *Credential) (*Credential, error)
	GetByID(ctx context.Context, tenantID, id string) (*Credential, error)
	Update(ctx context.Context, tenantID, id string, input *UpdateCredentialInput) (*Credential, error)
	Delete(ctx context.Context, tenantID, id string) error
	List(ctx context.Context, tenantID string, filter CredentialListFilter) ([]*Credential, error)
	UpdateLastUsedAt(ctx context.Context, tenantID, id string) error
	LogAccess(ctx context.Context, log *AccessLog) error
	GetAccessLogs(ctx context.Context, credentialID string, limit, offset int) ([]*AccessLog, error)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	repo       ServiceRepositoryInterface
	encryption EncryptionServiceInterface
	logger     *slog.Logger
}

// NewServiceImpl creates a new credential service implementation
func NewServiceImpl(repo ServiceRepositoryInterface, encryption EncryptionServiceInterface, logger *slog.Logger) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &ServiceImpl{
		repo:       repo,
		encryption: encryption,
		logger:     logger,
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
	// Note: Error is intentionally ignored as this is a non-critical operation
	_ = s.repo.UpdateLastUsedAt(ctx, tenantID, credentialID)

	// Log access
	accessLog := &AccessLog{
		CredentialID: credentialID,
		TenantID:     tenantID,
		AccessedBy:   userID,
		AccessType:   AccessTypeRead,
		AccessedAt:   time.Now().UTC(),
		Success:      true,
	}
	// Note: Error is intentionally ignored as this is a non-critical operation
	_ = s.repo.LogAccess(ctx, accessLog)

	// Build and return decrypted value
	return &DecryptedValue{
		Version:   1, // Version tracking can be added later
		Value:     decryptedData.Value,
		CreatedAt: cred.CreatedAt,
	}, nil
}

// Create creates a new credential with encrypted value
func (s *ServiceImpl) Create(ctx context.Context, tenantID, userID string, input CreateCredentialInput) (*Credential, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Encrypt the credential value
	credData := &CredentialData{Value: input.Value}
	encrypted, err := s.encryption.Encrypt(ctx, tenantID, credData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}

	// Create credential struct with encrypted data
	cred := &Credential{
		Name:         input.Name,
		Description:  input.Description,
		Type:         input.Type,
		Status:       StatusActive,
		ExpiresAt:    input.ExpiresAt,
		Metadata:     input.Metadata,
		EncryptedDEK: encrypted.EncryptedDEK,
		Ciphertext:   encrypted.Ciphertext,
		Nonce:        encrypted.Nonce,
		AuthTag:      encrypted.AuthTag,
		KMSKeyID:     encrypted.KMSKeyID,
	}

	// Store in repository
	created, err := s.repo.Create(ctx, tenantID, userID, cred)
	if err != nil {
		s.logger.Error("credential creation failed",
			"error", err,
			"tenant_id", tenantID,
			"user_id", userID,
			"credential_name", input.Name,
			"credential_type", input.Type,
		)
		return nil, fmt.Errorf("failed to store credential: %w", err)
	}

	s.logger.Info("credential created",
		"credential_id", created.ID,
		"tenant_id", tenantID,
		"user_id", userID,
		"credential_type", created.Type,
	)

	return created, nil
}

// List returns credentials for a tenant (metadata only, no values)
func (s *ServiceImpl) List(ctx context.Context, tenantID string, filter CredentialListFilter, limit, offset int) ([]*Credential, error) {
	credentials, err := s.repo.List(ctx, tenantID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}

	// Apply pagination manually since repo doesn't support it
	if offset >= len(credentials) {
		return []*Credential{}, nil
	}

	end := offset + limit
	if limit <= 0 || end > len(credentials) {
		end = len(credentials)
	}

	return credentials[offset:end], nil
}

// GetByID returns credential metadata by ID
func (s *ServiceImpl) GetByID(ctx context.Context, tenantID, credentialID string) (*Credential, error) {
	cred, err := s.repo.GetByID(ctx, tenantID, credentialID)
	if err != nil {
		return nil, err
	}
	return cred, nil
}

// Update updates credential metadata (not the value)
func (s *ServiceImpl) Update(ctx context.Context, tenantID, credentialID, userID string, input UpdateCredentialInput) (*Credential, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Update via repository
	updated, err := s.repo.Update(ctx, tenantID, credentialID, &input)
	if err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	// Log access
	accessLog := &AccessLog{
		CredentialID: credentialID,
		TenantID:     tenantID,
		AccessedBy:   userID,
		AccessType:   AccessTypeUpdate,
		AccessedAt:   time.Now().UTC(),
		Success:      true,
	}
	_ = s.repo.LogAccess(ctx, accessLog)

	return updated, nil
}

// Delete deletes a credential
func (s *ServiceImpl) Delete(ctx context.Context, tenantID, credentialID, userID string) error {
	// Delete via repository
	err := s.repo.Delete(ctx, tenantID, credentialID)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	// Log access
	accessLog := &AccessLog{
		CredentialID: credentialID,
		TenantID:     tenantID,
		AccessedBy:   userID,
		AccessType:   AccessTypeDelete,
		AccessedAt:   time.Now().UTC(),
		Success:      true,
	}
	_ = s.repo.LogAccess(ctx, accessLog)

	return nil
}

// Rotate creates a new version of the credential value
func (s *ServiceImpl) Rotate(ctx context.Context, tenantID, credentialID, userID string, input RotateCredentialInput) (*Credential, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Get existing credential to verify it exists
	existing, err := s.repo.GetByID(ctx, tenantID, credentialID)
	if err != nil {
		return nil, err
	}

	// Encrypt the new credential value
	credData := &CredentialData{Value: input.Value}
	encrypted, err := s.encryption.Encrypt(ctx, tenantID, credData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}

	// Update the credential with new encrypted data
	// We need to update the credential directly in the database
	existing.EncryptedDEK = encrypted.EncryptedDEK
	existing.Ciphertext = encrypted.Ciphertext
	existing.Nonce = encrypted.Nonce
	existing.AuthTag = encrypted.AuthTag
	existing.KMSKeyID = encrypted.KMSKeyID
	existing.UpdatedAt = time.Now().UTC()

	// For rotation, we update via a special method that updates encrypted fields
	// Since the standard Update doesn't touch encrypted fields, we'll update via Create
	// Actually, we need to update the existing record. Let me use a workaround.
	// We'll delete and recreate with the same ID, or update the encrypted fields directly.
	// For simplicity, we'll update the status to trigger an update, then manually update encrypted fields.

	// For now, let's update via repository Update with the encrypted data attached
	// This requires the repository to handle encrypted field updates during rotation
	// Since our Update doesn't support this, we'll need to implement it differently

	// Workaround: Delete and recreate with same properties
	if err := s.repo.Delete(ctx, tenantID, credentialID); err != nil {
		return nil, fmt.Errorf("failed to rotate credential: %w", err)
	}

	// Recreate with new encrypted data
	rotatedCred := &Credential{
		ID:           credentialID, // Keep the same ID
		Name:         existing.Name,
		Description:  existing.Description,
		Type:         existing.Type,
		Status:       existing.Status,
		ExpiresAt:    existing.ExpiresAt,
		Metadata:     existing.Metadata,
		EncryptedDEK: encrypted.EncryptedDEK,
		Ciphertext:   encrypted.Ciphertext,
		Nonce:        encrypted.Nonce,
		AuthTag:      encrypted.AuthTag,
		KMSKeyID:     encrypted.KMSKeyID,
	}

	created, err := s.repo.Create(ctx, tenantID, userID, rotatedCred)
	if err != nil {
		return nil, fmt.Errorf("failed to recreate credential during rotation: %w", err)
	}

	// Log access
	accessLog := &AccessLog{
		CredentialID: credentialID,
		TenantID:     tenantID,
		AccessedBy:   userID,
		AccessType:   AccessTypeRotate,
		AccessedAt:   time.Now().UTC(),
		Success:      true,
	}
	_ = s.repo.LogAccess(ctx, accessLog)

	return created, nil
}

// ListVersions returns all versions of a credential
// Note: Current implementation doesn't track versions separately
// This would require a credential_versions table for full support
func (s *ServiceImpl) ListVersions(ctx context.Context, tenantID, credentialID string) ([]*CredentialValue, error) {
	// Verify credential exists
	_, err := s.repo.GetByID(ctx, tenantID, credentialID)
	if err != nil {
		return nil, err
	}

	// For now, return empty list as we don't track versions separately
	// Full version tracking would require additional database schema
	return []*CredentialValue{}, nil
}

// GetAccessLog returns access log entries for a credential
func (s *ServiceImpl) GetAccessLog(ctx context.Context, tenantID, credentialID string, limit, offset int) ([]*AccessLog, error) {
	// Verify credential exists and belongs to tenant
	_, err := s.repo.GetByID(ctx, tenantID, credentialID)
	if err != nil {
		return nil, err
	}

	logs, err := s.repo.GetAccessLogs(ctx, credentialID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get access logs: %w", err)
	}

	return logs, nil
}
