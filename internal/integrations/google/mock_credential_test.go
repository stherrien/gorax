package google

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/credential"
)

// MockCredentialService for testing
type MockCredentialService struct {
	GetValueFunc func(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error)
}

func (m *MockCredentialService) Create(ctx context.Context, tenantID, userID string, input credential.CreateCredentialInput) (*credential.Credential, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockCredentialService) List(ctx context.Context, tenantID string, filter credential.CredentialListFilter, limit, offset int) ([]*credential.Credential, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockCredentialService) GetByID(ctx context.Context, tenantID, credentialID string) (*credential.Credential, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockCredentialService) GetValue(ctx context.Context, tenantID, credentialID, userID string) (*credential.DecryptedValue, error) {
	if m.GetValueFunc != nil {
		return m.GetValueFunc(ctx, tenantID, credentialID, userID)
	}
	return nil, nil
}

func (m *MockCredentialService) Update(ctx context.Context, tenantID, credentialID, userID string, input credential.UpdateCredentialInput) (*credential.Credential, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockCredentialService) Delete(ctx context.Context, tenantID, credentialID, userID string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockCredentialService) Rotate(ctx context.Context, tenantID, credentialID, userID string, input credential.RotateCredentialInput) (*credential.Credential, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockCredentialService) ListVersions(ctx context.Context, tenantID, credentialID string) ([]*credential.CredentialValue, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockCredentialService) GetAccessLog(ctx context.Context, tenantID, credentialID string, limit, offset int) ([]*credential.AccessLog, error) {
	return nil, fmt.Errorf("not implemented")
}
