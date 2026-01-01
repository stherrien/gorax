import { useState, useEffect, useCallback, useRef } from 'react'
import type {
  WebSocketMessage,
  UserPresence,
  EditLock,
  EditSession,
  EditOperation,
} from '../types/collaboration'

export interface UseCollaborationOptions {
  enabled?: boolean
  onUserJoined?: (user: UserPresence) => void
  onUserLeft?: (userId: string) => void
  onPresenceUpdate?: (userId: string, presence: UserPresence) => void
  onLockAcquired?: (lock: EditLock) => void
  onLockReleased?: (elementId: string) => void
  onLockFailed?: (elementId: string, reason: string, currentLock?: EditLock) => void
  onChangeApplied?: (operation: EditOperation) => void
  onError?: (error: string) => void
  baseURL?: string
}

export interface UseCollaborationReturn {
  // Connection state
  connected: boolean
  reconnecting: boolean

  // Session state
  session: EditSession | null
  users: UserPresence[]
  locks: EditLock[]

  // Actions
  join: () => void
  leave: () => void
  updateCursor: (x: number, y: number) => void
  updateSelection: (type: 'node' | 'edge', elementIds: string[]) => void
  acquireLock: (elementId: string, elementType: 'node' | 'edge') => void
  releaseLock: (elementId: string) => void
  broadcastChange: (operation: EditOperation) => void

  // Helpers
  isElementLocked: (elementId: string) => boolean
  getElementLock: (elementId: string) => EditLock | undefined
  isLockedByMe: (elementId: string) => boolean
  isLockedByOther: (elementId: string) => boolean
}

/**
 * Hook for real-time workflow collaboration
 */
