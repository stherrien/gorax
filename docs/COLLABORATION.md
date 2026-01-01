# Workflow Collaboration Feature

## Overview

The collaboration feature enables multiple users to edit workflows simultaneously in real-time. Users can see each other's cursors, track who is editing which nodes, and receive instant updates when changes are made.

## Architecture

### Backend Components

#### 1. Collaboration Service (`internal/collaboration/service.go`)

The service manages editing sessions and implements the core collaboration logic:

- **JoinSession**: Adds a user to an editing session
- **LeaveSession**: Removes a user and releases their locks
- **UpdatePresence**: Updates cursor position and selection
- **AcquireLock**: Locks an element for editing (prevents conflicts)
- **ReleaseLock**: Releases a lock on an element
- **GetActiveUsers**: Returns all active users in a session
- **GetActiveLocks**: Returns all active locks in a session
- **CleanupInactiveSessions**: Removes stale sessions

#### 2. Collaboration Hub (`internal/collaboration/hub.go`)

The hub manages WebSocket connections and message broadcasting:

- **RegisterClient**: Registers a WebSocket client for collaboration
- **UnregisterClient**: Unregisters a client
- **BroadcastToWorkflow**: Broadcasts to all users in a workflow
- **BroadcastToUser**: Sends message to specific user
- **BroadcastToOthers**: Broadcasts to all except sender

#### 3. Collaboration Handler (`internal/api/handlers/collaboration_handler.go`)

Handles WebSocket connections and processes collaboration messages:

- **HandleWorkflowCollaboration**: WebSocket endpoint `/api/v1/workflows/:id/collaborate`
- **Message Types**:
  - `join`: User joins session
  - `leave`: User leaves session
  - `presence`: Cursor/selection update
  - `lock_acquire`: Request element lock
  - `lock_release`: Release element lock
  - `change`: Broadcast workflow change

### Frontend Components

#### 1. Types (`web/src/types/collaboration.ts`)

TypeScript types for all collaboration data structures:
- `UserPresence`: User information, cursor, selection
- `EditLock`: Lock information
- `EditSession`: Session state with users and locks
- `EditOperation`: Change operations
- `WebSocketMessage`: All message types and payloads

#### 2. useCollaboration Hook (`web/src/hooks/useCollaboration.ts`)

React hook for collaboration features:

```typescript
const {
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
} = useCollaboration(workflowId, userId, userName, options)
```

#### 3. UI Components

- **UserPresenceIndicator**: Shows connection status and active users
- **CollaboratorList**: Lists all collaborators with avatars
- **CollaboratorCursors**: Displays other users' cursors on canvas
- **NodeLockIndicator**: Shows lock status on nodes/edges

## Usage

### Backend Setup

1. Create collaboration service and hub:

```go
collabService := collaboration.NewService()
collabHub := collaboration.NewHub(collabService, wsHub, logger)
```

2. Register WebSocket handler:

```go
collabHandler := handlers.NewCollaborationHandler(collabHub, wsHub, logger)
r.Get("/api/v1/workflows/{id}/collaborate", collabHandler.HandleWorkflowCollaboration)
```

3. Start session cleanup routine (optional):

```go
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        cleaned := collabService.CleanupInactiveSessions(1 * time.Hour)
        logger.Info("cleaned inactive sessions", "count", cleaned)
    }
}()
```

### Frontend Integration

1. Import the hook and components:

```typescript
import { useCollaboration } from '../hooks/useCollaboration'
import {
  UserPresenceIndicator,
  CollaboratorList,
  CollaboratorCursors,
  NodeLockIndicator,
} from '../components/collaboration'
```

2. Setup collaboration in WorkflowEditor:

