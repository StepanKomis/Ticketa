package login

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/security"
)

// userQuerier fetches the full local-login row including the bcrypt password hash.
// GetUserWithLocalLogin only returns rows where is_active = TRUE, so inactive
// accounts are rejected at the query level.
type userQuerier interface {
	GetUserWithLocalLogin(ctx context.Context, email string) (db.GetUserWithLocalLoginRow, error)
}

type sessionCreator interface {
	Create(ctx context.Context, userID int64, r *http.Request) (db.Session, error)
}

// ! ToJson serialises the request including the plaintext Password field.
// ! Never log or transmit the result — it exposes the user's password in cleartext.
// ? Consider returning ([]byte, error) so callers are not silently given nil on failure.
func (lr *LoginRequest) ToJson() []byte {
	data, err := json.Marshal(lr)
	if err != nil {
		return nil
	}

	return data
}

// Validate looks up the user by email, verifies the password against the stored
// bcrypt hash, and creates a new session on success. Returns the session token.
// Both "user not found" and "wrong password" return the same opaque error to
// prevent account-enumeration attacks.
func (lr *LoginRequest) Validate(q userQuerier, store sessionCreator, r *http.Request) (string, error) {
	user, err := q.GetUserWithLocalLogin(r.Context(), lr.Email)
	if err != nil {
		// * Intentionally opaque: distinguishing "user not found" from "wrong password"
		// * would let an attacker enumerate valid email addresses.
		return "", fmt.Errorf("invalid credentials")
	}

	if err := security.CheckPassword(lr.Password, user.PasswordHash); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	session, err := store.Create(r.Context(), int64(user.ID), r)
	if err != nil {
		return "", fmt.Errorf("creating session: %w", err)
	}

	return session.Token, nil
}
