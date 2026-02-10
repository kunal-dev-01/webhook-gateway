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
	s.sendVerificationEmailAsync(user)

	return nil
}

// ============================================================================
// Login
// ============================================================================

// Login authenticates user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest, ip, userAgent string) (*dto.LoginResponse, error) {
	// ✅ Step 1: Check Redis rate limit for this IP (max 10 attempts per 15 minutes)
	limited, err := s.tokenRepo.CheckLoginRateLimit(ctx, ip, 10, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}
	if limited {
		return nil, fmt.Errorf("too many login attempts from this IP, please try again later")
	}

	// ✅ Step 2: Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Increment failed attempts in DB
		s.userRepo.IncrementFailedAttempts(ctx, req.Email)
		return nil, fmt.Errorf("invalid credentials")
	}

	// ✅ Step 3: Check if user can login
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

	// ✅ Step 4: Verify password
	if !user.CheckPassword(req.Password) {
		// Increment failed attempts in DB
		s.userRepo.IncrementFailedAttempts(ctx, req.Email)
		return nil, fmt.Errorf("invalid credentials")
	}

	// ✅ Step 5: Check if 2FA is enabled
	if user.TwoFactorEnabled {
		// Generate temporary token for 2FA verification (valid for 5 minutes)
		tempToken, err := utils.GenerateRandomToken(32)
		if err != nil {
			return nil, fmt.Errorf("failed to generate temp token: %w", err)
		}

		// Store temp token in Redis (5 minutes)
		if err := s.tokenRepo.Store2FATempToken(ctx, user.ID, tempToken, 5*time.Minute); err != nil {
			return nil, fmt.Errorf("failed to store temp token: %w", err)
		}

		// Return temp token - user needs to verify 2FA code next
		return &dto.LoginResponse{
			Require2FA: true,
			TempToken:  tempToken,
		}, nil
	}

	// ✅ Step 6: Complete login (no 2FA required)
	return s.completeLogin(ctx, user, ip, userAgent)
}

// ============================================================================
// Verify2FA - UPDATED WITH BETTER ERROR MESSAGES ✅
// ============================================================================

