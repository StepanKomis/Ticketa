package userregistration

import "errors"

var ErrInvalidUserType = errors.New("neplatný user_type")

type RegistrationRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserType  string `json:"user_type"`
}