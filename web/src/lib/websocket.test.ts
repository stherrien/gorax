import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  WebSocketClient,
  createExecutionWebSocketURL,
  createWorkflowWebSocketURL,
  createTenantWebSocketURL,
  type ExecutionEvent,
  type EventHandler,
} from './websocket'

// Store instances created for testing
let mockWebSocketInstances: MockWebSocket[] = []
let webSocketConstructorSpy: ReturnType<typeof vi.fn>

// Mock WebSocket class
class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  url: string
  readyState: number = MockWebSocket.CONNECTING
  onopen: (() => void) | null = null
  onclose: (() => void) | null = null
  onerror: ((error: Event) => void) | null = null
  onmessage: ((event: { data: string }) => void) | null = null

  constructor(url: string) {
    this.url = url
    mockWebSocketInstances.push(this)
    webSocketConstructorSpy(url)
  }

  close() {
    this.readyState = MockWebSocket.CLOSED
    if (this.onclose) {
      this.onclose()
    }
  }

  // Helper to simulate connection
  simulateOpen() {
    this.readyState = MockWebSocket.OPEN
    if (this.onopen) {
      this.onopen()
    }
  }

  // Helper to simulate message
  simulateMessage(data: ExecutionEvent) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(data) })
    }
  }

  // Helper to simulate error
  simulateError(error: Event) {
    if (this.onerror) {
      this.onerror(error)
    }
  }

  // Helper to simulate close
  simulateClose() {
    this.readyState = MockWebSocket.CLOSED
    if (this.onclose) {
      this.onclose()
    }
  }
}

// Helper to get the latest instance
function getLatestMockWebSocket(): MockWebSocket | undefined {
  return mockWebSocketInstances[mockWebSocketInstances.length - 1]
}

