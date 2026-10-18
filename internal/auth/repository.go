package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sahilrana27582/taskflow/internal/utils"
)

// userColumns is the column list every query returns, in the order scanUser reads them.
const userColumns = `
	id, email, password_hash, full_name, status,
	email_verified_at, last_login_at, created_at, updated_at`

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create inserts a new active user. It returns utils.ErrEmailExists if a live
// account already uses the email.
func (r *UserRepository) Create(
	ctx context.Context,
	email string,
	passwordHash string,
	fullName string,
) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash, full_name)
		VALUES ($1, $2, $3)
		RETURNING ` + userColumns

	row := r.db.QueryRow(ctx, query, normalizeEmail(email), passwordHash, strings.TrimSpace(fullName))

	user, err := scanUser(row)
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return nil, utils.ErrEmailExists
		}

		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

// GetByID returns a live (not soft-deleted) user.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT ` + userColumns + `
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	user, err := scanUser(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail returns a live user by email. Used by login.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT ` + userColumns + `
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	user, err := scanUser(r.db.QueryRow(ctx, query, normalizeEmail(email)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

// UpdateLastLogin records a successful sign-in.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	return r.exec(ctx, "update last login", query, id)
}

// UpdatePassword replaces the stored password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $2
		WHERE id = $1 AND deleted_at IS NULL`

	return r.exec(ctx, "update password", query, id, passwordHash)
}

// MarkEmailVerified sets email_verified_at the first time it is called and
// leaves it unchanged after that.
func (r *UserRepository) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET email_verified_at = COALESCE(email_verified_at, NOW())
		WHERE id = $1 AND deleted_at IS NULL`

	return r.exec(ctx, "mark email verified", query, id)
}

// UpdateStatus activates or suspends a user.
func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status UserStatus) error {
	if !status.Valid() {
		return fmt.Errorf("update status: invalid user status %q", status)
	}

	query := `
		UPDATE users
		SET status = $2
		WHERE id = $1 AND deleted_at IS NULL`

	return r.exec(ctx, "update status", query, id, status)
}

// SoftDelete hides the user from every query above and frees their email for
// a new account. The row is kept for audit history.
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	return r.exec(ctx, "delete user", query, id)
}

// exec runs an UPDATE on a single user and returns utils.ErrUserNotFound if no
// live user has that id.
func (r *UserRepository) exec(ctx context.Context, op string, query string, args ...any) error {
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if tag.RowsAffected() == 0 {
		return utils.ErrUserNotFound
	}

	return nil
}

func scanUser(row pgx.Row) (*User, error) {
	var user User

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Status,
		&user.EmailVerifiedAt,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// normalizeEmail matches the users_email_format check, which only accepts
// lowercase emails.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
