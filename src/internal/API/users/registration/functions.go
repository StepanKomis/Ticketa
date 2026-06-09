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

// ValidatePassword ověří, že heslo splňuje minimální bezpečnostní požadavky:
// alespoň 8 znaků, nejvýše 72, jedna číslice a jeden speciální znak.
func ValidatePassword(rawPassword string) error {
	length := len(rawPassword)

	if length < minPasswordLength {
		return fmt.Errorf("heslo musí mít alespoň %d znaků", minPasswordLength)
	}

	if length > maxPasswordLength {
		return fmt.Errorf("heslo musí mít nejvýše %d znaků", maxPasswordLength)
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
		return fmt.Errorf("heslo musí obsahovat alespoň jednu číslici")
	}

	if !hasSpecial {
		return fmt.Errorf("heslo musí obsahovat alespoň jeden speciální znak")
	}

	return nil
}

var validUserTypes = map[string]db.UserType{
	"student":    db.UserTypeStudent,
	"staff":      db.UserTypeStaff,
	"maintainer": db.UserTypeMaintainer,
}

// RegisterNewLocalUser zaregistruje nového uživatele a jeho lokální přihlašovací údaje v rámci jedné transakce.
func RegisterNewLocalUser(b RegistrationRequest, psql *sql.DB) (int32, error) {
	userType, ok := validUserTypes[b.UserType]
	if !ok {
		return 0, fmt.Errorf("%w %q: musí být student, staff nebo maintainer", ErrInvalidUserType, b.UserType)
	}

	hash, err := security.HashPassword(b.Password)
	if err != nil {
		return 0, fmt.Errorf("nepodařilo se zahashovat heslo pro uživatele %s: %w", b.Email, err)
	}

	tx, err := psql.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("nepodařilo se zahájit transakci pro uživatele %s: %w", b.Email, err)
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
		return 0, fmt.Errorf("nepodařilo se vytvořit záznam uživatele %s: %w", b.Email, err)
	}

	loginParams := db.CreateLocalLoginParams{
		ID:           user.ID,
		PasswordHash: hash,
		MustChangePw: false,
	}

	if err = queries.CreateLocalLogin(context.Background(), loginParams); err != nil {
		return 0, fmt.Errorf("nepodařilo se vytvořit lokální přihlášení pro %s: %w", b.Email, err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("nepodařilo se potvrdit transakci pro uživatele %s: %w", b.Email, err)
	}

	return user.ID, nil
}