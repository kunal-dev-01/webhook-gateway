package services

import (
	"context"
	"encoding/base64" 
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"

	"authentication/internal/dto"
	"authentication/internal/repository"
	"authentication/internal/utils"

	"shared/errors"
	"shared/logger"
)

type TwoFactorService struct {
	userRepo   *repository.UserRepository
	tokenRepo  *repository.TokenRepository
	totpMgr    *utils.TOTPManager
	log        *logger.Logger
}

func NewTwoFactorService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	totpMgr *utils.TOTPManager,
	log *logger.Logger,
) *TwoFactorService {
	return &TwoFactorService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		totpMgr:   totpMgr,
		log:       log,
	}
}

// ============================================================================
// ENABLE 2FA (Step 1: Generate QR Code)
// ============================================================================

func (s *TwoFactorService) Enable2FA(ctx context.Context, userID uuid.UUID) (*dto.Enable2FAResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user for 2FA setup", err, "user_id", userID)
		return nil, err
	}

	if user.TwoFactorEnabled {
		return nil, errors.BadRequest("Two-factor authentication is already enabled")
	}

	// Generate TOTP secret
	rawSecret, encryptedSecret, provisioningURI, err := s.totpMgr.Generate(user.Email)
	if err != nil {
		s.log.Error("Failed to generate TOTP secret", err, "user_id", userID)
		return nil, err
	}

	// Store encrypted secret in Redis (10 mins)
	if err := s.tokenRepo.Store2FASetupSecret(ctx, userID, encryptedSecret, 10*time.Minute); err != nil {
    s.log.Error("Failed to store 2FA setup secret", err, "user_id", userID)
    return nil, err
}

	// Generate QR code (Returns []byte)
	qrCodeBytes, err := qrcode.Encode(provisioningURI, qrcode.Medium, 256)
	if err != nil {
		s.log.Error("Failed to generate QR code", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	// FIX: Encode Bytes to Base64 String
	qrBase64 := base64.StdEncoding.EncodeToString(qrCodeBytes)
	qrCodeDataURL := fmt.Sprintf("data:image/png;base64,%s", qrBase64)

	s.log.Info("2FA setup initiated", "user_id", userID)

	return &dto.Enable2FAResponse{
		QRCodeURL: qrCodeDataURL,
		Secret:    rawSecret,
	}, nil
}

// ============================================================================
// CONFIRM 2FA (Step 2: Verify TOTP Code)
// ============================================================================

func (s *TwoFactorService) Confirm2FA(ctx context.Context, userID uuid.UUID, req *dto.Confirm2FARequest) (*dto.Confirm2FAResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user for 2FA confirmation", err, "user_id", userID)
		return nil, err
	}

	// Check if 2FA is already enabled
	if user.TwoFactorEnabled {
		return nil, errors.BadRequest("Two-factor authentication is already enabled")
	}

	// Get the temporary encrypted secret from Redis
	encryptedSecret, err := s.tokenRepo.Get2FASetupSecret(ctx, userID)
	if err != nil {
		s.log.Warn("2FA setup secret not found or expired", "user_id", userID)
		return nil, errors.BadRequest("2FA setup session expired. Please restart the setup process")
	}

	// Validate TOTP code
	valid, err := s.totpMgr.Validate(encryptedSecret, req.Code)
	if err != nil {
		s.log.Error("Failed to validate TOTP code", err, "user_id", userID)
		return nil, err
	}

	if !valid {
		s.log.Warn("Invalid TOTP code during 2FA setup", "user_id", userID)
		return nil, errors.BadRequest("Invalid verification code. Please try again")
	}

	// Generate recovery codes
	rawCodes, encryptedCodes, err := s.totpMgr.GenerateRecoveryCodes(10)
	if err != nil {
		s.log.Error("Failed to generate recovery codes", err, "user_id", userID)
		return nil, err
	}

	// Enable 2FA in database
	if err := s.userRepo.Enable2FA(ctx, userID, encryptedSecret, encryptedCodes); err != nil {
		s.log.Error("Failed to enable 2FA in database", err, "user_id", userID)
		return nil, err
	}

	// Delete the temporary setup secret
	if err := s.tokenRepo.Delete2FASetupSecret(ctx, userID); err != nil {
		s.log.Error("Failed to delete 2FA setup secret", err, "user_id", userID)
		// Don't fail the request, just log
	}

	s.log.Info("2FA successfully enabled", "user_id", userID, "email", user.Email)

	return &dto.Confirm2FAResponse{
		Message:     "Two-factor authentication enabled successfully",
		BackupCodes: rawCodes, // Show to user ONCE
	}, nil
}

