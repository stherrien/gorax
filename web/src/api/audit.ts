import { apiClient } from './client'
import type {
  AuditEvent,
  QueryFilter,
  QueryResponse,
  AuditStats,
  TimeRange,
  ExportFormat,
} from '../types/audit'

/**
 * Audit log API client
 */
export const auditAPI = {
  /**
   * Query audit events with filters and pagination
   */
  async queryAuditEvents(filter: QueryFilter): Promise<QueryResponse> {
    const params: Record<string, any> = {}

    if (filter.userId) params.user_id = filter.userId
    if (filter.userEmail) params.user_email = filter.userEmail
    if (filter.categories?.length) params.categories = filter.categories.join(',')
    if (filter.eventTypes?.length) params.event_types = filter.eventTypes.join(',')
    if (filter.actions?.length) params.actions = filter.actions.join(',')
    if (filter.resourceType) params.resource_type = filter.resourceType
    if (filter.resourceId) params.resource_id = filter.resourceId
    if (filter.ipAddress) params.ip_address = filter.ipAddress
    if (filter.severities?.length) params.severities = filter.severities.join(',')
    if (filter.statuses?.length) params.statuses = filter.statuses.join(',')
    if (filter.startDate) params.start_date = filter.startDate
    if (filter.endDate) params.end_date = filter.endDate
    if (filter.limit) params.limit = filter.limit
    if (filter.offset) params.offset = filter.offset
    if (filter.sortBy) params.sort_by = filter.sortBy
    if (filter.sortDirection) params.sort_direction = filter.sortDirection

    const response = await apiClient.get('/api/v1/admin/audit/events', { params })

    return {
      events: response.events || [],
      total: response.total || 0,
      page: Math.floor((filter.offset || 0) / (filter.limit || 50)) + 1,
      limit: filter.limit || 50,
    }
  },

  /**
   * Get a single audit event by ID
   */
  async getAuditEvent(eventId: string): Promise<AuditEvent> {
    return apiClient.get(`/api/v1/admin/audit/events/${eventId}`)
  },

  /**
   * Get audit statistics for a time range
   */
  async getAuditStats(timeRange: TimeRange): Promise<AuditStats> {
    return apiClient.get('/api/v1/admin/audit/stats', {
      params: {
        start_date: timeRange.startDate,
        end_date: timeRange.endDate,
      },
    })
  },

  /**
   * Export audit events to CSV or JSON
   */
  async exportAuditEvents(
    filter: QueryFilter,
    format: ExportFormat
  ): Promise<Blob> {
    const params: Record<string, any> = {
      format,
    }

    if (filter.userId) params.user_id = filter.userId
    if (filter.userEmail) params.user_email = filter.userEmail
    if (filter.categories?.length) params.categories = filter.categories.join(',')
    if (filter.eventTypes?.length) params.event_types = filter.eventTypes.join(',')
    if (filter.actions?.length) params.actions = filter.actions.join(',')
    if (filter.resourceType) params.resource_type = filter.resourceType
    if (filter.resourceId) params.resource_id = filter.resourceId
    if (filter.ipAddress) params.ip_address = filter.ipAddress
    if (filter.severities?.length) params.severities = filter.severities.join(',')
    if (filter.statuses?.length) params.statuses = filter.statuses.join(',')
    if (filter.startDate) params.start_date = filter.startDate
    if (filter.endDate) params.end_date = filter.endDate

    // Build URL with query parameters
    const url = new URL('/api/v1/admin/audit/export', window.location.origin)
    Object.entries(params).forEach(([key, value]) => {
      url.searchParams.append(key, String(value))
    })

    // Fetch with auth headers
    const token = localStorage.getItem('auth_token')
    const headers: Record<string, string> = {}
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    } else {
      headers['X-Tenant-ID'] = '00000000-0000-0000-0000-000000000001'
    }

    const response = await fetch(url.toString(), { headers })

    if (!response.ok) {
      throw new Error(`Export failed: ${response.statusText}`)
    }

    return response.blob()
  },

  /**
   * Get available audit categories
   */
  getCategories(): string[] {
    return [
      'authentication',
      'authorization',
      'data_access',
      'configuration',
      'workflow',
      'integration',
      'credential',
      'user_management',
      'system',
    ]
  },

  /**
   * Get available audit event types
   */
  getEventTypes(): string[] {
    return [
      'create',
      'read',
      'update',
      'delete',
      'execute',
      'login',
      'logout',
      'permission_change',
      'export',
      'import',
      'access',
      'configure',
    ]
  },

  /**
   * Get available severity levels
   */
  getSeverities(): string[] {
    return ['info', 'warning', 'error', 'critical']
  },

  /**
   * Get available status values
   */
  getStatuses(): string[] {
    return ['success', 'failure', 'partial']
  },
}
