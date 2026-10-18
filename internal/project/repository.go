package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sahilrana7582/taskflow-cloud/internal/models"
	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
)

// projectColumns matches models.Project field for field, so rows scan with
// pgx.RowToStructByName.
const projectColumns = `
	id, workspace_id, key, name, description, next_job_number,
	created_by, archived_at, created_at, updated_at, deleted_at`

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, p models.Project) (*models.Project, error) {
	rows, _ := r.db.Query(ctx, `
		INSERT INTO projects (workspace_id, key, name, description, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+projectColumns,
		p.WorkspaceID, p.Key, p.Name, p.Description, p.CreatedBy)

	created, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.Project])
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return nil, utils.Conflict("a project with this key already exists in the workspace")
		}
		return nil, fmt.Errorf("create project: %w", err)
	}
	return created, nil
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID) ([]models.Project, error) {
	rows, _ := r.db.Query(ctx, `
		SELECT `+projectColumns+`
		FROM projects
		WHERE workspace_id = $1 AND deleted_at IS NULL
		ORDER BY name`,
		workspaceID)

	projects, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Project])
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projects, nil
}

// Get returns a live project. Filtering on workspace_id as well as id is what
// keeps one tenant from reading another tenant's project by guessing its ID.
func (r *Repository) Get(ctx context.Context, workspaceID, projectID uuid.UUID) (*models.Project, error) {
	rows, _ := r.db.Query(ctx, `
		SELECT `+projectColumns+`
		FROM projects
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
		projectID, workspaceID)

	p, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.Project])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.NotFound("project not found")
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	return p, nil
}

// Update saves the editable fields of p.
func (r *Repository) Update(ctx context.Context, p models.Project) (*models.Project, error) {
	rows, _ := r.db.Query(ctx, `
		UPDATE projects
		SET name = $3, description = $4, archived_at = $5
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING `+projectColumns,
		p.ID, p.WorkspaceID, p.Name, p.Description, p.ArchivedAt)

	updated, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.Project])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.NotFound("project not found")
		}
		return nil, fmt.Errorf("update project: %w", err)
	}
	return updated, nil
}

// SoftDelete hides the project and all of its jobs in one transaction.
func (r *Repository) SoftDelete(ctx context.Context, workspaceID, projectID uuid.UUID) error {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE projects
			SET deleted_at = NOW()
			WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
			projectID, workspaceID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return utils.NotFound("project not found")
		}

		_, err = tx.Exec(ctx, `
			UPDATE jobs
			SET deleted_at = NOW()
			WHERE project_id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
			projectID, workspaceID)
		return err
	})
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}
