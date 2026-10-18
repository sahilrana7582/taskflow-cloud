package auth

import (
	"context"

	"github.com/google/uuid"
)

// claimsKey is the context key for verified claims. Because the type is
// unexported, no other package can read or overwrite the value by accident.
type claimsKey struct{}

// ContextWithClaims returns a copy of ctx carrying verified token claims. Only
// the auth middleware should call it, after VerifyToken succeeds.
func ContextWithClaims(ctx context.Context, claims *JwtClaims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

// ClaimsFromContext returns the claims verified for this request. It returns
// false if the route is not behind the auth middleware.
func ClaimsFromContext(ctx context.Context) (*JwtClaims, bool) {
	claims, ok := ctx.Value(claimsKey{}).(*JwtClaims)
	return claims, ok
}

// UserIDFromContext returns the ID of the authenticated user, ready to pass to
// repository methods.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return uuid.Nil, false
	}

	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
