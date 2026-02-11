package models

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// CompanyStatus represents the status of a company
type CompanyStatus string

const (
	CompanyActive    CompanyStatus = "active"
	CompanyInactive  CompanyStatus = "inactive"
	CompanySuspended CompanyStatus = "suspended"
)

// Company represents a company/organization
type Company struct {
	// Primary identification
	ID uuid.UUID `db:"id" json:"id"`

	// Owner
	OwnerID uuid.UUID `db:"owner_id" json:"owner_id"`

	// Company details
	CompanyName        string  `db:"company_name" json:"company_name"`
	LegalName          *string `db:"legal_name" json:"legal_name,omitempty"`
	Industry           *string `db:"industry" json:"industry,omitempty"`
	EmployeeCountRange *string `db:"employee_count_range" json:"employee_count_range,omitempty"`

	// Contact information
	Email   *string `db:"email" json:"email,omitempty"`
	Phone   *string `db:"phone" json:"phone,omitempty"`
	Website *string `db:"website" json:"website,omitempty"`

	// Address
	AddressLine1 *string `db:"address_line1" json:"address_line1,omitempty"`
	AddressLine2 *string `db:"address_line2" json:"address_line2,omitempty"`
	City         *string `db:"city" json:"city,omitempty"`
	State        *string `db:"state" json:"state,omitempty"`
	Country      string  `db:"country" json:"country"`
	Pincode      *string `db:"pincode" json:"pincode,omitempty"`

	// India-specific Tax/Legal details
	GSTIN *string `db:"gstin" json:"gstin,omitempty"` // GST Number
	PAN   *string `db:"pan" json:"pan,omitempty"`     // PAN
	TAN   *string `db:"tan" json:"tan,omitempty"`     // TAN
	CIN   *string `db:"cin" json:"cin,omitempty"`     // Corporate Identity Number

	// Statutory registration
	PFRegistrationNumber  *string `db:"pf_registration_number" json:"pf_registration_number,omitempty"`
	PFSubCode             *string `db:"pf_sub_code" json:"pf_sub_code,omitempty"`
	ESIRegistrationNumber *string `db:"esi_registration_number" json:"esi_registration_number,omitempty"`
	PTRegistrationNumber  *string `db:"pt_registration_number" json:"pt_registration_number,omitempty"`
	LWFRegistrationNumber *string `db:"lwf_registration_number" json:"lwf_registration_number,omitempty"`

	// Branding
	LogoURL      *string `db:"logo_url" json:"logo_url,omitempty"`
	PrimaryColor *string `db:"primary_color" json:"primary_color,omitempty"`

	// Status
	Status CompanyStatus `db:"status" json:"status"`

	// Subscription/Plan (for future billing)
	PlanType     string     `db:"plan_type" json:"plan_type"`
	TrialEndsAt  *time.Time `db:"trial_ends_at" json:"trial_ends_at,omitempty"`

	// Timestamps
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// ============================================================================
// Validation Methods
// ============================================================================

// Validate validates the company struct
func (c *Company) Validate() error {
	if c.CompanyName == "" {
		return fmt.Errorf("company name is required")
	}

	if c.Email != nil && *c.Email != "" && !isValidCompanyEmail(*c.Email) {
		return fmt.Errorf("invalid email format")
	}

	if c.Phone != nil && *c.Phone != "" && !isValidCompanyPhone(*c.Phone) {
		return fmt.Errorf("invalid phone number format")
	}

	if c.GSTIN != nil && *c.GSTIN != "" && !isValidGSTIN(*c.GSTIN) {
		return fmt.Errorf("invalid GSTIN format")
	}

	if c.PAN != nil && *c.PAN != "" && !isValidPAN(*c.PAN) {
		return fmt.Errorf("invalid PAN format")
	}

	if c.TAN != nil && *c.TAN != "" && !isValidTAN(*c.TAN) {
		return fmt.Errorf("invalid TAN format")
	}

	if c.Pincode != nil && *c.Pincode != "" && !isValidPincode(*c.Pincode) {
		return fmt.Errorf("invalid pincode format")
	}

	return nil
}

// ============================================================================
// Status Methods
// ============================================================================

// IsActive checks if the company is active
func (c *Company) IsActive() bool {
	return c.Status == CompanyActive && c.DeletedAt == nil
}

// IsDeleted checks if the company is soft deleted
func (c *Company) IsDeleted() bool {
	return c.DeletedAt != nil
}

// SoftDelete soft deletes the company
func (c *Company) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.Status = CompanyInactive
}

// Restore restores a soft-deleted company
func (c *Company) Restore() error {
	if c.DeletedAt == nil {
		return fmt.Errorf("company is not deleted")
	}

	c.DeletedAt = nil
	c.Status = CompanyActive

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// isValidCompanyEmail validates email format
func isValidCompanyEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidCompanyPhone validates phone number format
func isValidCompanyPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

// isValidGSTIN validates GSTIN format (15 digits)
func isValidGSTIN(gstin string) bool {
	gstinRegex := regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)
	return gstinRegex.MatchString(gstin)
}

// isValidPAN validates PAN format (10 characters)
func isValidPAN(pan string) bool {
	panRegex := regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`)
	return panRegex.MatchString(pan)
}

// isValidTAN validates TAN format (10 characters)
func isValidTAN(tan string) bool {
	tanRegex := regexp.MustCompile(`^[A-Z]{4}[0-9]{5}[A-Z]{1}$`)
	return tanRegex.MatchString(tan)
}

// isValidPincode validates Indian pincode (6 digits)
func isValidPincode(pincode string) bool {
	pincodeRegex := regexp.MustCompile(`^\d{6}$`)
	return pincodeRegex.MatchString(pincode)
}

// ============================================================================
// Public Response Methods
// ============================================================================

// PublicCompany represents company data safe for public API responses
type PublicCompany struct {
	ID                    uuid.UUID     `json:"id"`
	CompanyName           string        `json:"company_name"`
	LegalName             *string       `json:"legal_name,omitempty"`
	Industry              *string       `json:"industry,omitempty"`
	Email                 *string       `json:"email,omitempty"`
	Phone                 *string       `json:"phone,omitempty"`
	City                  *string       `json:"city,omitempty"`
	State                 *string       `json:"state,omitempty"`
	Country               string        `json:"country"`
	LogoURL               *string       `json:"logo_url,omitempty"`
	Status                CompanyStatus `json:"status"`
	PlanType              string        `json:"plan_type"`
	CreatedAt             time.Time     `json:"created_at"`
}

// ToPublic converts Company to PublicCompany (safe for API response)
func (c *Company) ToPublic() *PublicCompany {
	return &PublicCompany{
		ID:          c.ID,
		CompanyName: c.CompanyName,
		LegalName:   c.LegalName,
		Industry:    c.Industry,
		Email:       c.Email,
		Phone:       c.Phone,
		City:        c.City,
		State:       c.State,
		Country:     c.Country,
		LogoURL:     c.LogoURL,
		Status:      c.Status,
		PlanType:    c.PlanType,
		CreatedAt:   c.CreatedAt,
	}
}