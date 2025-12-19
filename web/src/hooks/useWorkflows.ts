import { useState, useEffect, useCallback } from 'react'
import { workflowAPI } from '../api/workflows'
import type {
  Workflow,
  WorkflowListParams,
  WorkflowCreateInput,
  WorkflowUpdateInput,
  WorkflowExecutionResponse,
} from '../api/workflows'

/**
 * Hook to fetch and manage list of workflows
 */
export function useWorkflows(params?: WorkflowListParams) {
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const fetchWorkflows = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await workflowAPI.list(params)
      setWorkflows(response.workflows)
      setTotal(response.total)
    } catch (err) {
      setError(err as Error)
      setWorkflows([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [params])

  useEffect(() => {
    fetchWorkflows()
  }, [fetchWorkflows])

  return {
    workflows,
    total,
    loading,
    error,
    refetch: fetchWorkflows,
  }
}

/**
 * Hook to fetch a single workflow by ID
 */
export function useWorkflow(id: string | null) {
  const [workflow, setWorkflow] = useState<Workflow | null>(null)
  const [loading, setLoading] = useState(!!id)
  const [error, setError] = useState<Error | null>(null)

  const fetchWorkflow = useCallback(async () => {
    if (!id) {
      setLoading(false)
      return
    }

    try {
      setLoading(true)
      setError(null)
      const data = await workflowAPI.get(id)
      setWorkflow(data)
    } catch (err) {
      setError(err as Error)
      setWorkflow(null)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchWorkflow()
  }, [fetchWorkflow])

  return {
    workflow,
    loading,
    error,
    refetch: fetchWorkflow,
  }
}

/**
 * Hook for workflow CRUD mutations
 */
export function useWorkflowMutations() {
  const [creating, setCreating] = useState(false)
  const [updating, setUpdating] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [executing, setExecuting] = useState(false)

  const createWorkflow = async (input: WorkflowCreateInput): Promise<Workflow> => {
    try {
      setCreating(true)
      const workflow = await workflowAPI.create(input)
      return workflow
    } finally {
      setCreating(false)
    }
  }

  const updateWorkflow = async (
    id: string,
    updates: WorkflowUpdateInput
  ): Promise<Workflow> => {
    try {
      setUpdating(true)
      const workflow = await workflowAPI.update(id, updates)
      return workflow
    } finally {
      setUpdating(false)
    }
  }

  const deleteWorkflow = async (id: string): Promise<void> => {
    try {
      setDeleting(true)
      await workflowAPI.delete(id)
    } finally {
      setDeleting(false)
    }
  }

  const executeWorkflow = async (
    id: string,
    input?: Record<string, unknown>
  ): Promise<WorkflowExecutionResponse> => {
    try {
      setExecuting(true)
      const response = await workflowAPI.execute(id, input)
      return response
    } finally {
      setExecuting(false)
    }
  }

  return {
    createWorkflow,
    updateWorkflow,
    deleteWorkflow,
    executeWorkflow,
    creating,
    updating,
    deleting,
    executing,
  }
}
