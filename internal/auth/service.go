package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
)

type AuthService struct {
	users *UserRepository
}

func NewAuthService(users *UserRepository) *AuthService {
	return &AuthService{
		users: users,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	req RegisterRequest,
) (*UserResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if err := utils.ValidateEmail(email); err != nil {
		return nil, err
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	fullName := strings.TrimSpace(req.FullName)

	if err := utils.ValidateField(fullName); err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, email, passwordHash, fullName)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:     user.ID,
		Email:  user.Email,
		Status: user.Status,
	}, nil
}
