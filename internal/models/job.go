package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusTodo       JobStatus = "todo"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusBlocked    JobStatus = "blocked"
	JobStatusDone       JobStatus = "done"
	JobStatusCancelled  JobStatus = "cancelled"
)

func (s JobStatus) Valid() bool {
	switch s {
	case JobStatusTodo, JobStatusInProgress, JobStatusBlocked, JobStatusDone, JobStatusCancelled:
		return true
	}
	return false
}

type JobPriority string

const (
	JobPriorityLow    JobPriority = "low"
	JobPriorityMedium JobPriority = "medium"
	JobPriorityHigh   JobPriority = "high"
	JobPriorityUrgent JobPriority = "urgent"
)

func (p JobPriority) Valid() bool {
	switch p {
	case JobPriorityLow, JobPriorityMedium, JobPriorityHigh, JobPriorityUrgent:
		return true
	}
	return false
}

// Job is a unit of work inside a project. CompletedAt is set exactly when
// Status is done (enforced by a CHECK constraint). Version is used for
// optimistic locking on updates.
type Job struct {
	ID          uuid.UUID   `db:"id" json:"id"`
	WorkspaceID uuid.UUID   `db:"workspace_id" json:"workspace_id"`
	ProjectID   uuid.UUID   `db:"project_id" json:"project_id"`
	Number      int         `db:"number" json:"number"`
	Title       string      `db:"title" json:"title"`
	Description *string     `db:"description" json:"description,omitempty"`
	Status      JobStatus   `db:"status" json:"status"`
	Priority    JobPriority `db:"priority" json:"priority"`
	ReporterID  uuid.UUID   `db:"reporter_id" json:"reporter_id"`
	AssigneeID  *uuid.UUID  `db:"assignee_id" json:"assignee_id,omitempty"`
	DueAt       *time.Time  `db:"due_at" json:"due_at,omitempty"`
	CompletedAt *time.Time  `db:"completed_at" json:"completed_at,omitempty"`
	Version     int         `db:"version" json:"version"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time  `db:"deleted_at" json:"-"`
}

// Reference returns the human-readable ID, e.g. "ENG-42".
func (j Job) Reference(projectKey string) string {
	return fmt.Sprintf("%s-%d", projectKey, j.Number)
}

type JobComment struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	JobID     uuid.UUID  `db:"job_id" json:"job_id"`
	AuthorID  uuid.UUID  `db:"author_id" json:"author_id"`
	Body      string     `db:"body" json:"body"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"-"`
}
