import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rbacApi, Role, CreateRoleRequest, UpdateRoleRequest, Permission } from '../api/rbac';
import { toast } from 'react-hot-toast';

export const useRoles = () => {
  return useQuery({
    queryKey: ['roles'],
    queryFn: rbacApi.listRoles,
  });
};

export const useRole = (roleId: string) => {
  return useQuery({
    queryKey: ['roles', roleId],
    queryFn: () => rbacApi.getRole(roleId),
    enabled: !!roleId,
  });
};

export const useCreateRole = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateRoleRequest) => rbacApi.createRole(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      toast.success('Role created successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to create role');
    },
  });
};

export const useUpdateRole = (roleId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateRoleRequest) => rbacApi.updateRole(roleId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      queryClient.invalidateQueries({ queryKey: ['roles', roleId] });
      toast.success('Role updated successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to update role');
    },
  });
};

export const useDeleteRole = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (roleId: string) => rbacApi.deleteRole(roleId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      toast.success('Role deleted successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to delete role');
    },
  });
};

export const useRolePermissions = (roleId: string) => {
  return useQuery({
    queryKey: ['roles', roleId, 'permissions'],
    queryFn: () => rbacApi.getRolePermissions(roleId),
    enabled: !!roleId,
  });
};

export const useUpdateRolePermissions = (roleId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (permissionIds: string[]) => rbacApi.updateRolePermissions(roleId, permissionIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles', roleId, 'permissions'] });
      queryClient.invalidateQueries({ queryKey: ['roles', roleId] });
      toast.success('Role permissions updated successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to update permissions');
    },
  });
};

export const useUserRoles = (userId: string) => {
  return useQuery({
    queryKey: ['users', userId, 'roles'],
    queryFn: () => rbacApi.getUserRoles(userId),
    enabled: !!userId,
  });
};

export const useAssignUserRoles = (userId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (roleIds: string[]) => rbacApi.assignUserRoles(userId, roleIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users', userId, 'roles'] });
      queryClient.invalidateQueries({ queryKey: ['users', userId, 'permissions'] });
      toast.success('User roles updated successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to assign roles');
    },
  });
};

export const usePermissions = () => {
  return useQuery({
    queryKey: ['permissions'],
    queryFn: rbacApi.listPermissions,
  });
};

export const useUserPermissions = (userId: string) => {
  return useQuery({
    queryKey: ['users', userId, 'permissions'],
    queryFn: () => rbacApi.getUserPermissions(userId),
    enabled: !!userId,
  });
};

export const useCurrentUserPermissions = () => {
  return useQuery({
    queryKey: ['me', 'permissions'],
    queryFn: rbacApi.getCurrentUserPermissions,
  });
};

export const useAuditLogs = (limit = 50, offset = 0) => {
  return useQuery({
    queryKey: ['audit-logs', limit, offset],
    queryFn: () => rbacApi.getAuditLogs(limit, offset),
  });
};

// Helper hook to check if current user has a permission
export const useHasPermission = (resource: string, action: string): boolean => {
  const { data: permissions = [] } = useCurrentUserPermissions();
  return permissions.some(
    (p: Permission) => p.resource === resource && p.action === action
  );
};

// Helper hook to check if current user has any of the specified permissions
export const useHasAnyPermission = (checks: Array<{ resource: string; action: string }>): boolean => {
  const { data: permissions = [] } = useCurrentUserPermissions();
  return checks.some(check =>
    permissions.some((p: Permission) => p.resource === check.resource && p.action === check.action)
  );
};

// Helper hook to check if current user has all of the specified permissions
export const useHasAllPermissions = (checks: Array<{ resource: string; action: string }>): boolean => {
  const { data: permissions = [] } = useCurrentUserPermissions();
  return checks.every(check =>
    permissions.some((p: Permission) => p.resource === check.resource && p.action === check.action)
  );
};
