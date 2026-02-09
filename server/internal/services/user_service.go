package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"server/internal/repository"
	"server/internal/dto"
	"shared/models"
)

// UserService handles user business logic
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// ============================================================================
// Get User Operations
// ============================================================================

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

// GetMe retrieves current user information
func (s *UserService) GetMe(ctx context.Context, userID uuid.UUID) (*dto.MeResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.MeResponse{
		ID:                user.ID,
		Email:             user.Email,
		FullName:          user.FullName,
		Phone:             user.Phone,
		Role:              user.Role,
		Status:            user.Status,
		EmailVerified:     user.EmailVerified,
		TwoFactorEnabled:  user.TwoFactorEnabled,
		LastLoginAt:       user.LastLoginAt,
		CreatedAt:         user.CreatedAt,
		PasswordChangedAt: user.PasswordChangedAt,
	}, nil
}

// ============================================================================
// Update User Operations
// ============================================================================

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *dto.UpdateUserRequest, requestorID uuid.UUID, requestorRole models.UserRole) error {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check permissions
	if userID != requestorID {
		// Only admin and HR can update other users
		if requestorRole != models.RoleAdmin && requestorRole != models.RoleHR {
			return fmt.Errorf("unauthorized: you can only update your own profile")
		}
	}

	// Update fields if provided
	if req.FullName != nil {
		user.FullName = *req.FullName
	}

	if req.Phone != nil {
		user.Phone = req.Phone
	}

	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}

	if req.Language != nil {
		user.Language = *req.Language
	}

	// Only admin can update role and status
	if requestorRole == models.RoleAdmin {
		if req.Role != nil {
			user.Role = *req.Role
		}

		if req.Status != nil {
			user.Status = *req.Status
		}
	} else if req.Role != nil || req.Status != nil {
		return fmt.Errorf("unauthorized: only admins can update role or status")
	}

	// Validate updated user
	if err := user.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateProfile allows user to update their own profile (name, phone, timezone, language)
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateUserRequest) error {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update only profile fields (not role/status)
	if req.FullName != nil {
		user.FullName = *req.FullName
	}

	if req.Phone != nil {
		user.Phone = req.Phone
	}

	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}

	if req.Language != nil {
		user.Language = *req.Language
	}

	// Validate
	if err := user.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

// ============================================================================
// List Users Operations
// ============================================================================

// ListUsers retrieves paginated list of users
func (s *UserService) ListUsers(ctx context.Context, filters repository.ListUsersFilter, requestorRole models.UserRole) (*dto.ListUsersResponse, error) {
	// Check permissions - only admin, HR, and accountant can list users
	if requestorRole != models.RoleAdmin && requestorRole != models.RoleHR && requestorRole != models.RoleAccountant {
		return nil, fmt.Errorf("unauthorized: insufficient permissions to list users")
	}

	users, total, err := s.userRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to response DTOs
	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(user)
	}

	// Calculate page number from offset
	page := 1
	if filters.Limit > 0 {
		page = (filters.Offset / filters.Limit) + 1
	}

	return &dto.ListUsersResponse{
		Users: userResponses,
		Total: total,
		Page:  page,
		Limit: filters.Limit,
	}, nil
}

// ============================================================================
// Delete User Operations
// ============================================================================

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID, requestorRole models.UserRole) error {
	// Only admin and HR can delete users
	if requestorRole != models.RoleAdmin && requestorRole != models.RoleHR {
		return fmt.Errorf("unauthorized: insufficient permissions to delete users")
	}

	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Soft delete
	if err := s.userRepo.SoftDelete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// toUserResponse converts User model to UserResponse DTO
func (s *UserService) toUserResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:               user.ID,
		Email:            user.Email,
		FullName:         user.FullName,
		Phone:            user.Phone,
		Role:             user.Role,
		Status:           user.Status,
		EmailVerified:    user.EmailVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		LastLoginAt:      user.LastLoginAt,
		CreatedAt:        user.CreatedAt,
	}
}