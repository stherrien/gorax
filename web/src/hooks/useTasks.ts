import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  tasksApi,
  HumanTask,
  ListTasksParams,
  ApproveTaskRequest,
  RejectTaskRequest,
  SubmitTaskRequest,
} from '../api/tasks';

/**
 * Hook to fetch list of tasks
 */
export function useTasks(params?: ListTasksParams) {
  return useQuery({
    queryKey: ['tasks', params],
    queryFn: () => tasksApi.list(params),
    staleTime: 30000, // 30 seconds
  });
}

/**
 * Hook to fetch a single task
 */
export function useTask(taskId: string | undefined) {
  return useQuery({
    queryKey: ['tasks', taskId],
    queryFn: () => tasksApi.get(taskId!),
    enabled: !!taskId,
  });
}

/**
 * Hook to fetch pending task count
 */
export function usePendingTaskCount() {
  return useQuery({
    queryKey: ['tasks', 'pending-count'],
    queryFn: () => tasksApi.getPendingCount(),
    staleTime: 60000, // 1 minute
    refetchInterval: 60000, // Refetch every minute
  });
}

/**
 * Hook to approve a task
 */
export function useApproveTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ taskId, request }: { taskId: string; request: ApproveTaskRequest }) =>
      tasksApi.approve(taskId, request),
    onSuccess: () => {
      // Invalidate and refetch tasks
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });
}

/**
 * Hook to reject a task
 */
export function useRejectTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ taskId, request }: { taskId: string; request: RejectTaskRequest }) =>
      tasksApi.reject(taskId, request),
    onSuccess: () => {
      // Invalidate and refetch tasks
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });
}

/**
 * Hook to submit an input task
 */
export function useSubmitTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ taskId, request }: { taskId: string; request: SubmitTaskRequest }) =>
      tasksApi.submit(taskId, request),
    onSuccess: () => {
      // Invalidate and refetch tasks
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });
}

/**
 * Hook to get tasks assigned to current user
 */
export function useMyTasks(status?: string) {
  return useTasks({
    status,
    limit: 50,
  });
}

/**
 * Hook to get overdue tasks
 */
export function useOverdueTasks() {
  return useQuery({
    queryKey: ['tasks', 'overdue'],
    queryFn: async () => {
      const result = await tasksApi.list({ status: 'pending' });
      const now = new Date();
      return result.tasks.filter(
        (task) => task.due_date && new Date(task.due_date) < now
      );
    },
    staleTime: 60000,
  });
}
