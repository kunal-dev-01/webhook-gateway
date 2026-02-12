package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"
	stdErrors "errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	
	"shared/errors" // Tera Shared Error Package
)

// ───────────────── Lua Scripts ─────────────────

const (
	// Atomic session activity update (Sabse Important: Race condition avoid karta hai)
	updateActivityLua = `
local data = redis.call('GET', KEYS[1])
if not data then return nil end
local session = cjson.decode(data)
session.last_activity = ARGV[1]
redis.call('SET', KEYS[1], cjson.encode(session), 'KEEPTTL')
return 1
`

	// Atomic session creation with max sessions enforcement
	// Returns: {status, current_count}
	// Status: 1 (Success), -1 (Limit Reached)
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

// ───────────────── Structs ─────────────────

// SessionData represents the session payload
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
	// Default Max Sessions = 5 (Agar 0 pass kiya toh)
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

// ───────────────── Key Generators ─────────────────

func (s *Manager) sessionKey(userID uuid.UUID, sessionID string) string {
	return fmt.Sprintf("auth:session:%s:%s", userID, sessionID)
}

func (s *Manager) userSessionsSetKey(userID uuid.UUID) string {
	return fmt.Sprintf("auth:user_sessions:%s", userID)
}

// ───────────────── Core Methods ─────────────────

// CreateSession creates a new session securely using Lua
func (s *Manager) CreateSession(ctx context.Context, data *SessionData) error {
	data.CreatedAt = time.Now().UTC()
	data.LastActivity = time.Now().UTC()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.Internal(fmt.Errorf("failed to marshal session: %w", err))
	}

	ttlSeconds := int(s.defaultTTL.Seconds())
	keys := []string{s.sessionKey(data.UserID, data.SessionID), s.userSessionsSetKey(data.UserID)}
	args := []interface{}{jsonData, s.maxSessions, ttlSeconds, data.SessionID}

	// 1. Try to Create Session Atomicallly
	// Return type []interface{} hoga kyunki Lua table return kar raha hai
	result, err := s.client.Eval(ctx, createSessionWithLimitLua, keys, args...).Result()
	if err != nil {
		return errors.Internal(fmt.Errorf("redis lua error: %w", err))
	}

	// 2. Parse Lua Result
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 1 {
		return errors.Internal(fmt.Errorf("unexpected redis lua response"))
	}

	status := int(resultSlice[0].(int64))

	// 3. Handle Limit Reached (-1)
	if status == -1 {
		// Cleanup oldest session
		if err := s.enforceMaxSessions(ctx, data.UserID); err != nil {
			return err
		}
		
		// Retry Creation ONE time
		// Retry mein hum same Lua script use kar sakte hain, ab space honi chahiye
		retryResult, err := s.client.Eval(ctx, createSessionWithLimitLua, keys, args...).Result()
		if err != nil {
			return errors.Internal(fmt.Errorf("redis retry error: %w", err))
		}
		
		// Check retry result
		if retrySlice, ok := retryResult.([]interface{}); ok {
			if retryStatus := int(retrySlice[0].(int64)); retryStatus == -1 {
				return errors.Internal(fmt.Errorf("failed to create session after cleanup"))
			}
		}
	}

	return nil
}

// UpdateActivity updates the LastActivity timestamp ATOMICALLY
func (s *Manager) UpdateActivity(ctx context.Context, userID uuid.UUID, sessionID string) error {
	key := s.sessionKey(userID, sessionID)
	// Go ki time format RFC3339 standard JSON ke saath compatible hai
	nowStr := time.Now().UTC().Format(time.RFC3339Nano)

	// Direct Lua Call - No Get/Set roundtrip needed from Go
	res, err := s.client.Eval(ctx, updateActivityLua, []string{key}, nowStr).Result()
	
	if err != nil {
		if err == redis.Nil {
			return errors.NotFound("Session not found")
		}
		return errors.Internal(fmt.Errorf("redis update activity error: %w", err))
	}

	// Agar Lua ne nil return kiya (matlab key nahi mili)
	if res == nil {
		return errors.NotFound("Session expired or invalid")
	}

	return nil
}

// GetSession retrieves a session (Standard Get)
func (s *Manager) GetSession(ctx context.Context, userID uuid.UUID, sessionID string) (*SessionData, error) {
	data, err := s.client.Get(ctx, s.sessionKey(userID, sessionID)).Bytes()
	if err != nil {
		if stdErrors.Is(err, redis.Nil) {
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

// RevokeSession removes a specific session
func (s *Manager) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	pipe := s.client.Pipeline()
	pipe.Del(ctx, s.sessionKey(userID, sessionID))
	pipe.SRem(ctx, s.userSessionsSetKey(userID), sessionID)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.Internal(fmt.Errorf("failed to revoke session: %w", err))
	}
	return nil
}

// RevokeAllSessions removes all sessions for a user
func (s *Manager) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	setKey := s.userSessionsSetKey(userID)
	
	// Get all session IDs
	sessionIDs, err := s.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return errors.Internal(err)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	for _, sid := range sessionIDs {
		pipe.Del(ctx, s.sessionKey(userID, sid))
	}
	pipe.Del(ctx, setKey)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return errors.Internal(fmt.Errorf("failed to revoke all sessions: %w", err))
	}
	return nil
}

// ───────────────── Internal Helpers ─────────────────

// enforceMaxSessions removes the oldest session (LRU)
func (s *Manager) enforceMaxSessions(ctx context.Context, userID uuid.UUID) error {
	setKey := s.userSessionsSetKey(userID)
	sessionIDs, err := s.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return errors.Internal(err)
	}

	// Agar current sessions Max se kam hain, toh kuch karne ki zarurat nahi
	// Lekin agar hum yahan aaye hain, matlab limit hit hui hai.
	
	var sessions []SessionData
	for _, sid := range sessionIDs {
		sess, err := s.GetSession(ctx, userID, sid)
		if err == nil {
			sessions = append(sessions, *sess)
		} else {
			// Orphaned ID in set
			s.client.SRem(ctx, setKey, sid)
		}
	}

	// Sort by LastActivity (Oldest First)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.Before(sessions[j].LastActivity)
	})

	// Delete oldest until we have space (Target: Max - 1)
	// Example: Max=5, Current=5. Delete 1 -> Count=4. New one adds -> Count=5.
	// Agar race condition ki wajah se 6 hogaye, toh loop handle kar lega.
	currentCount := len(sessions)
	sessionsToDelete := currentCount - (s.maxSessions - 1)

	if sessionsToDelete > 0 {
		for i := range sessionsToDelete {
			_ = s.RevokeSession(ctx, userID, sessions[i].SessionID)
		}
	}

	return nil
}