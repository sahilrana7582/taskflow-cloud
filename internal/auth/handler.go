package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
)

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req RegisterRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrInvalidEmail):
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: "invalid email",
			})

		case errors.Is(err, utils.ErrInvalidPassword):
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: "password must be between 8 and 72 bytes",
			})

		case errors.Is(err, utils.ErrEmailExists):
			writeJSON(w, http.StatusConflict, errorResponse{
				Error: "email already exists",
			})

		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
			})
		}
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req LoginRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrInvalidCredentials):
			writeJSON(w, http.StatusUnauthorized, errorResponse{
				Error: "invalid email or password",
			})

		case errors.Is(err, utils.ErrAccountSuspended):
			writeJSON(w, http.StatusForbidden, errorResponse{
				Error: "account suspended",
			})

		default:
			log.Printf("login: %v", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
			})
		}
		return
	}

	// Responses carrying tokens must never be cached (RFC 6749 section 5.1).
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, resp)
}

// requireMethod writes a 405 response and returns false unless the request
// uses the given method.
func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}

	w.Header().Set("Allow", method)
	writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
		Error: "method not allowed",
	})
	return false
}

// decodeJSON reads exactly one JSON object from the request body into dst. On
// failure it writes a 400 response and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return false
	}

	// Reject additional JSON values after the first object.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "request must contain a single JSON object",
		})
		return false
	}

	return true
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
