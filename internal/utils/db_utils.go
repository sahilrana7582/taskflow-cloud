package utils

import (
	"errors"
	"net/mail"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func ValidateEmail(email string) error {
	if len(email) == 0 || len(email) > 254 {
		return ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return ErrInvalidEmail
	}

	return nil
}

func ValidatePassword(password string) error {
	if !utf8.ValidString(password) {
		return ErrInvalidPassword
	}

	if len(password) < 8 || len(password) > 72 {
		return ErrInvalidPassword
	}

	return nil
}

func ValidateField(fieldData string) error {
	if len(fieldData) < 0 {
		return ErrInvalidPassword
	}

	return nil
}
