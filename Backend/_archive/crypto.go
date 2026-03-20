package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"time"
	"unicode"

	"golang.org/x/crypto/pbkdf2"
)

const (
	pbkdf2Iterations = 100000
	pbkdf2KeyLen     = 64
)

// hashPassword reproduces the Node.js crypto.pbkdf2Sync behaviour:
//   crypto.pbkdf2Sync(password, Buffer.from(salt, 'hex'), 100000, 64, 'sha512') → hex string
//
// The salt argument is a hex-encoded string (as stored in persons.json).
// IMPORTANT: Passwords are case-sensitive and must be validated exactly as provided.
func hashPassword(password, saltHex string) (string, error) {
  salt, err := hex.DecodeString(saltHex)
  if err != nil {
    return "", fmt.Errorf("invalid salt hex: %w", err)
  }
  key := pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyLen, sha512.New)
  return hex.EncodeToString(key), nil
}

// verifyPassword returns true when the stored hash matches a newly computed one.
// The password comparison is case-sensitive - Pass123 and pass123 produce different hashes.
func verifyPassword(password, storedHash, saltHex string) (bool, error) {
	computed, err := hashPassword(password, saltHex)
	if err != nil {
		return false, err
	}
	return computed == storedHash, nil
}

// generateSalt creates a random 16-byte hex salt (matches Node.js randomBytes(16).toString('hex')).
func generateSalt() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateToken creates a 32-byte random hex session token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateID matches Node.js: Date.now().toString(36) + '-' + randomBytes(4).hex()
func generateID() string {
	b := make([]byte, 4)
	rand.Read(b) //nolint:errcheck
	return fmt.Sprintf("%s-%s",
		strconv36(time.Now().UnixMilli()),
		hex.EncodeToString(b),
	)
}

// strconv36 is a minimal base-36 encoder used only by generateID.
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

// validatePasswordStrength enforces the same rules as the JS implementation.
func validatePasswordStrength(password string) passwordValidation {
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
	isLong := len(password) >= 3

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
		errs = append(errs, "must be at least 3 characters long")
	}

	return passwordValidation{
		IsValid:            len(errs) == 0,
		Errors:             errs,
		HasCaseSensitivity: hasUpper && hasLower,
	}
}
