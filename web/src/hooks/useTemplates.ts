import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { templateAPI } from '../api/templates'
import type {
  TemplateListParams,
  CreateTemplateInput,
  UpdateTemplateInput,
  CreateFromWorkflowInput,
  InstantiateTemplateInput,
} from '../api/templates'

/**
 * Hook to fetch and manage list of templates
 */
export function useTemplates(params?: TemplateListParams) {
  const query = useQuery({
    queryKey: ['templates', params],
    queryFn: () => templateAPI.list(params),
    staleTime: 30000, // 30 seconds
  })

  return {
    templates: query.data ?? [],
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch a single template by ID
 */
export function useTemplate(id: string | null) {
  const query = useQuery({
    queryKey: ['template', id],
    queryFn: () => templateAPI.get(id!),
    enabled: !!id,
    staleTime: 30000, // 30 seconds
  })

  return {
    template: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook for template CRUD mutations
 */
export function useTemplateMutations() {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (input: CreateTemplateInput) => templateAPI.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['templates'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: UpdateTemplateInput }) =>
      templateAPI.update(id, updates),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['templates'] })
      queryClient.invalidateQueries({ queryKey: ['template', variables.id] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => templateAPI.delete(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['templates'] })
      queryClient.invalidateQueries({ queryKey: ['template', id] })
    },
  })

  const createFromWorkflowMutation = useMutation({
    mutationFn: ({ workflowId, input }: { workflowId: string; input: CreateFromWorkflowInput }) =>
      templateAPI.createFromWorkflow(workflowId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['templates'] })
    },
  })

  const instantiateMutation = useMutation({
    mutationFn: ({ templateId, input }: { templateId: string; input: InstantiateTemplateInput }) =>
      templateAPI.instantiate(templateId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] })
    },
  })

  return {
    createTemplate: createMutation.mutateAsync,
    updateTemplate: (id: string, updates: UpdateTemplateInput) =>
      updateMutation.mutateAsync({ id, updates }),
    deleteTemplate: deleteMutation.mutateAsync,
    createFromWorkflow: (workflowId: string, input: CreateFromWorkflowInput) =>
      createFromWorkflowMutation.mutateAsync({ workflowId, input }),
    instantiateTemplate: (templateId: string, input: InstantiateTemplateInput) =>
      instantiateMutation.mutateAsync({ templateId, input }),
    creating: createMutation.isPending || createFromWorkflowMutation.isPending,
    updating: updateMutation.isPending,
    deleting: deleteMutation.isPending,
    instantiating: instantiateMutation.isPending,
  }
}
