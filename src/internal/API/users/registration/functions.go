package userregistration

import (
	"context"
	"database/sql"
	"fmt"
	"unicode"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/security"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72
)

// ValidatePassword validates that a password meets minimum security requirements:
// at least 8 characters, at most 72, one digit, one special character.
func ValidatePassword(rawPassword string) error {
	length := len(rawPassword)

	if length < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", minPasswordLength)
	}

	if length > maxPasswordLength {
		return fmt.Errorf("password must be at most %d characters long", maxPasswordLength)
	}

	var hasDigit, hasSpecial bool

	for _, c := range rawPassword {
		switch {
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

var validUserTypes = map[string]db.UserType{
	"student":    db.UserTypeStudent,
	"staff":      db.UserTypeStaff,
	"maintainer": db.UserTypeMaintainer,
}

// RegisterNewUser registers a new user and their local login credentials within a single transaction.
func RegisterNewLocalUser(b RegistrationRequest, psql *sql.DB) (int32, error) {
	userType, ok := validUserTypes[b.UserType]
	if !ok {
		return 0, fmt.Errorf("%w %q: must be student, staff, or maintainer", ErrInvalidUserType, b.UserType)
	}

	hash, err := security.HashPassword(b.Password)
	if err != nil {
		return 0, fmt.Errorf("error hashing password for user %s: %w", b.Email, err)
	}

	tx, err := psql.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("error starting transaction for user %s: %w", b.Email, err)
	}
	defer tx.Rollback()

	queries := db.New(tx)

	userParams := db.CreateUserParams{
		Email:     b.Email,
		FirstName: sql.NullString{String: b.FirstName, Valid: b.FirstName != ""},
		LastName:  sql.NullString{String: b.LastName, Valid: b.LastName != ""},
		UserType:  userType,
		Provider:  db.AuthProviderLocal,
	}

	user, err := queries.CreateUser(context.Background(), userParams)
	if err != nil {
		return 0, fmt.Errorf("error creating user record for %s: %w", b.Email, err)
	}

	loginParams := db.CreateLocalLoginParams{
		ID:           user.ID,
		PasswordHash: hash,
		MustChangePw: false,
	}

	if err = queries.CreateLocalLogin(context.Background(), loginParams); err != nil {
		return 0, fmt.Errorf("error creating local login record for %s: %w", b.Email, err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("error committing transaction for user %s: %w", b.Email, err)
	}

	return user.ID, nil
}