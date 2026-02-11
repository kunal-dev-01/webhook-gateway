package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"shared/utils"
)

// ============================================================================
// ENUMS
// ============================================================================

type EmploymentStatus string

const (
	EmploymentStatusActive       EmploymentStatus = "active"
	EmploymentStatusProbation    EmploymentStatus = "probation"
	EmploymentStatusNoticePeriod EmploymentStatus = "notice_period"
	EmploymentStatusResigned     EmploymentStatus = "resigned"
	EmploymentStatusTerminated   EmploymentStatus = "terminated"
	EmploymentStatusRetired      EmploymentStatus = "retired"
	EmploymentStatusDeceased     EmploymentStatus = "deceased"
)

type EmploymentType string

const (
	EmploymentTypeFullTime   EmploymentType = "full_time"
	EmploymentTypePartTime   EmploymentType = "part_time"
	EmploymentTypeContract   EmploymentType = "contract"
	EmploymentTypeIntern     EmploymentType = "intern"
	EmploymentTypeConsultant EmploymentType = "consultant"
	EmploymentTypeTemporary  EmploymentType = "temporary"
)

type PaymentMode string

const (
	PaymentModeBankTransfer PaymentMode = "bank_transfer"
	PaymentModeCash         PaymentMode = "cash"
	PaymentModeCheque       PaymentMode = "cheque"
	PaymentModeUPI          PaymentMode = "upi"
)

type MaritalStatus string

const (
	MaritalStatusSingle   MaritalStatus = "single"
	MaritalStatusMarried  MaritalStatus = "married"
	MaritalStatusDivorced MaritalStatus = "divorced"
	MaritalStatusWidowed  MaritalStatus = "widowed"
)

// ============================================================================
// EMPLOYEE MODEL
// ============================================================================

