package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

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

type AuthService struct {
	userRepo         *repository.UserRepository
	organizationRepo *repository.OrganizationRepository
	planRepo         *repository.SubscriptionPlanRepository 
	tokenRepo        *repository.TokenRepository
	sessionMgr       *utils.Manager
	emailService     *utils.EmailService
	jwtManager       *auth.TokenManager
	totpMgr             *utils.TOTPManager
	cfg              *config.Config
	log              *logger.Logger
}

func NewAuthService(
	userRepo *repository.UserRepository,
	organizationRepo *repository.OrganizationRepository,
	planRepo *repository.SubscriptionPlanRepository,
	tokenRepo *repository.TokenRepository,
	sessionMgr *utils.Manager,
	emailService *utils.EmailService,
	jwtManager *auth.TokenManager,
	totpMgr *utils.TOTPManager,
	cfg *config.Config,
	log *logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		organizationRepo: organizationRepo,
		planRepo:         planRepo,
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
// REGISTER
// ============================================================================

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// Validate inputs
	if err := sharedUtils.ValidateEmail(req.Email); err != nil {
		return nil, errors.BadRequest("Invalid email format")
	}

	if err := sharedUtils.ValidatePasswordSimple(req.Password); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := sharedUtils.ValidateLength(req.FullName, 2, 255, "Full name"); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// Parse and validate plan selection
	planTier := models.PlanDeveloper // Default to free plan
	billingPeriod := models.BillingMonthly

	if req.PlanTier != nil {
		switch *req.PlanTier {
		case "developer":
			planTier = models.PlanDeveloper
		case "starter":
			planTier = models.PlanStarter
		case "growth":
			planTier = models.PlanGrowth
		case "scale":
			planTier = models.PlanScale
		default:
			return nil, errors.BadRequest("Invalid plan tier. Must be: developer, starter, growth, or scale")
		}
	}

	// Billing period validation (only for paid plans)
	if planTier != models.PlanDeveloper && req.BillingPeriod != nil {
		switch *req.BillingPeriod {
		case "monthly":
			billingPeriod = models.BillingMonthly
		case "annual":
			billingPeriod = models.BillingAnnual
		default:
			return nil, errors.BadRequest("Invalid billing period. Must be: monthly or annual")
		}
	}

	// Fetch plan from database
	plan, err := s.planRepo.GetByTierAndBilling(ctx, planTier, billingPeriod)
	if err != nil {
		s.log.Error("Failed to fetch subscription plan", err, "tier", planTier, "billing", billingPeriod)
		return nil, errors.BadRequest("Selected plan is not available")
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, sharedUtils.NormalizeEmail(req.Email))
	if err != nil {
		s.log.Error("Failed to check email existence", err, "email", req.Email)
		return nil, errors.Internal(err)
	}

	if exists {
		return nil, errors.Conflict("Email already registered")
	}

	// Create user
	user := &models.User{
		ID:            uuid.New(),
		Email:         sharedUtils.NormalizeEmail(req.Email),
		FullName:      req.FullName,
		EmailVerified: false,
		Preferences:   models.DefaultUserPreferences(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Hash password (SetPassword internally validates and hashes)
	if err := user.SetPassword(req.Password); err != nil {
		s.log.Error("Failed to hash password", err)
		return nil, errors.Internal(err)
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.log.Error("Failed to create user", err, "email", req.Email)
		return nil, errors.Internal(err)
	}

	// Create organization with selected plan limits
	orgSlug := sharedUtils.GenerateSlug(fmt.Sprintf("%s-workspace", user.FullName))
	
	// Ensure slug is unique
	slugExists, _ := s.organizationRepo.ExistsBySlug(ctx, orgSlug)
	if slugExists {
		orgSlug = fmt.Sprintf("%s-%s", orgSlug, uuid.New().String()[:8])
	}

	organization := &models.Organization{
		ID:                 uuid.New(),
		Name:               fmt.Sprintf("%s's Workspace", user.FullName),
		Slug:               orgSlug,
		OwnerID:            user.ID,
		PlanTier:           plan.Tier,
		BillingPeriod:      plan.BillingPeriod,
		SubscriptionStatus: models.SubscriptionActive,
		
		// Plan limits from database (subscription_plans table)
		MonthlyEventsLimit: plan.MonthlyEventsLimit,
		MonthlyEventsUsed:  0,
		RetentionDays:      plan.RetentionDays,
		TeamSeatsLimit:     plan.TeamSeatsLimit,
		RateLimitPerSecond: plan.RateLimitPerSecond,
		ProjectsLimit:      plan.ProjectsLimit,
		
		// Initialize metadata
		Metadata:  models.OrganizationMetadata{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// For paid plans, set billing period dates
	if !plan.IsFreePlan() {
		now := time.Now()
		periodEnd := now.AddDate(0, 1, 0) // +1 month
		if plan.IsAnnualPlan() {
			periodEnd = now.AddDate(1, 0, 0) // +1 year
		}
		organization.CurrentPeriodStart = &now
		organization.CurrentPeriodEnd = &periodEnd
		
		// Note: LemonSqueezy customer_id and subscription_id will be set later via webhook
		// when actual payment is completed
	}

	if err := s.organizationRepo.Create(ctx, organization); err != nil {
		s.log.Error("Failed to create organization", err, "user_id", user.ID)
		// Don't fail registration if org creation fails - user can create org later
		organization = nil
	}

	// Create organization membership (owner)
	if organization != nil && organization.ID != uuid.Nil {
		member := &models.OrganizationMember{
			ID:             uuid.New(),
			OrganizationID: organization.ID,
			UserID:         user.ID,
			Role:           models.RoleOwner,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := s.organizationRepo.CreateMembership(ctx, member); err != nil {
			s.log.Error("Failed to create organization membership", err)
		}
	}

	// Send verification email asynchronously
	s.sendVerificationEmailAsync(user)

	s.log.Info("User registered successfully", 
		"user_id", user.ID, 
		"email", user.Email,
		"plan", plan.Tier,
		"billing", plan.BillingPeriod,
		"plan_name", plan.Name,
	)

	return &dto.RegisterResponse{
		Message:      "Registration successful! Please check your email to verify your account.",
	}, nil
}

// ============================================================================
// LOGIN
// ============================================================================

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
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

func (s *AuthService) Verify2FA(ctx context.Context, req *dto.TwoFactorVerifyRequest) (*dto.TwoFactorVerifyResponse, error) {
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
// REFRESH TOKEN
// ============================================================================

func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		s.log.Warn("Invalid refresh token", "error", err)
		return nil, errors.Unauthorized("Invalid or expired refresh token")
	}

	// Check if session is still valid (not revoked)
	valid := s.sessionMgr.IsSessionValid(ctx, claims.UserID, claims.SessionID)
	if !valid {
		s.log.Warn("Session has been revoked", "user_id", claims.UserID, "session_id", claims.SessionID)
		return nil, errors.Unauthorized("Session has been revoked. Please login again")
	}

	// Generate new access token (preserves SessionID)
	newAccessToken, err := s.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		s.log.Error("Failed to refresh access token", err, "user_id", claims.UserID)
		return nil, errors.Internal(err)
	}

	// Update session last activity time
	_ = s.sessionMgr.UpdateActivity(ctx, claims.UserID, claims.SessionID)

	s.log.Debug("Access token refreshed",
		"user_id", claims.UserID,
		"session_id", claims.SessionID,
	)

	return &dto.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: req.RefreshToken, // Same refresh token
		ExpiresIn:    int64(s.cfg.JWT.AccessTokenDuration.Seconds()),
	}, nil
}

// ============================================================================
// VERIFY EMAIL
// ============================================================================

func (s *AuthService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.VerifyEmailResponse, error) {
	// Get user ID from token
	userID, err := s.tokenRepo.GetVerificationToken(ctx, req.Token)
	if err != nil {
		return nil, errors.BadRequest("Invalid or expired verification token")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user", err, "user_id", userID)
		return nil, errors.NotFound("User not found")
	}

	// Check if already verified
	if user.EmailVerified {
		return &dto.VerifyEmailResponse{
			Message: "Email already verified",
		}, nil
	}

	// Mark email as verified
	if err := s.userRepo.MarkEmailVerified(ctx, userID); err != nil {
		s.log.Error("Failed to mark email verified", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	// Delete verification token
	_ = s.tokenRepo.DeleteVerificationToken(ctx, req.Token)

	s.log.Info("Email verified successfully", "user_id", userID)

	return &dto.VerifyEmailResponse{
		Message: "Email verified successfully! You can now log in.",
	}, nil
}

// ============================================================================
// RESEND VERIFICATION EMAIL
// ============================================================================

func (s *AuthService) ResendVerificationEmail(ctx context.Context, req *dto.ResendVerificationRequest) (*dto.ResendVerificationResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, sharedUtils.NormalizeEmail(req.Email))
	if err != nil {
		// Don't reveal if email exists
		return &dto.ResendVerificationResponse{
			Message: "If the email exists, a verification link has been sent.",
		}, nil
	}

	// Check if already verified
	if user.EmailVerified {
		return &dto.ResendVerificationResponse{
			Message: "Email is already verified.",
		}, nil
	}

	// Send verification email
	s.sendVerificationEmailAsync(user)

	return &dto.ResendVerificationResponse{
		Message: "If the email exists, a verification link has been sent.",
	}, nil
}

// ============================================================================
// FORGOT PASSWORD
// ============================================================================

func (s *AuthService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) (*dto.ForgotPasswordResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, sharedUtils.NormalizeEmail(req.Email))
	if err != nil {
		// Don't reveal if email exists (security)
		return &dto.ForgotPasswordResponse{
			Message: "If the email exists, a password reset link has been sent.",
		}, nil
	}

	// Generate reset token
	token, err := sharedUtils.GenerateRandomToken(32)
	if err != nil {
		s.log.Error("Failed to generate reset token", err)
		return nil, errors.Internal(err)
	}

	// Store reset token (valid for 1 hour)
	if err := s.tokenRepo.StoreResetToken(ctx, user.ID, token, 1*time.Hour); err != nil {
		s.log.Error("Failed to store reset token", err)
		return nil, errors.Internal(err)
	}

	// Send password reset email asynchronously
	s.sendPasswordResetEmailAsync(user, token)

	return &dto.ForgotPasswordResponse{
		Message: "If the email exists, a password reset link has been sent.",
	}, nil
}

// ============================================================================
// RESET PASSWORD
// ============================================================================

func (s *AuthService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) (*dto.ResetPasswordResponse, error) {
	// Validate new password
	if err := sharedUtils.ValidatePasswordSimple(req.NewPassword); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// Get user ID from token
	userID, err := s.tokenRepo.GetResetToken(ctx, req.Token)
	if err != nil {
		return nil, errors.BadRequest("Invalid or expired reset token")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user", err, "user_id", userID)
		return nil, errors.NotFound("User not found")
	}

	// Check if email is verified
	if !user.EmailVerified {
		return nil, errors.BadRequest("Email is not verified")
	}

	// Hash new password
	passwordHash, err := sharedUtils.HashPassword(req.NewPassword)
	if err != nil {
		s.log.Error("Failed to hash password", err)
		return nil, errors.Internal(err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, passwordHash); err != nil {
		s.log.Error("Failed to update password", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	// Delete reset token
	_ = s.tokenRepo.DeleteResetToken(ctx, req.Token)

	// Revoke all sessions (force re-login)
	_ = s.sessionMgr.RevokeAllSessions(ctx, userID)

	s.log.Info("Password reset successfully", "user_id", userID)

	return &dto.ResetPasswordResponse{
		Message: "Password reset successfully! Please log in with your new password.",
	}, nil
}

// ============================================================================
// CHANGE PASSWORD
// ============================================================================

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, req *dto.ChangePasswordRequest) (*dto.ChangePasswordResponse, error) {
	// Validate new password
	if err := sharedUtils.ValidatePasswordSimple(req.NewPassword); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user", err, "user_id", userID)
		return nil, errors.NotFound("User not found")
	}

	// Verify current password
	if !user.CheckPassword(req.CurrentPassword) {
		return nil, errors.BadRequest("Current password is incorrect")
	}

	// Check if new password is same as current
	if user.CheckPassword(req.NewPassword) {
		return nil, errors.BadRequest("New password must be different from current password")
	}

	// Hash new password
	passwordHash, err := sharedUtils.HashPassword(req.NewPassword)
	if err != nil {
		s.log.Error("Failed to hash password", err)
		return nil, errors.Internal(err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, passwordHash); err != nil {
		s.log.Error("Failed to update password", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	s.log.Info("Password changed successfully", "user_id", userID)

	return &dto.ChangePasswordResponse{
		Message: "Password changed successfully!",
	}, nil
}

// ============================================================================
// LOGOUT
// ============================================================================

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, sessionID string, req *dto.LogoutRequest) (*dto.LogoutResponse, error) {
	if req.AllDevices {
		// Logout all sessions
		if err := s.sessionMgr.RevokeAllSessions(ctx, userID); err != nil {
			s.log.Error("Failed to revoke all sessions", err, "user_id", userID)
			return nil, errors.Internal(err)
		}

		s.log.Info("User logged out from all devices", "user_id", userID)

		return &dto.LogoutResponse{
			Message: "Logged out from all devices successfully",
		}, nil
	}

	// Logout current session only
	if err := s.sessionMgr.RevokeSession(ctx, userID, sessionID); err != nil {
		s.log.Error("Failed to revoke session", err, "user_id", userID, "session_id", sessionID)
		return nil, errors.Internal(err)
	}

	s.log.Info("User logged out", "user_id", userID, "session_id", sessionID)

	return &dto.LogoutResponse{
		Message: "Logged out successfully",
	}, nil
}

// ============================================================================
// GET ACTIVE SESSIONS
// ============================================================================

func (s *AuthService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]utils.SessionData, error) {
	sessions, err := s.sessionMgr.GetAllSessions(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get active sessions", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	return sessions, nil
}

// ============================================================================
// REVOKE SESSION BY ID
// ============================================================================

func (s *AuthService) RevokeSessionByID(ctx context.Context, userID uuid.UUID, sessionID string) error {
	if err := s.sessionMgr.RevokeSession(ctx, userID, sessionID); err != nil {
		s.log.Error("Failed to revoke session", err, "user_id", userID, "session_id", sessionID)
		return errors.Internal(err)
	}

	s.log.Info("Session revoked", "user_id", userID, "session_id", sessionID)
	return nil
}

// ============================================================================
// ASYNC EMAIL HELPERS
// ============================================================================

func (s *AuthService) sendVerificationEmailAsync(user *models.User) {
	email := user.Email
	userID := user.ID
	name := user.FullName

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		token, err := sharedUtils.GenerateRandomToken(32)
		if err != nil {
			s.log.Error("async: generate verification token failed", err)
			return
		}

		if err := s.tokenRepo.StoreVerificationToken(ctx, userID, token, 24*time.Hour); err != nil {
			s.log.Error("async: store verification token failed", err)
			return
		}

		if err := s.emailService.SendVerificationEmail(ctx, email, token, s.cfg.Frontend.VerificationUrl, name); err != nil {
			s.log.Error("async: send verification email failed", err)
		}
	}()
}

func (s *AuthService) sendPasswordResetEmailAsync(user *models.User, token string) {
	email := user.Email
	name := user.FullName

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := s.emailService.SendPasswordResetEmail(ctx, email, token, s.cfg.Frontend.ResetUrl, name); err != nil {
			s.log.Error("async: send password reset email failed", err)
		}
	}()
}


// ============================================================================
// HELPER: Remove Used Recovery Code
// ============================================================================

func (s *AuthService) removeUsedRecoveryCode(ctx context.Context, user *models.User, usedIndex int) error {
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