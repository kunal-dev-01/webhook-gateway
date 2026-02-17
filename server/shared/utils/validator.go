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
	if len(email) > 254 { // RFC 5321 max length
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

// NormalizeEmail normalizes email to lowercase and trims whitespace
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
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

// ValidatePasswordSimple validates password with simple requirements
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

// ValidatePassword validates password with strong requirements
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
	if len(strings.TrimSpace(value)) < minLength {
		return errors.New(fieldName + " must be at least " + string(rune(minLength+'0')) + " characters")
	}
	return nil
}

// ValidateMaxLength validates maximum string length
func ValidateMaxLength(value string, maxLength int, fieldName string) error {
	if len(value) > maxLength {
		return errors.New(fieldName + " must not exceed " + string(rune(maxLength+'0')) + " characters")
	}
	return nil
}

// ValidateLength validates string length range
func ValidateLength(value string, minLength, maxLength int, fieldName string) error {
	trimmed := strings.TrimSpace(value)
	length := len(trimmed)
	
	if length < minLength {
		return errors.New(fieldName + " must be at least " + string(rune(minLength+'0')) + " characters")
	}
	if length > maxLength {
		return errors.New(fieldName + " must not exceed " + string(rune(maxLength+'0')) + " characters")
	}
	return nil
}

// ============================================================================
// SLUG VALIDATION
// ============================================================================

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// IsValidSlug validates URL-friendly slug format
// Valid: "my-org", "acme-corp-2024"
// Invalid: "My Org", "acme_corp", "org-", "-org"
func IsValidSlug(slug string) bool {
	if slug == "" {
		return false
	}
	if len(slug) < 3 || len(slug) > 50 {
		return false
	}
	return slugRegex.MatchString(slug)
}

// ValidateSlug validates organization slug
func ValidateSlug(slug string) error {
	if slug == "" {
		return errors.New("slug is required")
	}
	if !IsValidSlug(slug) {
		return errors.New("invalid slug format (use lowercase letters, numbers, and hyphens only)")
	}
	return nil
}

// GenerateSlug generates URL-friendly slug from name
func GenerateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	
	// Remove special characters (keep only a-z, 0-9, and -)
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	slug = reg.ReplaceAllString(slug, "")
	
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")
	
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	
	// Limit length to 50 characters
	if len(slug) > 41 {
		slug = slug[:41]
		slug = strings.TrimRight(slug, "-")
	}
	
	return slug
}

// ============================================================================
// NUMERIC VALIDATIONS
// ============================================================================

// ValidateRange validates that a number is within range
func ValidateRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return errors.New(fieldName + " must be between " + string(rune(min+'0')) + " and " + string(rune(max+'0')))
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