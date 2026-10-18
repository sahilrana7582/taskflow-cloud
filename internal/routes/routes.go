package routes

import (
	"net/http"

	"github.com/sahilrana27582/taskflow/internal/auth"
)

func NewServeMux(authHandler *auth.AuthHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/register", authHandler.Register)

	return mux
}
