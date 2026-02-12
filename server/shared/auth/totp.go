package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"shared/config"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPManager handles TOTP (2FA) operations
type TOTPManager struct {
	config *config.TOTPConfig
	aesKey []byte
}

// NewTOTPManager creates new TOTP manager
func NewTOTPManager(cfg *config.TOTPConfig) (*TOTPManager, error) {
	// Decode AES key once during initialization
	aesKey, err := hex.DecodeString(cfg.AESKey)
	if err != nil {
		return nil, fmt.Errorf("invalid TOTP_AES_KEY format: %w", err)
	}

	if len(aesKey) != 32 {
		return nil, fmt.Errorf("TOTP_AES_KEY must be exactly 32 bytes (64 hex chars)")
	}

	return &TOTPManager{
		config: cfg,
		aesKey: aesKey,
	}, nil
}

// ============================================================================
// TOTP Generation
// ============================================================================

// GenerateSecret generates a new TOTP secret for a user
func (t *TOTPManager) GenerateSecret(email string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      t.config.Issuer,
		AccountName: email,
		Period:      30,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
}

// ============================================================================
// TOTP Validation - OPTIMIZED
// ============================================================================

// ValidateCode checks the code against the secret with default tolerance
func (t *TOTPManager) ValidateCode(secret, code string) bool {
	// Skew 1 = Allow 1 period before and 1 period after (30s * 3 window)
	return t.ValidateCodeWithSkew(secret, code, 1)
}

// ValidateCodeWithSkew validates with custom tolerance
func (t *TOTPManager) ValidateCodeWithSkew(secret, code string, skew uint) bool {
	// ✅ OPTIMIZED: Library's native implementation is faster than manual loops
	// It handles time window calculation internally without multiple GenerateCode calls
	valid, err := totp.ValidateCustom(
		code,
		secret,
		time.Now().UTC(),
		totp.ValidateOpts{
			Period:    30,
			Skew:      skew,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)

	if err != nil {
		return false
	}
	return valid
}

// ============================================================================
// Recovery Codes - BULK OPTIMIZED
// ============================================================================

// GenerateRecoveryCodes generates distinct recovery codes
func (t *TOTPManager) GenerateRecoveryCodes(count int) ([]string, error) {
	if count <= 0 {
		count = 10
	}

	// ✅ OPTIMIZATION: Bulk Random Read (1 Syscall instead of N)
	// Base32 needs 5 bytes to produce 8 characters
	// We need 'count' chunks of 5 bytes
	totalBytes := count * 5
	randomData := make([]byte, totalBytes)
	
	if _, err := rand.Read(randomData); err != nil {
		return nil, fmt.Errorf("failed to generate entropy: %w", err)
	}

	codes := make([]string, count)
	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)

	for i := 0; i < count; i++ {
		// Slice the bulk data
		start := i * 5
		end := start + 5
		chunk := randomData[start:end]
		
		// Encode to string (result is exactly 8 chars)
		codes[i] = encoder.EncodeToString(chunk)
	}

	return codes, nil
}

// Fix: Empty stored codes skip karo
func (t *TOTPManager) ValidateRecoveryCode(inputCode string, storedCodes []string) (bool, int) {
	if len(inputCode) == 0 || len(storedCodes) == 0 {
		return false, -1
	}

	normalizedInput := normalizeRecoveryCode(inputCode)
	if len(normalizedInput) == 0 {
		return false, -1 // Empty after normalization
	}
	
	inputBytes := []byte(normalizedInput)

	for i, storedCode := range storedCodes {
		if len(storedCode) == 0 {
			continue // Skip empty stored codes
		}
		
		normalizedStored := normalizeRecoveryCode(storedCode)
		if subtle.ConstantTimeCompare(inputBytes, []byte(normalizedStored)) == 1 {
			return true, i
		}
	}

	return false, -1
}

// normalizeRecoveryCode removes separators and uppercases (Optimized)
func normalizeRecoveryCode(code string) string {
	// ✅ OPTIMIZATION: strings.Builder minimizes allocations
	var b strings.Builder
	b.Grow(len(code))

	for _, char := range code {
		// Allow 0-9
		if char >= '0' && char <= '9' {
			b.WriteRune(char)
			continue
		}
		// Allow A-Z
		if char >= 'A' && char <= 'Z' {
			b.WriteRune(char)
			continue
		}
		// Convert a-z to A-Z
		if char >= 'a' && char <= 'z' {
			b.WriteRune(char - 32)
			continue
		}
		// Skip everything else (spaces, dashes)
	}
	return b.String()
}

// ============================================================================
// Encryption & Helpers
// ============================================================================

// GetAESKey returns the AES encryption key
func (t *TOTPManager) GetAESKey() []byte {
	return t.aesKey
}

// GetQRCodeURL returns URL for QR code generation
func (t *TOTPManager) GetQRCodeURL(key *otp.Key) string {
	return key.URL()
}

// GetProvisioningURI returns provisioning URI for manual entry
func (t *TOTPManager) GetProvisioningURI(email, secret string) string {
	// ✅ SAFETY: Use URL encoding to prevent broken links with special chars
	issuer := url.QueryEscape(t.config.Issuer)
	user := url.QueryEscape(email)
	
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, user, secret, issuer)
}

// GetCurrentCode generates current TOTP code (Testing helper)
func (t *TOTPManager) GetCurrentCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now().UTC())
}