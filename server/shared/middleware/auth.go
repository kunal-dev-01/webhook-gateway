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

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

// Context keys for auth data
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

// AuthMiddleware handles JWT authentication for Upnixo
type AuthMiddleware struct {
	jwtManager *auth.TokenManager
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
// Core Middleware Handlers
// ============================================================================

// RequireAuth validates JWT and checks Redis for session revocation
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := m.verifyRequest(c)
		if err != nil {
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
// Useful for endpoints that work both for authenticated and guest users
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

// ============================================================================
// Role-Based Access Control (RBAC)
// ============================================================================

// RequireRole enforces specific roles in current organization
// Example: RequireRole("owner", "admin")
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

// RequireOwner ensures only organization owner can access
// Used for sensitive operations: billing, subscription, delete org
func (m *AuthMiddleware) RequireOwner() fiber.Handler {
	return m.RequireRole("owner")
}

// RequireAdmin ensures admin or owner can access
// Used for: manage members, invitations, settings
func (m *AuthMiddleware) RequireAdmin() fiber.Handler {
	return m.RequireRole("owner", "admin")
}

// RequireMember ensures at least member access (not viewer)
// Used for: create/edit projects, send webhooks
func (m *AuthMiddleware) RequireMember() fiber.Handler {
	return m.RequireRole("owner", "admin", "member")
}

// RequireMinRole enforces minimum role level based on hierarchy
// Hierarchy: owner > admin > member > viewer
func (m *AuthMiddleware) RequireMinRole(minRole string) fiber.Handler {
	// Role hierarchy (higher value = more privilege)
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

// RequireOrganization ensures the organization in the JWT matches URL param
// Used to prevent cross-organization access
// Example: /api/organizations/:org_id/projects
func (m *AuthMiddleware) RequireOrganization() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get organization ID from JWT claims
		claimsOrgID, err := GetOrganizationID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "organization context not found",
			})
		}

		// Get organization ID from URL parameter
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

		// Verify they match (prevent cross-org access)
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

// RequirePermission checks specific permission for current user
// Uses helper methods from JWT claims
func (m *AuthMiddleware) RequirePermission(checkFunc func(*auth.Claims) bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get full auth context
		authCtx, err := GetAuthContext(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "unauthorized",
			})
		}

		// Reconstruct claims for permission check
		claims := &auth.Claims{
			UserID:         authCtx.UserID,
			OrganizationID: authCtx.OrganizationID,
			Email:          authCtx.Email,
			OrgRole:        authCtx.OrgRole,
		}

		// Check permission
		if !checkFunc(claims) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "forbidden: insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireCanManageBilling ensures user can manage billing (owner only)
func (m *AuthMiddleware) RequireCanManageBilling() fiber.Handler {
	return m.RequirePermission(func(c *auth.Claims) bool {
		return c.CanManageBilling()
	})
}

// RequireCanManageMembers ensures user can manage team members
func (m *AuthMiddleware) RequireCanManageMembers() fiber.Handler {
	return m.RequirePermission(func(c *auth.Claims) bool {
		return c.CanManageMembers()
	})
}

// RequireCanInviteMembers ensures user can invite new members
func (m *AuthMiddleware) RequireCanInviteMembers() fiber.Handler {
	return m.RequirePermission(func(c *auth.Claims) bool {
		return c.CanInviteMembers()
	})
}

// ============================================================================
// Internal Helpers
// ============================================================================

