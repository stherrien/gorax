import { apiClient } from './client'

export type TemplateCategory = 'security' | 'monitoring' | 'integration' | 'dataops' | 'devops' | 'other'

export interface Template {
  id: string
  tenantId?: string
  name: string
  description: string
  category: TemplateCategory
  definition: WorkflowDefinition
  tags: string[]
  isPublic: boolean
  createdBy: string
  createdAt: string
  updatedAt: string
}

export interface WorkflowDefinition {
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
}

export interface WorkflowNode {
  id: string
  type: string
  position: { x: number; y: number }
  data: {
    name: string
    config: unknown
  }
}

export interface WorkflowEdge {
  id: string
  source: string
  target: string
  sourceHandle?: string
  targetHandle?: string
  label?: string
}

export interface TemplateListParams {
  category?: string
  tags?: string[]
  isPublic?: boolean
  search?: string
}

export interface CreateTemplateInput {
  name: string
  description?: string
  category: TemplateCategory
  definition: WorkflowDefinition
  tags?: string[]
  isPublic?: boolean
}

export interface UpdateTemplateInput {
  name?: string
  description?: string
  category?: TemplateCategory
  definition?: WorkflowDefinition
  tags?: string[]
  isPublic?: boolean
}

export interface CreateFromWorkflowInput {
  name: string
  description?: string
  category: TemplateCategory
  definition: WorkflowDefinition
  tags?: string[]
  isPublic?: boolean
}

export interface InstantiateTemplateInput {
  workflowName: string
}

export interface InstantiateTemplateResult {
  workflowName: string
  definition: WorkflowDefinition
}

class TemplateAPI {
  /**
   * List all templates with optional filters
   */
  async list(params?: TemplateListParams): Promise<Template[]> {
    const queryParams: Record<string, string> = {}

    if (params?.category) {
      queryParams.category = params.category
    }

    if (params?.tags && params.tags.length > 0) {
      queryParams.tags = params.tags.join(',')
    }

    if (params?.isPublic !== undefined) {
      queryParams.is_public = String(params.isPublic)
    }

    if (params?.search) {
      queryParams.search = params.search
    }

    const response = await apiClient.get('/api/v1/templates', {
      params: queryParams
    })

    return response.data || response
  }

  /**
   * Get a single template by ID
   */
  async get(id: string): Promise<Template> {
    const response = await apiClient.get(`/api/v1/templates/${id}`)
    return response.data || response
  }

  /**
   * Create a new template
   */
  async create(template: CreateTemplateInput): Promise<Template> {
    const response = await apiClient.post('/api/v1/templates', template)
    return response.data || response
  }

  /**
   * Update an existing template
   */
  async update(id: string, updates: UpdateTemplateInput): Promise<void> {
    await apiClient.put(`/api/v1/templates/${id}`, updates)
  }

  /**
   * Delete a template
   */
  async delete(id: string): Promise<void> {
    await apiClient.delete(`/api/v1/templates/${id}`)
  }

  /**
   * Create a template from an existing workflow
   */
  async createFromWorkflow(
    workflowId: string,
    input: CreateFromWorkflowInput
  ): Promise<Template> {
    const response = await apiClient.post(
      `/api/v1/templates/from-workflow/${workflowId}`,
      input
    )
    return response.data || response
  }

  /**
   * Instantiate a template to create a workflow definition
   */
  async instantiate(
    templateId: string,
    input: InstantiateTemplateInput
  ): Promise<InstantiateTemplateResult> {
    const response = await apiClient.post(
      `/api/v1/templates/${templateId}/instantiate`,
      input
    )
    return response.data || response
  }
}

export const templateAPI = new TemplateAPI()
