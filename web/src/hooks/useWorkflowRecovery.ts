import { useState, useCallback, useEffect, useRef } from 'react'
import type { Node, Edge } from '@xyflow/react'

/**
 * Workflow state snapshot for recovery
 */
interface WorkflowSnapshot {
  id: string
  timestamp: number
  nodes: Node[]
  edges: Edge[]
  name: string
  description: string
}

/**
 * Recovery state and actions
 */
interface WorkflowRecoveryState {
  hasBackup: boolean
  lastBackupTime: Date | null
  isRecovering: boolean
  recoveryError: string | null
}

/**
 * Recovery hook return type
 */
interface UseWorkflowRecoveryReturn extends WorkflowRecoveryState {
  saveBackup: (workflowId: string, nodes: Node[], edges: Edge[], name: string, description: string) => void
  recoverFromBackup: (workflowId: string) => WorkflowSnapshot | null
  clearBackup: (workflowId: string) => void
  getBackup: (workflowId: string) => WorkflowSnapshot | null
  getAllBackups: () => WorkflowSnapshot[]
}

const STORAGE_KEY_PREFIX = 'gorax_workflow_backup_'
const MAX_BACKUPS = 5
const AUTO_SAVE_INTERVAL = 30000 // 30 seconds

/**
 * Generate storage key for a workflow
 */
function getStorageKey(workflowId: string): string {
  return `${STORAGE_KEY_PREFIX}${workflowId}`
}

/**
 * Hook for workflow state recovery
 *
 * Provides automatic backup and recovery functionality for workflow data.
 * Saves workflow state to localStorage periodically and allows recovery after errors.
 */
export function useWorkflowRecovery(
  workflowId: string | null,
  options?: {
    autoSave?: boolean
    autoSaveInterval?: number
  }
): UseWorkflowRecoveryReturn {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const { autoSave: _autoSave = true, autoSaveInterval: _autoSaveInterval = AUTO_SAVE_INTERVAL } = options ?? {}

  const [state, setState] = useState<WorkflowRecoveryState>({
    hasBackup: false,
    lastBackupTime: null,
    isRecovering: false,
    recoveryError: null,
  })

  const lastSaveRef = useRef<number>(0)

  // Check for existing backup on mount
  useEffect(() => {
    if (workflowId) {
      const backup = getBackup(workflowId)
      if (backup) {
        setState((prev) => ({
          ...prev,
          hasBackup: true,
          lastBackupTime: new Date(backup.timestamp),
        }))
      }
    }
  }, [workflowId])

  /**
   * Save a workflow backup to localStorage
   */
  const saveBackup = useCallback(
    (
      wfId: string,
      nodes: Node[],
      edges: Edge[],
      name: string,
      description: string
    ): void => {
      // Debounce saves
      const now = Date.now()
      if (now - lastSaveRef.current < 1000) {
        return
      }
      lastSaveRef.current = now

      try {
        const snapshot: WorkflowSnapshot = {
          id: wfId,
          timestamp: now,
          nodes: nodes.map((node) => ({
            ...node,
            // Remove non-serializable properties
            selected: undefined,
            dragging: undefined,
          })),
          edges: edges.map((edge) => ({
            ...edge,
            selected: undefined,
          })),
          name,
          description,
        }

        localStorage.setItem(getStorageKey(wfId), JSON.stringify(snapshot))

        setState((prev) => ({
          ...prev,
          hasBackup: true,
          lastBackupTime: new Date(now),
          recoveryError: null,
        }))

        // Cleanup old backups
        cleanupOldBackups()
      } catch (error) {
        console.error('[useWorkflowRecovery] Failed to save backup:', error)
      }
    },
    []
  )

  /**
   * Get a workflow backup from localStorage
   */
  const getBackup = useCallback((wfId: string): WorkflowSnapshot | null => {
    try {
      const stored = localStorage.getItem(getStorageKey(wfId))
      if (!stored) return null

      const snapshot = JSON.parse(stored) as WorkflowSnapshot

      // Validate snapshot structure
      if (!snapshot.id || !snapshot.nodes || !snapshot.edges) {
        return null
      }

      return snapshot
    } catch (error) {
      console.error('[useWorkflowRecovery] Failed to get backup:', error)
      return null
    }
  }, [])

  /**
   * Recover workflow from backup
   */
  const recoverFromBackup = useCallback(
    (wfId: string): WorkflowSnapshot | null => {
      setState((prev) => ({ ...prev, isRecovering: true, recoveryError: null }))

      try {
        const snapshot = getBackup(wfId)

        if (!snapshot) {
          setState((prev) => ({
            ...prev,
            isRecovering: false,
            recoveryError: 'No backup found for this workflow',
          }))
          return null
        }

        setState((prev) => ({
          ...prev,
          isRecovering: false,
          recoveryError: null,
        }))

        return snapshot
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown recovery error'
        setState((prev) => ({
          ...prev,
          isRecovering: false,
          recoveryError: errorMessage,
        }))
        return null
      }
    },
    [getBackup]
  )

  /**
   * Clear a workflow backup
   */
  const clearBackup = useCallback((wfId: string): void => {
    try {
      localStorage.removeItem(getStorageKey(wfId))
      setState((prev) => ({
        ...prev,
        hasBackup: false,
        lastBackupTime: null,
      }))
    } catch (error) {
      console.error('[useWorkflowRecovery] Failed to clear backup:', error)
    }
  }, [])

  /**
   * Get all workflow backups
   */
  const getAllBackups = useCallback((): WorkflowSnapshot[] => {
    const backups: WorkflowSnapshot[] = []

    try {
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i)
        if (key?.startsWith(STORAGE_KEY_PREFIX)) {
          const stored = localStorage.getItem(key)
          if (stored) {
            const snapshot = JSON.parse(stored) as WorkflowSnapshot
            if (snapshot.id && snapshot.nodes && snapshot.edges) {
              backups.push(snapshot)
            }
          }
        }
      }

      // Sort by timestamp, newest first
      backups.sort((a, b) => b.timestamp - a.timestamp)
    } catch (error) {
      console.error('[useWorkflowRecovery] Failed to get all backups:', error)
    }

    return backups
  }, [])

  /**
   * Cleanup old backups to prevent localStorage from filling up
   */
  const cleanupOldBackups = useCallback((): void => {
    try {
      const backups = getAllBackups()

      // Keep only the most recent backups
      if (backups.length > MAX_BACKUPS) {
        const toRemove = backups.slice(MAX_BACKUPS)
        for (const backup of toRemove) {
          localStorage.removeItem(getStorageKey(backup.id))
        }
      }
    } catch (error) {
      console.error('[useWorkflowRecovery] Failed to cleanup backups:', error)
    }
  }, [getAllBackups])

  return {
    ...state,
    saveBackup,
    recoverFromBackup,
    clearBackup,
    getBackup,
    getAllBackups,
  }
}

