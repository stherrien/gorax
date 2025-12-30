package graphql

import (
	"context"
	"errors"
)

// Context keys for extracting values from context
type contextKey string

const (
	tenantIDKey contextKey = "tenant_id"
	userIDKey   contextKey = "user_id"
)

// getTenantID extracts tenant ID from context (set by middleware)
func getTenantID(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(tenantIDKey).(string)
	if !ok || tenantID == "" {
		return "", errors.New("tenant ID not found in context")
	}
	return tenantID, nil
}

// getUserID extracts user ID from context (set by middleware)
func getUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}
