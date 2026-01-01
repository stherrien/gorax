package collaboration

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// Service handles collaboration business logic
type Service struct {
	sessions map[string]*EditSession
	mu       sync.RWMutex
}

// NewService creates a new collaboration service
func NewService() *Service {
	return &Service{
		sessions: make(map[string]*EditSession),
	}
}

// JoinSession adds a user to an editing session
func (s *Service) JoinSession(ctx context.Context, workflowID, userID, userName string) (*UserPresence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessions[workflowID]
	if session == nil {
		session = &EditSession{
			WorkflowID: workflowID,
			Users:      make(map[string]*UserPresence),
			Locks:      make(map[string]*EditLock),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		s.sessions[workflowID] = session
	}

	// Check if user already exists (rejoin)
	presence := session.Users[userID]
	if presence == nil {
		presence = &UserPresence{
			UserID:   userID,
			UserName: userName,
			Color:    generateUserColor(),
			JoinedAt: time.Now(),
		}
	} else {
		// Update user name if changed
		presence.UserName = userName
	}

	presence.UpdatedAt = time.Now()
	session.Users[userID] = presence
	session.UpdatedAt = time.Now()

	return presence, nil
}

// LeaveSession removes a user from an editing session
func (s *Service) LeaveSession(ctx context.Context, workflowID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessions[workflowID]
	if session == nil {
		return nil
	}

	// Remove user
	delete(session.Users, userID)

	// Release all locks held by this user
	for elementID, lock := range session.Locks {
		if lock.UserID == userID {
			delete(session.Locks, elementID)
		}
	}

	session.UpdatedAt = time.Now()

	// Clean up empty sessions
	if len(session.Users) == 0 {
		delete(s.sessions, workflowID)
	}

	return nil
}

// UpdatePresence updates a user's cursor position and/or selection
func (s *Service) UpdatePresence(ctx context.Context, workflowID, userID string, cursor *CursorPosition, selection *Selection) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessions[workflowID]
	if session == nil {
		return fmt.Errorf("session not found for workflow: %s", workflowID)
	}

	presence := session.Users[userID]
	if presence == nil {
		return fmt.Errorf("user not found in session: %s", userID)
	}

	if cursor != nil {
		presence.Cursor = cursor
	}

	if selection != nil {
		presence.Selection = selection
	}

	presence.UpdatedAt = time.Now()
	session.UpdatedAt = time.Now()

	return nil
}

// AcquireLock attempts to acquire a lock on an element (node or edge)
func (s *Service) AcquireLock(ctx context.Context, workflowID, userID, elementID, elementType string) (*EditLock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessions[workflowID]
	if session == nil {
		return nil, fmt.Errorf("session not found for workflow: %s", workflowID)
	}

	presence := session.Users[userID]
	if presence == nil {
		return nil, fmt.Errorf("user not found in session: %s", userID)
	}

	// Check if element is already locked
	existingLock := session.Locks[elementID]
	if existingLock != nil {
		// Allow same user to re-acquire their own lock (refresh)
		if existingLock.UserID != userID {
			return nil, fmt.Errorf("element already locked by user: %s", existingLock.UserName)
		}
	}

	// Create or update lock
	lock := &EditLock{
		ElementID:   elementID,
		ElementType: elementType,
		UserID:      userID,
		UserName:    presence.UserName,
		AcquiredAt:  time.Now(),
	}

	session.Locks[elementID] = lock
	session.UpdatedAt = time.Now()

	return lock, nil
}

// ReleaseLock releases a lock on an element
func (s *Service) ReleaseLock(ctx context.Context, workflowID, userID, elementID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessions[workflowID]
	if session == nil {
		return fmt.Errorf("session not found for workflow: %s", workflowID)
	}

	lock := session.Locks[elementID]
	if lock == nil {
		// No lock exists, nothing to release
		return nil
	}

	// Only lock owner can release
	if lock.UserID != userID {
		return fmt.Errorf("cannot release lock owned by another user")
	}

	delete(session.Locks, elementID)
	session.UpdatedAt = time.Now()

	return nil
}

// GetSession retrieves a session by workflow ID
func (s *Service) GetSession(workflowID string) *EditSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sessions[workflowID]
}

// GetActiveUsers returns all active users in a session
func (s *Service) GetActiveUsers(workflowID string) []*UserPresence {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.sessions[workflowID]
	if session == nil {
		return []*UserPresence{}
	}

	users := make([]*UserPresence, 0, len(session.Users))
	for _, user := range session.Users {
		users = append(users, user)
	}

	return users
}

// GetActiveLocks returns all active locks in a session
func (s *Service) GetActiveLocks(workflowID string) []*EditLock {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.sessions[workflowID]
	if session == nil {
		return []*EditLock{}
	}

	locks := make([]*EditLock, 0, len(session.Locks))
	for _, lock := range session.Locks {
		locks = append(locks, lock)
	}

	return locks
}

// CleanupInactiveSessions removes sessions that haven't been updated in the specified duration
func (s *Service) CleanupInactiveSessions(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	cleaned := 0

	for workflowID, session := range s.sessions {
		if session.UpdatedAt.Before(cutoff) {
			delete(s.sessions, workflowID)
			cleaned++
		}
	}

	return cleaned
}

// generateUserColor generates a random color for a user using cryptographically secure random
func generateUserColor() string {
	colors := []string{
		"#3B82F6", // blue
		"#10B981", // green
		"#F59E0B", // amber
		"#EF4444", // red
		"#8B5CF6", // purple
		"#EC4899", // pink
		"#14B8A6", // teal
		"#F97316", // orange
		"#6366F1", // indigo
		"#84CC16", // lime
	}

	// Use crypto/rand for secure random selection
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
	if err != nil {
		// Fallback to first color if random generation fails
		return colors[0]
	}

	return colors[n.Int64()]
}
