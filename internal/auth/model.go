package auth

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
)

func (s UserStatus) Valid() bool {
	switch s {
	case UserStatusActive, UserStatusSuspended:
		return true
	}
	return false
}

// User is an account that can sign in and belong to workspaces.
type User struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	Email           string     `db:"email" json:"email"`
	PasswordHash    string     `db:"password_hash" json:"-"`
	FullName        string     `db:"full_name" json:"full_name"`
	Status          UserStatus `db:"status" json:"status"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at" json:"-"`
}

// Session is a refresh-token session. TokenHash is the SHA-256 of the token
// handed to the client; the raw token is never stored.
type Session struct {
	ID         uuid.UUID     `db:"id" json:"id"`
	UserID     uuid.UUID     `db:"user_id" json:"user_id"`
	TokenHash  []byte        `db:"token_hash" json:"-"`
	UserAgent  *string       `db:"user_agent" json:"user_agent,omitempty"`
	IPAddress  *netip.Prefix `db:"ip_address" json:"ip_address,omitempty"`
	ExpiresAt  time.Time     `db:"expires_at" json:"expires_at"`
	RevokedAt  *time.Time    `db:"revoked_at" json:"revoked_at,omitempty"`
	LastUsedAt *time.Time    `db:"last_used_at" json:"last_used_at,omitempty"`
	CreatedAt  time.Time     `db:"created_at" json:"created_at"`
}

// Active reports whether the session can still be used to refresh tokens.
func (s Session) Active(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}
