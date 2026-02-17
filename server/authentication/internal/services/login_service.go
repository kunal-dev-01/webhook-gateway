package services

import (
	"context"
	"fmt"
	"time"

	"authentication/internal/dto"
	"authentication/internal/repository"
	"authentication/internal/utils"

	"shared/auth"
	"shared/config"
	"shared/errors"
	"shared/logger"
	"shared/models"
	sharedUtils "shared/utils"
)

type LoginService struct {
	userRepo         *repository.UserRepository
	organizationRepo *repository.OrganizationRepository
	tokenRepo        *repository.TokenRepository
	sessionMgr       *utils.Manager
	emailService     *utils.EmailService
	jwtManager       *auth.TokenManager
	totpMgr          *utils.TOTPManager
	cfg              *config.Config
	log              *logger.Logger
}

func NewLoginService(
	userRepo *repository.UserRepository,
	organizationRepo *repository.OrganizationRepository,
	tokenRepo *repository.TokenRepository,
	sessionMgr *utils.Manager,
	emailService *utils.EmailService,
	jwtManager *auth.TokenManager,
	totpMgr *utils.TOTPManager,
	cfg *config.Config,
	log *logger.Logger,
) *LoginService {
	return &LoginService{
		userRepo:         userRepo,
		organizationRepo: organizationRepo,
		tokenRepo:        tokenRepo,
		sessionMgr:       sessionMgr,
		emailService:     emailService,
		jwtManager:       jwtManager,
		totpMgr:          totpMgr,
		cfg:              cfg,
		log:              log,
	}
}

// ============================================================================
// LOGIN
// ============================================================================

func (s *LoginService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Validate inputs
	if err := sharedUtils.ValidateEmail(req.Email); err != nil {
		return nil, errors.BadRequest("Invalid email format")
	}

	if req.Password == "" {
		return nil, errors.BadRequest("Password is required")
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, sharedUtils.NormalizeEmail(req.Email))
	if err != nil {
		// Don't reveal if email exists (security)
		s.log.Warn("Login attempt with non-existent email", "email", req.Email)
		return nil, errors.Unauthorized("Invalid email or password")
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		s.log.Warn("Login attempt with invalid password", "user_id", user.ID, "email", user.Email)
		return nil, errors.Unauthorized("Invalid email or password")
	}

	// Check if email is verified
	if !user.EmailVerified {
		s.log.Warn("Login attempt with unverified email", "user_id", user.ID, "email", user.Email)
		return nil, errors.BadRequest("Email is not verified. Please check your inbox for verification link")
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		// Generate temporary 2FA token (valid for 5 minutes)
		tempToken, err := sharedUtils.GenerateRandomToken(32)
		if err != nil {
			s.log.Error("Failed to generate 2FA temp token", err, "user_id", user.ID)
			return nil, errors.Internal(err)
		}

		// Store temp token
		if err := s.tokenRepo.Store2FATempToken(ctx, user.ID, tempToken, 5*time.Minute); err != nil {
			s.log.Error("Failed to store 2FA temp token", err, "user_id", user.ID)
			return nil, errors.Internal(err)
		}

		s.log.Info("2FA required for login", "user_id", user.ID, "email", user.Email)

		return &dto.LoginResponse{
			TwoFactorRequired: true,
			TempToken:         tempToken,
		}, nil
	}

	// 2FA not enabled - proceed with normal login
	// Get user's primary organization (the one they own or most recently accessed)
	orgs, err := s.organizationRepo.ListByUser(ctx, user.ID)
	if err != nil || len(orgs) == 0 {
		s.log.Error("No organizations found for user", err, "user_id", user.ID)
		return nil, errors.Internal(fmt.Errorf("no organization found for user"))
	}

	// Use first organization (primary organization)
	primaryOrg := orgs[0]

	// Get user's role in this organization
	membership, err := s.organizationRepo.GetMembership(ctx, primaryOrg.ID, user.ID)
	if err != nil {
		s.log.Error("Failed to get organization membership", err, "user_id", user.ID, "org_id", primaryOrg.ID)
		return nil, errors.Internal(err)
	}

	// Generate JWT token pair
	tokenPair, err := s.jwtManager.GenerateTokenPair(
		user.ID,
		primaryOrg.ID,
		user.Email,
		string(membership.Role),
	)
	if err != nil {
		s.log.Error("Failed to generate token pair", err, "user_id", user.ID)
		return nil, errors.Internal(err)
	}

	// Store session in Redis
	if err := s.sessionMgr.CreateSession(ctx, &utils.SessionData{
		UserID:         user.ID,
		Email:          user.Email,
		OrganizationID: primaryOrg.ID,
		OrgRole:        string(membership.Role),
		SessionID:      tokenPair.SessionID,
		DeviceInfo:     "", // Will be set by handler
		IPAddress:      "", // Will be set by handler
		UserAgent:      "", // Will be set by handler
		CreatedAt:      time.Now(),
		LastActivity:   time.Now(),
	}); err != nil {
		s.log.Error("Failed to create session", err, "user_id", user.ID)
		return nil, errors.Internal(err)
	}

	s.log.Info("User logged in successfully",
		"user_id", user.ID,
		"email", user.Email,
		"org_id", primaryOrg.ID,
		"session_id", tokenPair.SessionID,
	)

	return &dto.LoginResponse{
		User:              dto.ToUserResponse(user),
		AccessToken:       tokenPair.AccessToken,
		RefreshToken:      tokenPair.RefreshToken,
		ExpiresIn:         tokenPair.ExpiresIn,
		TwoFactorRequired: false,
	}, nil
}

