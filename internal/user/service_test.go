package user

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
)

// MockRepository implements Repository for testing
type MockRepository struct {
	users                map[string]*User
	usersByKratosID      map[string]*User
	usersByEmail         map[string]*User
	createFunc           func(ctx context.Context, user *User) error
	getByIDFunc          func(ctx context.Context, id string) (*User, error)
	getByKratosIDFunc    func(ctx context.Context, kratosID string) (*User, error)
	getByEmailFunc       func(ctx context.Context, email string) (*User, error)
	listByTenantFunc     func(ctx context.Context, tenantID string) ([]*User, error)
	updateFunc           func(ctx context.Context, id string, input UpdateUserInput) (*User, error)
	deleteFunc           func(ctx context.Context, id string) error
}

func (m *MockRepository) Create(ctx context.Context, user *User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	if user.ID == "" {
		user.ID = "test-user-id"
	}
	m.users[user.ID] = user
	m.usersByKratosID[user.KratosIdentityID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	user, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (m *MockRepository) GetByKratosIdentityID(ctx context.Context, kratosID string) (*User, error) {
	if m.getByKratosIDFunc != nil {
		return m.getByKratosIDFunc(ctx, kratosID)
	}
	user, ok := m.usersByKratosID[kratosID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (m *MockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	user, ok := m.usersByEmail[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (m *MockRepository) ListByTenant(ctx context.Context, tenantID string) ([]*User, error) {
	if m.listByTenantFunc != nil {
		return m.listByTenantFunc(ctx, tenantID)
	}
	var users []*User
	for _, user := range m.users {
		if user.TenantID == tenantID {
			users = append(users, user)
		}
	}
	return users, nil
}

func (m *MockRepository) Update(ctx context.Context, id string, input UpdateUserInput) (*User, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, input)
	}
	user, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	if input.Role != "" {
		user.Role = input.Role
	}
	if input.Status != "" {
		user.Status = input.Status
	}
	return user, nil
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	user, ok := m.users[id]
	if !ok {
		return ErrUserNotFound
	}
	delete(m.users, id)
	delete(m.usersByKratosID, user.KratosIdentityID)
	delete(m.usersByEmail, user.Email)
	return nil
}

func newMockRepository() *MockRepository {
	return &MockRepository{
		users:           make(map[string]*User),
		usersByKratosID: make(map[string]*User),
		usersByEmail:    make(map[string]*User),
	}
}

func TestService_CreateUser(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := newMockRepository()
	service := NewService(repo, logger)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   CreateUserInput
		wantErr bool
	}{
		{
			name: "valid user",
			input: CreateUserInput{
				TenantID:         "tenant-1",
				KratosIdentityID: "kratos-1",
				Email:            "user1@example.com",
				Role:             "member",
			},
			wantErr: false,
		},
		{
			name: "duplicate kratos identity id",
			input: CreateUserInput{
				TenantID:         "tenant-1",
				KratosIdentityID: "kratos-1",
				Email:            "user2@example.com",
				Role:             "member",
			},
			wantErr: false, // Mock doesn't enforce uniqueness by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.CreateUser(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && user == nil {
				t.Error("CreateUser() returned nil user")
			}
			if !tt.wantErr {
				if user.Email != tt.input.Email {
					t.Errorf("CreateUser() email = %v, want %v", user.Email, tt.input.Email)
				}
			}
		})
	}
}

func TestService_GetUserByKratosIdentityID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := newMockRepository()
	service := NewService(repo, logger)
	ctx := context.Background()

	// Create test user
	testUser, _ := service.CreateUser(ctx, CreateUserInput{
		TenantID:         "tenant-1",
		KratosIdentityID: "kratos-test",
		Email:            "test@example.com",
		Role:             "member",
	})

	tests := []struct {
		name             string
		kratosIdentityID string
		wantErr          bool
	}{
		{
			name:             "existing user",
			kratosIdentityID: "kratos-test",
			wantErr:          false,
		},
		{
			name:             "non-existing user",
			kratosIdentityID: "kratos-nonexistent",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.GetUserByKratosIdentityID(ctx, tt.kratosIdentityID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByKratosIdentityID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && user == nil {
				t.Error("GetUserByKratosIdentityID() returned nil user")
			}
			if !tt.wantErr && user.ID != testUser.ID {
				t.Errorf("GetUserByKratosIdentityID() user.ID = %v, want %v", user.ID, testUser.ID)
			}
		})
	}
}

