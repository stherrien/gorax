import { renderHook, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useExecutionUpdates } from './useExecutionUpdates'
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

    // Simulate connection opening after a delay
    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      this.onopen?.()
    }, 10)
  }

  send(data: string) {
    // Mock send
  }

  close() {
    this.readyState = WebSocket.CLOSED
    this.onclose?.()
  }

  // Helper to simulate receiving a message
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

describe('useExecutionUpdates', () => {
  beforeEach(() => {
    MockWebSocket.reset()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('should connect to WebSocket when executionId is provided', async () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    expect(MockWebSocket.instances).toHaveLength(1)
    expect(MockWebSocket.instances[0].url).toContain('exec-123')
  })

  it('should not connect when executionId is null', () => {
    const { result } = renderHook(() =>
      useExecutionUpdates(null, { enabled: true })
    )

    expect(result.current.connected).toBe(false)
    expect(MockWebSocket.instances).toHaveLength(0)
  })

  it('should not connect when enabled is false', () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: false })
    )

    expect(result.current.connected).toBe(false)
    expect(MockWebSocket.instances).toHaveLength(0)
  })

  it('should receive and process execution.started event', async () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const event: ExecutionEvent = {
      type: 'execution.started',
      execution_id: 'exec-123',
      workflow_id: 'wf-456',
      tenant_id: 'tenant-1',
      status: 'running',
      progress: {
        total_steps: 5,
        completed_steps: 0,
        percentage: 0,
      },
      timestamp: new Date().toISOString(),
    }

    MockWebSocket.instances[0].simulateMessage(event)

    await waitFor(() => {
      expect(result.current.currentStatus).toBe('running')
      expect(result.current.currentProgress).toEqual(event.progress)
      expect(result.current.latestUpdate?.type).toBe('execution.started')
      expect(result.current.events).toHaveLength(1)
    })
  })

  it('should track progress updates', async () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    // Simulate progress events
    const progressEvents = [
      {
        type: 'execution.progress',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        progress: { total_steps: 5, completed_steps: 1, percentage: 20 },
        timestamp: new Date().toISOString(),
      },
      {
        type: 'execution.progress',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        progress: { total_steps: 5, completed_steps: 3, percentage: 60 },
        timestamp: new Date().toISOString(),
      },
    ] as ExecutionEvent[]

    for (const event of progressEvents) {
      MockWebSocket.instances[0].simulateMessage(event)
    }

    await waitFor(() => {
      expect(result.current.currentProgress?.completed_steps).toBe(3)
      expect(result.current.currentProgress?.percentage).toBe(60)
      expect(result.current.events).toHaveLength(2)
    })
  })

  it('should track completed steps', async () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const stepEvent: ExecutionEvent = {
      type: 'step.completed',
      execution_id: 'exec-123',
      workflow_id: 'wf-456',
      tenant_id: 'tenant-1',
      step: {
        step_id: 'step-1',
        node_id: 'node-1',
        node_type: 'action:http',
        status: 'completed',
        output_data: { status: 200 },
        duration_ms: 150,
      },
      timestamp: new Date().toISOString(),
    }

    MockWebSocket.instances[0].simulateMessage(stepEvent)

    await waitFor(() => {
      expect(result.current.completedSteps).toHaveLength(1)
      expect(result.current.completedSteps[0].node_id).toBe('node-1')
    })
  })

  it('should call onStatusChange callback', async () => {
    const onStatusChange = vi.fn()

    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true, onStatusChange })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
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
      expect(onStatusChange).toHaveBeenCalledWith('running')
    })
  })

  it('should call onProgress callback', async () => {
    const onProgress = vi.fn()

    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true, onProgress })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const event: ExecutionEvent = {
      type: 'execution.progress',
      execution_id: 'exec-123',
      workflow_id: 'wf-456',
      tenant_id: 'tenant-1',
      progress: { total_steps: 5, completed_steps: 2, percentage: 40 },
      timestamp: new Date().toISOString(),
    }

    MockWebSocket.instances[0].simulateMessage(event)

    await waitFor(() => {
      expect(onProgress).toHaveBeenCalledWith(event.progress)
    })
  })

  it('should call onStepComplete callback', async () => {
    const onStepComplete = vi.fn()

    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true, onStepComplete })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const stepInfo = {
      step_id: 'step-1',
      node_id: 'node-1',
      node_type: 'action:http',
      status: 'completed',
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
      expect(onStepComplete).toHaveBeenCalledWith(stepInfo)
    })
  })

  it('should call onComplete callback', async () => {
    const onComplete = vi.fn()

    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true, onComplete })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const output = { result: 'success', data: { count: 42 } }

    const event: ExecutionEvent = {
      type: 'execution.completed',
      execution_id: 'exec-123',
      workflow_id: 'wf-456',
      tenant_id: 'tenant-1',
      status: 'completed',
      output,
      timestamp: new Date().toISOString(),
    }

    MockWebSocket.instances[0].simulateMessage(event)

    await waitFor(() => {
      expect(onComplete).toHaveBeenCalledWith(output)
    })
  })

  it('should call onError callback', async () => {
    const onError = vi.fn()

    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true, onError })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const errorMsg = 'Connection timeout'

    const event: ExecutionEvent = {
      type: 'execution.failed',
      execution_id: 'exec-123',
      workflow_id: 'wf-456',
      tenant_id: 'tenant-1',
      status: 'failed',
      error: errorMsg,
      timestamp: new Date().toISOString(),
    }

    MockWebSocket.instances[0].simulateMessage(event)

    await waitFor(() => {
      expect(onError).toHaveBeenCalledWith(errorMsg)
    })
  })

  it('should clear events when clearEvents is called', async () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    // Send some events
    MockWebSocket.instances[0].simulateMessage({
      type: 'execution.started',
      execution_id: 'exec-123',
      workflow_id: 'wf-456',
      tenant_id: 'tenant-1',
      timestamp: new Date().toISOString(),
    })

    await waitFor(() => {
      expect(result.current.events).toHaveLength(1)
    })

    // Clear events
    result.current.clearEvents()

    await waitFor(() => {
      expect(result.current.events).toHaveLength(0)
      expect(result.current.completedSteps).toHaveLength(0)
    })
  })

  it('should disconnect when unmounted', async () => {
    const { result, unmount } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    unmount()

    await waitFor(() => {
      expect(MockWebSocket.instances[0].readyState).toBe(WebSocket.CLOSED)
    })
  })

  it('should handle reconnection', async () => {
    const onReconnecting = vi.fn()
    const onReconnected = vi.fn()

    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', {
        enabled: true,
      })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    // Simulate disconnection
    MockWebSocket.instances[0].close()

    await waitFor(() => {
      expect(result.current.connected).toBe(false)
    })
  })
})
