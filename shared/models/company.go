package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"shared/utils"
)

// ============================================================================
// ENUMS
// ============================================================================

type CompanyStatus string

const (
	CompanyStatusActive    CompanyStatus = "active"
	CompanyStatusInactive  CompanyStatus = "inactive"
	CompanyStatusSuspended CompanyStatus = "suspended"
)

// ============================================================================
// COMPANY SETTINGS
// ============================================================================

type CompanySettings struct {
	Payroll    PayrollSettings    `json:"payroll"`
	Leave      LeaveSettings      `json:"leave"`
	Attendance AttendanceSettings `json:"attendance"`
}

type PayrollSettings struct {
	Currency           string `json:"currency"`
	PayCycle           string `json:"pay_cycle"`
	PayDay             int    `json:"pay_day"`
	WorkDaysPerWeek    int    `json:"work_days_per_week"`
	WorkingHoursPerDay int    `json:"working_hours_per_day"`
}

type LeaveSettings struct {
	CasualLeaves    int  `json:"casual_leaves"`
	SickLeaves      int  `json:"sick_leaves"`
	EarnedLeaves    int  `json:"earned_leaves"`
	CarryForward    bool `json:"carry_forward"`
	MaxCarryForward int  `json:"max_carry_forward"`
}

type AttendanceSettings struct {
	LateMarkThresholdMinutes int  `json:"late_mark_threshold_minutes"`
	HalfDayHours             int  `json:"half_day_hours"`
	FullDayHours             int  `json:"full_day_hours"`
	AutoPunchOut             bool `json:"auto_punch_out"`
}

// Value implements driver.Valuer
func (cs CompanySettings) Value() (driver.Value, error) {
	return json.Marshal(cs)
}

// Scan implements sql.Scanner
func (cs *CompanySettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal CompanySettings")
	}
	return json.Unmarshal(bytes, cs)
}

// ============================================================================
// COMPANY MODEL
// ============================================================================

type Company struct {
	// Primary identification
	ID uuid.UUID `db:"id" json:"id"`

	// Owner
	OwnerID uuid.UUID `db:"owner_id" json:"owner_id"`

	// Company details
	CompanyName string  `db:"company_name" json:"company_name"`
	LegalName   *string `db:"legal_name" json:"legal_name,omitempty"`
	Industry    *string `db:"industry" json:"industry,omitempty"`
	Description *string `db:"description" json:"description,omitempty"`

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

	// India-specific Tax/Legal (ALL OPTIONAL)
	GSTIN *string `db:"gstin" json:"gstin,omitempty"` // GST Number
	PAN   *string `db:"pan" json:"pan,omitempty"`     // PAN
	TAN   *string `db:"tan" json:"tan,omitempty"`     // TAN
	CIN   *string `db:"cin" json:"cin,omitempty"`     // CIN

	// Statutory registrations
	PFRegistrationNumber   *string `db:"pf_registration_number" json:"pf_registration_number,omitempty"`
	PFSubCode              *string `db:"pf_sub_code" json:"pf_sub_code,omitempty"`
	ESIRegistrationNumber  *string `db:"esi_registration_number" json:"esi_registration_number,omitempty"`
	PTRegistrationNumber   *string `db:"pt_registration_number" json:"pt_registration_number,omitempty"`
	LWFRegistrationNumber  *string `db:"lwf_registration_number" json:"lwf_registration_number,omitempty"`
	MSMERegistrationNumber *string `db:"msme_registration_number" json:"msme_registration_number,omitempty"`

	// Branding
	LogoURL             *string `db:"logo_url" json:"logo_url,omitempty"`
	LogoThumbnailURL    *string `db:"logo_thumbnail_url" json:"logo_thumbnail_url,omitempty"`
	PrimaryColor        string  `db:"primary_color" json:"primary_color"`
	SecondaryColor      *string `db:"secondary_color" json:"secondary_color,omitempty"`

	// Settings
	Settings CompanySettings `db:"settings" json:"settings"`

	// Status
	Status CompanyStatus `db:"status" json:"status"`

	// Financial Year
	FinancialYearStartMonth int `db:"financial_year_start_month" json:"financial_year_start_month"`

	// Timestamps
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates company data
func (c *Company) Validate() error {
	if c.CompanyName == "" {
		return fmt.Errorf("company name is required")
	}

	// Email validation
	if c.Email != nil && *c.Email != "" && !utils.IsValidEmail(*c.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Phone validation
	if c.Phone != nil && *c.Phone != "" && !utils.IsValidIndianPhone(*c.Phone) {
		return fmt.Errorf("invalid phone number")
	}

	// GSTIN validation
	if c.GSTIN != nil && !utils.IsValidGSTIN(*c.GSTIN) {
		return fmt.Errorf("invalid GSTIN format")
	}

	// PAN validation
	if c.PAN != nil && !utils.IsValidPAN(*c.PAN) {
		return fmt.Errorf("invalid PAN format")
	}

	// TAN validation
	if c.TAN != nil && !utils.IsValidTAN(*c.TAN) {
		return fmt.Errorf("invalid TAN format")
	}

	// CIN validation
	if c.CIN != nil && !utils.IsValidCIN(*c.CIN) {
		return fmt.Errorf("invalid CIN format")
	}

	// Pincode validation
	if c.Pincode != nil && !utils.IsValidPincode(*c.Pincode) {
		return fmt.Errorf("invalid pincode format")
	}

	return nil
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// IsActive checks if company is active
func (c *Company) IsActive() bool {
	return c.Status == CompanyStatusActive && c.DeletedAt == nil
}

// SoftDelete soft deletes company
func (c *Company) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.Status = CompanyStatusInactive
}

// GetEmployeeCount returns employee count (implement in repository)
func (c *Company) GetEmployeeCount() int {
	// This will be implemented in repository layer
	return 0
}