package collaboration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_MultiUserEditSession tests multiple users joining an edit session
func TestIntegration_MultiUserEditSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	service := NewService()

	workflowID := "workflow-123"

	// Step 1: User1 joins
	presence1, err := service.JoinSession(ctx, workflowID, "user-1", "Alice")
	require.NoError(t, err)
	require.NotNil(t, presence1)
	assert.Equal(t, "user-1", presence1.UserID)
	assert.Equal(t, "Alice", presence1.UserName)
	assert.NotEmpty(t, presence1.Color)
	t.Logf("✓ User1 (Alice) joined session with color %s", presence1.Color)

	// Step 2: User2 joins
	presence2, err := service.JoinSession(ctx, workflowID, "user-2", "Bob")
	require.NoError(t, err)
	assert.Equal(t, "user-2", presence2.UserID)
	t.Logf("✓ User2 (Bob) joined session")

	// Step 3: Verify session state
	session := service.GetSession(workflowID)
	require.NotNil(t, session)
	assert.Len(t, session.Users, 2)
	t.Logf("✓ Session has %d active users", len(session.Users))

	// Step 4: User3 joins
	presence3, err := service.JoinSession(ctx, workflowID, "user-3", "Charlie")
	require.NoError(t, err)
	assert.Equal(t, "user-3", presence3.UserID)
	t.Logf("✓ User3 (Charlie) joined session")

	// Step 5: User1 leaves
	err = service.LeaveSession(ctx, workflowID, "user-1")
	require.NoError(t, err)
	t.Logf("✓ User1 (Alice) left session")

	// Step 6: Verify User1 is removed
	session = service.GetSession(workflowID)
	require.NotNil(t, session)
	assert.Len(t, session.Users, 2)
	assert.NotContains(t, session.Users, "user-1")
	assert.Contains(t, session.Users, "user-2")
	assert.Contains(t, session.Users, "user-3")
	t.Logf("✓ Session now has %d active users", len(session.Users))

	// Step 7: All users leave
	err = service.LeaveSession(ctx, workflowID, "user-2")
	require.NoError(t, err)
	err = service.LeaveSession(ctx, workflowID, "user-3")
	require.NoError(t, err)
	t.Logf("✓ All users left session")

	// Step 8: Session should be removed (empty)
	session = service.GetSession(workflowID)
	assert.Nil(t, session)
	t.Logf("✓ Session is now removed")
}

// TestIntegration_EditLockAcquisition tests lock acquisition and release
func TestIntegration_EditLockAcquisition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	service := NewService()

	workflowID := "workflow-456"

	// Step 1: Users join
	_, err := service.JoinSession(ctx, workflowID, "user-1", "Alice")
	require.NoError(t, err)
	_, err = service.JoinSession(ctx, workflowID, "user-2", "Bob")
	require.NoError(t, err)
	t.Logf("✓ Users joined session")

	// Step 2: User1 acquires lock on node-1
	lock1, err := service.AcquireLock(ctx, workflowID, "user-1", "node-1", "node")
	require.NoError(t, err)
	require.NotNil(t, lock1)
	assert.Equal(t, "node-1", lock1.ElementID)
	assert.Equal(t, "user-1", lock1.UserID)
	t.Logf("✓ User1 acquired lock on node-1")

	// Step 3: User2 tries to acquire same lock (should fail)
	_, err = service.AcquireLock(ctx, workflowID, "user-2", "node-1", "node")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "locked")
	t.Logf("✓ User2 blocked from acquiring locked node")

	// Step 4: User2 acquires lock on different node (should succeed)
	lock2, err := service.AcquireLock(ctx, workflowID, "user-2", "node-2", "node")
	require.NoError(t, err)
	assert.Equal(t, "node-2", lock2.ElementID)
	t.Logf("✓ User2 acquired lock on node-2")

	// Step 5: Verify session has 2 locks
	locks := service.GetActiveLocks(workflowID)
	assert.Len(t, locks, 2)
	t.Logf("✓ Session has %d active locks", len(locks))

	// Step 6: User1 releases lock
	err = service.ReleaseLock(ctx, workflowID, "user-1", "node-1")
	require.NoError(t, err)
	t.Logf("✓ User1 released lock on node-1")

	// Step 7: User2 can now acquire lock on node-1
	lock3, err := service.AcquireLock(ctx, workflowID, "user-2", "node-1", "node")
	require.NoError(t, err)
	assert.Equal(t, "node-1", lock3.ElementID)
	assert.Equal(t, "user-2", lock3.UserID)
	t.Logf("✓ User2 acquired lock on node-1 after release")

	// Step 8: User leaves - locks should be automatically released
	err = service.LeaveSession(ctx, workflowID, "user-2")
	require.NoError(t, err)

	session := service.GetSession(workflowID)
	require.NotNil(t, session) // user-1 still in session
	assert.Len(t, session.Locks, 0, "user-2's locks should be released when they left")
	t.Logf("✓ Locks automatically released when user left")
}

