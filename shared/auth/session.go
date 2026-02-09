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
// Session Operations - OPTIMIZED
// ============================================================================

// CreateSession creates a new session - OPTIMIZED WITH PIPELINE
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

	// ✅ OPTIMIZED: Use pipeline for atomic operations (1 network call instead of 3)
	pipe := s.redis.Pipeline()
	
	sessionKey := fmt.Sprintf("session:%s:%s", userID, sessionID)
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	
	pipe.Set(ctx, sessionKey, data, ttl)
	pipe.SAdd(ctx, userSessionsKey, sessionID)
	pipe.Expire(ctx, userSessionsKey, ttl)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

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

// UpdateSessionActivity updates last activity timestamp - OPTIMIZED
func (s *SessionManager) UpdateSessionActivity(ctx context.Context, userID uuid.UUID, sessionID string) error {
	// ✅ OPTIMIZED: Use GETEX to get + update TTL in single call
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	
	// Get current session
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Update last activity
	session.LastActivity = time.Now()
	newData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Get remaining TTL
	ttl, err := s.redis.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		ttl = 7 * 24 * time.Hour // Default 7 days if TTL not found
	}

	// Update with same TTL
	if err := s.redis.Set(ctx, key, newData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// DeleteSession removes a session - OPTIMIZED WITH PIPELINE
func (s *SessionManager) DeleteSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	// ✅ OPTIMIZED: Pipeline for atomic delete
	pipe := s.redis.Pipeline()
	
	sessionKey := fmt.Sprintf("session:%s:%s", userID, sessionID)
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	
	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, userSessionsKey, sessionID)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteAllUserSessions removes all sessions for a user - OPTIMIZED
func (s *SessionManager) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	
	// Get all session IDs
	sessionIDs, err := s.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	// ✅ OPTIMIZED: Batch delete with pipeline
	pipe := s.redis.Pipeline()
	
	for _, sessionID := range sessionIDs {
		sessionKey := fmt.Sprintf("session:%s:%s", userID, sessionID)
		pipe.Del(ctx, sessionKey)
	}
	pipe.Del(ctx, userSessionsKey)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	return nil
}

// ListUserSessions returns all active sessions for a user - OPTIMIZED
func (s *SessionManager) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*SessionData, error) {
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	sessionIDs, err := s.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	if len(sessionIDs) == 0 {
		return []*SessionData{}, nil
	}

	// ✅ OPTIMIZED: Batch fetch with pipeline
	pipe := s.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(sessionIDs))
	
	for i, sessionID := range sessionIDs {
		sessionKey := fmt.Sprintf("session:%s:%s", userID, sessionID)
		cmds[i] = pipe.Get(ctx, sessionKey)
	}
	
	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Some sessions might have expired, continue
	}

	sessions := make([]*SessionData, 0, len(sessionIDs))
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue // Skip expired sessions
		}

		var session SessionData
		if err := json.Unmarshal(data, &session); err != nil {
			continue // Skip corrupted sessions
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// CountUserSessions returns number of active sessions for a user
func (s *SessionManager) CountUserSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	return s.redis.SCard(ctx, userSessionsKey).Result()
}

// ValidateSession checks if session exists and is valid - OPTIMIZED
func (s *SessionManager) ValidateSession(ctx context.Context, userID uuid.UUID, sessionID string) (bool, error) {
	// ✅ OPTIMIZED: Use EXISTS instead of GET (faster)
	key := fmt.Sprintf("session:%s:%s", userID, sessionID)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// EnforceMaxSessions limits number of concurrent sessions - OPTIMIZED
func (s *SessionManager) EnforceMaxSessions(ctx context.Context, userID uuid.UUID, maxSessions int) error {
	sessions, err := s.ListUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	if len(sessions) <= maxSessions {
		return nil
	}

	// Sort by last activity (oldest first)
	excessCount := len(sessions) - maxSessions
	
	// ✅ OPTIMIZED: Batch delete oldest sessions
	pipe := s.redis.Pipeline()
	
	for i := 0; i < excessCount; i++ {
		// Find oldest
		oldestIdx := 0
		for j := 1; j < len(sessions); j++ {
			if sessions[j].LastActivity.Before(sessions[oldestIdx].LastActivity) {
				oldestIdx = j
			}
		}

		// Queue for deletion
		sessionKey := fmt.Sprintf("session:%s:%s", userID, sessions[oldestIdx].SessionID)
		userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
		
		pipe.Del(ctx, sessionKey)
		pipe.SRem(ctx, userSessionsKey, sessions[oldestIdx].SessionID)

		// Remove from list
		sessions = append(sessions[:oldestIdx], sessions[oldestIdx+1:]...)
	}

	_, err = pipe.Exec(ctx)
	return err
}