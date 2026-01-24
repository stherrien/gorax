package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
	"github.com/gorax/gorax/internal/api/response"
	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/validation"
)

// CredentialHandler handles credential-related HTTP requests
type CredentialHandler struct {
	service credential.Service
	logger  *slog.Logger
}

// NewCredentialHandler creates a new credential handler
func NewCredentialHandler(service credential.Service, logger *slog.Logger) *CredentialHandler {
	return &CredentialHandler{
		service: service,
		logger:  logger,
	}
}

// Create creates a new credential
func (h *CredentialHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)

	if tenantID == "" || user == nil {
		_ = response.InternalError(w, "tenant or user context missing")
		return
	}

	var input credential.CreateCredentialInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	cred, err := h.service.Create(r.Context(), tenantID, user.ID, input)
	if err != nil {
		if _, ok := err.(*credential.ValidationError); ok {
			_ = response.BadRequest(w, err.Error())
			return
		}
		h.logger.Error("failed to create credential",
			"error", err,
			"tenant_id", tenantID,
			"user_id", user.ID)
		_ = response.InternalError(w, "failed to create credential")
		return
	}

	_ = response.Created(w, map[string]any{
		"data": cred,
	})
}

// List returns all credentials for the tenant (metadata only)
func (h *CredentialHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	if tenantID == "" {
		_ = response.InternalError(w, "tenant context missing")
		return
	}

	// Parse query parameters
	limit, _ := validation.ParsePaginationLimit(
		r.URL.Query().Get("limit"),
		validation.DefaultPaginationLimit,
		validation.MaxPaginationLimit,
	)
	offset, _ := validation.ParsePaginationOffset(r.URL.Query().Get("offset"))

	// Parse filters
	filter := credential.CredentialListFilter{
		Type:   credential.CredentialType(r.URL.Query().Get("type")),
		Status: credential.CredentialStatus(r.URL.Query().Get("status")),
		Search: r.URL.Query().Get("search"),
	}

	credentials, err := h.service.List(r.Context(), tenantID, filter, limit, offset)
	if err != nil {
		h.logger.Error("failed to list credentials",
			"error", err,
			"tenant_id", tenantID)
		_ = response.InternalError(w, "failed to list credentials")
		return
	}

	_ = response.Paginated(w, credentials, limit, offset, 0)
}

// Get retrieves a single credential's metadata
func (h *CredentialHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" {
		_ = response.InternalError(w, "tenant context missing")
		return
	}

	cred, err := h.service.GetByID(r.Context(), tenantID, credentialID)
	if err != nil {
		if err == credential.ErrNotFound {
			_ = response.NotFound(w, "credential not found")
			return
		}
		h.logger.Error("failed to get credential",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID)
		_ = response.InternalError(w, "failed to get credential")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": cred,
	})
}

// GetValue retrieves the decrypted credential value (restricted access)
func (h *CredentialHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" || user == nil {
		_ = response.InternalError(w, "tenant or user context missing")
		return
	}

	value, err := h.service.GetValue(r.Context(), tenantID, credentialID, user.ID)
	if err != nil {
		if err == credential.ErrNotFound {
			_ = response.NotFound(w, "credential not found")
			return
		}
		if err == credential.ErrUnauthorized {
			_ = response.Forbidden(w, "unauthorized access to credential value")
			return
		}
		h.logger.Error("failed to get credential value",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID,
			"user_id", user.ID)
		_ = response.InternalError(w, "failed to get credential value")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": value,
	})
}

// Update updates a credential's metadata
func (h *CredentialHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" || user == nil {
		_ = response.InternalError(w, "tenant or user context missing")
		return
	}

	var input credential.UpdateCredentialInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	cred, err := h.service.Update(r.Context(), tenantID, credentialID, user.ID, input)
	if err != nil {
		if err == credential.ErrNotFound {
			_ = response.NotFound(w, "credential not found")
			return
		}
		if _, ok := err.(*credential.ValidationError); ok {
			_ = response.BadRequest(w, err.Error())
			return
		}
		h.logger.Error("failed to update credential",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID,
			"user_id", user.ID)
		_ = response.InternalError(w, "failed to update credential")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": cred,
	})
}

// Delete deletes a credential
func (h *CredentialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" || user == nil {
		_ = response.InternalError(w, "tenant or user context missing")
		return
	}

	err := h.service.Delete(r.Context(), tenantID, credentialID, user.ID)
	if err != nil {
		if err == credential.ErrNotFound {
			_ = response.NotFound(w, "credential not found")
			return
		}
		h.logger.Error("failed to delete credential",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID,
			"user_id", user.ID)
		_ = response.InternalError(w, "failed to delete credential")
		return
	}

	response.NoContent(w)
}

