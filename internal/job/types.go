package job

import (
	"time"

	"github.com/google/uuid"
	"github.com/sahilrana7582/taskflow-cloud/internal/models"
	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
)

type CreateRequest struct {
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	Priority    models.JobPriority `json:"priority"` // defaults to medium
	AssigneeID  *uuid.UUID         `json:"assignee_id"`
	DueAt       *time.Time         `json:"due_at"`
}

// UpdateRequest changes only the fields that are present. Version must be the
// version the client last read: if someone else saved the job since, the
// update fails with 409 Conflict instead of silently overwriting their change.
type UpdateRequest struct {
	Version     int                       `json:"version"`
	Title       *string                   `json:"title"`
	Description utils.Optional[string]    `json:"description"`
	Status      *models.JobStatus         `json:"status"`
	Priority    *models.JobPriority       `json:"priority"`
	AssigneeID  utils.Optional[uuid.UUID] `json:"assignee_id"`
	DueAt       utils.Optional[time.Time] `json:"due_at"`
}

// ListParams come from the query string: ?status=todo&limit=20&cursor=...
type ListParams struct {
	Status *models.JobStatus
	Limit  int
	Cursor string
}
