package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"server/internal/services"
	"server/internal/dto"
	"shared/models"
	"server/internal/repository"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ============================================================================
// Get User Handlers
// ============================================================================

// GetMe retrieves current user information
// GET /api/v1/users/me
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	// ✅ Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	user, err := h.userService.GetMe(c.Context(), userID)
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
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	// Get user ID from params
	userIDParam := c.Params("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	// ✅ Get requestor role from context
	roleStr, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}
	requestorRole := models.UserRole(roleStr)

	// Check permissions
	if requestorRole != models.RoleAdmin && 
	   requestorRole != models.RoleHR && 
	   requestorRole != models.RoleAccountant {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Insufficient permissions",
		})
	}

	user, err := h.userService.GetByID(c.Context(), userID)
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
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	// ✅ Get user ID from context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
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

	// Update profile
	if err := h.userService.UpdateProfile(c.Context(), userID, &req); err != nil {
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
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	// Get user ID from params
	userIDParam := c.Params("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	// ✅ Get requestor info from context
	requestorID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}

	roleStr, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}
	requestorRole := models.UserRole(roleStr)

	// Parse request body
	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Update user
	if err := h.userService.UpdateUser(c.Context(), userID, &req, requestorID, requestorRole); err != nil {
		if err.Error() == "unauthorized: you can only update your own profile" || 
		   err.Error() == "unauthorized: only admins can update role or status" {
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
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	// ✅ Get requestor role from context
	roleStr, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}
	requestorRole := models.UserRole(roleStr)

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

	if emailVerified := c.Query("email_verified"); emailVerified != "" {
		verified := emailVerified == "true"
		filters.EmailVerified = &verified
	}

	// Get users
	users, err := h.userService.ListUsers(c.Context(), filters, requestorRole)
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
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	// Get user ID from params
	userIDParam := c.Params("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Invalid user ID",
		})
	}

	// ✅ Get requestor role from context
	roleStr, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Error:   "Unauthorized",
		})
	}
	requestorRole := models.UserRole(roleStr)

	// Delete user
	if err := h.userService.DeleteUser(c.Context(), userID, requestorRole); err != nil {
		if err.Error() == "unauthorized: insufficient permissions to delete users" {
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