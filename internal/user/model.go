package user

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID               string    `db:"id" json:"id"`
	TenantID         string    `db:"tenant_id" json:"tenant_id"`
	KratosIdentityID string    `db:"kratos_identity_id" json:"kratos_identity_id"`
	Email            string    `db:"email" json:"email"`
	Role             string    `db:"role" json:"role"`
	Status           string    `db:"status" json:"status"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	TenantID         string `json:"tenant_id" validate:"required,uuid"`
	KratosIdentityID string `json:"kratos_identity_id" validate:"required"`
	Email            string `json:"email" validate:"required,email"`
	Role             string `json:"role" validate:"oneof=owner admin member"`
}

// UpdateUserInput represents input for updating a user
type UpdateUserInput struct {
	Role   string `json:"role,omitempty" validate:"omitempty,oneof=owner admin member"`
	Status string `json:"status,omitempty" validate:"omitempty,oneof=active inactive suspended"`
}

// KratosIdentityWebhook represents the webhook payload from Kratos
type KratosIdentityWebhook struct {
	IdentityID string                 `json:"identity_id"`
	Email      string                 `json:"email"`
	Name       map[string]interface{} `json:"name"`
	TenantID   string                 `json:"tenant_id"`
	CreatedAt  string                 `json:"created_at"`
	UpdatedAt  string                 `json:"updated_at"`
}
