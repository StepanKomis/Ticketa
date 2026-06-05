package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	userregistration "github.com/StepanKomis/Ticketa/src/internal/API/users/registration"
)

type UserHandler struct {
	httpLogger *logs.Logger
	userLogger *logs.Logger
	db         *sql.DB
}

type registrationResponse struct {
	ID int32 `json:"id"`
}

func NewUserHandler(httpLogger *logs.Logger, db *sql.DB) (*UserHandler, error) {
	uh := &UserHandler{}
	uh.httpLogger = httpLogger
	uh.db = db

	var err error
	uh.userLogger, err = logs.NewLogger("user")
	if err != nil {
		return nil, fmt.Errorf("failed to create user logger for user handler: %w", err)
	}

	return uh, nil
}

func (uh *UserHandler) post(w http.ResponseWriter, r *http.Request) {
	uh.httpLogger.Debugf("POST /api/users from %s", r.RemoteAddr)

	var body userregistration.RegistrationRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		uh.httpLogger.Debugf("error decoding registration request body: %s", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	uh.httpLogger.Debugf("registration request decoded: email=%s user_type=%s first_name=%s last_name=%s",
		body.Email, body.UserType, body.FirstName, body.LastName)

	err := userregistration.ValidatePassword(body.Password)
	if err != nil {
		uh.httpLogger.Debugf("password validation failed for %s: %s", body.Email, err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := userregistration.RegisterNewLocalUser(body, uh.db)
	if err != nil {
		uh.httpLogger.Debugf("registration failed for %s: %s", body.Email, err)
		if errors.Is(err, userregistration.ErrInvalidUserType) {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "failed to register user")
		}
		return
	}

	uh.httpLogger.Debugf("user registered successfully: id=%d email=%s user_type=%s", userID, body.Email, body.UserType)

	res := registrationResponse{ID: userID}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		uh.httpLogger.Debugf("error marshalling registration response for user %s: %s", body.Email, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonRes)
}

func (uh *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		uh.post(w, r)
	default:
		uh.httpLogger.Debugf("method not allowed: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		defaultResponse(w)
	}
}