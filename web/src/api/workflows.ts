import { apiClient } from './client'

// Workflow types
export interface WorkflowNode {
  id: string
  type: string
  position: { x: number; y: number }
  data: Record<string, unknown>
}

export interface WorkflowEdge {
  id: string
  source: string
  target: string
  sourceHandle?: string
  targetHandle?: string
}

export interface WorkflowDefinition {
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
}

export type WorkflowStatus = 'draft' | 'active' | 'inactive'

export interface Workflow {
  id: string
  tenantId: string
  name: string
  description?: string
  status: WorkflowStatus
  definition: WorkflowDefinition
  version: number
  createdAt: string
  updatedAt: string
}

export interface WorkflowListResponse {
  workflows: Workflow[]
  total: number
}

export interface WorkflowListParams {
  page?: number
  limit?: number
  status?: WorkflowStatus
  search?: string
}

export interface WorkflowCreateInput {
  name: string
  description?: string
  definition: WorkflowDefinition
  status?: WorkflowStatus
}

export interface WorkflowUpdateInput {
  name?: string
  description?: string
  definition?: WorkflowDefinition
  status?: WorkflowStatus
}

export interface WorkflowExecutionResponse {
  executionId: string
  workflowId: string
  status: string
  queuedAt: string
}

class WorkflowAPI {
  /**
   * List all workflows
   */
  async list(params?: WorkflowListParams): Promise<WorkflowListResponse> {
    const options = params ? { params } : undefined
    return await apiClient.get('/api/v1/workflows', options)
  }

  /**
   * Get a single workflow by ID
   */
  async get(id: string): Promise<Workflow> {
    return await apiClient.get(`/api/v1/workflows/${id}`)
  }

  /**
   * Create a new workflow
   */
  async create(workflow: WorkflowCreateInput): Promise<Workflow> {
    return await apiClient.post('/api/v1/workflows', workflow)
  }

  /**
   * Update an existing workflow
   */
  async update(id: string, updates: WorkflowUpdateInput): Promise<Workflow> {
    return await apiClient.put(`/api/v1/workflows/${id}`, updates)
  }

  /**
   * Delete a workflow
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/api/v1/workflows/${id}`)
  }

  /**
   * Execute a workflow
   */
  async execute(id: string, input?: Record<string, unknown>): Promise<WorkflowExecutionResponse> {
    return await apiClient.post(`/api/v1/workflows/${id}/execute`, input || {})
  }

  /**
   * Update workflow status (activate/deactivate)
   */
  async updateStatus(id: string, status: WorkflowStatus): Promise<Workflow> {
    return await this.update(id, { status })
  }
}

export const workflowAPI = new WorkflowAPI()
