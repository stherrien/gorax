import { apiClient } from './client'
import type {
  TenantOverview,
  WorkflowStats,
  ExecutionTrends,
  TopWorkflows,
  ErrorBreakdown,
  NodePerformance,
  AnalyticsParams,
} from '../types/analytics'

/**
 * Analytics API client
 */
export const analyticsAPI = {
  /**
   * Get overall tenant analytics
   */
  async getTenantOverview(params: AnalyticsParams): Promise<TenantOverview> {
    return apiClient.get('/api/v1/analytics/overview', {
      params: {
        start_date: params.startDate,
        end_date: params.endDate,
      },
    })
  },

  /**
   * Get statistics for a specific workflow
   */
  async getWorkflowStats(
    workflowId: string,
    params: AnalyticsParams
  ): Promise<WorkflowStats> {
    return apiClient.get(`/api/v1/analytics/workflows/${workflowId}`, {
      params: {
        start_date: params.startDate,
        end_date: params.endDate,
      },
    })
  },

  /**
   * Get execution trends over time
   */
  async getExecutionTrends(params: AnalyticsParams): Promise<ExecutionTrends> {
    return apiClient.get('/api/v1/analytics/trends', {
      params: {
        start_date: params.startDate,
        end_date: params.endDate,
        granularity: params.granularity || 'day',
      },
    })
  },

  /**
   * Get top workflows by execution count
   */
  async getTopWorkflows(params: AnalyticsParams): Promise<TopWorkflows> {
    return apiClient.get('/api/v1/analytics/top-workflows', {
      params: {
        start_date: params.startDate,
        end_date: params.endDate,
        limit: params.limit || 10,
      },
    })
  },

  /**
   * Get error breakdown analysis
   */
  async getErrorBreakdown(params: AnalyticsParams): Promise<ErrorBreakdown> {
    return apiClient.get('/api/v1/analytics/errors', {
      params: {
        start_date: params.startDate,
        end_date: params.endDate,
      },
    })
  },

  /**
   * Get node-level performance for a workflow
   */
  async getNodePerformance(workflowId: string): Promise<NodePerformance> {
    return apiClient.get(`/api/v1/analytics/workflows/${workflowId}/nodes`)
  },
}
