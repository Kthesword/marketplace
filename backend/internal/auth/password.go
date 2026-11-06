// Package auth handles password hashing, JWT tokens, and authentication HTTP handlers.
package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// ErrPasswordTooLong is returned when a password exceeds bcrypt's 72-byte limit.
var ErrPasswordTooLong = errors.New("password must not exceed 72 bytes")

// HashPassword hashes a plaintext password using bcrypt with DefaultCost (10).
// Each call generates a fresh random salt, so hashing the same password twice
// produces two different strings.
func HashPassword(plain string) (string, error) {
	if len(plain) > 72 {
		return "", ErrPasswordTooLong
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword compares a bcrypt hash against a candidate plaintext password.
// It extracts the salt and cost from the hash string itself and runs in
// constant time to prevent timing attacks.
func CheckPassword(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
