import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { useTemplates, useTemplate, useTemplateMutations } from './useTemplates'
import type { Template } from '../api/templates'

// Mock the template API
import { createQueryWrapper } from "../test/test-utils"
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

describe('useTemplates', () => {
  const mockTemplate: Template = {
    id: '12345678-1234-4234-8234-123456789abc',
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

      const { result } = renderHook(() => useTemplates(), { wrapper: createQueryWrapper() })

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

      const { result } = renderHook(() => useTemplates(), { wrapper: createQueryWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.templates).toEqual([])
      expect(result.current.error).toBeNull()
    })

    it('should handle errors', async () => {
      const error = new Error('Failed to fetch templates')
      vi.mocked(templateAPI.list).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useTemplates(), { wrapper: createQueryWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.templates).toEqual([])
      expect(result.current.error).toEqual(error)
    })

    it('should refetch templates', async () => {
      vi.mocked(templateAPI.list).mockResolvedValue([mockTemplate])

      const { result } = renderHook(() => useTemplates(), { wrapper: createQueryWrapper() })

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
      const { result } = renderHook(() => useTemplates(params), { wrapper: createQueryWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(templateAPI.list).toHaveBeenCalledWith(params)
    })
  })

  describe('useTemplate - single template hook', () => {
    it('should load template on mount', async () => {
      vi.mocked(templateAPI.get).mockResolvedValueOnce(mockTemplate)

      const { result } = renderHook(() => useTemplate('12345678-1234-4234-8234-123456789abc'), { wrapper: createQueryWrapper() })

      expect(result.current.loading).toBe(true)

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.template).toEqual(mockTemplate)
      expect(result.current.error).toBeNull()
    })

    it('should not load if id is null', async () => {
      const { result } = renderHook(() => useTemplate(null), { wrapper: createQueryWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.template).toBeNull()
      expect(templateAPI.get).not.toHaveBeenCalled()
    })

    it('should handle errors', async () => {
      const error = new Error('Template not found')
      vi.mocked(templateAPI.get).mockRejectedValueOnce(error)

      const { result } = renderHook(() => useTemplate('12345678-1234-4234-8234-123456789abc'), { wrapper: createQueryWrapper() })

      await waitFor(() => {
        expect(result.current.loading).toBe(false)
      })

      expect(result.current.template).toBeNull()
      expect(result.current.error).toEqual(error)
    })

    it('should refetch template', async () => {
      vi.mocked(templateAPI.get).mockResolvedValue(mockTemplate)

      const { result } = renderHook(() => useTemplate('12345678-1234-4234-8234-123456789abc'), { wrapper: createQueryWrapper() })

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

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createQueryWrapper() })

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

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createQueryWrapper() })

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

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createQueryWrapper() })

      await act(async () => {
        await result.current.updateTemplate('12345678-1234-4234-8234-123456789abc', updates)
      })

      expect(templateAPI.update).toHaveBeenCalledWith('12345678-1234-4234-8234-123456789abc', updates)
    })

    it('should delete a template', async () => {
      vi.mocked(templateAPI.delete).mockResolvedValueOnce()

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createQueryWrapper() })

      await act(async () => {
        await result.current.deleteTemplate('12345678-1234-4234-8234-123456789abc')
      })

      expect(templateAPI.delete).toHaveBeenCalledWith('12345678-1234-4234-8234-123456789abc')
    })

    it('should create template from workflow', async () => {
      const input = {
        name: 'From Workflow',
        category: 'integration' as const,
        definition: { nodes: [], edges: [] },
      }

      vi.mocked(templateAPI.createFromWorkflow).mockResolvedValueOnce(mockTemplate)

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createQueryWrapper() })

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

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createQueryWrapper() })

      let instantiateResult

      await act(async () => {
        instantiateResult = await result.current.instantiateTemplate('12345678-1234-4234-8234-123456789abc', input)
      })

      expect(instantiateResult).toEqual(mockResult)
      expect(templateAPI.instantiate).toHaveBeenCalledWith('12345678-1234-4234-8234-123456789abc', input)
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

      const { result } = renderHook(() => useTemplateMutations(), { wrapper: createQueryWrapper() })

      expect(result.current.instantiating).toBe(false)

      const promise = act(async () => {
        await result.current.instantiateTemplate('12345678-1234-4234-8234-123456789abc', input)
      })

      await waitFor(() => {
        expect(result.current.instantiating).toBe(false)
      })

      await promise
    })
  })
})
