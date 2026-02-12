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

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrMaxSessions     = errors.New("maximum sessions reached")
)

// Lua scripts for atomic operations
const (
	// Atomic increment with expiry initialization
	incrExpiryLua = `
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return current
`

	// Atomic session activity update
	updateActivityLua = `
local data = redis.call('GET', KEYS[1])
if not data then return nil end
local session = cjson.decode(data)
session.last_activity = ARGV[1]
redis.call('SET', KEYS[1], cjson.encode(session), 'KEEPTTL')
return 1
`

	// Atomic session creation with max sessions enforcement
	createSessionWithLimitLua = `
local current_sessions = redis.call('SCARD', KEYS[2])
if current_sessions >= tonumber(ARGV[2]) then
    return {-1, current_sessions}
end

redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[3])
redis.call('SADD', KEYS[2], ARGV[4])
redis.call('EXPIRE', KEYS[2], ARGV[3])
return {1, current_sessions + 1}
`
)

// SessionManager handles session operations
type SessionManager struct {
	redis      *redis.Client
	maxSessions int
	defaultTTL  time.Duration
}

// SessionConfig for manager configuration
type SessionConfig struct {
	MaxSessions int
	DefaultTTL  time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(redisClient *redis.Client, cfg SessionConfig) *SessionManager {
	if cfg.MaxSessions <= 0 {
		cfg.MaxSessions = 5 // Default
	}
	if cfg.DefaultTTL <= 0 {
		cfg.DefaultTTL = 24 * time.Hour
	}

	return &SessionManager{
		redis:       redisClient,
		maxSessions: cfg.MaxSessions,
		defaultTTL:  cfg.DefaultTTL,
	}
}

// SessionData represents session information
// UPDATED: Added CompanyID for multi-tenant context
type SessionData struct {
	UserID       uuid.UUID `json:"user_id"`
	CompanyID    uuid.UUID `json:"company_id"`    // NEW: Multi-tenant isolation
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	SessionID    string    `json:"session_id"`
	DeviceInfo   string    `json:"device_info,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// Key generators
func (s *SessionManager) sessionKey(userID uuid.UUID, sessionID string) string {
	return fmt.Sprintf("session:%s:%s", userID, sessionID)
}

func (s *SessionManager) userSessionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:sessions:%s", userID)
}

// ============================================================================
// Core Operations
// ============================================================================

// CreateSession creates a new session with max limit enforcement
// UPDATED: Now requires companyID parameter
func (s *SessionManager) CreateSession(ctx context.Context, userID, companyID uuid.UUID, email, role, sessionID, ipAddress, userAgent string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = s.defaultTTL
	}

	session := SessionData{
		UserID:       userID,
		CompanyID:    companyID, // NEW: Store company context
		Email:        email,
		Role:         role,
		SessionID:    sessionID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now().UTC(),
		LastActivity: time.Now().UTC(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	// Try atomic creation with limit
	result, err := s.redis.Eval(ctx, createSessionWithLimitLua, []string{
		s.sessionKey(userID, sessionID),
		s.userSessionsKey(userID),
	}, data, s.maxSessions, int(ttl.Seconds()), sessionID).Result()

	if err != nil {
		return fmt.Errorf("create session lua: %w", err)
	}

	resultSlice, ok := result.([]any)
	if !ok || len(resultSlice) != 2 {
		return fmt.Errorf("unexpected lua result")
	}

	status, _ := resultSlice[0].(int64)
	if status == -1 {
		// Max sessions reached - enforce cleanup and retry once
		if err := s.EnforceMaxSessions(ctx, userID, s.maxSessions-1); err != nil {
			return fmt.Errorf("enforce max sessions: %w", err)
		}
		
		// Retry once after cleanup
		return s.createSessionSimple(ctx, userID, sessionID, data, ttl)
	}

	return nil
}

// createSessionSimple creates session without limit check (used for retry)
func (s *SessionManager) createSessionSimple(ctx context.Context, userID uuid.UUID, sessionID string, data []byte, ttl time.Duration) error {
	pipe := s.redis.Pipeline()
	pipe.Set(ctx, s.sessionKey(userID, sessionID), data, ttl)
	pipe.SAdd(ctx, s.userSessionsKey(userID), sessionID)
	pipe.Expire(ctx, s.userSessionsKey(userID), ttl)

	_, err := pipe.Exec(ctx)
	return err
}

// GetSession retrieves session data
func (s *SessionManager) GetSession(ctx context.Context, userID uuid.UUID, sessionID string) (*SessionData, error) {
	key := s.sessionKey(userID, sessionID)

	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &session, nil
}

// UpdateSessionActivity atomically updates last activity
func (s *SessionManager) UpdateSessionActivity(ctx context.Context, userID uuid.UUID, sessionID string) error {
	key := s.sessionKey(userID, sessionID)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	result, err := s.redis.Eval(ctx, updateActivityLua, []string{key}, now).Result()
	if err != nil {
		return fmt.Errorf("update activity lua: %w", err)
	}

	if result == nil {
		return ErrSessionNotFound
	}

	return nil
}

// DeleteSession removes a specific session
func (s *SessionManager) DeleteSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	pipe := s.redis.Pipeline()
	pipe.Del(ctx, s.sessionKey(userID, sessionID))
	pipe.SRem(ctx, s.userSessionsKey(userID), sessionID)

	_, err := s.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, s.sessionKey(userID, sessionID))
		pipe.SRem(ctx, s.userSessionsKey(userID), sessionID)
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

// DeleteAllUserSessions removes all sessions for a user
func (s *SessionManager) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	userKey := s.userSessionsKey(userID)
	
	sessionIDs, err := s.redis.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("fetch sessions: %w", err)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	pipe := s.redis.Pipeline()
	for _, sid := range sessionIDs {
		pipe.Del(ctx, s.sessionKey(userID, sid))
	}
	pipe.Del(ctx, userKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete all sessions: %w", err)
	}

	return nil
}

// ListUserSessions returns all active sessions sorted by activity
func (s *SessionManager) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*SessionData, error) {
	userKey := s.userSessionsKey(userID)

	sessionIDs, err := s.redis.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("fetch session IDs: %w", err)
	}

	if len(sessionIDs) == 0 {
		return []*SessionData{}, nil
	}

	// Batch fetch
	pipe := s.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(sessionIDs))
	
	for i, sid := range sessionIDs {
		cmds[i] = pipe.Get(ctx, s.sessionKey(userID, sid))
	}

	_, _ = pipe.Exec(ctx) // Ignore errors, check individually

	sessions := make([]*SessionData, 0, len(sessionIDs))
	expiredSessions := make([]string, 0)

	for i, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				expiredSessions = append(expiredSessions, sessionIDs[i])
			}
			continue
		}

		var session SessionData
		if err := json.Unmarshal(data, &session); err != nil {
			expiredSessions = append(expiredSessions, sessionIDs[i])
			continue
		}
		sessions = append(sessions, &session)
	}

	// Cleanup expired entries asynchronously
	if len(expiredSessions) > 0 {
		go s.cleanupExpiredSessions(context.Background(), userID, expiredSessions)
	}

	// Sort: LastActivity DESC (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.After(sessions[j].LastActivity)
	})

	return sessions, nil
}

// CountUserSessions returns active session count
func (s *SessionManager) CountUserSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.redis.SCard(ctx, s.userSessionsKey(userID)).Result()
}

// ValidateSession checks if session exists
func (s *SessionManager) ValidateSession(ctx context.Context, userID uuid.UUID, sessionID string) (bool, error) {
	n, err := s.redis.Exists(ctx, s.sessionKey(userID, sessionID)).Result()
	if err != nil {
		return false, fmt.Errorf("check session exists: %w", err)
	}
	return n > 0, nil
}

// ============================================================================
// Max Sessions Enforcement
// ============================================================================

// EnforceMaxSessions keeps only N most recent sessions
func (s *SessionManager) EnforceMaxSessions(ctx context.Context, userID uuid.UUID, keep int) error {
	if keep <= 0 {
		keep = s.maxSessions
	}

	sessions, err := s.ListUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	if len(sessions) <= keep {
		return nil
	}

	// Sort by LastActivity ASC (oldest first) for deletion
	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].LastActivity.Equal(sessions[j].LastActivity) {
			return sessions[i].CreatedAt.Before(sessions[j].CreatedAt)
		}
		return sessions[i].LastActivity.Before(sessions[j].LastActivity)
	})

	toDelete := sessions[keep:]
	if len(toDelete) == 0 {
		return nil
	}

	pipe := s.redis.Pipeline()
	userKey := s.userSessionsKey(userID)

	for _, sess := range toDelete {
		pipe.Del(ctx, s.sessionKey(userID, sess.SessionID))
		pipe.SRem(ctx, userKey, sess.SessionID)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("enforce max sessions: %w", err)
	}

	return nil
}

// ============================================================================
// Cleanup
// ============================================================================

func (s *SessionManager) cleanupExpiredSessions(ctx context.Context, userID uuid.UUID, sessionIDs []string) {
	pipe := s.redis.Pipeline()
	userKey := s.userSessionsKey(userID)
	
	for _, sid := range sessionIDs {
		pipe.SRem(ctx, userKey, sid)
	}
	
	_, _ = pipe.Exec(ctx) // Best effort cleanup
}