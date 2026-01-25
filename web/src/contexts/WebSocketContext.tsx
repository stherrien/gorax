/**
 * WebSocket Context Provider
 *
 * Provides shared WebSocket connection management across the application.
 * Features:
 * - Tenant-wide WebSocket connection for global updates
 * - Authentication token management
 * - Connection state tracking
 * - Automatic reconnection handling
 * - Event subscription system
 *
 * @example
 * ```tsx
 * // In App.tsx
 * <WebSocketProvider>
 *   <App />
 * </WebSocketProvider>
 *
 * // In a component
 * const { connected, subscribe } = useWebSocketContext()
 * useEffect(() => {
 *   return subscribe((event) => {
 *     console.log('Received event:', event)
 *   })
 * }, [subscribe])
 * ```
 */

import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  useCallback,
  type ReactNode,
} from 'react'
import { WebSocketClient } from '../lib/websocket'
import type { ExecutionEvent, EventHandler } from '../types/websocket'
import {
  createTenantWebSocketUrl,
  getWebSocketConfig,
} from '../config/websocket'

// ============================================================================
// Types
// ============================================================================

interface WebSocketContextValue {
  /** Whether the WebSocket is connected */
  connected: boolean
  /** Whether the WebSocket is attempting to reconnect */
  reconnecting: boolean
  /** Current reconnection attempt number */
  reconnectAttempt: number
  /** Maximum reconnection attempts */
  maxReconnectAttempts: number
  /** Subscribe to WebSocket events */
  subscribe: (handler: EventHandler<ExecutionEvent>) => () => void
  /** Manually trigger reconnection */
  reconnect: () => void
  /** Disconnect the WebSocket */
  disconnect: () => void
  /** Last error message if any */
  lastError: string | null
}

interface WebSocketProviderProps {
  children: ReactNode
  /** Base URL for WebSocket connection (optional, uses env var by default) */
  baseURL?: string
  /** Whether to auto-connect on mount */
  autoConnect?: boolean
}

// ============================================================================
// Context
// ============================================================================

const WebSocketContext = createContext<WebSocketContextValue | null>(null)

// ============================================================================
// Provider
// ============================================================================

/**
 * WebSocket Provider component
 *
 * Manages a shared WebSocket connection for tenant-wide updates.
 * Should be placed near the root of your application.
 */
export function WebSocketProvider({
  children,
  baseURL,
  autoConnect = true,
}: WebSocketProviderProps) {
  // Connection state
  const [connected, setConnected] = useState(false)
  const [reconnecting, setReconnecting] = useState(false)
  const [reconnectAttempt, setReconnectAttempt] = useState(0)
  const [lastError, setLastError] = useState<string | null>(null)

  // WebSocket client ref
  const clientRef = useRef<WebSocketClient | null>(null)

  // Get configuration
  const config = getWebSocketConfig()
  const maxReconnectAttempts = config.maxReconnectAttempts

  /**
   * Subscribe to WebSocket events
   * Returns an unsubscribe function
   */
  const subscribe = useCallback((handler: EventHandler<ExecutionEvent>) => {
    if (!clientRef.current) {
      // Return no-op if client not initialized
      return () => {}
    }
    return clientRef.current.on(handler)
  }, [])

  /**
   * Manually trigger reconnection
   */
  const reconnect = useCallback(() => {
    if (clientRef.current) {
      setLastError(null)
      clientRef.current.disconnect()
      clientRef.current.connect()
    }
  }, [])

  /**
   * Disconnect the WebSocket
   */
  const disconnect = useCallback(() => {
    if (clientRef.current) {
      clientRef.current.disconnect()
    }
  }, [])

  /**
   * Initialize WebSocket connection
   */
  useEffect(() => {
    if (!autoConnect) {
      return
    }

    const wsURL = createTenantWebSocketUrl(baseURL)

    const client = new WebSocketClient({
      url: wsURL,
      reconnectDelay: config.reconnectDelay,
      maxReconnectAttempts: config.maxReconnectAttempts,
      onOpen: () => {
        setConnected(true)
        setReconnecting(false)
        setReconnectAttempt(0)
        setLastError(null)
      },
      onClose: () => {
        setConnected(false)
      },
      onError: () => {
        setLastError('WebSocket connection error')
      },
      onReconnecting: (attempt) => {
        setReconnecting(true)
        setReconnectAttempt(attempt)
      },
      onReconnected: () => {
        setReconnecting(false)
        setLastError(null)
      },
    })

    client.connect()
    clientRef.current = client

    return () => {
      client.disconnect()
      clientRef.current = null
    }
  }, [autoConnect, baseURL, config.reconnectDelay, config.maxReconnectAttempts])

  const value: WebSocketContextValue = {
    connected,
    reconnecting,
    reconnectAttempt,
    maxReconnectAttempts,
    subscribe,
    reconnect,
    disconnect,
    lastError,
  }

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  )
}

// ============================================================================
// Hook
// ============================================================================

/**
 * Hook to access WebSocket context
 *
 * @throws Error if used outside of WebSocketProvider
 *
 * @example
 * ```tsx
 * const { connected, reconnecting, subscribe, reconnect } = useWebSocketContext()
 * ```
 */
export function useWebSocketContext(): WebSocketContextValue {
  const context = useContext(WebSocketContext)
  if (!context) {
    throw new Error('useWebSocketContext must be used within a WebSocketProvider')
  }
  return context
}

/**
 * Hook to optionally access WebSocket context
 * Returns null if not within provider (useful for optional features)
 *
 * @example
 * ```tsx
 * const wsContext = useOptionalWebSocketContext()
 * if (wsContext?.connected) {
 *   // WebSocket is available and connected
 * }
 * ```
 */
export function useOptionalWebSocketContext(): WebSocketContextValue | null {
  return useContext(WebSocketContext)
}

// ============================================================================
// Utility Hooks
// ============================================================================

/**
 * Hook to subscribe to WebSocket events with automatic cleanup
 *
 * @example
 * ```tsx
 * useWebSocketSubscription((event) => {
 *   if (event.type === 'execution.completed') {
 *     // Handle completion
 *   }
 * })
 * ```
 */
export function useWebSocketSubscription(
  handler: EventHandler<ExecutionEvent>,
  deps: React.DependencyList = []
): void {
  const { subscribe } = useWebSocketContext()

  useEffect(() => {
    const unsubscribe = subscribe(handler)
    return unsubscribe
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [subscribe, ...deps])
}

/**
 * Hook to get just the connection status
 *
 * @example
 * ```tsx
 * const { connected, reconnecting } = useWebSocketStatus()
 * ```
 */
export function useWebSocketStatus(): Pick<
  WebSocketContextValue,
  'connected' | 'reconnecting' | 'reconnectAttempt' | 'lastError'
> {
  const { connected, reconnecting, reconnectAttempt, lastError } = useWebSocketContext()
  return { connected, reconnecting, reconnectAttempt, lastError }
}
