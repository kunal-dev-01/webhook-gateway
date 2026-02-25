package auth

import (
    "crypto/rand"
    "encoding/base64"
    "errors" 
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    
    appErrors "shared/errors"
)

// ============================================================================
// Auth Configuration
// ============================================================================

type Config struct {
    Secret               string
    AccessTokenDuration  time.Duration
    RefreshTokenDuration time.Duration
    Issuer               string
}

// ============================================================================
// Token Manager
// ============================================================================

type TokenManager struct {
    config      Config
    secretBytes []byte
}

func NewTokenManager(cfg Config) (*TokenManager, error) {
    if len(cfg.Secret) < 32 {
        return nil, errors.New("JWT secret must be at least 32 characters")
    }
    
    // Set defaults if missing
    if cfg.AccessTokenDuration == 0 { cfg.AccessTokenDuration = 15 * time.Minute }
    if cfg.RefreshTokenDuration == 0 { cfg.RefreshTokenDuration = 7 * 24 * time.Hour }
    if cfg.Issuer == "" { cfg.Issuer = "upnixo.com" }

    return &TokenManager{
        config:      cfg,
        secretBytes: []byte(cfg.Secret),
    }, nil
}

// ============================================================================
// Claims & Structs
// ============================================================================

type Claims struct {
    UserID         uuid.UUID `json:"user_id"`
    Email          string    `json:"email"`
    OrganizationID uuid.UUID `json:"organization_id"`
    OrgRole        string    `json:"org_role"`
    SessionID      string    `json:"session_id"`
    TokenType      string    `json:"token_type"`
    jwt.RegisteredClaims
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
    TokenType    string `json:"token_type"`
    SessionID    string `json:"session_id"`
}

// ============================================================================
// Token Generation
// ============================================================================

func (m *TokenManager) GenerateTokenPair(
    userID, organizationID uuid.UUID,
    email, orgRole string,
) (*TokenPair, *appErrors.AppError) {
    
    if userID == uuid.Nil || organizationID == uuid.Nil {
        return nil, appErrors.Internal(errors.New("missing user_id or organization_id in token generation"))
    }

    sessionID, err := generateSessionID()
    if err != nil {
        return nil, appErrors.Internal(fmt.Errorf("failed to generate session ID: %w", err))
    }

    accessToken, err := m.generateToken(
        userID, organizationID, email, orgRole,
        sessionID, "access", m.config.AccessTokenDuration,
    )
    if err != nil {
        return nil, appErrors.Internal(err)
    }

    refreshToken, err := m.generateToken(
        userID, organizationID, email, orgRole,
        sessionID, "refresh", m.config.RefreshTokenDuration,
    )
    if err != nil {
        return nil, appErrors.Internal(err)
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int64(m.config.AccessTokenDuration.Seconds()),
        TokenType:    "Bearer",
        SessionID:    sessionID,
    }, nil
}

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
            Issuer:    m.config.Issuer,
            Subject:   userID.String(),
            Audience:  jwt.ClaimStrings{organizationID.String()},
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
            NotBefore: jwt.NewNumericDate(now),
            ID:        uuid.New().String(),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secretBytes)
}

// ============================================================================
// Token Validation
// ============================================================================

func (m *TokenManager) ValidateAccessToken(tokenString string) (*Claims, *appErrors.AppError) {
    return m.validateToken(tokenString, "access")
}

func (m *TokenManager) ValidateRefreshToken(tokenString string) (*Claims, *appErrors.AppError) {
    return m.validateToken(tokenString, "refresh")
}

// validateToken returns *appErrors.AppError directly
func (m *TokenManager) validateToken(tokenString, expectedType string) (*Claims, *appErrors.AppError) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return m.secretBytes, nil
    })

    // Error Handling Logic: Mapping JWT errors to Shared Errors (HTTP Statuses)
    if err != nil {
        switch {
        case errors.Is(err, jwt.ErrTokenExpired):
            // 401 Unauthorized
            return nil, appErrors.Unauthorized("Token has expired")
        case errors.Is(err, jwt.ErrTokenSignatureInvalid):
            // 401 Unauthorized
            return nil, appErrors.Unauthorized("Invalid token signature")
        case errors.Is(err, jwt.ErrTokenMalformed):
            // 401 Unauthorized
            return nil, appErrors.Unauthorized("Malformed token format")
        default:
            // 401 Unauthorized (Generic)
            return nil, appErrors.Unauthorized("Invalid token")
        }
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, appErrors.Unauthorized("Invalid token claims")
    }

    // Logic Validation
    if claims.TokenType != expectedType {
        return nil, appErrors.Unauthorized(fmt.Sprintf("Invalid token type: expected %s", expectedType))
    }

    if issuer, err := claims.GetIssuer(); err != nil || issuer != m.config.Issuer {
        return nil, appErrors.Unauthorized("Invalid token issuer")
    }

    if claims.OrganizationID == uuid.Nil {
        return nil, appErrors.Forbidden("Token missing organization context")
    }

    return claims, nil
}

// ============================================================================
// Token Refresh & Switching
// ============================================================================

func (m *TokenManager) RefreshAccessToken(refreshTokenString string) (string, *appErrors.AppError) {
    claims, err := m.ValidateRefreshToken(refreshTokenString)
    if err != nil {
        return "", err // err is already *appErrors.AppError
    }

    newToken, genErr := m.generateToken(
        claims.UserID,
        claims.OrganizationID,
        claims.Email,
        claims.OrgRole,
        claims.SessionID,
        "access",
        m.config.AccessTokenDuration,
    )
    
    if genErr != nil {
        return "", appErrors.Internal(genErr)
    }

    return newToken, nil
}

func (m *TokenManager) SwitchOrganization(
    currentTokenString string,
    newOrganizationID uuid.UUID,
    newOrgRole string,
) (*TokenPair, *appErrors.AppError) {
    
    claims, err := m.ValidateAccessToken(currentTokenString)
    if err != nil {
        return nil, err
    }

    return m.GenerateTokenPair(
        claims.UserID,
        newOrganizationID,
        claims.Email,
        newOrgRole,
    )
}

// ============================================================================
// Helpers
// ============================================================================

func generateSessionID() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

// ExtractTokenFromBearer now returns AppError for immediate failure handling
func ExtractTokenFromBearer(authHeader string) (string, *appErrors.AppError) {
    if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
        return authHeader[7:], nil
    }
    return "", appErrors.Unauthorized("Invalid or missing Authorization header")
}

// Claims methods remain same
func (c *Claims) IsOwner() bool { return c.OrgRole == "owner" }
func (c *Claims) IsAdmin() bool { return c.OrgRole == "admin" || c.OrgRole == "owner" }
func (c *Claims) CanManageMembers() bool { return c.IsAdmin() }
func (c *Claims) CanManageBilling() bool { return c.IsOwner() }
func (c *Claims) CanInviteMembers() bool {
    return c.IsAdmin()
}
func (c *Claims) IsViewer() bool { return c.OrgRole == "viewer" }