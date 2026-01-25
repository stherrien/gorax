import { useQuery } from '@tanstack/react-query'
import { monitoringAPI, executionLogsAPI } from '../api/management'
import type { ExecutionLogListParams, LogLevel } from '../types/management'
import { isValidResourceId } from '../utils/routing'

/**
 * Hook to fetch monitoring statistics with auto-refresh
 */
export function useMonitoringStats(options?: { refetchInterval?: number }) {
  const query = useQuery({
    queryKey: ['monitoring-stats'],
    queryFn: () => monitoringAPI.getStats(),
    staleTime: 10000, // 10 seconds
    refetchInterval: options?.refetchInterval ?? 30000, // Auto-refresh every 30 seconds
  })

  return {
    stats: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch system health status
 */
export function useSystemHealth(options?: { refetchInterval?: number }) {
  const query = useQuery({
    queryKey: ['system-health'],
    queryFn: () => monitoringAPI.getHealth(),
    staleTime: 30000, // 30 seconds
    refetchInterval: options?.refetchInterval ?? 60000, // Auto-refresh every minute
  })

  return {
    health: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch workflow-specific statistics
 */
export function useWorkflowStats(workflowId?: string) {
  const query = useQuery({
    queryKey: ['workflow-stats', workflowId],
    queryFn: () => monitoringAPI.getWorkflowStats(workflowId),
    staleTime: 30000, // 30 seconds
  })

  return {
    workflowStats: query.data ?? [],
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch active executions for real-time monitoring
 */
export function useActiveExecutions(options?: { refetchInterval?: number }) {
  const query = useQuery({
    queryKey: ['active-executions'],
    queryFn: () => monitoringAPI.getActiveExecutions(),
    staleTime: 5000, // 5 seconds
    refetchInterval: options?.refetchInterval ?? 10000, // Auto-refresh every 10 seconds
  })

  return {
    activeExecutions: query.data ?? [],
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch execution logs
 */
export function useExecutionLogs(
  executionId: string | null,
  options?: {
    nodeId?: string
    level?: LogLevel | LogLevel[]
    search?: string
    limit?: number
    offset?: number
  }
) {
  const params: ExecutionLogListParams | null = executionId
    ? {
        executionId,
        nodeId: options?.nodeId,
        level: options?.level,
        search: options?.search,
        limit: options?.limit ?? 100,
        offset: options?.offset ?? 0,
      }
    : null

  const query = useQuery({
    queryKey: ['execution-logs', params],
    queryFn: () => executionLogsAPI.list(params!),
    enabled: isValidResourceId(executionId),
    staleTime: 30000, // 30 seconds
  })

  return {
    logs: query.data?.logs ?? [],
    total: query.data?.total ?? 0,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook for searching logs across executions
 */
export function useLogSearch(
  query: string,
  options?: { limit?: number; offset?: number }
) {
  const searchQuery = useQuery({
    queryKey: ['log-search', query, options],
    queryFn: () => executionLogsAPI.search(query, options),
    enabled: query.length >= 3, // Only search with at least 3 characters
    staleTime: 30000, // 30 seconds
  })

  return {
    logs: searchQuery.data?.logs ?? [],
    total: searchQuery.data?.total ?? 0,
    loading: searchQuery.isLoading,
    error: searchQuery.error as Error | null,
  }
}
