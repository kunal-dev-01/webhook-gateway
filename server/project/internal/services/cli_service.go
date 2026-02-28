package services

import (
	"context"

	"github.com/google/uuid"

	"project/internal/dto"
	"project/internal/repository"

	"shared/errors"
	"shared/logger"
)

// ============================================================================
// CLI Session Service
// ============================================================================

type CLISessionService struct {
	cliSessionRepo *repository.CLISessionRepository
	log            *logger.Logger
}

func NewCLISessionService(
	cliSessionRepo *repository.CLISessionRepository,
	log *logger.Logger,
) *CLISessionService {
	return &CLISessionService{
		cliSessionRepo: cliSessionRepo,
		log:            log,
	}
}

// ============================================================================
// LIST CLI SESSIONS
// ============================================================================

// ListSessions returns all CLI sessions for the current user.
// currentSessionID is used to mark the active session in the response.
func (s *CLISessionService) ListSessions(ctx context.Context, userID uuid.UUID, currentSessionID *uuid.UUID) (*dto.ListCLISessionsResponse, error) {
	sessions, total, err := s.cliSessionRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.CLISessionResponse, 0, len(sessions))
	for _, sess := range sessions {
		responses = append(responses, dto.ToCLISessionResponse(sess, currentSessionID))
	}

	return &dto.ListCLISessionsResponse{
		Sessions: responses,
		Total:    total,
	}, nil
}

// ============================================================================
// GET CLI SESSION
// ============================================================================

func (s *CLISessionService) GetSession(ctx context.Context, sessionID, userID uuid.UUID, currentSessionID *uuid.UUID) (*dto.GetCLISessionResponse, error) {
	session, err := s.cliSessionRepo.GetByIDAndUser(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	return &dto.GetCLISessionResponse{
		Session: dto.ToCLISessionResponse(session, currentSessionID),
	}, nil
}

// ============================================================================
// REVOKE CLI SESSION
// ============================================================================

func (s *CLISessionService) RevokeSession(ctx context.Context, sessionID, userID uuid.UUID, req *dto.RevokeCLISessionRequest) (*dto.RevokeCLISessionResponse, error) {
	// Verify ownership
	session, err := s.cliSessionRepo.GetByIDAndUser(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	if session.IsRevoked() {
		return nil, errors.BadRequest("session is already revoked")
	}

	reason := req.Reason
	if reason == "" {
		reason = "Revoked by user"
	}

	if err := s.cliSessionRepo.Revoke(sessionID, reason); err != nil {
		return nil, err
	}

	session.Revoke(reason)

	s.log.Info("CLI session revoked",
		"session_id", sessionID,
		"user_id", userID,
	)

	return &dto.RevokeCLISessionResponse{
		Session: dto.ToCLISessionResponse(session, nil),
		Message: "CLI session revoked successfully",
	}, nil
}

// ============================================================================
// REVOKE ALL CLI SESSIONS
// ============================================================================

func (s *CLISessionService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) (*dto.RevokeAllCLISessionsResponse, error) {
	count, err := s.cliSessionRepo.RevokeAllByUser(userID, "Revoked all sessions by user")
	if err != nil {
		return nil, err
	}

	s.log.Info("All CLI sessions revoked",
		"user_id", userID,
		"count", count,
	)

	msg := "All CLI sessions revoked successfully"
	if count == 0 {
		msg = "No active sessions to revoke"
	}

	return &dto.RevokeAllCLISessionsResponse{
		RevokedCount: count,
		Message:      msg,
	}, nil
}