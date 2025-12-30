import { describe, it, expect, beforeEach, vi } from 'vitest'
import { analyticsAPI } from './analytics'
import { apiClient } from './client'

vi.mock('./client')

describe('analyticsAPI', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getTenantOverview', () => {
    it('should fetch tenant overview', async () => {
      const mockOverview = {
        totalExecutions: 500,
        successfulExecutions: 450,
        failedExecutions: 50,
        cancelledExecutions: 0,
        pendingExecutions: 0,
        runningExecutions: 0,
        successRate: 0.9,
        avgDurationMs: 2000,
        activeWorkflows: 10,
        totalWorkflows: 15,
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockOverview)

      const result = await analyticsAPI.getTenantOverview({
        startDate: '2024-01-01T00:00:00Z',
        endDate: '2024-01-31T23:59:59Z',
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/analytics/overview', {
        params: {
          start_date: '2024-01-01T00:00:00Z',
          end_date: '2024-01-31T23:59:59Z',
        },
      })
      expect(result).toEqual(mockOverview)
    })
  })

  describe('getWorkflowStats', () => {
    it('should fetch workflow statistics', async () => {
      const mockStats = {
        workflowId: 'workflow-1',
        workflowName: 'Test Workflow',
        executionCount: 100,
        successCount: 90,
        failureCount: 10,
        successRate: 0.9,
        avgDurationMs: 1500,
        minDurationMs: 100,
        maxDurationMs: 5000,
        lastExecutedAt: '2024-01-31T12:00:00Z',
        totalDurationMs: 150000,
        cancelledCount: 0,
        pendingCount: 0,
        runningCount: 0,
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockStats)

      const result = await analyticsAPI.getWorkflowStats('workflow-1', {
        startDate: '2024-01-01T00:00:00Z',
        endDate: '2024-01-31T23:59:59Z',
      })

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/v1/analytics/workflows/workflow-1',
        {
          params: {
            start_date: '2024-01-01T00:00:00Z',
            end_date: '2024-01-31T23:59:59Z',
          },
        }
      )
      expect(result).toEqual(mockStats)
    })
  })

  describe('getExecutionTrends', () => {
    it('should fetch execution trends', async () => {
      const mockTrends = {
        granularity: 'day' as const,
        startDate: '2024-01-01T00:00:00Z',
        endDate: '2024-01-07T23:59:59Z',
        dataPoints: [
          {
            timestamp: '2024-01-01T00:00:00Z',
            executionCount: 50,
            successCount: 45,
            failureCount: 5,
            successRate: 0.9,
            avgDurationMs: 1500,
          },
        ],
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockTrends)

      const result = await analyticsAPI.getExecutionTrends({
        startDate: '2024-01-01T00:00:00Z',
        endDate: '2024-01-07T23:59:59Z',
        granularity: 'day',
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/analytics/trends', {
        params: {
          start_date: '2024-01-01T00:00:00Z',
          end_date: '2024-01-07T23:59:59Z',
          granularity: 'day',
        },
      })
      expect(result).toEqual(mockTrends)
    })

    it('should use default granularity if not provided', async () => {
      const mockTrends = {
        granularity: 'day' as const,
        startDate: '2024-01-01T00:00:00Z',
        endDate: '2024-01-07T23:59:59Z',
        dataPoints: [],
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockTrends)

      await analyticsAPI.getExecutionTrends({
        startDate: '2024-01-01T00:00:00Z',
        endDate: '2024-01-07T23:59:59Z',
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/analytics/trends', {
        params: {
          start_date: '2024-01-01T00:00:00Z',
          end_date: '2024-01-07T23:59:59Z',
          granularity: 'day',
        },
      })
    })
  })

  describe('getTopWorkflows', () => {
    it('should fetch top workflows', async () => {
      const mockTopWorkflows = {
        workflows: [
          {
            workflowId: 'workflow-1',
            workflowName: 'Top Workflow',
            executionCount: 500,
            successRate: 0.95,
            avgDurationMs: 1000,
            lastExecutedAt: '2024-01-31T12:00:00Z',
          },
        ],
        total: 1,
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockTopWorkflows)

      const result = await analyticsAPI.getTopWorkflows({
        startDate: '2024-01-01T00:00:00Z',
        endDate: '2024-01-31T23:59:59Z',
        limit: 10,
      })

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/v1/analytics/top-workflows',
        {
          params: {
            start_date: '2024-01-01T00:00:00Z',
            end_date: '2024-01-31T23:59:59Z',
            limit: 10,
          },
        }
      )
      expect(result).toEqual(mockTopWorkflows)
    })
  })

  describe('getErrorBreakdown', () => {
    it('should fetch error breakdown', async () => {
      const mockBreakdown = {
        totalErrors: 40,
        errorsByType: [
          {
            errorMessage: 'Connection timeout',
            errorCount: 20,
            workflowId: 'workflow-1',
            workflowName: 'Test Workflow',
            lastOccurrence: '2024-01-31T12:00:00Z',
            percentage: 50.0,
          },
        ],
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockBreakdown)

      const result = await analyticsAPI.getErrorBreakdown({
        startDate: '2024-01-01T00:00:00Z',
        endDate: '2024-01-31T23:59:59Z',
      })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/analytics/errors', {
        params: {
          start_date: '2024-01-01T00:00:00Z',
          end_date: '2024-01-31T23:59:59Z',
        },
      })
      expect(result).toEqual(mockBreakdown)
    })
  })

  describe('getNodePerformance', () => {
    it('should fetch node performance', async () => {
      const mockPerformance = {
        workflowId: 'workflow-1',
        workflowName: 'Test Workflow',
        nodes: [
          {
            nodeId: 'node-1',
            nodeType: 'action:http',
            executionCount: 100,
            successCount: 95,
            failureCount: 5,
            successRate: 0.95,
            avgDurationMs: 500,
            minDurationMs: 100,
            maxDurationMs: 2000,
          },
        ],
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockPerformance)

      const result = await analyticsAPI.getNodePerformance('workflow-1')

      expect(apiClient.get).toHaveBeenCalledWith(
        '/api/v1/analytics/workflows/workflow-1/nodes'
      )
      expect(result).toEqual(mockPerformance)
    })
  })
})
