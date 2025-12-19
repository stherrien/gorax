/**
 * WebSocket client utilities for real-time execution updates
 */

export type EventType =
  | 'execution.started'
  | 'execution.completed'
  | 'execution.failed'
  | 'step.started'
  | 'step.completed'
  | 'step.failed'
  | 'execution.progress'

export interface ProgressInfo {
  total_steps: number
  completed_steps: number
  percentage: number
}

export interface StepInfo {
  step_id: string
  node_id: string
  node_type: string
  status: string
  output_data?: any
  error?: string
  duration_ms?: number
  started_at?: string
  completed_at?: string
}

export interface ExecutionEvent {
  type: EventType
  execution_id: string
  workflow_id: string
  tenant_id: string
  status?: string
  progress?: ProgressInfo
  step?: StepInfo
  error?: string
  output?: any
  metadata?: Record<string, any>
  timestamp: string
}

export type EventHandler = (event: ExecutionEvent) => void

export interface WebSocketConfig {
  url: string
  reconnectDelay?: number
  maxReconnectAttempts?: number
  onOpen?: () => void
  onClose?: () => void
  onError?: (error: Event) => void
  onReconnecting?: (attempt: number) => void
  onReconnected?: () => void
}

/**
 * WebSocket client with automatic reconnection
 */
export class WebSocketClient {
  private ws: WebSocket | null = null
  private config: Required<WebSocketConfig>
  private handlers: Set<EventHandler> = new Set()
  private reconnectAttempt = 0
  private reconnectTimer: number | null = null
  private intentionallyClosed = false
  private isReconnecting = false

  constructor(config: WebSocketConfig) {
    this.config = {
      reconnectDelay: 3000,
      maxReconnectAttempts: 10,
      onOpen: () => {},
      onClose: () => {},
      onError: () => {},
      onReconnecting: () => {},
      onReconnected: () => {},
      ...config,
    }
  }

  /**
   * Connect to the WebSocket server
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    this.intentionallyClosed = false

    try {
      this.ws = new WebSocket(this.config.url)

      this.ws.onopen = () => {
        this.reconnectAttempt = 0
        if (this.isReconnecting) {
          this.isReconnecting = false
          this.config.onReconnected()
        }
        this.config.onOpen()
      }

      this.ws.onclose = () => {
        this.config.onClose()
        if (!this.intentionallyClosed) {
          this.scheduleReconnect()
        }
      }

      this.ws.onerror = (error) => {
        this.config.onError(error)
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as ExecutionEvent
          this.notifyHandlers(data)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }
    } catch (error) {
      console.error('Failed to create WebSocket:', error)
      this.scheduleReconnect()
    }
  }

  /**
   * Disconnect from the WebSocket server
   */
  disconnect(): void {
    this.intentionallyClosed = true
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  /**
   * Subscribe to events
   */
  on(handler: EventHandler): () => void {
    this.handlers.add(handler)
    return () => this.handlers.delete(handler)
  }

  /**
   * Get connection status
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  /**
   * Schedule a reconnection attempt
   */
  private scheduleReconnect(): void {
    if (this.intentionallyClosed) {
      return
    }

    if (this.reconnectAttempt >= this.config.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached')
      return
    }

    if (this.reconnectTimer !== null) {
      return
    }

    this.isReconnecting = true
    this.reconnectAttempt++
    this.config.onReconnecting(this.reconnectAttempt)

    const delay = this.config.reconnectDelay * Math.min(this.reconnectAttempt, 3)
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, delay)
  }

  /**
   * Notify all event handlers
   */
  private notifyHandlers(event: ExecutionEvent): void {
    this.handlers.forEach((handler) => {
      try {
        handler(event)
      } catch (error) {
        console.error('Error in event handler:', error)
      }
    })
  }
}

/**
 * Create WebSocket URL for execution updates
 */
export function createExecutionWebSocketURL(
  executionId: string,
  baseURL?: string
): string {
  const base = baseURL || import.meta.env.VITE_API_URL || 'http://localhost:8080'
  const wsProtocol = base.startsWith('https') ? 'wss' : 'ws'
  const url = base.replace(/^https?:\/\//, '')
  return `${wsProtocol}://${url}/api/v1/ws/executions/${executionId}`
}

/**
 * Create WebSocket URL for workflow updates
 */
export function createWorkflowWebSocketURL(
  workflowId: string,
  baseURL?: string
): string {
  const base = baseURL || import.meta.env.VITE_API_URL || 'http://localhost:8080'
  const wsProtocol = base.startsWith('https') ? 'wss' : 'ws'
  const url = base.replace(/^https?:\/\//, '')
  return `${wsProtocol}://${url}/api/v1/ws/workflows/${workflowId}`
}

/**
 * Create WebSocket URL for tenant-wide updates
 */
export function createTenantWebSocketURL(baseURL?: string): string {
  const base = baseURL || import.meta.env.VITE_API_URL || 'http://localhost:8080'
  const wsProtocol = base.startsWith('https') ? 'wss' : 'ws'
  const url = base.replace(/^https?:\/\//, '')
  return `${wsProtocol}://${url}/api/v1/ws?subscribe_tenant=true`
}