// Verify2FA verifies 2FA code (TOTP or recovery code) and completes login
// Accepts:
// - TOTP code (6 digits) from authenticator app
// - Recovery/Backup code (8 characters) from saved backup codes
func (s *AuthService) Verify2FA(ctx context.Context, req *dto.Verify2FARequest, ip, userAgent string) (*dto.LoginResponse, error) {
	// ✅ Step 1: Get user ID from temp token
	userID, err := s.tokenRepo.Get2FATempToken(ctx, req.TempToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired 2FA session - please login again")
	}

	// ✅ Step 2: Check 2FA rate limit (max 5 attempts per 5 minutes to prevent brute force)
	limitCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	limited, err := s.tokenRepo.Check2FARateLimit(limitCtx, userID, 5, 5*time.Minute)
	if err != nil {
		fmt.Printf("Rate limit check failed: %v\n", err)
	} else if limited {
		return nil, fmt.Errorf("too many 2FA verification attempts - please wait 5 minutes and try again")
	}

	// ✅ Step 3: Get user from database
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// ✅ Step 4: Verify 2FA is enabled
	if !user.TwoFactorEnabled || user.TwoFactorSecret == nil {
		return nil, fmt.Errorf("2FA is not enabled for this account")
	}

	// ✅ Step 5: Decrypt TOTP secret from database
	secret, err := utils.Decrypt(*user.TwoFactorSecret, s.totpManager.GetAESKey())
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}

	// ============================================================================
	// ✅ STRATEGY 1: Try TOTP Code First (6 digits from authenticator app)
	// ============================================================================
	if len(req.Code) == 6 {
		if s.totpManager.ValidateCode(secret, req.Code) {
			// ✅ SUCCESS: TOTP code is valid
			fmt.Printf("✅ 2FA: User %s authenticated with TOTP code\n", user.Email)

			// Reset rate limit counter
			s.tokenRepo.Reset2FARateLimit(ctx, userID)

			// Delete temp token (no longer needed)
			s.tokenRepo.Delete2FATempToken(ctx, req.TempToken)

			// Complete login
			return s.completeLogin(ctx, user, ip, userAgent)
		}
	}

	// ============================================================================
	// ✅ STRATEGY 2: Try Recovery/Backup Code (8 characters)
	// ============================================================================
	if len(user.TwoFactorRecoveryCodes) > 0 {
		valid, codeIndex := s.totpManager.ValidateRecoveryCode(req.Code, user.TwoFactorRecoveryCodes)

		if valid {
			// ✅ Check if code already used (prevent reuse attacks)
			used, err := s.tokenRepo.IsRecoveryCodeUsed(ctx, userID, codeIndex)
			if err == nil && used {
				return nil, fmt.Errorf("this recovery code has already been used - please use a different code")
			}

			// ✅ SUCCESS: Recovery code is valid
			fmt.Printf("✅ 2FA: User %s authenticated with recovery code (index %d)\n", user.Email, codeIndex)

			// Mark code as used in Redis (prevent reuse for 30 days)
			s.tokenRepo.MarkRecoveryCodeUsed(ctx, userID, codeIndex, 30*24*time.Hour)

			// Reset rate limit counter
			s.tokenRepo.Reset2FARateLimit(ctx, userID)

			// Delete temp token (no longer needed)
			s.tokenRepo.Delete2FATempToken(ctx, req.TempToken)

			// ✅ PERMANENTLY DELETE used recovery code from database
			// Copy all codes except the used one
			newRecoveryCodes := make([]string, 0, len(user.TwoFactorRecoveryCodes)-1)
			for i, code := range user.TwoFactorRecoveryCodes {
				if i != codeIndex {
					newRecoveryCodes = append(newRecoveryCodes, code)
				}
			}
			user.TwoFactorRecoveryCodes = newRecoveryCodes

			// Update user in database (removes used code permanently)
			if err := s.userRepo.Update(ctx, user); err != nil {
				// Log error but don't fail login
				fmt.Printf("⚠️ Warning: Failed to remove used recovery code: %v\n", err)
			}

			// Complete login
			return s.completeLogin(ctx, user, ip, userAgent)
		}
	}

	// ============================================================================
	// ❌ BOTH STRATEGIES FAILED
	// ============================================================================
	// Provide helpful error message based on code length
	if len(req.Code) == 6 {
		return nil, fmt.Errorf("invalid TOTP code - please check your authenticator app and try again")
	} else if len(req.Code) == 8 {
		return nil, fmt.Errorf("invalid recovery code - please check your saved backup codes")
	}
	
	return nil, fmt.Errorf("invalid 2FA code - please enter a 6-digit TOTP code or 8-character recovery code")
}

// ============================================================================
// Complete Login Helper
// ============================================================================

// completeLogin generates tokens and creates session
func (s *AuthService) completeLogin(ctx context.Context, user *models.User, ip, userAgent string) (*dto.LoginResponse, error) {
	// ✅ Step 1: Generate JWT tokens (access + refresh)
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// ✅ Step 2: Create session in Redis
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

	// ✅ Step 3: Update last login timestamp in database
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, ip); err != nil {
		// Log error but don't fail login
		fmt.Printf("⚠️ Warning: Failed to update last login: %v\n", err)
	}

	// ✅ Step 4: Return successful login response
	return &dto.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken, // For HTTP-only cookie
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

	s.sendVerificationEmailAsync(user)

	// Send verification email
	return nil
}

// sendVerificationEmail sends verification email
func (s *AuthService) sendVerificationEmailAsync(user *models.User) {
    go func() {
        ctx := context.Background()
        token, err := utils.GenerateRandomToken(32)
        if err != nil {
            fmt.Printf("Failed to generate token: %v\n", err)
            return
        }
        
        if err := s.tokenRepo.StoreVerificationToken(ctx, user.ID, token, 24*time.Hour); err != nil {
            fmt.Printf("Failed to store token: %v\n", err)
            return
        }
        
        s.emailService.SendVerificationEmail(
            ctx,
            user.Email,
            token,
            s.config.Frontend.VerificationUrl,
            user.FullName,
        )
    }()
}

