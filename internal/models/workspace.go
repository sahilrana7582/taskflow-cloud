package models

import (
	"time"

	"github.com/google/uuid"
)

// Workspace is the tenant boundary. Every project and job belongs to one.
type Workspace struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Slug      string     `db:"slug" json:"slug"`
	CreatedBy uuid.UUID  `db:"created_by" json:"created_by"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"-"`
}

type WorkspaceRole string

const (
	WorkspaceRoleOwner  WorkspaceRole = "owner"
	WorkspaceRoleAdmin  WorkspaceRole = "admin"
	WorkspaceRoleMember WorkspaceRole = "member"
	WorkspaceRoleViewer WorkspaceRole = "viewer"
)

func (r WorkspaceRole) Valid() bool {
	switch r {
	case WorkspaceRoleOwner, WorkspaceRoleAdmin, WorkspaceRoleMember, WorkspaceRoleViewer:
		return true
	}
	return false
}

// CanWrite reports whether the role may create and edit projects, jobs and comments.
func (r WorkspaceRole) CanWrite() bool {
	return r == WorkspaceRoleOwner || r == WorkspaceRoleAdmin || r == WorkspaceRoleMember
}

// CanManage reports whether the role may manage members and workspace settings.
func (r WorkspaceRole) CanManage() bool {
	return r == WorkspaceRoleOwner || r == WorkspaceRoleAdmin
}

type WorkspaceMember struct {
	WorkspaceID uuid.UUID     `db:"workspace_id" json:"workspace_id"`
	UserID      uuid.UUID     `db:"user_id" json:"user_id"`
	Role        WorkspaceRole `db:"role" json:"role"`
	InvitedBy   *uuid.UUID    `db:"invited_by" json:"invited_by,omitempty"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
}
