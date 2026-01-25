import { describe, it, expect, beforeEach, vi } from 'vitest'
import metricsApi from './metrics'
import type {
  ExecutionTrendsResponse,
  DurationStatsResponse,
  TopFailuresResponse,
  TriggerBreakdownResponse,
} from './metrics'

// Mock the API client
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
  },
}))

import { apiClient } from './client'

describe('Metrics API', () => {
  const mockTrendsResponse: ExecutionTrendsResponse = {
    trends: [
      { date: '2024-01-15', count: 100, success: 95, failed: 5 },
      { date: '2024-01-16', count: 120, success: 115, failed: 5 },
    ],
    startDate: '2024-01-15',
    endDate: '2024-01-16',
    groupBy: 'day',
  }

  const mockDurationResponse: DurationStatsResponse = {
    stats: [
      {
        workflowId: 'wf-1',
        workflowName: 'Data Sync',
        avgDuration: 1500,
        p50Duration: 1200,
        p90Duration: 2500,
        p99Duration: 5000,
        totalRuns: 100,
      },
    ],
    startDate: '2024-01-15',
    endDate: '2024-01-16',
  }

  const mockFailuresResponse: TopFailuresResponse = {
    failures: [
      {
        workflowId: 'wf-1',
        workflowName: 'Data Sync',
        failureCount: 10,
        lastFailedAt: '2024-01-16T10:00:00Z',
        errorPreview: 'Connection timeout',
      },
    ],
    startDate: '2024-01-15',
    endDate: '2024-01-16',
    limit: 10,
  }

  const mockTriggerBreakdown: TriggerBreakdownResponse = {
    breakdown: [
      { triggerType: 'webhook', count: 50, percentage: 50 },
      { triggerType: 'schedule', count: 30, percentage: 30 },
      { triggerType: 'manual', count: 20, percentage: 20 },
    ],
    startDate: '2024-01-15',
    endDate: '2024-01-16',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getExecutionTrends', () => {
    it('should fetch execution trends without params', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTrendsResponse)

      const result = await metricsApi.getExecutionTrends()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/metrics/trends')
      expect(result).toEqual(mockTrendsResponse)
    })

    it('should fetch execution trends with days param', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTrendsResponse)

      await metricsApi.getExecutionTrends({ days: 7 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/metrics/trends?days=7')
    })

    it('should fetch execution trends with date range', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTrendsResponse)

      await metricsApi.getExecutionTrends({
        startDate: '2024-01-01',
        endDate: '2024-01-31',
      })

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/v1/metrics/trends?startDate=2024-01-01&endDate=2024-01-31'
      )
    })

    it('should fetch execution trends with groupBy param', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTrendsResponse)

      await metricsApi.getExecutionTrends({ groupBy: 'hour' })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/metrics/trends?groupBy=hour')
    })

    it('should fetch execution trends with all params', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTrendsResponse)

      await metricsApi.getExecutionTrends({
        days: 30,
        startDate: '2024-01-01',
        endDate: '2024-01-31',
        groupBy: 'day',
      })

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/v1/metrics/trends?days=30&startDate=2024-01-01&endDate=2024-01-31&groupBy=day'
      )
    })

    it('should handle API error', async () => {
      const error = new Error('Network error')
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(metricsApi.getExecutionTrends()).rejects.toThrow('Network error')
    })
  })

  describe('getDurationStats', () => {
    it('should fetch duration stats without params', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockDurationResponse)

      const result = await metricsApi.getDurationStats()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/metrics/duration')
      expect(result).toEqual(mockDurationResponse)
    })

    it('should fetch duration stats with days param', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockDurationResponse)

      await metricsApi.getDurationStats({ days: 14 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/metrics/duration?days=14')
    })

    it('should fetch duration stats with date range', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockDurationResponse)

      await metricsApi.getDurationStats({
        startDate: '2024-01-01',
        endDate: '2024-01-31',
      })

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/v1/metrics/duration?startDate=2024-01-01&endDate=2024-01-31'
      )
    })

    it('should handle empty stats', async () => {
      const emptyResponse: DurationStatsResponse = {
        stats: [],
        startDate: '2024-01-15',
        endDate: '2024-01-16',
      }
      ;(apiClient.get as any).mockResolvedValueOnce(emptyResponse)

      const result = await metricsApi.getDurationStats()

      expect(result.stats).toEqual([])
    })
  })

  describe('getTopFailures', () => {
    it('should fetch top failures without params', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockFailuresResponse)

      const result = await metricsApi.getTopFailures()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/metrics/failures')
      expect(result).toEqual(mockFailuresResponse)
    })

    it('should fetch top failures with days param', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockFailuresResponse)

      await metricsApi.getTopFailures({ days: 7 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/metrics/failures?days=7')
    })

    it('should fetch top failures with limit param', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockFailuresResponse)

      await metricsApi.getTopFailures({ limit: 20 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/metrics/failures?limit=20')
    })

    it('should fetch top failures with all params', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockFailuresResponse)

      await metricsApi.getTopFailures({
        days: 30,
        startDate: '2024-01-01',
        endDate: '2024-01-31',
        limit: 50,
      })

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/v1/metrics/failures?days=30&startDate=2024-01-01&endDate=2024-01-31&limit=50'
      )
    })

    it('should handle empty failures', async () => {
      const emptyResponse: TopFailuresResponse = {
        failures: [],
        startDate: '2024-01-15',
        endDate: '2024-01-16',
        limit: 10,
      }
      ;(apiClient.get as any).mockResolvedValueOnce(emptyResponse)

      const result = await metricsApi.getTopFailures()

      expect(result.failures).toEqual([])
    })
  })

  describe('getTriggerBreakdown', () => {
    it('should fetch trigger breakdown without params', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTriggerBreakdown)

      const result = await metricsApi.getTriggerBreakdown()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/metrics/trigger-breakdown')
      expect(result).toEqual(mockTriggerBreakdown)
    })

    it('should fetch trigger breakdown with days param', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTriggerBreakdown)

      await metricsApi.getTriggerBreakdown({ days: 30 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/metrics/trigger-breakdown?days=30')
    })

    it('should fetch trigger breakdown with date range', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTriggerBreakdown)

      await metricsApi.getTriggerBreakdown({
        startDate: '2024-01-01',
        endDate: '2024-01-31',
      })

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/v1/metrics/trigger-breakdown?startDate=2024-01-01&endDate=2024-01-31'
      )
    })

    it('should return correct percentage totals', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockTriggerBreakdown)

      const result = await metricsApi.getTriggerBreakdown()

      const totalPercentage = result.breakdown.reduce((sum, item) => sum + item.percentage, 0)
      expect(totalPercentage).toBe(100)
    })
  })
})