```typescript
const WorkflowEditor = () => {
  const { id: workflowId } = useParams()
  const user = useAuth() // Get current user

  // Setup collaboration
  const collaboration = useCollaboration(
    workflowId,
    user.id,
    user.name,
    {
      enabled: true,
      onUserJoined: (user) => {
        console.log('User joined:', user.user_name)
      },
      onLockAcquired: (lock) => {
        console.log('Lock acquired:', lock.element_id)
      },
      onChangeApplied: (operation) => {
        // Apply remote change to local state
        applyOperation(operation)
      },
    }
  )

  // Track cursor movement on canvas
  const handleCanvasMouseMove = (event: MouseEvent) => {
    collaboration.updateCursor(event.clientX, event.clientY)
  }

  // Lock node before editing
  const handleNodeClick = (node: Node) => {
    collaboration.acquireLock(node.id, 'node')
    setSelectedNode(node)
  }

  // Broadcast changes
  const handleNodeUpdate = (nodeId: string, updates: any) => {
    const operation: EditOperation = {
      type: 'node_update',
      element_id: nodeId,
      data: updates,
      user_id: user.id,
      timestamp: new Date().toISOString(),
    }

    // Update local state
    updateNode(nodeId, updates)

    // Broadcast to others
    collaboration.broadcastChange(operation)
  }

  return (
    <div className="workflow-editor">
      {/* Presence indicator */}
      <UserPresenceIndicator
        connected={collaboration.connected}
        users={collaboration.users}
        currentUserId={user.id}
      />

      {/* Canvas with cursors */}
      <div onMouseMove={handleCanvasMouseMove}>
        <WorkflowCanvas
          nodes={nodes}
          edges={edges}
          onNodeClick={handleNodeClick}
        />

        <CollaboratorCursors
          users={collaboration.users}
          currentUserId={user.id}
        />
      </div>

      {/* Sidebar with collaborator list */}
      <CollaboratorList
        users={collaboration.users}
        currentUserId={user.id}
      />

      {/* Show lock indicators on nodes */}
      {nodes.map(node => {
        const lock = collaboration.getElementLock(node.id)
        return lock ? (
          <NodeLockIndicator
            lock={lock}
            currentUserId={user.id}
          />
        ) : null
      })}
    </div>
  )
}
```

## Message Flow

### Join Workflow

1. Client connects to WebSocket: `ws://host/api/v1/workflows/:id/collaborate`
2. Client sends `join` message with user info
3. Server adds user to session
4. Server sends session state to joining user
5. Server broadcasts `user_joined` to others

### Acquire Lock

1. User clicks on node
2. Client sends `lock_acquire` message
3. Server checks if element is already locked
4. If available: Server creates lock, broadcasts `lock_acquired`
5. If locked: Server sends `lock_failed` to requester

### Broadcast Change

1. User makes a change (add/update/delete node)
2. Client updates local state
3. Client sends `change` message with operation
4. Server broadcasts `change_applied` to others
5. Others receive and apply the change

### Presence Update

1. User moves cursor or selects nodes
2. Client throttles updates (e.g., every 100ms)
3. Client sends `presence` message
4. Server updates user presence
5. Server broadcasts `presence_update` to others

## Best Practices

### Lock Management

1. **Acquire locks early**: Lock elements before showing edit UI
2. **Release locks promptly**: Release when user finishes editing
3. **Auto-release on disconnect**: Server automatically releases locks when user disconnects
4. **Visual feedback**: Show lock indicators on nodes

### Change Broadcasting

1. **Local-first**: Apply changes locally immediately
2. **Then broadcast**: Send changes to others after local update
3. **Optimistic UI**: Don't wait for server confirmation
4. **Handle conflicts**: Use operational transform if needed

### Performance

1. **Throttle cursor updates**: Don't send every mouse move
2. **Debounce presence**: Wait for user to stop moving
3. **Batch changes**: Group multiple operations when possible
4. **Clean up sessions**: Remove inactive sessions periodically

### Error Handling

1. **Reconnect automatically**: Handle disconnections gracefully
2. **Show connection status**: Let users know if they're offline
3. **Queue operations**: Retry failed operations on reconnect
4. **Sync on reconnect**: Fetch latest state after reconnection

## Testing

Run backend tests:

```bash
go test ./internal/collaboration/...
```

Key test scenarios:
- Multiple users joining/leaving
- Concurrent lock acquisitions
- Presence updates
- Session cleanup
- Lock release on disconnect

## Security Considerations

1. **Authentication**: Users must be authenticated to join
2. **Authorization**: Verify user has access to workflow
3. **Tenant isolation**: Users can only join workflows in their tenant
4. **Rate limiting**: Prevent abuse of WebSocket messages
5. **Message validation**: Validate all incoming messages
6. **Lock ownership**: Only lock owner can release

## Future Enhancements

1. **Operational Transform**: Full conflict resolution for concurrent edits
2. **Change History**: Track and replay all changes
3. **Presence Awareness**: Show what users are typing
4. **Chat**: Built-in collaboration chat
5. **Permissions**: Fine-grained edit permissions
6. **Offline Mode**: Queue changes when offline
7. **Version Control**: Integrate with workflow versions
