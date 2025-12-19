package credential

import (
	"context"
)

// Service defines the credential service interface
type Service interface {
	// Create creates a new credential with encrypted value
	Create(ctx context.Context, tenantID, userID string, input CreateCredentialInput) (*Credential, error)

	// List returns credentials for a tenant (metadata only, no values)
	List(ctx context.Context, tenantID string, filter CredentialListFilter, limit, offset int) ([]*Credential, error)

	// GetByID returns credential metadata by ID
	GetByID(ctx context.Context, tenantID, credentialID string) (*Credential, error)

	// GetValue returns the decrypted credential value (requires special permissions)
	GetValue(ctx context.Context, tenantID, credentialID, userID string) (*DecryptedValue, error)

	// Update updates credential metadata (not the value)
	Update(ctx context.Context, tenantID, credentialID, userID string, input UpdateCredentialInput) (*Credential, error)

	// Delete soft-deletes a credential
	Delete(ctx context.Context, tenantID, credentialID, userID string) error

	// Rotate creates a new version of the credential value
	Rotate(ctx context.Context, tenantID, credentialID, userID string, input RotateCredentialInput) (*Credential, error)

	// ListVersions returns all versions of a credential
	ListVersions(ctx context.Context, tenantID, credentialID string) ([]*CredentialValue, error)

	// GetAccessLog returns access log entries for a credential
	GetAccessLog(ctx context.Context, tenantID, credentialID string, limit, offset int) ([]*AccessLog, error)
}
