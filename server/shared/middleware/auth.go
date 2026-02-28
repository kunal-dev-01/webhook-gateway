package middleware

import (
    "fmt"
    "slices"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"

    "shared/auth"
    appErrors "shared/errors" // Import shared errors
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
            // Type assertion: Check agar ye humara custom AppError hai
            if appErr, ok := err.(*appErrors.AppError); ok {
                return c.Status(appErr.Code).JSON(appErr)
            }
            // Agar koi aur anjaan error aaya, toh usko 500 Internal error bana do
            fallback := appErrors.Internal(err)
            return c.Status(fallback.Code).JSON(fallback)
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
        
        // If validation fails, we just proceed as unauthenticated (visitor)
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
            e := appErrors.Unauthorized("Unauthorized: Role not found in context")
            return c.Status(e.Code).JSON(e)
        }

        if slices.Contains(allowedRoles, orgRole) {
            return c.Next()
        }

        e := appErrors.Forbidden(fmt.Sprintf("Forbidden: Requires one of roles %v", allowedRoles))
        return c.Status(e.Code).JSON(e)
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
            e := appErrors.Unauthorized("Unauthorized: Role not found")
            return c.Status(e.Code).JSON(e)
        }
        
        minLevel, minExists := roleHierarchy[minRole]
        userLevel, userExists := roleHierarchy[orgRole]
        
        if !minExists || !userExists {
            e := appErrors.Forbidden("Invalid role configuration")
            return c.Status(e.Code).JSON(e)
        }
        
        if userLevel >= minLevel {
            return c.Next()
        }

        e := appErrors.Forbidden(fmt.Sprintf("Forbidden: Requires minimum role %s", minRole))
        return c.Status(e.Code).JSON(e)
    }
}

// ============================================================================
// Organization Context Validation
// ============================================================================

func (m *AuthMiddleware) RequireOrganization() fiber.Handler {
    return func(c *fiber.Ctx) error {
        claimsOrgID, err := GetOrganizationID(c)
        if err != nil {
            e := appErrors.Unauthorized("Organization context not found")
            return c.Status(e.Code).JSON(e)
        }
        
        orgIDParam := c.Params("org_id")
        if orgIDParam == "" {
            e := appErrors.BadRequest("Organization ID required in URL")
            return c.Status(e.Code).JSON(e)
        }
        
        paramOrgID, parseErr := uuid.Parse(orgIDParam)
        if parseErr != nil {
            e := appErrors.BadRequest("Invalid organization ID format")
            return c.Status(e.Code).JSON(e)
        }
        
        if claimsOrgID != paramOrgID {
            e := appErrors.Forbidden("Access denied: Organization mismatch")
            return c.Status(e.Code).JSON(e)
        }
        
        return c.Next()
    }
}

// ============================================================================
// Permission Helpers
// ============================================================================

// Helper: Reconstructs auth.Claims to use its methods
func (m *AuthMiddleware) RequirePermission(checkFunc func(*auth.Claims) bool) fiber.Handler {
    return func(c *fiber.Ctx) error {
        authCtx, err := GetAuthContext(c)
        if err != nil {
            e := appErrors.Unauthorized("Unauthorized: Context missing")
            return c.Status(e.Code).JSON(e)
        }

        // Reconstruct claims to access helper methods (IsOwner, CanInviteMembers, etc.)
        claims := &auth.Claims{
            UserID:         authCtx.UserID,
            OrganizationID: authCtx.OrganizationID,
            Email:          authCtx.Email,
            OrgRole:        authCtx.OrgRole,
            // SessionID not strictly needed for permission checks, but good to have
            SessionID:      authCtx.SessionID,
        }

        if !checkFunc(claims) {
            e := appErrors.Forbidden("Forbidden: Insufficient permissions")
            return c.Status(e.Code).JSON(e)
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

// Changed return type to *appErrors.AppError
func (m *AuthMiddleware) verifyRequest(c *fiber.Ctx) (*auth.Claims, error) {
    // 1. Extract Authorization header
    authHeader := c.Get("Authorization")
    if authHeader == "" {
        return nil, appErrors.Unauthorized("Missing authorization header")
    }

    // 2. Parse "Bearer <token>"
    // ExtractTokenFromBearer now returns *AppError directly
    tokenString, err := auth.ExtractTokenFromBearer(authHeader)
    if err != nil {
        return nil, err
    }

    // 3. Validate JWT
    // ValidateAccessToken now returns *AppError (handling expiration, signature, etc.)
    claims, jwtErr := m.jwtManager.ValidateAccessToken(tokenString)
    if jwtErr != nil {
        return nil, jwtErr
    }

    // 4. Redis session check
    key := sessionKey(claims.UserID, claims.SessionID)

    exists, redisErr := m.redis.Exists(c.Context(), key).Result()
    if redisErr != nil {
        // Fail closed: Redis down = deny access
        // Wrap Redis error into Internal Server Error
        return nil, appErrors.Internal(fmt.Errorf("system error: session check failed: %w", redisErr))
    }
    
    if exists == 0 {
        return nil, appErrors.Unauthorized("Session expired or revoked")
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

// Getters remain mostly unchanged, just strictly typed
func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
    if id, ok := c.Locals(CtxUserID).(uuid.UUID); ok {
        return id, nil
    }
    return uuid.Nil, fmt.Errorf("user ID not found in context")
}

func GetOrganizationID(c *fiber.Ctx) (uuid.UUID, error) {
    if id, ok := c.Locals(CtxOrganizationID).(uuid.UUID); ok {
        return id, nil
    }
    return uuid.Nil, fmt.Errorf("organization ID not found in context")
}

func GetUserEmail(c *fiber.Ctx) (string, error) {
    if email, ok := c.Locals(CtxUserEmail).(string); ok {
        return email, nil
    }
    return "", fmt.Errorf("email not found in context")
}

func GetOrgRole(c *fiber.Ctx) (string, error) {
    if role, ok := c.Locals(CtxOrgRole).(string); ok {
        return role, nil
    }
    return "", fmt.Errorf("organization role not found in context")
}

func GetSessionID(c *fiber.Ctx) (string, error) {
    if sid, ok := c.Locals(CtxSessionID).(string); ok {
        return sid, nil
    }
    return "", fmt.Errorf("session ID not found in context")
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
    if err != nil { return nil, err }
    
    organizationID, err := GetOrganizationID(c)
    if err != nil { return nil, err }
    
    email, err := GetUserEmail(c)
    if err != nil { return nil, err }
    
    orgRole, err := GetOrgRole(c)
    if err != nil { return nil, err }
    
    sessionID, err := GetSessionID(c)
    if err != nil { return nil, err }

    return &AuthContext{
        UserID:         userID,
        OrganizationID: organizationID,
        Email:          email,
        OrgRole:        orgRole,
        SessionID:      sessionID,
    }, nil
}

// ============================================================================
// Convenience Role Checkers (For Handlers)
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