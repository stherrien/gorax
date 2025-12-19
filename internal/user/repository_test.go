package user

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	// This would connect to a test database
	// For now, skip if DB_TEST_URL is not set
	t.Skip("Skipping integration test - set DB_TEST_URL environment variable to run")
	return nil
}

func TestRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			name: "valid user",
			user: &User{
				TenantID:         "00000000-0000-0000-0000-000000000001",
				KratosIdentityID: "test-identity-1",
				Email:            "test@example.com",
				Role:             "member",
			},
			wantErr: false,
		},
		{
			name: "duplicate kratos identity id",
			user: &User{
				TenantID:         "00000000-0000-0000-0000-000000000001",
				KratosIdentityID: "test-identity-1",
				Email:            "test2@example.com",
				Role:             "member",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepository_GetByKratosIdentityID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test user
	testUser := &User{
		TenantID:         "00000000-0000-0000-0000-000000000001",
		KratosIdentityID: "test-identity-get",
		Email:            "get@example.com",
		Role:             "member",
	}
	_ = repo.Create(ctx, testUser)

	tests := []struct {
		name             string
		kratosIdentityID string
		wantErr          bool
	}{
		{
			name:             "existing user",
			kratosIdentityID: "test-identity-get",
			wantErr:          false,
		},
		{
			name:             "non-existing user",
			kratosIdentityID: "non-existent",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByKratosIdentityID(ctx, tt.kratosIdentityID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByKratosIdentityID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && user == nil {
				t.Error("GetByKratosIdentityID() returned nil user")
			}
		})
	}
}

func TestRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	ctx := context.Background()

	// Create test user
	testUser := &User{
		TenantID:         "00000000-0000-0000-0000-000000000001",
		KratosIdentityID: "test-identity-update",
		Email:            "update@example.com",
		Role:             "member",
	}
	_ = repo.Create(ctx, testUser)

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
			id:   "non-existent-id",
			input: UpdateUserInput{
				Role: "admin",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.Update(ctx, tt.id, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && user == nil {
				t.Error("Update() returned nil user")
			}
		})
	}
}
