import { describe, it, expect, beforeEach, vi } from 'vitest'
import { templateAPI } from './templates'
import type {
  Template,
  CreateTemplateInput,
  UpdateTemplateInput,
  CreateFromWorkflowInput,
  InstantiateTemplateInput,
} from './templates'

// Mock the API client
vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}))

import { apiClient } from './client'

describe('Template API', () => {
  const mockTemplate: Template = {
    id: 'tmpl-123',
    tenantId: 'tenant-1',
    name: 'Security Scan Pipeline',
    description: 'Automated security scanning workflow',
    category: 'security',
    definition: {
      nodes: [
        {
          id: 'node-1',
          type: 'trigger:webhook',
          position: { x: 100, y: 100 },
          data: {
            name: 'Webhook Trigger',
            config: {},
          },
        },
      ],
      edges: [],
    },
    tags: ['security', 'scan'],
    isPublic: false,
    createdBy: 'user-123',
    createdAt: '2024-01-15T09:00:00Z',
    updatedAt: '2024-01-15T09:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should list all templates', async () => {
      const mockTemplates = [mockTemplate]
      vi.mocked(apiClient.get).mockResolvedValue({ data: mockTemplates })

      const result = await templateAPI.list()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/templates', {
        params: {}
      })
      expect(result).toEqual(mockTemplates)
    })

    it('should list templates with category filter', async () => {
      const mockTemplates = [mockTemplate]
      vi.mocked(apiClient.get).mockResolvedValue({ data: mockTemplates })

      const result = await templateAPI.list({ category: 'security' })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/templates', {
        params: { category: 'security' }
      })
      expect(result).toEqual(mockTemplates)
    })

    it('should list templates with tags filter', async () => {
      const mockTemplates = [mockTemplate]
      vi.mocked(apiClient.get).mockResolvedValue({ data: mockTemplates })

      const result = await templateAPI.list({ tags: ['security', 'scan'] })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/templates', {
        params: { tags: 'security,scan' }
      })
      expect(result).toEqual(mockTemplates)
    })

    it('should list templates with search query', async () => {
      const mockTemplates = [mockTemplate]
      vi.mocked(apiClient.get).mockResolvedValue({ data: mockTemplates })

      const result = await templateAPI.list({ search: 'security' })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/templates', {
        params: { search: 'security' }
      })
      expect(result).toEqual(mockTemplates)
    })

    it('should list templates with isPublic filter', async () => {
      const mockTemplates = [mockTemplate]
      vi.mocked(apiClient.get).mockResolvedValue({ data: mockTemplates })

      const result = await templateAPI.list({ isPublic: true })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/templates', {
        params: { is_public: 'true' }
      })
      expect(result).toEqual(mockTemplates)
    })
  })

  describe('get', () => {
    it('should get a single template', async () => {
      vi.mocked(apiClient.get).mockResolvedValue({ data: mockTemplate })

      const result = await templateAPI.get('tmpl-123')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/templates/tmpl-123')
      expect(result).toEqual(mockTemplate)
    })

    it('should handle API response without data wrapper', async () => {
      vi.mocked(apiClient.get).mockResolvedValue(mockTemplate)

      const result = await templateAPI.get('tmpl-123')

      expect(result).toEqual(mockTemplate)
    })
  })

  describe('create', () => {
    it('should create a new template', async () => {
      const input: CreateTemplateInput = {
        name: 'New Template',
        description: 'Test template',
        category: 'security',
        definition: {
          nodes: [],
          edges: [],
        },
        tags: ['test'],
        isPublic: false,
      }

      vi.mocked(apiClient.post).mockResolvedValue({ data: mockTemplate })

      const result = await templateAPI.create(input)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/templates', input)
      expect(result).toEqual(mockTemplate)
    })
  })

  describe('update', () => {
    it('should update a template', async () => {
      const updates: UpdateTemplateInput = {
        name: 'Updated Name',
        description: 'Updated description',
      }

      vi.mocked(apiClient.put).mockResolvedValue({})

      await templateAPI.update('tmpl-123', updates)

      expect(apiClient.put).toHaveBeenCalledWith(
        '/api/v1/templates/tmpl-123',
        updates
      )
    })
  })

  describe('delete', () => {
    it('should delete a template', async () => {
      vi.mocked(apiClient.delete).mockResolvedValue({})

      await templateAPI.delete('tmpl-123')

      expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/templates/tmpl-123')
    })
  })

  describe('createFromWorkflow', () => {
    it('should create a template from a workflow', async () => {
      const input: CreateFromWorkflowInput = {
        name: 'From Workflow',
        category: 'integration',
        definition: {
          nodes: [],
          edges: [],
        },
      }

      vi.mocked(apiClient.post).mockResolvedValue({ data: mockTemplate })

      const result = await templateAPI.createFromWorkflow('wf-123', input)

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/v1/templates/from-workflow/wf-123',
        input
      )
      expect(result).toEqual(mockTemplate)
    })
  })

  describe('instantiate', () => {
    it('should instantiate a template', async () => {
      const input: InstantiateTemplateInput = {
        workflowName: 'New Workflow',
      }

      const mockResult = {
        workflowName: 'New Workflow',
        definition: {
          nodes: [],
          edges: [],
        },
      }

      vi.mocked(apiClient.post).mockResolvedValue({ data: mockResult })

      const result = await templateAPI.instantiate('tmpl-123', input)

      expect(apiClient.post).toHaveBeenCalledWith(
        '/api/v1/templates/tmpl-123/instantiate',
        input
      )
      expect(result).toEqual(mockResult)
    })
  })
})
