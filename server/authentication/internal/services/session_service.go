package services

import (
	"context"

	"github.com/google/uuid"

	"authentication/internal/dto"
	"authentication/internal/repository"
	"authentication/internal/utils"
	"authentication/internal/config"

	"shared/auth"
	"shared/errors"
	"shared/logger"
)

type SessionService struct {
	userRepo         *repository.UserRepository
	sessionMgr       *utils.Manager
	emailService     *utils.EmailService
	jwtManager       *auth.TokenManager
	cfg              *config.Config
	log              *logger.Logger
}

func NewSessionService(
	userRepo *repository.UserRepository,
	sessionMgr *utils.Manager,
	emailService *utils.EmailService,
	jwtManager *auth.TokenManager,
	cfg *config.Config,
	log *logger.Logger,
) *SessionService {
	return &SessionService{
		userRepo:         userRepo,
		sessionMgr:       sessionMgr,
		emailService:     emailService,
		jwtManager:       jwtManager,
		cfg:              cfg,
		log:              log,
	}
}

// ============================================================================
// REFRESH TOKEN
// ============================================================================

func (s *SessionService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
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
// LOGOUT
// ============================================================================

func (s *SessionService) Logout(ctx context.Context, userID uuid.UUID, sessionID string, req *dto.LogoutRequest) (*dto.LogoutResponse, error) {
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

func (s *SessionService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]utils.SessionData, error) {
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

func (s *SessionService) RevokeSessionByID(ctx context.Context, userID uuid.UUID, sessionID string) error {
	if err := s.sessionMgr.RevokeSession(ctx, userID, sessionID); err != nil {
		s.log.Error("Failed to revoke session", err, "user_id", userID, "session_id", sessionID)
		return errors.Internal(err)
	}

	s.log.Info("Session revoked", "user_id", userID, "session_id", sessionID)
	return nil
}

