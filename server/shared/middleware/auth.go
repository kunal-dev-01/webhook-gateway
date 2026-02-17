package middleware

import (
	"errors"
	"fmt"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"shared/auth"
)

// ============================================================================
// Context Keys (Type-Safe)
// ============================================================================

type ContextKey string

const (
	CtxUserID         ContextKey = "user_id"
	CtxOrganizationID ContextKey = "organization_id"
	CtxUserEmail      ContextKey = "email"
	CtxOrgRole        ContextKey = "org_role"
	CtxSessionID      ContextKey = "session_id"
)

// ============================================================================
// Auth Middleware
// ============================================================================

type AuthMiddleware struct {
	jwtManager *auth.TokenManager
	redis      *redis.Client
}

func NewAuthMiddleware(jwtManager *auth.TokenManager, redisClient *redis.Client) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		redis:      redisClient,
	}
}

// ============================================================================
// Session Key Helper
// ============================================================================

// sessionKey must match SessionManager.sessionKey() format exactly:
// Format: auth:session:USER_ID:SESSION_ID
func sessionKey(userID uuid.UUID, sessionID string) string {
	return fmt.Sprintf("auth:session:%s:%s", userID, sessionID)
}

// ============================================================================
// Core Middleware Handlers
// ============================================================================

func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := m.verifyRequest(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}
		m.setContext(c, claims)
		return c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}
		if claims, err := m.verifyRequest(c); err == nil {
			m.setContext(c, claims)
		}
		return c.Next()
	}
}

// ============================================================================
// Role-Based Access Control (RBAC)
// ============================================================================

func (m *AuthMiddleware) RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgRole, err := GetOrgRole(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "unauthorized: role not found",
			})
		}
		if slices.Contains(allowedRoles, orgRole) {
			return c.Next()
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("forbidden: requires role %v", allowedRoles),
		})
	}
}

func (m *AuthMiddleware) RequireOwner() fiber.Handler  { return m.RequireRole("owner") }
func (m *AuthMiddleware) RequireAdmin() fiber.Handler  { return m.RequireRole("owner", "admin") }
func (m *AuthMiddleware) RequireMember() fiber.Handler { return m.RequireRole("owner", "admin", "member") }

func (m *AuthMiddleware) RequireMinRole(minRole string) fiber.Handler {
	roleHierarchy := map[string]int{
		"viewer": 1,
		"member": 2,
		"admin":  3,
		"owner":  4,
	}
	return func(c *fiber.Ctx) error {
		orgRole, err := GetOrgRole(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "unauthorized: role not found",
			})
		}
		minLevel, minExists := roleHierarchy[minRole]
		userLevel, userExists := roleHierarchy[orgRole]
		if !minExists || !userExists {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "invalid role configuration",
			})
		}
		if userLevel >= minLevel {
			return c.Next()
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("forbidden: requires minimum role %s", minRole),
		})
	}
}

// ============================================================================
// Organization Context Validation
// ============================================================================

func (m *AuthMiddleware) RequireOrganization() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claimsOrgID, err := GetOrganizationID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "organization context not found",
			})
		}
		orgIDParam := c.Params("org_id")
		if orgIDParam == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "organization ID required in URL",
			})
		}
		paramOrgID, err := uuid.Parse(orgIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "invalid organization ID format",
			})
		}
		if claimsOrgID != paramOrgID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "access denied: organization mismatch",
			})
		}
		return c.Next()
	}
}

// ============================================================================
// Permission Helpers
// ============================================================================

func (m *AuthMiddleware) RequirePermission(checkFunc func(*auth.Claims) bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx, err := GetAuthContext(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "unauthorized",
			})
		}
		claims := &auth.Claims{
			UserID:         authCtx.UserID,
			OrganizationID: authCtx.OrganizationID,
			Email:          authCtx.Email,
			OrgRole:        authCtx.OrgRole,
		}
		if !checkFunc(claims) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "forbidden: insufficient permissions",
			})
		}
		return c.Next()
	}
}

func (m *AuthMiddleware) RequireCanManageBilling() fiber.Handler {
	return m.RequirePermission(func(c *auth.Claims) bool { return c.CanManageBilling() })
}

func (m *AuthMiddleware) RequireCanManageMembers() fiber.Handler {
	return m.RequirePermission(func(c *auth.Claims) bool { return c.CanManageMembers() })
}

