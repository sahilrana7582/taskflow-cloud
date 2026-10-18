package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
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

// RequireUserID returns the authenticated user's ID for handlers behind the
// auth middleware. If there is none, it writes a 401 and returns false.
func RequireUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, ok := UserIDFromContext(r.Context())
	if !ok {
		utils.WriteError(w, r, utils.Unauthorized("authentication required"))
	}
	return id, ok
}
