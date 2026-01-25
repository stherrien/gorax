import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useFileUpload } from './useFileUpload'

describe('useFileUpload', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('initial state', () => {
    it('returns idle upload state', () => {
      const { result } = renderHook(() => useFileUpload())

      expect(result.current.uploadState.status).toBe('idle')
      expect(result.current.uploadState.progress).toBe(0)
      expect(result.current.isDragging).toBe(false)
    })

    it('provides fileInputRef', () => {
      const { result } = renderHook(() => useFileUpload())

      expect(result.current.fileInputRef).toBeDefined()
    })
  })

  describe('resetUpload', () => {
    it('resets upload state to initial', async () => {
      const { result } = renderHook(() => useFileUpload())

      // Create a mock file and start upload
      const content = JSON.stringify({
        nodes: [{ id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: {} }],
        edges: [],
      })
      const file = new File([content], 'workflow.json', { type: 'application/json' })

      await act(async () => {
        await result.current.uploadFile(file)
      })

      // Should be in success state
      expect(result.current.uploadState.status).toBe('success')

      // Reset
      act(() => {
        result.current.resetUpload()
      })

      expect(result.current.uploadState.status).toBe('idle')
      expect(result.current.uploadState.progress).toBe(0)
      expect(result.current.isDragging).toBe(false)
    })
  })

  describe('uploadFile', () => {
    it('successfully parses valid JSON workflow', async () => {
      const { result } = renderHook(() => useFileUpload())

      const content = JSON.stringify({
        nodes: [
          { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Start' } },
          { id: 'node-2', type: 'action', position: { x: 100, y: 100 }, data: { label: 'Action' } },
        ],
        edges: [
          { id: 'edge-1', source: 'node-1', target: 'node-2' },
        ],
      })
      const file = new File([content], 'workflow.json', { type: 'application/json' })

      await act(async () => {
        await result.current.uploadFile(file)
      })

      expect(result.current.uploadState.status).toBe('success')
      expect(result.current.uploadState.nodes).toHaveLength(2)
      expect(result.current.uploadState.edges).toHaveLength(1)
      expect(result.current.uploadState.progress).toBe(100)
    })

    it('rejects unsupported file types', async () => {
      const { result } = renderHook(() => useFileUpload())

      const file = new File(['content'], 'workflow.txt', { type: 'text/plain' })

      await act(async () => {
        await result.current.uploadFile(file)
      })

      expect(result.current.uploadState.status).toBe('error')
      expect(result.current.uploadState.error).toContain('Unsupported file type')
    })

    it('rejects files without extension', async () => {
      const { result } = renderHook(() => useFileUpload())

      const file = new File(['content'], 'workflow')

      await act(async () => {
        await result.current.uploadFile(file)
      })

      expect(result.current.uploadState.status).toBe('error')
      expect(result.current.uploadState.error).toContain('no extension')
    })

    it('rejects invalid JSON content', async () => {
      const { result } = renderHook(() => useFileUpload())

      const file = new File(['{ invalid json }'], 'workflow.json', { type: 'application/json' })

      await act(async () => {
        await result.current.uploadFile(file)
      })

      expect(result.current.uploadState.status).toBe('error')
      expect(result.current.uploadState.error).toContain('Invalid JSON')
    })

    it('calls onUploadStart callback', async () => {
      const onUploadStart = vi.fn()
      const { result } = renderHook(() => useFileUpload({ onUploadStart }))

      const content = JSON.stringify({
        nodes: [{ id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: {} }],
        edges: [],
      })
      const file = new File([content], 'workflow.json')

      await act(async () => {
        await result.current.uploadFile(file)
      })

      expect(onUploadStart).toHaveBeenCalled()
    })

    it('calls onUploadComplete callback on success', async () => {
      const onUploadComplete = vi.fn()
      const { result } = renderHook(() => useFileUpload({ onUploadComplete }))

      const content = JSON.stringify({
        nodes: [{ id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: {} }],
        edges: [],
      })
      const file = new File([content], 'workflow.json')

      await act(async () => {
        await result.current.uploadFile(file)
      })

      expect(onUploadComplete).toHaveBeenCalled()
      expect(onUploadComplete.mock.calls[0][0].status).toBe('success')
    })

    it('calls onUploadError callback on failure', async () => {
      const onUploadError = vi.fn()
      const { result } = renderHook(() => useFileUpload({ onUploadError }))

      const file = new File(['invalid'], 'workflow.txt')

      await act(async () => {
        await result.current.uploadFile(file)
      })

      expect(onUploadError).toHaveBeenCalled()
    })

    it('captures warnings from parser', async () => {
      const { result } = renderHook(() => useFileUpload())

      // Workflow without positions will generate warnings
      const content = JSON.stringify({
        nodes: [
          { id: 'node-1', type: 'action' },
        ],
        edges: [],
      })
      const file = new File([content], 'workflow.json')

      await act(async () => {
        await result.current.uploadFile(file)
      })

      expect(result.current.uploadState.status).toBe('success')
      expect(result.current.uploadState.warnings).toBeDefined()
      expect(result.current.uploadState.warnings?.length).toBeGreaterThan(0)
    })
  })

  describe('acceptUpload', () => {
    it('returns workflow data and resets state', async () => {
      const { result } = renderHook(() => useFileUpload())

      const content = JSON.stringify({
        name: 'My Workflow',
        description: 'Test workflow',
        nodes: [{ id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Start' } }],
        edges: [],
      })
      const file = new File([content], 'workflow.json')

      await act(async () => {
        await result.current.uploadFile(file)
      })

      let accepted: ReturnType<typeof result.current.acceptUpload>
      act(() => {
        accepted = result.current.acceptUpload()
      })

      expect(accepted).not.toBeNull()
      expect(accepted?.nodes).toHaveLength(1)
      expect(accepted?.name).toBe('My Workflow')
      expect(result.current.uploadState.status).toBe('idle')
    })

    it('returns null if not in success state', () => {
      const { result } = renderHook(() => useFileUpload())

      const accepted = result.current.acceptUpload()

      expect(accepted).toBeNull()
    })
  })

  describe('drag events', () => {
    it('sets isDragging true on dragEnter', () => {
      const { result } = renderHook(() => useFileUpload())

      const event = {
        preventDefault: vi.fn(),
        stopPropagation: vi.fn(),
        dataTransfer: { types: ['Files'] },
      } as unknown as React.DragEvent

      act(() => {
        result.current.handleDragEnter(event)
      })

      expect(result.current.isDragging).toBe(true)
    })

    it('sets isDragging false on dragLeave when counter reaches 0', () => {
      const { result } = renderHook(() => useFileUpload())

      const event = {
        preventDefault: vi.fn(),
        stopPropagation: vi.fn(),
        dataTransfer: { types: ['Files'] },
      } as unknown as React.DragEvent

      act(() => {
        result.current.handleDragEnter(event)
      })

      expect(result.current.isDragging).toBe(true)

      act(() => {
        result.current.handleDragLeave(event)
      })

      expect(result.current.isDragging).toBe(false)
    })

    it('handleDragOver prevents default and sets dropEffect', () => {
      const { result } = renderHook(() => useFileUpload())

      const event = {
        preventDefault: vi.fn(),
        stopPropagation: vi.fn(),
        dataTransfer: { dropEffect: '' },
      } as unknown as React.DragEvent

      act(() => {
        result.current.handleDragOver(event)
      })

      expect(event.preventDefault).toHaveBeenCalled()
      expect(event.dataTransfer.dropEffect).toBe('copy')
    })
  })

  describe('handleDrop', () => {
    it('processes dropped files', async () => {
      const { result } = renderHook(() => useFileUpload())

      const content = JSON.stringify({
        nodes: [{ id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: {} }],
        edges: [],
      })
      const file = new File([content], 'workflow.json')

      const event = {
        preventDefault: vi.fn(),
        stopPropagation: vi.fn(),
        dataTransfer: {
          files: [file],
        },
      } as unknown as React.DragEvent

      await act(async () => {
        await result.current.handleDrop(event)
      })

      expect(result.current.uploadState.status).toBe('success')
      expect(result.current.isDragging).toBe(false)
    })

    it('does nothing when no files dropped', async () => {
      const { result } = renderHook(() => useFileUpload())

      const event = {
        preventDefault: vi.fn(),
        stopPropagation: vi.fn(),
        dataTransfer: {
          files: [],
        },
      } as unknown as React.DragEvent

      await act(async () => {
        await result.current.handleDrop(event)
      })

      expect(result.current.uploadState.status).toBe('idle')
    })
  })

  describe('clearError', () => {
    it('clears error state', async () => {
      const { result } = renderHook(() => useFileUpload())

      const file = new File(['invalid'], 'workflow.txt')

      await act(async () => {
        await result.current.uploadFile(file)
      })

      expect(result.current.uploadState.status).toBe('error')

      act(() => {
        result.current.clearError()
      })

      expect(result.current.uploadState.status).toBe('idle')
      expect(result.current.uploadState.error).toBeUndefined()
    })
  })
})
