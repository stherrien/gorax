package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gorax/gorax/internal/rbac"
)

// RBACHandler handles RBAC-related HTTP requests
type RBACHandler struct {
	service *rbac.Service
}

// NewRBACHandler creates a new RBAC handler
func NewRBACHandler(service *rbac.Service) *RBACHandler {
	return &RBACHandler{
		service: service,
	}
}

// ListRoles handles GET /api/v1/roles
func (h *RBACHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	roles, err := h.service.ListRoles(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list roles")
		return
	}

	respondJSON(w, http.StatusOK, roles)
}

// CreateRole handles POST /api/v1/roles
func (h *RBACHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := r.Context().Value("user_id").(string)

	var req rbac.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	role, err := h.service.CreateRole(r.Context(), tenantID, userID, &req)
	if err != nil {
		if err == rbac.ErrRoleAlreadyExists {
			respondError(w, http.StatusConflict, "Role already exists")
			return
		}
		if err == rbac.ErrInvalidRoleName {
			respondError(w, http.StatusBadRequest, "Invalid role name")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create role")
		return
	}

	respondJSON(w, http.StatusCreated, role)
}

// GetRole handles GET /api/v1/roles/:id
func (h *RBACHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	roleID := chi.URLParam(r, "id")

	role, err := h.service.GetRole(r.Context(), roleID, tenantID)
	if err != nil {
		if err == rbac.ErrRoleNotFound {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get role")
		return
	}

	respondJSON(w, http.StatusOK, role)
}

// UpdateRole handles PUT /api/v1/roles/:id
func (h *RBACHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := r.Context().Value("user_id").(string)
	roleID := chi.URLParam(r, "id")

	var req rbac.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.UpdateRole(r.Context(), roleID, tenantID, userID, &req)
	if err != nil {
		if err == rbac.ErrRoleNotFound {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		if err == rbac.ErrSystemRoleCannotBeModified {
			respondError(w, http.StatusForbidden, "System role cannot be modified")
			return
		}
		if err == rbac.ErrInvalidRoleName {
			respondError(w, http.StatusBadRequest, "Invalid role name")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update role")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteRole handles DELETE /api/v1/roles/:id
func (h *RBACHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := r.Context().Value("user_id").(string)
	roleID := chi.URLParam(r, "id")

	err := h.service.DeleteRole(r.Context(), roleID, tenantID, userID)
	if err != nil {
		if err == rbac.ErrRoleNotFound {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		if err == rbac.ErrSystemRoleCannotBeDeleted {
			respondError(w, http.StatusForbidden, "System role cannot be deleted")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete role")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRolePermissions handles GET /api/v1/roles/:id/permissions
func (h *RBACHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	roleID := chi.URLParam(r, "id")

	permissions, err := h.service.GetRolePermissions(r.Context(), roleID, tenantID)
	if err != nil {
		if err == rbac.ErrRoleNotFound {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get role permissions")
		return
	}

	respondJSON(w, http.StatusOK, permissions)
}

// UpdateRolePermissions handles PUT /api/v1/roles/:id/permissions
func (h *RBACHandler) UpdateRolePermissions(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := r.Context().Value("user_id").(string)
	roleID := chi.URLParam(r, "id")

	var req struct {
		PermissionIDs []string `json:"permission_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.UpdateRolePermissions(r.Context(), roleID, tenantID, userID, req.PermissionIDs)
	if err != nil {
		if err == rbac.ErrRoleNotFound {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		if err == rbac.ErrSystemRoleCannotBeModified {
			respondError(w, http.StatusForbidden, "System role cannot be modified")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update role permissions")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserRoles handles GET /api/v1/users/:id/roles
func (h *RBACHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := chi.URLParam(r, "id")

	roles, err := h.service.GetUserRoles(r.Context(), userID, tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get user roles")
		return
	}

	respondJSON(w, http.StatusOK, roles)
}

// AssignUserRoles handles PUT /api/v1/users/:id/roles
func (h *RBACHandler) AssignUserRoles(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	grantedBy := r.Context().Value("user_id").(string)
	userID := chi.URLParam(r, "id")

	var req rbac.AssignRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.service.AssignRolesToUser(r.Context(), userID, tenantID, grantedBy, &req)
	if err != nil {
		if err == rbac.ErrRoleNotFound {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		if err == rbac.ErrNoRolesProvided {
			respondError(w, http.StatusBadRequest, "No roles provided")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to assign roles")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListPermissions handles GET /api/v1/permissions
func (h *RBACHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.service.ListPermissions(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list permissions")
		return
	}

	respondJSON(w, http.StatusOK, permissions)
}

// GetUserPermissions handles GET /api/v1/users/:id/permissions
func (h *RBACHandler) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := chi.URLParam(r, "id")

	permissions, err := h.service.GetUserPermissions(r.Context(), userID, tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get user permissions")
		return
	}

	respondJSON(w, http.StatusOK, permissions)
}

// GetCurrentUserPermissions handles GET /api/v1/me/permissions
func (h *RBACHandler) GetCurrentUserPermissions(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := r.Context().Value("user_id").(string)

	permissions, err := h.service.GetUserPermissions(r.Context(), userID, tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get user permissions")
		return
	}

	respondJSON(w, http.StatusOK, permissions)
}

// GetAuditLogs handles GET /api/v1/audit-logs
func (h *RBACHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	// Parse pagination parameters
	limit := 50
	offset := 0
	// TODO: Parse from query params

	logs, err := h.service.GetAuditLogs(r.Context(), tenantID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get audit logs")
		return
	}

	respondJSON(w, http.StatusOK, logs)
}