func TestService_SyncFromKratosWebhook(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := newMockRepository()
	service := NewService(repo, logger)
	ctx := context.Background()

	tests := []struct {
		name    string
		webhook KratosIdentityWebhook
		setup   func()
		wantErr bool
	}{
		{
			name: "new user",
			webhook: KratosIdentityWebhook{
				IdentityID: "new-kratos-id",
				Email:      "newuser@example.com",
				TenantID:   "tenant-1",
			},
			setup:   func() {},
			wantErr: false,
		},
		{
			name: "existing user",
			webhook: KratosIdentityWebhook{
				IdentityID: "existing-kratos-id",
				Email:      "existing@example.com",
				TenantID:   "tenant-1",
			},
			setup: func() {
				service.CreateUser(ctx, CreateUserInput{
					TenantID:         "tenant-1",
					KratosIdentityID: "existing-kratos-id",
					Email:            "existing@example.com",
					Role:             "member",
				})
			},
			wantErr: false,
		},
		{
			name: "webhook without tenant_id",
			webhook: KratosIdentityWebhook{
				IdentityID: "no-tenant-kratos-id",
				Email:      "notenant@example.com",
				TenantID:   "",
			},
			setup:   func() {},
			wantErr: false, // Should use default tenant
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			user, err := service.SyncFromKratosWebhook(ctx, tt.webhook)
			if (err != nil) != tt.wantErr {
				t.Errorf("SyncFromKratosWebhook() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && user == nil {
				t.Error("SyncFromKratosWebhook() returned nil user")
			}
			if !tt.wantErr && user.Email != tt.webhook.Email {
				t.Errorf("SyncFromKratosWebhook() user.Email = %v, want %v", user.Email, tt.webhook.Email)
			}
		})
	}
}

func TestService_UpdateUser(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := newMockRepository()
	service := NewService(repo, logger)
	ctx := context.Background()

	// Create test user
	testUser, _ := service.CreateUser(ctx, CreateUserInput{
		TenantID:         "tenant-1",
		KratosIdentityID: "kratos-update",
		Email:            "update@example.com",
		Role:             "member",
	})

	tests := []struct {
		name    string
		id      string
		input   UpdateUserInput
		wantErr bool
	}{
		{
			name: "update role",
			id:   testUser.ID,
			input: UpdateUserInput{
				Role: "admin",
			},
			wantErr: false,
		},
		{
			name: "update status",
			id:   testUser.ID,
			input: UpdateUserInput{
				Status: "inactive",
			},
			wantErr: false,
		},
		{
			name: "non-existing user",
			id:   "non-existent",
			input: UpdateUserInput{
				Role: "admin",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.UpdateUser(ctx, tt.id, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && user == nil {
				t.Error("UpdateUser() returned nil user")
			}
			if !tt.wantErr && tt.input.Role != "" && user.Role != tt.input.Role {
				t.Errorf("UpdateUser() user.Role = %v, want %v", user.Role, tt.input.Role)
			}
		})
	}
}

func TestService_DeleteUser(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := newMockRepository()
	service := NewService(repo, logger)
	ctx := context.Background()

	// Create test user
	testUser, _ := service.CreateUser(ctx, CreateUserInput{
		TenantID:         "tenant-1",
		KratosIdentityID: "kratos-delete",
		Email:            "delete@example.com",
		Role:             "member",
	})

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "existing user",
			id:      testUser.ID,
			wantErr: false,
		},
		{
			name:    "non-existing user",
			id:      "non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteUser(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_CreateUser_RepositoryError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := newMockRepository()

	// Mock repository error
	repo.createFunc = func(ctx context.Context, user *User) error {
		return errors.New("database error")
	}

	service := NewService(repo, logger)
	ctx := context.Background()

	_, err := service.CreateUser(ctx, CreateUserInput{
		TenantID:         "tenant-1",
		KratosIdentityID: "kratos-error",
		Email:            "error@example.com",
		Role:             "member",
	})

	if err == nil {
		t.Error("CreateUser() expected error, got nil")
	}
}
