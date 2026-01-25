import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useWorkflowRecovery, useAutoSaveWorkflow } from './useWorkflowRecovery'
import type { Node, Edge } from '@xyflow/react'

// Mock localStorage
const mockLocalStorage = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key]
    }),
    clear: vi.fn(() => {
      store = {}
    }),
    get length() {
      return Object.keys(store).length
    },
    key: vi.fn((i: number) => Object.keys(store)[i] ?? null),
  }
})()

Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

describe('useWorkflowRecovery', () => {
  beforeEach(() => {
    mockLocalStorage.clear()
    vi.clearAllMocks()
  })

  it('returns initial state with no backup', () => {
    const { result } = renderHook(() => useWorkflowRecovery('test-workflow'))

    expect(result.current.hasBackup).toBe(false)
    expect(result.current.lastBackupTime).toBeNull()
    expect(result.current.isRecovering).toBe(false)
    expect(result.current.recoveryError).toBeNull()
  })

  it('saves backup to localStorage', () => {
    const { result } = renderHook(() => useWorkflowRecovery('test-workflow'))

    const nodes: Node[] = [
      { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Start' } },
    ]
    const edges: Edge[] = []

    act(() => {
      result.current.saveBackup('test-workflow', nodes, edges, 'Test Workflow', 'Description')
    })

    expect(mockLocalStorage.setItem).toHaveBeenCalled()
    expect(result.current.hasBackup).toBe(true)
    expect(result.current.lastBackupTime).toBeInstanceOf(Date)
  })

  it('gets backup from localStorage', () => {
    const backup = {
      id: 'test-workflow',
      timestamp: Date.now(),
      nodes: [{ id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: {} }],
      edges: [],
      name: 'Test Workflow',
      description: 'Test description',
    }
    mockLocalStorage.setItem('gorax_workflow_backup_test-workflow', JSON.stringify(backup))

    const { result } = renderHook(() => useWorkflowRecovery('test-workflow'))

    const retrieved = result.current.getBackup('test-workflow')

    expect(retrieved).not.toBeNull()
    expect(retrieved?.id).toBe('test-workflow')
    expect(retrieved?.name).toBe('Test Workflow')
  })

  it('recovers from backup', () => {
    const backup = {
      id: 'test-workflow',
      timestamp: Date.now(),
      nodes: [{ id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: {} }],
      edges: [],
      name: 'Recovered Workflow',
      description: 'Recovered description',
    }
    mockLocalStorage.setItem('gorax_workflow_backup_test-workflow', JSON.stringify(backup))

    const { result } = renderHook(() => useWorkflowRecovery('test-workflow'))

    let recovered
    act(() => {
      recovered = result.current.recoverFromBackup('test-workflow')
    })

    expect(recovered).not.toBeNull()
    expect(recovered?.name).toBe('Recovered Workflow')
    expect(result.current.isRecovering).toBe(false)
    expect(result.current.recoveryError).toBeNull()
  })

  it('returns error when no backup found for recovery', () => {
    const { result } = renderHook(() => useWorkflowRecovery('test-workflow'))

    let recovered
    act(() => {
      recovered = result.current.recoverFromBackup('nonexistent-workflow')
    })

    expect(recovered).toBeNull()
    expect(result.current.recoveryError).toBe('No backup found for this workflow')
  })

  it('clears backup from localStorage', () => {
    const backup = {
      id: 'test-workflow',
      timestamp: Date.now(),
      nodes: [],
      edges: [],
      name: 'Test',
      description: '',
    }
    mockLocalStorage.setItem('gorax_workflow_backup_test-workflow', JSON.stringify(backup))

    const { result } = renderHook(() => useWorkflowRecovery('test-workflow'))

    act(() => {
      result.current.clearBackup('test-workflow')
    })

    expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('gorax_workflow_backup_test-workflow')
    expect(result.current.hasBackup).toBe(false)
  })

  it('returns all backups sorted by timestamp', () => {
    const backup1 = {
      id: 'workflow-1',
      timestamp: Date.now() - 1000,
      nodes: [],
      edges: [],
      name: 'First',
      description: '',
    }
    const backup2 = {
      id: 'workflow-2',
      timestamp: Date.now(),
      nodes: [],
      edges: [],
      name: 'Second',
      description: '',
    }

    mockLocalStorage.setItem('gorax_workflow_backup_workflow-1', JSON.stringify(backup1))
    mockLocalStorage.setItem('gorax_workflow_backup_workflow-2', JSON.stringify(backup2))

    const { result } = renderHook(() => useWorkflowRecovery('workflow-1'))

    const allBackups = result.current.getAllBackups()

    expect(allBackups).toHaveLength(2)
    expect(allBackups[0].name).toBe('Second') // Most recent first
    expect(allBackups[1].name).toBe('First')
  })

  it('handles invalid JSON in localStorage gracefully', () => {
    mockLocalStorage.setItem('gorax_workflow_backup_test-workflow', 'invalid json')

    const { result } = renderHook(() => useWorkflowRecovery('test-workflow'))

    const backup = result.current.getBackup('test-workflow')

    expect(backup).toBeNull()
  })

  it('handles missing required fields in backup', () => {
    const invalidBackup = {
      timestamp: Date.now(),
      // Missing id, nodes, edges
    }
    mockLocalStorage.setItem('gorax_workflow_backup_test-workflow', JSON.stringify(invalidBackup))

    const { result } = renderHook(() => useWorkflowRecovery('test-workflow'))

    const backup = result.current.getBackup('test-workflow')

    expect(backup).toBeNull()
  })
})

describe('useAutoSaveWorkflow', () => {
  beforeEach(() => {
    mockLocalStorage.clear()
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns initial state', () => {
    const nodes: Node[] = []
    const edges: Edge[] = []

    const { result } = renderHook(() =>
      useAutoSaveWorkflow('test-workflow', nodes, edges, 'Test', 'Description')
    )

    expect(result.current.lastSaved).toBeNull()
    expect(result.current.isSaving).toBe(false)
  })

  it('auto-saves after interval when data changes', () => {
    const nodes: Node[] = [
      { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Start' } },
    ]
    const edges: Edge[] = []

    const { result, rerender } = renderHook(
      ({ nodes, name }) =>
        useAutoSaveWorkflow('test-workflow', nodes, edges, name, 'Description', {
          interval: 1000,
        }),
      { initialProps: { nodes: [], name: 'Initial' } }
    )

    // Update with new nodes
    rerender({ nodes, name: 'Updated' })

    // Advance timer past the interval
    act(() => {
      vi.advanceTimersByTime(1500)
    })

    // Should have saved
    expect(mockLocalStorage.setItem).toHaveBeenCalled()
  })

  it('does not auto-save when disabled', () => {
    const nodes: Node[] = [
      { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Start' } },
    ]
    const edges: Edge[] = []

    renderHook(() =>
      useAutoSaveWorkflow('test-workflow', nodes, edges, 'Test', 'Description', {
        enabled: false,
        interval: 1000,
      })
    )

    act(() => {
      vi.advanceTimersByTime(5000)
    })

    expect(mockLocalStorage.setItem).not.toHaveBeenCalled()
  })

  it('does not auto-save when workflowId is null', () => {
    const nodes: Node[] = [
      { id: 'node-1', type: 'trigger', position: { x: 0, y: 0 }, data: { label: 'Start' } },
    ]
    const edges: Edge[] = []

    renderHook(() =>
      useAutoSaveWorkflow(null, nodes, edges, 'Test', 'Description', {
        interval: 1000,
      })
    )

    act(() => {
      vi.advanceTimersByTime(5000)
    })

    expect(mockLocalStorage.setItem).not.toHaveBeenCalled()
  })
})
