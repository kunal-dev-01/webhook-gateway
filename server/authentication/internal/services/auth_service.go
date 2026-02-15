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

	// Prepare response
	var orgResponse *dto.OrganizationResponse
	if organization != nil {
		orgResponse = dto.ToOrganizationResponse(organization, string(models.RoleOwner))
	}

	return &dto.RegisterResponse{
		Message:      "Registration successful! Please check your email to verify your account.",
		User:         dto.ToUserResponse(user),
		Organization: orgResponse,
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