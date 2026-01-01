package handlers

import (
	"context"
	"net/http"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/tenant"
)

// addTenantContext adds tenant context to a request for testing
func addTenantContext(req *http.Request, tenantID string) *http.Request {
	t := &tenant.Tenant{ID: tenantID}
	ctx := context.WithValue(req.Context(), middleware.TenantContextKey, t)
	return req.WithContext(ctx)
}
