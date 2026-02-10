package middleware

import (
	"slices"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"shared/auth"
)

// Context Keys (To avoid typos and use consistently)
const (
	CtxUserID    = "user_id"
	CtxUserEmail = "email"
	CtxUserRole  = "role"
	CtxSessionID = "session_id"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtManager *auth.TokenManager // Updated to TokenManager
	redis      *redis.Client
}

// NewAuthMiddleware creates new auth middleware
func NewAuthMiddleware(jwtManager *auth.TokenManager, redisClient *redis.Client) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		redis:      redisClient,
	}
}

// ============================================================================
// Core Middleware
// ============================================================================

// RequireAuth validates JWT and checks Redis for session revocation
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := m.verifyRequest(c)
		if err != nil {
			// Return specific error structure
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		// Inject user data into context
		m.setContext(c, claims)
		return c.Next()
	}
}

// OptionalAuth checks JWT if present, but doesn't block request if missing
func (m *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		// Try to verify, but ignore errors
		if claims, err := m.verifyRequest(c); err == nil {
			m.setContext(c, claims)
		}

		return c.Next()
	}
}

// RequireRole enforces RBAC (Role Based Access Control)
func (m *AuthMiddleware) RequireRole(allowedRoles ...string) fiber.Handler {
	// Optimization: Convert slice to map for O(1) lookup if roles > 3, 
	// but for small lists, loop is faster/negligible.
	return func(c *fiber.Ctx) error {
		userRole, err := GetUserRole(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "unauthorized: role not found",
			})
		}

		if slices.Contains(allowedRoles, userRole) {
				return c.Next()
			}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "forbidden: insufficient permissions",
		})
	}
}

// ============================================================================
// Internal Helpers (Optimization Logic)
// ============================================================================

// verifyRequest handles the heavy lifting: Header -> Token -> Claims -> Redis Check
func (m *AuthMiddleware) verifyRequest(c *fiber.Ctx) (*auth.Claims, error) {
	// 1. Extract Header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	// 2. Parse Token String
	tokenString, err := auth.ExtractTokenFromBearer(authHeader)
	if err != nil {
		return nil, errors.New("invalid authorization header format")
	}

	// 3. Validate JWT Signature & Expiry
	claims, err := m.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredToken) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	// 4. Redis Security Check (Is Session Active?)
	// Key format must match SessionManager: "session:USER_ID:SESSION_ID"
	sessionKey := fmt.Sprintf("session:%s:%s", claims.UserID, claims.SessionID)

	// Use Exists (Fast, no data fetch)
	exists, err := m.redis.Exists(c.Context(), sessionKey).Result()
	if err != nil {
		// Fail Closed: If Redis is down, we cannot guarantee security.
		return nil, errors.New("system error: session check failed")
	}
	
	if exists == 0 {
		return nil, errors.New("session expired or revoked")
	}

	return claims, nil
}

// setContext injects claims into Fiber locals using constants
func (m *AuthMiddleware) setContext(c *fiber.Ctx, claims *auth.Claims) {
	c.Locals(CtxUserID, claims.UserID)
	c.Locals(CtxUserEmail, claims.Email)
	c.Locals(CtxUserRole, claims.Role)
	c.Locals(CtxSessionID, claims.SessionID)
}

// ============================================================================
// Public Context Getters
// ============================================================================

// GetUserID retrieves user ID safely
func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	val := c.Locals(CtxUserID)
	if id, ok := val.(uuid.UUID); ok {
		return id, nil
	}
	return uuid.Nil, errors.New("user ID not found in context")
}

// GetUserEmail retrieves user email safely
func GetUserEmail(c *fiber.Ctx) (string, error) {
	val := c.Locals(CtxUserEmail)
	if email, ok := val.(string); ok {
		return email, nil
	}
	return "", errors.New("email not found in context")
}

// GetUserRole retrieves user role safely
func GetUserRole(c *fiber.Ctx) (string, error) {
	val := c.Locals(CtxUserRole)
	if role, ok := val.(string); ok {
		return role, nil
	}
	return "", errors.New("role not found in context")
}

// GetSessionID retrieves session ID safely
func GetSessionID(c *fiber.Ctx) (string, error) {
	val := c.Locals(CtxSessionID)
	if sid, ok := val.(string); ok {
		return sid, nil
	}
	return "", errors.New("session ID not found in context")
}

// IsAuthenticated fast check
func IsAuthenticated(c *fiber.Ctx) bool {
	_, ok := c.Locals(CtxUserID).(uuid.UUID)
	return ok
}