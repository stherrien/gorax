import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  scheduleTemplatesApi,
  ScheduleTemplateFilter,
  ApplyTemplateInput,
  Schedule,
} from '../api/scheduleTemplates';

/**
 * Hook to fetch schedule templates with optional filters
 */
export function useScheduleTemplates(filter?: ScheduleTemplateFilter) {
  return useQuery({
    queryKey: ['schedule-templates', filter],
    queryFn: () => scheduleTemplatesApi.list(filter),
    staleTime: 5 * 60 * 1000, // 5 minutes - templates don't change often
  });
}

/**
 * Hook to fetch a single schedule template by ID
 */
export function useScheduleTemplate(id: string | undefined) {
  return useQuery({
    queryKey: ['schedule-templates', id],
    queryFn: () => scheduleTemplatesApi.get(id!),
    enabled: !!id,
    staleTime: 5 * 60 * 1000,
  });
}

/**
 * Hook to apply a schedule template
 */
export function useApplyTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      templateId,
      input,
    }: {
      templateId: string;
      input: ApplyTemplateInput;
    }) => scheduleTemplatesApi.apply(templateId, input),
    onSuccess: (data: Schedule) => {
      // Invalidate schedules query to refetch the list
      queryClient.invalidateQueries({ queryKey: ['schedules'] });
      queryClient.invalidateQueries({ queryKey: ['schedules', data.workflow_id] });
    },
  });
}

/**
 * Hook to get schedule templates by category
 */
export function useScheduleTemplatesByCategory(category: string) {
  return useScheduleTemplates({ category });
}

/**
 * Hook to search schedule templates
 */
export function useSearchScheduleTemplates(searchQuery: string) {
  return useScheduleTemplates({ search: searchQuery });
}

/**
 * Hook to get system schedule templates only
 */
export function useSystemScheduleTemplates() {
  return useScheduleTemplates({ is_system: true });
}
