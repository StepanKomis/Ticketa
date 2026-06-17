package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/StepanKomis/Ticketa/src/cmd/server/env"
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
	// secureCookie nastaví Secure flag na session cookie. Výchozí false, protože
	// server zatím neumí TLS — nasazení za HTTPS proxy musí nastavit COOKIE_SECURE=true,
	// jinak prohlížeč cookie odešle i přes nešifrované HTTP.
	secureCookie bool
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
	uh.secureCookie = env.Get("COOKIE_SECURE", "false") == "true"

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
		Secure:   uh.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	res := currentUserResponse{
		ID:           user.ID,
		Email:        user.Email,
		FirstName:    user.FirstName.String,
		LastName:     user.LastName.String,
		UserType:     string(user.UserType),
		MustChangePw: user.MustChangePw,
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

	var mustChangePw bool
	if u.Provider == "local" {
		if ll, err := uh.queries.GetLocalLoginByUserID(r.Context(), u.ID); err == nil {
			mustChangePw = ll.MustChangePw
		}
	}

	res := currentUserResponse{
		ID:           u.ID,
		Email:        u.Email,
		FirstName:    u.FirstName.String,
		LastName:     u.LastName.String,
		UserType:     string(u.UserType),
		MustChangePw: mustChangePw,
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

// patchMe aktualizuje jméno a příjmení přihlášeného uživatele.
//
// @Summary      Aktualizovat vlastní profil
// @Description  Aktualizuje first_name a/nebo last_name. Email nelze změnit. Obě pole jsou volitelná.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      patchMeRequest        true  "Nové jméno a/nebo příjmení"
// @Success      200   {object}  currentUserResponse   "Aktualizovaný profil"
// @Failure      400   {object}  errorResponse         "Neplatné tělo požadavku"
// @Failure      401   {object}  errorResponse         "Chybí nebo vypršel session cookie"
// @Failure      500   {object}  errorResponse         "Interní chyba"
// @Security     cookieAuth
// @Router       /api/me [patch]
func (uh *UserHandler) patchMe(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	var body patchMeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}

	var firstName, lastName sql.NullString
	if body.FirstName != nil {
		firstName = sql.NullString{String: *body.FirstName, Valid: true}
	}
	if body.LastName != nil {
		lastName = sql.NullString{String: *body.LastName, Valid: true}
	}

	u, err := uh.queries.UpdateMyProfile(r.Context(), db.UpdateMyProfileParams{
		ID:        int32(session.UserID),
		FirstName: firstName,
		LastName:  lastName,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se aktualizovat profil")
		return
	}

	var mustChangePw bool
	if u.Provider == "local" {
		if ll, err := uh.queries.GetLocalLoginByUserID(r.Context(), u.ID); err == nil {
			mustChangePw = ll.MustChangePw
		}
	}

	res := currentUserResponse{
		ID:           u.ID,
		Email:        u.Email,
		FirstName:    u.FirstName.String,
		LastName:     u.LastName.String,
		UserType:     string(u.UserType),
		MustChangePw: mustChangePw,
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

// acceptInvite přijme pozvánku — ověří token, zaregistruje uživatele a zneplatní token.
//
// @Summary      Přijmout pozvánku
// @Description  Ověří pozvánkový token, vytvoří účet s rolí z pozvánky (přeskočí pending schvalování) a token označí jako použitý.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      acceptInviteRequest    true  "Token, heslo a jméno"
// @Success      201   {object}  acceptInviteResponse   "Vytvořený účet"
// @Failure      400   {object}  errorResponse          "Chybí povinné pole nebo neplatné heslo"
// @Failure      410   {object}  errorResponse          "Token vypršel nebo byl již použit"
// @Failure      500   {object}  errorResponse          "Interní chyba"
// @Router       /api/auth/invite/accept [post]
func (uh *UserHandler) acceptInvite(w http.ResponseWriter, r *http.Request) {
	var body acceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}
	if body.Token == "" || body.Password == "" {
		WriteError(w, http.StatusBadRequest, "token a password jsou povinné")
		return
	}

	if err := userregistration.ValidatePassword(body.Password); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	inv, err := uh.queries.GetInvitationByToken(ctx, body.Token)
	if err != nil {
		// Nerozlišujeme "nenalezeno" od "chyba DB" — oba případy jsou 410 z pohledu klienta.
		WriteError(w, http.StatusGone, "pozvánka neexistuje, vypršela nebo byla již použita")
		return
	}

	if inv.UsedAt.Valid || inv.ExpiresAt.Before(time.Now()) {
		WriteError(w, http.StatusGone, "pozvánka vypršela nebo byla již použita")
		return
	}

	hash, err := security.HashPassword(body.Password)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se zpracovat heslo")
		return
	}

	tx, err := uh.db.BeginTx(ctx, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "interní chyba serveru")
		return
	}
	defer tx.Rollback() //nolint:errcheck

	txQ := db.New(tx)

	user, err := txQ.CreateUser(ctx, db.CreateUserParams{
		Email:     inv.Email,
		FirstName: sql.NullString{String: body.FirstName, Valid: body.FirstName != ""},
		LastName:  sql.NullString{String: body.LastName, Valid: body.LastName != ""},
		UserType:  inv.UserType,
		Provider:  db.AuthProviderLocal,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se vytvořit účet")
		return
	}

	if err = txQ.CreateLocalLogin(ctx, db.CreateLocalLoginParams{
		ID:           user.ID,
		PasswordHash: hash,
		MustChangePw: false,
	}); err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se vytvořit přihlašovací údaje")
		return
	}

	if err = txQ.MarkInvitationUsed(ctx, inv.ID); err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se označit pozvánku jako použitou")
		return
	}

	if err = tx.Commit(); err != nil {
		WriteError(w, http.StatusInternalServerError, "interní chyba serveru")
		return
	}

	res := acceptInviteResponse{ID: user.ID, Email: user.Email}
	jsonRes, err := json.Marshal(res)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "interní chyba serveru")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonRes) //nolint:errcheck
}

