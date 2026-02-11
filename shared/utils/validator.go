package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// ============================================================================
// REGEX PATTERNS
// ============================================================================

var (
	// Email regex (RFC 5322 simplified)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Indian phone regex (10 digits, optionally with +91 or 91 prefix)
	indianPhoneRegex = regexp.MustCompile(`^(\+91|91)?[6-9]\d{9}$`)

	// GSTIN regex (15 characters: 2-state code, 10-PAN, 1-entity, 1-Z, 1-checksum)
	gstinRegex = regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)

	// PAN regex (10 characters: 3-letters, 1-P/C/H/F/A/T/B/L/J/G, 1-letter, 4-digits, 1-letter)
	panRegex = regexp.MustCompile(`^[A-Z]{3}[PCHFATBLJG]{1}[A-Z]{1}[0-9]{4}[A-Z]{1}$`)

	// TAN regex (10 characters: 4-letters, 5-digits, 1-letter)
	tanRegex = regexp.MustCompile(`^[A-Z]{4}[0-9]{5}[A-Z]{1}$`)

	// CIN regex (21 characters for company identification)
	cinRegex = regexp.MustCompile(`^([LUu]{1})([0-9]{5})([A-Za-z]{2})([0-9]{4})([A-Za-z]{3})([0-9]{6})$`)

	// Indian Pincode regex (6 digits)
	pincodeRegex = regexp.MustCompile(`^[1-9][0-9]{5}$`)

	// Aadhar number regex (12 digits, optionally with spaces)
	aadharRegex = regexp.MustCompile(`^[2-9][0-9]{11}$`)

	// IFSC code regex (11 characters: 4 letters, 1 zero, 6 alphanumeric)
	ifscRegex = regexp.MustCompile(`^[A-Z]{4}0[A-Z0-9]{6}$`)
)

// ============================================================================
// EMAIL VALIDATION
// ============================================================================

// IsValidEmail validates email format using RFC 5322 simplified regex
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	email = strings.TrimSpace(email)
	if len(email) > 254 { // RFC 5321
		return false
	}
	return emailRegex.MatchString(email)
}

