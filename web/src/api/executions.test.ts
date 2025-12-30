import { describe, it, expect, beforeEach, vi } from 'vitest'
import { executionAPI } from './executions'
import { apiClient } from './client'
import type { Execution, ExecutionStatus } from './executions'

vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

describe('Execution API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

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

  describe('list', () => {
    it('should fetch executions list', async () => {
      const mockResponse = {
        executions: [mockExecution],
        total: 1,
      }

      ;(apiClient.get as any).mockResolvedValueOnce(mockResponse)

      const result = await executionAPI.list()

      expect(apiClient.get).toHaveBeenCalledWith('/executions', { params: {} })
      expect(result).toEqual(mockResponse)
    })

    it('should support pagination', async () => {
      const mockResponse = {
        executions: [mockExecution],
        total: 100,
      }

      ;(apiClient.get as any).mockResolvedValueOnce(mockResponse)

      await executionAPI.list({ page: 2, limit: 20 })

      expect(apiClient.get).toHaveBeenCalledWith('/executions', {
        params: { page: 2, limit: 20 },
      })
    })

    it('should support filtering by workflow', async () => {
      const mockResponse = {
        executions: [mockExecution],
        total: 1,
      }

      ;(apiClient.get as any).mockResolvedValueOnce(mockResponse)

      await executionAPI.list({ workflowId: 'wf-123' })

      expect(apiClient.get).toHaveBeenCalledWith('/executions', {
        params: { workflowId: 'wf-123' },
      })
    })

    it('should support filtering by status', async () => {
      const mockResponse = {
        executions: [mockExecution],
        total: 1,
      }

      ;(apiClient.get as any).mockResolvedValueOnce(mockResponse)

      await executionAPI.list({ status: 'failed' })

      expect(apiClient.get).toHaveBeenCalledWith('/executions', {
        params: { status: 'failed' },
      })
    })

    it('should support date range filtering', async () => {
      const mockResponse = {
        executions: [mockExecution],
        total: 1,
      }

      ;(apiClient.get as any).mockResolvedValueOnce(mockResponse)

      await executionAPI.list({
        startDate: '2025-01-01',
        endDate: '2025-01-31',
      })

      expect(apiClient.get).toHaveBeenCalledWith('/executions', {
        params: {
          startDate: '2025-01-01',
          endDate: '2025-01-31',
        },
      })
    })
  })

  describe('get', () => {
    it('should fetch single execution by ID', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockExecution)

      const result = await executionAPI.get('exec-123')

      expect(apiClient.get).toHaveBeenCalledWith('/executions/exec-123')
      expect(result).toEqual(mockExecution)
    })

    it('should handle not found error', async () => {
      const error = new Error('Execution not found')
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(executionAPI.get('invalid-id')).rejects.toThrow(
        'Execution not found'
      )
    })
  })

  describe('cancel', () => {
    it('should cancel running execution', async () => {
      const cancelledExecution = {
        ...mockExecution,
        status: 'cancelled' as ExecutionStatus,
      }

      ;(apiClient.post as any).mockResolvedValueOnce(cancelledExecution)

      const result = await executionAPI.cancel('exec-123')

      expect(apiClient.post).toHaveBeenCalledWith('/executions/exec-123/cancel', {})
      expect(result.status).toBe('cancelled')
    })
  })

  describe('retry', () => {
    it('should retry failed execution', async () => {
      const newExecution = {
        ...mockExecution,
        id: 'exec-456',
        status: 'queued' as ExecutionStatus,
      }

      ;(apiClient.post as any).mockResolvedValueOnce(newExecution)

      const result = await executionAPI.retry('exec-123')

      expect(apiClient.post).toHaveBeenCalledWith('/executions/exec-123/retry', {})
      expect(result.id).toBe('exec-456')
      expect(result.status).toBe('queued')
    })
  })

  describe('getSteps', () => {
    it('should fetch execution steps', async () => {
      const mockSteps = [
        {
          id: 'step-1',
          executionId: 'exec-123',
          nodeId: 'node-1',
          nodeName: 'HTTP Request',
          status: 'completed' as ExecutionStatus,
          startedAt: '2025-01-15T10:00:00Z',
          completedAt: '2025-01-15T10:01:00Z',
          duration: 60000,
          input: { url: 'https://api.example.com' },
          output: { statusCode: 200, body: {} },
        },
      ]

      ;(apiClient.get as any).mockResolvedValueOnce({ steps: mockSteps })

      const result = await executionAPI.getSteps('exec-123')

      expect(apiClient.get).toHaveBeenCalledWith('/executions/exec-123/steps')
      expect(result.steps).toEqual(mockSteps)
    })
  })

  describe('getDashboardStats', () => {
    it('should fetch dashboard statistics from combined endpoints', async () => {
      const mockStats = {
        total_count: 1000,
        status_counts: {
          completed: 847,
          failed: 3,
        },
      }
      const mockWorkflows = {
        data: [{ id: 'wf-1' }],
        total_count: 12,
      }

      ;(apiClient.get as any)
        .mockResolvedValueOnce(mockStats)
        .mockResolvedValueOnce(mockWorkflows)

      const result = await executionAPI.getDashboardStats()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/executions/stats')
      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/workflows', { params: { limit: 1 } })
      expect(result.totalExecutions).toBe(1000)
      expect(result.failedToday).toBe(3)
      expect(result.activeWorkflows).toBe(12)
    })
  })

  describe('getRecentExecutions', () => {
    it('should fetch recent executions', async () => {
      const recentExecutions = [mockExecution]

      ;(apiClient.get as any).mockResolvedValueOnce({
        data: recentExecutions,
      })

      const result = await executionAPI.getRecentExecutions()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/executions', {
        params: { limit: 10 },
      })
      expect(result.executions).toEqual(recentExecutions)
    })

    it('should support custom limit', async () => {
      const recentExecutions = [mockExecution]

      ;(apiClient.get as any).mockResolvedValueOnce({
        data: recentExecutions,
      })

      await executionAPI.getRecentExecutions(5)

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/executions', {
        params: { limit: 5 },
      })
    })
  })
})
