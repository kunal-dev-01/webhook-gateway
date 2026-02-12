package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"server/internal/dto"
	"server/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"shared/auth"
	"shared/config"
	"shared/models"
	"shared/utils"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo       *repository.UserRepository
	companyRepo    *repository.CompanyRepository // NEW: For company operations
	tokenRepo      *repository.TokenRepository
	jwtManager     *auth.TokenManager
	totpManager    *auth.TOTPManager
	sessionMgr     *auth.SessionManager
	emailService   *utils.EmailService
	config         *config.Config
	logger         zerolog.Logger
}

// NewAuthService creates new auth service
// UPDATED: Added companyRepo parameter
func NewAuthService(
	userRepo *repository.UserRepository,
	companyRepo *repository.CompanyRepository, // NEW
	tokenRepo *repository.TokenRepository,
	jwtManager *auth.TokenManager,
	totpManager *auth.TOTPManager,
	sessionMgr *auth.SessionManager,
	emailService *utils.EmailService,
	config *config.Config,
	logger zerolog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		companyRepo:  companyRepo, // NEW
		tokenRepo:    tokenRepo,
		jwtManager:   jwtManager,
		totpManager:  totpManager,
		sessionMgr:   sessionMgr,
		emailService: emailService,
		config:       config,
		logger:       logger.With().Str("service", "auth_service").Logger(),
	}
}

// ============================================================================
// Registration
// ============================================================================

// Register creates a new company and owner user
// UPDATED: Now creates company + owner user (multi-tenant onboarding)
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) error {
	if err := utils.ValidatePassword(req.Password); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if email already exists (across all companies)
	exists, err := s.userRepo.Exists(ctx, email)
	if err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("failed to check user existence")
		return errors.New("internal server error")
	}
	if exists {
		return errors.New("email already registered")
	}

	// Create Company first
	company := &models.Company{
		ID:     uuid.New(),
		Name:   req.CompanyName,
		Status: models.CompanyStatusActive,
		// OwnerID will be set after user creation
	}

	if err := s.companyRepo.Create(ctx, company); err != nil {
		s.logger.Error().Err(err).Str("company_name", req.CompanyName).Msg("failed to create company")
		return errors.New("failed to create company")
	}

	// Create Owner User
	user := &models.User{
		ID:                uuid.New(),
		CompanyID:         company.ID, // NEW: Link to company
		Email:             email,
		FullName:          req.FullName,
		Role:              models.UserRoleOwner, // NEW: Owner role
		Status:            models.UserStatusPendingVerification,
		IsActive:          true,
		Timezone:          "UTC",
		Language:          "en",
		EmailVerified:     false,
		TwoFactorEnabled:  false,
		PasswordChangedAt: time.Now().UTC(),
	}

	if err := user.SetPassword(req.Password); err != nil {
		s.logger.Error().Err(err).Msg("failed to hash password")
		// Rollback company creation
		_ = s.companyRepo.Delete(ctx, company.ID)
		return errors.New("internal server error")
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("failed to create user")
		// Rollback company creation
		_ = s.companyRepo.Delete(ctx, company.ID)
		return errors.New("failed to create account")
	}

	// Update company with owner_id
	company.OwnerID = user.ID
	if err := s.companyRepo.Update(ctx, company); err != nil {
		s.logger.Error().Err(err).Msg("failed to set company owner")
		// Continue anyway - can be fixed manually
	}

	s.sendVerificationEmailAsync(user)
	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("company_id", company.ID.String()).
		Msg("company and owner registered")
	return nil
}

// ============================================================================
// Login & Authentication
// ============================================================================

// Login authenticates user and returns tokens
// NOTE: Email alone is not unique - we check across all companies
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest, ip, userAgent string) (*dto.LoginResponse, error) {
	limited, err := s.tokenRepo.CheckLoginRateLimit(ctx, ip, 10, 15*time.Minute)
	if err != nil {
		s.logger.Error().Err(err).Str("ip", ip).Msg("redis rate limit error")
		return nil, errors.New("internal error")
	}
	if limited {
		return nil, errors.New("too many login attempts, please try again later")
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	
	// NOTE: GetByEmail searches across all companies
	// If your DB has unique(company_id, email), this still works
	// because email can be duplicate across companies but unique within
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't increment for specific company - increment globally by email
		return nil, errors.New("invalid credentials")
	}

	if !user.CanLogin() {
		s.logger.Warn().
			Str("user_id", user.ID.String()).
			Str("status", string(user.Status)).
			Msg("login blocked")
		return nil, errors.New("account is locked or inactive")
	}

	if !user.CheckPassword(req.Password) {
		// Increment within company context
		_ = s.userRepo.IncrementFailedAttempts(ctx, email, user.CompanyID)
		return nil, errors.New("invalid credentials")
	}

	if user.TwoFactorEnabled {
		return s.initiate2FALogin(ctx, user)
	}

	return s.completeLogin(ctx, user, ip, userAgent)
}

