package routes

import (
	"net/http"

	"github.com/sahilrana27582/taskflow/internal/handler"
)

func NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /jobs", handler.GetJob)

	return mux
}
