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
	// Update session activity (Optimized: No fetch to Go, strictly Redis side)
	updateActivityLua = `
local data = redis.call('GET', KEYS[1])
if not data then return nil end
local session = cjson.decode(data)
session.last_activity = ARGV[1]
redis.call('SET', KEYS[1], cjson.encode(session), 'KEEPTTL')
return 1
`

	// Create session with limit check (Atomic check)
	createSessionWithLimitLua = `
local current_sessions = redis.call('SCARD', KEYS[2])
if current_sessions >= tonumber(ARGV[2]) then
    return -1
end

redis.call('SET', KEYS[1], ARGV[1], 'EX', ARGV[3])
redis.call('SADD', KEYS[2], ARGV[4])
redis.call('EXPIRE', KEYS[2], ARGV[3])
return 1
`
)

// ============================================================================
// SESSION DATA STRUCT
// ============================================================================

type SessionData struct {
	UserID         uuid.UUID `json:"user_id"`
	Email          string    `json:"email"`
	OrganizationID uuid.UUID `json:"organization_id"`
	OrgRole        string    `json:"org_role"`
	SessionID      string    `json:"session_id"`
	DeviceInfo     string    `json:"device_info,omitempty"`
	IPAddress      string    `json:"ip_address,omitempty"`
	UserAgent      string    `json:"user_agent,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivity   time.Time `json:"last_activity"`
}

// ============================================================================
// SESSION MANAGER
// ============================================================================

type Manager struct {
	client      *redis.Client
	maxSessions int
	defaultTTL  time.Duration
}

func NewManager(client *redis.Client, maxSessions int, ttl time.Duration) *Manager {
	if maxSessions <= 0 {
		maxSessions = 5
	}
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

func (m *Manager) sessionKey(userID uuid.UUID, sessionID string) string {
	return fmt.Sprintf("auth:session:%s:%s", userID, sessionID)
}

func (m *Manager) userSessionsSetKey(userID uuid.UUID) string {
	return fmt.Sprintf("auth:user_sessions:%s", userID)
}

// ============================================================================
// CREATE SESSION
// ============================================================================

func (m *Manager) CreateSession(ctx context.Context, data *SessionData) error {
	now := time.Now().UTC()
	data.CreatedAt = now
	data.LastActivity = now

	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.Internal(fmt.Errorf("failed to marshal session: %w", err))
	}

	keys := []string{
		m.sessionKey(data.UserID, data.SessionID),
		m.userSessionsSetKey(data.UserID),
	}
	args := []any{
		jsonData,
		m.maxSessions,
		int(m.defaultTTL.Seconds()),
		data.SessionID,
	}

	// Try to create session atomically
	result, err := m.client.Eval(ctx, createSessionWithLimitLua, keys, args...).Int()
	if err != nil && err != redis.Nil {
		return errors.Internal(fmt.Errorf("redis lua error: %w", err))
	}

	// If result is -1, Limit Exceeded -> Cleanup and Retry
	if result == -1 {
		if err := m.enforceMaxSessions(ctx, data.UserID); err != nil {
			return err
		}

		// Retry creation
		retryResult, err := m.client.Eval(ctx, createSessionWithLimitLua, keys, args...).Int()
		if err != nil {
			return errors.Internal(fmt.Errorf("redis retry error: %w", err))
		}

		if retryResult == -1 {
			return errors.Internal(fmt.Errorf("failed to create session after cleanup"))
		}
	}

	return nil
}

// ============================================================================
// GET SESSION
// ============================================================================

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

func (m *Manager) UpdateActivity(ctx context.Context, userID uuid.UUID, sessionID string) error {
	key := m.sessionKey(userID, sessionID)
	// Use RFC3339Nano ensuring Lua can parse it or Go can Unmarshal it back correctly
	nowStr := time.Now().UTC().Format(time.RFC3339Nano)

	err := m.client.Eval(ctx, updateActivityLua, []string{key}, nowStr).Err()
	if err != nil {
		if err == redis.Nil {
			return errors.NotFound("Session not found")
		}
		return errors.Internal(fmt.Errorf("redis update activity error: %w", err))
	}
	return nil
}

// ============================================================================
// UPDATE ORGANIZATION CONTEXT
// ============================================================================

func (m *Manager) UpdateOrganizationContext(
	ctx context.Context,
	userID uuid.UUID,
	sessionID string,
	newOrgID uuid.UUID,
	newOrgRole string,
) error {
	key := m.sessionKey(userID, sessionID)

	// Optimistic locking using WATCH could be used here, but for sessions GET/SET is usually fine
	data, err := m.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return errors.NotFound("Session not found")
		}
		return errors.Internal(err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return errors.Internal(fmt.Errorf("session data corrupted: %w", err))
	}

	session.OrganizationID = newOrgID
	session.OrgRole = newOrgRole
	session.LastActivity = time.Now().UTC()

	updatedData, err := json.Marshal(session)
	if err != nil {
		return errors.Internal(err)
	}

	if err := m.client.Set(ctx, key, updatedData, redis.KeepTTL).Err(); err != nil {
		return errors.Internal(err)
	}

	return nil
}

// ============================================================================
// REVOKE SESSIONS
// ============================================================================

func (m *Manager) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	pipe := m.client.Pipeline()
	pipe.Del(ctx, m.sessionKey(userID, sessionID))
	pipe.SRem(ctx, m.userSessionsSetKey(userID), sessionID)

	if _, err := pipe.Exec(ctx); err != nil {
		return errors.Internal(fmt.Errorf("failed to revoke session: %w", err))
	}
	return nil
}

func (m *Manager) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	setKey := m.userSessionsSetKey(userID)

	sessionIDs, err := m.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return errors.Internal(err)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	pipe := m.client.Pipeline()
	for _, sid := range sessionIDs {
		pipe.Del(ctx, m.sessionKey(userID, sid))
	}
	pipe.Del(ctx, setKey)

	if _, err := pipe.Exec(ctx); err != nil {
		return errors.Internal(fmt.Errorf("failed to revoke all sessions: %w", err))
	}
	return nil
}

// ============================================================================
// SESSION LISTING (OPTIMIZED WITH MGET)
// ============================================================================

func (m *Manager) GetAllSessions(ctx context.Context, userID uuid.UUID) ([]SessionData, error) {
	setKey := m.userSessionsSetKey(userID)

	// 1. Get all session IDs
	sessionIDs, err := m.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return nil, errors.Internal(err)
	}

	if len(sessionIDs) == 0 {
		return []SessionData{}, nil
	}

	// 2. Build keys for MGET
	keys := make([]string, len(sessionIDs))
	for i, sid := range sessionIDs {
		keys[i] = m.sessionKey(userID, sid)
	}

	// 3. Fetch all sessions in ONE call
	results, err := m.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, errors.Internal(err)
	}

	// 4. Parse results
	var sessions []SessionData
	// Clean up invalid sessions found during fetch
	var invalidSIDs []string

	for i, result := range results {
		if result == nil {
			// Session key missing, mark for cleanup
			invalidSIDs = append(invalidSIDs, sessionIDs[i])
			continue
		}

		strData, ok := result.(string)
		if !ok {
			continue
		}

		var sess SessionData
		if err := json.Unmarshal([]byte(strData), &sess); err == nil {
			sessions = append(sessions, sess)
		} else {
			invalidSIDs = append(invalidSIDs, sessionIDs[i])
		}
	}

	// 5. Async cleanup for invalid sessions (don't block response)
	if len(invalidSIDs) > 0 {
		go m.client.SRem(context.Background(), setKey, invalidSIDs)
	}

	// 6. Sort
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.After(sessions[j].LastActivity)
	})

	return sessions, nil
}

func (m *Manager) GetSessionCount(ctx context.Context, userID uuid.UUID) (int, error) {
	count, err := m.client.SCard(ctx, m.userSessionsSetKey(userID)).Result()
	if err != nil {
		return 0, errors.Internal(err)
	}
	return int(count), nil
}

// ============================================================================
// HELPERS (OPTIMIZED)
// ============================================================================

func (m *Manager) enforceMaxSessions(ctx context.Context, userID uuid.UUID) error {
	// Re-using GetAllSessions logic which is now optimized with MGET
	sessions, err := m.GetAllSessions(ctx, userID)
	if err != nil {
		return err
	}

	// Calculate excess
	currentCount := len(sessions)
	sessionsToDelete := currentCount - (m.maxSessions - 1)

	if sessionsToDelete <= 0 {
		return nil
	}

	// Sort oldest first for deletion (Reverse of GetAllSessions sort)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.Before(sessions[j].LastActivity)
	})

	// Delete oldest sessions using Pipeline
	pipe := m.client.Pipeline()
	setKey := m.userSessionsSetKey(userID)

	for i := range sessionsToDelete {
		sid := sessions[i].SessionID
		pipe.Del(ctx, m.sessionKey(userID, sid))
		pipe.SRem(ctx, setKey, sid)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return errors.Internal(fmt.Errorf("failed to enforce max sessions: %w", err))
	}

	return nil
}

// ============================================================================
// SESSION INFO HELPERS
// ============================================================================

func (m *Manager) IsSessionValid(ctx context.Context, userID uuid.UUID, sessionID string) bool {
	exists, err := m.client.Exists(ctx, m.sessionKey(userID, sessionID)).Result()
	return err == nil && exists > 0
}

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

func (m *Manager) ExtendSession(ctx context.Context, userID uuid.UUID, sessionID string, ttl time.Duration) error {
	key := m.sessionKey(userID, sessionID)
	// Using EXPIRE directly (returns 0 if key doesn't exist)
	updated, err := m.client.Expire(ctx, key, ttl).Result()
	if err != nil {
		return errors.Internal(err)
	}
	if !updated {
		return errors.NotFound("Session not found")
	}
	return nil
}