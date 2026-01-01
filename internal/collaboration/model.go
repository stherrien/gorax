package collaboration

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of collaboration message
type MessageType string

const (
	MessageTypeJoin          MessageType = "join"
	MessageTypeLeave         MessageType = "leave"
	MessageTypePresence      MessageType = "presence"
	MessageTypeLockAcquire   MessageType = "lock_acquire"
	MessageTypeLockRelease   MessageType = "lock_release"
	MessageTypeChange        MessageType = "change"
	MessageTypeUserJoined    MessageType = "user_joined"
	MessageTypeUserLeft      MessageType = "user_left"
	MessageTypePresenceUpdate MessageType = "presence_update"
	MessageTypeLockAcquired  MessageType = "lock_acquired"
	MessageTypeLockReleased  MessageType = "lock_released"
	MessageTypeLockFailed    MessageType = "lock_failed"
	MessageTypeChangeApplied MessageType = "change_applied"
	MessageTypeError         MessageType = "error"
)

// EditSession represents an active editing session for a workflow
type EditSession struct {
	WorkflowID string                   `json:"workflow_id"`
	Users      map[string]*UserPresence `json:"users"`
	Locks      map[string]*EditLock     `json:"locks"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
}

// UserPresence represents a user's presence in an editing session
type UserPresence struct {
	UserID    string          `json:"user_id"`
	UserName  string          `json:"user_name"`
	Color     string          `json:"color"`
	Cursor    *CursorPosition `json:"cursor,omitempty"`
	Selection *Selection      `json:"selection,omitempty"`
	JoinedAt  time.Time       `json:"joined_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// CursorPosition represents the position of a user's cursor on the canvas
type CursorPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Selection represents a user's current selection (node/edge)
type Selection struct {
	Type     string   `json:"type"` // "node" or "edge"
	ElementIDs []string `json:"element_ids"`
}

// EditLock represents a lock on a node or edge being edited
type EditLock struct {
	ElementID  string    `json:"element_id"`
	ElementType string   `json:"element_type"` // "node" or "edge"
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	AcquiredAt time.Time `json:"acquired_at"`
}

// EditOperation represents a change operation using operational transform
type EditOperation struct {
	Type      OperationType   `json:"type"`
	ElementID string          `json:"element_id"`
	Data      json.RawMessage `json:"data"`
	UserID    string          `json:"user_id"`
	Timestamp time.Time       `json:"timestamp"`
}

// OperationType represents the type of edit operation
type OperationType string

const (
	OperationTypeNodeAdd    OperationType = "node_add"
	OperationTypeNodeUpdate OperationType = "node_update"
	OperationTypeNodeDelete OperationType = "node_delete"
	OperationTypeNodeMove   OperationType = "node_move"
	OperationTypeEdgeAdd    OperationType = "edge_add"
	OperationTypeEdgeUpdate OperationType = "edge_update"
	OperationTypeEdgeDelete OperationType = "edge_delete"
)

// WebSocketMessage represents a WebSocket message between client and server
type WebSocketMessage struct {
	Type      MessageType     `json:"type"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// JoinPayload represents the payload for join message
type JoinPayload struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// LeavePayload represents the payload for leave message
type LeavePayload struct {
	UserID string `json:"user_id"`
}

// PresencePayload represents the payload for presence update
type PresencePayload struct {
	UserID    string          `json:"user_id"`
	Cursor    *CursorPosition `json:"cursor,omitempty"`
	Selection *Selection      `json:"selection,omitempty"`
}

// LockAcquirePayload represents the payload for lock acquisition
type LockAcquirePayload struct {
	ElementID   string `json:"element_id"`
	ElementType string `json:"element_type"`
}

// LockReleasePayload represents the payload for lock release
type LockReleasePayload struct {
	ElementID string `json:"element_id"`
}

// ChangePayload represents the payload for a change operation
type ChangePayload struct {
	Operation EditOperation `json:"operation"`
}

// UserJoinedPayload represents the payload when a user joins
type UserJoinedPayload struct {
	User *UserPresence `json:"user"`
}

// UserLeftPayload represents the payload when a user leaves
type UserLeftPayload struct {
	UserID string `json:"user_id"`
}

// LockAcquiredPayload represents the payload when a lock is acquired
type LockAcquiredPayload struct {
	Lock *EditLock `json:"lock"`
}

// LockReleasedPayload represents the payload when a lock is released
type LockReleasedPayload struct {
	ElementID string `json:"element_id"`
}

// LockFailedPayload represents the payload when lock acquisition fails
type LockFailedPayload struct {
	ElementID string     `json:"element_id"`
	Reason    string     `json:"reason"`
	CurrentLock *EditLock `json:"current_lock,omitempty"`
}

// ChangeAppliedPayload represents the payload when a change is applied
type ChangeAppliedPayload struct {
	Operation EditOperation `json:"operation"`
}

// ErrorPayload represents the payload for error messages
type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// SessionState represents the current state of an editing session
type SessionState struct {
	Session *EditSession `json:"session"`
}
