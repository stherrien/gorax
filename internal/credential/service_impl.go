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
	ListWithPagination(ctx context.Context, tenantID string, filter CredentialListFilter, limit, offset int) ([]*Credential, int, error)
	UpdateLastUsedAt(ctx context.Context, tenantID, id string) error
	LogAccess(ctx context.Context, log *AccessLog) error
	GetAccessLogs(ctx context.Context, credentialID string, limit, offset int) ([]*AccessLog, error)
	RotateCredential(ctx context.Context, tenantID, credentialID, userID string, newCred *Credential, reason string) (*Credential, error)
	GetVersions(ctx context.Context, tenantID, credentialID string) ([]*CredentialVersion, error)
	GetExpiredCredentials(ctx context.Context, tenantID string, withinDuration time.Duration) ([]*Credential, error)
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

	// Validate credential value based on type
	if err := ValidateCredentialValue(input.Type, input.Value); err != nil {
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

// Rotate creates a new version of the credential value with proper version tracking
func (s *ServiceImpl) Rotate(ctx context.Context, tenantID, credentialID, userID string, input RotateCredentialInput) (*Credential, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Get existing credential to verify it exists and get its type for validation
	existing, err := s.repo.GetByID(ctx, tenantID, credentialID)
	if err != nil {
		return nil, err
	}

	// Validate new credential value based on existing credential type
	if err := ValidateCredentialValue(existing.Type, input.Value); err != nil {
		return nil, err
	}

	// Encrypt the new credential value
	credData := &CredentialData{Value: input.Value}
	encrypted, err := s.encryption.Encrypt(ctx, tenantID, credData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}

	// Create new credential struct with encrypted data for rotation
	newCred := &Credential{
		EncryptedDEK: encrypted.EncryptedDEK,
		Ciphertext:   encrypted.Ciphertext,
		Nonce:        encrypted.Nonce,
		AuthTag:      encrypted.AuthTag,
		KMSKeyID:     encrypted.KMSKeyID,
	}

	// Use the repository's RotateCredential method which handles version tracking
	rotated, err := s.repo.RotateCredential(ctx, tenantID, credentialID, userID, newCred, "manual rotation")
	if err != nil {
		s.logger.Error("credential rotation failed",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID,
			"user_id", userID,
		)
		return nil, fmt.Errorf("failed to rotate credential: %w", err)
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

	s.logger.Info("credential rotated",
		"credential_id", credentialID,
		"tenant_id", tenantID,
		"user_id", userID,
	)

	return rotated, nil
}

// ListVersions returns all versions of a credential
func (s *ServiceImpl) ListVersions(ctx context.Context, tenantID, credentialID string) ([]*CredentialValue, error) {
	// Verify credential exists
	_, err := s.repo.GetByID(ctx, tenantID, credentialID)
	if err != nil {
		return nil, err
	}

	// Get versions from repository
	versions, err := s.repo.GetVersions(ctx, tenantID, credentialID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}

	// Convert CredentialVersion to CredentialValue for API compatibility
	var values []*CredentialValue
	for _, v := range versions {
		values = append(values, &CredentialValue{
			ID:           v.ID,
			CredentialID: v.CredentialID,
			Version:      v.Version,
			CreatedAt:    v.CreatedAt,
			CreatedBy:    v.CreatedBy,
			IsActive:     v.IsActive,
		})
	}

	if values == nil {
		values = []*CredentialValue{}
	}

	return values, nil
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