// ============================================================================
// VERIFY 2FA (Login completion with TOTP or Recovery Code)
// ============================================================================

func (s *LoginService) Verify2FA(ctx context.Context, req *dto.TwoFactorVerifyRequest) (*dto.TwoFactorVerifyResponse, error) {
	// Get user ID from temp token
	userID, err := s.tokenRepo.Get2FATempToken(ctx, req.TempToken)
	if err != nil {
		s.log.Warn("Invalid or expired 2FA temp token")
		return nil, errors.BadRequest("Invalid or expired 2FA token. Please login again")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user for 2FA verification", err, "user_id", userID)
		return nil, errors.NotFound("User not found")
	}

	// Verify 2FA is enabled
	if !user.TwoFactorEnabled || user.TwoFactorSecret == nil {
		s.log.Error("2FA not enabled for user during verification", nil, "user_id", userID)
		return nil, errors.BadRequest("Two-factor authentication is not enabled")
	}

	// Verify code based on type (TOTP or Recovery)
	verified := false

	if req.IsRecovery {
		// Using recovery code
		s.log.Info("2FA verification with recovery code attempted", "user_id", userID)

		valid, codeIndex, err := s.totpMgr.ValidateRecoveryCode(req.Code, user.TwoFactorBackupCodes)
		if err != nil {
			s.log.Error("Failed to validate recovery code", err, "user_id", userID)
			return nil, errors.Internal(err)
		}

		if !valid {
			s.log.Warn("Invalid recovery code during 2FA verification", "user_id", userID)
			return nil, errors.BadRequest("Invalid recovery code")
		}

		// Check if code was already used
		used, err := s.tokenRepo.IsRecoveryCodeUsed(ctx, userID, codeIndex)
		if err != nil {
			s.log.Error("Failed to check recovery code usage", err, "user_id", userID)
			return nil, errors.Internal(err)
		}

		if used {
			s.log.Warn("Attempted to use already-used recovery code", "user_id", userID)
			return nil, errors.BadRequest("This recovery code has already been used")
		}

		// Mark code as used (24 hour TTL)
		if err := s.tokenRepo.MarkRecoveryCodeUsed(ctx, userID, codeIndex, 24*3600); err != nil {
			s.log.Error("Failed to mark recovery code as used", err, "user_id", userID)
			// Don't fail the request, just log
		}

		// Remove used recovery code from database
		if err := s.removeUsedRecoveryCode(ctx, user, codeIndex); err != nil {
			s.log.Error("Failed to remove used recovery code from database", err, "user_id", userID)
			// Don't fail the request, just log
		}

		verified = true
		s.log.Info("2FA verification with recovery code successful", "user_id", userID)

	} else {
		// Using TOTP code
		valid, err := s.totpMgr.Validate(*user.TwoFactorSecret, req.Code)
		if err != nil {
			s.log.Error("Failed to validate TOTP code", err, "user_id", userID)
			return nil, errors.Internal(err)
		}

		if !valid {
			s.log.Warn("Invalid TOTP code during 2FA verification", "user_id", userID)
			return nil, errors.BadRequest("Invalid verification code")
		}

		verified = true
		s.log.Info("2FA verification with TOTP successful", "user_id", userID)
	}

	if !verified {
		return nil, errors.BadRequest("Invalid verification code")
	}

	// Delete temp token
	_ = s.tokenRepo.Delete2FATempToken(ctx, req.TempToken)

	// Get user's primary organization
	orgs, err := s.organizationRepo.ListByUser(ctx, userID)
	if err != nil || len(orgs) == 0 {
		s.log.Error("No organizations found for user", err, "user_id", userID)
		return nil, errors.Internal(fmt.Errorf("no organization found for user"))
	}

	primaryOrg := orgs[0]

	// Get user's role in this organization
	membership, err := s.organizationRepo.GetMembership(ctx, primaryOrg.ID, userID)
	if err != nil {
		s.log.Error("Failed to get organization membership", err, "user_id", userID, "org_id", primaryOrg.ID)
		return nil, errors.Internal(err)
	}

	// Generate JWT token pair
	tokenPair, err := s.jwtManager.GenerateTokenPair(
		user.ID,
		primaryOrg.ID,
		user.Email,
		string(membership.Role),
	)
	if err != nil {
		s.log.Error("Failed to generate token pair", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	// Store session in Redis
	if err := s.sessionMgr.CreateSession(ctx, &utils.SessionData{
		UserID:         userID,
		Email:          user.Email,
		OrganizationID: primaryOrg.ID,
		OrgRole:        string(membership.Role),
		SessionID:      tokenPair.SessionID,
		DeviceInfo:     "", // Will be set by handler
		IPAddress:      "", // Will be set by handler
		UserAgent:      "", // Will be set by handler
		CreatedAt:      time.Now(),
		LastActivity:   time.Now(),
	}); err != nil {
		s.log.Error("Failed to create session", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	s.log.Info("2FA verification completed, user logged in",
		"user_id", userID,
		"email", user.Email,
		"org_id", primaryOrg.ID,
		"session_id", tokenPair.SessionID,
		"used_recovery", req.IsRecovery,
	)

	return &dto.TwoFactorVerifyResponse{
		User:         dto.ToUserResponse(user),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// ============================================================================
// HELPER: Remove Used Recovery Code
// ============================================================================

func (s *LoginService) removeUsedRecoveryCode(ctx context.Context, user *models.User, usedIndex int) error {
	// Create new slice without the used code
	newCodes := make([]string, 0, len(user.TwoFactorBackupCodes)-1)
	for i, code := range user.TwoFactorBackupCodes {
		if i != usedIndex {
			newCodes = append(newCodes, code)
		}
	}

	// Update user with new backup codes (removing the used one)
	if err := s.userRepo.Enable2FA(ctx, user.ID, *user.TwoFactorSecret, newCodes); err != nil {
		return err
	}

	s.log.Info("Removed used recovery code", "user_id", user.ID, "remaining_codes", len(newCodes))
	return nil
}
