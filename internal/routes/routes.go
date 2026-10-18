package routes

import (
	"net/http"

	"github.com/sahilrana7582/taskflow-cloud/internal/auth"
	"github.com/sahilrana7582/taskflow-cloud/internal/middleware"
)

func NewServeMux(authHandler *auth.AuthHandler, tokenManager *auth.TokenManager) *http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)

	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /me", meHandler)

	mux.Handle(
		"/api/",
		http.StripPrefix(
			"/api",
			middleware.Auth(tokenManager)(protectedMux),
		),
	)

	handler := middleware.RequestID(mux)
	handler = middleware.Recovery(handler)

	return &handler
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("Authenticated user: " + userID.String()))
}
