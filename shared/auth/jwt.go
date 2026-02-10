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

// Standard errors for consistent handling upstream
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrInvalidAudience  = errors.New("invalid token audience")
)

// TokenManager handles JWT operations
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

// Claims represents JWT claims
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	SessionID string    `json:"session_id"`
	TokenType string    `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"` // "Bearer"
	SessionID    string `json:"session_id"`
}

// ============================================================================
// Token Generation
// ============================================================================

// GenerateTokenPair generates both access and refresh tokens
func (m *TokenManager) GenerateTokenPair(userID uuid.UUID, email, role string) (*TokenPair, error) {
	// Generate a cryptographic random session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Generate Access Token
	accessToken, err := m.generateToken(userID, email, role, sessionID, "access", m.config.AccessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate Refresh Token
	refreshToken, err := m.generateToken(userID, email, role, sessionID, "refresh", m.config.RefreshTokenDuration)
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

func (m *TokenManager) generateToken(userID uuid.UUID, email, role, sessionID, tokenType string, duration time.Duration) (string, error) {
	now := time.Now()
	
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		SessionID: sessionID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	// Use SigningMethodHS256 explicitly
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Use pre-converted bytes
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

func (m *TokenManager) validateToken(tokenString, expectedType string) (*Claims, error) {
	// Parse and validate signature
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Strict Algorithm Check: Must be HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Use pre-converted bytes
		return m.secretBytes, nil
	})

	if err != nil {
		// Precise Error Mapping for v5
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

	// Logic Checks
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	// Verify Issuer using built-in validator
	issuer, err := claims.GetIssuer()
	if err != nil || issuer != m.config.Issuer {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ============================================================================
// Token Refresh
// ============================================================================

// RefreshAccessToken generates a new access token using a valid refresh token.
// IMPORTANT: It preserves the original SessionID to maintain session continuity.
func (m *TokenManager) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := m.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// Issue new access token, but KEEP the original SessionID
	// This ensures that if the session is revoked in Redis, this new token 
	// (which shares the sessionID) will also be blocked by middleware.
	return m.generateToken(
		claims.UserID,
		claims.Email,
		claims.Role,
		claims.SessionID,
		"access",
		m.config.AccessTokenDuration,
	)
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateSessionID creates a cryptographically secure random ID
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