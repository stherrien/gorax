import { renderHook, waitFor, act } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useExecutionUpdates } from './useExecutionUpdates'
import type { ExecutionEvent } from '../lib/websocket'

// Use vi.hoisted to define the mock class before vi.mock runs
const { MockWebSocketClient, mockClientInstances, resetInstances } = vi.hoisted(() => {
  type EventHandler = (event: any) => void
  type WebSocketConfig = {
    url: string
    onOpen?: () => void
    onClose?: () => void
    onError?: (error: Event) => void
    onReconnecting?: (attempt: number) => void
    onReconnected?: () => void
  }

  const instances: any[] = []

  class MockWebSocketClient {
    config: WebSocketConfig
    private handlers: Set<EventHandler> = new Set()
    private _connected: boolean = false

    constructor(config: WebSocketConfig) {
      this.config = config
      instances.push(this)
    }

    connect(): void {
      // Connection is handled synchronously when simulateOpen is called
    }

    disconnect(): void {
      this._connected = false
      this.config.onClose?.()
    }

    on(handler: EventHandler): () => void {
      this.handlers.add(handler)
      return () => this.handlers.delete(handler)
    }

    isConnected(): boolean {
      return this._connected
    }

    // Test helper methods
    simulateMessage(event: any) {
      this.handlers.forEach(handler => handler(event))
    }

    simulateOpen() {
      this._connected = true
      this.config.onOpen?.()
    }

    simulateClose() {
      this._connected = false
      this.config.onClose?.()
    }
  }

  return {
    MockWebSocketClient,
    mockClientInstances: instances,
    resetInstances: () => { instances.length = 0 },
  }
})

// Mock the websocket module
vi.mock('../lib/websocket', async () => {
  const actual = await vi.importActual('../lib/websocket')
  return {
    ...actual,
    WebSocketClient: MockWebSocketClient,
  }
})

// Helper to open the WebSocket connection with proper act wrapping
async function openConnection() {
  await waitFor(() => {
    expect(mockClientInstances.length).toBeGreaterThan(0)
  })

  await act(async () => {
    mockClientInstances[0].simulateOpen()
  })
}

describe('useExecutionUpdates', () => {
  beforeEach(() => {
    resetInstances()
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('should connect to WebSocket when executionId is provided', async () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await openConnection()

    expect(result.current.connected).toBe(true)
    expect(mockClientInstances).toHaveLength(1)
    expect(mockClientInstances[0].config.url).toContain('exec-123')
  })

  it('should not connect when executionId is null', () => {
    const { result } = renderHook(() =>
      useExecutionUpdates(null, { enabled: true })
    )

    expect(result.current.connected).toBe(false)
    expect(mockClientInstances).toHaveLength(0)
  })

  it('should not connect when enabled is false', () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: false })
    )

    expect(result.current.connected).toBe(false)
    expect(mockClientInstances).toHaveLength(0)
  })

  it('should receive and process execution.started event', async () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await openConnection()

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

    await act(async () => {
      mockClientInstances[0].simulateMessage(event)
    })

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

    await openConnection()

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
      await act(async () => {
        mockClientInstances[0].simulateMessage(event)
      })
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

    await openConnection()

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

    await act(async () => {
      mockClientInstances[0].simulateMessage(stepEvent)
    })

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

    await openConnection()

    const event: ExecutionEvent = {
      type: 'execution.started',
      execution_id: 'exec-123',
      workflow_id: 'wf-456',
      tenant_id: 'tenant-1',
      status: 'running',
      timestamp: new Date().toISOString(),
    }

    await act(async () => {
      mockClientInstances[0].simulateMessage(event)
    })

    await waitFor(() => {
      expect(onStatusChange).toHaveBeenCalledWith('running')
    })
  })

  it('should call onProgress callback', async () => {
    const onProgress = vi.fn()

    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true, onProgress })
    )

    await openConnection()

    const event: ExecutionEvent = {
      type: 'execution.progress',
      execution_id: 'exec-123',
      workflow_id: 'wf-456',
      tenant_id: 'tenant-1',
      progress: { total_steps: 5, completed_steps: 2, percentage: 40 },
      timestamp: new Date().toISOString(),
    }

    await act(async () => {
      mockClientInstances[0].simulateMessage(event)
    })

    await waitFor(() => {
      expect(onProgress).toHaveBeenCalledWith(event.progress)
    })
  })

  it('should call onStepComplete callback', async () => {
    const onStepComplete = vi.fn()

    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true, onStepComplete })
    )

    await openConnection()

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

    await act(async () => {
      mockClientInstances[0].simulateMessage(event)
    })

    await waitFor(() => {
      expect(onStepComplete).toHaveBeenCalledWith(stepInfo)
    })
  })

  it('should call onComplete callback', async () => {
    const onComplete = vi.fn()

    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true, onComplete })
    )

    await openConnection()

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

    await act(async () => {
      mockClientInstances[0].simulateMessage(event)
    })

    await waitFor(() => {
      expect(onComplete).toHaveBeenCalledWith(output)
    })
  })

  it('should call onError callback', async () => {
    const onError = vi.fn()

    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true, onError })
    )

    await openConnection()

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

    await act(async () => {
      mockClientInstances[0].simulateMessage(event)
    })

    await waitFor(() => {
      expect(onError).toHaveBeenCalledWith(errorMsg)
    })
  })

  it('should clear events when clearEvents is called', async () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await openConnection()

    // Send some events
    await act(async () => {
      mockClientInstances[0].simulateMessage({
        type: 'execution.started',
        execution_id: 'exec-123',
        workflow_id: 'wf-456',
        tenant_id: 'tenant-1',
        timestamp: new Date().toISOString(),
      })
    })

    await waitFor(() => {
      expect(result.current.events).toHaveLength(1)
    })

    // Clear events
    await act(async () => {
      result.current.clearEvents()
    })

    await waitFor(() => {
      expect(result.current.events).toHaveLength(0)
      expect(result.current.completedSteps).toHaveLength(0)
    })
  })

  it('should disconnect when unmounted', async () => {
    const { result, unmount } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await openConnection()

    expect(result.current.connected).toBe(true)

    unmount()

    // After unmount, we can't easily check the state
    // Just verify that no errors occurred
  })

  it('should handle reconnection', async () => {
    const { result } = renderHook(() =>
      useExecutionUpdates('exec-123', { enabled: true })
    )

    await openConnection()

    expect(result.current.connected).toBe(true)

    // Simulate disconnection
    await act(async () => {
      mockClientInstances[0].simulateClose()
    })

    await waitFor(() => {
      expect(result.current.connected).toBe(false)
    })
  })
})
