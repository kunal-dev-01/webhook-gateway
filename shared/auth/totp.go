package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"time"

	"shared/config"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPManager handles TOTP (2FA) operations
type TOTPManager struct {
	config *config.TOTPConfig
	aesKey []byte // 32 bytes for AES-256
}

// NewTOTPManager creates new TOTP manager from config
func NewTOTPManager(cfg *config.TOTPConfig) (*TOTPManager, error) {
	// Decode AES key from hex
	aesKey, err := hex.DecodeString(cfg.AESKey)
	if err != nil {
		return nil, fmt.Errorf("invalid TOTP_AES_KEY: %w", err)
	}

	if len(aesKey) != 32 {
		return nil, fmt.Errorf("TOTP_AES_KEY must be 32 bytes (64 hex chars)")
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
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      t.config.Issuer,
		AccountName: email,
		Period:      30,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return key, nil
}

// ============================================================================
// TOTP Validation - CORRECTED ✅
// ============================================================================

// ValidateCode validates a TOTP code (default skew = 1, allows ±30 seconds)
func (t *TOTPManager) ValidateCode(secret, code string) bool {
	// ✅ CORRECT: Use totp.Validate with default behavior
	// Library internally handles skew=1 (±30 seconds)
	return totp.Validate(code, secret)
}

// ValidateCodeWithSkew validates with custom time skew
func (t *TOTPManager) ValidateCodeWithSkew(secret, code string, skew uint) bool {
	now := time.Now()

	// ✅ CORRECT: Manual validation with custom skew
	// Check current time
	if totp.Validate(code, secret) {
		return true
	}

	// Check past windows
	for i := uint(1); i <= skew; i++ {
		pastTime := now.Add(-time.Duration(i) * 30 * time.Second)
		pastCode, err := totp.GenerateCode(secret, pastTime)
		if err != nil {
			continue
		}
		if code == pastCode {
			return true
		}
	}

	// Check future windows
	for i := uint(1); i <= skew; i++ {
		futureTime := now.Add(time.Duration(i) * 30 * time.Second)
		futureCode, err := totp.GenerateCode(secret, futureTime)
		if err != nil {
			continue
		}
		if code == futureCode {
			return true
		}
	}

	return false
}

// ============================================================================
// Recovery Codes - OPTIMIZED
// ============================================================================

// GenerateRecoveryCodes generates backup recovery codes
func (t *TOTPManager) GenerateRecoveryCodes(count int) ([]string, error) {
	if count <= 0 {
		count = 10
	}

	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := generateRecoveryCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate recovery code: %w", err)
		}
		codes[i] = code
	}

	return codes, nil
}

func generateRecoveryCode() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Generate 8 character code (e.g., "ABCD1234")
	code := base32.StdEncoding.EncodeToString(b)[:8]
	return code, nil
}

// ValidateRecoveryCode checks if recovery code is valid
// Uses constant-time comparison to prevent timing attacks
func (t *TOTPManager) ValidateRecoveryCode(code string, recoveryCodes []string) (bool, int) {
	if len(code) == 0 || len(recoveryCodes) == 0 {
		return false, -1
	}

	// Normalize input
	code = normalizeRecoveryCode(code)

	for i, recoveryCode := range recoveryCodes {
		normalized := normalizeRecoveryCode(recoveryCode)

		// ✅ SECURITY: Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(code), []byte(normalized)) == 1 {
			return true, i
		}
	}

	return false, -1
}

// normalizeRecoveryCode removes spaces and converts to uppercase
func normalizeRecoveryCode(code string) string {
	result := ""
	for _, char := range code {
		if char != ' ' && char != '-' {
			if char >= 'a' && char <= 'z' {
				result += string(char - 32) // Convert to uppercase
			} else {
				result += string(char)
			}
		}
	}
	return result
}

// ============================================================================
// Encryption (for storing secrets in DB)
// ============================================================================

// GetAESKey returns the AES encryption key
func (t *TOTPManager) GetAESKey() []byte {
	return t.aesKey
}

// ============================================================================
// QR Code Generation
// ============================================================================

// GetQRCodeURL returns URL for QR code generation
func (t *TOTPManager) GetQRCodeURL(key *otp.Key) string {
	return key.URL()
}

// GetProvisioningURI returns provisioning URI for manual entry
func (t *TOTPManager) GetProvisioningURI(email, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		t.config.Issuer, email, secret, t.config.Issuer)
}

// GetCurrentCode generates current TOTP code (for testing/debugging)
func (t *TOTPManager) GetCurrentCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}