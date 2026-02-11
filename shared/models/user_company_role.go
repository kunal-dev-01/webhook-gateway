package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CompanyRole represents a user's role within a company
type CompanyRole string

const (
	CompanyRoleAdmin      CompanyRole = "admin"
	CompanyRoleHR         CompanyRole = "hr"
	CompanyRoleAccountant CompanyRole = "accountant"
	CompanyRoleViewer     CompanyRole = "viewer"
)

// UserCompanyRole represents the association between a user and a company with a specific role
type UserCompanyRole struct {
	// Primary identification
	ID uuid.UUID `db:"id" json:"id"`

	// Relationships
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	CompanyID uuid.UUID `db:"company_id" json:"company_id"`

	// Role in this specific company
	Role CompanyRole `db:"role" json:"role"`

	// Is this the user's primary/default company?
	IsPrimary bool `db:"is_primary" json:"is_primary"`

	// Status (can disable access without deleting)
	IsActive bool `db:"is_active" json:"is_active"`

	// Custom permissions (for future fine-grained control)
	CustomPermissions *string `db:"custom_permissions" json:"custom_permissions,omitempty"` // JSON string

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// Permission Methods (Company-specific)
// ============================================================================

// IsAdmin checks if user is admin in this company
func (ucr *UserCompanyRole) IsAdmin() bool {
	return ucr.Role == CompanyRoleAdmin && ucr.IsActive
}

// CanManageUsers checks if user can manage other users in this company
func (ucr *UserCompanyRole) CanManageUsers() bool {
	return (ucr.Role == CompanyRoleAdmin || ucr.Role == CompanyRoleHR) && ucr.IsActive
}

// CanRunPayroll checks if user can run payroll in this company
func (ucr *UserCompanyRole) CanRunPayroll() bool {
	return (ucr.Role == CompanyRoleAdmin || 
	        ucr.Role == CompanyRoleHR || 
	        ucr.Role == CompanyRoleAccountant) && ucr.IsActive
}

// CanViewReports checks if user can view reports in this company
func (ucr *UserCompanyRole) CanViewReports() bool {
	return ucr.IsActive // All active roles can view reports
}

// CanManageCompanySettings checks if user can manage company settings
func (ucr *UserCompanyRole) CanManageCompanySettings() bool {
	return ucr.Role == CompanyRoleAdmin && ucr.IsActive
}

// CanDeleteEmployees checks if user can delete employees
func (ucr *UserCompanyRole) CanDeleteEmployees() bool {
	return (ucr.Role == CompanyRoleAdmin || ucr.Role == CompanyRoleHR) && ucr.IsActive
}

// HasRole checks if user has a specific role in this company
func (ucr *UserCompanyRole) HasRole(role CompanyRole) bool {
	return ucr.Role == role && ucr.IsActive
}

// ============================================================================
// Validation Methods
// ============================================================================

// Validate validates the user company role struct
func (ucr *UserCompanyRole) Validate() error {
	if ucr.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}

	if ucr.CompanyID == uuid.Nil {
		return fmt.Errorf("company_id is required")
	}

	validRoles := []CompanyRole{
		CompanyRoleAdmin,
		CompanyRoleHR,
		CompanyRoleAccountant,
		CompanyRoleViewer,
	}

	isValidRole := false
	for _, validRole := range validRoles {
		if ucr.Role == validRole {
			isValidRole = true
			break
		}
	}

	if !isValidRole {
		return fmt.Errorf("invalid role: must be admin, hr, accountant, or viewer")
	}

	return nil
}

// ============================================================================
// Status Methods
// ============================================================================

// Activate activates the user's access to this company
func (ucr *UserCompanyRole) Activate() {
	ucr.IsActive = true
}

// Deactivate deactivates the user's access to this company
func (ucr *UserCompanyRole) Deactivate() {
	ucr.IsActive = false
}

// SetAsPrimary sets this company as the user's primary company
func (ucr *UserCompanyRole) SetAsPrimary() {
	ucr.IsPrimary = true
}

// UnsetAsPrimary unsets this company as the user's primary company
func (ucr *UserCompanyRole) UnsetAsPrimary() {
	ucr.IsPrimary = false
}

// ============================================================================
// Public Response Methods
// ============================================================================

// PublicUserCompanyRole represents safe data for API responses
type PublicUserCompanyRole struct {
	ID        uuid.UUID   `json:"id"`
	UserID    uuid.UUID   `json:"user_id"`
	CompanyID uuid.UUID   `json:"company_id"`
	Role      CompanyRole `json:"role"`
	IsPrimary bool        `json:"is_primary"`
	IsActive  bool        `json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
}

// ToPublic converts UserCompanyRole to PublicUserCompanyRole
func (ucr *UserCompanyRole) ToPublic() *PublicUserCompanyRole {
	return &PublicUserCompanyRole{
		ID:        ucr.ID,
		UserID:    ucr.UserID,
		CompanyID: ucr.CompanyID,
		Role:      ucr.Role,
		IsPrimary: ucr.IsPrimary,
		IsActive:  ucr.IsActive,
		CreatedAt: ucr.CreatedAt,
	}
}

// ============================================================================
// Helper Types (For Joins)
// ============================================================================

// UserWithCompanyRole represents a user with their company role
type UserWithCompanyRole struct {
	User
	CompanyRole     CompanyRole `db:"company_role" json:"company_role"`
	CompanyIsPrimary bool       `db:"company_is_primary" json:"company_is_primary"`
	CompanyIsActive  bool       `db:"company_is_active" json:"company_is_active"`
}

// CompanyWithUserRole represents a company with the current user's role
type CompanyWithUserRole struct {
	Company
	UserRole      CompanyRole `db:"user_role" json:"user_role"`
	UserIsPrimary bool        `db:"user_is_primary" json:"user_is_primary"`
	UserIsActive  bool        `db:"user_is_active" json:"user_is_active"`
}