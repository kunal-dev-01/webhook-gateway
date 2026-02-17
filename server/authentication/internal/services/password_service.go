package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"authentication/internal/dto"
	"authentication/internal/repository"
	"authentication/internal/utils"

	"shared/config"
	"shared/errors"
	"shared/logger"
	"shared/models"
	sharedUtils "shared/utils"
)

type PasswordService struct {
	userRepo         *repository.UserRepository
	tokenRepo        *repository.TokenRepository
	emailService     *utils.EmailService
	sessionMgr       *utils.Manager
	cfg              *config.Config
	log              *logger.Logger
}

func NewPasswordService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	emailService *utils.EmailService,
	sessionMgr       *utils.Manager,
	cfg *config.Config,
	log *logger.Logger,
) *PasswordService {
	return &PasswordService{
		userRepo:         userRepo,
		tokenRepo:        tokenRepo,
		emailService:     emailService,
		sessionMgr:       sessionMgr,
		cfg:              cfg,
		log:              log,
	}
}

// ============================================================================
// FORGOT PASSWORD
// ============================================================================

func (s *PasswordService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) (*dto.ForgotPasswordResponse, error) {
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

func (s *PasswordService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) (*dto.ResetPasswordResponse, error) {
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

func (s *PasswordService) ChangePassword(ctx context.Context, userID uuid.UUID, req *dto.ChangePasswordRequest) (*dto.ChangePasswordResponse, error) {
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
// ASYNC EMAIL HELPERS
// ============================================================================

func (s *PasswordService) sendPasswordResetEmailAsync(user *models.User, token string) {
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
