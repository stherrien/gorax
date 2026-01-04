import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { scheduleAPI } from '../api/schedules'
import type {
  ScheduleListParams,
  ScheduleCreateInput,
  ScheduleUpdateInput,
} from '../api/schedules'

/**
 * Hook to fetch and manage list of schedules
 */
export function useSchedules(params?: ScheduleListParams) {
  const query = useQuery({
    queryKey: ['schedules', params],
    queryFn: () => scheduleAPI.list(params),
    staleTime: 30000, // 30 seconds
  })

  return {
    schedules: query.data?.schedules ?? [],
    total: query.data?.total ?? 0,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook to fetch a single schedule by ID
 */
export function useSchedule(id: string | null) {
  const query = useQuery({
    queryKey: ['schedule', id],
    queryFn: () => scheduleAPI.get(id!),
    enabled: !!id,
    staleTime: 30000, // 30 seconds
  })

  return {
    schedule: query.data ?? null,
    loading: query.isLoading,
    error: query.error as Error | null,
    refetch: query.refetch,
  }
}

/**
 * Hook for schedule CRUD mutations
 */
export function useScheduleMutations() {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: ({ workflowId, input }: { workflowId: string; input: ScheduleCreateInput }) =>
      scheduleAPI.create(workflowId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['schedules'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, updates }: { id: string; updates: ScheduleUpdateInput }) =>
      scheduleAPI.update(id, updates),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['schedules'] })
      queryClient.invalidateQueries({ queryKey: ['schedule', variables.id] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => scheduleAPI.delete(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['schedules'] })
      queryClient.invalidateQueries({ queryKey: ['schedule', id] })
    },
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      scheduleAPI.toggle(id, enabled),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['schedules'] })
      queryClient.invalidateQueries({ queryKey: ['schedule', variables.id] })
    },
  })

  return {
    createSchedule: (workflowId: string, input: ScheduleCreateInput) =>
      createMutation.mutateAsync({ workflowId, input }),
    updateSchedule: (id: string, updates: ScheduleUpdateInput) =>
      updateMutation.mutateAsync({ id, updates }),
    deleteSchedule: deleteMutation.mutateAsync,
    toggleSchedule: (id: string, enabled: boolean) =>
      toggleMutation.mutateAsync({ id, enabled }),
    creating: createMutation.isPending,
    updating: updateMutation.isPending || toggleMutation.isPending,
    deleting: deleteMutation.isPending,
  }
}
