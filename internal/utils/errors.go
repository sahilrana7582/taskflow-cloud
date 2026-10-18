package utils

import "errors"

var (
	// Auth Errors
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")

	// Validation Errors
	InvalidField = errors.New("invalid filed value")
)
