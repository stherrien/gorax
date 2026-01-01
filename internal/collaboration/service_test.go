package collaboration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_JoinSession(t *testing.T) {
	tests := []struct {
		name          string
		workflowID    string
		userID        string
		userName      string
		expectSession bool
		expectUser    bool
	}{
		{
			name:          "user joins new session",
			workflowID:    "workflow-1",
			userID:        "user-1",
			userName:      "Alice",
			expectSession: true,
			expectUser:    true,
		},
		{
			name:          "second user joins existing session",
			workflowID:    "workflow-1",
			userID:        "user-2",
			userName:      "Bob",
			expectSession: true,
			expectUser:    true,
		},
		{
			name:          "same user joins again (should update)",
			workflowID:    "workflow-1",
			userID:        "user-1",
			userName:      "Alice Updated",
			expectSession: true,
			expectUser:    true,
		},
	}

	service := NewService()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presence, err := service.JoinSession(ctx, tt.workflowID, tt.userID, tt.userName)
			require.NoError(t, err)
			require.NotNil(t, presence)

			assert.Equal(t, tt.userID, presence.UserID)
			assert.Equal(t, tt.userName, presence.UserName)
			assert.NotEmpty(t, presence.Color)
			assert.False(t, presence.JoinedAt.IsZero())

			session := service.GetSession(tt.workflowID)
			if tt.expectSession {
				require.NotNil(t, session)
				assert.Equal(t, tt.workflowID, session.WorkflowID)
			}

			if tt.expectUser {
				require.Contains(t, session.Users, tt.userID)
				assert.Equal(t, tt.userName, session.Users[tt.userID].UserName)
			}
		})
	}
}

func TestService_LeaveSession(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	workflowID := "workflow-1"
	userID := "user-1"

	// Join session first
	_, err := service.JoinSession(ctx, workflowID, userID, "Alice")
	require.NoError(t, err)

	// Leave session
	err = service.LeaveSession(ctx, workflowID, userID)
	require.NoError(t, err)

	// Verify session is removed (no users left)
	session := service.GetSession(workflowID)
	assert.Nil(t, session)

	// Leave non-existent session should not error
	err = service.LeaveSession(ctx, "non-existent", userID)
	assert.NoError(t, err)

	// Leave non-existent user should not error
	err = service.LeaveSession(ctx, workflowID, "non-existent-user")
	assert.NoError(t, err)
}

func TestService_LeaveSession_RemovesLocksAndCleansUp(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	workflowID := "workflow-1"
	userID := "user-1"
	nodeID := "node-1"

	// Join and acquire lock
	_, err := service.JoinSession(ctx, workflowID, userID, "Alice")
	require.NoError(t, err)

	lock, err := service.AcquireLock(ctx, workflowID, userID, nodeID, "node")
	require.NoError(t, err)
	require.NotNil(t, lock)

	// Leave session
	err = service.LeaveSession(ctx, workflowID, userID)
	require.NoError(t, err)

	// Verify session is removed (no users left, so locks are also removed)
	session := service.GetSession(workflowID)
	assert.Nil(t, session)
}

func TestService_UpdatePresence(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	workflowID := "workflow-1"
	userID := "user-1"

	// Join session first
	_, err := service.JoinSession(ctx, workflowID, userID, "Alice")
	require.NoError(t, err)

	// Update cursor position
	cursor := &CursorPosition{X: 100, Y: 200}
	err = service.UpdatePresence(ctx, workflowID, userID, cursor, nil)
	require.NoError(t, err)

	session := service.GetSession(workflowID)
	require.NotNil(t, session.Users[userID].Cursor)
	assert.Equal(t, 100.0, session.Users[userID].Cursor.X)
	assert.Equal(t, 200.0, session.Users[userID].Cursor.Y)

	// Update selection
	selection := &Selection{Type: "node", ElementIDs: []string{"node-1", "node-2"}}
	err = service.UpdatePresence(ctx, workflowID, userID, nil, selection)
	require.NoError(t, err)

	session = service.GetSession(workflowID)
	require.NotNil(t, session.Users[userID].Selection)
	assert.Equal(t, "node", session.Users[userID].Selection.Type)
	assert.Len(t, session.Users[userID].Selection.ElementIDs, 2)

	// Update non-existent session should error
	err = service.UpdatePresence(ctx, "non-existent", userID, cursor, nil)
	assert.Error(t, err)

	// Update non-existent user should error
	err = service.UpdatePresence(ctx, workflowID, "non-existent", cursor, nil)
	assert.Error(t, err)
}

