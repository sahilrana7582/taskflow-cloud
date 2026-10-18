package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sahilrana7582/taskflow-cloud/internal/auth"
)

// Auth rejects any request without a valid "Authorization: Bearer <token>"
// header. For valid requests it stores the verified claims on the request
// context, where handlers read them with auth.ClaimsFromContext or
// auth.UserIDFromContext.
func Auth(tokens *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				// No error code when credentials are missing (RFC 6750 section 3.1).
				w.Header().Set("WWW-Authenticate", "Bearer")
				writeError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}

			claims, err := tokens.VerifyToken(token)
			if err != nil {
				// Telling expired apart lets clients know to log in again
				// rather than treat the token as forged.
				msg := "invalid token"
				if errors.Is(err, jwt.ErrTokenExpired) {
					msg = "token expired"
				}

				w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", error_description="`+msg+`"`)
				writeError(w, http.StatusUnauthorized, msg)
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.ContextWithClaims(r.Context(), claims)))
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	// Auth scheme names are case-insensitive (RFC 7235 section 2.1).
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}

	token = strings.TrimSpace(token)
	return token, token != ""
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}
