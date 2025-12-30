import { useState, useEffect, useCallback } from 'react'
import { analyticsAPI } from '../api/analytics'
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
 * Hook to fetch tenant overview analytics
 */
export function useTenantOverview(params: AnalyticsParams) {
  const [data, setData] = useState<TenantOverview | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchData = useCallback(async () => {
    if (!params.startDate || !params.endDate) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const result = await analyticsAPI.getTenantOverview(params)
      setData(result)
    } catch (err) {
      setError(err as Error)
      setData(null)
    } finally {
      setLoading(false)
    }
  }, [params.startDate, params.endDate])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  return { data, loading, error, refetch: fetchData }
}

/**
 * Hook to fetch workflow statistics
 */
export function useWorkflowStats(workflowId: string, params: AnalyticsParams) {
  const [data, setData] = useState<WorkflowStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchData = useCallback(async () => {
    if (!workflowId || !params.startDate || !params.endDate) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const result = await analyticsAPI.getWorkflowStats(workflowId, params)
      setData(result)
    } catch (err) {
      setError(err as Error)
      setData(null)
    } finally {
      setLoading(false)
    }
  }, [workflowId, params.startDate, params.endDate])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  return { data, loading, error, refetch: fetchData }
}

/**
 * Hook to fetch execution trends
 */
export function useExecutionTrends(params: AnalyticsParams) {
  const [data, setData] = useState<ExecutionTrends | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchData = useCallback(async () => {
    if (!params.startDate || !params.endDate) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const result = await analyticsAPI.getExecutionTrends(params)
      setData(result)
    } catch (err) {
      setError(err as Error)
      setData(null)
    } finally {
      setLoading(false)
    }
  }, [params.startDate, params.endDate, params.granularity])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  return { data, loading, error, refetch: fetchData }
}

/**
 * Hook to fetch top workflows
 */
export function useTopWorkflows(params: AnalyticsParams) {
  const [data, setData] = useState<TopWorkflows | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchData = useCallback(async () => {
    if (!params.startDate || !params.endDate) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const result = await analyticsAPI.getTopWorkflows(params)
      setData(result)
    } catch (err) {
      setError(err as Error)
      setData(null)
    } finally {
      setLoading(false)
    }
  }, [params.startDate, params.endDate, params.limit])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  return { data, loading, error, refetch: fetchData }
}

/**
 * Hook to fetch error breakdown
 */
export function useErrorBreakdown(params: AnalyticsParams) {
  const [data, setData] = useState<ErrorBreakdown | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchData = useCallback(async () => {
    if (!params.startDate || !params.endDate) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const result = await analyticsAPI.getErrorBreakdown(params)
      setData(result)
    } catch (err) {
      setError(err as Error)
      setData(null)
    } finally {
      setLoading(false)
    }
  }, [params.startDate, params.endDate])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  return { data, loading, error, refetch: fetchData }
}

/**
 * Hook to fetch node performance
 */
export function useNodePerformance(workflowId: string) {
  const [data, setData] = useState<NodePerformance | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchData = useCallback(async () => {
    if (!workflowId) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const result = await analyticsAPI.getNodePerformance(workflowId)
      setData(result)
    } catch (err) {
      setError(err as Error)
      setData(null)
    } finally {
      setLoading(false)
    }
  }, [workflowId])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  return { data, loading, error, refetch: fetchData }
}
