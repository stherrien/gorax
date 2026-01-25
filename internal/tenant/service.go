package tenant

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gorax/gorax/internal/pkg/tenantctx"
)

// Service handles tenant business logic
type Service struct {
	repo   *Repository
	logger *slog.Logger
}

// NewService creates a new tenant service
func NewService(repo *Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new tenant
func (s *Service) Create(ctx context.Context, input CreateTenantInput) (*Tenant, error) {
	// Set default tier if not specified
	if input.Tier == "" {
		input.Tier = "free"
	}

	tenant, err := s.repo.Create(ctx, input)
	if err != nil {
		s.logger.Error("failed to create tenant", "error", err, "name", input.Name)
		return nil, err
	}

	s.logger.Info("tenant created", "tenant_id", tenant.ID, "name", tenant.Name)
	return tenant, nil
}

// GetByID retrieves a tenant by ID
func (s *Service) GetByID(ctx context.Context, id string) (*Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

// GetBySubdomain retrieves a tenant by subdomain
func (s *Service) GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error) {
	return s.repo.GetBySubdomain(ctx, subdomain)
}

// Update updates a tenant
func (s *Service) Update(ctx context.Context, id string, input UpdateTenantInput) (*Tenant, error) {
	tenant, err := s.repo.Update(ctx, id, input)
	if err != nil {
		s.logger.Error("failed to update tenant", "error", err, "tenant_id", id)
		return nil, err
	}

	s.logger.Info("tenant updated", "tenant_id", tenant.ID)
	return tenant, nil
}

// Delete deletes a tenant
func (s *Service) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Error("failed to delete tenant", "error", err, "tenant_id", id)
		return err
	}

	s.logger.Info("tenant deleted", "tenant_id", id)
	return nil
}

// List retrieves all tenants with pagination
func (s *Service) List(ctx context.Context, limit, offset int) ([]*Tenant, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.List(ctx, limit, offset)
}

// UpdateQuotas updates tenant quotas
func (s *Service) UpdateQuotas(ctx context.Context, id string, quotas TenantQuotas) (*Tenant, error) {
	tenant, err := s.repo.UpdateQuotas(ctx, id, quotas)
	if err != nil {
		s.logger.Error("failed to update tenant quotas", "error", err, "tenant_id", id)
		return nil, err
	}

	s.logger.Info("tenant quotas updated", "tenant_id", tenant.ID)
	return tenant, nil
}

// GetWorkflowCount returns the count of active workflows for a tenant
func (s *Service) GetWorkflowCount(ctx context.Context, tenantID string) (int, error) {
	return s.repo.GetWorkflowCount(ctx, tenantID)
}

// GetExecutionStats returns execution statistics for a tenant
func (s *Service) GetExecutionStats(ctx context.Context, tenantID string) (*UsageStats, error) {
	return s.repo.GetExecutionStats(ctx, tenantID)
}

// GetConcurrentExecutions returns the count of currently running executions
func (s *Service) GetConcurrentExecutions(ctx context.Context, tenantID string) (int, error) {
	return s.repo.GetConcurrentExecutions(ctx, tenantID)
}

// Count returns the total number of active tenants
func (s *Service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// SwitchTenant switches the tenant context for an admin user
// Returns a new context with the switched tenant ID
func (s *Service) SwitchTenant(ctx context.Context, targetTenantID string) (context.Context, *Tenant, error) {
	// Verify target tenant exists and is active
	targetTenant, err := s.repo.GetByID(ctx, targetTenantID)
	if err != nil {
		if err == ErrNotFound {
			return ctx, nil, fmt.Errorf("target tenant not found: %w", err)
		}
		return ctx, nil, fmt.Errorf("failed to get target tenant: %w", err)
	}

	if targetTenant.Status != string(StatusActive) {
		return ctx, nil, fmt.Errorf("cannot switch to inactive tenant")
	}

	// Create new context with switched tenant
	newCtx := tenantctx.WithSwitchedTenant(ctx, targetTenantID)

	s.logger.Info("tenant context switched",
		"original_tenant_id", tenantctx.GetOriginalTenantID(newCtx),
		"target_tenant_id", targetTenantID,
	)

	return newCtx, targetTenant, nil
}

// GetOrCreateDefault gets the default tenant or creates one if it doesn't exist
// This is used in single-tenant mode to ensure a default tenant always exists
func (s *Service) GetOrCreateDefault(ctx context.Context) (*Tenant, error) {
	// Try to get existing default tenant
	tenant, err := s.repo.GetBySubdomain(ctx, DefaultTenantSubdomain)
	if err == nil {
		return tenant, nil
	}

	if err != ErrNotFound {
		return nil, fmt.Errorf("failed to check for default tenant: %w", err)
	}

	// Create default tenant
	input := CreateTenantInput{
		Name:      DefaultTenantName,
		Subdomain: DefaultTenantSubdomain,
		Tier:      string(TierEnterprise), // Default tenant gets enterprise tier
	}

	tenant, err = s.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create default tenant: %w", err)
	}

	s.logger.Info("default tenant created", "tenant_id", tenant.ID)
	return tenant, nil
}

// SetStatus updates the status of a tenant
func (s *Service) SetStatus(ctx context.Context, id string, status TenantStatus) (*Tenant, error) {
	if !IsValidStatus(string(status)) {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	input := UpdateTenantInput{
		Status: string(status),
	}

	tenant, err := s.repo.Update(ctx, id, input)
	if err != nil {
		s.logger.Error("failed to set tenant status", "error", err, "tenant_id", id, "status", status)
		return nil, err
	}

	s.logger.Info("tenant status updated", "tenant_id", tenant.ID, "status", status)
	return tenant, nil
}

// Activate activates a tenant
func (s *Service) Activate(ctx context.Context, id string) (*Tenant, error) {
	return s.SetStatus(ctx, id, StatusActive)
}

// Suspend suspends a tenant
func (s *Service) Suspend(ctx context.Context, id string) (*Tenant, error) {
	return s.SetStatus(ctx, id, StatusSuspended)
}

// Deactivate deactivates a tenant
func (s *Service) Deactivate(ctx context.Context, id string) (*Tenant, error) {
	return s.SetStatus(ctx, id, StatusInactive)
}