/**
 * Hook for auto-saving workflow state
 */
export function useAutoSaveWorkflow(
  workflowId: string | null,
  nodes: Node[],
  edges: Edge[],
  name: string,
  description: string,
  options?: {
    enabled?: boolean
    interval?: number
  }
): { lastSaved: Date | null; isSaving: boolean } {
  const { enabled = true, interval = AUTO_SAVE_INTERVAL } = options ?? {}
  const { saveBackup } = useWorkflowRecovery(workflowId)
  const [lastSaved, setLastSaved] = useState<Date | null>(null)
  const [isSaving, setIsSaving] = useState(false)

  // Track if data has changed
  const dataRef = useRef({ nodes, edges, name, description })
  const hasChangedRef = useRef(false)

  useEffect(() => {
    const hasChanged =
      JSON.stringify(dataRef.current.nodes) !== JSON.stringify(nodes) ||
      JSON.stringify(dataRef.current.edges) !== JSON.stringify(edges) ||
      dataRef.current.name !== name ||
      dataRef.current.description !== description

    if (hasChanged) {
      hasChangedRef.current = true
      dataRef.current = { nodes, edges, name, description }
    }
  }, [nodes, edges, name, description])

  // Auto-save on interval
  useEffect(() => {
    if (!enabled || !workflowId) return

    const saveInterval = setInterval(() => {
      if (hasChangedRef.current && nodes.length > 0) {
        setIsSaving(true)
        saveBackup(workflowId, nodes, edges, name, description)
        setLastSaved(new Date())
        hasChangedRef.current = false
        setIsSaving(false)
      }
    }, interval)

    return () => clearInterval(saveInterval)
  }, [enabled, workflowId, nodes, edges, name, description, interval, saveBackup])

  // Save on unmount if there are unsaved changes
  useEffect(() => {
    return () => {
      if (hasChangedRef.current && workflowId && nodes.length > 0) {
        saveBackup(workflowId, nodes, edges, name, description)
      }
    }
  }, [workflowId]) // eslint-disable-line react-hooks/exhaustive-deps

  return { lastSaved, isSaving }
}

export default useWorkflowRecovery
