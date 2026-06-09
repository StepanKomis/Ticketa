package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPassword zahashuje plaintext heslo pomocí bcrypt.
// Vrátí zahashované heslo jako řetězec nebo chybu.
func HashPassword(rawPassword string) (string, error) {
	if rawPassword == "" {
		return "", fmt.Errorf("heslo nesmí být prázdné")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("nepodařilo se zahashovat heslo: %w", err)
	}

	return string(hash), nil
}

// CheckPassword porovná plaintext heslo s bcrypt hashem.
// Vrátí nil při shodě, chybu při neshodě nebo selhání.
func CheckPassword(rawPassword, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(rawPassword))
	if err != nil {
		return fmt.Errorf("heslo se neshoduje: %w", err)
	}

	return nil
}