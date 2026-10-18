package utils

import "errors"

var (
	// Auth Errors
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountSuspended   = errors.New("account suspended")

	// Validation Errors
	InvalidField = errors.New("invalid filed value")
)
