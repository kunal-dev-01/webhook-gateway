package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"server/internal/dto"
	"server/internal/repository"

	"github.com/google/uuid"
	"shared/auth"
	"shared/config"
	"shared/models"
	"shared/utils"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo      *repository.UserRepository
	tokenRepo     *repository.TokenRepository
	jwtManager    *auth.Manager
	totpManager   *auth.TOTPManager
	sessionMgr    *auth.SessionManager
	emailService  *utils.EmailService
	config        *config.Config
}

// NewAuthService creates new auth service
func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	jwtManager *auth.Manager,
	totpManager *auth.TOTPManager,
	sessionMgr *auth.SessionManager,
	emailService *utils.EmailService,
	config *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		jwtManager:   jwtManager,
		totpManager:  totpManager,
		sessionMgr:   sessionMgr,
		emailService: emailService,
		config:       config,
	}
}

// ============================================================================
// Registration
// ============================================================================

// Register creates new user account
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) error {
	// Validate password strength
	if err := utils.ValidatePassword(req.Password); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Check if email already exists
	exists, err := s.userRepo.Exists(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return fmt.Errorf("email already registered")
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        strings.ToLower(req.Email),
		FullName:     req.FullName,
		Role:         models.RoleAdmin,
		Status:       models.StatusActive,
		Timezone:     "UTC",
		Language:     "en",
		EmailVerified: false,
		TwoFactorEnabled: false,
		PasswordChangedAt: time.Now(),
	}

	// Hash password
	if err := user.SetPassword(req.Password); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Save user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Send verification email
	if err := s.sendVerificationEmail(ctx, user); err != nil {
		fmt.Printf("User created but failed to send verification email: %v\n", err)
		return nil
	}

	return nil
}

// ============================================================================
// Login
// ============================================================================

// Login authenticates user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest, ip, userAgent string) (*dto.LoginResponse, error) {
	// Check Redis rate limit for this IP (max 10 attempts per 15 minutes)
	limited, err := s.tokenRepo.CheckLoginRateLimit(ctx, ip, 10, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}
	if limited {
		return nil, fmt.Errorf("too many login attempts from this IP, please try again later")
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Increment failed attempts in DB
		s.userRepo.IncrementFailedAttempts(ctx, req.Email)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user can login
	if !user.CanLogin() {
		if user.IsLocked() {
			return nil, fmt.Errorf("account is locked due to too many failed attempts")
		}
		if user.Status == models.StatusInactive {
			return nil, fmt.Errorf("account is inactive")
		}
		if user.Status == models.StatusSuspended {
			return nil, fmt.Errorf("account is suspended")
		}
		if user.IsDeleted() {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("account cannot login")
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		// Increment failed attempts in DB
		s.userRepo.IncrementFailedAttempts(ctx, req.Email)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		// Generate temporary token for 2FA verification
		tempToken, err := utils.GenerateRandomToken(32)
		if err != nil {
			return nil, fmt.Errorf("failed to generate temp token: %w", err)
		}

		// Store temp token in Redis (5 minutes)
		if err := s.tokenRepo.Store2FATempToken(ctx, user.ID, tempToken, 5*time.Minute); err != nil {
			return nil, fmt.Errorf("failed to store temp token: %w", err)
		}

		return &dto.LoginResponse{
			Require2FA: true,
			TempToken:  tempToken,
		}, nil
	}

	// Complete login
	return s.completeLogin(ctx, user, ip, userAgent)
}

// Verify2FA verifies 2FA code and completes login
func (s *AuthService) Verify2FA(ctx context.Context, req *dto.Verify2FARequest, ip, userAgent string) (*dto.LoginResponse, error) {
	// Get user ID from temp token
	userID, err := s.tokenRepo.Get2FATempToken(ctx, req.TempToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired 2FA token")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Verify TOTP code
	if user.TwoFactorSecret == nil {
		return nil, fmt.Errorf("2FA not enabled for this user")
	}

	// Decrypt TOTP secret
	secret, err := utils.Decrypt(*user.TwoFactorSecret, s.totpManager.GetAESKey())
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}

	// Validate code
	if !s.totpManager.ValidateCode(secret, req.Code) {
		return nil, fmt.Errorf("invalid 2FA code")
	}

	// Delete temp token
	s.tokenRepo.Delete2FATempToken(ctx, req.TempToken)

	// Complete login
	return s.completeLogin(ctx, user, ip, userAgent)
}

// completeLogin generates tokens and creates session
func (s *AuthService) completeLogin(ctx context.Context, user *models.User, ip, userAgent string) (*dto.LoginResponse, error) {
	// Generate JWT tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	if err := s.sessionMgr.CreateSession(
		ctx,
		user.ID,
		user.Email,
		string(user.Role),
		tokenPair.SessionID,
		ip,
		userAgent,
		s.config.JWT.RefreshTokenDuration,
	); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, ip); err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken, // For cookie
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         user.ToPublic(),
		Require2FA:   false,
	}, nil
}

