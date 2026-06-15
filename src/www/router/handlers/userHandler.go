package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/API/users/login"
	userregistration "github.com/StepanKomis/Ticketa/src/internal/API/users/registration"
	"github.com/StepanKomis/Ticketa/src/internal/security"
)

type UserHandler struct {
	httpLogger *logs.Logger
	userLogger *logs.Logger
	db         *sql.DB
	queries    *db.Queries
	store      *security.SessionStore
}

type registrationResponse struct {
	ID int32 `json:"id"`
}

func NewUserHandler(httpLogger *logs.Logger, sqlDB *sql.DB, store *security.SessionStore, cfg *config.Config) (*UserHandler, error) {
	uh := &UserHandler{}
	uh.httpLogger = httpLogger
	uh.db = sqlDB
	uh.queries = db.New(sqlDB)
	uh.store = store

	var err error
	uh.userLogger, err = logs.NewLogger("user", cfg)
	if err != nil {
		return nil, fmt.Errorf("nepodařilo se vytvořit logger pro userHandler: %w", err)
	}

	return uh, nil
}

// register vytvoří nový lokální účet.
// Po úspěšné registraci je potřeba se přihlásit přes POST /api/login — session cookie není nastaven automaticky.
//
// @Summary      Registrace
// @Description  Vytvoří nový lokální účet. Heslo musí mít alespoň 8 znaků a obsahovat velké písmeno, číslici a speciální znak.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      registerRequest       true  "Registrační údaje"
// @Success      201   {object}  registrationResponse  "ID nově vytvořeného uživatele"
// @Failure      400   {object}  errorResponse         "Chybí povinné pole, slabé heslo nebo neplatný user_type"
// @Failure      500   {object}  errorResponse         "Interní chyba (např. duplicitní e-mail)"
// @Router       /api/register [post]
func (uh *UserHandler) register(w http.ResponseWriter, r *http.Request) {
	uh.httpLogger.Debugf("POST /api/register od %s", r.RemoteAddr)

	var body userregistration.RegistrationRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		uh.httpLogger.Debugf("chyba dekódování těla požadavku registrace: %s", err)
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}

	uh.httpLogger.Debugf("tělo požadavku registrace dekódováno: email=%s user_type=%s first_name=%s last_name=%s",
		body.Email, body.UserType, body.FirstName, body.LastName)

	err := userregistration.ValidatePassword(body.Password)
	if err != nil {
		uh.httpLogger.Debugf("validace hesla selhala pro %s: %s", body.Email, err)
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := userregistration.RegisterNewLocalUser(body, uh.db)
	if err != nil {
		uh.httpLogger.Debugf("registrace selhala pro %s: %s", body.Email, err)
		if errors.Is(err, userregistration.ErrInvalidUserType) {
			WriteError(w, http.StatusBadRequest, err.Error())
		} else {
			WriteError(w, http.StatusInternalServerError, "nepodařilo se registrovat uživatele")
		}
		return
	}

	uh.httpLogger.Debugf("uživatel úspěšně zaregistrován: id=%d email=%s user_type=%s", userID, body.Email, body.UserType)

	res := registrationResponse{ID: userID}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		uh.httpLogger.Debugf("chyba serializace odpovědi registrace pro uživatele %s: %s", body.Email, err)
		WriteError(w, http.StatusInternalServerError, "interní chyba serveru")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonRes) //nolint:errcheck // a failed write after WriteHeader cannot be recovered
}

// login ověří přihlašovací údaje, nastaví session cookie a vrátí data přihlášeného uživatele.
//
// @Summary      Přihlášení
// @Description  Ověří e-mail a heslo. Při úspěchu nastaví HTTP-only cookie `session_token` platný 7 dní a vrátí profil uživatele.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      loginRequest   true  "Přihlašovací údaje"
// @Success      200   {object}  currentUserResponse  "Profil přihlášeného uživatele"
// @Failure      400   {object}  errorResponse        "Neplatné tělo požadavku"
// @Failure      401   {object}  errorResponse        "Špatné přihlašovací údaje nebo neaktivní účet"
// @Router       /api/login [post]
func (uh *UserHandler) login(w http.ResponseWriter, r *http.Request) {
	uh.httpLogger.Debugf("POST /api/login od %s", r.RemoteAddr)

	var body login.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		uh.httpLogger.Debugf("chyba dekódování těla požadavku přihlášení: %s", err)
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}

	user, token, err := body.Validate(uh.queries, uh.store, r)
	if err != nil {
		uh.httpLogger.Debugf("přihlášení selhalo pro %s: %s", body.Email, err)
		WriteError(w, http.StatusUnauthorized, "neplatné přihlašovací údaje")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     security.TokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(security.SessionTTLSeconds),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	res := currentUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
		UserType:  string(user.UserType),
	}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "interní chyba serveru")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes) //nolint:errcheck
}

// me vrátí profil aktuálně přihlášeného uživatele.
//
// @Summary      Profil přihlášeného uživatele
// @Description  Vrátí id, e-mail, jméno, příjmení a roli uživatele identifikovaného platným session cookie.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  currentUserResponse  "Profil uživatele"
// @Failure      401  {object}  errorResponse  "Chybí nebo vypršel session cookie"
// @Failure      500  {object}  errorResponse  "Interní chyba"
// @Security     cookieAuth
// @Router       /api/me [get]
func (uh *UserHandler) me(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	u, err := uh.queries.GetUserByID(r.Context(), int32(session.UserID))
	if err != nil {
		uh.httpLogger.Debugf("GET /api/me: GetUserByID selhalo pro user_id=%d: %s", session.UserID, err)
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst uživatele")
		return
	}

	res := currentUserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName.String,
		LastName:  u.LastName.String,
		UserType:  string(u.UserType),
	}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "interní chyba serveru")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes) //nolint:errcheck
}

// logout zneplatní serverovou session a smaže cookie v prohlížeči.
//
// @Summary      Odhlášení
// @Description  Smaže session cookie a označí session jako smazanou v databázi. Po odhlášení je session token okamžitě neplatný.
// @Tags         auth
// @Success      204  "Úspěšně odhlášen"
// @Failure      401  {object}  errorResponse  "Chybí nebo vypršel session cookie"
// @Failure      500  {object}  errorResponse  "Interní chyba při rušení session"
// @Security     cookieAuth
// @Router       /api/logout [post]
func (uh *UserHandler) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(security.TokenCookieName)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "nepřihlášen")
		return
	}

	if err := uh.store.Invalidate(r.Context(), cookie.Value); err != nil {
		uh.httpLogger.Debugf("POST /api/logout: Invalidate selhalo: %s", err)
		WriteError(w, http.StatusInternalServerError, "nepodařilo se zrušit session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     security.TokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (uh *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/register":
		uh.register(w, r)
	case "/api/login":
		uh.login(w, r)
	case "/api/me":
		uh.me(w, r)
	case "/api/logout":
		uh.logout(w, r)
	default:
		uh.httpLogger.Debugf("neošetřená cesta: %s %s od %s", r.Method, r.URL.Path, r.RemoteAddr)
		defaultResponse(w)
	}
}
