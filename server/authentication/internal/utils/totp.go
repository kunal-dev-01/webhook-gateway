package utils

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
	"shared/errors" 
	sharedUtils "shared/utils"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPManager handles Two-Factor Authentication
type TOTPManager struct {
	issuer string
	aesKey []byte // Decoded key for encryption/decryption
}

// NewTOTPManager creates a new manager instance
func NewTOTPManager(cfg *config.TOTPConfig) (*TOTPManager, error) {
	// 1. Decode Key (Hex to Bytes)
	keyBytes, err := hex.DecodeString(cfg.AESKey)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("invalid TOTP_AES_KEY format: %w", err))
	}

	// 2. Validate Key Length (Must be 32 bytes for AES-256)
	if len(keyBytes) != 32 {
		return nil, errors.Internal(fmt.Errorf("TOTP_AES_KEY must be exactly 32 bytes (64 hex chars)"))
	}

	return &TOTPManager{
		issuer: cfg.Issuer,
		aesKey: keyBytes,
	}, nil
}

// ============================================================================
// SECRET MANAGEMENT (Encryption Wrapper)
// ============================================================================

// EncryptSecret encrypts the raw TOTP secret for DB storage
func (t *TOTPManager) EncryptSecret(secret string) (string, error) {
	// Hum same package mein hain, isliye Encrypt direct call kar sakte hain
	encrypted, err := sharedUtils.Encrypt(secret, t.aesKey)
	if err != nil {
		return "", errors.Internal(fmt.Errorf("failed to encrypt totp secret: %w", err))
	}
	return encrypted, nil
}

// DecryptSecret decrypts the stored TOTP secret
func (t *TOTPManager) DecryptSecret(encryptedSecret string) (string, error) {
	decrypted, err := sharedUtils.Decrypt(encryptedSecret, t.aesKey)
	if err != nil {
		return "", errors.Internal(fmt.Errorf("failed to decrypt totp secret: %w", err))
	}
	return decrypted, nil
}

// ============================================================================
// TOTP OPERATIONS
// ============================================================================

// Generate generates a new Secret and QR Code URL
// Returns: (RawSecret, EncryptedSecret, QRCodeURL, error)
func (t *TOTPManager) Generate(email string) (string, string, string, error) {
	// 1. Generate Key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      t.issuer,
		AccountName: email,
		Period:      30,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", "", errors.Internal(fmt.Errorf("failed to generate totp key: %w", err))
	}

	rawSecret := key.Secret()

	// 2. Encrypt Secret for Storage
	encryptedSecret, err := t.EncryptSecret(rawSecret)
	if err != nil {
		return "", "", "", err
	}

	// 3. Generate QR URL
	return rawSecret, encryptedSecret, key.URL(), nil
}

// Validate verifies a code against an ENCRYPTED secret
func (t *TOTPManager) Validate(encryptedSecret, code string) (bool, error) {
	// 1. Decrypt the secret first
	secret, err := t.DecryptSecret(encryptedSecret)
	if err != nil {
		return false, err
	}

	// 2. Validate with Skew (Tolerance of +/- 1 period i.e., 30 seconds)
	// Library handles the time window logic efficiently
	valid, err := totp.ValidateCustom(
		code,
		secret,
		time.Now().UTC(),
		totp.ValidateOpts{
			Period:    30,
			Skew:      1, // Allow 1 step back/forward for clock drift
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)

	if err != nil {
		// Ye error tab aata hai agar input format galat ho (e.g. non-numeric code)
		// Hum isse "Internal" nahi maanege, bas validation fail
		return false, nil
	}

	return valid, nil
}

// ============================================================================
// RECOVERY CODES
// ============================================================================

// GenerateRecoveryCodes generates distinct recovery codes
// Returns: RawCodes (for user), EncryptedCodes (for DB)
func (t *TOTPManager) GenerateRecoveryCodes(count int) ([]string, []string, error) {
	if count <= 0 {
		count = 10
	}

	// 1. Bulk Random Generation (Optimized)
	totalBytes := count * 5 // Base32 needs 5 bytes for 8 chars
	randomData := make([]byte, totalBytes)
	
	if _, err := rand.Read(randomData); err != nil {
		return nil, nil, errors.Internal(fmt.Errorf("failed to generate entropy: %w", err))
	}

	rawCodes := make([]string, count)
	encryptedCodes := make([]string, count)
	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)

	// 2. Process chunks
	for i := 0; i < count; i++ {
		start := i * 5
		chunk := randomData[start : start+5]
		
		code := encoder.EncodeToString(chunk)
		rawCodes[i] = code

		// Encrypt immediately
		encrypted, err := t.EncryptSecret(code)
		if err != nil {
			return nil, nil, err
		}
		encryptedCodes[i] = encrypted
	}

	return rawCodes, encryptedCodes, nil
}

// ValidateRecoveryCode checks a code against ENCRYPTED stored codes
// Returns: (isValid, indexUsed, error)
func (t *TOTPManager) ValidateRecoveryCode(inputCode string, encryptedStoredCodes []string) (bool, int, error) {
	if len(inputCode) == 0 || len(encryptedStoredCodes) == 0 {
		return false, -1, nil
	}

	// 1. Normalize Input
	normalizedInput := normalizeRecoveryCode(inputCode)
	if len(normalizedInput) == 0 {
		return false, -1, nil
	}
	inputBytes := []byte(normalizedInput)

	// 2. Iterate and Decrypt one by one
	// Note: Ye thoda heavy operation hai, but recovery codes rarely use hote hain.
	for i, encryptedCode := range encryptedStoredCodes {
		// Decrypt
		decrypted, err := t.DecryptSecret(encryptedCode)
		if err != nil {
			// Skip corrupt data, log error internally if needed
			continue 
		}

		normalizedStored := normalizeRecoveryCode(decrypted)
		
		// Constant Time Compare to prevent timing attacks
		if subtle.ConstantTimeCompare(inputBytes, []byte(normalizedStored)) == 1 {
			return true, i, nil
		}
	}

	return false, -1, nil
}

// normalizeRecoveryCode cleanups the code (removes dashes, spaces, uppercases)
func normalizeRecoveryCode(code string) string {
	var b strings.Builder
	b.Grow(len(code))

	for _, char := range code {
		if char >= '0' && char <= '9' {
			b.WriteRune(char)
		} else if char >= 'A' && char <= 'Z' {
			b.WriteRune(char)
		} else if char >= 'a' && char <= 'z' {
			b.WriteRune(char - 32) // Convert to Upper
		}
	}
	return b.String()
}

// ============================================================================
// URI HELPER
// ============================================================================

// GetProvisioningURI returns the URI for manual entry in Authenticator apps
func (t *TOTPManager) GetProvisioningURI(email, rawSecret string) string {
	issuer := url.QueryEscape(t.issuer)
	user := url.QueryEscape(email)
	
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, user, rawSecret, issuer)
}