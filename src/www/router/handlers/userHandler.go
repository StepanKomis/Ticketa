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
	store      *security.SessionStore
}

type registrationResponse struct {
	ID int32 `json:"id"`
}

func NewUserHandler(httpLogger *logs.Logger, db *sql.DB, store *security.SessionStore, cfg *config.Config) (*UserHandler, error) {
	uh := &UserHandler{}
	uh.httpLogger = httpLogger
	uh.db = db
	uh.store = store

	var err error
	uh.userLogger, err = logs.NewLogger("user", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create user logger for user handler: %w", err)
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
	uh.httpLogger.Debugf("POST /api/register from %s", r.RemoteAddr)

	var body userregistration.RegistrationRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		uh.httpLogger.Debugf("error decoding registration request body: %s", err)
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	uh.httpLogger.Debugf("registration request decoded: email=%s user_type=%s first_name=%s last_name=%s",
		body.Email, body.UserType, body.FirstName, body.LastName)

	err := userregistration.ValidatePassword(body.Password)
	if err != nil {
		uh.httpLogger.Debugf("password validation failed for %s: %s", body.Email, err)
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := userregistration.RegisterNewLocalUser(body, uh.db)
	if err != nil {
		uh.httpLogger.Debugf("registration failed for %s: %s", body.Email, err)
		if errors.Is(err, userregistration.ErrInvalidUserType) {
			WriteError(w, http.StatusBadRequest, err.Error())
		} else {
			WriteError(w, http.StatusInternalServerError, "failed to register user")
		}
		return
	}

	uh.httpLogger.Debugf("user registered successfully: id=%d email=%s user_type=%s", userID, body.Email, body.UserType)

	res := registrationResponse{ID: userID}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		uh.httpLogger.Debugf("error marshalling registration response for user %s: %s", body.Email, err)
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonRes) //nolint:errcheck // a failed write after WriteHeader cannot be recovered
}

// login ověří přihlašovací údaje a nastaví session cookie.
// Cookie `session_token` je HTTP-only a platí 7 dní.
//
// @Summary      Přihlášení
// @Description  Ověří e-mail a heslo. Při úspěchu nastaví HTTP-only cookie `session_token` platný 7 dní.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  loginRequest  true  "Přihlašovací údaje"
// @Success      200   "Session cookie nastaven"
// @Failure      400   {object}  errorResponse  "Neplatné tělo požadavku"
// @Failure      401   {object}  errorResponse  "Špatné přihlašovací údaje nebo neaktivní účet"
// @Router       /api/login [post]
func (uh *UserHandler) login(w http.ResponseWriter, r *http.Request) {
	uh.httpLogger.Debugf("POST /api/login from %s", r.RemoteAddr)

	var body login.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		uh.httpLogger.Debugf("error decoding login request body: %s", err)
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	q := db.New(uh.db)
	token, err := body.Validate(q, uh.store, r)
	if err != nil {
		uh.httpLogger.Debugf("login failed for %s: %s", body.Email, err)
		WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     security.TokenCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusOK)
}

func (uh *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/register":
		uh.register(w, r)
	case "/api/login":
		uh.login(w, r)
	default:
		uh.httpLogger.Debugf("unhandled path: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		defaultResponse(w)
	}
}