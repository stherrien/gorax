import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import {
  useExecutions,
  useExecution,
  useDashboardStats,
  useRecentExecutions,
} from './useExecutions'
import type { Execution, DashboardStats } from '../api/executions'

// Mock the execution API
vi.mock('../api/executions', () => ({
  executionAPI: {
    list: vi.fn(),
    get: vi.fn(),
    getDashboardStats: vi.fn(),
    getRecentExecutions: vi.fn(),
  },
}))

import { executionAPI } from '../api/executions'

describe('useExecutions', () => {
  const mockExecution: Execution = {
    id: 'exec-123',
    workflowId: 'wf-123',
    workflowName: 'Test Workflow',
    status: 'completed',
    trigger: {
      type: 'webhook',
      source: 'api',
    },
    startedAt: '2025-01-15T10:00:00Z',
    completedAt: '2025-01-15T10:05:00Z',
    duration: 300000,
    stepCount: 5,
    completedSteps: 5,
    failedSteps: 0,
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('useExecutions - list hook', () => {
    it('should load executions on mount', async () => {
      (executionAPI.list as any).mockResolvedValueOnce({
        executions: [mockExecution],
        total: 1,
      })

      const { result } = renderHook(() => useExecutions())

      expect(result.current.loading).toBe(true)
      expect(result.current.executions).toEqual([])

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.executions).toEqual([mockExecution])
      expect(result.current.total).toBe(1)
      expect(result.current.error).toBeNull()
    })

    it('should handle empty list', async () => {
      (executionAPI.list as any).mockResolvedValueOnce({
        executions: [],
        total: 0,
      })

      const { result } = renderHook(() => useExecutions())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.executions).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('should handle errors', async () => {
      const error = new Error('Failed to load executions')
      ;(executionAPI.list as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useExecutions())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.executions).toEqual([])
    })

    it('should support refetch', async () => {
      (executionAPI.list as any).mockResolvedValue({
        executions: [mockExecution],
        total: 1,
      })

      const { result } = renderHook(() => useExecutions())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(executionAPI.list).toHaveBeenCalledTimes(1)

      // Trigger refetch
      result.current.refetch()

      await waitFor(() => {
        expect(executionAPI.list).toHaveBeenCalledTimes(2)
      })
    })

    it('should pass filter params to API', async () => {
      (executionAPI.list as any).mockResolvedValueOnce({
        executions: [],
        total: 0,
      })

      renderHook(() => useExecutions({ workflowId: 'wf-123', status: 'failed' }))

      await waitFor(() => {
        expect(executionAPI.list).toHaveBeenCalledWith({
          workflowId: 'wf-123',
          status: 'failed',
        })
      })
    })
  })

  describe('useExecution - single execution hook', () => {
    it('should load execution by ID on mount', async () => {
      (executionAPI.get as any).mockResolvedValueOnce(mockExecution)

      const { result } = renderHook(() => useExecution('exec-123'))

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.execution).toEqual(mockExecution)
      expect(result.current.error).toBeNull()
    })

    it('should not load if ID is null', () => {
      const { result } = renderHook(() => useExecution(null))

      expect(result.current.loading).toBe(false)
      expect(result.current.execution).toBeNull()
      expect(executionAPI.get).not.toHaveBeenCalled()
    })

    it('should handle not found error', async () => {
      const error = new Error('Execution not found')
      ;(executionAPI.get as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useExecution('invalid-id'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.execution).toBeNull()
    })

    it('should refetch execution', async () => {
      (executionAPI.get as any).mockResolvedValue(mockExecution)

      const { result } = renderHook(() => useExecution('exec-123'))

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      result.current.refetch()

      await waitFor(() => {
        expect(executionAPI.get).toHaveBeenCalledTimes(2)
      })
    })
  })

  describe('useDashboardStats', () => {
    const mockStats: DashboardStats = {
      totalExecutions: 1000,
      executionsToday: 847,
      failedToday: 3,
      successRateToday: 99.6,
      averageDuration: 45000,
      activeWorkflows: 12,
    }

    it('should load dashboard stats on mount', async () => {
      (executionAPI.getDashboardStats as any).mockResolvedValueOnce(mockStats)

      const { result } = renderHook(() => useDashboardStats())

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.stats).toEqual(mockStats)
      expect(result.current.error).toBeNull()
    })

    it('should handle errors', async () => {
      const error = new Error('Failed to load stats')
      ;(executionAPI.getDashboardStats as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useDashboardStats())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.stats).toBeNull()
    })

    it('should support refetch', async () => {
      (executionAPI.getDashboardStats as any).mockResolvedValue(mockStats)

      const { result } = renderHook(() => useDashboardStats())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      result.current.refetch()

      await waitFor(() => {
        expect(executionAPI.getDashboardStats).toHaveBeenCalledTimes(2)
      })
    })

    it('should pass date range params to API', async () => {
      (executionAPI.getDashboardStats as any).mockResolvedValueOnce(mockStats)

      renderHook(() =>
        useDashboardStats({ startDate: '2025-01-01', endDate: '2025-01-31' })
      )

      await waitFor(() => {
        expect(executionAPI.getDashboardStats).toHaveBeenCalledWith({
          startDate: '2025-01-01',
          endDate: '2025-01-31',
        })
      })
    })
  })

  describe('useRecentExecutions', () => {
    it('should load recent executions on mount', async () => {
      (executionAPI.getRecentExecutions as any).mockResolvedValueOnce({
        executions: [mockExecution],
      })

      const { result } = renderHook(() => useRecentExecutions())

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.executions).toEqual([mockExecution])
      expect(result.current.error).toBeNull()
    })

    it('should handle empty list', async () => {
      (executionAPI.getRecentExecutions as any).mockResolvedValueOnce({
        executions: [],
      })

      const { result } = renderHook(() => useRecentExecutions())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.executions).toEqual([])
    })

    it('should handle errors', async () => {
      const error = new Error('Failed to load recent executions')
      ;(executionAPI.getRecentExecutions as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useRecentExecutions())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.executions).toEqual([])
    })

    it('should support custom limit', async () => {
      (executionAPI.getRecentExecutions as any).mockResolvedValueOnce({
        executions: [mockExecution],
      })

      renderHook(() => useRecentExecutions(5))

      await waitFor(() => {
        expect(executionAPI.getRecentExecutions).toHaveBeenCalledWith(5)
      })
    })

    it('should support refetch', async () => {
      (executionAPI.getRecentExecutions as any).mockResolvedValue({
        executions: [mockExecution],
      })

      const { result } = renderHook(() => useRecentExecutions())

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      result.current.refetch()

      await waitFor(() => {
        expect(executionAPI.getRecentExecutions).toHaveBeenCalledTimes(2)
      })
    })
  })
})