type Employee struct {
	// Primary identification
	ID uuid.UUID `db:"id" json:"id"`

	// Company relationship
	CompanyID uuid.UUID `db:"company_id" json:"company_id"`

	// Optional: Link to user (if employee gets system access)
	UserID *uuid.UUID `db:"user_id" json:"user_id,omitempty"`

	// Employee identification
	EmployeeCode string  `db:"employee_code" json:"employee_code"`
	BiometricID  *string `db:"biometric_id" json:"biometric_id,omitempty"`

	// Personal Information
	FirstName  string  `db:"first_name" json:"first_name"`
	MiddleName *string `db:"middle_name" json:"middle_name,omitempty"`
	LastName   string  `db:"last_name" json:"last_name"`
	FullName   string  `db:"full_name" json:"full_name"` // Auto-generated

	// Contact Information
	Email          *string `db:"email" json:"email,omitempty"`
	Phone          string  `db:"phone" json:"phone"`
	AlternatePhone *string `db:"alternate_phone" json:"alternate_phone,omitempty"`

	// Personal Details
	DateOfBirth   time.Time      `db:"date_of_birth" json:"date_of_birth"`
	Age           int            `db:"age" json:"age"` // Auto-calculated
	Gender        string         `db:"gender" json:"gender"`
	BloodGroup    *string        `db:"blood_group" json:"blood_group,omitempty"`
	MaritalStatus *MaritalStatus `db:"marital_status" json:"marital_status,omitempty"`
	Nationality   string         `db:"nationality" json:"nationality"`

	// Avatar/Photo
	PhotoURL          *string `db:"photo_url" json:"photo_url,omitempty"`
	PhotoThumbnailURL *string `db:"photo_thumbnail_url" json:"photo_thumbnail_url,omitempty"`

	// Current Address
	CurrentAddressLine1 *string `db:"current_address_line1" json:"current_address_line1,omitempty"`
	CurrentAddressLine2 *string `db:"current_address_line2" json:"current_address_line2,omitempty"`
	CurrentCity         *string `db:"current_city" json:"current_city,omitempty"`
	CurrentState        *string `db:"current_state" json:"current_state,omitempty"`
	CurrentCountry      string  `db:"current_country" json:"current_country"`
	CurrentPincode      *string `db:"current_pincode" json:"current_pincode,omitempty"`

	// Permanent Address
	PermanentAddressLine1 *string `db:"permanent_address_line1" json:"permanent_address_line1,omitempty"`
	PermanentAddressLine2 *string `db:"permanent_address_line2" json:"permanent_address_line2,omitempty"`
	PermanentCity         *string `db:"permanent_city" json:"permanent_city,omitempty"`
	PermanentState        *string `db:"permanent_state" json:"permanent_state,omitempty"`
	PermanentCountry      string  `db:"permanent_country" json:"permanent_country"`
	PermanentPincode      *string `db:"permanent_pincode" json:"permanent_pincode,omitempty"`
	IsSameAddress         bool    `db:"is_same_address" json:"is_same_address"`

	// Government IDs (India-specific) - ALL SHOULD BE ENCRYPTED
	PANNumber        *string `db:"pan_number" json:"pan_number,omitempty"`
	AadharNumber     *string `db:"aadhar_number" json:"-"` // Never expose - encrypt
	UANNumber        *string `db:"uan_number" json:"uan_number,omitempty"`
	ESICNumber       *string `db:"esic_number" json:"esic_number,omitempty"`
	DrivingLicense   *string `db:"driving_license" json:"driving_license,omitempty"`
	PassportNumber   *string `db:"passport_number" json:"passport_number,omitempty"`

	// Employment Details
	Department           *string           `db:"department" json:"department,omitempty"`
	Designation          string            `db:"designation" json:"designation"`
	ReportingManagerID   *uuid.UUID        `db:"reporting_manager_id" json:"reporting_manager_id,omitempty"`
	EmploymentType       EmploymentType    `db:"employment_type" json:"employment_type"`
	EmploymentStatus     EmploymentStatus  `db:"employment_status" json:"employment_status"`

	// Dates
	DateOfJoining      time.Time  `db:"date_of_joining" json:"date_of_joining"`
	ProbationEndDate   *time.Time `db:"probation_end_date" json:"probation_end_date,omitempty"`
	ConfirmationDate   *time.Time `db:"confirmation_date" json:"confirmation_date,omitempty"`
	ResignationDate    *time.Time `db:"resignation_date" json:"resignation_date,omitempty"`
	LastWorkingDate    *time.Time `db:"last_working_date" json:"last_working_date,omitempty"`
	NoticePeriodDays   int        `db:"notice_period_days" json:"notice_period_days"`

	// Work Schedule
	WorkShift    string `db:"work_shift" json:"work_shift"`
	WorkLocation *string `db:"work_location" json:"work_location,omitempty"`
	IsRemote     bool   `db:"is_remote" json:"is_remote"`

	// Salary Information
	CurrentCTC  *float64     `db:"current_ctc" json:"current_ctc,omitempty"`     // Annual CTC
	BasicSalary *float64     `db:"basic_salary" json:"basic_salary,omitempty"`   // Monthly
	PaymentMode *PaymentMode `db:"payment_mode" json:"payment_mode,omitempty"`

	// Bank Details
	BankName          *string `db:"bank_name" json:"bank_name,omitempty"`
	BankAccountNumber *string `db:"bank_account_number" json:"bank_account_number,omitempty"`
	BankIFSCCode      *string `db:"bank_ifsc_code" json:"bank_ifsc_code,omitempty"`
	BankBranch        *string `db:"bank_branch" json:"bank_branch,omitempty"`

	// Statutory (PF, ESI, PT)
	IsPFApplicable  bool    `db:"is_pf_applicable" json:"is_pf_applicable"`
	PFNumber        *string `db:"pf_number" json:"pf_number,omitempty"`
	IsESIApplicable bool    `db:"is_esi_applicable" json:"is_esi_applicable"`
	IsPTApplicable  bool    `db:"is_pt_applicable" json:"is_pt_applicable"`

	// Emergency Contact
	EmergencyContactName         *string `db:"emergency_contact_name" json:"emergency_contact_name,omitempty"`
	EmergencyContactRelationship *string `db:"emergency_contact_relationship" json:"emergency_contact_relationship,omitempty"`
	EmergencyContactPhone        *string `db:"emergency_contact_phone" json:"emergency_contact_phone,omitempty"`

	// Family Details
	SpouseName     *string    `db:"spouse_name" json:"spouse_name,omitempty"`
	SpouseDOB      *time.Time `db:"spouse_dob" json:"spouse_dob,omitempty"`
	ChildrenCount  int        `db:"children_count" json:"children_count"`

	// Education
	HighestQualification *string `db:"highest_qualification" json:"highest_qualification,omitempty"`

	// Additional Info
	Skills    pq.StringArray `db:"skills" json:"skills,omitempty"`
	Languages pq.StringArray `db:"languages" json:"languages,omitempty"`
	Hobbies   *string        `db:"hobbies" json:"hobbies,omitempty"`

	// Medical
	HasDisabilities   bool    `db:"has_disabilities" json:"has_disabilities"`
	DisabilityDetails *string `db:"disability_details" json:"disability_details,omitempty"`

	// Status
	IsActive bool `db:"is_active" json:"is_active"`

	// Exit
	ExitReason           *string `db:"exit_reason" json:"exit_reason,omitempty"`
	ExitInterviewNotes   *string `db:"exit_interview_notes" json:"exit_interview_notes,omitempty"`

	// HR Notes
	HRNotes *string `db:"hr_notes" json:"hr_notes,omitempty"`

	// Timestamps
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	CreatedBy *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	UpdatedBy *uuid.UUID `db:"updated_by" json:"updated_by,omitempty"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// ============================================================================
// VALIDATION
// ============================================================================

// Validate validates employee data
func (e *Employee) Validate() error {
	if e.FirstName == "" || e.LastName == "" {
		return fmt.Errorf("first name and last name are required")
	}
	if e.EmployeeCode == "" {
		return fmt.Errorf("employee code is required")
	}
	if e.Phone == "" {
		return fmt.Errorf("phone number is required")
	}
	if !utils.IsValidIndianPhone(e.Phone) {
		return fmt.Errorf("invalid phone number format")
	}
	if e.Email != nil && !utils.IsValidEmail(*e.Email) {
		return fmt.Errorf("invalid email format")
	}
	if e.PANNumber != nil && !utils.IsValidPAN(*e.PANNumber) {
		return fmt.Errorf("invalid PAN format")
	}
	if e.AadharNumber != nil && !utils.IsValidAadhar(*e.AadharNumber) {
		return fmt.Errorf("invalid Aadhar format")
	}
	if e.BankIFSCCode != nil && !utils.IsValidIFSC(*e.BankIFSCCode) {
		return fmt.Errorf("invalid IFSC code")
	}
	if e.CurrentPincode != nil && !utils.IsValidPincode(*e.CurrentPincode) {
		return fmt.Errorf("invalid current pincode")
	}
	if e.PermanentPincode != nil && !utils.IsValidPincode(*e.PermanentPincode) {
		return fmt.Errorf("invalid permanent pincode")
	}
	return nil
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// IsEmployeeActive checks if employee is active
func (e *Employee) IsEmployeeActive() bool {
	return e.IsActive && e.EmploymentStatus == EmploymentStatusActive && e.DeletedAt == nil
}

// GetYearsOfService calculates years of service
func (e *Employee) GetYearsOfService() float64 {
	return time.Since(e.DateOfJoining).Hours() / (24 * 365.25)
}

// IsOnProbation checks if employee is on probation
func (e *Employee) IsOnProbation() bool {
	return e.EmploymentStatus == EmploymentStatusProbation
}

// ShouldAutoConfirm checks if probation period ended
func (e *Employee) ShouldAutoConfirm() bool {
	if e.ProbationEndDate == nil {
		return false
	}
	return e.IsOnProbation() && time.Now().After(*e.ProbationEndDate)
}

// SoftDelete soft deletes employee
func (e *Employee) SoftDelete() {
	now := time.Now()
	e.DeletedAt = &now
	e.IsActive = false
	e.EmploymentStatus = EmploymentStatusTerminated
	if e.Email != nil {
		email := fmt.Sprintf("%s.deleted.%d", *e.Email, now.Unix())
		e.Email = &email
	}
}