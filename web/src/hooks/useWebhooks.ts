import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { webhookAPI } from '../api/webhooks'
import type {
  WebhookListParams,
  WebhookCreateInput,
  WebhookUpdateInput,
  TestWebhookInput,
} from '../api/webhooks'

/**
 * Hook to fetch and manage list of webhooks
 */
export function useWebhooks(params?: WebhookListParams) {
  const query = useQuery({
    queryKey: ['webhooks', params],
    queryFn: () => webhookAPI.list(params),
    staleTime: 30000, // 30 seconds
  })

  return {
    webhooks: query.data?.webhooks ?? [],
    total: query.data?.total ?? 0,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch a single webhook by ID
 */
export function useWebhook(id: string | null) {
  const query = useQuery({
    queryKey: ['webhook', id],
    queryFn: () => webhookAPI.get(id!),
    enabled: !!id,
    staleTime: 30000, // 30 seconds
  })

  return {
    webhook: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook for webhook event history
 */
export function useWebhookEvents(
  webhookId: string | null,
  params?: { page?: number; limit?: number }
) {
  const query = useQuery({
    queryKey: ['webhook-events', webhookId, params],
    queryFn: () => webhookAPI.getEvents(webhookId!, params),
    enabled: !!webhookId,
    staleTime: 30000, // 30 seconds
  })

  return {
    events: query.data?.events ?? [],
    total: query.data?.total ?? 0,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook for webhook CRUD mutations
 */
export function useWebhookMutations() {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (input: WebhookCreateInput) => webhookAPI.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: WebhookUpdateInput }) =>
      webhookAPI.update(id, updates),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
      queryClient.invalidateQueries({ queryKey: ['webhook', variables.id] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => webhookAPI.delete(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
      queryClient.invalidateQueries({ queryKey: ['webhook', id] })
    },
  })

  const regenerateMutation = useMutation({
    mutationFn: (id: string) => webhookAPI.regenerateSecret(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['webhook', id] })
    },
  })

  const testMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input?: TestWebhookInput }) =>
      webhookAPI.test(id, input || {}),
  })

  return {
    createWebhook: createMutation.mutateAsync,
    updateWebhook: (id: string, updates: WebhookUpdateInput) =>
      updateMutation.mutateAsync({ id, updates }),
    deleteWebhook: deleteMutation.mutateAsync,
    regenerateSecret: regenerateMutation.mutateAsync,
    testWebhook: (id: string, input?: TestWebhookInput) =>
      testMutation.mutateAsync({ id, input }),
    creating: createMutation.isPending,
    updating: updateMutation.isPending,
    deleting: deleteMutation.isPending,
    regenerating: regenerateMutation.isPending,
    testing: testMutation.isPending,
  }
}
