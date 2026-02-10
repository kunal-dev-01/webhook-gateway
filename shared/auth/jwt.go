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

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrInvalidClaims    = errors.New("invalid token claims")
)

// Manager handles JWT operations
type Manager struct {
	config *config.JWTConfig
}

// NewManager creates new JWT manager from config
func NewManager(cfg *config.JWTConfig) *Manager {
	return &Manager{config: cfg}
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
func (m *Manager) GenerateTokenPair(userID uuid.UUID, email, role string) (*TokenPair, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	accessToken, err := m.generateToken(userID, email, role, sessionID, "access", m.config.AccessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

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

// generateToken creates a JWT token
func (m *Manager) generateToken(userID uuid.UUID, email, role, sessionID, tokenType string, duration time.Duration) (string, error) {
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.Secret))
}

// ============================================================================
// Token Validation
// ============================================================================

// ValidateAccessToken validates and parses access token
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, "access")
}

// ValidateRefreshToken validates and parses refresh token
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, "refresh")
}

// validateToken validates JWT token
func (m *Manager) validateToken(tokenString, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrInvalidSignature
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("expected %s token, got %s", expectedType, claims.TokenType)
	}

	if claims.Issuer != m.config.Issuer {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ============================================================================
// Token Refresh
// ============================================================================

// RefreshAccessToken generates new access token from refresh token
func (m *Manager) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := m.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", err
	}

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

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ExtractTokenFromBearer extracts token from "Bearer {token}" string
func ExtractTokenFromBearer(authHeader string) (string, error) {
	const prefix = "Bearer "
	if len(authHeader) < len(prefix) {
		return "", errors.New("invalid authorization header format")
	}
	if authHeader[:len(prefix)] != prefix {
		return "", errors.New("missing Bearer prefix")
	}
	return authHeader[len(prefix):], nil
}