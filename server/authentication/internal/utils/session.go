package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"shared/errors"
)

// ============================================================================
// LUA SCRIPTS
// ============================================================================

const (
	// Update session activity without fetching entire session
	updateActivityLua = `
local data = redis.call('GET', KEYS[1])
if not data then return nil end
local session = cjson.decode(data)
session.last_activity = ARGV[1]
redis.call('SET', KEYS[1], cjson.encode(session), 'KEEPTTL')
return 1
`

	// Create session with max sessions limit enforcement
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

// ============================================================================
// SESSION DATA STRUCT
// ============================================================================

// SessionData represents a user session stored in Redis
type SessionData struct {
	// User identification
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`

	// Organization context (active organization)
	OrganizationID uuid.UUID `json:"organization_id"`
	OrgRole        string    `json:"org_role"` // owner, admin, member, viewer

	// Session identification
	SessionID string `json:"session_id"`

	// Device & Location info
	DeviceInfo string `json:"device_info,omitempty"` // "Chrome on Windows", "Safari on iPhone"
	IPAddress  string `json:"ip_address,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`

	// Timestamps
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// ============================================================================
// SESSION MANAGER
// ============================================================================

// Manager handles session operations in Redis
type Manager struct {
	client      *redis.Client
	maxSessions int           // Max concurrent sessions per user
	defaultTTL  time.Duration // Session TTL (default: 7 days for refresh token)
}

// NewManager creates a new session manager
func NewManager(client *redis.Client, maxSessions int, ttl time.Duration) *Manager {
	// Default: max 5 sessions per user
	if maxSessions <= 0 {
		maxSessions = 5
	}
	// Default: 7 days TTL (matches refresh token)
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	return &Manager{
		client:      client,
		maxSessions: maxSessions,
		defaultTTL:  ttl,
	}
}

// ============================================================================
// KEY GENERATORS
// ============================================================================

// sessionKey generates Redis key for a specific session
// Format: auth:session:USER_ID:SESSION_ID
func (m *Manager) sessionKey(userID uuid.UUID, sessionID string) string {
	return fmt.Sprintf("auth:session:%s:%s", userID, sessionID)
}

// userSessionsSetKey generates Redis key for user's session set
// Format: auth:user_sessions:USER_ID
func (m *Manager) userSessionsSetKey(userID uuid.UUID) string {
	return fmt.Sprintf("auth:user_sessions:%s", userID)
}

// ============================================================================
// CREATE SESSION
// ============================================================================

// CreateSession creates a new session in Redis with max sessions enforcement
func (m *Manager) CreateSession(ctx context.Context, data *SessionData) error {
	now := time.Now().UTC()
	data.CreatedAt = now
	data.LastActivity = now

	// Marshal session data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.Internal(fmt.Errorf("failed to marshal session: %w", err))
	}

	// Lua script keys and args
	keys := []string{
		m.sessionKey(data.UserID, data.SessionID),
		m.userSessionsSetKey(data.UserID),
	}
	args := []interface{}{
		jsonData,
		m.maxSessions,
		int(m.defaultTTL.Seconds()),
		data.SessionID,
	}

	// Execute Lua script
	result, err := m.client.Eval(ctx, createSessionWithLimitLua, keys, args...).Result()
	if err != nil {
		return errors.Internal(fmt.Errorf("redis lua error: %w", err))
	}

	// Parse result
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 1 {
		return errors.Internal(fmt.Errorf("unexpected redis response"))
	}

	status := int(resultSlice[0].(int64))
	
	// If max sessions exceeded, enforce limit and retry
	if status == -1 {
		if err := m.enforceMaxSessions(ctx, data.UserID); err != nil {
			return err
		}

		// Retry after cleanup
		retryResult, err := m.client.Eval(ctx, createSessionWithLimitLua, keys, args...).Result()
		if err != nil {
			return errors.Internal(fmt.Errorf("redis retry error: %w", err))
		}

		if retrySlice, ok := retryResult.([]interface{}); ok {
			if retryStatus := int(retrySlice[0].(int64)); retryStatus == -1 {
				return errors.Internal(fmt.Errorf("failed to create session after cleanup"))
			}
		}
	}

	return nil
}

// ============================================================================
// GET SESSION
// ============================================================================

// GetSession retrieves session data from Redis
func (m *Manager) GetSession(ctx context.Context, userID uuid.UUID, sessionID string) (*SessionData, error) {
	data, err := m.client.Get(ctx, m.sessionKey(userID, sessionID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.NotFound("Session expired or invalid")
		}
		return nil, errors.Internal(err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, errors.Internal(fmt.Errorf("session data corrupted: %w", err))
	}
	
	return &session, nil
}

// ============================================================================
// UPDATE ACTIVITY
// ============================================================================

// UpdateActivity updates last activity timestamp for a session
func (m *Manager) UpdateActivity(ctx context.Context, userID uuid.UUID, sessionID string) error {
	key := m.sessionKey(userID, sessionID)
	nowStr := time.Now().UTC().Format(time.RFC3339Nano)

	res, err := m.client.Eval(ctx, updateActivityLua, []string{key}, nowStr).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.NotFound("Session not found")
		}
		return errors.Internal(fmt.Errorf("redis update activity error: %w", err))
	}

	if res == nil {
		return errors.NotFound("Session expired or invalid")
	}

	return nil
}

// ============================================================================
// UPDATE ORGANIZATION CONTEXT
// ============================================================================

