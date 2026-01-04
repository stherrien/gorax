import { useQuery } from '@tanstack/react-query'
import { executionAPI } from '../api/executions'
import type {
  ExecutionListParams,
  DashboardStatsParams,
} from '../api/executions'
import { isValidResourceId } from '../utils/routing'

/**
 * Hook to fetch and manage list of executions
 */
export function useExecutions(params?: ExecutionListParams) {
  const query = useQuery({
    queryKey: ['executions', params],
    queryFn: () => executionAPI.list(params),
    staleTime: 30000, // 30 seconds
  })

  return {
    executions: query.data?.executions ?? [],
    total: query.data?.total ?? 0,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch a single execution by ID
 */
export function useExecution(id: string | null) {
  const query = useQuery({
    queryKey: ['execution', id],
    queryFn: () => executionAPI.get(id!),
    enabled: isValidResourceId(id),
    staleTime: 30000, // 30 seconds
  })

  return {
    execution: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch dashboard statistics
 */
export function useDashboardStats(params?: DashboardStatsParams) {
  const query = useQuery({
    queryKey: ['dashboard-stats', params],
    queryFn: () => executionAPI.getDashboardStats(params),
    staleTime: 30000, // 30 seconds
  })

  return {
    stats: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch recent executions (for dashboard)
 */
export function useRecentExecutions(limit: number = 10) {
  const query = useQuery({
    queryKey: ['recent-executions', limit],
    queryFn: () => executionAPI.getRecentExecutions(limit),
    staleTime: 30000, // 30 seconds
  })

  return {
    executions: query.data?.executions ?? [],
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}
