package models

import (
	"time"

	"github.com/google/uuid"
)

// Project groups jobs inside a workspace. Key is the short prefix used in
// job references, e.g. "ENG" in ENG-42.
type Project struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	WorkspaceID   uuid.UUID  `db:"workspace_id" json:"workspace_id"`
	Key           string     `db:"key" json:"key"`
	Name          string     `db:"name" json:"name"`
	Description   *string    `db:"description" json:"description,omitempty"`
	NextJobNumber int        `db:"next_job_number" json:"-"`
	CreatedBy     uuid.UUID  `db:"created_by" json:"created_by"`
	ArchivedAt    *time.Time `db:"archived_at" json:"archived_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at" json:"-"`
}
