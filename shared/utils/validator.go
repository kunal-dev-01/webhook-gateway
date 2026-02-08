package utils

import (
	"errors"
	"regexp"
	"unicode"
)

var (
	// Email regex (RFC 5322 simplified)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	
	// Phone regex (E.164 format)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
)

// ============================================================================
// Email Validation
// ============================================================================

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ============================================================================
// Phone Validation
// ============================================================================

// IsValidPhone validates phone number format
func IsValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// ============================================================================
// Password Validation
// ============================================================================

// PasswordStrength represents password strength
type PasswordStrength int

const (
	PasswordWeak PasswordStrength = iota
	PasswordMedium
	PasswordStrong
	PasswordVeryStrong
)

// ValidatePassword validates password requirements
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