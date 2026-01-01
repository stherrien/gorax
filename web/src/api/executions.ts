import { apiClient } from './client'

/**
 * Execution status
 */
export type ExecutionStatus =
  | 'queued'
  | 'running'
  | 'completed'
  | 'failed'
  | 'cancelled'
  | 'timeout'

/**
 * Execution trigger information
 */
export interface ExecutionTrigger {
  type: 'webhook' | 'schedule' | 'manual'
  source?: string
  userId?: string
}

/**
 * Execution model
 */
export interface Execution {
  id: string
  workflowId: string
  workflowName: string
  status: ExecutionStatus
  trigger: ExecutionTrigger
  startedAt: string
  completedAt?: string
  duration?: number
  stepCount: number
  completedSteps: number
  failedSteps: number
  error?: string
}

/**
 * Execution step model
 */
export interface ExecutionStep {
  id: string
  executionId: string
  nodeId: string
  nodeName: string
  status: ExecutionStatus
  startedAt: string
  completedAt?: string
  duration?: number
  input?: Record<string, unknown>
  output?: Record<string, unknown>
  error?: string
}

/**
 * List executions parameters
 */
export interface ExecutionListParams {
  page?: number
  limit?: number
  workflowId?: string
  status?: ExecutionStatus | ExecutionStatus[]
  triggerType?: string | string[]
  startDate?: string
  endDate?: string
  errorSearch?: string
  executionIdPrefix?: string
  minDurationMs?: number
  maxDurationMs?: number
}

/**
 * Execution list response
 */
export interface ExecutionListResponse {
  executions: Execution[]
  total: number
}

/**
 * Execution steps response
 */
export interface ExecutionStepsResponse {
  steps: ExecutionStep[]
}

/**
 * Dashboard statistics
 */
export interface DashboardStats {
  totalExecutions: number
  executionsToday: number
  failedToday: number
  successRateToday: number
  averageDuration: number
  activeWorkflows: number
}

/**
 * Dashboard stats parameters
 */
export interface DashboardStatsParams {
  startDate?: string
  endDate?: string
}

/**
 * Recent executions response
 */
export interface RecentExecutionsResponse {
  executions: Execution[]
}

/**
 * Execution API client
 */
export const executionAPI = {
  /**
   * List executions with optional filtering and pagination
   */
  async list(params?: ExecutionListParams): Promise<ExecutionListResponse> {
    return apiClient.get('/executions', { params: params || {} })
  },

  /**
   * Get single execution by ID
   */
  async get(id: string): Promise<Execution> {
    return apiClient.get(`/executions/${id}`)
  },

  /**
   * Cancel a running execution
   */
  async cancel(id: string): Promise<Execution> {
    return apiClient.post(`/executions/${id}/cancel`, {})
  },

  /**
   * Retry a failed execution
   */
  async retry(id: string): Promise<Execution> {
    return apiClient.post(`/executions/${id}/retry`, {})
  },

  /**
   * Get execution steps (detailed step-by-step results)
   */
  async getSteps(id: string): Promise<ExecutionStepsResponse> {
    return apiClient.get(`/executions/${id}/steps`)
  },

  /**
   * Get dashboard statistics
   * Uses the existing stats endpoint and transforms to dashboard format
   */
  async getDashboardStats(_params?: DashboardStatsParams): Promise<DashboardStats> {
    // Fetch stats and workflows count in parallel
    const [stats, workflowsResponse] = await Promise.all([
      apiClient.get('/api/v1/executions/stats'),
      apiClient.get('/api/v1/workflows', { params: { limit: 1 } }),
    ])

    const statusCounts = stats.status_counts || {}
    const completed = statusCounts.completed || 0
    const failed = statusCounts.failed || 0
    const total = stats.total_count || 0
    const workflows = workflowsResponse.data || []

    return {
      totalExecutions: total,
      executionsToday: completed + failed, // Approximate - would need date filtering
      failedToday: failed,
      successRateToday: total > 0 ? (completed / total) * 100 : 100,
      averageDuration: 0, // Not available from this endpoint
      activeWorkflows: workflows.length > 0 ? (workflowsResponse.total_count || workflows.length) : 0,
    }
  },

  /**
   * Get recent executions (for dashboard)
   * Uses the list endpoint with a limit
   */
  async getRecentExecutions(limit: number = 10): Promise<RecentExecutionsResponse> {
    const response = await apiClient.get('/api/v1/executions', { params: { limit } })
    return {
      executions: response.data || [],
    }
  },

  /**
   * Bulk delete executions
   */
  async bulkDelete(ids: string[]): Promise<BulkOperationResult> {
    // Use POST to /bulk/delete since DELETE doesn't support body in our API client
    return apiClient.post('/api/v1/executions/bulk/delete', { ids })
  },

  /**
   * Bulk retry failed executions
   */
  async bulkRetry(ids: string[]): Promise<BulkOperationResult> {
    return apiClient.post('/api/v1/executions/bulk/retry', { ids })
  },
}

/**
 * Bulk operation result
 */
export interface BulkOperationResult {
  success: string[]
  failed: BulkOperationError[]
}

/**
 * Bulk operation error
 */
export interface BulkOperationError {
  id: string
  error: string
}
