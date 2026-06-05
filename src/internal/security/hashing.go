package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPassword hashes a plaintext password using bcrypt.
// Returns the hashed password as a string or an error.
func HashPassword(rawPassword string) (string, error) {
	if rawPassword == "" {
		return "", fmt.Errorf("password must not be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
// Returns nil on match, error on mismatch or failure.
func CheckPassword(rawPassword, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(rawPassword))
	if err != nil {
		return fmt.Errorf("password does not match: %w", err)
	}

	return nil
}