// ============================================================================
// Email Verification
// ============================================================================

// VerifyEmail verifies user email with token
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// Get user ID from token
	userID, err := s.tokenRepo.GetVerificationToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid or expired verification token")
	}

	// Mark email as verified
	if err := s.userRepo.MarkEmailAsVerified(ctx, userID); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	// Delete token
	s.tokenRepo.DeleteVerificationToken(ctx, token)

	return nil
}

// ResendVerification sends new verification email
func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	// Get user
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Check if already verified
	if user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	// Check rate limit (max 3 per hour)
	limited, err := s.tokenRepo.CheckEmailRateLimit(ctx, email, 3, time.Hour)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if limited {
		return fmt.Errorf("too many verification emails sent, please try again later")
	}

	// Send verification email
	return s.sendVerificationEmail(ctx, user)
}

// sendVerificationEmail sends verification email
func (s *AuthService) sendVerificationEmail(ctx context.Context, user *models.User) error {
	// Generate token
	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in Redis (24 hours)
	if err := s.tokenRepo.StoreVerificationToken(ctx, user.ID, token, 24*time.Hour); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Send email
	return s.emailService.SendVerificationEmail(
		ctx,
		user.Email,
		token,
		s.config.Frontend.VerificationUrl,
		user.FullName,
	)
}

// ============================================================================
// Password Reset
// ============================================================================

// ForgotPassword sends password reset email
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	// Get user
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists
		return nil
	}

	// Check rate limit (max 3 per hour)
	limited, err := s.tokenRepo.CheckEmailRateLimit(ctx, email, 3, time.Hour)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if limited {
		return nil // Don't reveal rate limit
	}

	// Generate token
	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in Redis (1 hour)
	if err := s.tokenRepo.StoreResetToken(ctx, user.ID, token, time.Hour); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Send email
	return s.emailService.SendPasswordResetEmail(
		ctx,
		user.Email,
		token,
		s.config.Frontend.ResetUrl,
		user.FullName,
	)
}

// ResetPassword resets password with token
func (s *AuthService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	// Validate password
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Get user ID from token
	userID, err := s.tokenRepo.GetResetToken(ctx, req.Token)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Hash new password
	if err := user.SetPassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, user.ID, user.PasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete reset token
	s.tokenRepo.DeleteResetToken(ctx, req.Token)

	// Invalidate all sessions (force re-login)
	s.sessionMgr.DeleteAllUserSessions(ctx, user.ID)

	return nil
}

// ChangePassword changes password for logged-in user
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, req *dto.ChangePasswordRequest) error {
	// Validate new password
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify current password
	if !user.CheckPassword(req.CurrentPassword) {
		return fmt.Errorf("current password is incorrect")
	}

	// Check if new password is same as current
	if user.CheckPassword(req.NewPassword) {
		return fmt.Errorf("new password must be different from current password")
	}

	// Hash new password
	if err := user.SetPassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, user.ID, user.PasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all other sessions (keep current one)
	// This should be done by deleting all sessions except current one
	// For simplicity, we'll invalidate all and user can re-login

	return nil
}

// ============================================================================
// GET CURRENT USER
// ============================================================================

