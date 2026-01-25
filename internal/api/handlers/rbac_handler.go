package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
// @Summary List roles
// @Description Returns all roles configured for the tenant
// @Tags RBAC
// @Accept json
// @Produce json
// @Security TenantID
// @Security UserID
// @Success 200 {array} rbac.Role "List of roles"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/roles [get]
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
// @Summary Create role
// @Description Creates a new custom role with specified permissions
// @Tags RBAC
// @Accept json
// @Produce json
// @Param role body rbac.CreateRoleRequest true "Role creation data"
// @Security TenantID
// @Security UserID
// @Success 201 {object} rbac.Role "Created role"
// @Failure 400 {object} map[string]string "Invalid request or role name"
// @Failure 409 {object} map[string]string "Role already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/roles [post]
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
		if errors.Is(err, rbac.ErrRoleAlreadyExists) {
			respondError(w, http.StatusConflict, "Role already exists")
			return
		}
		if errors.Is(err, rbac.ErrInvalidRoleName) {
			respondError(w, http.StatusBadRequest, "Invalid role name")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create role")
		return
	}

	respondJSON(w, http.StatusCreated, role)
}

// GetRole handles GET /api/v1/roles/:id
// @Summary Get role
// @Description Retrieves details of a specific role
// @Tags RBAC
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Security TenantID
// @Security UserID
// @Success 200 {object} rbac.Role "Role details"
// @Failure 404 {object} map[string]string "Role not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/roles/{id} [get]
func (h *RBACHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	roleID := chi.URLParam(r, "id")

	role, err := h.service.GetRole(r.Context(), roleID, tenantID)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get role")
		return
	}

	respondJSON(w, http.StatusOK, role)
}

// UpdateRole handles PUT /api/v1/roles/:id
// @Summary Update role
// @Description Updates a role's name and description (system roles cannot be modified)
// @Tags RBAC
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body rbac.UpdateRoleRequest true "Role update data"
// @Security TenantID
// @Security UserID
// @Success 204 "Role updated successfully"
// @Failure 400 {object} map[string]string "Invalid request or role name"
// @Failure 403 {object} map[string]string "System role cannot be modified"
// @Failure 404 {object} map[string]string "Role not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/roles/{id} [put]
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
		if errors.Is(err, rbac.ErrRoleNotFound) {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		if errors.Is(err, rbac.ErrSystemRoleCannotBeModified) {
			respondError(w, http.StatusForbidden, "System role cannot be modified")
			return
		}
		if errors.Is(err, rbac.ErrInvalidRoleName) {
			respondError(w, http.StatusBadRequest, "Invalid role name")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update role")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteRole handles DELETE /api/v1/roles/:id
// @Summary Delete role
// @Description Deletes a custom role (system roles cannot be deleted)
// @Tags RBAC
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Security TenantID
// @Security UserID
// @Success 204 "Role deleted successfully"
// @Failure 403 {object} map[string]string "System role cannot be deleted"
// @Failure 404 {object} map[string]string "Role not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/roles/{id} [delete]
func (h *RBACHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := r.Context().Value("user_id").(string)
	roleID := chi.URLParam(r, "id")

	err := h.service.DeleteRole(r.Context(), roleID, tenantID, userID)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		if errors.Is(err, rbac.ErrSystemRoleCannotBeDeleted) {
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
		if errors.Is(err, rbac.ErrRoleNotFound) {
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
		if errors.Is(err, rbac.ErrRoleNotFound) {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		if errors.Is(err, rbac.ErrSystemRoleCannotBeModified) {
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
// @Summary Assign roles to user
// @Description Assigns one or more roles to a user
// @Tags RBAC
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param roles body rbac.AssignRolesRequest true "Role assignment data"
// @Security TenantID
// @Security UserID
// @Success 204 "Roles assigned successfully"
// @Failure 400 {object} map[string]string "Invalid request or no roles provided"
// @Failure 404 {object} map[string]string "Role not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/users/{id}/roles [put]
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
		if errors.Is(err, rbac.ErrRoleNotFound) {
			respondError(w, http.StatusNotFound, "Role not found")
			return
		}
		if errors.Is(err, rbac.ErrNoRolesProvided) {
			respondError(w, http.StatusBadRequest, "No roles provided")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to assign roles")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListPermissions handles GET /api/v1/permissions
// @Summary List all permissions
// @Description Returns all available permissions in the system
// @Tags RBAC
// @Accept json
// @Produce json
// @Success 200 {array} rbac.Permission "List of permissions"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/permissions [get]
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
// @Summary Get RBAC audit logs
// @Description Returns paginated audit logs of role and permission changes
// @Tags RBAC
// @Accept json
// @Produce json
// @Param limit query int false "Maximum results (max 100)" default(50)
// @Param offset query int false "Pagination offset" default(0)
// @Security TenantID
// @Security UserID
// @Success 200 {array} rbac.AuditLog "List of audit logs"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/audit-logs [get]
func (h *RBACHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	// Parse pagination parameters with defaults and bounds
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	logs, err := h.service.GetAuditLogs(r.Context(), tenantID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get audit logs")
		return
	}

	respondJSON(w, http.StatusOK, logs)
}
