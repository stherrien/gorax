/**
 * WebSocket types for real-time workflow execution updates
 *
 * This module contains all TypeScript interfaces and types
 * for WebSocket communication between the frontend and backend.
 */

// ============================================================================
// Connection Types
// ============================================================================

/**
 * WebSocket connection state
 */
export type ConnectionState =
  | 'connecting'
  | 'connected'
  | 'disconnected'
  | 'reconnecting'
  | 'error'

/**
 * WebSocket connection info
 */
export interface ConnectionInfo {
  state: ConnectionState
  reconnectAttempt: number
  maxReconnectAttempts: number
  lastError?: string
}

// ============================================================================
// Event Types
// ============================================================================

/**
 * Types of execution events sent via WebSocket
 */
export type ExecutionEventType =
  | 'execution.started'
  | 'execution.completed'
  | 'execution.failed'
  | 'execution.progress'
  | 'step.started'
  | 'step.completed'
  | 'step.failed'

/**
 * Types of collaboration events sent via WebSocket
 */
export type CollaborationEventType =
  | 'user.joined'
  | 'user.left'
  | 'cursor.moved'
  | 'node.selected'
  | 'node.editing'
  | 'workflow.updated'

/**
 * All WebSocket event types
 */
export type WebSocketEventType = ExecutionEventType | CollaborationEventType

// ============================================================================
// Execution Event Payloads
// ============================================================================

/**
 * Progress information for execution events
 */
export interface ProgressInfo {
  total_steps: number
  completed_steps: number
  percentage: number
}

/**
 * Step information for step-related events
 */
export interface StepInfo {
  step_id: string
  node_id: string
  node_type: string
  status: string
  output_data?: unknown
  error?: string
  duration_ms?: number
  started_at?: string
  completed_at?: string
}

/**
 * Execution event payload
 */
export interface ExecutionEvent {
  type: ExecutionEventType
  execution_id: string
  workflow_id: string
  tenant_id: string
  status?: string
  progress?: ProgressInfo
  step?: StepInfo
  error?: string
  output?: unknown
  metadata?: Record<string, unknown>
  timestamp: string
}

// ============================================================================
// Collaboration Event Payloads
// ============================================================================

/**
 * User presence information
 */
export interface UserPresence {
  user_id: string
  user_name: string
  color: string
  cursor_x?: number
  cursor_y?: number
  selected_node_id?: string
  editing_node_id?: string
  last_seen: string
}

/**
 * Collaboration event payload
 */
export interface CollaborationEvent {
  type: CollaborationEventType
  workflow_id: string
  user: UserPresence
  timestamp: string
}

// ============================================================================
// WebSocket Message Types
// ============================================================================

/**
 * Incoming WebSocket message (from server)
 */
export type IncomingMessage = ExecutionEvent | CollaborationEvent

/**
 * Outgoing WebSocket message (to server)
 */
export interface OutgoingMessage {
  type: string
  payload: Record<string, unknown>
}

// ============================================================================
// Configuration Types
// ============================================================================

/**
 * WebSocket configuration options
 */
export interface WebSocketConfig {
  /** WebSocket URL to connect to */
  url: string
  /** Delay before reconnection attempt (ms) */
  reconnectDelay?: number
  /** Maximum number of reconnection attempts */
  maxReconnectAttempts?: number
  /** Connection timeout (ms) */
  connectionTimeout?: number
  /** Heartbeat interval (ms) */
  heartbeatInterval?: number
}

/**
 * WebSocket client configuration with callbacks
 */
export interface WebSocketClientConfig extends WebSocketConfig {
  /** Called when connection opens */
  onOpen?: () => void
  /** Called when connection closes */
  onClose?: () => void
  /** Called on connection error */
  onError?: (error: Event) => void
  /** Called when attempting to reconnect */
  onReconnecting?: (attempt: number) => void
  /** Called after successful reconnection */
  onReconnected?: () => void
}

// ============================================================================
// Hook Return Types
// ============================================================================

/**
 * Return type for useWebSocket hook
 */
export interface UseWebSocketReturn<T = IncomingMessage> {
  /** Whether the WebSocket is connected */
  connected: boolean
  /** Whether the WebSocket is reconnecting */
  reconnecting: boolean
  /** Current reconnection attempt number */
  reconnectAttempt: number
  /** Latest message received */
  latestMessage: T | null
  /** Send a message through the WebSocket */
  send: (message: OutgoingMessage) => void
  /** Manually reconnect */
  reconnect: () => void
  /** Disconnect the WebSocket */
  disconnect: () => void
}

/**
 * Return type for useExecutionUpdates hook
 */
export interface UseExecutionUpdatesReturn {
  /** Whether the WebSocket is connected */
  connected: boolean
  /** Whether the WebSocket is reconnecting */
  reconnecting: boolean
  /** Current reconnection attempt number */
  reconnectAttempt: number
  /** Latest execution update */
  latestUpdate: ExecutionUpdate | null
  /** Current execution status */
  currentStatus: string | null
  /** Current progress info */
  currentProgress: ProgressInfo | null
  /** List of completed steps */
  completedSteps: StepInfo[]
  /** All events received */
  events: ExecutionEvent[]
  /** Manually reconnect */
  reconnect: () => void
  /** Clear events history */
  clearEvents: () => void
}

/**
 * Execution update (transformed from ExecutionEvent)
 */
export interface ExecutionUpdate {
  type: ExecutionEventType
  executionId: string
  workflowId: string
  status?: string
  progress?: ProgressInfo
  step?: StepInfo
  error?: string
  output?: unknown
  timestamp: string
}

// ============================================================================
// Event Handler Types
// ============================================================================

/**
 * Handler function for execution events
 */
export type ExecutionEventHandler = (event: ExecutionEvent) => void

/**
 * Handler function for collaboration events
 */
export type CollaborationEventHandler = (event: CollaborationEvent) => void

/**
 * Generic event handler
 */
export type EventHandler<T = IncomingMessage> = (event: T) => void

// ============================================================================
// Subscription Types
// ============================================================================

/**
 * Subscription options for execution updates
 */
export interface ExecutionSubscriptionOptions {
  /** Whether the subscription is enabled */
  enabled?: boolean
  /** Called when execution status changes */
  onStatusChange?: (status: string) => void
  /** Called when progress updates */
  onProgress?: (progress: ProgressInfo) => void
  /** Called when a step completes */
  onStepComplete?: (step: StepInfo) => void
  /** Called when execution completes */
  onComplete?: (output: unknown) => void
  /** Called on error */
  onError?: (error: string) => void
  /** Base URL for WebSocket connection */
  baseURL?: string
}

/**
 * Subscription options for collaboration updates
 */
export interface CollaborationSubscriptionOptions {
  /** Whether the subscription is enabled */
  enabled?: boolean
  /** Called when a user joins */
  onUserJoined?: (user: UserPresence) => void
  /** Called when a user leaves */
  onUserLeft?: (user: UserPresence) => void
  /** Called when a cursor moves */
  onCursorMoved?: (user: UserPresence) => void
  /** Base URL for WebSocket connection */
  baseURL?: string
}