// GetCurrentUser returns current logged-in user info
// GetMe returns current user info
func (s *AuthService) GetMe(ctx context.Context, userID uuid.UUID) (*dto.MeResponse, error) {

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.MeResponse{
		ID:               user.ID,
		Email:            user.Email,
		FullName:         user.FullName,
		Phone:            user.Phone,
		Role:             user.Role,
		Status:           user.Status,
		EmailVerified:    user.EmailVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		LastLoginAt:      user.LastLoginAt,
		CreatedAt:        user.CreatedAt,
		PasswordChangedAt: user.PasswordChangedAt,
	}, nil
}

// ============================================================================
// Token Refresh
// ============================================================================

// RefreshAccessToken generates new access token from refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token")
	}

	// Validate session exists
	valid, err := s.sessionMgr.ValidateSession(ctx, claims.UserID, claims.SessionID)
	if err != nil || !valid {
		return "", fmt.Errorf("session not found or expired")
	}

	// Generate new access token
	accessToken, err := s.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update session activity
	s.sessionMgr.UpdateSessionActivity(ctx, claims.UserID, claims.SessionID)

	return accessToken, nil
}

// ============================================================================
// Logout
// ============================================================================

// Logout invalidates current session
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, sessionID string) error {
	return s.sessionMgr.DeleteSession(ctx, userID, sessionID)
}

// LogoutAll invalidates all user sessions
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.sessionMgr.DeleteAllUserSessions(ctx, userID)
}

// ============================================================================
// Two-Factor Authentication
// ============================================================================

// Setup2FA generates 2FA secret and recovery codes
func (s *AuthService) Setup2FA(ctx context.Context, userID uuid.UUID) (*dto.Setup2FAResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.TwoFactorEnabled {
		return nil, fmt.Errorf("2FA is already enabled")
	}

	// Generate TOTP secret
	key, err := s.totpManager.GenerateSecret(user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	// Generate recovery codes
	recoveryCodes, err := s.totpManager.GenerateRecoveryCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recovery codes: %w", err)
	}

	return &dto.Setup2FAResponse{
		Secret:        key.Secret(),
		QRCodeURL:     key.URL(),
		RecoveryCodes: recoveryCodes,
	}, nil
}

// Enable2FA verifies code and enables 2FA
func (s *AuthService) Enable2FA(ctx context.Context, userID uuid.UUID, req *dto.Enable2FARequest, secret string, recoveryCodes []string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Validate code
	if !s.totpManager.ValidateCode(secret, req.Code) {
		return fmt.Errorf("invalid verification code")
	}

	// Encrypt secret before storing
	encryptedSecret, err := utils.Encrypt(secret, s.totpManager.GetAESKey())
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Enable 2FA
	user.Enable2FA(encryptedSecret, recoveryCodes)

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return nil
}

// Disable2FA disables 2FA with password verification
func (s *AuthService) Disable2FA(ctx context.Context, userID uuid.UUID, req *dto.Disable2FARequest) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		return fmt.Errorf("incorrect password")
	}

	// Disable 2FA
	user.Disable2FA()

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	return nil
}

// ============================================================================
// Session Management
// ============================================================================

// ListSessions returns all active sessions for user
func (s *AuthService) ListSessions(ctx context.Context, userID uuid.UUID) (*dto.ListSessionsResponse, error) {
	sessions, err := s.sessionMgr.ListUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessionResponses := make([]*dto.SessionResponse, len(sessions))
	for i, session := range sessions {
		sessionResponses[i] = &dto.SessionResponse{
			SessionID:    session.SessionID,
			DeviceInfo:   session.DeviceInfo,
			IPAddress:    session.IPAddress,
			UserAgent:    session.UserAgent,
			CreatedAt:    session.CreatedAt,
			LastActivity: session.LastActivity,
		}
	}

	return &dto.ListSessionsResponse{
		Sessions: sessionResponses,
		Total:    len(sessionResponses),
	}, nil
}

// DeleteSession deletes specific session
func (s *AuthService) DeleteSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	return s.sessionMgr.DeleteSession(ctx, userID, sessionID)
}