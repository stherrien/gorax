import { apiClient } from './client';

export interface Permission {
  id: string;
  resource: string;
  action: string;
  description: string;
  created_at: string;
}

export interface Role {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
  permissions?: Permission[];
}

export interface CreateRoleRequest {
  name: string;
  description: string;
  permission_ids?: string[];
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
  permission_ids?: string[];
}

export interface AssignRolesRequest {
  role_ids: string[];
}

export interface AuditLog {
  id: string;
  tenant_id: string;
  user_id: string;
  action: string;
  target_type?: string;
  target_id?: string;
  details: Record<string, any>;
  created_at: string;
}

export const rbacApi = {
  // Roles
  listRoles: async (): Promise<Role[]> => {
    const response = await apiClient.get('/api/v1/roles');
    return response.data;
  },

  createRole: async (data: CreateRoleRequest): Promise<Role> => {
    const response = await apiClient.post('/api/v1/roles', data);
    return response.data;
  },

  getRole: async (id: string): Promise<Role> => {
    const response = await apiClient.get(`/api/v1/roles/${id}`);
    return response.data;
  },

  updateRole: async (id: string, data: UpdateRoleRequest): Promise<void> => {
    await apiClient.put(`/api/v1/roles/${id}`, data);
  },

  deleteRole: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/v1/roles/${id}`);
  },

  // Role Permissions
  getRolePermissions: async (roleId: string): Promise<Permission[]> => {
    const response = await apiClient.get(`/api/v1/roles/${roleId}/permissions`);
    return response.data;
  },

  updateRolePermissions: async (roleId: string, permissionIds: string[]): Promise<void> => {
    await apiClient.put(`/api/v1/roles/${roleId}/permissions`, {
      permission_ids: permissionIds,
    });
  },

  // User Roles
  getUserRoles: async (userId: string): Promise<Role[]> => {
    const response = await apiClient.get(`/api/v1/users/${userId}/roles`);
    return response.data;
  },

  assignUserRoles: async (userId: string, roleIds: string[]): Promise<void> => {
    await apiClient.put(`/api/v1/users/${userId}/roles`, {
      role_ids: roleIds,
    });
  },

  // Permissions
  listPermissions: async (): Promise<Permission[]> => {
    const response = await apiClient.get('/api/v1/permissions');
    return response.data;
  },

  getUserPermissions: async (userId: string): Promise<Permission[]> => {
    const response = await apiClient.get(`/api/v1/users/${userId}/permissions`);
    return response.data;
  },

  getCurrentUserPermissions: async (): Promise<Permission[]> => {
    const response = await apiClient.get('/api/v1/me/permissions');
    return response.data;
  },

  // Audit Logs
  getAuditLogs: async (limit = 50, offset = 0): Promise<AuditLog[]> => {
    const response = await apiClient.get('/api/v1/audit-logs', {
      params: { limit, offset },
    });
    return response.data;
  },
};
