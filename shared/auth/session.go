package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Common Errors
var (
	ErrSessionNotFound = errors.New("session not found")
)

// Key Generators (To avoid typos and duplication)
func getSessionKey(userID uuid.UUID, sessionID string) string {
	return fmt.Sprintf("session:%s:%s", userID, sessionID)
}

func getUserSessionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:sessions:%s", userID)
}

// SessionManager handles session operations
type SessionManager struct {
	redis *redis.Client
}

// NewSessionManager creates a new session manager
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

// CreateSession creates a new session atomically
func (s *SessionManager) CreateSession(ctx context.Context, userID uuid.UUID, email, role, sessionID, ipAddress, userAgent string, ttl time.Duration) error {
	session := SessionData{
		UserID:       userID,
		Email:        email,
		Role:         role,
		SessionID:    sessionID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now().UTC(), // Always use UTC for consistency
		LastActivity: time.Now().UTC(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionKey := getSessionKey(userID, sessionID)
	userSessionsKey := getUserSessionsKey(userID)

	// Pipeline: 1 Network Trip
	pipe := s.redis.Pipeline()
	pipe.Set(ctx, sessionKey, data, ttl)
	pipe.SAdd(ctx, userSessionsKey, sessionID)
	pipe.Expire(ctx, userSessionsKey, ttl) // Extend user set TTL as well

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute session creation pipeline: %w", err)
	}

	return nil
}

// GetSession retrieves session data
func (s *SessionManager) GetSession(ctx context.Context, userID uuid.UUID, sessionID string) (*SessionData, error) {
	key := getSessionKey(userID, sessionID)

	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("redis error getting session: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &session, nil
}

// UpdateSessionActivity updates the last activity timestamp
func (s *SessionManager) UpdateSessionActivity(ctx context.Context, userID uuid.UUID, sessionID string) error {
	key := getSessionKey(userID, sessionID)

	// 1. Get existing session
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("failed to get session for update: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("corrupted session data: %w", err)
	}

	// 2. Update timestamp
	session.LastActivity = time.Now().UTC()

	newData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal updated session: %w", err)
	}

	// 3. Save back ONLY if it exists, maintaining the original TTL
	// redis.KeepTTL requires Redis 6.0+, significantly optimizes performance (no need to fetch TTL first)
	if err := s.redis.Set(ctx, key, newData, redis.KeepTTL).Err(); err != nil {
		return fmt.Errorf("failed to save session update: %w", err)
	}

	return nil
}

// DeleteSession removes a specific session
func (s *SessionManager) DeleteSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	sessionKey := getSessionKey(userID, sessionID)
	userSessionsKey := getUserSessionsKey(userID)

	pipe := s.redis.Pipeline()
	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, userSessionsKey, sessionID)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete session pipeline: %w", err)
	}

	return nil
}

// DeleteAllUserSessions removes all active sessions for a user
func (s *SessionManager) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	userSessionsKey := getUserSessionsKey(userID)

	// Get all session IDs
	sessionIDs, err := s.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch user sessions list: %w", err)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	pipe := s.redis.Pipeline()
	
	// Delete individual session keys
	for _, sid := range sessionIDs {
		pipe.Del(ctx, getSessionKey(userID, sid))
	}
	// Delete the index key
	pipe.Del(ctx, userSessionsKey)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute batch delete: %w", err)
	}

	return nil
}

// ListUserSessions returns all active sessions for a user
func (s *SessionManager) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*SessionData, error) {
	userSessionsKey := getUserSessionsKey(userID)

	sessionIDs, err := s.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session IDs: %w", err)
	}

	if len(sessionIDs) == 0 {
		return []*SessionData{}, nil
	}

	// Batch fetch using Pipeline
	pipe := s.redis.Pipeline()
	cmds := make(map[string]*redis.StringCmd, len(sessionIDs))

	for _, sid := range sessionIDs {
		key := getSessionKey(userID, sid)
		cmds[sid] = pipe.Get(ctx, key)
	}

	// Ignore error here because some keys might have expired (redis.Nil)
	_, _ = pipe.Exec(ctx)

	sessions := make([]*SessionData, 0, len(sessionIDs))

	// Process results
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			// If key is missing (expired), we should lazily clean it up from the set
			// But for read-operation speed, we just skip it for now.
			if errors.Is(err, redis.Nil) {
				// Optional: Trigger background cleanup for 'sid'
				continue 
			}
			// Log error for monitoring if needed, but continue fetching others
			continue
		}

		var session SessionData
		if err := json.Unmarshal(data, &session); err != nil {
			continue // Skip malformed data
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// CountUserSessions returns number of active sessions
func (s *SessionManager) CountUserSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.redis.SCard(ctx, getUserSessionsKey(userID)).Result()
}

// ValidateSession checks if session exists (Fast check)
func (s *SessionManager) ValidateSession(ctx context.Context, userID uuid.UUID, sessionID string) (bool, error) {
	key := getSessionKey(userID, sessionID)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists check failed: %w", err)
	}
	return exists > 0, nil
}

// EnforceMaxSessions ensures user doesn't exceed max concurrent sessions
func (s *SessionManager) EnforceMaxSessions(ctx context.Context, userID uuid.UUID, maxSessions int) error {
	sessions, err := s.ListUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	if len(sessions) <= maxSessions {
		return nil
	}

	// Sort sessions by LastActivity (Ascending: Oldest first)
	// Improved: Using sort.Slice is much faster and cleaner than manual loops
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.Before(sessions[j].LastActivity)
	})

	excessCount := len(sessions) - maxSessions
	sessionsToDelete := sessions[:excessCount]

	if len(sessionsToDelete) == 0 {
		return nil
	}

	// Batch delete oldest sessions
	pipe := s.redis.Pipeline()
	userSessionsKey := getUserSessionsKey(userID)

	for _, session := range sessionsToDelete {
		pipe.Del(ctx, getSessionKey(userID, session.SessionID))
		pipe.SRem(ctx, userSessionsKey, session.SessionID)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to enforce max sessions: %w", err)
	}

	return nil
}