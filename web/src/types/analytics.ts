/**
 * Time range for analytics queries
 */
export interface TimeRange {
  startDate: string
  endDate: string
}

/**
 * Granularity for time series data
 */
export type Granularity = 'hour' | 'day' | 'week' | 'month'

/**
 * Workflow statistics
 */
export interface WorkflowStats {
  workflowId: string
  workflowName: string
  executionCount: number
  successCount: number
  failureCount: number
  successRate: number
  avgDurationMs: number
  minDurationMs: number
  maxDurationMs: number
  lastExecutedAt: string
  totalDurationMs: number
  cancelledCount: number
  pendingCount: number
  runningCount: number
}

/**
 * Tenant overview statistics
 */
export interface TenantOverview {
  totalExecutions: number
  successfulExecutions: number
  failedExecutions: number
  cancelledExecutions: number
  pendingExecutions: number
  runningExecutions: number
  successRate: number
  avgDurationMs: number
  activeWorkflows: number
  totalWorkflows: number
}

/**
 * Time series data point
 */
export interface TimeSeriesPoint {
  timestamp: string
  executionCount: number
  successCount: number
  failureCount: number
  successRate: number
  avgDurationMs: number
}

/**
 * Execution trends over time
 */
export interface ExecutionTrends {
  granularity: Granularity
  startDate: string
  endDate: string
  dataPoints: TimeSeriesPoint[]
}

/**
 * Top workflow item
 */
export interface TopWorkflow {
  workflowId: string
  workflowName: string
  executionCount: number
  successRate: number
  avgDurationMs: number
  lastExecutedAt: string
}

/**
 * Top workflows list
 */
export interface TopWorkflows {
  workflows: TopWorkflow[]
  total: number
}

/**
 * Error information
 */
export interface ErrorInfo {
  errorMessage: string
  errorCount: number
  workflowId: string
  workflowName: string
  lastOccurrence: string
  percentage: number
}

/**
 * Error breakdown analysis
 */
export interface ErrorBreakdown {
  totalErrors: number
  errorsByType: ErrorInfo[]
}

/**
 * Node performance statistics
 */
export interface NodeStats {
  nodeId: string
  nodeType: string
  executionCount: number
  successCount: number
  failureCount: number
  successRate: number
  avgDurationMs: number
  minDurationMs: number
  maxDurationMs: number
}

/**
 * Node-level performance data
 */
export interface NodePerformance {
  workflowId: string
  workflowName: string
  nodes: NodeStats[]
}

/**
 * Analytics query parameters
 */
export interface AnalyticsParams {
  startDate?: string
  endDate?: string
  granularity?: Granularity
  limit?: number
}
