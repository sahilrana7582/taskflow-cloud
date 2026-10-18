package routes

import (
	"net/http"

	"github.com/sahilrana7582/taskflow-cloud/internal/auth"
)

func NewServeMux(authHandler *auth.AuthHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/register", authHandler.Register)

	return mux
}