// ============================================================================
// DISABLE 2FA
// ============================================================================

func (s *TwoFactorService) Disable2FA(ctx context.Context, userID uuid.UUID, req *dto.Disable2FARequest) (*dto.Disable2FAResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user for 2FA disable", err, "user_id", userID)
		return nil, err
	}

	// Check if 2FA is enabled
	if !user.TwoFactorEnabled {
		return nil, errors.BadRequest("Two-factor authentication is not enabled")
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		s.log.Warn("Invalid password during 2FA disable attempt", "user_id", userID)
		return nil, errors.Unauthorized("Invalid password")
	}

	// Verify either TOTP code OR recovery code
	verified := false

	// Try TOTP code first
	if req.Code != "" && user.TwoFactorSecret != nil {
		valid, err := s.totpMgr.Validate(*user.TwoFactorSecret, req.Code)
		if err != nil {
			s.log.Error("Failed to validate TOTP during 2FA disable", err, "user_id", userID)
			return nil, err
		}
		verified = valid
	}

	// If TOTP failed, try recovery code
	if !verified && req.RecoveryCode != "" && len(user.TwoFactorBackupCodes) > 0 {
		valid, codeIndex, err := s.totpMgr.ValidateRecoveryCode(req.RecoveryCode, user.TwoFactorBackupCodes)
		if err != nil {
			s.log.Error("Failed to validate recovery code during 2FA disable", err, "user_id", userID)
			return nil, err
		}

		if valid {
			// Check if code was already used
			used, err := s.tokenRepo.IsRecoveryCodeUsed(ctx, userID, codeIndex)
			if err != nil {
				s.log.Error("Failed to check recovery code usage", err, "user_id", userID)
				return nil, err
			}

			if used {
				s.log.Warn("Attempted to use already-used recovery code for 2FA disable", "user_id", userID)
				return nil, errors.BadRequest("This recovery code has already been used")
			}

			// Mark code as used (24 hour TTL)
			if err := s.tokenRepo.MarkRecoveryCodeUsed(ctx, userID, codeIndex, 24*3600); err != nil {
				s.log.Error("Failed to mark recovery code as used", err, "user_id", userID)
				// Don't fail the request
			}

			verified = true
		}
	}

	if !verified {
		s.log.Warn("Invalid code during 2FA disable attempt", "user_id", userID)
		return nil, errors.BadRequest("Invalid verification code or recovery code")
	}

	// Disable 2FA in database
	if err := s.userRepo.Disable2FA(ctx, userID); err != nil {
		s.log.Error("Failed to disable 2FA in database", err, "user_id", userID)
		return nil, err
	}

	s.log.Info("2FA successfully disabled", "user_id", userID, "email", user.Email)

	return &dto.Disable2FAResponse{
		Message: "Two-factor authentication disabled successfully",
	}, nil
}

// ============================================================================
// REGENERATE BACKUP CODES
// ============================================================================

func (s *TwoFactorService) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, req *dto.RegenerateBackupCodesRequest) (*dto.RegenerateBackupCodesResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user for backup codes regeneration", err, "user_id", userID)
		return nil, err
	}

	// Check if 2FA is enabled
	if !user.TwoFactorEnabled {
		return nil, errors.BadRequest("Two-factor authentication is not enabled")
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		s.log.Warn("Invalid password during backup codes regeneration", "user_id", userID)
		return nil, errors.Unauthorized("Invalid password")
	}

	// Generate new recovery codes
	rawCodes, encryptedCodes, err := s.totpMgr.GenerateRecoveryCodes(10)
	if err != nil {
		s.log.Error("Failed to generate new recovery codes", err, "user_id", userID)
		return nil, err
	}

	// Update in database (keep the same TOTP secret)
	if err := s.userRepo.Enable2FA(ctx, userID, *user.TwoFactorSecret, encryptedCodes); err != nil {
		s.log.Error("Failed to update recovery codes in database", err, "user_id", userID)
		return nil, err
	}

	s.log.Info("Backup codes regenerated", "user_id", userID, "email", user.Email)

	return &dto.RegenerateBackupCodesResponse{
		BackupCodes: rawCodes,
	}, nil
}

// ============================================================================
// GET 2FA STATUS
// ============================================================================

func (s *TwoFactorService) Get2FAStatus(ctx context.Context, userID uuid.UUID) (*dto.TwoFactorStatusResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user for 2FA status", err, "user_id", userID)
		return nil, err
	}

	return &dto.TwoFactorStatusResponse{
		Enabled:     user.TwoFactorEnabled,
		ConfirmedAt: user.TwoFactorConfirmedAt,
	}, nil
}