package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// InMemorySessionStore implements SessionStore using in-memory storage
type InMemorySessionStore struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// NewInMemorySessionStore creates a new in-memory session store
func NewInMemorySessionStore() *InMemorySessionStore {
	store := &InMemorySessionStore{
		sessions: make(map[string]*Session),
	}
	
	// Start cleanup goroutine
	go store.cleanup()
	
	return store
}

// CreateSession creates a new session for a user
func (s *InMemorySessionStore) CreateSession(userID string) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
		Active:    true,
	}

	s.mutex.Lock()
	s.sessions[sessionID] = session
	s.mutex.Unlock()

	return session, nil
}

// GetSession retrieves a session by ID
func (s *InMemorySessionStore) GetSession(sessionID string) (*Session, error) {
	s.mutex.RLock()
	session, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists {
		return nil, errors.New("session not found")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(sessionID)
		return nil, errors.New("session expired")
	}

	return session, nil
}

// DeleteSession removes a session
func (s *InMemorySessionStore) DeleteSession(sessionID string) error {
	s.mutex.Lock()
	delete(s.sessions, sessionID)
	s.mutex.Unlock()
	return nil
}

// IsSessionValid checks if a session is valid
func (s *InMemorySessionStore) IsSessionValid(sessionID string) bool {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return false
	}
	return session.Active && time.Now().Before(session.ExpiresAt)
}

// ExtendSession extends the expiration time of a session
func (s *InMemorySessionStore) ExtendSession(sessionID string, duration time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}

	session.ExpiresAt = time.Now().Add(duration)
	return nil
}

// GetUserSessions returns all active sessions for a user
func (s *InMemorySessionStore) GetUserSessions(userID string) []*Session {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var userSessions []*Session
	for _, session := range s.sessions {
		if session.UserID == userID && session.Active && time.Now().Before(session.ExpiresAt) {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions
}

// RevokeUserSessions revokes all sessions for a user
func (s *InMemorySessionStore) RevokeUserSessions(userID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for sessionID, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, sessionID)
		}
	}

	return nil
}

// UpdateSessionActivity updates session activity information
func (s *InMemorySessionStore) UpdateSessionActivity(sessionID, ipAddress, userAgent string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}

	session.IPAddress = ipAddress
	session.UserAgent = userAgent
	return nil
}

// cleanup removes expired sessions periodically
func (s *InMemorySessionStore) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for sessionID, session := range s.sessions {
			if now.After(session.ExpiresAt) {
				delete(s.sessions, sessionID)
			}
		}
		s.mutex.Unlock()
	}
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// RedisSessionStore can be implemented for production use
// This would store sessions in Redis for scalability and persistence

// DatabaseSessionStore can be implemented for persistent sessions
// This would store sessions in the database