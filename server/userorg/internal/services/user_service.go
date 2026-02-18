package services

import (
	"context"

	"github.com/google/uuid"

	"userorg/internal/dto"
	"userorg/internal/repository"

	"shared/errors"
	"shared/logger"
	"shared/models"
	sharedUtils "shared/utils"
)

type UserService struct {
	userRepo *repository.UserRepository
	log      *logger.Logger
}

func NewUserService(
	userRepo *repository.UserRepository,
	log *logger.Logger,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		log:      log,
	}
}

// ============================================================================
// GET ME
// ============================================================================

func (s *UserService) GetMe(ctx context.Context, userID uuid.UUID) (*dto.GetMeResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.GetMeResponse{
		User: dto.ToUserResponse(user),
	}, nil
}

// ============================================================================
// UPDATE USER
// ============================================================================

func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *dto.UpdateUserRequest) (*dto.UpdateUserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.FullName != nil {
		if err := sharedUtils.ValidateLength(*req.FullName, 2, 255, "full name"); err != nil {
			return nil, errors.BadRequest(err.Error())
		}
		user.FullName = *req.FullName
	}

	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error("Failed to update user", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	return &dto.UpdateUserResponse{
		FullName:  user.FullName,
		AvatarURL: user.AvatarURL,
	}, nil
}

// ============================================================================
// UPDATE PREFERENCES
// ============================================================================

func (s *UserService) UpdatePreferences(ctx context.Context, userID uuid.UUID, req *dto.UpdateUserPreferencesRequest) (*dto.UpdateUserPreferencesResponse, error) {
	if err := validatePreferences(req.Preferences); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := s.userRepo.UpdatePreferences(ctx, userID, req.Preferences); err != nil {
		s.log.Error("Failed to update preferences", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	return &dto.UpdateUserPreferencesResponse{
		Preferences: req.Preferences,
	}, nil
}

// ============================================================================
// HELPERS
// ============================================================================

func validatePreferences(p models.UserPreferences) error {
	validThemes := map[string]bool{"light": true, "dark": true, "system": true}
	if p.Theme != "" && !validThemes[p.Theme] {
		return errors.BadRequest("theme must be one of: light, dark, system")
	}
	if p.Timezone != "" {
		if err := sharedUtils.ValidateLength(p.Timezone, 1, 100, "timezone"); err != nil {
			return err
		}
	}
	return nil
}