func TestService_AcquireLock(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	workflowID := "workflow-1"
	user1ID := "user-1"
	user2ID := "user-2"
	nodeID := "node-1"

	// Join session
	_, err := service.JoinSession(ctx, workflowID, user1ID, "Alice")
	require.NoError(t, err)

	_, err = service.JoinSession(ctx, workflowID, user2ID, "Bob")
	require.NoError(t, err)

	// User 1 acquires lock
	lock, err := service.AcquireLock(ctx, workflowID, user1ID, nodeID, "node")
	require.NoError(t, err)
	require.NotNil(t, lock)
	assert.Equal(t, nodeID, lock.ElementID)
	assert.Equal(t, "node", lock.ElementType)
	assert.Equal(t, user1ID, lock.UserID)
	assert.Equal(t, "Alice", lock.UserName)

	// User 2 tries to acquire same lock - should fail
	lock, err = service.AcquireLock(ctx, workflowID, user2ID, nodeID, "node")
	assert.Error(t, err)
	assert.Nil(t, lock)

	// User 1 can re-acquire their own lock (refresh)
	lock, err = service.AcquireLock(ctx, workflowID, user1ID, nodeID, "node")
	require.NoError(t, err)
	require.NotNil(t, lock)
	assert.Equal(t, user1ID, lock.UserID)

	// Acquire lock on non-existent session should error
	lock, err = service.AcquireLock(ctx, "non-existent", user1ID, nodeID, "node")
	assert.Error(t, err)
	assert.Nil(t, lock)

	// Acquire lock for non-existent user should error
	lock, err = service.AcquireLock(ctx, workflowID, "non-existent", nodeID, "node")
	assert.Error(t, err)
	assert.Nil(t, lock)
}

func TestService_ReleaseLock(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	workflowID := "workflow-1"
	userID := "user-1"
	nodeID := "node-1"

	// Join session and acquire lock
	_, err := service.JoinSession(ctx, workflowID, userID, "Alice")
	require.NoError(t, err)

	_, err = service.AcquireLock(ctx, workflowID, userID, nodeID, "node")
	require.NoError(t, err)

	// Release lock
	err = service.ReleaseLock(ctx, workflowID, userID, nodeID)
	require.NoError(t, err)

	// Verify lock is released
	session := service.GetSession(workflowID)
	assert.NotContains(t, session.Locks, nodeID)

	// Another user tries to acquire - should succeed now
	user2ID := "user-2"
	_, err = service.JoinSession(ctx, workflowID, user2ID, "Bob")
	require.NoError(t, err)

	lock, err := service.AcquireLock(ctx, workflowID, user2ID, nodeID, "node")
	require.NoError(t, err)
	require.NotNil(t, lock)
	assert.Equal(t, user2ID, lock.UserID)

	// Release non-existent lock should not error
	err = service.ReleaseLock(ctx, workflowID, userID, "non-existent")
	assert.NoError(t, err)

	// Release lock for non-existent session should error
	err = service.ReleaseLock(ctx, "non-existent", userID, nodeID)
	assert.Error(t, err)
}

func TestService_ReleaseLock_OnlyOwnerCanRelease(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	workflowID := "workflow-1"
	user1ID := "user-1"
	user2ID := "user-2"
	nodeID := "node-1"

	// Join sessions
	_, err := service.JoinSession(ctx, workflowID, user1ID, "Alice")
	require.NoError(t, err)

	_, err = service.JoinSession(ctx, workflowID, user2ID, "Bob")
	require.NoError(t, err)

	// User 1 acquires lock
	_, err = service.AcquireLock(ctx, workflowID, user1ID, nodeID, "node")
	require.NoError(t, err)

	// User 2 tries to release user 1's lock - should fail
	err = service.ReleaseLock(ctx, workflowID, user2ID, nodeID)
	assert.Error(t, err)

	// Lock should still exist
	session := service.GetSession(workflowID)
	assert.Contains(t, session.Locks, nodeID)
	assert.Equal(t, user1ID, session.Locks[nodeID].UserID)
}

