package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"authentication/internal/dto"
	"authentication/internal/repository"
	"authentication/internal/utils"
	"authentication/internal/config"

	"shared/auth"
	"shared/errors"
	"shared/logger"
	"shared/models"
	sharedUtils "shared/utils"
)

type RegisterService struct {
	userRepo         *repository.UserRepository
	organizationRepo *repository.OrganizationRepository
	planRepo         *repository.SubscriptionPlanRepository
	tokenRepo        *repository.TokenRepository
	sessionMgr       *utils.Manager
	emailService     *utils.EmailService
	jwtManager       *auth.TokenManager
	totpMgr          *utils.TOTPManager
	cfg              *config.Config
	log              *logger.Logger
}

func NewRegisterService(
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
) *RegisterService {
	return &RegisterService{
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

func (s *RegisterService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
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
	orgSlug := fmt.Sprintf("%s-%s", sharedUtils.GenerateSlug(fmt.Sprintf("%s-workspace", user.FullName)), uuid.New().String()[:8])

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
		Message: "Registration successful! Please check your email to verify your account.",
	}, nil
}

// ============================================================================
// VERIFY EMAIL
// ============================================================================

func (s *RegisterService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.VerifyEmailResponse, error) {
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

func (s *RegisterService) ResendVerificationEmail(ctx context.Context, req *dto.ResendVerificationRequest) (*dto.ResendVerificationResponse, error) {
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
// ASYNC EMAIL HELPERS
// ============================================================================

func (s *RegisterService) sendVerificationEmailAsync(user *models.User) {
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

