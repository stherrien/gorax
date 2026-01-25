import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { userAPI, tenantAPI, executionLogsAPI, monitoringAPI } from './management'
import type {
  User,
  UserCreateInput,
  UserUpdateInput,
  Tenant,
  TenantCreateInput,
  TenantUpdateInput,
  ExecutionLog,
  MonitoringStats,
  SystemHealth,
  WorkflowStats,
  BulkOperationResult,
} from '../types/management'

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

describe('Management API', () => {
  const mockUser: User = {
    id: 'user-123',
    tenantId: 'tenant-1',
    email: 'user@example.com',
    name: 'Test User',
    role: 'operator',
    status: 'active',
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
  }

  const mockTenant: Tenant = {
    id: 'tenant-123',
    name: 'Acme Corp',
    slug: 'acme-corp',
    status: 'active',
    plan: 'professional',
    limits: {
      maxWorkflows: 100,
      maxExecutionsPerMonth: 10000,
      maxUsers: 50,
      maxCredentials: 100,
      retentionDays: 90,
    },
    usage: {
      workflowCount: 25,
      executionsThisMonth: 1500,
      userCount: 10,
      credentialCount: 15,
      storageBytes: 1048576,
    },
    ownerId: 'user-1',
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
  }

  const mockLog: ExecutionLog = {
    id: 'log-123',
    executionId: 'exec-456',
    nodeId: 'node-1',
    nodeName: 'HTTP Request',
    level: 'info',
    message: 'Request completed successfully',
    data: { statusCode: 200 },
    timestamp: '2024-01-15T10:00:00Z',
  }

  const mockMonitoringStats: MonitoringStats = {
    activeExecutions: 5,
    queuedExecutions: 10,
    executionsPerMinute: 2.5,
    averageExecutionTime: 1500,
    successRate: 98.5,
    errorRate: 1.5,
    lastHourExecutions: 150,
    lastHourFailures: 2,
  }

  const mockSystemHealth: SystemHealth = {
    overall: 'healthy',
    services: [
      { name: 'database', status: 'healthy', responseTime: 5, lastCheck: '2024-01-15T10:00:00Z' },
      { name: 'redis', status: 'healthy', responseTime: 2, lastCheck: '2024-01-15T10:00:00Z' },
    ],
    uptime: 86400,
    lastUpdated: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('User API', () => {
    describe('list', () => {
      it('should fetch list of users', async () => {
        const users = [mockUser]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: users, total: 1 })

        const result = await userAPI.list()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/users', { params: {} })
        expect(result.users).toEqual(users)
        expect(result.total).toBe(1)
      })

      it('should fetch users with pagination', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await userAPI.list({ page: 2, limit: 20 })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/users', {
          params: { page: 2, limit: 20 },
        })
      })

      it('should fetch users with role filter', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await userAPI.list({ role: 'admin' })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/users', {
          params: { role: 'admin' },
        })
      })

      it('should fetch users with status filter', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await userAPI.list({ status: 'active' })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/users', {
          params: { status: 'active' },
        })
      })

      it('should fetch users with search', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await userAPI.list({ search: 'john' })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/users', {
          params: { search: 'john' },
        })
      })

      it('should handle response with users property', async () => {
        const users = [mockUser]
        ;(apiClient.get as any).mockResolvedValueOnce({ users, total: 1 })

        const result = await userAPI.list()

        expect(result.users).toEqual(users)
      })
    })

    describe('get', () => {
      it('should fetch single user by ID', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: mockUser })

        const result = await userAPI.get('user-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/users/user-123')
        expect(result).toEqual(mockUser)
      })

      it('should handle direct response', async () => {
        (apiClient.get as any).mockResolvedValueOnce(mockUser)

        const result = await userAPI.get('user-123')

        expect(result).toEqual(mockUser)
      })

      it('should throw error for invalid ID', async () => {
        const error = new Error('User not found')
        ;(apiClient.get as any).mockRejectedValueOnce(error)

        await expect(userAPI.get('invalid-id')).rejects.toThrow('User not found')
      })
    })

    describe('create', () => {
      it('should create new user', async () => {
        const input: UserCreateInput = {
          email: 'new@example.com',
          name: 'New User',
          role: 'viewer',
          sendInvite: true,
        }
        ;(apiClient.post as any).mockResolvedValueOnce({ data: { ...mockUser, ...input } })

        const result = await userAPI.create(input)

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/users', input)
        expect(result.email).toBe('new@example.com')
      })

      it('should create user without invite', async () => {
        const input: UserCreateInput = {
          email: 'new@example.com',
          name: 'New User',
          role: 'viewer',
          sendInvite: false,
        }
        ;(apiClient.post as any).mockResolvedValueOnce({ data: { ...mockUser, ...input } })

        await userAPI.create(input)

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/users', input)
      })

      it('should handle validation error', async () => {
        const error = new Error('Email is required')
        ;(apiClient.post as any).mockRejectedValueOnce(error)

        await expect(
          userAPI.create({ email: '', name: 'Test', role: 'viewer' })
        ).rejects.toThrow('Email is required')
      })
    })

    describe('update', () => {
      it('should update user', async () => {
        const updates: UserUpdateInput = {
          name: 'Updated Name',
          role: 'admin',
        }
        ;(apiClient.put as any).mockResolvedValueOnce({ data: { ...mockUser, ...updates } })

        const result = await userAPI.update('user-123', updates)

        expect(apiClient.put).toHaveBeenCalledWith('/api/v1/users/user-123', updates)
        expect(result.name).toBe('Updated Name')
        expect(result.role).toBe('admin')
      })

      it('should update user status', async () => {
        const updates: UserUpdateInput = { status: 'suspended' }
        ;(apiClient.put as any).mockResolvedValueOnce({ data: { ...mockUser, ...updates } })

        const result = await userAPI.update('user-123', updates)

        expect(result.status).toBe('suspended')
      })
    })

    describe('delete', () => {
      it('should delete user', async () => {
        (apiClient.delete as any).mockResolvedValueOnce({})

        await userAPI.delete('user-123')

        expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/users/user-123')
      })

      it('should throw error for non-existent user', async () => {
        const error = new Error('User not found')
        ;(apiClient.delete as any).mockRejectedValueOnce(error)

        await expect(userAPI.delete('invalid-id')).rejects.toThrow('User not found')
      })
    })

    describe('resendInvite', () => {
      it('should resend invite to user', async () => {
        (apiClient.post as any).mockResolvedValueOnce({})

        await userAPI.resendInvite('user-123')

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/users/user-123/resend-invite', {})
      })
    })

    describe('bulkDelete', () => {
      it('should bulk delete users', async () => {
        const result: BulkOperationResult = {
          successCount: 3,
          failureCount: 1,
          failures: [{ id: 'user-4', error: 'Cannot delete admin user' }],
        }
        ;(apiClient.post as any).mockResolvedValueOnce(result)

        const response = await userAPI.bulkDelete(['user-1', 'user-2', 'user-3', 'user-4'])

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/users/bulk/delete', {
          ids: ['user-1', 'user-2', 'user-3', 'user-4'],
        })
        expect(response.successCount).toBe(3)
        expect(response.failureCount).toBe(1)
      })
    })

    describe('bulkUpdateRole', () => {
      it('should bulk update user roles', async () => {
        const result: BulkOperationResult = {
          successCount: 2,
          failureCount: 0,
          failures: [],
        }
        ;(apiClient.post as any).mockResolvedValueOnce(result)

        const response = await userAPI.bulkUpdateRole(['user-1', 'user-2'], 'admin')

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/users/bulk/role', {
          ids: ['user-1', 'user-2'],
          role: 'admin',
        })
        expect(response.successCount).toBe(2)
      })
    })
  })

  describe('Tenant API', () => {
    describe('list', () => {
      it('should fetch list of tenants', async () => {
        const tenants = [mockTenant]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: tenants, total: 1 })

        const result = await tenantAPI.list()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/tenants', { params: {} })
        expect(result.tenants).toEqual(tenants)
        expect(result.total).toBe(1)
      })

      it('should fetch tenants with pagination', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await tenantAPI.list({ page: 1, limit: 10 })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/tenants', {
          params: { page: 1, limit: 10 },
        })
      })

      it('should fetch tenants with status filter', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await tenantAPI.list({ status: 'active' })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/tenants', {
          params: { status: 'active' },
        })
      })

      it('should fetch tenants with plan filter', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await tenantAPI.list({ plan: 'enterprise' })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/tenants', {
          params: { plan: 'enterprise' },
        })
      })

      it('should handle response with tenants property', async () => {
        const tenants = [mockTenant]
        ;(apiClient.get as any).mockResolvedValueOnce({ tenants, total: 1 })

        const result = await tenantAPI.list()

        expect(result.tenants).toEqual(tenants)
      })
    })

    describe('get', () => {
      it('should fetch single tenant by ID', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: mockTenant })

        const result = await tenantAPI.get('tenant-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/tenants/tenant-123')
        expect(result).toEqual(mockTenant)
      })

      it('should throw error for invalid ID', async () => {
        const error = new Error('Tenant not found')
        ;(apiClient.get as any).mockRejectedValueOnce(error)

        await expect(tenantAPI.get('invalid-id')).rejects.toThrow('Tenant not found')
      })
    })

    describe('create', () => {
      it('should create new tenant', async () => {
        const input: TenantCreateInput = {
          name: 'New Corp',
          slug: 'new-corp',
          plan: 'starter',
          ownerEmail: 'owner@newcorp.com',
        }
        ;(apiClient.post as any).mockResolvedValueOnce({ data: { ...mockTenant, ...input } })

        const result = await tenantAPI.create(input)

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/tenants', input)
        expect(result.name).toBe('New Corp')
      })

      it('should handle validation error', async () => {
        const error = new Error('Slug already exists')
        ;(apiClient.post as any).mockRejectedValueOnce(error)

        await expect(
          tenantAPI.create({ name: 'Test', slug: 'test', plan: 'free', ownerEmail: 'test@test.com' })
        ).rejects.toThrow('Slug already exists')
      })
    })

    describe('update', () => {
      it('should update tenant', async () => {
        const updates: TenantUpdateInput = {
          name: 'Updated Corp',
          plan: 'enterprise',
        }
        ;(apiClient.put as any).mockResolvedValueOnce({ data: { ...mockTenant, ...updates } })

        const result = await tenantAPI.update('tenant-123', updates)

        expect(apiClient.put).toHaveBeenCalledWith('/api/v1/tenants/tenant-123', updates)
        expect(result.name).toBe('Updated Corp')
        expect(result.plan).toBe('enterprise')
      })

      it('should update tenant limits', async () => {
        const updates: TenantUpdateInput = {
          limits: { maxWorkflows: 200, maxUsers: 100 },
        }
        ;(apiClient.put as any).mockResolvedValueOnce({
          data: { ...mockTenant, limits: { ...mockTenant.limits, ...updates.limits } },
        })

        const result = await tenantAPI.update('tenant-123', updates)

        expect(result.limits.maxWorkflows).toBe(200)
      })
    })

    describe('delete', () => {
      it('should delete tenant', async () => {
        (apiClient.delete as any).mockResolvedValueOnce({})

        await tenantAPI.delete('tenant-123')

        expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/tenants/tenant-123')
      })
    })

    describe('getUsage', () => {
      it('should fetch tenant usage', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: mockTenant.usage })

        const result = await tenantAPI.getUsage('tenant-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/tenants/tenant-123/usage')
        expect(result).toEqual(mockTenant.usage)
      })
    })

    describe('suspend', () => {
      it('should suspend tenant without reason', async () => {
        const suspendedTenant = { ...mockTenant, status: 'suspended' as const }
        ;(apiClient.post as any).mockResolvedValueOnce({ data: suspendedTenant })

        const result = await tenantAPI.suspend('tenant-123')

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/tenants/tenant-123/suspend', {
          reason: undefined,
        })
        expect(result.status).toBe('suspended')
      })

      it('should suspend tenant with reason', async () => {
        const suspendedTenant = { ...mockTenant, status: 'suspended' as const }
        ;(apiClient.post as any).mockResolvedValueOnce({ data: suspendedTenant })

        await tenantAPI.suspend('tenant-123', 'Payment overdue')

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/tenants/tenant-123/suspend', {
          reason: 'Payment overdue',
        })
      })
    })

    describe('reactivate', () => {
      it('should reactivate suspended tenant', async () => {
        const activeTenant = { ...mockTenant, status: 'active' as const }
        ;(apiClient.post as any).mockResolvedValueOnce({ data: activeTenant })

        const result = await tenantAPI.reactivate('tenant-123')

        expect(apiClient.post).toHaveBeenCalledWith('/api/v1/tenants/tenant-123/reactivate', {})
        expect(result.status).toBe('active')
      })
    })
  })

  describe('Execution Logs API', () => {
    describe('list', () => {
      it('should fetch logs for execution', async () => {
        const logs = [mockLog]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: logs, total: 1 })

        const result = await executionLogsAPI.list({ executionId: 'exec-456' })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/executions/exec-456/logs', {
          params: {},
        })
        expect(result.logs).toEqual(logs)
        expect(result.total).toBe(1)
      })

      it('should fetch logs with node filter', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await executionLogsAPI.list({ executionId: 'exec-456', nodeId: 'node-1' })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/executions/exec-456/logs', {
          params: { nodeId: 'node-1' },
        })
      })

      it('should fetch logs with level filter', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await executionLogsAPI.list({ executionId: 'exec-456', level: 'error' })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/executions/exec-456/logs', {
          params: { level: 'error' },
        })
      })

      it('should fetch logs with pagination', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await executionLogsAPI.list({ executionId: 'exec-456', limit: 100, offset: 50 })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/executions/exec-456/logs', {
          params: { limit: 100, offset: 50 },
        })
      })

      it('should handle response with logs property', async () => {
        const logs = [mockLog]
        ;(apiClient.get as any).mockResolvedValueOnce({ logs, total: 1 })

        const result = await executionLogsAPI.list({ executionId: 'exec-456' })

        expect(result.logs).toEqual(logs)
      })
    })

    describe('get', () => {
      it('should fetch single log entry', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: mockLog })

        const result = await executionLogsAPI.get('exec-456', 'log-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/executions/exec-456/logs/log-123')
        expect(result).toEqual(mockLog)
      })
    })

    describe('search', () => {
      it('should search logs with query', async () => {
        const logs = [mockLog]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: logs, total: 1 })

        const result = await executionLogsAPI.search('error')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/logs/search', {
          params: { q: 'error' },
        })
        expect(result.logs).toEqual(logs)
      })

      it('should search logs with pagination', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [], total: 0 })

        await executionLogsAPI.search('error', { limit: 50, offset: 100 })

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/logs/search', {
          params: { q: 'error', limit: 50, offset: 100 },
        })
      })
    })

    describe('export', () => {
      const originalFetch = global.fetch
      const originalCreateObjectURL = URL.createObjectURL
      const originalRevokeObjectURL = URL.revokeObjectURL

      beforeEach(() => {
        vi.stubGlobal('fetch', vi.fn())
        URL.createObjectURL = vi.fn().mockReturnValue('blob:url')
        URL.revokeObjectURL = vi.fn()
      })

      afterEach(() => {
        global.fetch = originalFetch
        URL.createObjectURL = originalCreateObjectURL
        URL.revokeObjectURL = originalRevokeObjectURL
      })

      it('should export logs as JSON', async () => {
        const mockBlob = new Blob(['[]'], { type: 'application/json' })
        ;(global.fetch as any).mockResolvedValueOnce({
          ok: true,
          blob: () => Promise.resolve(mockBlob),
        })

        const clickMock = vi.fn()
        const appendChildMock = vi.spyOn(document.body, 'appendChild').mockImplementation(() => null as any)
        const removeChildMock = vi.spyOn(document.body, 'removeChild').mockImplementation(() => null as any)

        vi.spyOn(document, 'createElement').mockReturnValue({
          href: '',
          download: '',
          click: clickMock,
        } as any)

        await executionLogsAPI.export('exec-456', 'json')

        expect(global.fetch).toHaveBeenCalledWith(
          '/api/v1/executions/exec-456/logs/export?format=json',
          expect.objectContaining({
            headers: expect.objectContaining({
              'X-Tenant-ID': '00000000-0000-0000-0000-000000000001',
            }),
          })
        )
        expect(clickMock).toHaveBeenCalled()

        appendChildMock.mockRestore()
        removeChildMock.mockRestore()
      })

      it('should throw error on export failure', async () => {
        (global.fetch as any).mockResolvedValueOnce({
          ok: false,
          statusText: 'Internal Server Error',
        })

        await expect(executionLogsAPI.export('exec-456', 'json')).rejects.toThrow(
          'Failed to export logs: Internal Server Error'
        )
      })
    })
  })

  describe('Monitoring API', () => {
    describe('getStats', () => {
      it('should fetch monitoring stats', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: mockMonitoringStats })

        const result = await monitoringAPI.getStats()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/monitoring/stats')
        expect(result).toEqual(mockMonitoringStats)
      })

      it('should handle direct response', async () => {
        (apiClient.get as any).mockResolvedValueOnce(mockMonitoringStats)

        const result = await monitoringAPI.getStats()

        expect(result).toEqual(mockMonitoringStats)
      })
    })

    describe('getHealth', () => {
      it('should fetch system health', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: mockSystemHealth })

        const result = await monitoringAPI.getHealth()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/health')
        expect(result).toEqual(mockSystemHealth)
      })

      it('should handle degraded health status', async () => {
        const degradedHealth: SystemHealth = {
          ...mockSystemHealth,
          overall: 'degraded',
          services: [
            { name: 'database', status: 'healthy', responseTime: 5, lastCheck: '2024-01-15T10:00:00Z' },
            { name: 'redis', status: 'degraded', responseTime: 500, lastCheck: '2024-01-15T10:00:00Z', message: 'High latency' },
          ],
        }
        ;(apiClient.get as any).mockResolvedValueOnce({ data: degradedHealth })

        const result = await monitoringAPI.getHealth()

        expect(result.overall).toBe('degraded')
        expect(result.services[1].status).toBe('degraded')
      })
    })

    describe('getWorkflowStats', () => {
      it('should fetch all workflow stats', async () => {
        const stats: WorkflowStats[] = [
          {
            workflowId: 'wf-1',
            workflowName: 'Workflow 1',
            totalExecutions: 100,
            successfulExecutions: 95,
            failedExecutions: 5,
            averageDuration: 1500,
            lastExecutedAt: '2024-01-15T10:00:00Z',
          },
        ]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: stats })

        const result = await monitoringAPI.getWorkflowStats()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/monitoring/workflows', {
          params: undefined,
        })
        expect(result).toEqual(stats)
      })

      it('should fetch stats for specific workflow', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [] })

        await monitoringAPI.getWorkflowStats('wf-123')

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/monitoring/workflows', {
          params: { workflowId: 'wf-123' },
        })
      })

      it('should handle response with workflows property', async () => {
        const stats: WorkflowStats[] = []
        ;(apiClient.get as any).mockResolvedValueOnce({ workflows: stats })

        const result = await monitoringAPI.getWorkflowStats()

        expect(result).toEqual(stats)
      })
    })

    describe('getActiveExecutions', () => {
      it('should fetch active executions', async () => {
        const activeExecutions = [
          {
            id: 'exec-1',
            workflowId: 'wf-1',
            workflowName: 'Test Workflow',
            status: 'running',
            startedAt: '2024-01-15T10:00:00Z',
            currentNode: 'HTTP Request',
            progress: 50,
          },
        ]
        ;(apiClient.get as any).mockResolvedValueOnce({ data: activeExecutions })

        const result = await monitoringAPI.getActiveExecutions()

        expect(apiClient.get).toHaveBeenCalledWith('/api/v1/monitoring/active')
        expect(result).toEqual(activeExecutions)
      })

      it('should handle response with executions property', async () => {
        const activeExecutions = [
          {
            id: 'exec-1',
            workflowId: 'wf-1',
            workflowName: 'Test Workflow',
            status: 'running',
            startedAt: '2024-01-15T10:00:00Z',
            progress: 75,
          },
        ]
        ;(apiClient.get as any).mockResolvedValueOnce({ executions: activeExecutions })

        const result = await monitoringAPI.getActiveExecutions()

        expect(result).toEqual(activeExecutions)
      })

      it('should handle empty active executions', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [] })

        const result = await monitoringAPI.getActiveExecutions()

        expect(result).toEqual([])
      })
    })
  })
})