// TestIntegration_PresenceUpdates tests cursor and selection tracking
func TestIntegration_PresenceUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	service := NewService()

	workflowID := "workflow-789"

	// Step 1: User joins
	_, err := service.JoinSession(ctx, workflowID, "user-1", "Alice")
	require.NoError(t, err)
	t.Logf("✓ User joined session")

	// Step 2: Update cursor position
	cursor := &CursorPosition{X: 100.5, Y: 200.5}
	err = service.UpdatePresence(ctx, workflowID, "user-1", cursor, nil)
	require.NoError(t, err)
	t.Logf("✓ Updated cursor position to (%.1f, %.1f)", cursor.X, cursor.Y)

	// Step 3: Verify cursor updated
	session := service.GetSession(workflowID)
	require.NotNil(t, session)
	userPresence := session.Users["user-1"]
	require.NotNil(t, userPresence)
	require.NotNil(t, userPresence.Cursor)
	assert.Equal(t, cursor.X, userPresence.Cursor.X)
	assert.Equal(t, cursor.Y, userPresence.Cursor.Y)
	t.Logf("✓ Cursor position verified")

	// Step 4: Update selection
	selection := &Selection{
		Type:       "node",
		ElementIDs: []string{"node-1", "node-2"},
	}
	err = service.UpdatePresence(ctx, workflowID, "user-1", nil, selection)
	require.NoError(t, err)
	t.Logf("✓ Updated selection: %d nodes", len(selection.ElementIDs))

	// Step 5: Verify selection updated
	session = service.GetSession(workflowID)
	require.NotNil(t, session)
	userPresence = session.Users["user-1"]
	require.NotNil(t, userPresence.Selection)
	assert.Equal(t, "node", userPresence.Selection.Type)
	assert.ElementsMatch(t, selection.ElementIDs, userPresence.Selection.ElementIDs)
	t.Logf("✓ Selection verified")

	// Step 6: Update both cursor and selection
	newCursor := &CursorPosition{X: 300, Y: 400}
	newSelection := &Selection{
		Type:       "edge",
		ElementIDs: []string{"edge-1"},
	}
	err = service.UpdatePresence(ctx, workflowID, "user-1", newCursor, newSelection)
	require.NoError(t, err)
	t.Logf("✓ Updated both cursor and selection")

	// Step 7: Verify both updated
	session = service.GetSession(workflowID)
	require.NotNil(t, session)
	userPresence = session.Users["user-1"]
	assert.Equal(t, newCursor.X, userPresence.Cursor.X)
	assert.Equal(t, "edge", userPresence.Selection.Type)
	t.Logf("✓ Both cursor and selection verified")
}

