package workspace

import (
	"time"

	"github.com/google/uuid"
	"github.com/sahilrana7582/taskflow-cloud/internal/models"
)

type CreateRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// UpdateRequest changes only the fields that are present.
type UpdateRequest struct {
	Name *string `json:"name"`
	Slug *string `json:"slug"`
}

type AddMemberRequest struct {
	Email string               `json:"email"`
	Role  models.WorkspaceRole `json:"role"`
}

// WorkspaceResponse is a workspace together with the caller's role in it.
type WorkspaceResponse struct {
	models.Workspace
	Role models.WorkspaceRole `db:"role" json:"role"`
}

type MemberResponse struct {
	UserID   uuid.UUID            `db:"user_id" json:"user_id"`
	Email    string               `db:"email" json:"email"`
	FullName string               `db:"full_name" json:"full_name"`
	Role     models.WorkspaceRole `db:"role" json:"role"`
	JoinedAt time.Time            `db:"joined_at" json:"joined_at"`
}