// UpdateOrganizationContext updates organization context when user switches orgs
func (m *Manager) UpdateOrganizationContext(
	ctx context.Context,
	userID uuid.UUID,
	sessionID string,
	newOrgID uuid.UUID,
	newOrgRole string,
) error {
	key := m.sessionKey(userID, sessionID)

	// Get current session
	data, err := m.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return errors.NotFound("Session not found")
		}
		return errors.Internal(err)
	}

	// Unmarshal, update, marshal
	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return errors.Internal(fmt.Errorf("session data corrupted: %w", err))
	}

	// Update organization context
	session.OrganizationID = newOrgID
	session.OrgRole = newOrgRole
	session.LastActivity = time.Now().UTC()

	updatedData, err := json.Marshal(session)
	if err != nil {
		return errors.Internal(err)
	}

	// Save with KEEPTTL to preserve expiration
	if err := m.client.Set(ctx, key, updatedData, redis.KeepTTL).Err(); err != nil {
		return errors.Internal(err)
	}

	return nil
}

// ============================================================================
// REVOKE SESSIONS
// ============================================================================

// RevokeSession revokes a specific session
func (m *Manager) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	pipe := m.client.Pipeline()
	pipe.Del(ctx, m.sessionKey(userID, sessionID))
	pipe.SRem(ctx, m.userSessionsSetKey(userID), sessionID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.Internal(fmt.Errorf("failed to revoke session: %w", err))
	}
	
	return nil
}

// RevokeAllSessions revokes all sessions for a user (logout all devices)
func (m *Manager) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	setKey := m.userSessionsSetKey(userID)

	// Get all session IDs
	sessionIDs, err := m.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return errors.Internal(err)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	// Delete all sessions
	pipe := m.client.Pipeline()
	for _, sid := range sessionIDs {
		pipe.Del(ctx, m.sessionKey(userID, sid))
	}
	pipe.Del(ctx, setKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return errors.Internal(fmt.Errorf("failed to revoke all sessions: %w", err))
	}
	
	return nil
}

// ============================================================================
// SESSION LISTING
// ============================================================================

// GetAllSessions retrieves all active sessions for a user
func (m *Manager) GetAllSessions(ctx context.Context, userID uuid.UUID) ([]SessionData, error) {
	setKey := m.userSessionsSetKey(userID)
	
	// Get all session IDs
	sessionIDs, err := m.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return nil, errors.Internal(err)
	}

	var sessions []SessionData
	for _, sid := range sessionIDs {
		sess, err := m.GetSession(ctx, userID, sid)
		if err == nil {
			sessions = append(sessions, *sess)
		} else {
			// Clean up invalid session
			m.client.SRem(ctx, setKey, sid)
		}
	}

	// Sort by last activity (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.After(sessions[j].LastActivity)
	})

	return sessions, nil
}

// GetSessionCount returns count of active sessions for a user
func (m *Manager) GetSessionCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := m.client.SCard(ctx, m.userSessionsSetKey(userID)).Result()
	if err != nil {
		return 0, errors.Internal(err)
	}
	
	return int(count), nil
}

// ============================================================================
// HELPERS
// ============================================================================

// enforceMaxSessions removes oldest sessions when max limit is exceeded
func (m *Manager) enforceMaxSessions(ctx context.Context, userID uuid.UUID) error {
	setKey := m.userSessionsSetKey(userID)
	
	// Get all session IDs
	sessionIDs, err := m.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return errors.Internal(err)
	}

	// Fetch all sessions
	var sessions []SessionData
	for _, sid := range sessionIDs {
		sess, err := m.GetSession(ctx, userID, sid)
		if err == nil {
			sessions = append(sessions, *sess)
		} else {
			// Clean up invalid sessions
			m.client.SRem(ctx, setKey, sid)
		}
	}

	// Sort by last activity (oldest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.Before(sessions[j].LastActivity)
	})

	// Calculate how many to delete
	currentCount := len(sessions)
	sessionsToDelete := currentCount - (m.maxSessions - 1)

	// Delete oldest sessions
	if sessionsToDelete > 0 {
		for i := 0; i < sessionsToDelete; i++ {
			_ = m.RevokeSession(ctx, userID, sessions[i].SessionID)
		}
	}

	return nil
}

// ============================================================================
// SESSION INFO HELPERS
// ============================================================================

// IsSessionValid checks if session exists and is valid
func (m *Manager) IsSessionValid(ctx context.Context, userID uuid.UUID, sessionID string) bool {
	exists, err := m.client.Exists(ctx, m.sessionKey(userID, sessionID)).Result()
	return err == nil && exists > 0
}

// GetSessionTTL returns remaining TTL for a session
func (m *Manager) GetSessionTTL(ctx context.Context, userID uuid.UUID, sessionID string) (time.Duration, error) {
	ttl, err := m.client.TTL(ctx, m.sessionKey(userID, sessionID)).Result()
	if err != nil {
		return 0, errors.Internal(err)
	}
	
	if ttl < 0 {
		return 0, errors.NotFound("Session expired or not found")
	}
	
	return ttl, nil
}

// ExtendSession extends session TTL
func (m *Manager) ExtendSession(ctx context.Context, userID uuid.UUID, sessionID string, ttl time.Duration) error {
	key := m.sessionKey(userID, sessionID)
	
	// Check if session exists
	exists, err := m.client.Exists(ctx, key).Result()
	if err != nil {
		return errors.Internal(err)
	}
	
	if exists == 0 {
		return errors.NotFound("Session not found")
	}

	// Extend TTL
	if err := m.client.Expire(ctx, key, ttl).Err(); err != nil {
		return errors.Internal(err)
	}

	return nil
}