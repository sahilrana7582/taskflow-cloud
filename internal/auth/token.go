package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// minSecretKeyLength is the smallest HS256 key RFC 7518 allows: 256 bits, the
// size of the SHA-256 output. Generate one with: openssl rand -base64 32
const minSecretKeyLength = 32

// clockSkew tolerates small clock differences between the instance that
// issued a token and the one verifying it.
const clockSkew = 5 * time.Second

type JwtClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`

	jwt.RegisteredClaims
}

// TokenManager issues and verifies HS256-signed access tokens. Create one at
// startup and share it; it is safe for concurrent use.
type TokenManager struct {
	secretKey []byte
	issuer    string
	ttl       time.Duration
	now       func() time.Time // replaced in tests
}

func NewTokenManager(secretKey, issuer string, ttl time.Duration) (*TokenManager, error) {
	if len(secretKey) < minSecretKeyLength {
		return nil, fmt.Errorf("JWT secret key must be at least %d bytes, got %d", minSecretKeyLength, len(secretKey))
	}
	if issuer == "" {
		return nil, errors.New("JWT issuer must not be empty")
	}
	if ttl <= 0 {
		return nil, errors.New("JWT lifetime must be positive")
	}

	return &TokenManager{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		ttl:       ttl,
		now:       time.Now,
	}, nil
}

// GenerateToken returns a signed access token for the user and the time it
// expires.
func (m *TokenManager) GenerateToken(userID, email string) (string, time.Time, error) {
	if userID == "" {
		return "", time.Time{}, errors.New("generate token: user ID must not be empty")
	}

	now := m.now()
	claims := JwtClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	// Report the expiry as encoded in the token, which is rounded to seconds.
	return token, claims.ExpiresAt.Time, nil
}

// VerifyToken checks a token's signature, algorithm, issuer and expiry and
// returns its claims. Errors wrap the jwt sentinels, so callers can tell an
// expired token apart with errors.Is(err, jwt.ErrTokenExpired).
func (m *TokenManager) VerifyToken(tokenString string) (*JwtClaims, error) {
	claims := &JwtClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims,
		func(*jwt.Token) (any, error) { return m.secretKey, nil },
		// Pinning the algorithm blocks "alg: none" and algorithm-swap attacks.
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(clockSkew),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}
	return claims, nil
}