// Rotate creates a new version of the credential value
func (h *CredentialHandler) Rotate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" || user == nil {
		_ = response.InternalError(w, "tenant or user context missing")
		return
	}

	var input credential.RotateCredentialInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	cred, err := h.service.Rotate(r.Context(), tenantID, credentialID, user.ID, input)
	if err != nil {
		if err == credential.ErrNotFound {
			_ = response.NotFound(w, "credential not found")
			return
		}
		if _, ok := err.(*credential.ValidationError); ok {
			_ = response.BadRequest(w, err.Error())
			return
		}
		h.logger.Error("failed to rotate credential",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID,
			"user_id", user.ID)
		_ = response.InternalError(w, "failed to rotate credential")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": cred,
	})
}

// ListVersions returns all versions of a credential
func (h *CredentialHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" {
		_ = response.InternalError(w, "tenant context missing")
		return
	}

	versions, err := h.service.ListVersions(r.Context(), tenantID, credentialID)
	if err != nil {
		if err == credential.ErrNotFound {
			_ = response.NotFound(w, "credential not found")
			return
		}
		h.logger.Error("failed to list credential versions",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID)
		_ = response.InternalError(w, "failed to list credential versions")
		return
	}

	_ = response.OK(w, map[string]any{
		"data": versions,
	})
}

// GetAccessLog returns access log entries for a credential
func (h *CredentialHandler) GetAccessLog(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" {
		_ = response.InternalError(w, "tenant context missing")
		return
	}

	limit, _ := validation.ParsePaginationLimit(
		r.URL.Query().Get("limit"),
		validation.DefaultPaginationLimit,
		validation.MaxPaginationLimit,
	)
	offset, _ := validation.ParsePaginationOffset(r.URL.Query().Get("offset"))

	logs, err := h.service.GetAccessLog(r.Context(), tenantID, credentialID, limit, offset)
	if err != nil {
		if err == credential.ErrNotFound {
			_ = response.NotFound(w, "credential not found")
			return
		}
		h.logger.Error("failed to get access log",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID)
		_ = response.InternalError(w, "failed to get access log")
		return
	}

	_ = response.Paginated(w, logs, limit, offset, 0)
}

// Test validates a credential by attempting to decrypt and verify its value
// This endpoint does NOT return the decrypted value, only confirmation of validity
func (h *CredentialHandler) Test(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" || user == nil {
		_ = response.InternalError(w, "tenant or user context missing")
		return
	}

	// Attempt to get the credential value - this verifies decryption works
	_, err := h.service.GetValue(r.Context(), tenantID, credentialID, user.ID)
	if err != nil {
		if err == credential.ErrNotFound {
			_ = response.NotFound(w, "credential not found")
			return
		}
		h.logger.Error("credential test failed",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID)
		_ = response.OK(w, map[string]any{
			"valid":   false,
			"message": "credential validation failed: unable to decrypt",
		})
		return
	}

	_ = response.OK(w, map[string]any{
		"valid":   true,
		"message": "credential is valid and decryptable",
	})
}

// GetTypes returns the list of supported credential types with their schemas
func (h *CredentialHandler) GetTypes(w http.ResponseWriter, r *http.Request) {
	schemas := credential.GetAllCredentialTypeSchemas()

	_ = response.OK(w, map[string]any{
		"data": schemas,
	})
}

// ValidateType validates a credential value against a specific type's schema
// This endpoint does NOT store anything, just validates the structure
func (h *CredentialHandler) ValidateType(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Type  credential.CredentialType `json:"type"`
		Value map[string]any            `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		_ = response.BadRequest(w, "invalid request body")
		return
	}

	if input.Type == "" {
		_ = response.BadRequest(w, "type is required")
		return
	}

	if len(input.Value) == 0 {
		_ = response.BadRequest(w, "value is required")
		return
	}

	// Validate the value against the type's schema
	if err := credential.ValidateCredentialValue(input.Type, input.Value); err != nil {
		_ = response.OK(w, map[string]any{
			"valid":   false,
			"message": err.Error(),
			"schema":  credential.GetCredentialTypeSchema(input.Type),
		})
		return
	}

	_ = response.OK(w, map[string]any{
		"valid":   true,
		"message": "credential value is valid for type " + string(input.Type),
	})
}
