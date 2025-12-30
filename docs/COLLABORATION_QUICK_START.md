# Collaboration Feature - Quick Start Guide

## 5-Minute Integration

### Backend (3 steps)

1. **Create services** (in `main.go` or wherever you initialize services):

```go
import "github.com/gorax/gorax/internal/collaboration"

// After creating wsHub
collabService := collaboration.NewService()
collabHub := collaboration.NewHub(collabService, wsHub, logger)
```

2. **Register handler**:

```go
import "github.com/gorax/gorax/internal/api/handlers"

collabHandler := handlers.NewCollaborationHandler(collabHub, wsHub, logger)
r.Get("/api/v1/workflows/{id}/collaborate", collabHandler.HandleWorkflowCollaboration)
```

3. **Optional cleanup routine**:

```go
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        collabService.CleanupInactiveSessions(1 * time.Hour)
    }
}()
```

### Frontend (4 steps)

1. **Import**:

```typescript
import { useCollaboration } from '../hooks/useCollaboration'
import { UserPresenceIndicator } from '../components/collaboration'
```

2. **Initialize hook**:

```typescript
const collaboration = useCollaboration(
  workflowId,
  currentUser.id,
  currentUser.name,
  {
    onChangeApplied: (operation) => {
      // Apply remote changes
      applyRemoteChange(operation)
    }
  }
)
```

3. **Add presence indicator**:

```tsx
<UserPresenceIndicator
  connected={collaboration.connected}
  users={collaboration.users}
  currentUserId={currentUser.id}
/>
```

4. **Broadcast changes**:

```typescript
const handleNodeUpdate = (nodeId: string, updates: any) => {
  // Update local
  updateNode(nodeId, updates)

  // Broadcast
  collaboration.broadcastChange({
    type: 'node_update',
    element_id: nodeId,
    data: updates,
    user_id: currentUser.id,
    timestamp: new Date().toISOString(),
  })
}
```

## Key Patterns

### Acquire Lock Before Edit

```typescript
const handleNodeClick = (node: Node) => {
  if (collaboration.isLockedByOther(node.id)) {
    alert('Node is being edited by another user')
    return
  }
  collaboration.acquireLock(node.id, 'node')
  setSelectedNode(node)
}
```

### Release Lock After Edit

```typescript
const handleNodeDeselect = () => {
  if (selectedNode) {
    collaboration.releaseLock(selectedNode.id)
  }
  setSelectedNode(null)
}
```

### Track Cursor

```typescript
const handleMouseMove = (e: MouseEvent) => {
  // Throttle to every 100ms
  if (Date.now() % 100 < 16) {
    collaboration.updateCursor(e.clientX, e.clientY)
  }
}
```

### Apply Remote Changes

```typescript
const applyRemoteChange = (operation: EditOperation) => {
  switch (operation.type) {
    case 'node_update':
      setNodes(prev =>
        prev.map(n =>
          n.id === operation.element_id
            ? { ...n, data: { ...n.data, ...operation.data }}
            : n
        )
      )
      break
    case 'node_add':
      setNodes(prev => [...prev, operation.data])
      break
    case 'node_delete':
      setNodes(prev => prev.filter(n => n.id !== operation.element_id))
      break
  }
}
```

## Complete Example

See `/web/src/examples/WorkflowEditorWithCollaboration.tsx` for a complete working example.

## Testing

```bash
# Backend tests
go test ./internal/collaboration/... -v

# Expected: All tests pass (10 tests)
```

## Troubleshooting

### "WebSocket connection failed"
- Check that handler is registered at correct endpoint
- Verify authentication middleware is configured
- Check CORS settings allow WebSocket upgrade

### "Lock failed" errors
- Another user has the lock - this is expected behavior
- Show user feedback about who has the lock
- User can wait for lock to be released

### "User not receiving updates"
- Check WebSocket connection status
- Verify user successfully joined session
- Check browser console for errors
- Verify messages are being broadcast to correct room

### "Session not cleaning up"
- Ensure cleanup routine is running
- Check session timeout configuration
- Verify LeaveSession is called on disconnect

## Performance Tips

1. **Throttle cursor updates**: Max 10 per second
2. **Debounce presence**: Wait 200ms after movement stops
3. **Batch changes**: Combine multiple operations
4. **Local-first**: Always update local state immediately
5. **Lazy load**: Only load collaboration for active workflows

## Security Checklist

- ✅ Require authentication for WebSocket connections
- ✅ Verify user has access to workflow
- ✅ Validate all incoming messages
- ✅ Enforce tenant isolation
- ✅ Prevent lock stealing (only owner can release)
- ⚠️ Consider rate limiting for production

## Need More Help?

- Full documentation: `/docs/COLLABORATION.md`
- Implementation details: `/docs/COLLABORATION_IMPLEMENTATION.md`
- Example code: `/web/src/examples/WorkflowEditorWithCollaboration.tsx`
