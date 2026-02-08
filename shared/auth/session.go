package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionManager handles session operations
type SessionManager struct {
	redis *redis.Client
}

// NewSessionManager creates new session manager
func NewSessionManager(redisClient *redis.Client) *SessionManager {
	return &SessionManager{
		redis: redisClient,
	}
}

// SessionData represents session information
type SessionData struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	SessionID    string    `json:"session_id"`
	DeviceInfo   string    `json:"device_info,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// ============================================================================
// Session Operations
// ============================================================================

// CreateSession creates a new session
func (s *SessionManager) CreateSession(ctx context.Context, userID uuid.UUID, email, role, sessionID, ipAddress, userAgent string, ttl time.Duration) error {
	session := SessionData{
		UserID:       userID,
		Email:        email,
		Role:         role,
		SessionID:    sessionID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Store session with TTL
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	if err := s.redis.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Track user's active sessions
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	if err := s.redis.SAdd(ctx, userSessionsKey, sessionID).Err(); err != nil {
		return fmt.Errorf("failed to add session to user sessions: %w", err)
	}
	s.redis.Expire(ctx, userSessionsKey, ttl)

	return nil
}

// GetSession retrieves session data
func (s *SessionManager) GetSession(ctx context.Context, userID uuid.UUID, sessionID string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// UpdateSessionActivity updates last activity timestamp
func (s *SessionManager) UpdateSessionActivity(ctx context.Context, userID uuid.UUID, sessionID string) error {
	session, err := s.GetSession(ctx, userID, sessionID)
	if err != nil {
		return err
	}

	session.LastActivity = time.Now()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	
	// Get remaining TTL
	ttl, err := s.redis.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}

	// Update with same TTL
	if err := s.redis.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// DeleteSession removes a session
func (s *SessionManager) DeleteSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Remove from user's active sessions
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	s.redis.SRem(ctx, userSessionsKey, sessionID)

	return nil
}

// DeleteAllUserSessions removes all sessions for a user
func (s *SessionManager) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	// Get all session IDs for user
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	sessionIDs, err := s.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Delete each session
	for _, sessionID := range sessionIDs {
		key := fmt.Sprintf("session:%s:%s", userID, sessionID)
		s.redis.Del(ctx, key)
	}

	// Delete user sessions set
	s.redis.Del(ctx, userSessionsKey)

	return nil
}

// ListUserSessions returns all active sessions for a user
func (s *SessionManager) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*SessionData, error) {
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	sessionIDs, err := s.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	sessions := make([]*SessionData, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := s.GetSession(ctx, userID, sessionID)
		if err != nil {
			// Skip invalid/expired sessions
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// CountUserSessions returns number of active sessions for a user
func (s *SessionManager) CountUserSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	return s.redis.SCard(ctx, userSessionsKey).Result()
}

// ValidateSession checks if session exists and is valid
func (s *SessionManager) ValidateSession(ctx context.Context, userID uuid.UUID, sessionID string) (bool, error) {
	_, err := s.GetSession(ctx, userID, sessionID)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// EnforceMaxSessions limits number of concurrent sessions (e.g., max 5 devices)
func (s *SessionManager) EnforceMaxSessions(ctx context.Context, userID uuid.UUID, maxSessions int) error {
	sessions, err := s.ListUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	// If under limit, nothing to do
	if len(sessions) <= maxSessions {
		return nil
	}

	// Sort by last activity (oldest first)
	// Delete oldest sessions until under limit
	excessCount := len(sessions) - maxSessions
	for i := 0; i < excessCount; i++ {
		// Find oldest session
		oldestIdx := 0
		for j := 1; j < len(sessions); j++ {
			if sessions[j].LastActivity.Before(sessions[oldestIdx].LastActivity) {
				oldestIdx = j
			}
		}

		// Delete oldest session
		if err := s.DeleteSession(ctx, userID, sessions[oldestIdx].SessionID); err != nil {
			return err
		}

		// Remove from list
		sessions = append(sessions[:oldestIdx], sessions[oldestIdx+1:]...)
	}

	return nil
}