package job

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sahilrana7582/taskflow-cloud/internal/models"
	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
)

// jobColumns matches models.Job field for field, so rows scan with
// pgx.RowToStructByName.
const jobColumns = `
	id, workspace_id, project_id, number, title, description, status, priority,
	reporter_id, assignee_id, due_at, completed_at, version,
	created_at, updated_at, deleted_at`

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// Create takes the next job number from the project and inserts the job in
// one transaction. The UPDATE locks the project row, so concurrent creates in
// the same project wait their turn instead of getting the same number.
func (r *Repository) Create(ctx context.Context, j models.Job) (*models.Job, error) {
	var created *models.Job

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var archivedAt *time.Time
		err := tx.QueryRow(ctx, `
			UPDATE projects
			SET next_job_number = next_job_number + 1
			WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
			RETURNING next_job_number - 1, archived_at`,
			j.ProjectID, j.WorkspaceID).Scan(&j.Number, &archivedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return utils.NotFound("project not found")
		}
		if err != nil {
			return err
		}
		// Returning an error rolls back the counter increment as well.
		if archivedAt != nil {
			return utils.Conflict("cannot add jobs to an archived project")
		}

		rows, _ := tx.Query(ctx, `
			INSERT INTO jobs (
				workspace_id, project_id, number, title, description,
				priority, reporter_id, assignee_id, due_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING `+jobColumns,
			j.WorkspaceID, j.ProjectID, j.Number, j.Title, j.Description,
			j.Priority, j.ReporterID, j.AssigneeID, j.DueAt)

		created, err = pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.Job])
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}
	return created, nil
}

// listFilter selects one page of a project's jobs, newest first.
type listFilter struct {
	Status *models.JobStatus
	After  *cursor // start after this job; nil for the first page
	Limit  int
}

func (r *Repository) List(ctx context.Context, workspaceID, projectID uuid.UUID, f listFilter) ([]models.Job, error) {
	query := `
		SELECT ` + jobColumns + `
		FROM jobs
		WHERE workspace_id = $1 AND project_id = $2 AND deleted_at IS NULL`
	args := []any{workspaceID, projectID}

	if f.Status != nil {
		args = append(args, *f.Status)
		query += fmt.Sprintf(" AND status = $%d", len(args))
	}

	// Keyset pagination: the row comparison lets jobs_project_created_idx seek
	// straight to the page, where OFFSET would rescan every skipped row.
	if f.After != nil {
		args = append(args, f.After.CreatedAt, f.After.ID)
		query += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", len(args)-1, len(args))
	}

	args = append(args, f.Limit)
	query += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", len(args))

	rows, _ := r.db.Query(ctx, query, args...)
	jobs, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Job])
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	return jobs, nil
}

func (r *Repository) ProjectExists(ctx context.Context, workspaceID, projectID uuid.UUID) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM projects
			WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		)`,
		projectID, workspaceID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check project: %w", err)
	}
	return exists, nil
}

func (r *Repository) Get(ctx context.Context, workspaceID, jobID uuid.UUID) (*models.Job, error) {
	rows, _ := r.db.Query(ctx, `
		SELECT `+jobColumns+`
		FROM jobs
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
		jobID, workspaceID)

	j, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.Job])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.NotFound("job not found")
		}
		return nil, fmt.Errorf("get job: %w", err)
	}
	return j, nil
}

// Update saves the editable fields of j, but only if the stored version is
// still expectedVersion, and bumps the version. No matching row means someone
// else saved first, so the caller must reload and retry.
func (r *Repository) Update(ctx context.Context, j models.Job, expectedVersion int) (*models.Job, error) {
	rows, _ := r.db.Query(ctx, `
		UPDATE jobs
		SET title = $4,
		    description = $5,
		    status = $6,
		    priority = $7,
		    assignee_id = $8,
		    due_at = $9,
		    -- Keeps the original completion time if the job was already done.
		    completed_at = CASE WHEN $6 = 'done' THEN COALESCE(completed_at, NOW()) END,
		    version = version + 1
		WHERE id = $1 AND workspace_id = $2 AND version = $3 AND deleted_at IS NULL
		RETURNING `+jobColumns,
		j.ID, j.WorkspaceID, expectedVersion,
		j.Title, j.Description, j.Status, j.Priority, j.AssigneeID, j.DueAt)

	updated, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.Job])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.Conflict("the job was changed by someone else; reload it and try again")
		}
		return nil, fmt.Errorf("update job: %w", err)
	}
	return updated, nil
}

func (r *Repository) SoftDelete(ctx context.Context, workspaceID, jobID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE jobs
		SET deleted_at = NOW()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
		jobID, workspaceID)
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return utils.NotFound("job not found")
	}
	return nil
}