func (m *AuthMiddleware) RequireCanInviteMembers() fiber.Handler {
	return m.RequirePermission(func(c *auth.Claims) bool { return c.CanInviteMembers() })
}

// ============================================================================
// Internal: Core Verification Flow
// ============================================================================

func (m *AuthMiddleware) verifyRequest(c *fiber.Ctx) (*auth.Claims, error) {
	// 1. Extract Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	// 2. Parse "Bearer <token>"
	tokenString, err := auth.ExtractTokenFromBearer(authHeader)
	if err != nil {
		return nil, errors.New("invalid authorization header format")
	}

	// 3. Validate JWT
	claims, err := m.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrExpiredToken):
			return nil, errors.New("token expired")
		case errors.Is(err, auth.ErrMissingOrgID):
			return nil, errors.New("invalid token: missing organization")
		default:
			return nil, errors.New("invalid token")
		}
	}

	// 4. Redis session check
	// Key format matches SessionManager.sessionKey(): auth:session:USER_ID:SESSION_ID
	key := sessionKey(claims.UserID, claims.SessionID)

	exists, err := m.redis.Exists(c.Context(), key).Result()
	if err != nil {
		// Fail closed: Redis down = deny access
		return nil, errors.New("system error: session check failed")
	}
	if exists == 0 {
		return nil, errors.New("session expired or revoked")
	}

	return claims, nil
}

// ============================================================================
// Context Setters / Getters
// ============================================================================

func (m *AuthMiddleware) setContext(c *fiber.Ctx, claims *auth.Claims) {
	c.Locals(CtxUserID, claims.UserID)
	c.Locals(CtxOrganizationID, claims.OrganizationID)
	c.Locals(CtxUserEmail, claims.Email)
	c.Locals(CtxOrgRole, claims.OrgRole)
	c.Locals(CtxSessionID, claims.SessionID)
}

func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	if id, ok := c.Locals(CtxUserID).(uuid.UUID); ok {
		return id, nil
	}
	return uuid.Nil, errors.New("user ID not found in context")
}

func GetOrganizationID(c *fiber.Ctx) (uuid.UUID, error) {
	if id, ok := c.Locals(CtxOrganizationID).(uuid.UUID); ok {
		return id, nil
	}
	return uuid.Nil, errors.New("organization ID not found in context")
}

func GetUserEmail(c *fiber.Ctx) (string, error) {
	if email, ok := c.Locals(CtxUserEmail).(string); ok {
		return email, nil
	}
	return "", errors.New("email not found in context")
}

func GetOrgRole(c *fiber.Ctx) (string, error) {
	if role, ok := c.Locals(CtxOrgRole).(string); ok {
		return role, nil
	}
	return "", errors.New("organization role not found in context")
}

func GetSessionID(c *fiber.Ctx) (string, error) {
	if sid, ok := c.Locals(CtxSessionID).(string); ok {
		return sid, nil
	}
	return "", errors.New("session ID not found in context")
}

func IsAuthenticated(c *fiber.Ctx) bool {
	_, ok := c.Locals(CtxUserID).(uuid.UUID)
	return ok
}

// ============================================================================
// Combined Auth Context
// ============================================================================

type AuthContext struct {
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Email          string    `json:"email"`
	OrgRole        string    `json:"org_role"`
	SessionID      string    `json:"session_id"`
}

func GetAuthContext(c *fiber.Ctx) (*AuthContext, error) {
	userID, err := GetUserID(c)
	if err != nil {
		return nil, err
	}
	organizationID, err := GetOrganizationID(c)
	if err != nil {
		return nil, err
	}
	email, err := GetUserEmail(c)
	if err != nil {
		return nil, err
	}
	orgRole, err := GetOrgRole(c)
	if err != nil {
		return nil, err
	}
	sessionID, err := GetSessionID(c)
	if err != nil {
		return nil, err
	}
	return &AuthContext{
		UserID:         userID,
		OrganizationID: organizationID,
		Email:          email,
		OrgRole:        orgRole,
		SessionID:      sessionID,
	}, nil
}

// ============================================================================
// Convenience Role Checkers
// ============================================================================

func IsOwner(c *fiber.Ctx) bool {
	role, err := GetOrgRole(c)
	return err == nil && role == "owner"
}

func IsAdmin(c *fiber.Ctx) bool {
	role, err := GetOrgRole(c)
	return err == nil && (role == "owner" || role == "admin")
}

func IsViewer(c *fiber.Ctx) bool {
	role, err := GetOrgRole(c)
	return err == nil && role == "viewer"
}