describe('websocket utilities', () => {
  let originalWebSocket: typeof WebSocket

  beforeEach(() => {
    originalWebSocket = global.WebSocket
    mockWebSocketInstances = []
    webSocketConstructorSpy = vi.fn()

    // Replace global WebSocket with our mock class
    // @ts-expect-error - Mocking WebSocket
    global.WebSocket = MockWebSocket

    vi.useFakeTimers()
  })

  afterEach(() => {
    global.WebSocket = originalWebSocket
    mockWebSocketInstances = []
    vi.useRealTimers()
    vi.clearAllMocks()
  })

  describe('WebSocketClient', () => {
    describe('connect', () => {
      it('should create WebSocket connection with configured URL', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })

        client.connect()

        expect(webSocketConstructorSpy).toHaveBeenCalledWith('ws://localhost:8080/ws')
      })

      it('should not create new connection if already connected', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        client.connect() // Second call

        expect(webSocketConstructorSpy).toHaveBeenCalledTimes(1)
      })

      it('should call onOpen callback when connection opens', () => {
        const onOpen = vi.fn()
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
          onOpen,
        })

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()

        expect(onOpen).toHaveBeenCalledTimes(1)
      })

      it('should call onClose callback when connection closes', () => {
        const onClose = vi.fn()
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
          onClose,
        })

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        getLatestMockWebSocket()?.simulateClose()

        expect(onClose).toHaveBeenCalledTimes(1)
      })

      it('should call onError callback on connection error', () => {
        const onError = vi.fn()
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
          onError,
        })

        client.connect()
        const errorEvent = new Event('error')
        getLatestMockWebSocket()?.simulateError(errorEvent)

        expect(onError).toHaveBeenCalledWith(errorEvent)
      })
    })

    describe('disconnect', () => {
      it('should close WebSocket connection', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        client.disconnect()

        expect(getLatestMockWebSocket()?.readyState).toBe(MockWebSocket.CLOSED)
      })

      it('should clear reconnect timer on disconnect', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
          reconnectDelay: 1000,
        })

        client.connect()
        getLatestMockWebSocket()?.simulateClose() // Triggers reconnect scheduling
        client.disconnect()

        // Advance timers - should not attempt reconnection
        vi.advanceTimersByTime(5000)
        expect(webSocketConstructorSpy).toHaveBeenCalledTimes(1)
      })
    })

    describe('isConnected', () => {
      it('should return true when WebSocket is open', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()

        expect(client.isConnected()).toBe(true)
      })

      it('should return false when WebSocket is not connected', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })

        expect(client.isConnected()).toBe(false)
      })

      it('should return false when WebSocket is closed', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        client.disconnect()

        expect(client.isConnected()).toBe(false)
      })
    })

    describe('on (event subscription)', () => {
      it('should register event handler', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })
        const handler: EventHandler = vi.fn()

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        client.on(handler)

        const mockEvent: ExecutionEvent = {
          type: 'execution.started',
          execution_id: 'exec-123',
          workflow_id: 'wf-456',
          tenant_id: 'tenant-789',
          timestamp: new Date().toISOString(),
        }

        getLatestMockWebSocket()?.simulateMessage(mockEvent)

        expect(handler).toHaveBeenCalledWith(mockEvent)
      })

      it('should support multiple handlers', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })
        const handler1: EventHandler = vi.fn()
        const handler2: EventHandler = vi.fn()

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        client.on(handler1)
        client.on(handler2)

        const mockEvent: ExecutionEvent = {
          type: 'execution.completed',
          execution_id: 'exec-123',
          workflow_id: 'wf-456',
          tenant_id: 'tenant-789',
          timestamp: new Date().toISOString(),
        }

        getLatestMockWebSocket()?.simulateMessage(mockEvent)

        expect(handler1).toHaveBeenCalledWith(mockEvent)
        expect(handler2).toHaveBeenCalledWith(mockEvent)
      })

      it('should return unsubscribe function', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })
        const handler: EventHandler = vi.fn()

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        const unsubscribe = client.on(handler)

        // Unsubscribe
        unsubscribe()

        const mockEvent: ExecutionEvent = {
          type: 'step.started',
          execution_id: 'exec-123',
          workflow_id: 'wf-456',
          tenant_id: 'tenant-789',
          timestamp: new Date().toISOString(),
        }

        getLatestMockWebSocket()?.simulateMessage(mockEvent)

        expect(handler).not.toHaveBeenCalled()
      })

      it('should handle errors in event handlers gracefully', () => {
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })
        const errorHandler: EventHandler = vi.fn(() => {
          throw new Error('Handler error')
        })
        const normalHandler: EventHandler = vi.fn()

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        client.on(errorHandler)
        client.on(normalHandler)

        const mockEvent: ExecutionEvent = {
          type: 'execution.failed',
          execution_id: 'exec-123',
          workflow_id: 'wf-456',
          tenant_id: 'tenant-789',
          timestamp: new Date().toISOString(),
        }

        getLatestMockWebSocket()?.simulateMessage(mockEvent)

        // Error should be logged
        expect(consoleErrorSpy).toHaveBeenCalled()
        // Other handler should still be called
        expect(normalHandler).toHaveBeenCalledWith(mockEvent)

        consoleErrorSpy.mockRestore()
      })

      it('should handle malformed messages gracefully', () => {
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })
        const handler: EventHandler = vi.fn()

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        client.on(handler)

        // Send malformed JSON
        const mockWs = getLatestMockWebSocket()
        if (mockWs?.onmessage) {
          mockWs.onmessage({ data: 'invalid json' })
        }

        expect(consoleErrorSpy).toHaveBeenCalled()
        expect(handler).not.toHaveBeenCalled()

        consoleErrorSpy.mockRestore()
      })
    })

    describe('automatic reconnection', () => {
      it('should attempt reconnection after unexpected close', () => {
        const onReconnecting = vi.fn()
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
          reconnectDelay: 3000,
          onReconnecting,
        })

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        getLatestMockWebSocket()?.simulateClose()

        expect(onReconnecting).toHaveBeenCalledWith(1)

        // Advance timer to trigger reconnect
        vi.advanceTimersByTime(3000)

        expect(webSocketConstructorSpy).toHaveBeenCalledTimes(2)
      })

      it('should not reconnect after intentional disconnect', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
          reconnectDelay: 1000,
        })

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        client.disconnect()

        // Advance timer - should not reconnect
        vi.advanceTimersByTime(5000)

        expect(webSocketConstructorSpy).toHaveBeenCalledTimes(1)
      })

      it('should stop reconnecting after max attempts', () => {
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
        const onReconnecting = vi.fn()
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
          reconnectDelay: 1000,
          maxReconnectAttempts: 3,
          onReconnecting,
        })

        client.connect()
        const firstInstance = getLatestMockWebSocket()
        firstInstance?.simulateOpen()

        // Simulate first disconnect (attempt 1)
        firstInstance?.simulateClose()
        expect(onReconnecting).toHaveBeenCalledWith(1)

        vi.advanceTimersByTime(1000)
        const secondInstance = getLatestMockWebSocket()
        secondInstance?.simulateClose()
        expect(onReconnecting).toHaveBeenCalledWith(2)

        vi.advanceTimersByTime(2000)
        const thirdInstance = getLatestMockWebSocket()
        thirdInstance?.simulateClose()
        expect(onReconnecting).toHaveBeenCalledWith(3)

        vi.advanceTimersByTime(3000)
        const fourthInstance = getLatestMockWebSocket()
        fourthInstance?.simulateClose()

        // Should only have called onReconnecting 3 times (max attempts)
        expect(onReconnecting).toHaveBeenCalledTimes(3)
        expect(consoleErrorSpy).toHaveBeenCalledWith('Max reconnection attempts reached')

        consoleErrorSpy.mockRestore()
      })

      it('should call onReconnected after successful reconnection', () => {
        const onReconnected = vi.fn()
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
          reconnectDelay: 1000,
          onReconnected,
        })

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        getLatestMockWebSocket()?.simulateClose()

        // Advance timer to trigger reconnect
        vi.advanceTimersByTime(1000)

        // Simulate successful reconnection
        getLatestMockWebSocket()?.simulateOpen()

        expect(onReconnected).toHaveBeenCalledTimes(1)
      })

      it('should reset reconnect counter after successful connection', () => {
        const onReconnecting = vi.fn()
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
          reconnectDelay: 1000,
          maxReconnectAttempts: 3,
          onReconnecting,
        })

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        getLatestMockWebSocket()?.simulateClose()

        // First reconnect
        vi.advanceTimersByTime(1000)
        getLatestMockWebSocket()?.simulateOpen() // Successful reconnect

        // Another disconnect
        getLatestMockWebSocket()?.simulateClose()

        // Should start from 1 again
        expect(onReconnecting).toHaveBeenCalledWith(1)
        expect(onReconnecting).toHaveBeenLastCalledWith(1)
      })
    })

    describe('event types', () => {
      it('should handle step.completed events with step info', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })
        const handler: EventHandler = vi.fn()

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        client.on(handler)

        const mockEvent: ExecutionEvent = {
          type: 'step.completed',
          execution_id: 'exec-123',
          workflow_id: 'wf-456',
          tenant_id: 'tenant-789',
          step: {
            step_id: 'step-1',
            node_id: 'node-1',
            node_type: 'http',
            status: 'completed',
            output_data: { response: { status: 200 } },
            duration_ms: 150,
            started_at: '2025-01-01T00:00:00Z',
            completed_at: '2025-01-01T00:00:00.150Z',
          },
          timestamp: new Date().toISOString(),
        }

        getLatestMockWebSocket()?.simulateMessage(mockEvent)

        expect(handler).toHaveBeenCalledWith(mockEvent)
        expect(handler.mock.calls[0][0].step?.output_data).toEqual({ response: { status: 200 } })
      })

      it('should handle execution.progress events', () => {
        const client = new WebSocketClient({
          url: 'ws://localhost:8080/ws',
        })
        const handler: EventHandler = vi.fn()

        client.connect()
        getLatestMockWebSocket()?.simulateOpen()
        client.on(handler)

        const mockEvent: ExecutionEvent = {
          type: 'execution.progress',
          execution_id: 'exec-123',
          workflow_id: 'wf-456',
          tenant_id: 'tenant-789',
          progress: {
            total_steps: 5,
            completed_steps: 2,
            percentage: 40,
          },
          timestamp: new Date().toISOString(),
        }

        getLatestMockWebSocket()?.simulateMessage(mockEvent)

        expect(handler).toHaveBeenCalledWith(mockEvent)
        expect(handler.mock.calls[0][0].progress?.percentage).toBe(40)
      })
    })
  })

  describe('createExecutionWebSocketURL', () => {
    it('should create correct URL with default base', () => {
      const url = createExecutionWebSocketURL('exec-123')
      expect(url).toBe('ws://localhost:8080/api/v1/ws/executions/exec-123')
    })

    it('should create correct URL with http base URL', () => {
      const url = createExecutionWebSocketURL('exec-123', 'http://api.example.com')
      expect(url).toBe('ws://api.example.com/api/v1/ws/executions/exec-123')
    })

    it('should create wss URL with https base URL', () => {
      const url = createExecutionWebSocketURL('exec-123', 'https://api.example.com')
      expect(url).toBe('wss://api.example.com/api/v1/ws/executions/exec-123')
    })

    it('should handle base URL with port', () => {
      const url = createExecutionWebSocketURL('exec-123', 'http://localhost:3000')
      expect(url).toBe('ws://localhost:3000/api/v1/ws/executions/exec-123')
    })
  })

  describe('createWorkflowWebSocketURL', () => {
    it('should create correct URL with default base', () => {
      const url = createWorkflowWebSocketURL('wf-456')
      expect(url).toBe('ws://localhost:8080/api/v1/ws/workflows/wf-456')
    })

    it('should create correct URL with custom base URL', () => {
      const url = createWorkflowWebSocketURL('wf-456', 'http://api.example.com')
      expect(url).toBe('ws://api.example.com/api/v1/ws/workflows/wf-456')
    })

    it('should create wss URL with https base URL', () => {
      const url = createWorkflowWebSocketURL('wf-456', 'https://api.example.com')
      expect(url).toBe('wss://api.example.com/api/v1/ws/workflows/wf-456')
    })
  })

  describe('createTenantWebSocketURL', () => {
    it('should create correct URL with default base', () => {
      const url = createTenantWebSocketURL()
      expect(url).toBe('ws://localhost:8080/api/v1/ws?subscribe_tenant=true')
    })

    it('should create correct URL with custom base URL', () => {
      const url = createTenantWebSocketURL('http://api.example.com')
      expect(url).toBe('ws://api.example.com/api/v1/ws?subscribe_tenant=true')
    })

    it('should create wss URL with https base URL', () => {
      const url = createTenantWebSocketURL('https://api.example.com')
      expect(url).toBe('wss://api.example.com/api/v1/ws?subscribe_tenant=true')
    })
  })
})
