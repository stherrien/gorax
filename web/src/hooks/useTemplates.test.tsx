import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { useTemplates, useTemplate, useTemplateMutations } from './useTemplates'
import type { Template } from '../api/templates'

// Mock the template API
vi.mock('../api/templates', () => ({
  templateAPI: {
    list: vi.fn(),
    get: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    createFromWorkflow: vi.fn(),
    instantiate: vi.fn(),
  },
}))

import { templateAPI } from '../api/templates'

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

describe('useTemplates', () => {
  const mockTemplate: Template = {
    id: 'tmpl-123',
    tenantId: 'tenant-1',
    name: 'Security Scan Pipeline',
    description: 'Automated security scanning',
    category: 'security',
    definition: {
      nodes: [],
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

  describe('useTemplates - list hook', () => {
    it('should load templates on mount', async () => {
      vi.mocked(templateAPI.list).mockResolvedValueOnce([mockTemplate])

      const { result } = renderHook(() => useTemplates(), { wrapper: createWrapper() })

      expect(result.current.loading).toBe(true)
      expect(result.current.templates).toEqual([])

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.templates).toEqual([mockTemplate])
      expect(result.current.error).toBeNull()
    })

    it('should handle empty list', async () => {
      vi.mocked(templateAPI.list).mockResolvedValueOnce([])

      const { result } = renderHook(() => useTemplates(), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.templates).toEqual([])
      expect(result.current.error).toBeNull()
    })

    it('should handle errors', async () => {
      const error = new Error('Failed to fetch templates')
      vi.mocked(templateAPI.list).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useTemplates(), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.templates).toEqual([])
      expect(result.current.error).toEqual(error)
    })

    it('should refetch templates', async () => {
      vi.mocked(templateAPI.list).mockResolvedValue([mockTemplate])

      const { result } = renderHook(() => useTemplates(), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      act(() => {
        result.current.refetch()
      })

      await waitFor(() => {
        expect(templateAPI.list).toHaveBeenCalledTimes(2)
      })
    })

    it('should pass filter params to API', async () => {
      vi.mocked(templateAPI.list).mockResolvedValueOnce([mockTemplate])

      const params = { category: 'security', tags: ['scan'] }
      const { result } = renderHook(() => useTemplates(params), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(templateAPI.list).toHaveBeenCalledWith(params)
    })
  })

  describe('useTemplate - single template hook', () => {
    it('should load template on mount', async () => {
      vi.mocked(templateAPI.get).mockResolvedValueOnce(mockTemplate)

      const { result } = renderHook(() => useTemplate('tmpl-123'), { wrapper: createWrapper() })

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.template).toEqual(mockTemplate)
      expect(result.current.error).toBeNull()
    })

    it('should not load if id is null', async () => {
      const { result } = renderHook(() => useTemplate(null), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.template).toBeNull()
      expect(templateAPI.get).not.toHaveBeenCalled()
    })

    it('should handle errors', async () => {
      const error = new Error('Template not found')
      vi.mocked(templateAPI.get).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useTemplate('tmpl-123'), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.template).toBeNull()
      expect(result.current.error).toEqual(error)
    })

    it('should refetch template', async () => {
      vi.mocked(templateAPI.get).mockResolvedValue(mockTemplate)

      const { result } = renderHook(() => useTemplate('tmpl-123'), { wrapper: createWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      act(() => {
        result.current.refetch()
      })

      await waitFor(() => {
        expect(templateAPI.get).toHaveBeenCalledTimes(2)
      })
    })
  })

  describe('useTemplateMutations - mutation hook', () => {
    it('should create a template', async () => {
      const input = {
        name: 'New Template',
        category: 'security' as const,
        definition: { nodes: [], edges: [] },
      }

      vi.mocked(templateAPI.create).mockResolvedValueOnce(mockTemplate)

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createWrapper() })

      let createdTemplate: Template | undefined

      await act(async () => {
        createdTemplate = await result.current.createTemplate(input)
      })

      expect(createdTemplate).toEqual(mockTemplate)
      expect(templateAPI.create).toHaveBeenCalledWith(input)
    })

    it('should track creating state', async () => {
      const input = {
        name: 'New Template',
        category: 'security' as const,
        definition: { nodes: [], edges: [] },
      }

      vi.mocked(templateAPI.create).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockTemplate), 100))
      )

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createWrapper() })

      expect(result.current.creating).toBe(false)

      const promise = act(async () => {
        await result.current.createTemplate(input)
      })

      await waitFor(() => {
        expect(result.current.creating).toBe(false)
      })

      await promise
    })

    it('should update a template', async () => {
      const updates = { name: 'Updated Name' }

      vi.mocked(templateAPI.update).mockResolvedValueOnce()

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createWrapper() })

      await act(async () => {
        await result.current.updateTemplate('tmpl-123', updates)
      })

      expect(templateAPI.update).toHaveBeenCalledWith('tmpl-123', updates)
    })

    it('should delete a template', async () => {
      vi.mocked(templateAPI.delete).mockResolvedValueOnce()

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createWrapper() })

      await act(async () => {
        await result.current.deleteTemplate('tmpl-123')
      })

      expect(templateAPI.delete).toHaveBeenCalledWith('tmpl-123')
    })

    it('should create template from workflow', async () => {
      const input = {
        name: 'From Workflow',
        category: 'integration' as const,
        definition: { nodes: [], edges: [] },
      }

      vi.mocked(templateAPI.createFromWorkflow).mockResolvedValueOnce(mockTemplate)

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createWrapper() })

      let createdTemplate: Template | undefined

      await act(async () => {
        createdTemplate = await result.current.createFromWorkflow('wf-123', input)
      })

      expect(createdTemplate).toEqual(mockTemplate)
      expect(templateAPI.createFromWorkflow).toHaveBeenCalledWith('wf-123', input)
    })

    it('should instantiate template', async () => {
      const input = { workflowName: 'New Workflow' }
      const mockResult = {
        workflowName: 'New Workflow',
        definition: { nodes: [], edges: [] },
      }

      vi.mocked(templateAPI.instantiate).mockResolvedValueOnce(mockResult)

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createWrapper() })

      let instantiateResult

      await act(async () => {
        instantiateResult = await result.current.instantiateTemplate('tmpl-123', input)
      })

      expect(instantiateResult).toEqual(mockResult)
      expect(templateAPI.instantiate).toHaveBeenCalledWith('tmpl-123', input)
    })

    it('should track instantiating state', async () => {
      const input = { workflowName: 'New Workflow' }
      const mockResult = {
        workflowName: 'New Workflow',
        definition: { nodes: [], edges: [] },
      }

      vi.mocked(templateAPI.instantiate).mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve(mockResult), 100))
      )

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createWrapper() })

      expect(result.current.instantiating).toBe(false)

      const promise = act(async () => {
        await result.current.instantiateTemplate('tmpl-123', input)
      })

      await waitFor(() => {
        expect(result.current.instantiating).toBe(false)
      })

      await promise
    })
  })
})
