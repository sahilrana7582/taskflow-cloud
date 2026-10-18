package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sahilrana27582/taskflow/internal/utils"
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
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{
			Error: "method not allowed",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req RegisterRequest

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	// Reject additional JSON values after the first object.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "request must contain a single JSON object",
		})
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

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