// ValidateEmail validates email and returns error if invalid
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if !IsValidEmail(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// ============================================================================
// PHONE VALIDATION
// ============================================================================

// IsValidIndianPhone validates Indian phone number format
// Accepts: 9876543210, +919876543210, 919876543210
func IsValidIndianPhone(phone string) bool {
	if phone == "" {
		return false
	}
	// Remove spaces and hyphens
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	return indianPhoneRegex.MatchString(phone)
}

// ValidateIndianPhone validates Indian phone and returns error if invalid
func ValidateIndianPhone(phone string) error {
	if phone == "" {
		return errors.New("phone number is required")
	}
	if !IsValidIndianPhone(phone) {
		return errors.New("invalid Indian phone number format (expected 10 digits starting with 6-9)")
	}
	return nil
}

// NormalizeIndianPhone normalizes Indian phone to E.164 format (+91XXXXXXXXXX)
func NormalizeIndianPhone(phone string) string {
	// Remove spaces and hyphens
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Remove +91 or 91 prefix if present
	phone = strings.TrimPrefix(phone, "+91")
	phone = strings.TrimPrefix(phone, "91")

	// Add +91 prefix
	return "+91" + phone
}

// ============================================================================
// INDIAN TAX/LEGAL VALIDATIONS
// ============================================================================

// IsValidGSTIN validates GSTIN (Goods and Services Tax Identification Number)
// Format: 22AAAAA0000A1Z5 (15 characters)
func IsValidGSTIN(gstin string) bool {
	if gstin == "" {
		return false
	}
	gstin = strings.ToUpper(strings.TrimSpace(gstin))
	if len(gstin) != 15 {
		return false
	}
	return gstinRegex.MatchString(gstin)
}

// ValidateGSTIN validates GSTIN and returns error if invalid
func ValidateGSTIN(gstin string) error {
	if gstin == "" {
		return errors.New("GSTIN is required")
	}
	if !IsValidGSTIN(gstin) {
		return errors.New("invalid GSTIN format (expected 15 characters)")
	}
	return nil
}

// IsValidPAN validates PAN (Permanent Account Number)
// Format: AAAAA9999A (10 characters)
func IsValidPAN(pan string) bool {
	if pan == "" {
		return false
	}
	pan = strings.ToUpper(strings.TrimSpace(pan))
	if len(pan) != 10 {
		return false
	}
	return panRegex.MatchString(pan)
}

// ValidatePAN validates PAN and returns error if invalid
func ValidatePAN(pan string) error {
	if pan == "" {
		return errors.New("PAN is required")
	}
	if !IsValidPAN(pan) {
		return errors.New("invalid PAN format (expected 10 characters)")
	}
	return nil
}

// IsValidTAN validates TAN (Tax Deduction Account Number)
// Format: AAAA99999A (10 characters)
func IsValidTAN(tan string) bool {
	if tan == "" {
		return false
	}
	tan = strings.ToUpper(strings.TrimSpace(tan))
	if len(tan) != 10 {
		return false
	}
	return tanRegex.MatchString(tan)
}

// ValidateTAN validates TAN and returns error if invalid
func ValidateTAN(tan string) error {
	if tan == "" {
		return errors.New("TAN is required")
	}
	if !IsValidTAN(tan) {
		return errors.New("invalid TAN format (expected 10 characters)")
	}
	return nil
}

// IsValidCIN validates CIN (Corporate Identification Number)
// Format: L12345AA1234AAA123456 (21 characters)
func IsValidCIN(cin string) bool {
	if cin == "" {
		return false
	}
	cin = strings.TrimSpace(cin)
	if len(cin) != 21 {
		return false
	}
	return cinRegex.MatchString(cin)
}

// ValidateCIN validates CIN and returns error if invalid
func ValidateCIN(cin string) error {
	if cin == "" {
		return errors.New("CIN is required")
	}
	if !IsValidCIN(cin) {
		return errors.New("invalid CIN format (expected 21 characters)")
	}
	return nil
}

// IsValidPincode validates Indian pincode (6 digits, first digit cannot be 0)
func IsValidPincode(pincode string) bool {
	if pincode == "" {
		return false
	}
	pincode = strings.TrimSpace(pincode)
	if len(pincode) != 6 {
		return false
	}
	return pincodeRegex.MatchString(pincode)
}

// ValidatePincode validates Indian pincode and returns error if invalid
func ValidatePincode(pincode string) error {
	if pincode == "" {
		return errors.New("pincode is required")
	}
	if !IsValidPincode(pincode) {
		return errors.New("invalid pincode format (expected 6 digits, first digit cannot be 0)")
	}
	return nil
}


// IsValidAadhar validates Aadhar number (12 digits, first digit cannot be 0 or 1)
func IsValidAadhar(aadhar string) bool {
	if aadhar == "" {
		return false
	}
	// Remove spaces and hyphens
	aadhar = strings.ReplaceAll(aadhar, " ", "")
	aadhar = strings.ReplaceAll(aadhar, "-", "")
	
	if len(aadhar) != 12 {
		return false
	}
	return aadharRegex.MatchString(aadhar)
}

// ValidateAadhar validates Aadhar number and returns error if invalid
func ValidateAadhar(aadhar string) error {
	if aadhar == "" {
		return errors.New("Aadhar number is required")
	}
	if !IsValidAadhar(aadhar) {
		return errors.New("invalid Aadhar number format (expected 12 digits, first digit must be 2-9)")
	}
	return nil
}

// NormalizeAadhar removes spaces and hyphens from Aadhar number
func NormalizeAadhar(aadhar string) string {
	aadhar = strings.ReplaceAll(aadhar, " ", "")
	aadhar = strings.ReplaceAll(aadhar, "-", "")
	return aadhar
}

// FormatAadhar formats Aadhar number with spaces (XXXX XXXX XXXX)
func FormatAadhar(aadhar string) string {
	normalized := NormalizeAadhar(aadhar)
	if len(normalized) != 12 {
		return aadhar // Return original if invalid
	}
	return normalized[:4] + " " + normalized[4:8] + " " + normalized[8:]
}

// IsValidIFSC validates IFSC code format
// Format: ABCD0123456 (4 letters, 1 zero, 6 alphanumeric)
func IsValidIFSC(ifsc string) bool {
	if ifsc == "" {
		return false
	}
	ifsc = strings.ToUpper(strings.TrimSpace(ifsc))
	if len(ifsc) != 11 {
		return false
	}
	return ifscRegex.MatchString(ifsc)
}

// ValidateIFSC validates IFSC code and returns error if invalid
func ValidateIFSC(ifsc string) error {
	if ifsc == "" {
		return errors.New("IFSC code is required")
	}
	if !IsValidIFSC(ifsc) {
		return errors.New("invalid IFSC code format (expected 11 characters: 4 letters, 0, 6 alphanumeric)")
	}
	return nil
}

// NormalizeIFSC normalizes IFSC code to uppercase
func NormalizeIFSC(ifsc string) string {
	return strings.ToUpper(strings.TrimSpace(ifsc))
}

// ============================================================================
// PASSWORD VALIDATION
// ============================================================================

// PasswordStrength represents password strength levels
type PasswordStrength int

const (
	PasswordWeak PasswordStrength = iota
	PasswordMedium
	PasswordStrong
	PasswordVeryStrong
)

// String returns string representation of password strength
func (ps PasswordStrength) String() string {
	switch ps {
	case PasswordWeak:
		return "weak"
	case PasswordMedium:
		return "medium"
	case PasswordStrong:
		return "strong"
	case PasswordVeryStrong:
		return "very strong"
	default:
		return "unknown"
	}
}

// ValidatePassword validates password requirements
// Requirements:
// - At least 8 characters
// - At most 72 characters (bcrypt limit)
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one number
// - At least one special character
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 72 {
		return errors.New("password must not exceed 72 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// ValidatePasswordSimple validates password with simpler requirements (for User model compatibility)
// Requirements:
// - At least 8 characters
// - At most 72 characters (bcrypt limit)
// - At least one letter
// - At least one number
func ValidatePasswordSimple(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 72 { // bcrypt limit
		return errors.New("password must be less than 72 characters")
	}

	// Check for at least one number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	// Check for at least one letter
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)

	if !hasNumber || !hasLetter {
		return errors.New("password must contain both letters and numbers")
	}

	return nil
}

// CheckPasswordStrength returns password strength level
func CheckPasswordStrength(password string) PasswordStrength {
	var score int

	// Length bonus
	if len(password) >= 8 {
		score++
	}
	if len(password) >= 12 {
		score++
	}
	if len(password) >= 16 {
		score++
	}

	// Character variety
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsNumber(char) {
			hasNumber = true
		}
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			hasSpecial = true
		}
	}

	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasNumber {
		score++
	}
	if hasSpecial {
		score++
	}

	// Return strength based on score
	switch {
	case score >= 7:
		return PasswordVeryStrong
	case score >= 5:
		return PasswordStrong
	case score >= 3:
		return PasswordMedium
	default:
		return PasswordWeak
	}
}