// verifyRequest handles the complete verification flow
// Header → Token → Claims → Redis Session Check
func (m *AuthMiddleware) verifyRequest(c *fiber.Ctx) (*auth.Claims, error) {
	// 1. Extract Authorization Header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	// 2. Parse Token String from "Bearer <token>"
	tokenString, err := auth.ExtractTokenFromBearer(authHeader)
	if err != nil {
		return nil, errors.New("invalid authorization header format")
	}

	// 3. Validate JWT (signature, expiry, claims)
	claims, err := m.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredToken) {
			return nil, errors.New("token expired")
		}
		if errors.Is(err, auth.ErrMissingOrgID) {
			return nil, errors.New("invalid token: missing organization")
		}
		return nil, errors.New("invalid token")
	}

	// 4. Redis Session Check (Is session still active?)
	// Key format: "session:USER_ID:SESSION_ID"
	sessionKey := fmt.Sprintf("session:%s:%s", claims.UserID, claims.SessionID)

	// Use Exists (fast, no data fetch)
	exists, err := m.redis.Exists(c.Context(), sessionKey).Result()
	if err != nil {
		// Fail closed: If Redis is down, deny access for security
		return nil, errors.New("system error: session check failed")
	}

	if exists == 0 {
		return nil, errors.New("session expired or revoked")
	}

	return claims, nil
}

// setContext injects claims into Fiber locals using type-safe keys
func (m *AuthMiddleware) setContext(c *fiber.Ctx, claims *auth.Claims) {
	c.Locals(CtxUserID, claims.UserID)
	c.Locals(CtxOrganizationID, claims.OrganizationID)
	c.Locals(CtxUserEmail, claims.Email)
	c.Locals(CtxOrgRole, claims.OrgRole)
	c.Locals(CtxSessionID, claims.SessionID)
}

// ============================================================================
// Public Context Getters (Type-Safe)
// ============================================================================

// GetUserID retrieves user ID from context
func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	val := c.Locals(CtxUserID)
	if id, ok := val.(uuid.UUID); ok {
		return id, nil
	}
	return uuid.Nil, errors.New("user ID not found in context")
}

// GetOrganizationID retrieves organization ID from context
// CRITICAL: This is the active organization the user is working in
func GetOrganizationID(c *fiber.Ctx) (uuid.UUID, error) {
	val := c.Locals(CtxOrganizationID)
	if id, ok := val.(uuid.UUID); ok {
		return id, nil
	}
	return uuid.Nil, errors.New("organization ID not found in context")
}

// GetUserEmail retrieves user email from context
func GetUserEmail(c *fiber.Ctx) (string, error) {
	val := c.Locals(CtxUserEmail)
	if email, ok := val.(string); ok {
		return email, nil
	}
	return "", errors.New("email not found in context")
}

// GetOrgRole retrieves user's role in current organization
func GetOrgRole(c *fiber.Ctx) (string, error) {
	val := c.Locals(CtxOrgRole)
	if role, ok := val.(string); ok {
		return role, nil
	}
	return "", errors.New("organization role not found in context")
}

// GetSessionID retrieves session ID from context
func GetSessionID(c *fiber.Ctx) (string, error) {
	val := c.Locals(CtxSessionID)
	if sid, ok := val.(string); ok {
		return sid, nil
	}
	return "", errors.New("session ID not found in context")
}

// IsAuthenticated fast check if user is authenticated
func IsAuthenticated(c *fiber.Ctx) bool {
	_, ok := c.Locals(CtxUserID).(uuid.UUID)
	return ok
}

// ============================================================================
// Combined Getters (Convenience)
// ============================================================================

// AuthContext represents complete authentication context
type AuthContext struct {
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Email          string    `json:"email"`
	OrgRole        string    `json:"org_role"`
	SessionID      string    `json:"session_id"`
}

// GetAuthContext retrieves complete auth context in one call
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
// Helper Functions
// ============================================================================

// IsOwner checks if current user is owner in active organization
func IsOwner(c *fiber.Ctx) bool {
	role, err := GetOrgRole(c)
	if err != nil {
		return false
	}
	return role == "owner"
}

// IsAdmin checks if current user is admin or owner
func IsAdmin(c *fiber.Ctx) bool {
	role, err := GetOrgRole(c)
	if err != nil {
		return false
	}
	return role == "owner" || role == "admin"
}

// IsViewer checks if current user is viewer (read-only)
func IsViewer(c *fiber.Ctx) bool {
	role, err := GetOrgRole(c)
	if err != nil {
		return false
	}
	return role == "viewer"
}