package auth

import "github.com/google/uuid"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type UserResponse struct {
	ID     uuid.UUID  `json:"id"`
	Email  string     `json:"email"`
	Status UserStatus `json:"status"`
}