// ============================================================================
// Password Reset
// ============================================================================

// ForgotPassword sends password reset email
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	// ✅ OPTIMIZED: Check rate limit FIRST (before DB query)
	limitCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	
	limited, err := s.tokenRepo.CheckEmailRateLimit(limitCtx, email, 3, time.Hour)
	if err != nil {
		// Log but don't reveal
		fmt.Printf("Rate limit check failed: %v\n", err)
	} else if limited {
		// Don't reveal rate limit - silent fail
		return nil
	}

	// Get user
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists
		return nil
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

	// ✅ OPTIMIZED: Send email asynchronously (non-blocking)
	go func() {
		bgCtx := context.Background()
		if err := s.emailService.SendPasswordResetEmail(
			bgCtx,
			user.Email,
			token,
			s.config.Frontend.ResetUrl,
			user.FullName,
		); err != nil {
			fmt.Printf("Failed to send password reset email: %v\n", err)
		}
	}()

	return nil
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

// Setup2FA generates secret and stores it in Redis temporarily
func (s *AuthService) Setup2FA(ctx context.Context, userID uuid.UUID) (*dto.Setup2FAResponse, error) {
    // 1. Get user
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("user not found")
    }

    if user.TwoFactorEnabled {
        return nil, fmt.Errorf("2FA is already enabled")
    }

    // 2. Generate TOTP secret
    key, err := s.totpManager.GenerateSecret(user.Email)
    if err != nil {
        return nil, fmt.Errorf("failed to generate secret: %w", err)
    }

    // 3. STORE SECRET IN REDIS (Valid for 10 minutes)
    // Yeh step missing tha pehle. Server ko yaad rakhna padega.
    err = s.tokenRepo.Store2FASetupSecret(ctx, userID, key.Secret(), 10*time.Minute)
    if err != nil {
        return nil, fmt.Errorf("failed to save setup state: %w", err)
    }

    // 4. Return Secret & QR URL (No recovery codes yet)
    return &dto.Setup2FAResponse{
        Secret:    key.Secret(),
        QRCodeURL: key.URL(),
    }, nil
}

// Enable2FA verifies the code against the stored secret and finalizes 2FA
func (s *AuthService) Enable2FA(ctx context.Context, userID uuid.UUID, req *dto.Enable2FARequest) (*dto.Enable2FAResponse, error) {
    // 1. Get User
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("user not found")
    }

    // 2. RETRIEVE SECRET FROM REDIS
    secret, err := s.tokenRepo.Get2FASetupSecret(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("setup session expired, please try setup again")
    }

    // 3. Validate Code
    if !s.totpManager.ValidateCode(secret, req.Code) {
        return nil, fmt.Errorf("invalid verification code")
    }

    // 4. Generate Recovery Codes (Now is the time to generate them)
    recoveryCodes, err := s.totpManager.GenerateRecoveryCodes(10)
    if err != nil {
        return nil, fmt.Errorf("failed to generate recovery codes")
    }

    // 5. Encrypt Secret for Database
    encryptedSecret, err := utils.Encrypt(secret, s.totpManager.GetAESKey())
    if err != nil {
        return nil, fmt.Errorf("failed to encrypt secret: %w", err)
    }

    // 6. Update User Model
    user.Enable2FA(encryptedSecret, recoveryCodes)

    if err := s.userRepo.Update(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to enable 2FA: %w", err)
    }

    // 7. Cleanup Redis (Remove the temp secret)
    _ = s.tokenRepo.Delete2FASetupSecret(ctx, userID)

    return &dto.Enable2FAResponse{
        Success:       true,
        Message:       "2FA enabled successfully",
        RecoveryCodes: recoveryCodes,
    }, nil
}

// Disable2FA remains mostly the same...
func (s *AuthService) Disable2FA(ctx context.Context, userID uuid.UUID, req *dto.Disable2FARequest) error {
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return fmt.Errorf("user not found")
    }

    if !user.CheckPassword(req.Password) {
        return fmt.Errorf("incorrect password")
    }

    user.Disable2FA()

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