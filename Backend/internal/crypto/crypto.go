package crypto

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"time"
	"unicode"
	"pulse-backend/internal/config"
	"pulse-backend/internal/models"
	"golang.org/x/crypto/pbkdf2"
)

// HashPassword hashes a password with salt using PBKDF2.
func HashPassword(password, saltHex string) (string, error) {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return "", fmt.Errorf("invalid salt hex: %w", err)
	}
	key := pbkdf2.Key([]byte(password), salt, config.PBKDF2Iterations, config.PBKDF2KeyLen, sha512.New)
	return hex.EncodeToString(key), nil
}

// VerifyPassword verifies a password against its stored hash.
func VerifyPassword(password, storedHash, saltHex string) (bool, error) {
	computed, err := HashPassword(password, saltHex)
	if err != nil {
		return false, err
	}
	return computed == storedHash, nil
}

// GenerateSalt creates a random 16-byte hex salt.
func GenerateSalt() (string, error) {
	b := make([]byte, config.SaltBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateToken creates a 32-byte random hex session token.
func GenerateToken() (string, error) {
	b := make([]byte, config.TokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateID generates a unique user ID.
func GenerateID() string {
	b := make([]byte, 4)
	rand.Read(b) //nolint:errcheck
	return fmt.Sprintf("%s-%s",
		strconv36(time.Now().UnixMilli()),
		hex.EncodeToString(b),
	)
}

// strconv36 is a base-36 encoder.
func strconv36(n int64) string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 13)
	for n > 0 {
		buf = append([]byte{digits[n%36]}, buf...)
		n /= 36
	}
	return string(buf)
}

// ValidatePasswordStrength validates password requirements.
func ValidatePasswordStrength(password string) models.PasswordValidation {
	var hasUpper, hasLower, hasNumber bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsNumber(r):
			hasNumber = true
		}
	}
	isLong := len(password) >= config.MinPasswordLength

	var errs []string
	if !hasUpper {
		errs = append(errs, "must contain at least one uppercase letter")
	}
	if !hasLower {
		errs = append(errs, "must contain at least one lowercase letter")
	}
	if !hasNumber {
		errs = append(errs, "must contain at least one number")
	}
	if !isLong {
		errs = append(errs, fmt.Sprintf("must be at least %d characters long", config.MinPasswordLength))
	}

	return models.PasswordValidation{
		IsValid:            len(errs) == 0,
		Errors:             errs,
		HasCaseSensitivity: hasUpper && hasLower,
	}
}
