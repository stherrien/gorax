package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// Repository defines the interface for user data access
type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByKratosIdentityID(ctx context.Context, kratosIdentityID string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ListByTenant(ctx context.Context, tenantID string) ([]*User, error)
	Update(ctx context.Context, id string, input UpdateUserInput) (*User, error)
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new user repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Set defaults
	if user.Role == "" {
		user.Role = "member"
	}
	if user.Status == "" {
		user.Status = "active"
	}

	query := `
		INSERT INTO users (id, tenant_id, kratos_identity_id, email, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.TenantID,
		user.KratosIdentityID,
		user.Email,
		user.Role,
		user.Status,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_kratos_identity_id_key\"" {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*User, error) {
	var user User
	query := `SELECT * FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (r *repository) GetByKratosIdentityID(ctx context.Context, kratosIdentityID string) (*User, error) {
	var user User
	query := `SELECT * FROM users WHERE kratos_identity_id = $1`

	err := r.db.GetContext(ctx, &user, query, kratosIdentityID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by kratos identity id: %w", err)
	}

	return &user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `SELECT * FROM users WHERE email = $1`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

func (r *repository) ListByTenant(ctx context.Context, tenantID string) ([]*User, error) {
	var users []*User
	query := `SELECT * FROM users WHERE tenant_id = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &users, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list users by tenant: %w", err)
	}

	return users, nil
}

func (r *repository) Update(ctx context.Context, id string, input UpdateUserInput) (*User, error) {
	// Build dynamic update query
	query := `UPDATE users SET updated_at = NOW()`
	args := []interface{}{}
	argCount := 1

	if input.Role != "" {
		query += fmt.Sprintf(", role = $%d", argCount)
		args = append(args, input.Role)
		argCount++
	}

	if input.Status != "" {
		query += fmt.Sprintf(", status = $%d", argCount)
		args = append(args, input.Status)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d RETURNING *", argCount)
	args = append(args, id)

	var user User
	err := r.db.GetContext(ctx, &user, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &user, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
