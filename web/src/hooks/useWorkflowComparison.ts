/**
 * useWorkflowComparison - Hook for managing workflow version comparison state
 * Handles version fetching, selection, and diff computation
 */

import { useState, useMemo, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { workflowAPI, type WorkflowVersion } from '../api/workflows'
import type { WorkflowDiff } from '../types/diff'
import { computeWorkflowDiff } from '../components/workflow/comparison/DiffHighlight'

interface UseWorkflowComparisonOptions {
  workflowId: string
  enabled?: boolean
}

interface UseWorkflowComparisonResult {
  versions: WorkflowVersion[]
  loading: boolean
  error: string | null
  baseVersion: WorkflowVersion | null
  compareVersion: WorkflowVersion | null
  diff: WorkflowDiff | null
  setBaseVersionId: (id: string | null) => void
  setCompareVersionId: (id: string | null) => void
  selectVersionPair: (baseId: string, compareId: string) => void
  swapVersions: () => void
  refetch: () => void
}

export function useWorkflowComparison({
  workflowId,
  enabled = true,
}: UseWorkflowComparisonOptions): UseWorkflowComparisonResult {
  const [baseVersionId, setBaseVersionId] = useState<string | null>(null)
  const [compareVersionId, setCompareVersionId] = useState<string | null>(null)

  // Fetch version history
  const {
    data: versions = [],
    isLoading: loading,
    error: queryError,
    refetch,
  } = useQuery({
    queryKey: ['workflowVersions', workflowId],
    queryFn: () => workflowAPI.listVersions(workflowId),
    enabled: enabled && Boolean(workflowId),
    staleTime: 30000,
  })

  const error = queryError instanceof Error ? queryError.message : null

  // Get selected versions
  const baseVersion = useMemo(() => {
    return versions.find((v) => v.id === baseVersionId) || null
  }, [versions, baseVersionId])

  const compareVersion = useMemo(() => {
    return versions.find((v) => v.id === compareVersionId) || null
  }, [versions, compareVersionId])

  // Compute diff when both versions are selected
  const diff = useMemo<WorkflowDiff | null>(() => {
    if (!baseVersion || !compareVersion) return null

    return computeWorkflowDiff(
      baseVersion.definition,
      compareVersion.definition,
      baseVersion.version,
      compareVersion.version
    )
  }, [baseVersion, compareVersion])

  // Select a pair of versions at once
  const selectVersionPair = useCallback((baseId: string, compareId: string) => {
    setBaseVersionId(baseId)
    setCompareVersionId(compareId)
  }, [])

  // Swap base and compare versions
  const swapVersions = useCallback(() => {
    const tempBase = baseVersionId
    setBaseVersionId(compareVersionId)
    setCompareVersionId(tempBase)
  }, [baseVersionId, compareVersionId])

  return {
    versions,
    loading,
    error,
    baseVersion,
    compareVersion,
    diff,
    setBaseVersionId,
    setCompareVersionId,
    selectVersionPair,
    swapVersions,
    refetch,
  }
}

/**
 * Utility hook for comparing specific versions
 */
interface UseVersionComparisonOptions {
  workflowId: string
  baseVersionNumber: number
  compareVersionNumber: number
  enabled?: boolean
}

interface UseVersionComparisonResult {
  baseVersion: WorkflowVersion | null
  compareVersion: WorkflowVersion | null
  diff: WorkflowDiff | null
  loading: boolean
  error: string | null
}

export function useVersionComparison({
  workflowId,
  baseVersionNumber,
  compareVersionNumber,
  enabled = true,
}: UseVersionComparisonOptions): UseVersionComparisonResult {
  // Fetch base version
  const {
    data: baseVersion,
    isLoading: baseLoading,
    error: baseError,
  } = useQuery({
    queryKey: ['workflowVersion', workflowId, baseVersionNumber],
    queryFn: () => workflowAPI.getVersion(workflowId, baseVersionNumber),
    enabled: enabled && Boolean(workflowId) && baseVersionNumber > 0,
    staleTime: 60000,
  })

  // Fetch compare version
  const {
    data: compareVersion,
    isLoading: compareLoading,
    error: compareError,
  } = useQuery({
    queryKey: ['workflowVersion', workflowId, compareVersionNumber],
    queryFn: () => workflowAPI.getVersion(workflowId, compareVersionNumber),
    enabled: enabled && Boolean(workflowId) && compareVersionNumber > 0,
    staleTime: 60000,
  })

  const loading = baseLoading || compareLoading
  const error = baseError instanceof Error
    ? baseError.message
    : compareError instanceof Error
    ? compareError.message
    : null

  // Compute diff
  const diff = useMemo<WorkflowDiff | null>(() => {
    if (!baseVersion || !compareVersion) return null

    return computeWorkflowDiff(
      baseVersion.definition,
      compareVersion.definition,
      baseVersion.version,
      compareVersion.version
    )
  }, [baseVersion, compareVersion])

  return {
    baseVersion: baseVersion || null,
    compareVersion: compareVersion || null,
    diff,
    loading,
    error,
  }
}
