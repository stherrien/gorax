import { renderHook, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useExecutionTrace } from './useExecutionTrace'
import { useExecutionTraceStore } from '../stores/executionTraceStore'
import type { ExecutionEvent } from '../lib/websocket'

// Mock WebSocket
class MockWebSocket {
  static instances: MockWebSocket[] = []

  url: string
  onopen: (() => void) | null = null
  onclose: (() => void) | null = null
  onerror: ((error: Event) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  readyState: number = WebSocket.CONNECTING

  constructor(url: string) {
    this.url = url
    MockWebSocket.instances.push(this)

    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      this.onopen?.()
    }, 10)
  }

  close() {
    this.readyState = WebSocket.CLOSED
    this.onclose?.()
  }

  simulateMessage(data: any) {
    const event = new MessageEvent('message', {
      data: JSON.stringify(data),
    })
    this.onmessage?.(event)
  }

  static reset() {
    MockWebSocket.instances = []
  }
}

// @ts-ignore
global.WebSocket = MockWebSocket

// Mock the store
vi.mock('../stores/executionTraceStore', () => ({
  useExecutionTraceStore: {
    getState: vi.fn(),
  },
}))

describe('useExecutionTrace', () => {
  let mockStore: any

  beforeEach(() => {
    MockWebSocket.reset()

    mockStore = {
      setCurrentExecutionId: vi.fn(),
      setNodeStatus: vi.fn(),
      addStepLog: vi.fn(),
      setEdgeAnimated: vi.fn(),
      addTimelineEvent: vi.fn(),
      reset: vi.fn(),
    }

    vi.mocked(useExecutionTraceStore.getState).mockReturnValue(mockStore)
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('Connection management', () => {
    it('should connect when execution ID is provided', async () => {
      const { result } = renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      expect(MockWebSocket.instances).toHaveLength(1)
    })

    it('should not connect when execution ID is null', () => {
      const { result } = renderHook(() => useExecutionTrace(null))

      expect(result.current.connected).toBe(false)
      expect(MockWebSocket.instances).toHaveLength(0)
    })

    it('should set current execution ID on connect', async () => {
      renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(mockStore.setCurrentExecutionId).toHaveBeenCalledWith('exec-123')
      })
    })

    it('should disconnect and reset store on unmount', async () => {
      const { unmount } = renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.OPEN)
      })

      unmount()

      await waitFor(() => {
        expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.CLOSED)
        expect(mockStore.reset).toHaveBeenCalled()
      })
    })

    it('should update execution ID when it changes', async () => {
      const { rerender } = renderHook(
        ({ execId }) => useExecutionTrace(execId),
        { initialProps: { execId: 'exec-123' } }
      )

      await waitFor(() => {
        expect(mockStore.setCurrentExecutionId).toHaveBeenCalledWith('exec-123')
      })

      rerender({ execId: 'exec-456' })

      await waitFor(() => {
        expect(mockStore.setCurrentExecutionId).toHaveBeenCalledWith('exec-456')
      })
    })
  })

  describe('Event handling - execution.started', () => {
    it('should update node status on execution.started', async () => {
      renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.OPEN)
      })

      const event: ExecutionEvent = {
        type: 'execution.started',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        status: 'running',
        timestamp: new Date().toISOString(),
      }

      MockWebSocket.instances[0].simulateMessage(event)

      await waitFor(() => {
        expect(mockStore.addTimelineEvent).toHaveBeenCalledWith(
          expect.objectContaining({
            type: 'started',
            message: expect.any(String),
          })
        )
      })
    })
  })

  describe('Event handling - step.started', () => {
    it('should set node status to running on step.started', async () => {
      renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.OPEN)
      })

      const event: ExecutionEvent = {
        type: 'step.started',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        step: {
          step_id: 'step-1',
          node_id: 'node-1',
          node_type: 'action:http',
          status: 'running',
        },
        timestamp: new Date().toISOString(),
      }

      MockWebSocket.instances[0].simulateMessage(event)

      await waitFor(() => {
        expect(mockStore.setNodeStatus).toHaveBeenCalledWith('node-1', 'running')
      })
    })

    it('should add timeline event on step.started', async () => {
      renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.OPEN)
      })

      const event: ExecutionEvent = {
        type: 'step.started',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        step: {
          step_id: 'step-1',
          node_id: 'node-1',
          node_type: 'action:http',
          status: 'running',
        },
        timestamp: new Date().toISOString(),
      }

      MockWebSocket.instances[0].simulateMessage(event)

      await waitFor(() => {
        expect(mockStore.addTimelineEvent).toHaveBeenCalledWith(
          expect.objectContaining({
            nodeId: 'node-1',
            type: 'started',
          })
        )
      })
    })
  })

  describe('Event handling - step.completed', () => {
    it('should set node status to completed on step.completed', async () => {
      renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.OPEN)
      })

      const event: ExecutionEvent = {
        type: 'step.completed',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        step: {
          step_id: 'step-1',
          node_id: 'node-1',
          node_type: 'action:http',
          status: 'completed',
          output_data: { result: 'success' },
          duration_ms: 1500,
        },
        timestamp: new Date().toISOString(),
      }

      MockWebSocket.instances[0].simulateMessage(event)

      await waitFor(() => {
        expect(mockStore.setNodeStatus).toHaveBeenCalledWith('node-1', 'completed')
      })
    })

    it('should add step log on step.completed', async () => {
      renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.OPEN)
      })

      const stepInfo = {
        step_id: 'step-1',
        node_id: 'node-1',
        node_type: 'action:http',
        status: 'completed',
        output_data: { result: 'success' },
      }

      const event: ExecutionEvent = {
        type: 'step.completed',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        step: stepInfo,
        timestamp: new Date().toISOString(),
      }

      MockWebSocket.instances[0].simulateMessage(event)

      await waitFor(() => {
        expect(mockStore.addStepLog).toHaveBeenCalledWith('node-1', stepInfo)
      })
    })
  })

  describe('Event handling - step.failed', () => {
    it('should set node status to failed on step.failed', async () => {
      renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.OPEN)
      })

      const event: ExecutionEvent = {
        type: 'step.failed',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        step: {
          step_id: 'step-1',
          node_id: 'node-1',
          node_type: 'action:http',
          status: 'failed',
          error: 'Connection timeout',
        },
        timestamp: new Date().toISOString(),
      }

      MockWebSocket.instances[0].simulateMessage(event)

      await waitFor(() => {
        expect(mockStore.setNodeStatus).toHaveBeenCalledWith('node-1', 'failed')
      })
    })

    it('should add timeline event with error info on step.failed', async () => {
      renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.OPEN)
      })

      const event: ExecutionEvent = {
        type: 'step.failed',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        step: {
          step_id: 'step-1',
          node_id: 'node-1',
          node_type: 'action:http',
          status: 'failed',
          error: 'Connection timeout',
        },
        timestamp: new Date().toISOString(),
      }

      MockWebSocket.instances[0].simulateMessage(event)

      await waitFor(() => {
        expect(mockStore.addTimelineEvent).toHaveBeenCalledWith(
          expect.objectContaining({
            type: 'failed',
            nodeId: 'node-1',
            message: expect.stringContaining('Connection timeout'),
          })
        )
      })
    })
  })

  describe('Event handling - execution.progress', () => {
    it('should add progress timeline event', async () => {
      renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.OPEN)
      })

      const event: ExecutionEvent = {
        type: 'execution.progress',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        progress: {
          total_steps: 5,
          completed_steps: 3,
          percentage: 60,
        },
        timestamp: new Date().toISOString(),
      }

      MockWebSocket.instances[0].simulateMessage(event)

      await waitFor(() => {
        expect(mockStore.addTimelineEvent).toHaveBeenCalledWith(
          expect.objectContaining({
            type: 'progress',
            message: expect.stringContaining('60%'),
          })
        )
      })
    })
  })

  describe('Connection state', () => {
    it('should expose connected state', async () => {
      const { result } = renderHook(() => useExecutionTrace('exec-123'))

      expect(result.current.connected).toBe(false)

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })
    })

    it('should expose reconnecting state', async () => {
      const { result } = renderHook(() => useExecutionTrace('exec-123'))

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      // Initially not reconnecting
      expect(result.current.reconnecting).toBe(false)
    })
  })
})
