package user

import (
	"context"
	"fmt"
	"log/slog"
)

// Service handles user business logic
type Service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new user service
func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	s.logger.Info("creating user",
		"email", input.Email,
		"tenant_id", input.TenantID,
	)

	user := &User{
		TenantID:         input.TenantID,
		KratosIdentityID: input.KratosIdentityID,
		Email:            input.Email,
		Role:             input.Role,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user",
			"error", err,
			"email", input.Email,
		)
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.logger.Info("user created successfully",
		"user_id", user.ID,
		"email", user.Email,
	)

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// GetUserByKratosIdentityID retrieves a user by Kratos identity ID
func (s *Service) GetUserByKratosIdentityID(ctx context.Context, kratosIdentityID string) (*User, error) {
	user, err := s.repo.GetByKratosIdentityID(ctx, kratosIdentityID)
	if err != nil {
		return nil, fmt.Errorf("get user by kratos identity id: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

// ListUsersByTenant retrieves all users for a tenant
func (s *Service) ListUsersByTenant(ctx context.Context, tenantID string) ([]*User, error) {
	users, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list users by tenant: %w", err)
	}
	return users, nil
}

// UpdateUser updates a user
func (s *Service) UpdateUser(ctx context.Context, id string, input UpdateUserInput) (*User, error) {
	s.logger.Info("updating user", "user_id", id)

	user, err := s.repo.Update(ctx, id, input)
	if err != nil {
		s.logger.Error("failed to update user",
			"error", err,
			"user_id", id,
		)
		return nil, fmt.Errorf("update user: %w", err)
	}

	s.logger.Info("user updated successfully", "user_id", id)
	return user, nil
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	s.logger.Info("deleting user", "user_id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete user",
			"error", err,
			"user_id", id,
		)
		return fmt.Errorf("delete user: %w", err)
	}

	s.logger.Info("user deleted successfully", "user_id", id)
	return nil
}

// SyncFromKratosWebhook syncs user data from Kratos webhook
func (s *Service) SyncFromKratosWebhook(ctx context.Context, webhook KratosIdentityWebhook) (*User, error) {
	s.logger.Info("syncing user from Kratos webhook",
		"identity_id", webhook.IdentityID,
		"email", webhook.Email,
	)

	// Check if user already exists
	existingUser, err := s.repo.GetByKratosIdentityID(ctx, webhook.IdentityID)
	if err == nil {
		// User exists, update if needed
		s.logger.Info("user already exists, skipping sync",
			"user_id", existingUser.ID,
			"identity_id", webhook.IdentityID,
		)
		return existingUser, nil
	}

	// Determine tenant ID
	tenantID := webhook.TenantID
	if tenantID == "" {
		// For development/testing, use a default tenant
		// In production, this should be required
		s.logger.Warn("no tenant_id in webhook, using default")
		tenantID = "00000000-0000-0000-0000-000000000001"
	}

	// Create new user
	input := CreateUserInput{
		TenantID:         tenantID,
		KratosIdentityID: webhook.IdentityID,
		Email:            webhook.Email,
		Role:             "member", // Default role
	}

	user, err := s.CreateUser(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("sync from kratos webhook: %w", err)
	}

	return user, nil
}
