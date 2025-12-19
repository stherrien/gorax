import { useState, useEffect, useCallback, useRef } from 'react'
import {
  WebSocketClient,
  ExecutionEvent,
  createExecutionWebSocketURL,
} from '../lib/websocket'
import { useExecutionTraceStore } from '../stores/executionTraceStore'
import type { TimelineEvent } from '../stores/executionTraceStore'

export interface UseExecutionTraceOptions {
  enabled?: boolean
  baseURL?: string
}

export interface UseExecutionTraceReturn {
  connected: boolean
  reconnecting: boolean
  reconnectAttempt: number
  reconnect: () => void
}

/**
 * Hook to subscribe to execution updates and update the trace store
 *
 * This hook:
 * - Connects to WebSocket for execution events
 * - Updates execution trace store with real-time data
 * - Manages node status, step logs, and timeline events
 * - Handles connection lifecycle
 *
 * @example
 * ```tsx
 * const { connected, reconnecting } = useExecutionTrace('exec-123')
 * ```
 */
export function useExecutionTrace(
  executionId: string | null,
  options: UseExecutionTraceOptions = {}
): UseExecutionTraceReturn {
  const { enabled = true, baseURL } = options

  const [connected, setConnected] = useState(false)
  const [reconnecting, setReconnecting] = useState(false)
  const [reconnectAttempt, setReconnectAttempt] = useState(0)

  const clientRef = useRef<WebSocketClient | null>(null)
  const storeRef = useRef(useExecutionTraceStore.getState())

  /**
   * Handle incoming execution events
   */
  const handleEvent = useCallback(
    (event: ExecutionEvent) => {
      switch (event.type) {
        case 'execution.started':
          handleExecutionStarted(event)
          break

        case 'step.started':
          handleStepStarted(event)
          break

        case 'step.completed':
          handleStepCompleted(event)
          break

        case 'step.failed':
          handleStepFailed(event)
          break

        case 'execution.progress':
          handleExecutionProgress(event)
          break

        case 'execution.completed':
          handleExecutionCompleted(event)
          break

        case 'execution.failed':
          handleExecutionFailed(event)
          break
      }
    },
    []
  )

  /**
   * Handle execution started event
   */
  const handleExecutionStarted = (event: ExecutionEvent) => {
    const timelineEvent: TimelineEvent = {
      timestamp: event.timestamp,
      nodeId: 'execution',
      type: 'started',
      message: 'Execution started',
      metadata: event.progress
        ? {
            total_steps: event.progress.total_steps,
          }
        : undefined,
    }
    storeRef.current.addTimelineEvent(timelineEvent)
  }

  /**
   * Handle step started event
   */
  const handleStepStarted = (event: ExecutionEvent) => {
    if (!event.step) return

    storeRef.current.setNodeStatus(event.step.node_id, 'running')

    const timelineEvent: TimelineEvent = {
      timestamp: event.timestamp,
      nodeId: event.step.node_id,
      type: 'started',
      message: `Started ${event.step.node_type}`,
    }
    storeRef.current.addTimelineEvent(timelineEvent)
  }

  /**
   * Handle step completed event
   */
  const handleStepCompleted = (event: ExecutionEvent) => {
    if (!event.step) return

    storeRef.current.setNodeStatus(event.step.node_id, 'completed')
    storeRef.current.addStepLog(event.step.node_id, event.step)

    const timelineEvent: TimelineEvent = {
      timestamp: event.timestamp,
      nodeId: event.step.node_id,
      type: 'completed',
      message: `Completed ${event.step.node_type}`,
      metadata: event.step.duration_ms
        ? {
            duration_ms: event.step.duration_ms,
          }
        : undefined,
    }
    storeRef.current.addTimelineEvent(timelineEvent)
  }

  /**
   * Handle step failed event
   */
  const handleStepFailed = (event: ExecutionEvent) => {
    if (!event.step) return

    storeRef.current.setNodeStatus(event.step.node_id, 'failed')
    storeRef.current.addStepLog(event.step.node_id, event.step)

    const timelineEvent: TimelineEvent = {
      timestamp: event.timestamp,
      nodeId: event.step.node_id,
      type: 'failed',
      message: `Failed: ${event.step.error || 'Unknown error'}`,
      metadata: event.step.error
        ? {
            error: event.step.error,
          }
        : undefined,
    }
    storeRef.current.addTimelineEvent(timelineEvent)
  }

  /**
   * Handle execution progress event
   */
  const handleExecutionProgress = (event: ExecutionEvent) => {
    if (!event.progress) return

    const timelineEvent: TimelineEvent = {
      timestamp: event.timestamp,
      nodeId: 'execution',
      type: 'progress',
      message: `Progress: ${event.progress.completed_steps}/${event.progress.total_steps} steps (${event.progress.percentage}%)`,
      metadata: {
        completed_steps: event.progress.completed_steps,
        total_steps: event.progress.total_steps,
        percentage: event.progress.percentage,
      },
    }
    storeRef.current.addTimelineEvent(timelineEvent)
  }

  /**
   * Handle execution completed event
   */
  const handleExecutionCompleted = (event: ExecutionEvent) => {
    const timelineEvent: TimelineEvent = {
      timestamp: event.timestamp,
      nodeId: 'execution',
      type: 'completed',
      message: 'Execution completed successfully',
    }
    storeRef.current.addTimelineEvent(timelineEvent)
  }

  /**
   * Handle execution failed event
   */
  const handleExecutionFailed = (event: ExecutionEvent) => {
    const timelineEvent: TimelineEvent = {
      timestamp: event.timestamp,
      nodeId: 'execution',
      type: 'failed',
      message: `Execution failed: ${event.error || 'Unknown error'}`,
      metadata: event.error
        ? {
            error: event.error,
          }
        : undefined,
    }
    storeRef.current.addTimelineEvent(timelineEvent)
  }

  /**
   * Reconnect function
   */
  const reconnect = useCallback(() => {
    if (clientRef.current) {
      clientRef.current.disconnect()
      clientRef.current.connect()
    }
  }, [])

  /**
   * Setup WebSocket connection
   */
  useEffect(() => {
    if (!executionId || !enabled) {
      return
    }

    // Set current execution ID in store
    storeRef.current.setCurrentExecutionId(executionId)

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
      onError: (error) => {
        console.error('WebSocket error:', error)
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
      storeRef.current.reset()
    }
  }, [executionId, enabled, baseURL, handleEvent])

  return {
    connected,
    reconnecting,
    reconnectAttempt,
    reconnect,
  }
}
