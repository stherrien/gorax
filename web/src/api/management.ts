import { apiClient } from './client'
import type {
  User,
  UserCreateInput,
  UserUpdateInput,
  UserListParams,
  UserListResponse,
  Tenant,
  TenantCreateInput,
  TenantUpdateInput,
  TenantListParams,
  TenantListResponse,
  ExecutionLog,
  ExecutionLogListParams,
  ExecutionLogListResponse,
  SystemHealth,
  MonitoringStats,
  WorkflowStats,
  BulkOperationResult,
} from '../types/management'

// ============================================================================
// User API
// ============================================================================

export const userAPI = {
  /**
   * List users with optional filtering and pagination
   */
  async list(params?: UserListParams): Promise<UserListResponse> {
    const response = await apiClient.get('/api/v1/users', { params: params || {} })
    return {
      users: response.data || response.users || [],
      total: response.total || (response.data?.length ?? 0),
    }
  },

  /**
   * Get single user by ID
   */
  async get(id: string): Promise<User> {
    const response = await apiClient.get(`/api/v1/users/${id}`)
    return response.data || response
  },

  /**
   * Create a new user
   */
  async create(input: UserCreateInput): Promise<User> {
    const response = await apiClient.post('/api/v1/users', input)
    return response.data || response
  },

  /**
   * Update an existing user
   */
  async update(id: string, updates: UserUpdateInput): Promise<User> {
    const response = await apiClient.put(`/api/v1/users/${id}`, updates)
    return response.data || response
  },

  /**
   * Delete a user
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/api/v1/users/${id}`)
  },

  /**
   * Resend invitation to a pending user
   */
  async resendInvite(id: string): Promise<void> {
    await apiClient.post(`/api/v1/users/${id}/resend-invite`, {})
  },

  /**
   * Bulk delete users
   */
  async bulkDelete(ids: string[]): Promise<BulkOperationResult> {
    const response = await apiClient.post('/api/v1/users/bulk/delete', { ids })
    return response
  },

  /**
   * Bulk update user roles
   */
  async bulkUpdateRole(ids: string[], role: string): Promise<BulkOperationResult> {
    const response = await apiClient.post('/api/v1/users/bulk/role', { ids, role })
    return response
  },
}

// ============================================================================
// Tenant API
// ============================================================================

export const tenantAPI = {
  /**
   * List tenants with optional filtering and pagination
   */
  async list(params?: TenantListParams): Promise<TenantListResponse> {
    const response = await apiClient.get('/api/v1/tenants', { params: params || {} })
    return {
      tenants: response.data || response.tenants || [],
      total: response.total || (response.data?.length ?? 0),
    }
  },

  /**
   * Get single tenant by ID
   */
  async get(id: string): Promise<Tenant> {
    const response = await apiClient.get(`/api/v1/tenants/${id}`)
    return response.data || response
  },

  /**
   * Create a new tenant
   */
  async create(input: TenantCreateInput): Promise<Tenant> {
    const response = await apiClient.post('/api/v1/tenants', input)
    return response.data || response
  },

  /**
   * Update an existing tenant
   */
  async update(id: string, updates: TenantUpdateInput): Promise<Tenant> {
    const response = await apiClient.put(`/api/v1/tenants/${id}`, updates)
    return response.data || response
  },

  /**
   * Delete a tenant
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/api/v1/tenants/${id}`)
  },

  /**
   * Get tenant usage statistics
   */
  async getUsage(id: string): Promise<Tenant['usage']> {
    const response = await apiClient.get(`/api/v1/tenants/${id}/usage`)
    return response.data || response
  },

  /**
   * Suspend a tenant
   */
  async suspend(id: string, reason?: string): Promise<Tenant> {
    const response = await apiClient.post(`/api/v1/tenants/${id}/suspend`, { reason })
    return response.data || response
  },

  /**
   * Reactivate a suspended tenant
   */
  async reactivate(id: string): Promise<Tenant> {
    const response = await apiClient.post(`/api/v1/tenants/${id}/reactivate`, {})
    return response.data || response
  },
}

// ============================================================================
// Execution Logs API
// ============================================================================

export const executionLogsAPI = {
  /**
   * List execution logs with filtering
   */
  async list(params: ExecutionLogListParams): Promise<ExecutionLogListResponse> {
    const { executionId, ...otherParams } = params
    const response = await apiClient.get(`/api/v1/executions/${executionId}/logs`, {
      params: otherParams,
    })
    return {
      logs: response.data || response.logs || [],
      total: response.total || (response.data?.length ?? 0),
    }
  },

  /**
   * Get single log entry
   */
  async get(executionId: string, logId: string): Promise<ExecutionLog> {
    const response = await apiClient.get(`/api/v1/executions/${executionId}/logs/${logId}`)
    return response.data || response
  },

  /**
   * Search logs across executions
   */
  async search(query: string, params?: { limit?: number; offset?: number }): Promise<ExecutionLogListResponse> {
    const response = await apiClient.get('/api/v1/logs/search', {
      params: { q: query, ...params },
    })
    return {
      logs: response.data || response.logs || [],
      total: response.total || (response.data?.length ?? 0),
    }
  },

  /**
   * Export logs to file
   */
  async export(
    executionId: string,
    format: 'txt' | 'json' | 'csv'
  ): Promise<void> {
    const baseURL = import.meta.env.VITE_API_URL || ''
    const token = localStorage.getItem('auth_token')
    const headers: HeadersInit = {
      'X-Tenant-ID': '00000000-0000-0000-0000-000000000001',
    }
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    const response = await fetch(
      `${baseURL}/api/v1/executions/${executionId}/logs/export?format=${format}`,
      { headers }
    )

    if (!response.ok) {
      throw new Error(`Failed to export logs: ${response.statusText}`)
    }

    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `execution-${executionId}-logs.${format}`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  },
}

// ============================================================================
// Monitoring API
// ============================================================================

export const monitoringAPI = {
  /**
   * Get current monitoring statistics
   */
  async getStats(): Promise<MonitoringStats> {
    const response = await apiClient.get('/api/v1/monitoring/stats')
    return response.data || response
  },

  /**
   * Get system health status
   */
  async getHealth(): Promise<SystemHealth> {
    const response = await apiClient.get('/api/v1/health')
    return response.data || response
  },

  /**
   * Get workflow-specific statistics
   */
  async getWorkflowStats(workflowId?: string): Promise<WorkflowStats[]> {
    const params = workflowId ? { workflowId } : undefined
    const response = await apiClient.get('/api/v1/monitoring/workflows', { params })
    return response.data || response.workflows || []
  },

  /**
   * Get active executions
   */
  async getActiveExecutions(): Promise<Array<{
    id: string
    workflowId: string
    workflowName: string
    status: string
    startedAt: string
    currentNode?: string
    progress: number
  }>> {
    const response = await apiClient.get('/api/v1/monitoring/active')
    return response.data || response.executions || []
  },
}
