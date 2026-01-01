import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { useBulkWorkflows } from './useBulkWorkflows'
import { workflowAPI } from '../api/workflows'

// Mock the workflow API
vi.mock('../api/workflows', () => ({
  workflowAPI: {
    bulkDelete: vi.fn(),
    bulkEnable: vi.fn(),
    bulkDisable: vi.fn(),
    bulkExport: vi.fn(),
    bulkClone: vi.fn(),
  },
}))

// Mock URL and document APIs for export download
const mockCreateObjectURL = vi.fn(() => 'blob:mock-url')
const mockRevokeObjectURL = vi.fn()
const mockClick = vi.fn()

// Store original createElement to avoid recursive mocking
const originalCreateElement = document.createElement.bind(document)

beforeEach(() => {
  vi.clearAllMocks()

  // Setup URL mocks
  global.URL.createObjectURL = mockCreateObjectURL
  global.URL.revokeObjectURL = mockRevokeObjectURL

  // Mock createElement to track link click
  vi.spyOn(document, 'createElement').mockImplementation((tagName: string) => {
    if (tagName === 'a') {
      const link = originalCreateElement('a')
      link.click = mockClick
      return link
    }
    return originalCreateElement(tagName)
  })
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('useBulkWorkflows', () => {
  describe('Bulk Delete', () => {
    it('should delete multiple workflows successfully', async () => {
      const mockResult = {
        success_count: 3,
        failures: [],
      }
      vi.mocked(workflowAPI.bulkDelete).mockResolvedValue(mockResult)

      const { result } = renderHook(() => useBulkWorkflows())

      let deleteResult
      await act(async () => {
        deleteResult = await result.current.bulkDelete(['wf-1', 'wf-2', 'wf-3'])
      })

      expect(workflowAPI.bulkDelete).toHaveBeenCalledWith(['wf-1', 'wf-2', 'wf-3'])
      expect(deleteResult).toEqual(mockResult)
    })

    it('should handle partial failures in bulk delete', async () => {
      const mockResult = {
        success_count: 2,
        failures: [
          { workflow_id: 'wf-3', error: 'Workflow not found' },
        ],
      }
      vi.mocked(workflowAPI.bulkDelete).mockResolvedValue(mockResult)

      const { result } = renderHook(() => useBulkWorkflows())

      let deleteResult
      await act(async () => {
        deleteResult = await result.current.bulkDelete(['wf-1', 'wf-2', 'wf-3'])
      })

      expect(deleteResult?.success_count).toBe(2)
      expect(deleteResult?.failures).toHaveLength(1)
      expect(deleteResult?.failures[0].workflow_id).toBe('wf-3')
    })

    it('should throw error on complete failure', async () => {
      vi.mocked(workflowAPI.bulkDelete).mockRejectedValue(new Error('Network error'))

      const { result } = renderHook(() => useBulkWorkflows())

      await expect(
        act(async () => {
          await result.current.bulkDelete(['wf-1'])
        })
      ).rejects.toThrow('Network error')
    })

    it('should show loading state during delete', async () => {
      let resolvePromise: (value: unknown) => void
      const pendingPromise = new Promise((resolve) => {
        resolvePromise = resolve
      })
      vi.mocked(workflowAPI.bulkDelete).mockReturnValue(pendingPromise as Promise<any>)

      const { result } = renderHook(() => useBulkWorkflows())

      expect(result.current.bulkDeleting).toBe(false)
      expect(result.current.isLoading).toBe(false)

      // Start delete without awaiting
      let deletePromise: Promise<any>
      act(() => {
        deletePromise = result.current.bulkDelete(['wf-1'])
      })

      // Check loading state
      await waitFor(() => {
        expect(result.current.bulkDeleting).toBe(true)
        expect(result.current.isLoading).toBe(true)
      })

      // Resolve the promise
      await act(async () => {
        resolvePromise!({ success_count: 1, failures: [] })
        await deletePromise
      })

      // Check loading state is reset
      expect(result.current.bulkDeleting).toBe(false)
      expect(result.current.isLoading).toBe(false)
    })

    it('should reset loading state on error', async () => {
      vi.mocked(workflowAPI.bulkDelete).mockRejectedValue(new Error('Failed'))

      const { result } = renderHook(() => useBulkWorkflows())

      await act(async () => {
        try {
          await result.current.bulkDelete(['wf-1'])
        } catch {
          // Expected error
        }
      })

      expect(result.current.bulkDeleting).toBe(false)
      expect(result.current.isLoading).toBe(false)
    })
  })

  describe('Bulk Enable', () => {
    it('should enable multiple workflows', async () => {
      const mockResult = { success_count: 2, failures: [] }
      vi.mocked(workflowAPI.bulkEnable).mockResolvedValue(mockResult)

      const { result } = renderHook(() => useBulkWorkflows())

      let enableResult
      await act(async () => {
        enableResult = await result.current.bulkEnable(['wf-1', 'wf-2'])
      })

      expect(workflowAPI.bulkEnable).toHaveBeenCalledWith(['wf-1', 'wf-2'])
      expect(enableResult).toEqual(mockResult)
    })

    it('should show loading state during enable', async () => {
      let resolvePromise: (value: unknown) => void
      vi.mocked(workflowAPI.bulkEnable).mockImplementation(
        () => new Promise((resolve) => { resolvePromise = resolve })
      )

      const { result } = renderHook(() => useBulkWorkflows())

      expect(result.current.bulkEnabling).toBe(false)

      let enablePromise: Promise<any>
      act(() => {
        enablePromise = result.current.bulkEnable(['wf-1'])
      })

      await waitFor(() => {
        expect(result.current.bulkEnabling).toBe(true)
      })

      await act(async () => {
        resolvePromise!({ success_count: 1, failures: [] })
        await enablePromise
      })

      expect(result.current.bulkEnabling).toBe(false)
    })

    it('should handle enable failures', async () => {
      const mockResult = {
        success_count: 1,
        failures: [{ workflow_id: 'wf-2', error: 'Already enabled' }],
      }
      vi.mocked(workflowAPI.bulkEnable).mockResolvedValue(mockResult)

      const { result } = renderHook(() => useBulkWorkflows())

      let enableResult
      await act(async () => {
        enableResult = await result.current.bulkEnable(['wf-1', 'wf-2'])
      })

      expect(enableResult?.failures).toHaveLength(1)
      expect(enableResult?.failures[0].workflow_id).toBe('wf-2')
    })
  })

  describe('Bulk Disable', () => {
    it('should disable multiple workflows', async () => {
      const mockResult = { success_count: 2, failures: [] }
      vi.mocked(workflowAPI.bulkDisable).mockResolvedValue(mockResult)

      const { result } = renderHook(() => useBulkWorkflows())

      let disableResult
      await act(async () => {
        disableResult = await result.current.bulkDisable(['wf-1', 'wf-2'])
      })

      expect(workflowAPI.bulkDisable).toHaveBeenCalledWith(['wf-1', 'wf-2'])
      expect(disableResult).toEqual(mockResult)
    })

    it('should show loading state during disable', async () => {
      let resolvePromise: (value: unknown) => void
      vi.mocked(workflowAPI.bulkDisable).mockImplementation(
        () => new Promise((resolve) => { resolvePromise = resolve })
      )

      const { result } = renderHook(() => useBulkWorkflows())

      expect(result.current.bulkDisabling).toBe(false)

      let disablePromise: Promise<any>
      act(() => {
        disablePromise = result.current.bulkDisable(['wf-1'])
      })

      await waitFor(() => {
        expect(result.current.bulkDisabling).toBe(true)
      })

      await act(async () => {
        resolvePromise!({ success_count: 1, failures: [] })
        await disablePromise
      })

      expect(result.current.bulkDisabling).toBe(false)
    })
  })

  describe('Bulk Export', () => {
    it('should export multiple workflows', async () => {
      const mockExport = {
        export: {
          workflows: [
            { id: 'wf-1', name: 'Workflow 1', definition: {}, status: 'active', version: 1 },
            { id: 'wf-2', name: 'Workflow 2', definition: {}, status: 'draft', version: 1 },
          ],
          exported_at: '2025-01-15T00:00:00Z',
          version: '1.0',
        },
        result: { success_count: 2, failures: [] },
      }
      vi.mocked(workflowAPI.bulkExport).mockResolvedValue(mockExport)

      const { result } = renderHook(() => useBulkWorkflows())

      let exportResult
      await act(async () => {
        exportResult = await result.current.bulkExport(['wf-1', 'wf-2'])
      })

      expect(workflowAPI.bulkExport).toHaveBeenCalledWith(['wf-1', 'wf-2'])
      expect(exportResult?.export.workflows).toHaveLength(2)
      expect(exportResult?.result.success_count).toBe(2)
    })

    it('should trigger file download on export', async () => {
      const mockExport = {
        export: {
          workflows: [{ id: 'wf-1', name: 'Workflow 1' }],
          exported_at: '2025-01-15T00:00:00Z',
          version: '1.0',
        },
        result: { success_count: 1, failures: [] },
      }
      vi.mocked(workflowAPI.bulkExport).mockResolvedValue(mockExport)

      const { result } = renderHook(() => useBulkWorkflows())

      await act(async () => {
        await result.current.bulkExport(['wf-1'])
      })

      // Verify download was triggered
      expect(mockCreateObjectURL).toHaveBeenCalled()
      expect(mockClick).toHaveBeenCalled()
      expect(mockRevokeObjectURL).toHaveBeenCalledWith('blob:mock-url')
    })

    it('should show loading state during export', async () => {
      let resolvePromise: (value: unknown) => void
      vi.mocked(workflowAPI.bulkExport).mockImplementation(
        () => new Promise((resolve) => { resolvePromise = resolve })
      )

      const { result } = renderHook(() => useBulkWorkflows())

      expect(result.current.bulkExporting).toBe(false)

      let exportPromise: Promise<any>
      act(() => {
        exportPromise = result.current.bulkExport(['wf-1'])
      })

      await waitFor(() => {
        expect(result.current.bulkExporting).toBe(true)
      })

      await act(async () => {
        resolvePromise!({
          export: { workflows: [], exported_at: '', version: '1.0' },
          result: { success_count: 0, failures: [] },
        })
        await exportPromise
      })

      expect(result.current.bulkExporting).toBe(false)
    })

    it('should handle export with partial failures', async () => {
      const mockExport = {
        export: {
          workflows: [{ id: 'wf-1', name: 'Workflow 1' }],
          exported_at: '2025-01-15T00:00:00Z',
          version: '1.0',
        },
        result: {
          success_count: 1,
          failures: [{ workflow_id: 'wf-2', error: 'Permission denied' }],
        },
      }
      vi.mocked(workflowAPI.bulkExport).mockResolvedValue(mockExport)

      const { result } = renderHook(() => useBulkWorkflows())

      let exportResult
      await act(async () => {
        exportResult = await result.current.bulkExport(['wf-1', 'wf-2'])
      })

      expect(exportResult?.result.success_count).toBe(1)
      expect(exportResult?.result.failures).toHaveLength(1)
      expect(exportResult?.export.workflows).toHaveLength(1)
    })
  })

  describe('Bulk Clone', () => {
    it('should clone multiple workflows', async () => {
      const mockClones = {
        clones: [
          { id: 'wf-1-copy', name: 'Workflow 1 (Copy)', status: 'draft' },
          { id: 'wf-2-copy', name: 'Workflow 2 (Copy)', status: 'draft' },
        ],
        result: { success_count: 2, failures: [] },
      }
      vi.mocked(workflowAPI.bulkClone).mockResolvedValue(mockClones)

      const { result } = renderHook(() => useBulkWorkflows())

      let cloneResult
      await act(async () => {
        cloneResult = await result.current.bulkClone(['wf-1', 'wf-2'])
      })

      expect(workflowAPI.bulkClone).toHaveBeenCalledWith(['wf-1', 'wf-2'])
      expect(cloneResult?.clones).toHaveLength(2)
      expect(cloneResult?.clones[0].name).toContain('(Copy)')
    })

    it('should show loading state during clone', async () => {
      let resolvePromise: (value: unknown) => void
      vi.mocked(workflowAPI.bulkClone).mockImplementation(
        () => new Promise((resolve) => { resolvePromise = resolve })
      )

      const { result } = renderHook(() => useBulkWorkflows())

      expect(result.current.bulkCloning).toBe(false)

      let clonePromise: Promise<any>
      act(() => {
        clonePromise = result.current.bulkClone(['wf-1'])
      })

      await waitFor(() => {
        expect(result.current.bulkCloning).toBe(true)
      })

      await act(async () => {
        resolvePromise!({ clones: [], result: { success_count: 0, failures: [] } })
        await clonePromise
      })

      expect(result.current.bulkCloning).toBe(false)
    })

    it('should handle clone failures', async () => {
      const mockClones = {
        clones: [{ id: 'wf-1-copy', name: 'Workflow 1 (Copy)', status: 'draft' }],
        result: {
          success_count: 1,
          failures: [{ workflow_id: 'wf-2', error: 'Original not found' }],
        },
      }
      vi.mocked(workflowAPI.bulkClone).mockResolvedValue(mockClones)

      const { result } = renderHook(() => useBulkWorkflows())

      let cloneResult
      await act(async () => {
        cloneResult = await result.current.bulkClone(['wf-1', 'wf-2'])
      })

      expect(cloneResult?.clones).toHaveLength(1)
      expect(cloneResult?.result.failures).toHaveLength(1)
      expect(cloneResult?.result.failures[0].workflow_id).toBe('wf-2')
    })
  })

  describe('Combined Loading State', () => {
    it('isLoading should be true when any operation is in progress', async () => {
      let resolveDelete: (value: unknown) => void
      let resolveEnable: (value: unknown) => void

      vi.mocked(workflowAPI.bulkDelete).mockImplementation(
        () => new Promise((resolve) => { resolveDelete = resolve })
      )
      vi.mocked(workflowAPI.bulkEnable).mockImplementation(
        () => new Promise((resolve) => { resolveEnable = resolve })
      )

      const { result } = renderHook(() => useBulkWorkflows())

      expect(result.current.isLoading).toBe(false)

      // Start delete
      let deletePromise: Promise<any>
      act(() => {
        deletePromise = result.current.bulkDelete(['wf-1'])
      })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(true)
        expect(result.current.bulkDeleting).toBe(true)
      })

      // Start enable while delete is still running
      let enablePromise: Promise<any>
      act(() => {
        enablePromise = result.current.bulkEnable(['wf-2'])
      })

      await waitFor(() => {
        expect(result.current.isLoading).toBe(true)
        expect(result.current.bulkDeleting).toBe(true)
        expect(result.current.bulkEnabling).toBe(true)
      })

      // Complete delete
      await act(async () => {
        resolveDelete!({ success_count: 1, failures: [] })
        await deletePromise
      })

      // Still loading because enable is running
      expect(result.current.isLoading).toBe(true)
      expect(result.current.bulkDeleting).toBe(false)
      expect(result.current.bulkEnabling).toBe(true)

      // Complete enable
      await act(async () => {
        resolveEnable!({ success_count: 1, failures: [] })
        await enablePromise
      })

      expect(result.current.isLoading).toBe(false)
      expect(result.current.bulkEnabling).toBe(false)
    })
  })

  describe('Error Handling', () => {
    it('should provide error details for all operations', async () => {
      const mockResult = {
        success_count: 1,
        failures: [
          { workflow_id: 'wf-2', error: 'Workflow is locked' },
          { workflow_id: 'wf-3', error: 'Permission denied' },
        ],
      }
      vi.mocked(workflowAPI.bulkDelete).mockResolvedValue(mockResult)

      const { result } = renderHook(() => useBulkWorkflows())

      let deleteResult
      await act(async () => {
        deleteResult = await result.current.bulkDelete(['wf-1', 'wf-2', 'wf-3'])
      })

      expect(deleteResult?.failures).toHaveLength(2)
      expect(deleteResult?.failures[0].workflow_id).toBe('wf-2')
      expect(deleteResult?.failures[0].error).toBe('Workflow is locked')
      expect(deleteResult?.failures[1].workflow_id).toBe('wf-3')
      expect(deleteResult?.failures[1].error).toBe('Permission denied')
    })

    it('should allow retrying after failure', async () => {
      vi.mocked(workflowAPI.bulkDelete)
        .mockResolvedValueOnce({
          success_count: 1,
          failures: [{ workflow_id: 'wf-2', error: 'Temporary error' }],
        })
        .mockResolvedValueOnce({
          success_count: 1,
          failures: [],
        })

      const { result } = renderHook(() => useBulkWorkflows())

      // First attempt
      let firstResult
      await act(async () => {
        firstResult = await result.current.bulkDelete(['wf-1', 'wf-2'])
      })

      expect(firstResult?.failures).toHaveLength(1)

      // Retry with failed ID
      let retryResult
      await act(async () => {
        retryResult = await result.current.bulkDelete(['wf-2'])
      })

      expect(retryResult?.success_count).toBe(1)
      expect(retryResult?.failures).toHaveLength(0)
    })
  })

  describe('Edge Cases', () => {
    it('should handle empty workflow array', async () => {
      vi.mocked(workflowAPI.bulkDelete).mockResolvedValue({
        success_count: 0,
        failures: [],
      })

      const { result } = renderHook(() => useBulkWorkflows())

      let deleteResult
      await act(async () => {
        deleteResult = await result.current.bulkDelete([])
      })

      expect(workflowAPI.bulkDelete).toHaveBeenCalledWith([])
      expect(deleteResult?.success_count).toBe(0)
    })

    it('should handle single workflow', async () => {
      vi.mocked(workflowAPI.bulkClone).mockResolvedValue({
        clones: [{ id: 'wf-1-copy', name: 'Workflow (Copy)', status: 'draft' }],
        result: { success_count: 1, failures: [] },
      })

      const { result } = renderHook(() => useBulkWorkflows())

      let cloneResult
      await act(async () => {
        cloneResult = await result.current.bulkClone(['wf-1'])
      })

      expect(cloneResult?.clones).toHaveLength(1)
    })

    it('should handle all operations failing', async () => {
      vi.mocked(workflowAPI.bulkEnable).mockResolvedValue({
        success_count: 0,
        failures: [
          { workflow_id: 'wf-1', error: 'Not found' },
          { workflow_id: 'wf-2', error: 'Not found' },
        ],
      })

      const { result } = renderHook(() => useBulkWorkflows())

      let enableResult
      await act(async () => {
        enableResult = await result.current.bulkEnable(['wf-1', 'wf-2'])
      })

      expect(enableResult?.success_count).toBe(0)
      expect(enableResult?.failures).toHaveLength(2)
    })
  })
})
