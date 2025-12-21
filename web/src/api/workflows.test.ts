import { describe, it, expect, beforeEach, vi } from 'vitest'
import { workflowAPI } from './workflows'
import type { Workflow, WorkflowCreateInput, WorkflowUpdateInput } from './workflows'

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

describe('Workflow API', () => {
  const mockWorkflow: Workflow = {
    id: 'wf-123',
    tenantId: 'tenant-1',
    name: 'Test Workflow',
    description: 'Test description',
    status: 'draft',
    definition: {
      nodes: [
        {
          id: 'node-1',
          type: 'trigger',
          position: { x: 0, y: 0 },
          data: { label: 'Webhook', triggerType: 'webhook', config: {} },
        },
      ],
      edges: [],
    },
    version: 1,
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('list', () => {
    it('should fetch list of workflows', async () => {
      const mockWorkflows = [mockWorkflow]
      ;(apiClient.get as any).mockResolvedValueOnce({ workflows: mockWorkflows, total: 1 })

      const result = await workflowAPI.list()

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/workflows', undefined)
      expect(result).toEqual({ workflows: mockWorkflows, total: 1 })
    })

    it('should handle empty list', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ workflows: [], total: 0 })

      const result = await workflowAPI.list()

      expect(result.workflows).toEqual([])
      expect(result.total).toBe(0)
    })

    it('should support pagination parameters', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ workflows: [], total: 0 })

      await workflowAPI.list({ page: 2, limit: 20 })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/workflows', {
        params: { page: 2, limit: 20 },
      })
    })

    it('should support status filter', async () => {
      (apiClient.get as any).mockResolvedValueOnce({ workflows: [], total: 0 })

      await workflowAPI.list({ status: 'active' })

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/workflows', {
        params: { status: 'active' },
      })
    })
  })

  describe('get', () => {
    it('should fetch single workflow by ID', async () => {
      (apiClient.get as any).mockResolvedValueOnce(mockWorkflow)

      const result = await workflowAPI.get('wf-123')

      expect(apiClient.get).toHaveBeenCalledWith('/api/v1/workflows/wf-123')
      expect(result).toEqual(mockWorkflow)
    })

    it('should throw NotFoundError for invalid ID', async () => {
      const error = new Error('Not found')
      error.name = 'NotFoundError'
      ;(apiClient.get as any).mockRejectedValueOnce(error)

      await expect(workflowAPI.get('invalid-id')).rejects.toThrow('Not found')
    })
  })

  describe('create', () => {
    it('should create new workflow', async () => {
      const createInput: WorkflowCreateInput = {
        name: 'New Workflow',
        description: 'New description',
        definition: {
          nodes: [],
          edges: [],
        },
      }

      const createdWorkflow = { ...mockWorkflow, ...createInput }
      ;(apiClient.post as any).mockResolvedValueOnce(createdWorkflow)

      const result = await workflowAPI.create(createInput)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/workflows', createInput)
      expect(result).toEqual(createdWorkflow)
      expect(result.id).toBeDefined()
    })

    it('should validate required fields', async () => {
      const error = new Error('Name is required')
      error.name = 'ValidationError'
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      const invalidInput = {
        name: '',
        definition: { nodes: [], edges: [] },
      } as WorkflowCreateInput

      await expect(workflowAPI.create(invalidInput)).rejects.toThrow('Name is required')
    })

    it('should create workflow with minimal data', async () => {
      const minimalInput: WorkflowCreateInput = {
        name: 'Minimal Workflow',
        definition: {
          nodes: [],
          edges: [],
        },
      }

      ;(apiClient.post as any).mockResolvedValueOnce({ ...mockWorkflow, ...minimalInput })

      const result = await workflowAPI.create(minimalInput)

      expect(result.name).toBe('Minimal Workflow')
    })
  })

  describe('update', () => {
    it('should update existing workflow', async () => {
      const updates: WorkflowUpdateInput = {
        name: 'Updated Workflow',
        description: 'Updated description',
      }

      const updatedWorkflow = { ...mockWorkflow, ...updates }
      ;(apiClient.put as any).mockResolvedValueOnce(updatedWorkflow)

      const result = await workflowAPI.update('wf-123', updates)

      expect(apiClient.put).toHaveBeenCalledWith('/api/v1/workflows/wf-123', updates)
      expect(result.name).toBe('Updated Workflow')
      expect(result.description).toBe('Updated description')
    })

    it('should preserve unchanged fields', async () => {
      const updates: WorkflowUpdateInput = {
        name: 'Updated Name Only',
      }

      const updatedWorkflow = { ...mockWorkflow, ...updates }
      ;(apiClient.put as any).mockResolvedValueOnce(updatedWorkflow)

      const result = await workflowAPI.update('wf-123', updates)

      expect(result.name).toBe('Updated Name Only')
      expect(result.description).toBe(mockWorkflow.description) // Unchanged
    })

    it('should update workflow definition', async () => {
      const updates: WorkflowUpdateInput = {
        definition: {
          nodes: [
            {
              id: 'node-2',
              type: 'action',
              position: { x: 100, y: 100 },
              data: { label: 'HTTP Request', actionType: 'http', config: {} },
            },
          ],
          edges: [],
        },
      }

      ;(apiClient.put as any).mockResolvedValueOnce({ ...mockWorkflow, ...updates })

      const result = await workflowAPI.update('wf-123', updates)

      expect(result.definition.nodes).toHaveLength(1)
      expect(result.definition.nodes[0].id).toBe('node-2')
    })

    it('should throw NotFoundError for non-existent workflow', async () => {
      const error = new Error('Workflow not found')
      error.name = 'NotFoundError'
      ;(apiClient.put as any).mockRejectedValueOnce(error)

      await expect(workflowAPI.update('invalid-id', { name: 'Test' })).rejects.toThrow(
        'Workflow not found'
      )
    })
  })

  describe('delete', () => {
    it('should delete workflow by ID', async () => {
      (apiClient.delete as any).mockResolvedValueOnce({})

      await workflowAPI.delete('wf-123')

      expect(apiClient.delete).toHaveBeenCalledWith('/api/v1/workflows/wf-123')
    })

    it('should throw NotFoundError for non-existent workflow', async () => {
      const error = new Error('Workflow not found')
      error.name = 'NotFoundError'
      ;(apiClient.delete as any).mockRejectedValueOnce(error)

      await expect(workflowAPI.delete('invalid-id')).rejects.toThrow('Workflow not found')
    })
  })

  describe('execute', () => {
    it('should trigger workflow execution', async () => {
      const executionResponse = {
        executionId: 'exec-123',
        workflowId: 'wf-123',
        status: 'queued',
        queuedAt: '2024-01-15T10:00:00Z',
      }

      ;(apiClient.post as any).mockResolvedValueOnce(executionResponse)

      const result = await workflowAPI.execute('wf-123')

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/workflows/wf-123/execute', {})
      expect(result.executionId).toBe('exec-123')
      expect(result.status).toBe('queued')
    })

    it('should support execution with input data', async () => {
      const inputData = { userId: '123', action: 'test' }
      const executionResponse = {
        executionId: 'exec-123',
        workflowId: 'wf-123',
        status: 'queued',
        queuedAt: '2024-01-15T10:00:00Z',
      }

      ;(apiClient.post as any).mockResolvedValueOnce(executionResponse)

      await workflowAPI.execute('wf-123', inputData)

      expect(apiClient.post).toHaveBeenCalledWith('/api/v1/workflows/wf-123/execute', inputData)
    })

    it('should throw ValidationError for invalid workflow state', async () => {
      const error = new Error('Workflow must be active to execute')
      error.name = 'ValidationError'
      ;(apiClient.post as any).mockRejectedValueOnce(error)

      await expect(workflowAPI.execute('wf-123')).rejects.toThrow(
        'Workflow must be active to execute'
      )
    })
  })

  describe('updateStatus', () => {
    it('should update workflow status to active', async () => {
      const updatedWorkflow = { ...mockWorkflow, status: 'active' as const }
      ;(apiClient.put as any).mockResolvedValueOnce(updatedWorkflow)

      const result = await workflowAPI.updateStatus('wf-123', 'active')

      expect(apiClient.put).toHaveBeenCalledWith('/api/v1/workflows/wf-123', { status: 'active' })
      expect(result.status).toBe('active')
    })

    it('should update workflow status to inactive', async () => {
      const updatedWorkflow = { ...mockWorkflow, status: 'inactive' as const }
      ;(apiClient.put as any).mockResolvedValueOnce(updatedWorkflow)

      const result = await workflowAPI.updateStatus('wf-123', 'inactive')

      expect(result.status).toBe('inactive')
    })
  })

  describe('Version Management', () => {
    describe('listVersions', () => {
      it('should list all versions for a workflow', async () => {
        const workflowId = 'wf-123'
        const mockVersions = [
          {
            id: 'version-3',
            workflowId,
            version: 3,
            definition: { nodes: [], edges: [] },
            createdBy: 'user-1',
            createdAt: '2025-12-19T12:00:00Z',
          },
          {
            id: 'version-2',
            workflowId,
            version: 2,
            definition: { nodes: [], edges: [] },
            createdBy: 'user-1',
            createdAt: '2025-12-19T11:00:00Z',
          },
        ]

        ;(apiClient.get as any).mockResolvedValueOnce({ data: mockVersions })

        const result = await workflowAPI.listVersions(workflowId)

        expect(apiClient.get).toHaveBeenCalledWith(`/api/v1/workflows/${workflowId}/versions`)
        expect(result).toEqual(mockVersions)
      })

      it('should handle empty version list', async () => {
        (apiClient.get as any).mockResolvedValueOnce({ data: [] })

        const result = await workflowAPI.listVersions('wf-123')

        expect(result).toEqual([])
      })
    })

    describe('getVersion', () => {
      it('should get a specific version', async () => {
        const workflowId = 'wf-123'
        const version = 2
        const mockVersion = {
          id: 'version-2',
          workflowId,
          version,
          definition: { nodes: [{ id: 'node1', type: 'trigger:webhook', position: { x: 0, y: 0 }, data: {} }], edges: [] },
          createdBy: 'user-1',
          createdAt: '2025-12-19T11:00:00Z',
        }

        ;(apiClient.get as any).mockResolvedValueOnce({ data: mockVersion })

        const result = await workflowAPI.getVersion(workflowId, version)

        expect(apiClient.get).toHaveBeenCalledWith(`/api/v1/workflows/${workflowId}/versions/${version}`)
        expect(result).toEqual(mockVersion)
      })

      it('should throw error when version not found', async () => {
        const error = new Error('version not found')
        error.name = 'NotFoundError'
        ;(apiClient.get as any).mockRejectedValueOnce(error)

        await expect(workflowAPI.getVersion('wf-123', 999)).rejects.toThrow('version not found')
      })
    })

    describe('restoreVersion', () => {
      it('should restore a workflow to a previous version', async () => {
        const workflowId = 'wf-123'
        const version = 1
        const mockRestoredWorkflow = {
          ...mockWorkflow,
          version: 4, // New version after restore
          updatedAt: '2025-12-19T13:00:00Z',
        }

        ;(apiClient.post as any).mockResolvedValueOnce({ data: mockRestoredWorkflow })

        const result = await workflowAPI.restoreVersion(workflowId, version)

        expect(apiClient.post).toHaveBeenCalledWith(`/api/v1/workflows/${workflowId}/versions/${version}/restore`, {})
        expect(result).toEqual(mockRestoredWorkflow)
        expect(result.version).toBe(4) // Version incremented
      })

      it('should throw error when workflow not found', async () => {
        const error = new Error('workflow not found')
        error.name = 'NotFoundError'
        ;(apiClient.post as any).mockRejectedValueOnce(error)

        await expect(workflowAPI.restoreVersion('invalid-id', 1)).rejects.toThrow('workflow not found')
      })

      it('should throw error when version not found', async () => {
        const error = new Error('version not found')
        error.name = 'NotFoundError'
        ;(apiClient.post as any).mockRejectedValueOnce(error)

        await expect(workflowAPI.restoreVersion('wf-123', 999)).rejects.toThrow('version not found')
      })
    })
  })
})
