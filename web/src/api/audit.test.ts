import { describe, it, expect, vi, beforeEach } from 'vitest'
import { auditAPI } from './audit'
import { apiClient } from './client'
import {
  AuditCategory,
  AuditEventType,
  AuditSeverity,
  AuditStatus,
  ExportFormat,
} from '../types/audit'

import { createQueryWrapper } from "../test/test-utils"
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
  },
}))

describe('auditAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('queryAuditEvents', () => {
    it('should query audit events with filters', async () => {
      const mockResponse = {
        events: [
          {
            id: '1',
            tenantId: 'tenant1',
            userId: 'user1',
            userEmail: 'user@example.com',
            category: AuditCategory.Authentication,
            eventType: AuditEventType.Login,
            action: 'user.login',
            resourceType: 'user',
            resourceId: 'user1',
            resourceName: 'User 1',
            ipAddress: '192.168.1.1',
            userAgent: 'Mozilla/5.0',
            severity: AuditSeverity.Info,
            status: AuditStatus.Success,
            metadata: {},
            createdAt: '2024-01-01T00:00:00Z',
          },
        ],
        total: 1,
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      const result = await auditAPI.queryAuditEvents({
        userId: 'user1',
        categories: [AuditCategory.Authentication],
        limit: 50,
        offset: 10,
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/admin/audit/events', {
        params: expect.objectContaining({
          user_id: 'user1',
          categories: 'authentication',
          limit: 50,
          offset: 10,
        }),
      })

      expect(result.events).toHaveLength(1)
      expect(result.total).toBe(1)
      expect(result.page).toBe(1)
      expect(result.limit).toBe(50)
    })

    it('should handle empty filter', async () => {
      const mockResponse = { events: [], total: 0 }
      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      const result = await auditAPI.queryAuditEvents({})

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/admin/audit/events', {
        params: {},
      })

      expect(result.events).toHaveLength(0)
      expect(result.total).toBe(0)
    })

    it('should handle multiple filter values', async () => {
      const mockResponse = { events: [], total: 0 }
      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      await auditAPI.queryAuditEvents({
        categories: [AuditCategory.Authentication, AuditCategory.Authorization],
        eventTypes: [AuditEventType.Login, AuditEventType.Logout],
        severities: [AuditSeverity.Warning, AuditSeverity.Error],
        statuses: [AuditStatus.Success, AuditStatus.Failure],
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/admin/audit/events', {
        params: {
          categories: 'authentication,authorization',
          event_types: 'login,logout',
          severities: 'warning,error',
          statuses: 'success,failure',
        },
      })
    })

    it('should calculate correct page number', async () => {
      const mockResponse = { events: [], total: 100 }
      vi.mocked(apiClient.get).mockResolvedValue(mockResponse)

      const result = await auditAPI.queryAuditEvents({
        limit: 25,
        offset: 50,
      })

      expect(result.page).toBe(3) // offset 50 / limit 25 + 1 = 3
    })
  })

  describe('getAuditEvent', () => {
    it('should get a single audit event by ID', async () => {
      const mockEvent = {
        id: 'event1',
        tenantId: 'tenant1',
        userId: 'user1',
        userEmail: 'user@example.com',
        category: AuditCategory.Workflow,
        eventType: AuditEventType.Execute,
        action: 'workflow.execute',
        resourceType: 'workflow',
        resourceId: 'wf1',
        resourceName: 'Test Workflow',
        ipAddress: '192.168.1.1',
        userAgent: 'Mozilla/5.0',
        severity: AuditSeverity.Info,
        status: AuditStatus.Success,
        metadata: {},
        createdAt: '2024-01-01T00:00:00Z',
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockEvent)

      const result = await auditAPI.getAuditEvent('event1')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/admin/audit/events/event1')
      expect(result).toEqual(mockEvent)
    })
  })

  describe('getAuditStats', () => {
    it('should get audit statistics for a time range', async () => {
      const mockStats = {
        totalEvents: 100,
        eventsByCategory: {
          [AuditCategory.Authentication]: 30,
          [AuditCategory.Workflow]: 40,
        },
        eventsBySeverity: {
          [AuditSeverity.Info]: 80,
          [AuditSeverity.Warning]: 15,
        },
        eventsByStatus: {
          [AuditStatus.Success]: 90,
          [AuditStatus.Failure]: 10,
        },
        topUsers: [],
        topActions: [],
        criticalEvents: 5,
        failedEvents: 10,
        recentCritical: [],
        timeRange: {
          startDate: '2024-01-01T00:00:00Z',
          endDate: '2024-01-31T23:59:59Z',
        },
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockStats)

      const result = await auditAPI.getAuditStats({
        startDate: '2024-01-01T00:00:00Z',
        endDate: '2024-01-31T23:59:59Z',
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/admin/audit/stats', {
        params: {
          start_date: '2024-01-01T00:00:00Z',
          end_date: '2024-01-31T23:59:59Z',
        },
      })

      expect(result.totalEvents).toBe(100)
      expect(result.criticalEvents).toBe(5)
      expect(result.failedEvents).toBe(10)
    })
  })

  describe('exportAuditEvents', () => {
    it('should export audit events as CSV', async () => {
      const mockBlob = new Blob(['csv,data'], { type: 'text/csv' })

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        blob: () => Promise.resolve(mockBlob),
      })

      const result = await auditAPI.exportAuditEvents(
        {
          startDate: '2024-01-01T00:00:00Z',
          endDate: '2024-01-31T23:59:59Z',
        },
        ExportFormat.CSV
      )

      expect(global.fetch).toHaveBeenCalled()
      expect(result).toBeInstanceOf(Blob)
    })

    it('should throw error on failed export', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        statusText: 'Internal Server Error',
      })

      await expect(
        auditAPI.exportAuditEvents({}, ExportFormat.JSON)
      ).rejects.toThrow('Export failed: Internal Server Error')
    })
  })

  describe('static data methods', () => {
    it('should return available categories', () => {
      const categories = auditAPI.getCategories()
      expect(categories).toContain('authentication')
      expect(categories).toContain('workflow')
      expect(categories).toContain('credential')
    })

    it('should return available event types', () => {
      const eventTypes = auditAPI.getEventTypes()
      expect(eventTypes).toContain('create')
      expect(eventTypes).toContain('login')
      expect(eventTypes).toContain('execute')
    })

    it('should return available severities', () => {
      const severities = auditAPI.getSeverities()
      expect(severities).toEqual(['info', 'warning', 'error', 'critical'])
    })

    it('should return available statuses', () => {
      const statuses = auditAPI.getStatuses()
      expect(statuses).toEqual(['success', 'failure', 'partial'])
    })
  })
})
