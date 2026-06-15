package login

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/security"
)

// userQuerier načte kompletní řádek lokálního přihlášení včetně bcrypt hashe hesla.
// GetUserWithLocalLogin vrací pouze řádky kde is_active = TRUE, takže neaktivní
// účty jsou odmítnuty na úrovni dotazu.
type userQuerier interface {
	GetUserWithLocalLogin(ctx context.Context, email string) (db.GetUserWithLocalLoginRow, error)
}

type sessionCreator interface {
	Create(ctx context.Context, userID int64, r *http.Request) (db.Session, error)
}

// ToJson serializuje požadavek včetně plaintext pole Password.
// Výsledek nikdy nelogujte ani nepřeposílejte — odhaluje heslo uživatele v čitelné podobě.
func (lr *LoginRequest) ToJson() []byte {
	data, err := json.Marshal(lr)
	if err != nil {
		return nil
	}

	return data
}

// Validate vyhledá uživatele podle e-mailu, ověří heslo vůči uloženému bcrypt hashi
// a při úspěchu vytvoří novou session. Vrátí ověřený řádek uživatele a token session.
// Chyby "uživatel nenalezen" i "špatné heslo" vrátí stejnou nepřímou chybu,
// aby nebylo možné provádět enumeraci platných e-mailových adres.
func (lr *LoginRequest) Validate(q userQuerier, store sessionCreator, r *http.Request) (db.GetUserWithLocalLoginRow, string, error) {
	user, err := q.GetUserWithLocalLogin(r.Context(), lr.Email)
	if err != nil {
		// Záměrně nepřímé: rozlišení "uživatel nenalezen" od "špatné heslo"
		// by útočníkovi umožnilo enumerovat platné e-mailové adresy.
		return db.GetUserWithLocalLoginRow{}, "", fmt.Errorf("neplatné přihlašovací údaje")
	}

	if err := security.CheckPassword(lr.Password, user.PasswordHash); err != nil {
		return db.GetUserWithLocalLoginRow{}, "", fmt.Errorf("neplatné přihlašovací údaje")
	}

	session, err := store.Create(r.Context(), int64(user.ID), r)
	if err != nil {
		return db.GetUserWithLocalLoginRow{}, "", fmt.Errorf("vytváření session: %w", err)
	}

	return user, session.Token, nil
}
