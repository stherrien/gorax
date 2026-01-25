import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { tenantAPI } from '../api/management'
import type {
  TenantListParams,
  TenantCreateInput,
  TenantUpdateInput,
} from '../types/management'
import { isValidResourceId } from '../utils/routing'

/**
 * Hook to fetch and manage list of tenants
 */
export function useTenants(params?: TenantListParams) {
  const query = useQuery({
    queryKey: ['tenants', params],
    queryFn: () => tenantAPI.list(params),
    staleTime: 30000, // 30 seconds
  })

  return {
    tenants: query.data?.tenants ?? [],
    total: query.data?.total ?? 0,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch a single tenant by ID
 */
export function useTenant(id: string | null) {
  const query = useQuery({
    queryKey: ['tenant', id],
    queryFn: () => tenantAPI.get(id!),
    enabled: isValidResourceId(id),
    staleTime: 30000, // 30 seconds
  })

  return {
    tenant: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch tenant usage statistics
 */
export function useTenantUsage(id: string | null) {
  const query = useQuery({
    queryKey: ['tenant-usage', id],
    queryFn: () => tenantAPI.getUsage(id!),
    enabled: isValidResourceId(id),
    staleTime: 60000, // 1 minute
  })

  return {
    usage: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook for tenant CRUD mutations
 */
export function useTenantMutations() {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (input: TenantCreateInput) => tenantAPI.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: TenantUpdateInput }) =>
      tenantAPI.update(id, updates),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      queryClient.invalidateQueries({ queryKey: ['tenant', variables.id] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => tenantAPI.delete(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      queryClient.invalidateQueries({ queryKey: ['tenant', id] })
    },
  })

  const suspendMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason?: string }) =>
      tenantAPI.suspend(id, reason),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      queryClient.invalidateQueries({ queryKey: ['tenant', variables.id] })
    },
  })

  const reactivateMutation = useMutation({
    mutationFn: (id: string) => tenantAPI.reactivate(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      queryClient.invalidateQueries({ queryKey: ['tenant', id] })
    },
  })

  return {
    createTenant: createMutation.mutateAsync,
    updateTenant: (id: string, updates: TenantUpdateInput) =>
      updateMutation.mutateAsync({ id, updates }),
    deleteTenant: deleteMutation.mutateAsync,
    suspendTenant: (id: string, reason?: string) =>
      suspendMutation.mutateAsync({ id, reason }),
    reactivateTenant: reactivateMutation.mutateAsync,
    creating: createMutation.isPending,
    updating: updateMutation.isPending,
    deleting: deleteMutation.isPending,
    suspending: suspendMutation.isPending,
    reactivating: reactivateMutation.isPending,
  }
}
