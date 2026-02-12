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

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"shared/config"
	"shared/errors"
	sharedUtils "shared/utils"
)

type TOTPManager struct {
	issuer string
	aesKey []byte
}

func NewTOTPManager(cfg *config.TOTPConfig) (*TOTPManager, error) {
	keyBytes, err := hex.DecodeString(cfg.AESKey)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("invalid TOTP_AES_KEY format: %w", err))
	}

	if len(keyBytes) != 32 {
		return nil, errors.Internal(fmt.Errorf("TOTP_AES_KEY must be exactly 32 bytes"))
	}

	return &TOTPManager{
		issuer: cfg.Issuer,
		aesKey: keyBytes,
	}, nil
}

// ============================================================================
// SECRET MANAGEMENT
// ============================================================================

func (t *TOTPManager) EncryptSecret(secret string) (string, error) {
	encrypted, err := sharedUtils.Encrypt(secret, t.aesKey)
	if err != nil {
		return "", errors.Internal(fmt.Errorf("failed to encrypt totp secret: %w", err))
	}
	return encrypted, nil
}

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

func (t *TOTPManager) Generate(email string) (string, string, string, error) {
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
	encryptedSecret, err := t.EncryptSecret(rawSecret)
	if err != nil {
		return "", "", "", err
	}

	return rawSecret, encryptedSecret, key.URL(), nil
}

func (t *TOTPManager) Validate(encryptedSecret, code string) (bool, error) {
	secret, err := t.DecryptSecret(encryptedSecret)
	if err != nil {
		return false, err
	}

	valid, err := totp.ValidateCustom(
		code,
		secret,
		time.Now().UTC(),
		totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)

	if err != nil {
		return false, nil
	}

	return valid, nil
}

// ============================================================================
// RECOVERY CODES
// ============================================================================

func (t *TOTPManager) GenerateRecoveryCodes(count int) ([]string, []string, error) {
	if count <= 0 {
		count = 10
	}

	totalBytes := count * 5
	randomData := make([]byte, totalBytes)

	if _, err := rand.Read(randomData); err != nil {
		return nil, nil, errors.Internal(fmt.Errorf("failed to generate entropy: %w", err))
	}

	rawCodes := make([]string, count)
	encryptedCodes := make([]string, count)
	encoder := base32.StdEncoding.WithPadding(base32.NoPadding)

	for i := 0; i < count; i++ {
		start := i * 5
		chunk := randomData[start : start+5]

		code := encoder.EncodeToString(chunk)
		rawCodes[i] = code

		encrypted, err := t.EncryptSecret(code)
		if err != nil {
			return nil, nil, err
		}
		encryptedCodes[i] = encrypted
	}

	return rawCodes, encryptedCodes, nil
}

func (t *TOTPManager) ValidateRecoveryCode(inputCode string, encryptedStoredCodes []string) (bool, int, error) {
	if len(inputCode) == 0 || len(encryptedStoredCodes) == 0 {
		return false, -1, nil
	}

	normalizedInput := normalizeRecoveryCode(inputCode)
	if len(normalizedInput) == 0 {
		return false, -1, nil
	}
	inputBytes := []byte(normalizedInput)

	for i, encryptedCode := range encryptedStoredCodes {
		decrypted, err := t.DecryptSecret(encryptedCode)
		if err != nil {
			continue
		}

		normalizedStored := normalizeRecoveryCode(decrypted)

		if subtle.ConstantTimeCompare(inputBytes, []byte(normalizedStored)) == 1 {
			return true, i, nil
		}
	}

	return false, -1, nil
}

func normalizeRecoveryCode(code string) string {
	var b strings.Builder
	b.Grow(len(code))

	for _, char := range code {
		if char >= '0' && char <= '9' {
			b.WriteRune(char)
		} else if char >= 'A' && char <= 'Z' {
			b.WriteRune(char)
		} else if char >= 'a' && char <= 'z' {
			b.WriteRune(char - 32)
		}
	}
	return b.String()
}

// ============================================================================
// URI HELPER
// ============================================================================

func (t *TOTPManager) GetProvisioningURI(email, rawSecret string) string {
	issuer := url.QueryEscape(t.issuer)
	user := url.QueryEscape(email)

	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, user, rawSecret, issuer)
}