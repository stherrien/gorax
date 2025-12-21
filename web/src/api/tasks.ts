import { apiClient } from './client'

export interface HumanTask {
  id: string
  tenant_id: string
  execution_id: string
  step_id: string
  task_type: 'approval' | 'input' | 'review'
  title: string
  description: string
  assignees_list: string[]
  status: 'pending' | 'approved' | 'rejected' | 'expired' | 'cancelled'
  due_date?: string
  completed_at?: string
  completed_by?: string
  response_map?: Record<string, unknown>
  config_data?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export interface ListTasksParams {
  status?: string
  task_type?: string
  assignee?: string
  execution_id?: string
  limit?: number
  offset?: number
}

export interface ApproveTaskRequest {
  comment?: string
  data?: Record<string, unknown>
}

export interface RejectTaskRequest {
  reason?: string
  data?: Record<string, unknown>
}

export interface SubmitTaskRequest {
  data: Record<string, unknown>
}

export interface ListTasksResponse {
  tasks: HumanTask[]
  count: number
}

export const tasksApi = {
  async list(params?: ListTasksParams): Promise<ListTasksResponse> {
    return apiClient.get('/api/v1/tasks', { params })
  },

  async get(id: string): Promise<HumanTask> {
    return apiClient.get(`/api/v1/tasks/${id}`)
  },

  async approve(id: string, request?: ApproveTaskRequest): Promise<HumanTask> {
    return apiClient.post(`/api/v1/tasks/${id}/approve`, request || {})
  },

  async reject(id: string, request?: RejectTaskRequest): Promise<HumanTask> {
    return apiClient.post(`/api/v1/tasks/${id}/reject`, request || {})
  },

  async submit(id: string, request: SubmitTaskRequest): Promise<HumanTask> {
    return apiClient.post(`/api/v1/tasks/${id}/submit`, request)
  },
}
