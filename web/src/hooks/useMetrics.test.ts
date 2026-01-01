import { renderHook, waitFor, act } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  useExecutionTrends,
  useDurationStats,
  useTopFailures,
  useTriggerBreakdown,
  useAllMetrics,
} from './useMetrics'
import metricsApi from '../api/metrics'

// Mock the metrics API
vi.mock('../api/metrics', () => ({
  default: {
    getExecutionTrends: vi.fn(),
    getDurationStats: vi.fn(),
    getTopFailures: vi.fn(),
    getTriggerBreakdown: vi.fn(),
  },
}))

describe('useMetrics hooks', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.resetAllMocks()
  })

  describe('useExecutionTrends', () => {
    const mockTrends = [
      { date: '2025-01-01', count: 10, success: 8, failed: 2 },
      { date: '2025-01-02', count: 15, success: 12, failed: 3 },
    ]

    it('should fetch execution trends on mount', async () => {
      vi.mocked(metricsApi.getExecutionTrends).mockResolvedValue({
        trends: mockTrends,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
        groupBy: 'day',
      })

      const { result } = renderHook(() => useExecutionTrends())

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.trends).toEqual(mockTrends)
      expect(result.current.error).toBeNull()
    })

    it('should handle fetch error', async () => {
      vi.mocked(metricsApi.getExecutionTrends).mockRejectedValue(
        new Error('Network error')
      )

      const { result } = renderHook(() => useExecutionTrends())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe('Network error')
      expect(result.current.trends).toEqual([])
    })

    it('should handle non-Error rejection', async () => {
      vi.mocked(metricsApi.getExecutionTrends).mockRejectedValue('Unknown error')

      const { result } = renderHook(() => useExecutionTrends())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe('Failed to fetch trends')
    })

    it('should refetch when params change', async () => {
      vi.mocked(metricsApi.getExecutionTrends).mockResolvedValue({
        trends: mockTrends,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
        groupBy: 'day',
      })

      const { result, rerender } = renderHook(
        ({ params }) => useExecutionTrends(params),
        { initialProps: { params: { days: 7 } } }
      )

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(metricsApi.getExecutionTrends).toHaveBeenCalledWith({ days: 7 })

      rerender({ params: { days: 30 } })

      await waitFor(() => {
        expect(metricsApi.getExecutionTrends).toHaveBeenCalledWith({ days: 30 })
      })
    })

    it('should provide refetch function', async () => {
      vi.mocked(metricsApi.getExecutionTrends).mockResolvedValue({
        trends: mockTrends,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
        groupBy: 'day',
      })

      const { result } = renderHook(() => useExecutionTrends())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(metricsApi.getExecutionTrends).toHaveBeenCalledTimes(1)

      await act(async () => {
        await result.current.refetch()
      })

      expect(metricsApi.getExecutionTrends).toHaveBeenCalledTimes(2)
    })
  })

  describe('useDurationStats', () => {
    const mockStats = [
      {
        workflowId: 'wf-1',
        workflowName: 'Workflow 1',
        avgDuration: 1500,
        p50Duration: 1200,
        p90Duration: 2000,
        p99Duration: 2500,
        totalRuns: 100,
      },
    ]

    it('should fetch duration stats on mount', async () => {
      vi.mocked(metricsApi.getDurationStats).mockResolvedValue({
        stats: mockStats,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
      })

      const { result } = renderHook(() => useDurationStats())

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.stats).toEqual(mockStats)
      expect(result.current.error).toBeNull()
    })

    it('should handle fetch error', async () => {
      vi.mocked(metricsApi.getDurationStats).mockRejectedValue(
        new Error('API error')
      )

      const { result } = renderHook(() => useDurationStats())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe('API error')
    })

    it('should handle non-Error rejection', async () => {
      vi.mocked(metricsApi.getDurationStats).mockRejectedValue('Unknown')

      const { result } = renderHook(() => useDurationStats())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe('Failed to fetch duration stats')
    })

    it('should pass params to API', async () => {
      vi.mocked(metricsApi.getDurationStats).mockResolvedValue({
        stats: mockStats,
        startDate: '2025-01-01',
        endDate: '2025-01-07',
      })

      const { result } = renderHook(() =>
        useDurationStats({ startDate: '2025-01-01', endDate: '2025-01-07' })
      )

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(metricsApi.getDurationStats).toHaveBeenCalledWith({
        startDate: '2025-01-01',
        endDate: '2025-01-07',
      })
    })
  })

  describe('useTopFailures', () => {
    const mockFailures = [
      {
        workflowId: 'wf-1',
        workflowName: 'Failing Workflow',
        failureCount: 25,
        lastFailedAt: '2025-01-20T10:00:00Z',
        errorPreview: 'Connection timeout',
      },
    ]

    it('should fetch top failures on mount', async () => {
      vi.mocked(metricsApi.getTopFailures).mockResolvedValue({
        failures: mockFailures,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
        limit: 10,
      })

      const { result } = renderHook(() => useTopFailures())

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.failures).toEqual(mockFailures)
      expect(result.current.error).toBeNull()
    })

    it('should handle fetch error', async () => {
      vi.mocked(metricsApi.getTopFailures).mockRejectedValue(
        new Error('Server error')
      )

      const { result } = renderHook(() => useTopFailures())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe('Server error')
    })

    it('should handle non-Error rejection', async () => {
      vi.mocked(metricsApi.getTopFailures).mockRejectedValue(null)

      const { result } = renderHook(() => useTopFailures())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe('Failed to fetch failures')
    })

    it('should pass limit param to API', async () => {
      vi.mocked(metricsApi.getTopFailures).mockResolvedValue({
        failures: mockFailures,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
        limit: 5,
      })

      const { result } = renderHook(() => useTopFailures({ limit: 5 }))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(metricsApi.getTopFailures).toHaveBeenCalledWith({ limit: 5 })
    })
  })

  describe('useTriggerBreakdown', () => {
    const mockBreakdown = [
      { triggerType: 'webhook', count: 50, percentage: 50 },
      { triggerType: 'schedule', count: 30, percentage: 30 },
      { triggerType: 'manual', count: 20, percentage: 20 },
    ]

    it('should fetch trigger breakdown on mount', async () => {
      vi.mocked(metricsApi.getTriggerBreakdown).mockResolvedValue({
        breakdown: mockBreakdown,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
      })

      const { result } = renderHook(() => useTriggerBreakdown())

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.breakdown).toEqual(mockBreakdown)
      expect(result.current.error).toBeNull()
    })

    it('should handle fetch error', async () => {
      vi.mocked(metricsApi.getTriggerBreakdown).mockRejectedValue(
        new Error('Timeout')
      )

      const { result } = renderHook(() => useTriggerBreakdown())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe('Timeout')
    })

    it('should handle non-Error rejection', async () => {
      vi.mocked(metricsApi.getTriggerBreakdown).mockRejectedValue(undefined)

      const { result } = renderHook(() => useTriggerBreakdown())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe('Failed to fetch trigger breakdown')
    })

    it('should refetch on param changes', async () => {
      vi.mocked(metricsApi.getTriggerBreakdown).mockResolvedValue({
        breakdown: mockBreakdown,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
      })

      const { result, rerender } = renderHook(
        ({ params }) => useTriggerBreakdown(params),
        { initialProps: { params: { days: 7 } } }
      )

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      rerender({ params: { days: 14 } })

      await waitFor(() => {
        expect(metricsApi.getTriggerBreakdown).toHaveBeenCalledWith({ days: 14 })
      })
    })
  })

  describe('useAllMetrics', () => {
    const mockTrends = [{ date: '2025-01-01', count: 10, success: 8, failed: 2 }]
    const mockStats = [{
      workflowId: 'wf-1',
      workflowName: 'Workflow 1',
      avgDuration: 1500,
      p50Duration: 1200,
      p90Duration: 2000,
      p99Duration: 2500,
      totalRuns: 100,
    }]
    const mockFailures = [{
      workflowId: 'wf-1',
      workflowName: 'Failing Workflow',
      failureCount: 25,
    }]
    const mockBreakdown = [{ triggerType: 'webhook', count: 50, percentage: 50 }]

    beforeEach(() => {
      vi.mocked(metricsApi.getExecutionTrends).mockResolvedValue({
        trends: mockTrends,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
        groupBy: 'day',
      })
      vi.mocked(metricsApi.getDurationStats).mockResolvedValue({
        stats: mockStats,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
      })
      vi.mocked(metricsApi.getTopFailures).mockResolvedValue({
        failures: mockFailures,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
        limit: 10,
      })
      vi.mocked(metricsApi.getTriggerBreakdown).mockResolvedValue({
        breakdown: mockBreakdown,
        startDate: '2025-01-01',
        endDate: '2025-01-02',
      })
    })

    it('should fetch all metrics on mount', async () => {
      const { result } = renderHook(() => useAllMetrics())

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.trends).toEqual(mockTrends)
      expect(result.current.durationStats).toEqual(mockStats)
      expect(result.current.failures).toEqual(mockFailures)
      expect(result.current.triggerBreakdown).toEqual(mockBreakdown)
      expect(result.current.error).toBeNull()
    })

    it('should show loading when any metric is loading', async () => {
      // Make one API call take longer
      let resolveTrends: (value: any) => void
      vi.mocked(metricsApi.getExecutionTrends).mockImplementation(
        () =>
          new Promise((resolve) => {
            resolveTrends = resolve
          })
      )

      const { result } = renderHook(() => useAllMetrics())

      // While trends is still loading
      expect(result.current.loading).toBe(true)

      // Resolve the pending call
      act(() => {
        resolveTrends!({
          trends: mockTrends,
          startDate: '2025-01-01',
          endDate: '2025-01-02',
          groupBy: 'day',
        })
      })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })
    })

    it('should show error when any metric fails', async () => {
      vi.mocked(metricsApi.getTopFailures).mockRejectedValue(
        new Error('Failures API error')
      )

      const { result } = renderHook(() => useAllMetrics())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe('Failures API error')
    })

    it('should provide refetch function that refetches all metrics', async () => {
      const { result } = renderHook(() => useAllMetrics())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      // Clear call counts
      vi.clearAllMocks()

      await act(async () => {
        result.current.refetch()
      })

      await waitFor(() => {
        expect(metricsApi.getExecutionTrends).toHaveBeenCalled()
        expect(metricsApi.getDurationStats).toHaveBeenCalled()
        expect(metricsApi.getTopFailures).toHaveBeenCalled()
        expect(metricsApi.getTriggerBreakdown).toHaveBeenCalled()
      })
    })

    it('should pass params to all API calls', async () => {
      const params = { days: 30 }

      const { result } = renderHook(() => useAllMetrics(params))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(metricsApi.getExecutionTrends).toHaveBeenCalledWith(params)
      expect(metricsApi.getDurationStats).toHaveBeenCalledWith(params)
      expect(metricsApi.getTopFailures).toHaveBeenCalledWith(params)
      expect(metricsApi.getTriggerBreakdown).toHaveBeenCalledWith(params)
    })
  })
})
