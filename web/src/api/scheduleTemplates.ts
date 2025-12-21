import { apiClient } from './client';

export interface ScheduleTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  cron_expression: string;
  timezone: string;
  tags: string[];
  is_system: boolean;
  created_at: string;
}

export interface ScheduleTemplateFilter {
  category?: string;
  tags?: string[];
  is_system?: boolean;
  search?: string;
}

export interface ApplyTemplateInput {
  workflow_id: string;
  name?: string;
  timezone?: string;
}

export interface Schedule {
  id: string;
  tenant_id: string;
  workflow_id: string;
  name: string;
  cron_expression: string;
  timezone: string;
  enabled: boolean;
  next_run_at?: string;
  last_run_at?: string;
  last_execution_id?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export const scheduleTemplatesApi = {
  /**
   * List all schedule templates with optional filters
   */
  list: async (filter?: ScheduleTemplateFilter): Promise<ScheduleTemplate[]> => {
    const params = new URLSearchParams();

    if (filter?.category) {
      params.append('category', filter.category);
    }

    if (filter?.tags && filter.tags.length > 0) {
      params.append('tags', filter.tags.join(','));
    }

    if (filter?.is_system !== undefined) {
      params.append('is_system', filter.is_system.toString());
    }

    if (filter?.search) {
      params.append('search', filter.search);
    }

    const queryString = params.toString();
    const url = `/api/v1/schedule-templates${queryString ? `?${queryString}` : ''}`;

    const response = await apiClient.get(url);
    return response as ScheduleTemplate[];
  },

  /**
   * Get a single schedule template by ID
   */
  get: async (id: string): Promise<ScheduleTemplate> => {
    const response = await apiClient.get(`/api/v1/schedule-templates/${id}`);
    return response as ScheduleTemplate;
  },

  /**
   * Apply a template to create a schedule
   */
  apply: async (templateId: string, input: ApplyTemplateInput): Promise<Schedule> => {
    const response = await apiClient.post(
      `/api/v1/schedule-templates/${templateId}/apply`,
      input
    );
    return response as Schedule;
  },
};
