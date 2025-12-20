package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/api/middleware"
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
		h.respondError(w, http.StatusInternalServerError, "tenant or user context missing")
		return
	}

	var input credential.CreateCredentialInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cred, err := h.service.Create(r.Context(), tenantID, user.ID, input)
	if err != nil {
		if _, ok := err.(*credential.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("failed to create credential",
			"error", err,
			"tenant_id", tenantID,
			"user_id", user.ID)
		h.respondError(w, http.StatusInternalServerError, "failed to create credential")
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"data": cred,
	})
}

// List returns all credentials for the tenant (metadata only)
func (h *CredentialHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)

	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant context missing")
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
		h.respondError(w, http.StatusInternalServerError, "failed to list credentials")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   credentials,
		"limit":  limit,
		"offset": offset,
	})
}

// Get retrieves a single credential's metadata
func (h *CredentialHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant context missing")
		return
	}

	cred, err := h.service.GetByID(r.Context(), tenantID, credentialID)
	if err != nil {
		if err == credential.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "credential not found")
			return
		}
		h.logger.Error("failed to get credential",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID)
		h.respondError(w, http.StatusInternalServerError, "failed to get credential")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": cred,
	})
}

// GetValue retrieves the decrypted credential value (restricted access)
func (h *CredentialHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" || user == nil {
		h.respondError(w, http.StatusInternalServerError, "tenant or user context missing")
		return
	}

	value, err := h.service.GetValue(r.Context(), tenantID, credentialID, user.ID)
	if err != nil {
		if err == credential.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "credential not found")
			return
		}
		if err == credential.ErrUnauthorized {
			h.respondError(w, http.StatusForbidden, "unauthorized access to credential value")
			return
		}
		h.logger.Error("failed to get credential value",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID,
			"user_id", user.ID)
		h.respondError(w, http.StatusInternalServerError, "failed to get credential value")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": value,
	})
}

// Update updates a credential's metadata
func (h *CredentialHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" || user == nil {
		h.respondError(w, http.StatusInternalServerError, "tenant or user context missing")
		return
	}

	var input credential.UpdateCredentialInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cred, err := h.service.Update(r.Context(), tenantID, credentialID, user.ID, input)
	if err != nil {
		if err == credential.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "credential not found")
			return
		}
		if _, ok := err.(*credential.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("failed to update credential",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID,
			"user_id", user.ID)
		h.respondError(w, http.StatusInternalServerError, "failed to update credential")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": cred,
	})
}

// Delete deletes a credential
func (h *CredentialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" || user == nil {
		h.respondError(w, http.StatusInternalServerError, "tenant or user context missing")
		return
	}

	err := h.service.Delete(r.Context(), tenantID, credentialID, user.ID)
	if err != nil {
		if err == credential.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "credential not found")
			return
		}
		h.logger.Error("failed to delete credential",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID,
			"user_id", user.ID)
		h.respondError(w, http.StatusInternalServerError, "failed to delete credential")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Rotate creates a new version of the credential value
func (h *CredentialHandler) Rotate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	user := middleware.GetUser(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" || user == nil {
		h.respondError(w, http.StatusInternalServerError, "tenant or user context missing")
		return
	}

	var input credential.RotateCredentialInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cred, err := h.service.Rotate(r.Context(), tenantID, credentialID, user.ID, input)
	if err != nil {
		if err == credential.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "credential not found")
			return
		}
		if _, ok := err.(*credential.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("failed to rotate credential",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID,
			"user_id", user.ID)
		h.respondError(w, http.StatusInternalServerError, "failed to rotate credential")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": cred,
	})
}

// ListVersions returns all versions of a credential
func (h *CredentialHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant context missing")
		return
	}

	versions, err := h.service.ListVersions(r.Context(), tenantID, credentialID)
	if err != nil {
		if err == credential.ErrNotFound {
			h.respondError(w, http.StatusNotFound, "credential not found")
			return
		}
		h.logger.Error("failed to list credential versions",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID)
		h.respondError(w, http.StatusInternalServerError, "failed to list credential versions")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": versions,
	})
}

// GetAccessLog returns access log entries for a credential
func (h *CredentialHandler) GetAccessLog(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r)
	credentialID := chi.URLParam(r, "credentialID")

	if tenantID == "" {
		h.respondError(w, http.StatusInternalServerError, "tenant context missing")
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
			h.respondError(w, http.StatusNotFound, "credential not found")
			return
		}
		h.logger.Error("failed to get access log",
			"error", err,
			"tenant_id", tenantID,
			"credential_id", credentialID)
		h.respondError(w, http.StatusInternalServerError, "failed to get access log")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

// Helper methods

func (h *CredentialHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *CredentialHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
