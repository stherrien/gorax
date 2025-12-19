import { apiClient } from './client'

// Schedule types
export interface Schedule {
  id: string
  tenantId: string
  workflowId: string
  name: string
  cronExpression: string
  timezone: string
  enabled: boolean
  nextRunAt?: string
  lastRunAt?: string
  lastExecutionId?: string
  createdBy: string
  createdAt: string
  updatedAt: string
}

export interface ScheduleWithWorkflow extends Schedule {
  workflowName: string
  workflowStatus: string
}

export interface ScheduleListResponse {
  schedules: Schedule[]
  total: number
}

export interface ScheduleListParams {
  page?: number
  limit?: number
  workflowId?: string
  enabled?: boolean
  search?: string
}

export interface ScheduleCreateInput {
  name: string
  cronExpression: string
  timezone?: string
  enabled?: boolean
}

export interface ScheduleUpdateInput {
  name?: string
  cronExpression?: string
  timezone?: string
  enabled?: boolean
}

export interface ParseCronResponse {
  valid: boolean
  next_run?: string
}

export interface PreviewScheduleResponse {
  valid: boolean
  next_runs: string[]
  count: number
  timezone: string
}

class ScheduleAPI {
  /**
   * List all schedules
   */
  async list(params?: ScheduleListParams): Promise<ScheduleListResponse> {
    const options = params ? { params } : undefined
    const response = await apiClient.get('/api/v1/schedules', options)
    // Backend returns { data: [], limit, offset }
    if (response.data && Array.isArray(response.data)) {
      return { schedules: response.data, total: response.data.length }
    }
    return response
  }

  /**
   * Get a single schedule by ID
   */
  async get(id: string): Promise<Schedule> {
    const response = await apiClient.get(`/api/v1/schedules/${id}`)
    return response.data || response
  }

  /**
   * Create a new schedule for a workflow
   */
  async create(workflowId: string, schedule: ScheduleCreateInput): Promise<Schedule> {
    const response = await apiClient.post(
      `/api/v1/workflows/${workflowId}/schedules`,
      schedule
    )
    return response.data || response
  }

  /**
   * Update an existing schedule
   */
  async update(id: string, updates: ScheduleUpdateInput): Promise<Schedule> {
    const response = await apiClient.put(`/api/v1/schedules/${id}`, updates)
    return response.data || response
  }

  /**
   * Delete a schedule
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/api/v1/schedules/${id}`)
  }

  /**
   * Toggle schedule enabled state
   */
  async toggle(id: string, enabled: boolean): Promise<Schedule> {
    return await this.update(id, { enabled })
  }

  /**
   * Parse and validate a cron expression
   */
  async parseCron(cronExpression: string, timezone: string = 'UTC'): Promise<ParseCronResponse> {
    const response = await apiClient.post('/api/v1/schedules/parse-cron', {
      cron_expression: cronExpression,
      timezone,
    })
    return response.data || response
  }

  /**
   * Preview schedule execution times
   */
  async preview(
    cronExpression: string,
    timezone: string = 'UTC',
    count: number = 10
  ): Promise<PreviewScheduleResponse> {
    const response = await apiClient.post('/api/v1/schedules/preview', {
      cron_expression: cronExpression,
      timezone,
      count,
    })
    return response.data || response
  }

  /**
   * Bulk update schedules (enable/disable/delete)
   */
  async bulkUpdate(ids: string[], action: 'enable' | 'disable' | 'delete'): Promise<BulkOperationResult> {
    const response = await apiClient.patch('/api/v1/schedules/bulk', {
      ids,
      action,
    })
    return response.data || response
  }

  /**
   * Export schedules as JSON
   */
  async export(ids?: string[]): Promise<Schedule[]> {
    const params = ids && ids.length > 0 ? { ids: ids.join(',') } : undefined
    const response = await apiClient.get('/api/v1/schedules/export', { params })
    return response.data || response
  }
}

export interface BulkOperationResult {
  success: string[]
  failed: BulkOperationError[]
}

export interface BulkOperationError {
  id: string
  error: string
}

export const scheduleAPI = new ScheduleAPI()