func TestService_GetActiveUsers(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	workflowID := "workflow-1"

	// No users initially
	users := service.GetActiveUsers(workflowID)
	assert.Empty(t, users)

	// Add users
	_, err := service.JoinSession(ctx, workflowID, "user-1", "Alice")
	require.NoError(t, err)

	_, err = service.JoinSession(ctx, workflowID, "user-2", "Bob")
	require.NoError(t, err)

	users = service.GetActiveUsers(workflowID)
	assert.Len(t, users, 2)

	// Remove one user
	err = service.LeaveSession(ctx, workflowID, "user-1")
	require.NoError(t, err)

	users = service.GetActiveUsers(workflowID)
	assert.Len(t, users, 1)
	assert.Equal(t, "Bob", users[0].UserName)
}

func TestService_GetActiveLocks(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	workflowID := "workflow-1"
	userID := "user-1"

	// No locks initially
	locks := service.GetActiveLocks(workflowID)
	assert.Empty(t, locks)

	// Join and acquire locks
	_, err := service.JoinSession(ctx, workflowID, userID, "Alice")
	require.NoError(t, err)

	_, err = service.AcquireLock(ctx, workflowID, userID, "node-1", "node")
	require.NoError(t, err)

	_, err = service.AcquireLock(ctx, workflowID, userID, "node-2", "node")
	require.NoError(t, err)

	locks = service.GetActiveLocks(workflowID)
	assert.Len(t, locks, 2)

	// Release one lock
	err = service.ReleaseLock(ctx, workflowID, userID, "node-1")
	require.NoError(t, err)

	locks = service.GetActiveLocks(workflowID)
	assert.Len(t, locks, 1)
	assert.Equal(t, "node-2", locks[0].ElementID)
}

func TestService_CleanupInactiveSessions(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	workflowID := "workflow-1"
	userID := "user-1"

	// Create session
	_, err := service.JoinSession(ctx, workflowID, userID, "Alice")
	require.NoError(t, err)

	// Manually set UpdatedAt to old time
	session := service.GetSession(workflowID)
	session.UpdatedAt = time.Now().Add(-2 * time.Hour)

	// Clean up sessions older than 1 hour
	cleaned := service.CleanupInactiveSessions(1 * time.Hour)
	assert.Equal(t, 1, cleaned)

	// Session should be removed
	session = service.GetSession(workflowID)
	assert.Nil(t, session)
}

func TestService_ConcurrentOperations(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	workflowID := "workflow-1"
	numUsers := 10

	// Concurrent joins
	done := make(chan bool, numUsers)
	for i := 0; i < numUsers; i++ {
		go func(id int) {
			userID := "user-" + string(rune(id))
			userName := "User" + string(rune(id))
			_, err := service.JoinSession(ctx, workflowID, userID, userName)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < numUsers; i++ {
		<-done
	}

	users := service.GetActiveUsers(workflowID)
	assert.Len(t, users, numUsers)
}

func TestGenerateUserColor_SecureRandom(t *testing.T) {
	// Test that generateUserColor returns valid color codes
	colors := make(map[string]int)

	// Generate many colors to test randomness
	for i := 0; i < 100; i++ {
		color := generateUserColor()

		// Verify it's a valid hex color
		assert.True(t, strings.HasPrefix(color, "#"), "color should start with #")
		assert.Len(t, color, 7, "color should be 7 characters (#RRGGBB)")

		// Count occurrences
		colors[color]++
	}

	// Should have generated at least 2 different colors from 100 tries
	// This is a probabilistic test - with 10 colors, getting only 1 is extremely unlikely
	assert.GreaterOrEqual(t, len(colors), 2, "should generate multiple different colors")

	// Verify all colors are from the expected set
	expectedColors := map[string]bool{
		"#3B82F6": true, "#10B981": true, "#F59E0B": true,
		"#EF4444": true, "#8B5CF6": true, "#EC4899": true,
		"#14B8A6": true, "#F97316": true, "#6366F1": true,
		"#84CC16": true,
	}

	for color := range colors {
		assert.True(t, expectedColors[color], "generated color should be from expected set")
	}
}

func TestGenerateUserColor_Deterministic(t *testing.T) {
	// Even though it's random, it should always return a valid color
	color := generateUserColor()
	assert.NotEmpty(t, color)
	assert.True(t, strings.HasPrefix(color, "#"))
	assert.Len(t, color, 7)
}
