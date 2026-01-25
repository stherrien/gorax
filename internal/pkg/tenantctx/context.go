// Package tenantctx provides utilities for managing tenant context throughout request lifecycle
package tenantctx

import (
	"context"
	"errors"
)

// contextKey is a private type used for context keys to prevent collisions
type contextKey string

const (
	// tenantIDKey is the context key for storing the tenant ID
	tenantIDKey contextKey = "tenant_id"
	// originalTenantIDKey stores the original tenant ID when admin switches tenant
	originalTenantIDKey contextKey = "original_tenant_id"
	// isSwitchedKey indicates if an admin has switched to a different tenant
	isSwitchedKey contextKey = "tenant_switched"
)

// ErrNoTenant is returned when no tenant ID is found in context
var ErrNoTenant = errors.New("no tenant ID in context")

// WithTenantID returns a new context with the tenant ID set
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// GetTenantID retrieves the tenant ID from the context
// Returns an empty string if not present
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(tenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

// MustGetTenantID retrieves the tenant ID from the context or returns an error
func MustGetTenantID(ctx context.Context) (string, error) {
	tenantID := GetTenantID(ctx)
	if tenantID == "" {
		return "", ErrNoTenant
	}
	return tenantID, nil
}

// WithSwitchedTenant returns a new context with a switched tenant ID,
// preserving the original tenant ID for audit purposes
func WithSwitchedTenant(ctx context.Context, newTenantID string) context.Context {
	originalTenantID := GetTenantID(ctx)
	ctx = context.WithValue(ctx, originalTenantIDKey, originalTenantID)
	ctx = context.WithValue(ctx, isSwitchedKey, true)
	return WithTenantID(ctx, newTenantID)
}

// GetOriginalTenantID returns the original tenant ID if a switch occurred
// Returns empty string if no switch has occurred
func GetOriginalTenantID(ctx context.Context) string {
	if originalID, ok := ctx.Value(originalTenantIDKey).(string); ok {
		return originalID
	}
	return ""
}

// IsTenantSwitched returns true if an admin has switched to a different tenant
func IsTenantSwitched(ctx context.Context) bool {
	if switched, ok := ctx.Value(isSwitchedKey).(bool); ok {
		return switched
	}
	return false
}

// ResetToOriginalTenant returns a context with the original tenant ID restored
// If no switch occurred, returns the context unchanged
func ResetToOriginalTenant(ctx context.Context) context.Context {
	if !IsTenantSwitched(ctx) {
		return ctx
	}
	originalID := GetOriginalTenantID(ctx)
	if originalID == "" {
		return ctx
	}
	// Reset without the switched flag
	ctx = context.WithValue(ctx, isSwitchedKey, false)
	return WithTenantID(ctx, originalID)
}
