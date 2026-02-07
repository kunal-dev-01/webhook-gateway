package dto

import (
	"time"

	"github.com/google/uuid"
	"shared/models"
)

// CreateUserRequest represents request to create a user
type CreateUserRequest struct {
	Email    string           `json:"email" validate:"required,email"`
	Password string           `json:"password" validate:"required,min=8,max=72"`
	FullName string           `json:"full_name" validate:"required,min=2,max=255"`
	Phone    *string          `json:"phone" validate:"omitempty,e164"`
	Role     models.UserRole  `json:"role" validate:"required,oneof=admin hr accountant viewer"`
	Timezone string           `json:"timezone" validate:"omitempty"`
	Language string           `json:"language" validate:"omitempty,len=2"`
}

// UpdateUserRequest represents request to update a user
type UpdateUserRequest struct {
	FullName *string          `json:"full_name" validate:"omitempty,min=2,max=255"`
	Phone    *string          `json:"phone" validate:"omitempty,e164"`
	Role     *models.UserRole `json:"role" validate:"omitempty,oneof=admin hr accountant viewer"`
	Status   *models.UserStatus `json:"status" validate:"omitempty,oneof=active inactive suspended"`
	Timezone *string          `json:"timezone" validate:"omitempty"`
	Language *string          `json:"language" validate:"omitempty,len=2"`
}

// ChangePasswordRequest represents request to change password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken  string              `json:"access_token"`
	RefreshToken string              `json:"refresh_token"`
	User         *models.PublicUser  `json:"user"`
	ExpiresIn    int                 `json:"expires_in"` // seconds
}

// UserResponse represents user response
type UserResponse struct {
	ID               uuid.UUID          `json:"id"`
	Email            string             `json:"email"`
	FullName         string             `json:"full_name"`
	Phone            *string            `json:"phone,omitempty"`
	Role             models.UserRole    `json:"role"`
	Status           models.UserStatus  `json:"status"`
	EmailVerified    bool               `json:"email_verified"`
	TwoFactorEnabled bool               `json:"two_factor_enabled"`
	LastLoginAt      *time.Time         `json:"last_login_at,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
}

// ListUsersResponse represents response for listing users
type ListUsersResponse struct {
	Users []*UserResponse `json:"users"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}