func (s *AuthService) initiate2FALogin(ctx context.Context, user *models.User) (*dto.LoginResponse, error) {
	tempToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return nil, errors.New("internal error")
	}

	if err := s.tokenRepo.Store2FATempToken(ctx, user.ID, tempToken, 5*time.Minute); err != nil {
		s.logger.Error().Err(err).Msg("failed to store 2fa temp token")
		return nil, errors.New("login failed")
	}

	return &dto.LoginResponse{Require2FA: true, TempToken: tempToken}, nil
}

// ============================================================================
// Verify 2FA
// ============================================================================

func (s *AuthService) Verify2FA(ctx context.Context, req *dto.Verify2FARequest, ip, userAgent string) (*dto.LoginResponse, error) {
	userID, err := s.tokenRepo.Get2FATempToken(ctx, req.TempToken)
	if err != nil {
		return nil, errors.New("session expired, please login again")
	}

	limited, _ := s.tokenRepo.Check2FARateLimit(ctx, userID, 5, 5*time.Minute)
	if limited {
		return nil, errors.New("too many attempts, try again later")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	isValid := false
	isRecovery := len(req.Code) == 8

	if isRecovery {
		var idx int
		isValid, idx = s.totpManager.ValidateRecoveryCode(req.Code, user.TwoFactorRecoveryCodes)
		if isValid {
			if used, _ := s.tokenRepo.IsRecoveryCodeUsed(ctx, userID, idx); used {
				s.logger.Warn().Str("user_id", userID.String()).Int("idx", idx).Msg("recovery code reuse attempt")
				return nil, errors.New("recovery code already used")
			}
			_ = s.tokenRepo.MarkRecoveryCodeUsed(ctx, userID, idx, 30*24*time.Hour)
			s.burnRecoveryCode(ctx, user, idx)
		}
	} else {
		if user.TwoFactorSecret != nil {
			secret, err := utils.Decrypt(*user.TwoFactorSecret, s.totpManager.GetAESKey())
			if err == nil {
				isValid = s.totpManager.ValidateCode(secret, req.Code)
			}
		}
	}

	if !isValid {
		return nil, errors.New("invalid verification code")
	}

	_ = s.tokenRepo.Delete2FATempToken(ctx, req.TempToken)
	_ = s.tokenRepo.Reset2FARateLimit(ctx, userID)

	return s.completeLogin(ctx, user, ip, userAgent)
}

// ============================================================================
// Token Management (Refresh & Complete Login)
// ============================================================================

// RefreshAccessToken generates new access token from refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	valid, err := s.sessionMgr.ValidateSession(ctx, claims.UserID, claims.SessionID)
	if err != nil || !valid {
		s.logger.Warn().Str("user_id", claims.UserID.String()).Msg("refresh on revoked session")
		return "", errors.New("session expired or revoked")
	}

	accessToken, err := s.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate access token")
		return "", errors.New("internal error")
	}

	if err := s.sessionMgr.UpdateSessionActivity(ctx, claims.UserID, claims.SessionID); err != nil {
		s.logger.Warn().Err(err).Msg("failed to extend session ttl")
	}

	return accessToken, nil
}

// completeLogin generates tokens and creates session
// UPDATED: Now passes CompanyID to token and session creation
func (s *AuthService) completeLogin(ctx context.Context, user *models.User, ip, userAgent string) (*dto.LoginResponse, error) {
	// Generate tokens with CompanyID
	tokenPair, err := s.jwtManager.GenerateTokenPair(
		user.ID,
		user.CompanyID, // NEW: Pass company context
		user.Email,
		string(user.Role),
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("token generation failed")
		return nil, errors.New("internal error")
	}

	// Create session with CompanyID
	err = s.sessionMgr.CreateSession(
		ctx,
		user.ID,
		user.CompanyID, // NEW: Pass company context
		user.Email,
		string(user.Role),
		tokenPair.SessionID,
		ip,
		userAgent,
		s.config.JWT.RefreshTokenDuration,
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("session creation failed")
		return nil, errors.New("login failed")
	}

	// Update last login asynchronously
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.userRepo.UpdateLastLogin(bgCtx, user.ID, ip)
	}()

	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         user.ToPublic(), // Already has CompanyID
		Require2FA:   false,
	}, nil
}

