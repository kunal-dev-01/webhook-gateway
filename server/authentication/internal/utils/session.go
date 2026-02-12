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
	updateActivityLua = `
local data = redis.call('GET', KEYS[1])
if not data then return nil end
local session = cjson.decode(data)
session.last_activity = ARGV[1]
redis.call('SET', KEYS[1], cjson.encode(session), 'KEEPTTL')
return 1
`

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
// STRUCTS
// ============================================================================

type SessionData struct {
	UserID       uuid.UUID `json:"user_id"`
	CompanyID    uuid.UUID `json:"company_id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	SessionID    string    `json:"session_id"`
	DeviceInfo   string    `json:"device_info,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

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
		ttl = 24 * time.Hour
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
	args := []interface{}{jsonData, m.maxSessions, int(m.defaultTTL.Seconds()), data.SessionID}

	result, err := m.client.Eval(ctx, createSessionWithLimitLua, keys, args...).Result()
	if err != nil {
		return errors.Internal(fmt.Errorf("redis lua error: %w", err))
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 1 {
		return errors.Internal(fmt.Errorf("unexpected redis response"))
	}

	status := int(resultSlice[0].(int64))
	if status == -1 {
		if err := m.enforceMaxSessions(ctx, data.UserID); err != nil {
			return err
		}

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
// REVOKE SESSIONS
// ============================================================================

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

	_, err = pipe.Exec(ctx)
	if err != nil {
		return errors.Internal(fmt.Errorf("failed to revoke all sessions: %w", err))
	}
	return nil
}

// ============================================================================
// HELPERS
// ============================================================================

func (m *Manager) enforceMaxSessions(ctx context.Context, userID uuid.UUID) error {
	setKey := m.userSessionsSetKey(userID)
	sessionIDs, err := m.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return errors.Internal(err)
	}

	var sessions []SessionData
	for _, sid := range sessionIDs {
		sess, err := m.GetSession(ctx, userID, sid)
		if err == nil {
			sessions = append(sessions, *sess)
		} else {
			m.client.SRem(ctx, setKey, sid)
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.Before(sessions[j].LastActivity)
	})

	currentCount := len(sessions)
	sessionsToDelete := currentCount - (m.maxSessions - 1)

	if sessionsToDelete > 0 {
		for i := 0; i < sessionsToDelete; i++ {
			_ = m.RevokeSession(ctx, userID, sessions[i].SessionID)
		}
	}

	return nil
}

func (m *Manager) GetAllSessions(ctx context.Context, userID uuid.UUID) ([]SessionData, error) {
	setKey := m.userSessionsSetKey(userID)
	sessionIDs, err := m.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return nil, errors.Internal(err)
	}

	var sessions []SessionData
	for _, sid := range sessionIDs {
		sess, err := m.GetSession(ctx, userID, sid)
		if err == nil {
			sessions = append(sessions, *sess)
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.After(sessions[j].LastActivity)
	})

	return sessions, nil
}