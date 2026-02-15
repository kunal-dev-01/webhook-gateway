package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"shared/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ============================================================================
// Standard Errors
// ============================================================================

// Standard errors for consistent handling upstream
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrInvalidAudience  = errors.New("invalid token audience")
	ErrMissingOrgID     = errors.New("organization ID required in token")
)

// ============================================================================
// Token Manager
// ============================================================================

// TokenManager handles JWT operations for Upnixo
type TokenManager struct {
	config      *config.JWTConfig
	secretBytes []byte // Pre-converted secret to avoid allocation on every request
}

// NewTokenManager creates new JWT manager with optimized configuration
func NewTokenManager(cfg *config.JWTConfig) *TokenManager {
	return &TokenManager{
		config:      cfg,
		secretBytes: []byte(cfg.Secret), // Optimize: Convert once, use everywhere
	}
}

// ============================================================================
// Claims Structure
// ============================================================================

// Claims represents JWT claims for Upnixo webhook gateway
type Claims struct {
	// User identification
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`

	// Organization context (current active organization)
	OrganizationID uuid.UUID `json:"organization_id"` // Active organization
	OrgRole        string    `json:"org_role"`        // Role in this organization: owner, admin, member, viewer

	// Session tracking
	SessionID string `json:"session_id"`
	TokenType string `json:"token_type"` // "access" or "refresh"

	// Standard JWT claims
	jwt.RegisteredClaims
}

// ============================================================================
// Token Pair
// ============================================================================

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expires
	TokenType    string `json:"token_type"` // "Bearer"
	SessionID    string `json:"session_id"`
}

// ============================================================================
// Token Generation
// ============================================================================

// GenerateTokenPair generates both access and refresh tokens
// organizationID: The active organization user is working in
// orgRole: User's role in this organization (owner, admin, member, viewer)
func (m *TokenManager) GenerateTokenPair(
	userID, organizationID uuid.UUID,
	email, orgRole string,
) (*TokenPair, error) {
	// Validate required fields
	if userID == uuid.Nil {
		return nil, errors.New("user ID is required")
	}
	if organizationID == uuid.Nil {
		return nil, errors.New("organization ID is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if orgRole == "" {
		return nil, errors.New("organization role is required")
	}

	// Generate a cryptographic random session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Generate Access Token (short-lived: 15 minutes)
	accessToken, err := m.generateToken(
		userID, organizationID, email, orgRole,
		sessionID, "access", m.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate Refresh Token (long-lived: 7 days)
	refreshToken, err := m.generateToken(
		userID, organizationID, email, orgRole,
		sessionID, "refresh", m.config.RefreshTokenDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.config.AccessTokenDuration.Seconds()),
		TokenType:    "Bearer",
		SessionID:    sessionID,
	}, nil
}

// generateToken creates a signed JWT token
func (m *TokenManager) generateToken(
	userID, organizationID uuid.UUID,
	email, orgRole, sessionID, tokenType string,
	duration time.Duration,
) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:         userID,
		Email:          email,
		OrganizationID: organizationID,
		OrgRole:        orgRole,
		SessionID:      sessionID,
		TokenType:      tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,    // "upnixo.com"
			Subject:   userID.String(),     // User ID
			Audience:  jwt.ClaimStrings{organizationID.String()}, // Organization ID
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(), // Unique token ID (jti)
		},
	}

	// Use SigningMethodHS256 explicitly
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Use pre-converted bytes for performance
	return token.SignedString(m.secretBytes)
}

// ============================================================================
// Token Validation
// ============================================================================

// ValidateAccessToken validates and parses access token
func (m *TokenManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, "access")
}

// ValidateRefreshToken validates and parses refresh token
func (m *TokenManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, "refresh")
}

// validateToken performs comprehensive token validation
func (m *TokenManager) validateToken(tokenString, expectedType string) (*Claims, error) {
	// Parse and validate signature
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Strict Algorithm Check: Must be HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Use pre-converted bytes for performance
		return m.secretBytes, nil
	})

	if err != nil {
		// Precise Error Mapping for jwt/v5
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrExpiredToken
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, ErrInvalidSignature
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, ErrInvalidToken
		default:
			return nil, fmt.Errorf("token validation failed: %w", err)
		}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate token type
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	// Verify Issuer
	issuer, err := claims.GetIssuer()
	if err != nil || issuer != m.config.Issuer {
		return nil, ErrInvalidClaims
	}

	// CRITICAL: Validate OrganizationID is present
	if claims.OrganizationID == uuid.Nil {
		return nil, ErrMissingOrgID
	}

	// Validate UserID
	if claims.UserID == uuid.Nil {
		return nil, ErrInvalidClaims
	}

	// Validate Email
	if claims.Email == "" {
		return nil, ErrInvalidClaims
	}

	// Validate OrgRole
	validRoles := map[string]bool{
		"owner":  true,
		"admin":  true,
		"member": true,
		"viewer": true,
	}
	if !validRoles[claims.OrgRole] {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ============================================================================
// Token Refresh
// ============================================================================

// RefreshAccessToken generates a new access token using a valid refresh token.
// IMPORTANT: Preserves the original SessionID to maintain session continuity.
// If the session is revoked in Redis, both old and new tokens will be blocked.
func (m *TokenManager) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := m.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// Issue new access token, preserving SessionID and OrganizationID
	// This ensures session revocation in Redis affects all tokens with this sessionID
	return m.generateToken(
		claims.UserID,
		claims.OrganizationID, // Preserve organization context
		claims.Email,
		claims.OrgRole,        // Preserve role in organization
		claims.SessionID,      // CRITICAL: Preserve session ID
		"access",
		m.config.AccessTokenDuration,
	)
}

// ============================================================================
// Organization Switching
// ============================================================================

// SwitchOrganization generates new tokens for a different organization
// Used when user switches between organizations in the dashboard
func (m *TokenManager) SwitchOrganization(
	currentTokenString string,
	newOrganizationID uuid.UUID,
	newOrgRole string,
) (*TokenPair, error) {
	// Validate current token
	claims, err := m.ValidateAccessToken(currentTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid current token: %w", err)
	}

	// Validate new organization ID
	if newOrganizationID == uuid.Nil {
		return nil, errors.New("new organization ID is required")
	}

	// Generate new token pair for new organization
	// Note: This creates a NEW session (new SessionID)
	// The old session remains valid until expiry/revocation
	return m.GenerateTokenPair(
		claims.UserID,
		newOrganizationID,
		claims.Email,
		newOrgRole,
	)
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateSessionID creates a cryptographically secure random session ID
func generateSessionID() (string, error) {
	// 32 bytes gives 256 bits of entropy
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// URL-safe base64 is better for web headers than hex
	return base64.URLEncoding.EncodeToString(b), nil
}

// ExtractTokenFromBearer extracts token from "Bearer <token>" header
func ExtractTokenFromBearer(authHeader string) (string, error) {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:], nil
	}
	return "", errors.New("invalid authorization header format")
}

// ============================================================================
// Claim Helpers
// ============================================================================

// IsOwner checks if user is owner in the current organization
func (c *Claims) IsOwner() bool {
	return c.OrgRole == "owner"
}

// IsAdmin checks if user is admin or owner in the current organization
func (c *Claims) IsAdmin() bool {
	return c.OrgRole == "admin" || c.OrgRole == "owner"
}

// CanManageMembers checks if user can manage team members
func (c *Claims) CanManageMembers() bool {
	return c.IsAdmin()
}

// CanManageBilling checks if user can manage billing
func (c *Claims) CanManageBilling() bool {
	return c.IsOwner() // Only owner can manage billing
}

// CanInviteMembers checks if user can invite new members
func (c *Claims) CanInviteMembers() bool {
	return c.IsAdmin()
}

// IsViewer checks if user is viewer (read-only)
func (c *Claims) IsViewer() bool {
	return c.OrgRole == "viewer"
}