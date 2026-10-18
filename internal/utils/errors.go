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

	// Client error kinds. Services return them through Invalid, NotFound, etc.,
	// and WriteError turns each kind into its HTTP status.
	ErrInvalidInput = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
)

// ClientError is an error whose message is safe to return to API clients.
// errors.Is(err, ErrNotFound) and friends match on its kind.
type ClientError struct {
	Kind    error
	Message string
}

func (e *ClientError) Error() string { return e.Message }
func (e *ClientError) Unwrap() error { return e.Kind }

func Invalid(msg string) error      { return &ClientError{Kind: ErrInvalidInput, Message: msg} }
func Unauthorized(msg string) error { return &ClientError{Kind: ErrUnauthorized, Message: msg} }
func Forbidden(msg string) error    { return &ClientError{Kind: ErrForbidden, Message: msg} }
func NotFound(msg string) error     { return &ClientError{Kind: ErrNotFound, Message: msg} }
func Conflict(msg string) error     { return &ClientError{Kind: ErrConflict, Message: msg} }
