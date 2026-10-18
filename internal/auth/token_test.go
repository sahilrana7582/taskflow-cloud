package auth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecretKey = "test-secret-key-at-least-32-bytes"
	testIssuer    = "taskflow-cloud"
)

func newTestTokenManager(t *testing.T, secretKey, issuer string) *TokenManager {
	t.Helper()

	m, err := NewTokenManager(secretKey, issuer, time.Hour)
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}
	return m
}

func mustGenerateToken(t *testing.T, m *TokenManager, userID string) string {
	t.Helper()

	token, _, err := m.GenerateToken(userID, "ada@example.com")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	return token
}

func TestGenerateAndVerifyToken(t *testing.T) {
	m := newTestTokenManager(t, testSecretKey, testIssuer)

	token, expiresAt, err := m.GenerateToken("user-123", "ada@example.com")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := m.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}
	if claims.UserID != "user-123" || claims.Email != "ada@example.com" || claims.Issuer != testIssuer {
		t.Fatalf("VerifyToken() claims = %+v, want user-123 / ada@example.com from %s", claims, testIssuer)
	}
	if !claims.ExpiresAt.Time.Equal(expiresAt) {
		t.Fatalf("GenerateToken() expiresAt = %v, token says %v", expiresAt, claims.ExpiresAt.Time)
	}
}

func TestVerifyTokenRejectsInvalidTokens(t *testing.T) {
	m := newTestTokenManager(t, testSecretKey, testIssuer)

	expired := newTestTokenManager(t, testSecretKey, testIssuer)
	expired.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }

	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, JwtClaims{UserID: "user-123"}).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign unsigned token: %v", err)
	}

	hs512, err := jwt.NewWithClaims(jwt.SigningMethodHS512, JwtClaims{UserID: "user-123"}).
		SignedString([]byte(testSecretKey))
	if err != nil {
		t.Fatalf("sign HS512 token: %v", err)
	}

	// Swap in another user's payload but keep the original signature.
	original := strings.Split(mustGenerateToken(t, m, "user-123"), ".")
	other := strings.Split(mustGenerateToken(t, m, "admin"), ".")
	tampered := original[0] + "." + other[1] + "." + original[2]

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{"expired", mustGenerateToken(t, expired, "user-123"), jwt.ErrTokenExpired},
		{"wrong secret", mustGenerateToken(t, newTestTokenManager(t, strings.Repeat("x", 32), testIssuer), "user-123"), jwt.ErrTokenSignatureInvalid},
		{"wrong issuer", mustGenerateToken(t, newTestTokenManager(t, testSecretKey, "someone-else"), "user-123"), jwt.ErrTokenInvalidIssuer},
		{"alg none", unsigned, jwt.ErrTokenSignatureInvalid},
		{"different algorithm", hs512, jwt.ErrTokenSignatureInvalid},
		{"tampered payload", tampered, jwt.ErrTokenSignatureInvalid},
		{"malformed", "not-a-jwt", jwt.ErrTokenMalformed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := m.VerifyToken(tt.token)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("VerifyToken() error = %v, want %v", err, tt.wantErr)
			}
			if claims != nil {
				t.Fatalf("VerifyToken() claims = %+v, want nil", claims)
			}
		})
	}
}

func TestNewTokenManagerRejectsShortSecret(t *testing.T) {
	if _, err := NewTokenManager("too-short", testIssuer, time.Hour); err == nil {
		t.Fatal("expected a secret shorter than 32 bytes to be rejected")
	}
}
