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
  status?: ExecutionStatus
  startDate?: string
  endDate?: string
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
   */
  async getDashboardStats(params?: DashboardStatsParams): Promise<DashboardStats> {
    return apiClient.get('/executions/stats/dashboard', { params: params || {} })
  },

  /**
   * Get recent executions (for dashboard)
   */
  async getRecentExecutions(limit: number = 10): Promise<RecentExecutionsResponse> {
    return apiClient.get('/executions/recent', { params: { limit } })
  },
}