// ============================================================================
// GENERAL STRING VALIDATIONS
// ============================================================================

// IsEmpty checks if string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if string is not empty
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// ValidateRequired validates that a field is not empty
func ValidateRequired(value, fieldName string) error {
	if IsEmpty(value) {
		return errors.New(fieldName + " is required")
	}
	return nil
}

// ValidateMinLength validates minimum string length
func ValidateMinLength(value string, minLength int, fieldName string) error {
	if len(value) < minLength {
		return errors.New(fieldName + " must be at least " + string(rune(minLength)) + " characters")
	}
	return nil
}

// ValidateMaxLength validates maximum string length
func ValidateMaxLength(value string, maxLength int, fieldName string) error {
	if len(value) > maxLength {
		return errors.New(fieldName + " must not exceed " + string(rune(maxLength)) + " characters")
	}
	return nil
}

// ValidateLength validates string length range
func ValidateLength(value string, minLength, maxLength int, fieldName string) error {
	if err := ValidateMinLength(value, minLength, fieldName); err != nil {
		return err
	}
	if err := ValidateMaxLength(value, maxLength, fieldName); err != nil {
		return err
	}
	return nil
}

// ============================================================================
// NUMERIC VALIDATIONS
// ============================================================================

// ValidateRange validates that a number is within range
func ValidateRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return errors.New(fieldName + " must be between " + string(rune(min)) + " and " + string(rune(max)))
	}
	return nil
}

// ValidatePositive validates that a number is positive
func ValidatePositive(value int, fieldName string) error {
	if value <= 0 {
		return errors.New(fieldName + " must be positive")
	}
	return nil
}

// ValidateNonNegative validates that a number is non-negative
func ValidateNonNegative(value int, fieldName string) error {
	if value < 0 {
		return errors.New(fieldName + " must be non-negative")
	}
	return nil
}