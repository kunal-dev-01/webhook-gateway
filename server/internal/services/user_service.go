package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"server/internal/repository"
	"server/internal/dto"
	"shared/models"
)

// UserService handles user business logic
type UserService struct {
	userRepo *repository.UserRepository
	logger   zerolog.Logger
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository, logger zerolog.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger.With().Str("service", "user_service").Logger(),
	}
}

// ============================================================================
// Get User Operations
// ============================================================================

// GetByID retrieves a user by ID with company verification
// UPDATED: Enforces multi-tenant isolation
func (s *UserService) GetByID(ctx context.Context, userID, companyID uuid.UUID) (*dto.UserResponse, error) {
	// Use company-scoped query for multi-tenant safety
	user, err := s.userRepo.GetByIDAndCompany(ctx, userID, companyID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return s.toUserResponse(user), nil
}

// GetMe retrieves current user information
// UPDATED: Now includes CompanyID and new fields
func (s *UserService) GetMe(ctx context.Context, userID, companyID uuid.UUID) (*dto.MeResponse, error) {
	// Use company-scoped query
	user, err := s.userRepo.GetByIDAndCompany(ctx, userID, companyID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &dto.MeResponse{
		ID:                user.ID,
		CompanyID:         user.CompanyID,
		Email:             user.Email,
		FullName:          user.FullName,
		Phone:             user.Phone,
		Role:              user.Role,
		Department:        user.Department,
		Designation:       user.Designation,
		EmployeeID:        user.EmployeeID,
		AvatarURL:         user.AvatarURL,
		Status:            user.Status,
		EmailVerified:     user.EmailVerified,
		TwoFactorEnabled:  user.TwoFactorEnabled,
		Permissions:       user.Permissions,
		LastLoginAt:       user.LastLoginAt,
		LastActiveAt:      user.LastActiveAt,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}, nil
}

// ============================================================================
// Create User Operations
// ============================================================================

// CreateUser creates a new user in the company (admin only)
// NEW: Multi-tenant user creation
func (s *UserService) CreateUser(ctx context.Context, companyID uuid.UUID, req *dto.CreateUserRequest, creatorID uuid.UUID) (uuid.UUID, error) {
	// Check if email already exists in company
	exists, err := s.userRepo.ExistsInCompany(ctx, req.Email, companyID)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to check email existence")
		return uuid.Nil, fmt.Errorf("internal error")
	}
	if exists {
		return uuid.Nil, fmt.Errorf("email already exists in company")
	}

	// Create user
	user := &models.User{
		ID:                uuid.New(),
		CompanyID:         companyID,
		Email:             req.Email,
		FullName:          req.FullName,
		Phone:             req.Phone,
		Role:              req.Role,
		Department:        req.Department,
		Designation:       req.Designation,
		EmployeeID:        req.EmployeeID,
		Status:            models.UserStatusActive,
		IsActive:          true,
		Timezone:          req.Timezone,
		Language:          req.Language,
		EmailVerified:     false, // Must verify email
		TwoFactorEnabled:  false,
		PasswordChangedAt: time.Now(),
		InvitedBy:         &creatorID,
		InvitedAt:         timePtr(time.Now()),
	}

	// Set default values if not provided
	if user.Timezone == "" {
		user.Timezone = "UTC"
	}
	if user.Language == "" {
		user.Language = "en"
	}

	// Set password
	if err := user.SetPassword(req.Password); err != nil {
		return uuid.Nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Validate
	if err := user.Validate(); err != nil {
		return uuid.Nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error().Err(err).Msg("failed to create user")
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("company_id", companyID.String()).
		Str("creator_id", creatorID.String()).
		Msg("user created by admin")

	// TODO: Send invitation/verification email
	return user.ID, nil
}

// ============================================================================
// Update User Operations
// ============================================================================

// UpdateUser updates user information (admin/HR)
// UPDATED: Enforces multi-tenant isolation and role hierarchy
func (s *UserService) UpdateUser(
	ctx context.Context,
	userID, companyID uuid.UUID,
	req *dto.UpdateUserRequest,
	requestorID uuid.UUID,
	requestorRole models.UserRole,
) error {
	// Get existing user (with company verification)
	user, err := s.userRepo.GetByIDAndCompany(ctx, userID, companyID)
	if err != nil {
		return fmt.Errorf("unauthorized: user not found in your company")
	}

	// Check permissions
	if userID != requestorID {
		// Only owner, admin and HR can update other users
		if requestorRole != models.UserRoleOwner &&
		   requestorRole != models.UserRoleAdmin && 
		   requestorRole != models.UserRoleHR {
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
	if req.Department != nil {
		user.Department = req.Department
	}
	if req.Designation != nil {
		user.Designation = req.Designation
	}
	if req.EmployeeID != nil {
		user.EmployeeID = req.EmployeeID
	}

	// Only owner and admin can update role and status
	if requestorRole == models.UserRoleOwner || requestorRole == models.UserRoleAdmin {
		if req.Role != nil {
			// Cannot change owner role
			if user.Role == models.UserRoleOwner {
				return fmt.Errorf("cannot change owner role. Transfer ownership first")
			}
			user.Role = *req.Role
		}

		if req.Status != nil {
			user.Status = *req.Status
		}
	} else if req.Role != nil || req.Status != nil {
		return fmt.Errorf("unauthorized: only admins and owners can update role or status")
	}

	// Validate updated user
	if err := user.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error().Err(err).Msg("failed to update user")
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateProfile allows user to update their own profile
// NEW: Self-service profile updates with limited fields
func (s *UserService) UpdateProfile(ctx context.Context, userID, companyID uuid.UUID, req *dto.UpdateMeRequest) error {
	// Get existing user (with company verification)
	user, err := s.userRepo.GetByIDAndCompany(ctx, userID, companyID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update only profile fields (not role/status/permissions)
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
	if req.DateOfBirth != nil {
		user.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		user.Gender = req.Gender
	}

	// Validate
	if err := user.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error().Err(err).Msg("failed to update profile")
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

// ============================================================================
// List Users Operations
// ============================================================================

// ListUsers retrieves paginated list of users within company
// UPDATED: Always scoped to company for multi-tenant isolation
func (s *UserService) ListUsers(
	ctx context.Context,
	companyID uuid.UUID,
	filters repository.ListUsersFilter,
	requestorRole models.UserRole,
) (*dto.ListUsersResponse, error) {
	// Check permissions - only owner, admin, HR, and accountant can list users
	if requestorRole != models.UserRoleOwner &&
	   requestorRole != models.UserRoleAdmin && 
	   requestorRole != models.UserRoleHR && 
	   requestorRole != models.UserRoleAccountant {
		return nil, fmt.Errorf("unauthorized: insufficient permissions to list users")
	}

	// List users (automatically scoped to company)
	users, total, err := s.userRepo.List(ctx, companyID, filters)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to list users")
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
// UPDATED: Enforces company isolation and prevents owner deletion
func (s *UserService) DeleteUser(ctx context.Context, userID, companyID uuid.UUID, requestorRole models.UserRole) error {
	// Only owner and admin can delete users
	if requestorRole != models.UserRoleOwner && requestorRole != models.UserRoleAdmin {
		return fmt.Errorf("unauthorized: insufficient permissions to delete users")
	}

	// Check if user exists in company
	user, err := s.userRepo.GetByIDAndCompany(ctx, userID, companyID)
	if err != nil {
		return fmt.Errorf("unauthorized: user not found in your company")
	}

	// Cannot delete company owner
	if user.Role == models.UserRoleOwner {
		return fmt.Errorf("cannot delete company owner")
	}

	// Soft delete
	if err := s.userRepo.SoftDelete(ctx, userID); err != nil {
		s.logger.Error().Err(err).Msg("failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Str("company_id", companyID.String()).
		Msg("user soft deleted")

	return nil
}

// ============================================================================
// User Statistics Operations
// ============================================================================

// GetUserStats retrieves user statistics for company dashboard
// NEW: Company-scoped statistics
func (s *UserService) GetUserStats(ctx context.Context, companyID uuid.UUID) (*dto.UserStatsResponse, error) {
	// Get all users in company
	allUsers, _, err := s.userRepo.List(ctx, companyID, repository.ListUsersFilter{
		Limit:  1000, // Get all users for stats
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	stats := &dto.UserStatsResponse{
		TotalUsers:    len(allUsers),
		ActiveUsers:   0,
		InactiveUsers: 0,
		PendingUsers:  0,
		UsersByRole:   make(map[models.UserRole]int),
		UsersByDept:   make(map[string]int),
		RecentLogins:  0,
		With2FA:       0,
	}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// Calculate stats
	for _, user := range allUsers {
		// Status counts
		switch user.Status {
		case models.UserStatusActive:
			if user.IsActive {
				stats.ActiveUsers++
			}
		case models.UserStatusInactive:
			stats.InactiveUsers++
		case models.UserStatusPendingVerification:
			stats.PendingUsers++
		}

		// Role counts
		stats.UsersByRole[user.Role]++

		// Department counts
		if user.Department != nil && *user.Department != "" {
			stats.UsersByDept[*user.Department]++
		}

		// Recent logins (last 24 hours)
		if user.LastLoginAt != nil && user.LastLoginAt.After(yesterday) {
			stats.RecentLogins++
		}

		// 2FA enabled count
		if user.TwoFactorEnabled {
			stats.With2FA++
		}
	}

	return stats, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// toUserResponse converts User model to UserResponse DTO
// UPDATED: Now includes CompanyID and new fields
func (s *UserService) toUserResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:               user.ID,
		CompanyID:        user.CompanyID,
		Email:            user.Email,
		FullName:         user.FullName,
		Phone:            user.Phone,
		Role:             user.Role,
		Department:       user.Department,
		Designation:      user.Designation,
		EmployeeID:       user.EmployeeID,
		AvatarURL:        user.AvatarURL,
		Status:           user.Status,
		EmailVerified:    user.EmailVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		Permissions:      user.Permissions,
		LastLoginAt:      user.LastLoginAt,
		LastActiveAt:     user.LastActiveAt,
		CreatedAt:        user.CreatedAt,
	}
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}