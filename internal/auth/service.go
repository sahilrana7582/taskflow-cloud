package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
)

// maxPasswordBytes is bcrypt's input limit; longer passwords make it error.
const maxPasswordBytes = 72

// dummyPasswordHash is checked against when no user has the given email, so
// that login takes as long as it does for a wrong password and the response
// time does not reveal which emails are registered.
var dummyPasswordHash = sync.OnceValue(func() string {
	hash, err := HashPassword("taskflow-dummy-password")
	if err != nil {
		panic(err) // unreachable: the input is a fixed, valid password
	}
	return hash
})

type AuthService struct {
	users  *UserRepository
	tokens *TokenManager
}

func NewAuthService(users *UserRepository, tokens *TokenManager) *AuthService {
	return &AuthService{
		users:  users,
		tokens: tokens,
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

// Login checks the email and password and returns a signed access token. An
// unknown email and a wrong password both return utils.ErrInvalidCredentials,
// so the response never reveals whether an account exists.
func (s *AuthService) Login(
	ctx context.Context,
	req LoginRequest,
) (*LoginResponse, error) {
	// No stored password can be empty or longer than bcrypt accepts. Only the
	// bcrypt limits are checked here, not the register rules, so tightening
	// those later does not lock out existing users.
	if req.Password == "" || len(req.Password) > maxPasswordBytes {
		return nil, utils.ErrInvalidCredentials
	}

	user, err := s.users.GetByEmail(ctx, req.Email)
	if errors.Is(err, utils.ErrUserNotFound) {
		_, _ = VerifyPassword(req.Password, dummyPasswordHash())
		return nil, utils.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	ok, err := VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, utils.ErrInvalidCredentials
	}

	// Checked only after the password, so suspension is never revealed to
	// someone who does not know it.
	if user.Status != UserStatusActive {
		return nil, utils.ErrAccountSuspended
	}

	token, expiresAt, err := s.tokens.GenerateToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	if err := s.users.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("record login: %w", err)
	}

	return &LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User: UserResponse{
			ID:     user.ID,
			Email:  user.Email,
			Status: user.Status,
		},
	}, nil
}
