/**
 * Collaboration types for real-time workflow editing
 */

export type MessageType =
  | 'join'
  | 'leave'
  | 'presence'
  | 'lock_acquire'
  | 'lock_release'
  | 'change'
  | 'user_joined'
  | 'user_left'
  | 'presence_update'
  | 'lock_acquired'
  | 'lock_released'
  | 'lock_failed'
  | 'change_applied'
  | 'error'

export interface CursorPosition {
  x: number
  y: number
}

export interface Selection {
  type: 'node' | 'edge'
  element_ids: string[]
}

export interface UserPresence {
  user_id: string
  user_name: string
  color: string
  cursor?: CursorPosition
  selection?: Selection
  joined_at: string
  updated_at: string
}

export interface EditLock {
  element_id: string
  element_type: 'node' | 'edge'
  user_id: string
  user_name: string
  acquired_at: string
}

export interface EditSession {
  workflow_id: string
  users: Record<string, UserPresence>
  locks: Record<string, EditLock>
  created_at: string
  updated_at: string
}

export type OperationType =
  | 'node_add'
  | 'node_update'
  | 'node_delete'
  | 'node_move'
  | 'edge_add'
  | 'edge_update'
  | 'edge_delete'

export interface EditOperation {
  type: OperationType
  element_id: string
  data: any
  user_id: string
  timestamp: string
}

export interface WebSocketMessage {
  type: MessageType
  payload?: any
  timestamp: string
}

// Message Payloads

export interface JoinPayload {
  user_id: string
  user_name: string
}

export interface LeavePayload {
  user_id: string
}

export interface PresencePayload {
  user_id: string
  cursor?: CursorPosition
  selection?: Selection
}

export interface LockAcquirePayload {
  element_id: string
  element_type: 'node' | 'edge'
}

export interface LockReleasePayload {
  element_id: string
}

export interface ChangePayload {
  operation: EditOperation
}

export interface UserJoinedPayload {
  user: UserPresence
}

export interface UserLeftPayload {
  user_id: string
}

export interface LockAcquiredPayload {
  lock: EditLock
}

export interface LockReleasedPayload {
  element_id: string
}

export interface LockFailedPayload {
  element_id: string
  reason: string
  current_lock?: EditLock
}

export interface ChangeAppliedPayload {
  operation: EditOperation
}

export interface ErrorPayload {
  message: string
  code?: string
}

export interface SessionState {
  session: EditSession
}
