import { useState, useEffect, useCallback, useRef } from 'react'
import {
  WebSocketClient,
  ExecutionEvent,
  EventType,
  ProgressInfo,
  StepInfo,
  createExecutionWebSocketURL,
} from '../lib/websocket'

export interface ExecutionUpdate {
  type: EventType
  executionId: string
  workflowId: string
  status?: string
  progress?: ProgressInfo
  step?: StepInfo
  error?: string
  output?: any
  timestamp: string
}

export interface UseExecutionUpdatesOptions {
  enabled?: boolean
  onStatusChange?: (status: string) => void
  onProgress?: (progress: ProgressInfo) => void
  onStepComplete?: (step: StepInfo) => void
  onComplete?: (output: any) => void
  onError?: (error: string) => void
  baseURL?: string
}

export interface UseExecutionUpdatesReturn {
  // Connection state
  connected: boolean
  reconnecting: boolean
  reconnectAttempt: number

  // Latest updates
  latestUpdate: ExecutionUpdate | null
  currentStatus: string | null
  currentProgress: ProgressInfo | null
  completedSteps: StepInfo[]

  // All events
  events: ExecutionEvent[]

  // Actions
  reconnect: () => void
  clearEvents: () => void
}

/**
 * Hook to subscribe to real-time execution updates via WebSocket
 *
 * @example
 * ```tsx
 * const { connected, currentStatus, currentProgress } = useExecutionUpdates(
 *   executionId,
 *   {
 *     onStatusChange: (status) => {
 *       // Handle status change
 *     },
 *     onComplete: (output) => {
 *       // Handle completion
 *     },
 *   }
 * )
 * ```
 */
export function useExecutionUpdates(
  executionId: string | null,
  options: UseExecutionUpdatesOptions = {}
): UseExecutionUpdatesReturn {
  const {
    enabled = true,
    onStatusChange,
    onProgress,
    onStepComplete,
    onComplete,
    onError,
    baseURL,
  } = options

  const [connected, setConnected] = useState(false)
  const [reconnecting, setReconnecting] = useState(false)
  const [reconnectAttempt, setReconnectAttempt] = useState(0)
  const [latestUpdate, setLatestUpdate] = useState<ExecutionUpdate | null>(null)
  const [currentStatus, setCurrentStatus] = useState<string | null>(null)
  const [currentProgress, setCurrentProgress] = useState<ProgressInfo | null>(null)
  const [completedSteps, setCompletedSteps] = useState<StepInfo[]>([])
  const [events, setEvents] = useState<ExecutionEvent[]>([])

  const clientRef = useRef<WebSocketClient | null>(null)
  const callbacksRef = useRef({ onStatusChange, onProgress, onStepComplete, onComplete, onError })

  // Keep callbacks ref up to date
  useEffect(() => {
    callbacksRef.current = { onStatusChange, onProgress, onStepComplete, onComplete, onError }
  }, [onStatusChange, onProgress, onStepComplete, onComplete, onError])

  // Handle incoming events
  const handleEvent = useCallback((event: ExecutionEvent) => {
    // Add to events list
    setEvents((prev) => [...prev, event])

    // Create update object
    const update: ExecutionUpdate = {
      type: event.type,
      executionId: event.execution_id,
      workflowId: event.workflow_id,
      status: event.status,
      progress: event.progress,
      step: event.step,
      error: event.error,
      output: event.output,
      timestamp: event.timestamp,
    }
    setLatestUpdate(update)

    // Update current status
    if (event.status) {
      setCurrentStatus(event.status)
      callbacksRef.current.onStatusChange?.(event.status)
    }

    // Handle specific event types
    switch (event.type) {
      case 'execution.started':
        if (event.progress) {
          setCurrentProgress(event.progress)
          callbacksRef.current.onProgress?.(event.progress)
        }
        break

      case 'execution.progress':
        if (event.progress) {
          setCurrentProgress(event.progress)
          callbacksRef.current.onProgress?.(event.progress)
        }
        break

      case 'step.completed':
        if (event.step) {
          setCompletedSteps((prev) => [...prev, event.step!])
          callbacksRef.current.onStepComplete?.(event.step)
        }
        break

      case 'execution.completed':
        if (event.output) {
          callbacksRef.current.onComplete?.(event.output)
        }
        break

      case 'execution.failed':
      case 'step.failed':
        if (event.error) {
          callbacksRef.current.onError?.(event.error)
        }
        break
    }
  }, [])

  // Reconnect function
  const reconnect = useCallback(() => {
    if (clientRef.current) {
      clientRef.current.disconnect()
      clientRef.current.connect()
    }
  }, [])

  // Clear events function
  const clearEvents = useCallback(() => {
    setEvents([])
    setCompletedSteps([])
  }, [])

  // Setup WebSocket connection
  useEffect(() => {
    if (!executionId || !enabled) {
      return
    }

    const wsURL = createExecutionWebSocketURL(executionId, baseURL)

    const client = new WebSocketClient({
      url: wsURL,
      reconnectDelay: 3000,
      maxReconnectAttempts: 10,
      onOpen: () => {
        setConnected(true)
        setReconnecting(false)
        setReconnectAttempt(0)
      },
      onClose: () => {
        setConnected(false)
      },
      onError: () => {
        // WebSocket error - could be logged to error tracking service
        callbacksRef.current.onError?.('WebSocket connection error')
      },
      onReconnecting: (attempt) => {
        setReconnecting(true)
        setReconnectAttempt(attempt)
      },
      onReconnected: () => {
        setReconnecting(false)
      },
    })

    const unsubscribe = client.on(handleEvent)
    client.connect()

    clientRef.current = client

    return () => {
      unsubscribe()
      client.disconnect()
      clientRef.current = null
    }
  }, [executionId, enabled, baseURL, handleEvent])

  return {
    connected,
    reconnecting,
    reconnectAttempt,
    latestUpdate,
    currentStatus,
    currentProgress,
    completedSteps,
    events,
    reconnect,
    clearEvents,
  }
}
