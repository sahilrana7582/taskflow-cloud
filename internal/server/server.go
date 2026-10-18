package server

import (
	"net/http"
	"time"
)

func NewServer(port string, mux http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}
