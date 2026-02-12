package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"server/internal/services"
	"server/internal/dto"
	"shared/models"
	"server/internal/repository"
	"shared/middleware"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *services.UserService
	validate    *validator.Validate
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

// ============================================================================
// Get User Handlers
// ============================================================================

// GetMe retrieves current user information
// GET /api/v1/users/me
// UPDATED: Now includes CompanyID and new fields in response
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	// Get auth context (includes UserID and CompanyID)
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	// Get user info with company context
	user, err := h.userService.GetMe(c.Context(), authCtx.UserID, authCtx.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Failed to get user information",
		})
	}

	return c.JSON(user)
}

// GetUserByID retrieves a user by ID
// GET /api/v1/users/:id
// UPDATED: Enforces company isolation
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	// Get auth context
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	// Get user ID from params
	userIDParam := c.Params("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	// Check permissions
	requestorRole := models.UserRole(authCtx.Role)
	if requestorRole != models.UserRoleOwner &&
	   requestorRole != models.UserRoleAdmin && 
	   requestorRole != models.UserRoleHR && 
	   requestorRole != models.UserRoleAccountant {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Insufficient permissions",
		})
	}

	// Get user with company verification (multi-tenant safe)
	user, err := h.userService.GetByID(c.Context(), userID, authCtx.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "User not found",
		})
	}

	return c.JSON(user)
}

// ============================================================================
// Update User Handlers
// ============================================================================

// UpdateProfile updates current user's profile
// PATCH /api/v1/users/me
// UPDATED: Uses new UpdateMeRequest DTO for self-service updates
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	// Get auth context
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	// Parse request body
	var req dto.UpdateMeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Update profile (self-service, limited fields)
	if err := h.userService.UpdateProfile(c.Context(), authCtx.UserID, authCtx.CompanyID, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "Profile updated successfully",
	})
}

// UpdateUser updates a user by ID (admin/HR only)
// PATCH /api/v1/users/:id
// UPDATED: Enforces company isolation and role-based permissions
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	// Get auth context
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	// Get user ID from params
	userIDParam := c.Params("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	// Parse request body
	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Update user (with company and role verification)
	requestorRole := models.UserRole(authCtx.Role)
	if err := h.userService.UpdateUser(
		c.Context(), 
		userID, 
		authCtx.CompanyID, 
		&req, 
		authCtx.UserID, 
		requestorRole,
	); err != nil {
		if err.Error() == "unauthorized: you can only update your own profile" || 
		   err.Error() == "unauthorized: only admins and owners can update role or status" ||
		   err.Error() == "unauthorized: user not found in your company" {
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
		}

		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

// ============================================================================
// List Users Handler
// ============================================================================

// ListUsers retrieves paginated list of users
// GET /api/v1/users
// UPDATED: Scoped to company, includes new filters
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	// Get auth context
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	// Parse query parameters
	filters := repository.ListUsersFilter{
		Search:   c.Query("search", ""),
		SortBy:   c.Query("sort_by", "created_at"),
		SortDesc: c.Query("sort_desc", "true") == "true",
		Limit:    c.QueryInt("limit", 10),
		Offset:   c.QueryInt("offset", 0),
	}

	// Parse optional filters
	if status := c.Query("status"); status != "" {
		userStatus := models.UserStatus(status)
		filters.Status = &userStatus
	}

	if role := c.Query("role"); role != "" {
		userRole := models.UserRole(role)
		filters.Role = &userRole
	}

	if department := c.Query("department"); department != "" {
		filters.Department = &department
	}

	if emailVerified := c.Query("email_verified"); emailVerified != "" {
		verified := emailVerified == "true"
		filters.EmailVerified = &verified
	}

	if isActive := c.Query("is_active"); isActive != "" {
		active := isActive == "true"
		filters.IsActive = &active
	}

	// Get users (scoped to company)
	requestorRole := models.UserRole(authCtx.Role)
	users, err := h.userService.ListUsers(
		c.Context(), 
		authCtx.CompanyID, 
		filters, 
		requestorRole,
	)
	if err != nil {
		if err.Error() == "unauthorized: insufficient permissions to list users" {
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Failed to list users",
		})
	}

	return c.JSON(users)
}

// ============================================================================
// Delete User Handler
// ============================================================================

// DeleteUser soft deletes a user
// DELETE /api/v1/users/:id
// UPDATED: Enforces company isolation
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	// Get auth context
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	// Get user ID from params
	userIDParam := c.Params("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	// Delete user (with company and role verification)
	requestorRole := models.UserRole(authCtx.Role)
	if err := h.userService.DeleteUser(
		c.Context(), 
		userID, 
		authCtx.CompanyID, 
		requestorRole,
	); err != nil {
		if err.Error() == "unauthorized: insufficient permissions to delete users" ||
		   err.Error() == "unauthorized: user not found in your company" ||
		   err.Error() == "cannot delete company owner" {
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
		}

		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.MessageResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// ============================================================================
// Create User Handler (Admin Only)
// ============================================================================

// CreateUser creates a new user in the company
// POST /api/v1/users
// NEW: Admin can create users within their company
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	// Get auth context
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	// Only owner and admin can create users
	requestorRole := models.UserRole(authCtx.Role)
	if requestorRole != models.UserRoleOwner && requestorRole != models.UserRoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Insufficient permissions",
		})
	}

	// Parse request
	var req dto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Create user (automatically uses authCtx.CompanyID)
	userID, err := h.userService.CreateUser(c.Context(), authCtx.CompanyID, &req, authCtx.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
		"user_id": userID,
	})
}

// ============================================================================
// User Statistics Handler (Admin Dashboard)
// ============================================================================

// GetUserStats retrieves user statistics for admin dashboard
// GET /api/v1/users/stats
// NEW: Company-scoped statistics
func (h *UserHandler) GetUserStats(c *fiber.Ctx) error {
	// Get auth context
	authCtx, err := middleware.GetAuthContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	// Only owner and admin can view stats
	requestorRole := models.UserRole(authCtx.Role)
	if requestorRole != models.UserRoleOwner && requestorRole != models.UserRoleAdmin {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Insufficient permissions",
		})
	}

	// Get stats
	stats, err := h.userService.GetUserStats(c.Context(), authCtx.CompanyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Failed to get user statistics",
		})
	}

	return c.JSON(stats)
}