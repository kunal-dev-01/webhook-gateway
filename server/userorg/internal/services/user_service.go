package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"userorg/internal/dto"
	"userorg/internal/repository"
	"userorg/internal/utils"

	"shared/errors"
	"shared/logger"
	"shared/models"
	sharedUtils "shared/utils"
)

type UserService struct {
	userRepo    *repository.UserRepository
	r2          *utils.R2Client
	avatarQueue *utils.AvatarUploadQueue
	log         *logger.Logger
}

func NewUserService(
	userRepo *repository.UserRepository,
	r2 *utils.R2Client,
	avatarQueue *utils.AvatarUploadQueue,
	log *logger.Logger,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		r2:          r2,
		avatarQueue: avatarQueue,
		log:         log,
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
// UPDATE USER (full_name only via JSON)
// ============================================================================

func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *dto.UpdateUserRequest) (*dto.UpdateUserResponse, error) {
    if req.FullName == nil {
        return nil, errors.BadRequest("full_name is required")
    }

    if err := sharedUtils.ValidateLength(*req.FullName, 2, 255, "full name"); err != nil {
        return nil, errors.BadRequest(err.Error())
    }

    if err := s.userRepo.UpdateFullName(ctx, userID, *req.FullName); err != nil {
        s.log.Error("Failed to update full name", err, "user_id", userID)
        return nil, err
    }

    return &dto.UpdateUserResponse{
        FullName: *req.FullName,
    }, nil
}

// ============================================================================
// UPDATE AVATAR
// ============================================================================

// UpdateAvatar reads file bytes, saves pre-computed URL to DB, enqueues R2 upload.
// URL is returned immediately — worker uploads to R2 in background.
func (s *UserService) UpdateAvatar(ctx context.Context, userID uuid.UUID, file io.Reader, contentType string) (*dto.UpdateAvatarResponse, error) {
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to read file: %w", err))
	}

	// Generate deterministic key & pre-compute public URL
	ext := extensionFromContentType(contentType)
	objectKey := fmt.Sprintf("avatars/%s/%d%s", userID.String(), time.Now().UnixMilli(), ext)
	publicURL := s.r2.PublicURL(objectKey)

	// Save URL to DB immediately (optimistic — worker will fulfil the upload)
	if err := s.userRepo.UpdateAvatarURL(ctx, userID, publicURL); err != nil {
		s.log.Error("Failed to save avatar URL", err, "user_id", userID)
		return nil, errors.Internal(err)
	}

	// Enqueue background R2 upload
	job := utils.AvatarUploadJob{
		UserID:      userID,
		ObjectKey:   objectKey,
		FileData:    fileData,
		ContentType: contentType,
	}
	if err := s.avatarQueue.Enqueue(ctx, job); err != nil {
		s.log.Error("Failed to enqueue avatar upload", err, "user_id", userID)
		// Non-fatal: URL saved, log and continue
	}

	return &dto.UpdateAvatarResponse{
		AvatarURL: publicURL,
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

func extensionFromContentType(ct string) string {
	if ct == "image/png" {
		return ".png"
	}
	return ".jpg"
}