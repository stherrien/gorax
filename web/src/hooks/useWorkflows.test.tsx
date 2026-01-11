import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { useWorkflows, useWorkflow, useWorkflowMutations } from './useWorkflows'
import type { Workflow } from '../api/workflows'

// Mock the workflow API
vi.mock('../api/workflows', () => ({
  workflowAPI: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    execute: vi.fn(),
  },
}))

import { workflowAPI } from '../api/workflows'

// Helper to create a wrapper with QueryClient
function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('useWorkflows', () => {
  const mockWorkflow: Workflow = {
    id: 'wf-123',
    tenantId: 'tenant-1',
    name: 'Test Workflow',
    description: 'Test description',
    status: 'draft',
    definition: { nodes: [], edges: [] },
    version: 1,
    createdAt: '2024-01-15T10:00:00Z',
    updatedAt: '2024-01-15T10:00:00Z',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('useWorkflows - list hook', () => {
    it('should load workflows on mount', async () => {
      (workflowAPI.list as any).mockResolvedValueOnce({
        workflows: [mockWorkflow],
        total: 1,
      })

      const { result } = renderHook(() => useWorkflows(), { wrapper: createWrapper() })

      expect(result.current.loading).toBe(true)
      expect(result.current.workflows).toEqual([])

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.workflows).toEqual([mockWorkflow])
      expect(result.current.total).toBe(1)
      expect(result.current.error).toBeNull()
    })

    it('should handle empty list', async () => {
      (workflowAPI.list as any).mockResolvedValueOnce({
        workflows: [],
        total: 0,
      })

      const { result } = renderHook(() => useWorkflows(), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.workflows).toEqual([])
      expect(result.current.total).toBe(0)
    })

    it('should handle errors', async () => {
      const error = new Error('Failed to load workflows')
      ;(workflowAPI.list as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWorkflows(), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.workflows).toEqual([])
    })

    it('should support refetch', async () => {
      (workflowAPI.list as any).mockResolvedValue({
        workflows: [mockWorkflow],
        total: 1,
      })

      const { result } = renderHook(() => useWorkflows(), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(workflowAPI.list).toHaveBeenCalledTimes(1)

      // Trigger refetch
      result.current.refetch()

      await waitFor(() => {
        expect(workflowAPI.list).toHaveBeenCalledTimes(2)
      })
    })

    it('should pass filter params to API', async () => {
      (workflowAPI.list as any).mockResolvedValueOnce({
        workflows: [],
        total: 0,
      })

      renderHook(() => useWorkflows({ status: 'active', page: 2 }), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(workflowAPI.list).toHaveBeenCalledWith({ status: 'active', page: 2 })
      })
    })
  })

  describe('useWorkflow - single workflow hook', () => {
    it('should load workflow by ID on mount', async () => {
      (workflowAPI.get as any).mockResolvedValueOnce(mockWorkflow)

      const { result } = renderHook(() => useWorkflow('wf-123'), { wrapper: createWrapper() })

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.workflow).toEqual(mockWorkflow)
      expect(result.current.error).toBeNull()
    })

    it('should not load if ID is null', () => {
      const { result } = renderHook(() => useWorkflow(null), { wrapper: createWrapper() })

      expect(result.current.loading).toBe(false)
      expect(result.current.workflow).toBeNull()
      expect(workflowAPI.get).not.toHaveBeenCalled()
    })

    it('should handle not found error', async () => {
      const error = new Error('Workflow not found')
      error.name = 'NotFoundError'
      ;(workflowAPI.get as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWorkflow('invalid-id'), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.error).toBe(error)
      expect(result.current.workflow).toBeNull()
    })

    it('should refetch workflow', async () => {
      (workflowAPI.get as any).mockResolvedValue(mockWorkflow)

      const { result } = renderHook(() => useWorkflow('wf-123'), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      result.current.refetch()

      await waitFor(() => {
        expect(workflowAPI.get).toHaveBeenCalledTimes(2)
      })
    })
  })

  describe('useWorkflowMutations - CRUD operations', () => {
    it('should create workflow', async () => {
      const newWorkflow = {
        name: 'New Workflow',
        definition: { nodes: [], edges: [] },
      }

      ;(workflowAPI.create as any).mockResolvedValueOnce({
        ...mockWorkflow,
        ...newWorkflow,
      })

      const { result } = renderHook(() => useWorkflowMutations(), { wrapper: createWrapper() })

      expect(result.current.creating).toBe(false)

      const created = await result.current.createWorkflow(newWorkflow)

      await waitFor(() => {
        expect(result.current.creating).toBe(false)
      })

      expect(created.name).toBe('New Workflow')
      expect(workflowAPI.create).toHaveBeenCalledWith(newWorkflow)
    })

    it('should handle create error', async () => {
      const error = new Error('Validation failed')
      ;(workflowAPI.create as any).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useWorkflowMutations(), { wrapper: createWrapper() })

      await expect(
        result.current.createWorkflow({
          name: '',
          definition: { nodes: [], edges: [] },
        })
      ).rejects.toThrow('Validation failed')

      expect(result.current.creating).toBe(false)
    })

    it('should update workflow', async () => {
      const updates = { name: 'Updated Workflow' }
      ;(workflowAPI.update as any).mockResolvedValueOnce({
        ...mockWorkflow,
        ...updates,
      })

      const { result } = renderHook(() => useWorkflowMutations(), { wrapper: createWrapper() })

      expect(result.current.updating).toBe(false)

      const updated = await result.current.updateWorkflow('wf-123', updates)

      await waitFor(() => {
        expect(result.current.updating).toBe(false)
      })

      expect(updated.name).toBe('Updated Workflow')
      expect(workflowAPI.update).toHaveBeenCalledWith('wf-123', updates)
    })

    it('should delete workflow', async () => {
      (workflowAPI.delete as any).mockResolvedValueOnce(undefined)

      const { result } = renderHook(() => useWorkflowMutations(), { wrapper: createWrapper() })

      expect(result.current.deleting).toBe(false)

      await result.current.deleteWorkflow('wf-123')

      await waitFor(() => {
        expect(result.current.deleting).toBe(false)
      })

      expect(workflowAPI.delete).toHaveBeenCalledWith('wf-123')
    })

    it('should execute workflow', async () => {
      const executionResponse = {
        executionId: 'exec-123',
        workflowId: 'wf-123',
        status: 'queued',
        queuedAt: '2024-01-15T10:00:00Z',
      }

      ;(workflowAPI.execute as any).mockResolvedValueOnce(executionResponse)

      const { result } = renderHook(() => useWorkflowMutations(), { wrapper: createWrapper() })

      expect(result.current.executing).toBe(false)

      const response = await result.current.executeWorkflow('wf-123')

      await waitFor(() => {
        expect(result.current.executing).toBe(false)
      })

      expect(response.executionId).toBe('exec-123')
      expect(workflowAPI.execute).toHaveBeenCalledWith('wf-123', undefined)
    })

    it('should execute workflow with input data', async () => {
      const inputData = { userId: '123' }
      const executionResponse = {
        executionId: 'exec-123',
        workflowId: 'wf-123',
        status: 'queued',
        queuedAt: '2024-01-15T10:00:00Z',
      }

      ;(workflowAPI.execute as any).mockResolvedValueOnce(executionResponse)

      const { result } = renderHook(() => useWorkflowMutations(), { wrapper: createWrapper() })

      await result.current.executeWorkflow('wf-123', inputData)

      expect(workflowAPI.execute).toHaveBeenCalledWith('wf-123', inputData)
    })
  })
})
