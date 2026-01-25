import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userAPI } from '../api/management'
import type {
  UserListParams,
  UserCreateInput,
  UserUpdateInput,
  UserRole,
} from '../types/management'
import { isValidResourceId } from '../utils/routing'

/**
 * Hook to fetch and manage list of users
 */
export function useUsers(params?: UserListParams) {
  const query = useQuery({
    queryKey: ['users', params],
    queryFn: () => userAPI.list(params),
    staleTime: 30000, // 30 seconds
  })

  return {
    users: query.data?.users ?? [],
    total: query.data?.total ?? 0,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch a single user by ID
 */
export function useUser(id: string | null) {
  const query = useQuery({
    queryKey: ['user', id],
    queryFn: () => userAPI.get(id!),
    enabled: isValidResourceId(id),
    staleTime: 30000, // 30 seconds
  })

  return {
    user: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook for user CRUD mutations
 */
export function useUserMutations() {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (input: UserCreateInput) => userAPI.create(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: UserUpdateInput }) =>
      userAPI.update(id, updates),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      queryClient.invalidateQueries({ queryKey: ['user', variables.id] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => userAPI.delete(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      queryClient.invalidateQueries({ queryKey: ['user', id] })
    },
  })

  const resendInviteMutation = useMutation({
    mutationFn: (id: string) => userAPI.resendInvite(id),
  })

  const bulkDeleteMutation = useMutation({
    mutationFn: (ids: string[]) => userAPI.bulkDelete(ids),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  const bulkUpdateRoleMutation = useMutation({
    mutationFn: ({ ids, role }: { ids: string[]; role: UserRole }) =>
      userAPI.bulkUpdateRole(ids, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  return {
    createUser: createMutation.mutateAsync,
    updateUser: (id: string, updates: UserUpdateInput) =>
      updateMutation.mutateAsync({ id, updates }),
    deleteUser: deleteMutation.mutateAsync,
    resendInvite: resendInviteMutation.mutateAsync,
    bulkDeleteUsers: bulkDeleteMutation.mutateAsync,
    bulkUpdateRole: (ids: string[], role: UserRole) =>
      bulkUpdateRoleMutation.mutateAsync({ ids, role }),
    creating: createMutation.isPending,
    updating: updateMutation.isPending,
    deleting: deleteMutation.isPending,
    sendingInvite: resendInviteMutation.isPending,
    bulkDeleting: bulkDeleteMutation.isPending,
    bulkUpdatingRole: bulkUpdateRoleMutation.isPending,
  }
}
