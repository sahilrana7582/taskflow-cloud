package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// RequiredText trims s and checks it is 1 to max characters long, matching
// the char_length CHECK constraints in the migrations.
func RequiredText(field, s string, max int) (string, error) {
	s = strings.TrimSpace(s)
	if n := utf8.RuneCountInString(s); n == 0 || n > max {
		return "", Invalid(fmt.Sprintf("%s must be between 1 and %d characters", field, max))
	}
	return s, nil
}

// OptionalText trims s and checks it is at most max characters long. Blank
// text becomes nil, so the column stores NULL instead of "".
func OptionalText(field string, s *string, max int) (*string, error) {
	if s == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > max {
		return nil, Invalid(fmt.Sprintf("%s must be at most %d characters", field, max))
	}
	return &trimmed, nil
}