// ============================================================================
// Logout & Session Management
// ============================================================================

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, sessionID string) error {
	if err := s.sessionMgr.DeleteSession(ctx, userID, sessionID); err != nil {
		s.logger.Error().Err(err).Str("session_id", sessionID).Msg("failed to delete session")
		return fmt.Errorf("logout failed: %w", err)
	}
	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.sessionMgr.DeleteAllUserSessions(ctx, userID); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to delete all sessions")
		return fmt.Errorf("logout all failed: %w", err)
	}
	s.logger.Info().Str("user_id", userID.String()).Msg("all sessions invalidated")
	return nil
}

// ListSessions returns all active sessions for user
// UPDATED: Now includes CompanyID in response
func (s *AuthService) ListSessions(ctx context.Context, userID uuid.UUID) (*dto.ListSessionsResponse, error) {
	sessions, err := s.sessionMgr.ListUserSessions(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to list sessions")
		return nil, fmt.Errorf("failed to retrieve sessions: %w", err)
	}

	// Optimization: Pre-allocate slice
	response := make([]*dto.SessionResponse, len(sessions))
	for i, session := range sessions {
		response[i] = &dto.SessionResponse{
			SessionID:    session.SessionID,
			CompanyID:    session.CompanyID, // NEW: Include company context
			DeviceInfo:   session.DeviceInfo,
			IPAddress:    session.IPAddress,
			UserAgent:    session.UserAgent,
			CreatedAt:    session.CreatedAt,
			LastActivity: session.LastActivity,
		}
	}

	return &dto.ListSessionsResponse{
		Sessions: response,
		Total:    len(response),
	}, nil
}

func (s *AuthService) DeleteSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	return s.sessionMgr.DeleteSession(ctx, userID, sessionID)
}

// ============================================================================
// 2FA Management (Setup, Enable, Disable)
// ============================================================================

func (s *AuthService) Setup2FA(ctx context.Context, userID uuid.UUID) (*dto.Setup2FAResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.TwoFactorEnabled {
		return nil, errors.New("2FA is already enabled")
	}

	key, err := s.totpManager.GenerateSecret(user.Email)
	if err != nil {
		s.logger.Error().Err(err).Msg("totp secret generation failed")
		return nil, errors.New("internal error")
	}

	// Store secret temporarily in Redis (10 minutes)
	err = s.tokenRepo.Store2FASetupSecret(ctx, userID, key.Secret(), 10*time.Minute)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to store setup secret")
		return nil, errors.New("internal error")
	}

	return &dto.Setup2FAResponse{
		Secret:    key.Secret(),
		QRCodeURL: key.URL(),
	}, nil
}

func (s *AuthService) Enable2FA(ctx context.Context, userID uuid.UUID, req *dto.Enable2FARequest) (*dto.Enable2FAResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Retrieve secret from Redis
	secret, err := s.tokenRepo.Get2FASetupSecret(ctx, userID)
	if err != nil {
		return nil, errors.New("setup session expired, please try again")
	}

	// Validate Code
	if !s.totpManager.ValidateCode(secret, req.Code) {
		return nil, errors.New("invalid verification code")
	}

	// Generate Recovery Codes
	recoveryCodes, err := s.totpManager.GenerateRecoveryCodes(10)
	if err != nil {
		return nil, errors.New("failed to generate recovery codes")
	}

	// Encrypt Secret
	encryptedSecret, err := utils.Encrypt(secret, s.totpManager.GetAESKey())
	if err != nil {
		s.logger.Error().Err(err).Msg("encryption failed")
		return nil, errors.New("internal error")
	}

	// Update User
	user.Enable2FA(encryptedSecret, recoveryCodes)
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error().Err(err).Msg("db update failed for 2fa enable")
		return nil, errors.New("internal error")
	}

	// Cleanup Redis
	_ = s.tokenRepo.Delete2FASetupSecret(ctx, userID)

	s.logger.Info().Str("user_id", userID.String()).Msg("2FA enabled successfully")

	return &dto.Enable2FAResponse{
		Success:       true,
		Message:       "2FA enabled successfully",
		RecoveryCodes: recoveryCodes,
	}, nil
}

