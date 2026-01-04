import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { workflowAPI } from '../api/workflows'
import type {
  WorkflowListParams,
  WorkflowCreateInput,
  WorkflowUpdateInput,
} from '../api/workflows'

/**
 * Hook to fetch and manage list of workflows
 */
export function useWorkflows(params?: WorkflowListParams) {
  const query = useQuery({
    queryKey: ['workflows', params],
    queryFn: () => workflowAPI.list(params),
    staleTime: 30000, // 30 seconds
  })

  return {
    workflows: query.data?.workflows ?? [],
    total: query.data?.total ?? 0,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch a single workflow by ID
 */
export function useWorkflow(id: string | null) {
  const query = useQuery({
    queryKey: ['workflow', id],
    queryFn: () => workflowAPI.get(id!),
    enabled: !!id,
    staleTime: 30000, // 30 seconds
  })

  return {
    workflow: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook for workflow CRUD mutations
 */
export function useWorkflowMutations() {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (input: WorkflowCreateInput) => workflowAPI.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: WorkflowUpdateInput }) =>
      workflowAPI.update(id, updates),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
      queryClient.invalidateQueries({ queryKey: ['workflow', variables.id] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => workflowAPI.delete(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
      queryClient.invalidateQueries({ queryKey: ['workflow', id] })
    },
  })

  const executeMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input?: Record<string, unknown> }) =>
      workflowAPI.execute(id, input),
  })

  return {
    createWorkflow: createMutation.mutateAsync,
    updateWorkflow: (id: string, updates: WorkflowUpdateInput) =>
      updateMutation.mutateAsync({ id, updates }),
    deleteWorkflow: deleteMutation.mutateAsync,
    executeWorkflow: (id: string, input?: Record<string, unknown>) =>
      executeMutation.mutateAsync({ id, input }),
    creating: createMutation.isPending,
    updating: updateMutation.isPending,
    deleting: deleteMutation.isPending,
    executing: executeMutation.isPending,
  }
}
