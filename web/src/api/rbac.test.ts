import { describe, it, expect, beforeEach, vi } from 'vitest'
import { rbacApi } from './rbac'
import type { Role, Permission, CreateRoleRequest, UpdateRoleRequest, AuditLog } from './rbac'

// Mock the API client
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

import { apiClient } from './client'

describe('RBAC API', () => {
  const mockPermission: Permission = {
    id: 'perm-123',
    resource: 'workflows',
    action: 'read',
    description: 'Read workflows',
    created_at: '2024-01-15T10:00:00Z',
  }

  const mockRole: Role = {
    id: 'role-123',
    tenant_id: 'tenant-1',
    name: 'Admin',
    description: 'Administrator role',
    is_system: false,
    created_at: '2024-01-15T10:00:00Z',
    updated_at: '2024-01-15T10:00:00Z',
    permissions: [mockPermission],
  }

  const mockAuditLog: AuditLog = {
    id: 'audit-123',
    tenant_id: 'tenant-1',
    user_id: 'user-1',
    action: 'role.created',
    target_type: 'role',
    target_id: 'role-123',
    details: { name: 'Admin' },
    created_at: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Roles', () => {
    describe('listRoles', () => {
      it('should fetch list of roles', async () => {
        const mockRoles = [mockRole]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: mockRoles })

        const result = await rbacApi.listRoles()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/roles')
        expect(result).toEqual(mockRoles)
      })

      it('should handle empty list', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [] })

        const result = await rbacApi.listRoles()

        expect(result).toEqual([])
      })

      it('should handle API error', async () => {
        const error = new Error('Network error')
        ;(apiClient.get as any).mockRejectedValueOnce(error)

        await expect(rbacApi.listRoles()).rejects.toThrow('Network error')
      })
    })

    describe('createRole', () => {
      it('should create new role', async () => {
        const createData: CreateRoleRequest = {
          name: 'Editor',
          description: 'Editor role',
          permission_ids: ['perm-1', 'perm-2'],
        }
        const createdRole = { ...mockRole, name: 'Editor', id: 'role-456' }
        ;(apiClient.post as any).mockResolvedValueOnce({ data: createdRole })

        const result = await rbacApi.createRole(createData)

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/roles', createData)
        expect(result).toEqual(createdRole)
      })

      it('should create role without permissions', async () => {
        const createData: CreateRoleRequest = {
          name: 'Viewer',
          description: 'Viewer role',
        }
        ;(apiClient.post as any).mockResolvedValueOnce({ data: { ...mockRole, name: 'Viewer' } })

        const result = await rbacApi.createRole(createData)

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/roles', createData)
        expect(result.name).toBe('Viewer')
      })

      it('should handle validation error', async () => {
        const error = new Error('Name is required')
        error.name = 'ValidationError'
        ;(apiClient.post as any).mockRejectedValueOnce(error)

        await expect(
          rbacApi.createRole({ name: '', description: '' })
        ).rejects.toThrow('Name is required')
      })
    })

    describe('getRole', () => {
      it('should fetch single role by ID', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: mockRole })

        const result = await rbacApi.getRole('role-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/roles/role-123')
        expect(result).toEqual(mockRole)
      })

      it('should throw NotFoundError for invalid ID', async () => {
        const error = new Error('Role not found')
        error.name = 'NotFoundError'
        ;(apiClient.get as any).mockRejectedValueOnce(error)

        await expect(rbacApi.getRole('invalid-id')).rejects.toThrow('Role not found')
      })
    })

    describe('updateRole', () => {
      it('should update existing role', async () => {
        const updates: UpdateRoleRequest = {
          name: 'Updated Admin',
          description: 'Updated description',
        }
        ;(apiClient.put as any).mockResolvedValueOnce({})

        await rbacApi.updateRole('role-123', updates)

        expect(apiClient.put).toHaveBeenCalledWith('/api/v1/roles/role-123', updates)
      })

      it('should update role with permission IDs', async () => {
        const updates: UpdateRoleRequest = {
          permission_ids: ['perm-1', 'perm-2', 'perm-3'],
        }
        ;(apiClient.put as any).mockResolvedValueOnce({})

        await rbacApi.updateRole('role-123', updates)

        expect(apiClient.put).toHaveBeenCalledWith('/api/v1/roles/role-123', updates)
      })

      it('should throw error for system role modification', async () => {
        const error = new Error('Cannot modify system role')
        ;(apiClient.put as any).mockRejectedValueOnce(error)

        await expect(
          rbacApi.updateRole('system-role', { name: 'Test' })
        ).rejects.toThrow('Cannot modify system role')
      })
    })

    describe('deleteRole', () => {
      it('should delete role by ID', async () => {
        (apiClient.delete as any).mockResolvedValueOnce({})

        await rbacApi.deleteRole('role-123')

        expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/roles/role-123')
      })

      it('should throw NotFoundError for non-existent role', async () => {
        const error = new Error('Role not found')
        error.name = 'NotFoundError'
        ;(apiClient.delete as any).mockRejectedValueOnce(error)

        await expect(rbacApi.deleteRole('invalid-id')).rejects.toThrow('Role not found')
      })

      it('should throw error when role is in use', async () => {
        const error = new Error('Role is assigned to users')
        ;(apiClient.delete as any).mockRejectedValueOnce(error)

        await expect(rbacApi.deleteRole('role-123')).rejects.toThrow('Role is assigned to users')
      })
    })
  })

  describe('Role Permissions', () => {
    describe('getRolePermissions', () => {
      it('should fetch permissions for a role', async () => {
        const permissions = [mockPermission]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: permissions })

        const result = await rbacApi.getRolePermissions('role-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/roles/role-123/permissions')
        expect(result).toEqual(permissions)
      })

      it('should handle role with no permissions', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [] })

        const result = await rbacApi.getRolePermissions('role-123')

        expect(result).toEqual([])
      })
    })

    describe('updateRolePermissions', () => {
      it('should update role permissions', async () => {
        const permissionIds = ['perm-1', 'perm-2', 'perm-3']
        ;(apiClient.put as any).mockResolvedValueOnce({})

        await rbacApi.updateRolePermissions('role-123', permissionIds)

        expect(apiClient.put).toHaveBeenCalledWith('/api/v1/roles/role-123/permissions', {
          permission_ids: permissionIds,
        })
      })

      it('should clear all permissions with empty array', async () => {
        (apiClient.put as any).mockResolvedValueOnce({})

        await rbacApi.updateRolePermissions('role-123', [])

        expect(apiClient.put).toHaveBeenCalledWith('/api/v1/roles/role-123/permissions', {
          permission_ids: [],
        })
      })
    })
  })

  describe('User Roles', () => {
    describe('getUserRoles', () => {
      it('should fetch roles for a user', async () => {
        const roles = [mockRole]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: roles })

        const result = await rbacApi.getUserRoles('user-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/users/user-123/roles')
        expect(result).toEqual(roles)
      })

      it('should handle user with no roles', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [] })

        const result = await rbacApi.getUserRoles('user-123')

        expect(result).toEqual([])
      })
    })

    describe('assignUserRoles', () => {
      it('should assign roles to user', async () => {
        const roleIds = ['role-1', 'role-2']
        ;(apiClient.put as any).mockResolvedValueOnce({})

        await rbacApi.assignUserRoles('user-123', roleIds)

        expect(apiClient.put).toHaveBeenCalledWith('/api/v1/users/user-123/roles', {
          role_ids: roleIds,
        })
      })

      it('should remove all roles with empty array', async () => {
        (apiClient.put as any).mockResolvedValueOnce({})

        await rbacApi.assignUserRoles('user-123', [])

        expect(apiClient.put).toHaveBeenCalledWith('/api/v1/users/user-123/roles', {
          role_ids: [],
        })
      })
    })
  })

  describe('Permissions', () => {
    describe('listPermissions', () => {
      it('should fetch all permissions', async () => {
        const permissions = [
          mockPermission,
          { ...mockPermission, id: 'perm-456', action: 'write' },
        ]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: permissions })

        const result = await rbacApi.listPermissions()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/permissions')
        expect(result).toEqual(permissions)
        expect(result).toHaveLength(2)
      })

      it('should handle empty permissions list', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [] })

        const result = await rbacApi.listPermissions()

        expect(result).toEqual([])
      })
    })

    describe('getUserPermissions', () => {
      it('should fetch permissions for specific user', async () => {
        const permissions = [mockPermission]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: permissions })

        const result = await rbacApi.getUserPermissions('user-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/users/user-123/permissions')
        expect(result).toEqual(permissions)
      })

      it('should return empty array for user with no permissions', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [] })

        const result = await rbacApi.getUserPermissions('user-123')

        expect(result).toEqual([])
      })
    })

    describe('getCurrentUserPermissions', () => {
      it('should fetch permissions for current user', async () => {
        const permissions = [mockPermission]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: permissions })

        const result = await rbacApi.getCurrentUserPermissions()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/me/permissions')
        expect(result).toEqual(permissions)
      })

      it('should handle unauthenticated user', async () => {
        const error = new Error('Unauthorized')
        error.name = 'UnauthorizedError'
        ;(apiClient.get as any).mockRejectedValueOnce(error)

        await expect(rbacApi.getCurrentUserPermissions()).rejects.toThrow('Unauthorized')
      })
    })
  })

  describe('Audit Logs', () => {
    describe('getAuditLogs', () => {
      it('should fetch audit logs with default pagination', async () => {
        const auditLogs = [mockAuditLog]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: auditLogs })

        const result = await rbacApi.getAuditLogs()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/audit-logs', {
          params: { limit: 50, offset: 0 },
        })
        expect(result).toEqual(auditLogs)
      })

      it('should fetch audit logs with custom limit', async () => {
        const auditLogs = [mockAuditLog]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: auditLogs })

        const result = await rbacApi.getAuditLogs(100)

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/audit-logs', {
          params: { limit: 100, offset: 0 },
        })
        expect(result).toEqual(auditLogs)
      })

      it('should fetch audit logs with custom offset', async () => {
        const auditLogs = [mockAuditLog]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: auditLogs })

        const result = await rbacApi.getAuditLogs(50, 100)

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/audit-logs', {
          params: { limit: 50, offset: 100 },
        })
        expect(result).toEqual(auditLogs)
      })

      it('should handle empty audit logs', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [] })

        const result = await rbacApi.getAuditLogs()

        expect(result).toEqual([])
      })
    })
  })
})