func (s *AuthService) Disable2FA(ctx context.Context, userID uuid.UUID, req *dto.Disable2FARequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !user.CheckPassword(req.Password) {
		return errors.New("incorrect password")
	}

	user.Disable2FA()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error().Err(err).Msg("failed to disable 2fa in db")
		return errors.New("internal error")
	}

	s.logger.Info().Str("user_id", userID.String()).Msg("2FA disabled")
	return nil
}

// ============================================================================
// Password Management (Forgot, Reset, Change)
// ============================================================================

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))

	// Silent fail checks
	limited, _ := s.tokenRepo.CheckEmailRateLimit(ctx, email, 3, time.Hour)
	if limited {
		return nil
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil // Silent fail
	}

	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		return errors.New("internal error")
	}

	if err := s.tokenRepo.StoreResetToken(ctx, user.ID, token, time.Hour); err != nil {
		s.logger.Error().Err(err).Msg("failed to store reset token")
		return errors.New("internal error")
	}

	s.sendPasswordResetEmailAsync(user, token)
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	userID, err := s.tokenRepo.GetResetToken(ctx, req.Token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		return errors.New("failed to hash password")
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, user.PasswordHash); err != nil {
		return errors.New("internal error")
	}

	_ = s.tokenRepo.DeleteResetToken(ctx, req.Token)
	_ = s.sessionMgr.DeleteAllUserSessions(ctx, user.ID)

	s.logger.Info().Str("user_id", user.ID.String()).Msg("password reset success")
	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, req *dto.ChangePasswordRequest) error {
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return errors.New("current password is incorrect")
	}

	if user.CheckPassword(req.NewPassword) {
		return errors.New("new password must be different")
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		return errors.New("failed to hash password")
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, user.PasswordHash); err != nil {
		s.logger.Error().Err(err).Msg("failed to update password in db")
		return errors.New("internal error")
	}

	s.logger.Info().Str("user_id", userID.String()).Msg("password changed successfully")
	return nil
}

// ============================================================================
// Email Verification Helpers
// ============================================================================

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	userID, err := s.tokenRepo.GetVerificationToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired verification token")
	}

	if err := s.userRepo.MarkEmailAsVerified(ctx, userID); err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to mark email verified")
		return errors.New("internal error")
	}

	_ = s.tokenRepo.DeleteVerificationToken(ctx, token)
	return nil
}

func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	limited, _ := s.tokenRepo.CheckEmailRateLimit(ctx, email, 3, time.Hour)
	if limited {
		return errors.New("too many emails sent, please wait")
	}

	s.sendVerificationEmailAsync(user)
	return nil
}

// ============================================================================
// Internal Async Workers
// ============================================================================

func (s *AuthService) sendVerificationEmailAsync(user *models.User) {
	email := user.Email
	userID := user.ID
	name := user.FullName

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		token, err := utils.GenerateRandomToken(32)
		if err != nil {
			s.logger.Error().Err(err).Msg("async: generate token failed")
			return
		}

		if err := s.tokenRepo.StoreVerificationToken(ctx, userID, token, 24*time.Hour); err != nil {
			s.logger.Error().Err(err).Msg("async: store token failed")
			return
		}

		if err := s.emailService.SendVerificationEmail(ctx, email, token, s.config.Frontend.VerificationUrl, name); err != nil {
			s.logger.Error().Err(err).Str("email", email).Msg("async: send email failed")
		}
	}()
}

func (s *AuthService) sendPasswordResetEmailAsync(user *models.User, token string) {
	email := user.Email
	name := user.FullName

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := s.emailService.SendPasswordResetEmail(ctx, email, token, s.config.Frontend.ResetUrl, name); err != nil {
			s.logger.Error().Err(err).Str("email", email).Msg("async: send reset email failed")
		}
	}()
}

func (s *AuthService) burnRecoveryCode(ctx context.Context, user *models.User, indexToRemove int) {
	newCodes := make([]string, 0, len(user.TwoFactorRecoveryCodes)-1)
	for i, code := range user.TwoFactorRecoveryCodes {
		if i != indexToRemove {
			newCodes = append(newCodes, code)
		}
	}
	user.TwoFactorRecoveryCodes = newCodes

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error().Err(err).Str("user_id", user.ID.String()).Msg("CRITICAL: failed to burn recovery code")
	}
}