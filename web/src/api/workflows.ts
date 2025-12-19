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

export interface WorkflowVersion {
  id: string
  workflowId: string
  version: number
  definition: WorkflowDefinition
  createdBy: string
  createdAt: string
}

export interface DryRunWarning {
  nodeId: string
  message: string
}

export interface DryRunError {
  nodeId: string
  field: string
  message: string
}

export interface DryRunResult {
  valid: boolean
  executionOrder: string[]
  variableMapping: Record<string, string>
  warnings: DryRunWarning[]
  errors: DryRunError[]
}

export interface DryRunInput {
  testData?: Record<string, unknown>
}

class WorkflowAPI {
  /**
   * List all workflows
   */
  async list(params?: WorkflowListParams): Promise<WorkflowListResponse> {
    const options = params ? { params } : undefined
    const response = await apiClient.get('/api/v1/workflows', options)
    // Backend returns { data: [], limit, offset } or { workflows: [], total }
    if (response.data && Array.isArray(response.data)) {
      return { workflows: response.data, total: response.data.length }
    }
    return response
  }

  /**
   * Get a single workflow by ID
   */
  async get(id: string): Promise<Workflow> {
    const response = await apiClient.get(`/api/v1/workflows/${id}`)
    return response.data || response
  }

  /**
   * Create a new workflow
   */
  async create(workflow: WorkflowCreateInput): Promise<Workflow> {
    const response = await apiClient.post('/api/v1/workflows', workflow)
    return response.data || response
  }

  /**
   * Update an existing workflow
   */
  async update(id: string, updates: WorkflowUpdateInput): Promise<Workflow> {
    const response = await apiClient.put(`/api/v1/workflows/${id}`, updates)
    return response.data || response
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

  /**
   * List all versions for a workflow
   */
  async listVersions(workflowId: string): Promise<WorkflowVersion[]> {
    const response = await apiClient.get(`/api/v1/workflows/${workflowId}/versions`)
    return response.data || response
  }

  /**
   * Get a specific version of a workflow
   */
  async getVersion(workflowId: string, version: number): Promise<WorkflowVersion> {
    const response = await apiClient.get(`/api/v1/workflows/${workflowId}/versions/${version}`)
    return response.data || response
  }

  /**
   * Restore a workflow to a previous version
   */
  async restoreVersion(workflowId: string, version: number): Promise<Workflow> {
    const response = await apiClient.post(`/api/v1/workflows/${workflowId}/versions/${version}/restore`, {})
    return response.data || response
  }

  /**
   * Perform a dry-run validation of a workflow
   */
  async dryRun(id: string, testData?: Record<string, unknown>): Promise<DryRunResult> {
    const response = await apiClient.post(`/api/v1/workflows/${id}/dry-run`, {
      test_data: testData || {}
    })
    return response.data || response
  }

  /**
   * Export execution logs in specified format
   * @param executionId - Execution ID
   * @param format - Export format (txt, json, csv)
   */
  async exportLogs(executionId: string, format: 'txt' | 'json' | 'csv'): Promise<void> {
    const response = await apiClient.get(
      `/api/v1/executions/${executionId}/logs/export?format=${format}`,
      { responseType: 'blob' }
    )

    const blob = response instanceof Blob ? response : response.data
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${executionId}.${format}`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  }
}

export const workflowAPI = new WorkflowAPI()
