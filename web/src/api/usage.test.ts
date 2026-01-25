import { describe, it, expect, beforeEach, vi } from 'vitest'
import { usageApi } from './usage'
import type { UsageResponse, UsageHistoryResponse } from './usage'

// Mock the API client
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
  },
}))

import { apiClient } from './client'

describe('Usage API', () => {
  const mockUsageResponse: UsageResponse = {
    tenant_id: 'tenant-123',
    current_period: {
      workflow_executions: 100,
      step_executions: 500,
      period: '2024-01',
    },
    month_to_date: {
      workflow_executions: 1500,
      step_executions: 7500,
      period: '2024-01',
    },
    quotas: {
      max_executions_per_day: 500,
      max_executions_per_month: 10000,
      executions_remaining: 8500,
      quota_percent_used: 15,
      max_concurrent_executions: 10,
      max_workflows: 100,
    },
    rate_limits: {
      requests_per_minute: 60,
      requests_per_hour: 1000,
      requests_per_day: 10000,
      hits_today: 250,
    },
  }

  const mockUsageHistoryResponse: UsageHistoryResponse = {
    usage: [
      { date: '2024-01-15', workflow_executions: 100, step_executions: 500 },
      { date: '2024-01-14', workflow_executions: 120, step_executions: 600 },
      { date: '2024-01-13', workflow_executions: 90, step_executions: 450 },
    ],
    total: 30,
    page: 1,
    limit: 30,
    start_date: '2023-12-16',
    end_date: '2024-01-15',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getCurrentUsage', () => {
    it('should fetch current usage for tenant', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockUsageResponse)

      const result = await usageApi.getCurrentUsage('tenant-123')

      expect(apiClient.get).toHaveBeenCalledWith('/api/tenants/tenant-123/usage')
      expect(result).toEqual(mockUsageResponse)
    })

    it('should return correct quota information', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockUsageResponse)

      const result = await usageApi.getCurrentUsage('tenant-123')

      expect(result.quotas.max_executions_per_month).toBe(10000)
      expect(result.quotas.executions_remaining).toBe(8500)
      expect(result.quotas.quota_percent_used).toBe(15)
    })

    it('should return correct rate limit information', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockUsageResponse)

      const result = await usageApi.getCurrentUsage('tenant-123')

      expect(result.rate_limits.requests_per_minute).toBe(60)
      expect(result.rate_limits.hits_today).toBe(250)
    })

    it('should handle API error', async () => {
      const error = new Error('Tenant not found')
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(usageApi.getCurrentUsage('invalid-tenant')).rejects.toThrow('Tenant not found')
    })
  })

  describe('getUsageHistory', () => {
    it('should fetch usage history with default params', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockUsageHistoryResponse)

      const result = await usageApi.getUsageHistory('tenant-123')

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/tenants/tenant-123/usage/history?page=1&limit=30'
      )
      expect(result).toEqual(mockUsageHistoryResponse)
    })

    it('should fetch usage history with date range', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockUsageHistoryResponse)

      await usageApi.getUsageHistory('tenant-123', '2024-01-01', '2024-01-15')

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/tenants/tenant-123/usage/history?page=1&limit=30&start_date=2024-01-01&end_date=2024-01-15'
      )
    })

    it('should fetch usage history with custom pagination', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockUsageHistoryResponse)

      await usageApi.getUsageHistory('tenant-123', undefined, undefined, 2, 50)

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/tenants/tenant-123/usage/history?page=2&limit=50'
      )
    })

    it('should fetch usage history with all params', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockUsageHistoryResponse)

      await usageApi.getUsageHistory('tenant-123', '2024-01-01', '2024-01-31', 3, 100)

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/tenants/tenant-123/usage/history?page=3&limit=100&start_date=2024-01-01&end_date=2024-01-31'
      )
    })

    it('should handle empty history', async () => {
      const emptyResponse: UsageHistoryResponse = {
        usage: [],
        total: 0,
        page: 1,
        limit: 30,
        start_date: '2024-01-01',
        end_date: '2024-01-31',
      }
      ;(apiClient.get as any).mockResolvedValueOnce(emptyResponse)

      const result = await usageApi.getUsageHistory('tenant-123')

      expect(result.usage).toEqual([])
      expect(result.total).toBe(0)
    })

    it('should return correct pagination info', async () => {
      const paginatedResponse: UsageHistoryResponse = {
        ...mockUsageHistoryResponse,
        page: 2,
        limit: 10,
        total: 50,
      }
      ;(apiClient.get as any).mockResolvedValueOnce(paginatedResponse)

      const result = await usageApi.getUsageHistory('tenant-123', undefined, undefined, 2, 10)

      expect(result.page).toBe(2)
      expect(result.limit).toBe(10)
      expect(result.total).toBe(50)
    })
  })
})