// TestIntegration_ConcurrentLockAcquisition tests race conditions
func TestIntegration_ConcurrentLockAcquisition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	service := NewService()

	workflowID := "workflow-concurrent"
	numUsers := 10
	elementID := "node-1"

	// All users join
	for i := 0; i < numUsers; i++ {
		userID := "user-" + string(rune('A'+i))
		userName := "User " + string(rune('A'+i))
		_, err := service.JoinSession(ctx, workflowID, userID, userName)
		require.NoError(t, err)
	}
	t.Logf("✓ %d users joined session", numUsers)

	// All users try to acquire the same lock concurrently
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		userID := "user-" + string(rune('A'+i))
		go func(uid string) {
			defer wg.Done()

			_, err := service.AcquireLock(ctx, workflowID, uid, elementID, "node")
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(userID)
	}

	wg.Wait()

	// Only one user should have successfully acquired the lock
	assert.Equal(t, 1, successCount, "only one user should acquire the lock")
	t.Logf("✓ Only 1 out of %d concurrent lock attempts succeeded", numUsers)

	// Verify session has exactly one lock
	locks := service.GetActiveLocks(workflowID)
	assert.Len(t, locks, 1)
	t.Logf("✓ Session has exactly 1 lock")
}

// TestIntegration_SessionCleanup tests automatic cleanup
func TestIntegration_SessionCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	service := NewService()

	workflowID := "workflow-cleanup"

	// Step 1: User joins
	_, err := service.JoinSession(ctx, workflowID, "user-1", "Alice")
	require.NoError(t, err)

	// Step 2: Acquire multiple locks
	locks := []string{"node-1", "node-2", "node-3"}
	for _, lockID := range locks {
		_, err := service.AcquireLock(ctx, workflowID, "user-1", lockID, "node")
		require.NoError(t, err)
	}
	t.Logf("✓ User acquired %d locks", len(locks))

	// Step 3: Verify locks exist
	activeLocks := service.GetActiveLocks(workflowID)
	assert.Len(t, activeLocks, len(locks))

	// Step 4: User leaves
	err = service.LeaveSession(ctx, workflowID, "user-1")
	require.NoError(t, err)
	t.Logf("✓ User left session")

	// Step 5: Session should be cleaned up (empty)
	session := service.GetSession(workflowID)
	assert.Nil(t, session, "session should be removed when all users leave")
	t.Logf("✓ Session cleaned up")
}

// TestIntegration_MultiWorkflowSessions tests isolation between workflows
func TestIntegration_MultiWorkflowSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	service := NewService()

	workflow1 := "workflow-A"
	workflow2 := "workflow-B"

	// Step 1: Join both workflows
	_, err := service.JoinSession(ctx, workflow1, "user-1", "Alice")
	require.NoError(t, err)
	_, err = service.JoinSession(ctx, workflow2, "user-1", "Alice")
	require.NoError(t, err)
	t.Logf("✓ User joined 2 different workflow sessions")

	// Step 2: Acquire lock in workflow1
	_, err = service.AcquireLock(ctx, workflow1, "user-1", "node-1", "node")
	require.NoError(t, err)

	// Step 3: Should be able to acquire same element ID in workflow2
	_, err = service.AcquireLock(ctx, workflow2, "user-1", "node-1", "node")
	require.NoError(t, err)
	t.Logf("✓ Same node ID can be locked in different workflows")

	// Step 4: Verify sessions are independent
	session1 := service.GetSession(workflow1)
	session2 := service.GetSession(workflow2)

	require.NotNil(t, session1)
	require.NotNil(t, session2)

	assert.Len(t, session1.Users, 1)
	assert.Len(t, session2.Users, 1)
	assert.Len(t, session1.Locks, 1)
	assert.Len(t, session2.Locks, 1)
	t.Logf("✓ Sessions are properly isolated")

	// Step 5: Leave workflow1
	err = service.LeaveSession(ctx, workflow1, "user-1")
	require.NoError(t, err)

	// Step 6: User should still be in workflow2
	session2 = service.GetSession(workflow2)
	require.NotNil(t, session2)
	assert.Len(t, session2.Users, 1)
	assert.Len(t, session2.Locks, 1)
	t.Logf("✓ Leaving one session doesn't affect others")
}