// setupStatus vrátí, zda je systém inicializovaný (alespoň jeden uživatel existuje).
// Pokud needs_setup = true, první registrace automaticky vytvoří administrátora.
//
// @Summary      Stav inicializace systému
// @Description  Vrátí needs_setup=true pokud žádný uživatel neexistuje — první registrace pak automaticky dostane roli admin.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  setupStatusResponse
// @Failure      500  {object}  errorResponse
// @Router       /api/setup-status [get]
func (uh *UserHandler) setupStatus(w http.ResponseWriter, r *http.Request) {
	count, err := uh.queries.CountUsers(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se ověřit stav systému")
		return
	}
	res := setupStatusResponse{NeedsSetup: count == 0}
	jsonRes, err := json.Marshal(res)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "interní chyba serveru")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes) //nolint:errcheck
}

// patchMyPassword změní heslo přihlášeného lokálního uživatele.
// Vyžaduje ověření aktuálního hesla. Po úspěchu nastaví must_change_pw = FALSE a pw_changed_at = NOW().
//
// @Summary      Změna hesla
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  patchMyPasswordRequest  true  "Aktuální a nové heslo"
// @Success      204   "Heslo bylo úspěšně změněno"
// @Failure      400   {object}  errorResponse  "Neplatné tělo požadavku"
// @Failure      401   {object}  errorResponse  "Aktuální heslo je nesprávné nebo chybí session"
// @Failure      422   {object}  errorResponse  "Nové heslo nesplňuje požadavky"
// @Failure      500   {object}  errorResponse  "Interní chyba"
// @Security     cookieAuth
// @Router       /api/me/password [patch]
func (uh *UserHandler) patchMyPassword(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	var body patchMyPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}

	if err := userregistration.ValidatePassword(body.NewPassword); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	ll, err := uh.queries.GetLocalLoginByUserID(r.Context(), int32(session.UserID))
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "účet nepodporuje lokální přihlášení")
		return
	}

	if err := security.CheckPassword(body.CurrentPassword, ll.PasswordHash); err != nil {
		WriteError(w, http.StatusUnauthorized, "aktuální heslo je nesprávné")
		return
	}

	newHash, err := security.HashPassword(body.NewPassword)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "interní chyba")
		return
	}

	if err := uh.queries.UpdateLocalLoginPassword(r.Context(), db.UpdateLocalLoginPasswordParams{
		ID:           int32(session.UserID),
		PasswordHash: newHash,
	}); err != nil {
		WriteError(w, http.StatusInternalServerError, "interní chyba")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (uh *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/register":
		uh.register(w, r)
	case "/api/login":
		uh.login(w, r)
	case "/api/me":
		if r.Method == http.MethodPatch {
			uh.patchMe(w, r)
		} else {
			uh.me(w, r)
		}
	case "/api/me/password":
		uh.patchMyPassword(w, r)
	case "/api/logout":
		uh.logout(w, r)
	case "/api/setup-status":
		uh.setupStatus(w, r)
	case "/api/auth/invite/accept":
		uh.acceptInvite(w, r)
	default:
		uh.httpLogger.Debugf("neošetřená cesta: %s %s od %s", r.Method, r.URL.Path, r.RemoteAddr)
		defaultResponse(w)
	}
}
