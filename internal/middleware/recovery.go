package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf(
					"panic recovered: request_id=%s method=%s path=%s panic=%v\n%s",
					RequestIDFromContext(r.Context()),
					r.Method,
					r.URL.Path,
					recovered,
					debug.Stack(),
				)

				http.Error(
					w,
					`{"error":"internal server error"}`,
					http.StatusInternalServerError,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
