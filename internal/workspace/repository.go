package workspace

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

// workspaceColumns matches models.Workspace field for field, so rows scan with
// pgx.RowToStructByName. Queries alias the workspaces table as w.
const workspaceColumns = `
	w.id, w.name, w.slug, w.created_by, w.created_at, w.updated_at, w.deleted_at`

const memberColumns = `
	m.user_id, u.email, u.full_name, m.role, m.created_at AS joined_at`

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

// Create inserts the workspace and makes ownerID its owner in one transaction,
// so a workspace can never exist without an owner.
func (r *Repository) Create(ctx context.Context, name, slug string, ownerID uuid.UUID) (*WorkspaceResponse, error) {
	var created *WorkspaceResponse

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		rows, _ := tx.Query(ctx, `
			INSERT INTO workspaces AS w (name, slug, created_by)
			VALUES ($1, $2, $3)
			RETURNING `+workspaceColumns,
			name, slug, ownerID)

		ws, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Workspace])
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO workspace_members (workspace_id, user_id, role)
			VALUES ($1, $2, $3)`,
			ws.ID, ownerID, models.WorkspaceRoleOwner); err != nil {
			return err
		}

		created = &WorkspaceResponse{Workspace: ws, Role: models.WorkspaceRoleOwner}
		return nil
	})
	if err != nil {
		if utils.IsUniqueViolation(err) {
			return nil, utils.Conflict("workspace slug is already taken")
		}
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	return created, nil
}

// ListForUser returns every live workspace the user belongs to.
func (r *Repository) ListForUser(ctx context.Context, userID uuid.UUID) ([]WorkspaceResponse, error) {
	rows, _ := r.db.Query(ctx, `
		SELECT `+workspaceColumns+`, m.role
		FROM workspaces w
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE m.user_id = $1 AND w.deleted_at IS NULL
		ORDER BY w.name`,
		userID)

	workspaces, err := pgx.CollectRows(rows, pgx.RowToStructByName[WorkspaceResponse])
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	return workspaces, nil
}

// GetForUser returns the workspace if the user is a member of it.
func (r *Repository) GetForUser(ctx context.Context, workspaceID, userID uuid.UUID) (*WorkspaceResponse, error) {
	rows, _ := r.db.Query(ctx, `
		SELECT `+workspaceColumns+`, m.role
		FROM workspaces w
		JOIN workspace_members m ON m.workspace_id = w.id
		WHERE w.id = $1 AND m.user_id = $2 AND w.deleted_at IS NULL`,
		workspaceID, userID)

	ws, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[WorkspaceResponse])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.NotFound("workspace not found")
		}
		return nil, fmt.Errorf("get workspace: %w", err)
	}
	return ws, nil
}

// MemberRole returns the user's role in a live workspace.
func (r *Repository) MemberRole(ctx context.Context, workspaceID, userID uuid.UUID) (models.WorkspaceRole, error) {
	var role models.WorkspaceRole

	err := r.db.QueryRow(ctx, `
		SELECT m.role
		FROM workspace_members m
		JOIN workspaces w ON w.id = m.workspace_id
		WHERE m.workspace_id = $1 AND m.user_id = $2 AND w.deleted_at IS NULL`,
		workspaceID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", utils.NotFound("workspace not found")
		}
		return "", fmt.Errorf("get member role: %w", err)
	}
	return role, nil
}

// Update changes the name and/or slug. A nil argument keeps the current value.
func (r *Repository) Update(ctx context.Context, workspaceID uuid.UUID, name, slug *string) (*models.Workspace, error) {
	rows, _ := r.db.Query(ctx, `
		UPDATE workspaces AS w
		SET name = COALESCE($2, w.name),
		    slug = COALESCE($3, w.slug)
		WHERE w.id = $1 AND w.deleted_at IS NULL
		RETURNING `+workspaceColumns,
		workspaceID, name, slug)

	ws, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.Workspace])
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, utils.NotFound("workspace not found")
		case utils.IsUniqueViolation(err):
			return nil, utils.Conflict("workspace slug is already taken")
		}
		return nil, fmt.Errorf("update workspace: %w", err)
	}
	return ws, nil
}

// SoftDelete hides the workspace, and with it every project and job inside,
// because all access goes through a membership check on a live workspace.
func (r *Repository) SoftDelete(ctx context.Context, workspaceID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE workspaces
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`,
		workspaceID)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return utils.NotFound("workspace not found")
	}
	return nil
}

func (r *Repository) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]MemberResponse, error) {
	rows, _ := r.db.Query(ctx, `
		SELECT `+memberColumns+`
		FROM workspace_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.workspace_id = $1 AND u.deleted_at IS NULL
		ORDER BY m.created_at`,
		workspaceID)

	members, err := pgx.CollectRows(rows, pgx.RowToStructByName[MemberResponse])
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	return members, nil
}

// AddMember adds the live user with the given email. Looking the user up and
// inserting happen in one statement, so there is no gap between the two.
func (r *Repository) AddMember(
	ctx context.Context,
	workspaceID uuid.UUID,
	email string,
	role models.WorkspaceRole,
	invitedBy uuid.UUID,
) (*MemberResponse, error) {
	rows, _ := r.db.Query(ctx, `
		WITH m AS (
			INSERT INTO workspace_members (workspace_id, user_id, role, invited_by)
			-- Casts are required: untyped parameters in a SELECT list default to text.
			SELECT $1::uuid, u.id, $3::text, $4::uuid
			FROM users u
			WHERE u.email = $2 AND u.deleted_at IS NULL
			RETURNING user_id, role, created_at
		)
		SELECT `+memberColumns+`
		FROM m
		JOIN users u ON u.id = m.user_id`,
		workspaceID, email, role, invitedBy)

	member, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[MemberResponse])
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, utils.NotFound("no user with that email")
		case utils.IsUniqueViolation(err):
			return nil, utils.Conflict("user is already a member of this workspace")
		}
		return nil, fmt.Errorf("add member: %w", err)
	}
	return member, nil
}

func (r *Repository) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return utils.NotFound("member not found")
	}
	return nil
}