export function useCollaboration(
  workflowId: string | null,
  userId: string,
  userName: string,
  options: UseCollaborationOptions = {}
): UseCollaborationReturn {
  const {
    enabled = true,
    onUserJoined,
    onUserLeft,
    onPresenceUpdate,
    onLockAcquired,
    onLockReleased,
    onLockFailed,
    onChangeApplied,
    onError,
    baseURL,
  } = options

  const [connected, setConnected] = useState(false)
  const [reconnecting, setReconnecting] = useState(false)
  const [session, setSession] = useState<EditSession | null>(null)
  const [users, setUsers] = useState<UserPresence[]>([])
  const [locks, setLocks] = useState<EditLock[]>([])

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<number | null>(null)
  const reconnectAttemptsRef = useRef(0)

  // Keep callbacks up to date
  const callbacksRef = useRef({
    onUserJoined,
    onUserLeft,
    onPresenceUpdate,
    onLockAcquired,
    onLockReleased,
    onLockFailed,
    onChangeApplied,
    onError,
  })

  useEffect(() => {
    callbacksRef.current = {
      onUserJoined,
      onUserLeft,
      onPresenceUpdate,
      onLockAcquired,
      onLockReleased,
      onLockFailed,
      onChangeApplied,
      onError,
    }
  }, [
    onUserJoined,
    onUserLeft,
    onPresenceUpdate,
    onLockAcquired,
    onLockReleased,
    onLockFailed,
    onChangeApplied,
    onError,
  ])

  // Send message helper
  const sendMessage = useCallback((message: WebSocketMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message))
    }
  }, [])

  // Join session
  const join = useCallback(() => {
    sendMessage({
      type: 'join',
      payload: { user_id: userId, user_name: userName },
      timestamp: new Date().toISOString(),
    })
  }, [sendMessage, userId, userName])

  // Leave session
  const leave = useCallback(() => {
    sendMessage({
      type: 'leave',
      payload: { user_id: userId },
      timestamp: new Date().toISOString(),
    })
  }, [sendMessage, userId])

  // Update cursor position
  const updateCursor = useCallback(
    (x: number, y: number) => {
      sendMessage({
        type: 'presence',
        payload: { user_id: userId, cursor: { x, y } },
        timestamp: new Date().toISOString(),
      })
    },
    [sendMessage, userId]
  )

  // Update selection
  const updateSelection = useCallback(
    (type: 'node' | 'edge', elementIds: string[]) => {
      sendMessage({
        type: 'presence',
        payload: { user_id: userId, selection: { type, element_ids: elementIds } },
        timestamp: new Date().toISOString(),
      })
    },
    [sendMessage, userId]
  )

  // Acquire lock on element
  const acquireLock = useCallback(
    (elementId: string, elementType: 'node' | 'edge') => {
      sendMessage({
        type: 'lock_acquire',
        payload: { element_id: elementId, element_type: elementType },
        timestamp: new Date().toISOString(),
      })
    },
    [sendMessage]
  )

  // Release lock on element
  const releaseLock = useCallback(
    (elementId: string) => {
      sendMessage({
        type: 'lock_release',
        payload: { element_id: elementId },
        timestamp: new Date().toISOString(),
      })
    },
    [sendMessage]
  )

  // Broadcast change
  const broadcastChange = useCallback(
    (operation: EditOperation) => {
      sendMessage({
        type: 'change',
        payload: { operation },
        timestamp: new Date().toISOString(),
      })
    },
    [sendMessage]
  )

  // Handle incoming messages
  const handleMessage = useCallback(
    (event: MessageEvent) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data)

        switch (message.type) {
          case 'user_joined': {
            const payload = message.payload
            if (payload.session) {
              // Initial session state
              setSession(payload.session)
              setUsers(Object.values(payload.session.users))
              setLocks(Object.values(payload.session.locks))
            } else if (payload.user) {
              // New user joined
              setUsers((prev) => [...prev, payload.user])
              callbacksRef.current.onUserJoined?.(payload.user)
            }
            break
          }

          case 'user_left': {
            const { user_id } = message.payload
            setUsers((prev) => prev.filter((u) => u.user_id !== user_id))
            callbacksRef.current.onUserLeft?.(user_id)
            break
          }

          case 'presence_update': {
            const { user_id, cursor, selection } = message.payload
            setUsers((prev) =>
              prev.map((u) =>
                u.user_id === user_id ? { ...u, cursor, selection, updated_at: new Date().toISOString() } : u
              )
            )
            const user = users.find((u) => u.user_id === user_id)
            if (user) {
              callbacksRef.current.onPresenceUpdate?.(user_id, { ...user, cursor, selection })
            }
            break
          }

          case 'lock_acquired': {
            const { lock } = message.payload
            setLocks((prev) => [...prev.filter((l) => l.element_id !== lock.element_id), lock])
            callbacksRef.current.onLockAcquired?.(lock)
            break
          }

          case 'lock_released': {
            const { element_id } = message.payload
            setLocks((prev) => prev.filter((l) => l.element_id !== element_id))
            callbacksRef.current.onLockReleased?.(element_id)
            break
          }

          case 'lock_failed': {
            const { element_id, reason, current_lock } = message.payload
            callbacksRef.current.onLockFailed?.(element_id, reason, current_lock)
            break
          }

          case 'change_applied': {
            const { operation } = message.payload
            callbacksRef.current.onChangeApplied?.(operation)
            break
          }

          case 'error': {
            const { message: errorMsg } = message.payload
            callbacksRef.current.onError?.(errorMsg)
            break
          }
        }
      } catch (error) {
        console.error('Failed to handle message:', error)
      }
    },
    [users]
  )

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (!workflowId || !enabled) return

    const base = baseURL || import.meta.env.VITE_API_URL || 'http://localhost:8080'
    const wsProtocol = base.startsWith('https') ? 'wss' : 'ws'
    const url = base.replace(/^https?:\/\//, '')
    const wsURL = `${wsProtocol}://${url}/api/v1/workflows/${workflowId}/collaborate`

    try {
      const ws = new WebSocket(wsURL)

      ws.onopen = () => {
        setConnected(true)
        setReconnecting(false)
        reconnectAttemptsRef.current = 0

        // Auto-join on connection
        setTimeout(() => {
          join()
        }, 100)
      }

      ws.onclose = () => {
        setConnected(false)

        // Attempt to reconnect
        if (reconnectAttemptsRef.current < 10) {
          setReconnecting(true)
          const delay = Math.min(3000 * (reconnectAttemptsRef.current + 1), 30000)
          reconnectTimerRef.current = window.setTimeout(() => {
            reconnectAttemptsRef.current++
            connect()
          }, delay)
        }
      }

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        callbacksRef.current.onError?.('WebSocket connection error')
      }

      ws.onmessage = handleMessage

      wsRef.current = ws
    } catch (error) {
      console.error('Failed to create WebSocket:', error)
    }
  }, [workflowId, enabled, baseURL, join, handleMessage])

  // Connect on mount
  useEffect(() => {
    if (workflowId && enabled) {
      connect()
    }

    return () => {
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current)
      }
      if (wsRef.current) {
        leave()
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [workflowId, enabled, connect, leave])

  // Helper functions
  const isElementLocked = useCallback(
    (elementId: string) => {
      return locks.some((lock) => lock.element_id === elementId)
    },
    [locks]
  )

  const getElementLock = useCallback(
    (elementId: string) => {
      return locks.find((lock) => lock.element_id === elementId)
    },
    [locks]
  )

  const isLockedByMe = useCallback(
    (elementId: string) => {
      const lock = getElementLock(elementId)
      return lock ? lock.user_id === userId : false
    },
    [getElementLock, userId]
  )

  const isLockedByOther = useCallback(
    (elementId: string) => {
      const lock = getElementLock(elementId)
      return lock ? lock.user_id !== userId : false
    },
    [getElementLock, userId]
  )

  return {
    connected,
    reconnecting,
    session,
    users,
    locks,
    join,
    leave,
    updateCursor,
    updateSelection,
    acquireLock,
    releaseLock,
    broadcastChange,
    isElementLocked,
    getElementLock,
    isLockedByMe,
    isLockedByOther,
  }
}
