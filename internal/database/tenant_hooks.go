package database

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// TenantContextKey is the context key for storing tenant ID
type TenantContextKey string

const (
	// ContextKeyTenantID is the key used to store tenant ID in context
	ContextKeyTenantID TenantContextKey = "tenant_id"
)

// TenantDB wraps sqlx.DB with tenant-aware hooks
type TenantDB struct {
	*sqlx.DB
}

// NewTenantDB creates a new tenant-aware database wrapper
func NewTenantDB(db *sqlx.DB) *TenantDB {
	return &TenantDB{DB: db}
}

// ExecContext executes a query with automatic tenant_id injection
func (db *TenantDB) ExecContext(ctx context.Context, query string, args ...interface{}) (driver.Result, error) {
	tenantID := GetTenantIDFromContext(ctx)
	if tenantID != "" && shouldInjectTenantID(query) {
		// Set tenant context in session before executing query
		_, err := db.DB.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to set tenant context: %w", err)
		}
	}
	return db.DB.ExecContext(ctx, query, args...)
}

// QueryContext executes a query with automatic tenant_id injection
func (db *TenantDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	tenantID := GetTenantIDFromContext(ctx)
	if tenantID != "" && shouldInjectTenantID(query) {
		// Set tenant context in session before executing query
		_, err := db.DB.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to set tenant context: %w", err)
		}
	}
	return db.DB.QueryxContext(ctx, query, args...)
}

// QueryRowContext executes a query returning a single row with automatic tenant_id injection
func (db *TenantDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	tenantID := GetTenantIDFromContext(ctx)
	if tenantID != "" && shouldInjectTenantID(query) {
		// Set tenant context in session before executing query
		_, err := db.DB.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
		if err != nil {
			// Return a row with the error
			return db.DB.QueryRowxContext(ctx, "SELECT $1::text", err.Error())
		}
	}
	return db.DB.QueryRowxContext(ctx, query, args...)
}

// GetContext is a helper that uses QueryRowContext
func (db *TenantDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	tenantID := GetTenantIDFromContext(ctx)
	if tenantID != "" && shouldInjectTenantID(query) {
		// Set tenant context in session before executing query
		_, err := db.DB.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
		if err != nil {
			return fmt.Errorf("failed to set tenant context: %w", err)
		}
	}
	return db.DB.GetContext(ctx, dest, query, args...)
}

// SelectContext is a helper that uses QueryContext
func (db *TenantDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	tenantID := GetTenantIDFromContext(ctx)
	if tenantID != "" && shouldInjectTenantID(query) {
		// Set tenant context in session before executing query
		_, err := db.DB.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
		if err != nil {
			return fmt.Errorf("failed to set tenant context: %w", err)
		}
	}
	return db.DB.SelectContext(ctx, dest, query, args...)
}

// BeginTxx begins a transaction with tenant context
func (db *TenantDB) BeginTxx(ctx context.Context, opts *driver.TxOptions) (*sqlx.Tx, error) {
	tx, err := db.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Set tenant context in transaction
	tenantID := GetTenantIDFromContext(ctx)
	if tenantID != "" {
		_, err = tx.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to set tenant context in transaction: %w", err)
		}
	}

	return tx, nil
}

// TenantScoped returns a new context with the tenant ID set
func TenantScoped(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, ContextKeyTenantID, tenantID)
}

// GetTenantIDFromContext extracts the tenant ID from the context
func GetTenantIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	tenantID, _ := ctx.Value(ContextKeyTenantID).(string)
	return tenantID
}

// shouldInjectTenantID checks if the query needs tenant_id injection
// This is a simple heuristic - skip admin queries and DDL statements
func shouldInjectTenantID(query string) bool {
	queryLower := strings.ToLower(strings.TrimSpace(query))

	// Skip if querying tenants table itself (admin operations)
	if strings.Contains(queryLower, "from tenants") || strings.Contains(queryLower, "into tenants") {
		return false
	}

	// Skip DDL statements
	if strings.HasPrefix(queryLower, "create") ||
		strings.HasPrefix(queryLower, "alter") ||
		strings.HasPrefix(queryLower, "drop") {
		return false
	}

	// Skip if already setting tenant context
	if strings.Contains(queryLower, "app.current_tenant_id") {
		return false
	}

	// Skip SHOW/SET statements
	if strings.HasPrefix(queryLower, "show") || strings.HasPrefix(queryLower, "set") {
		return false
	}

	return true
}

// WithTenantID is a helper to wrap a database operation with tenant context
func WithTenantID(ctx context.Context, tenantID string, fn func(context.Context) error) error {
	tenantCtx := TenantScoped(ctx, tenantID)
	return fn(tenantCtx)
}