// TestIntegration_SessionCleanupByAge tests cleanup of inactive sessions
func TestIntegration_SessionCleanupByAge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	service := NewService()

	// Create sessions
	_, err := service.JoinSession(ctx, "workflow-old", "user-1", "Alice")
	require.NoError(t, err)
	_, err = service.JoinSession(ctx, "workflow-new", "user-2", "Bob")
	require.NoError(t, err)

	// Get both sessions and verify they exist
	oldSession := service.GetSession("workflow-old")
	newSession := service.GetSession("workflow-new")
	require.NotNil(t, oldSession)
	require.NotNil(t, newSession)

	// Manipulate the old session's UpdatedAt to simulate old age
	service.mu.Lock()
	if s, ok := service.sessions["workflow-old"]; ok {
		s.UpdatedAt = time.Now().Add(-2 * time.Hour)
	}
	service.mu.Unlock()

	// Cleanup sessions older than 1 hour
	cleaned := service.CleanupInactiveSessions(1 * time.Hour)
	assert.Equal(t, 1, cleaned)
	t.Logf("✓ Cleaned up %d inactive session(s)", cleaned)

	// Verify old session is gone
	oldSession = service.GetSession("workflow-old")
	assert.Nil(t, oldSession)

	// Verify new session still exists
	newSession = service.GetSession("workflow-new")
	assert.NotNil(t, newSession)
	t.Logf("✓ New session still exists")
}

// TestIntegration_GetActiveUsers tests retrieving active users
func TestIntegration_GetActiveUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	service := NewService()

	workflowID := "workflow-users"

	// Add multiple users
	users := []struct {
		id   string
		name string
	}{
		{"user-1", "Alice"},
		{"user-2", "Bob"},
		{"user-3", "Charlie"},
	}

	for _, u := range users {
		_, err := service.JoinSession(ctx, workflowID, u.id, u.name)
		require.NoError(t, err)
	}

	// Get active users
	activeUsers := service.GetActiveUsers(workflowID)
	assert.Len(t, activeUsers, 3)
	t.Logf("✓ Retrieved %d active users", len(activeUsers))

	// Verify user details
	userNames := make([]string, 0, len(activeUsers))
	for _, u := range activeUsers {
		userNames = append(userNames, u.UserName)
	}
	assert.Contains(t, userNames, "Alice")
	assert.Contains(t, userNames, "Bob")
	assert.Contains(t, userNames, "Charlie")

	// Remove one user
	err := service.LeaveSession(ctx, workflowID, "user-2")
	require.NoError(t, err)

	// Verify user count
	activeUsers = service.GetActiveUsers(workflowID)
	assert.Len(t, activeUsers, 2)
	t.Logf("✓ After removal: %d active users", len(activeUsers))
}

// TestIntegration_LockRefresh tests that the same user can re-acquire their own lock
func TestIntegration_LockRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	service := NewService()

	workflowID := "workflow-refresh"

	// User joins
	_, err := service.JoinSession(ctx, workflowID, "user-1", "Alice")
	require.NoError(t, err)

	// Acquire lock
	lock1, err := service.AcquireLock(ctx, workflowID, "user-1", "node-1", "node")
	require.NoError(t, err)
	originalAcquiredAt := lock1.AcquiredAt

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Re-acquire the same lock (refresh)
	lock2, err := service.AcquireLock(ctx, workflowID, "user-1", "node-1", "node")
	require.NoError(t, err)
	assert.Equal(t, lock1.ElementID, lock2.ElementID)
	assert.Equal(t, lock1.UserID, lock2.UserID)
	assert.True(t, lock2.AcquiredAt.After(originalAcquiredAt) || lock2.AcquiredAt.Equal(originalAcquiredAt))
	t.Logf("✓ Lock refresh successful")
}
