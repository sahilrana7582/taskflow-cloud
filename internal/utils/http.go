package utils

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
)

const maxBodyBytes = 1 << 20

type errorResponse struct {
	Error string `json:"error"`
}

// ListResponse wraps every list endpoint's result, so fields such as a
// pagination cursor can be added later without breaking clients.
type ListResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

// WriteError writes a ClientError with its matching status and message. Any
// other error is logged and hidden behind a generic 500, so internal details
// such as SQL errors never reach the client.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var clientErr *ClientError
	if !errors.As(err, &clientErr) {
		log.Printf("internal error: method=%s path=%s err=%v", r.Method, r.URL.Path, err)
		WriteJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
		})
		return
	}

	WriteJSON(w, statusFor(clientErr.Kind), errorResponse{
		Error: clientErr.Message,
	})
}

func statusFor(kind error) int {
	switch kind {
	case ErrInvalidInput:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrConflict:
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

// DecodeJSON reads exactly one JSON object from the request body into dst.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return Invalid("invalid request body")
	}

	// Reject additional JSON values after the first object.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Invalid("request must contain a single JSON object")
	}

	return nil
}

// PathUUID parses a UUID path parameter such as {workspaceID}.
func PathUUID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		return uuid.Nil, Invalid("invalid " + name)
	}
	return id, nil
}
