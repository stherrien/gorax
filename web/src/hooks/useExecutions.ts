import { useState, useEffect, useCallback } from 'react'
import { executionAPI } from '../api/executions'
import type {
  Execution,
  ExecutionListParams,
  DashboardStats,
  DashboardStatsParams,
} from '../api/executions'

/**
 * Hook to fetch and manage list of executions
 */
export function useExecutions(params?: ExecutionListParams) {
  const [executions, setExecutions] = useState<Execution[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchExecutions = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await executionAPI.list(params)
      setExecutions(response.executions)
      setTotal(response.total)
    } catch (err) {
      setError(err as Error)
      setExecutions([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchExecutions()
  }, [fetchExecutions])

  return {
    executions,
    total,
    loading,
    error,
    refetch: fetchExecutions,
  }
}

/**
 * Hook to fetch a single execution by ID
 */
export function useExecution(id: string | null) {
  const [execution, setExecution] = useState<Execution | null>(null)
  const [loading, setLoading] = useState(!!id)
  const [error, setError] = useState<Error | null>(null)

  const fetchExecution = useCallback(async () => {
    if (!id) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const data = await executionAPI.get(id)
      setExecution(data)
    } catch (err) {
      setError(err as Error)
      setExecution(null)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchExecution()
  }, [fetchExecution])

  return {
    execution,
    loading,
    error,
    refetch: fetchExecution,
  }
}

/**
 * Hook to fetch dashboard statistics
 */
export function useDashboardStats(params?: DashboardStatsParams) {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchStats = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await executionAPI.getDashboardStats(params)
      setStats(data)
    } catch (err) {
      setError(err as Error)
      setStats(null)
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchStats()
  }, [fetchStats])

  return {
    stats,
    loading,
    error,
    refetch: fetchStats,
  }
}

/**
 * Hook to fetch recent executions (for dashboard)
 */
export function useRecentExecutions(limit: number = 10) {
  const [executions, setExecutions] = useState<Execution[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchRecentExecutions = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await executionAPI.getRecentExecutions(limit)
      setExecutions(response.executions || [])
    } catch (err) {
      setError(err as Error)
      setExecutions([])
    } finally {
      setLoading(false)
    }
  }, [limit])

  useEffect(() => {
    fetchRecentExecutions()
  }, [fetchRecentExecutions])

  return {
    executions,
    loading,
    error,
    refetch: fetchRecentExecutions,
  }
}
