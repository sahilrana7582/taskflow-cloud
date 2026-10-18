package job

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sahilrana7582/taskflow-cloud/internal/models"
	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

// workspaceAccess is the part of workspace.Service this package needs.
type workspaceAccess interface {
	Authorize(ctx context.Context, workspaceID, userID uuid.UUID, allowed func(models.WorkspaceRole) bool) (models.WorkspaceRole, error)
	IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error)
}

type Service struct {
	repo       *Repository
	workspaces workspaceAccess
}

func NewService(repo *Repository, workspaces workspaceAccess) *Service {
	return &Service{
		repo:       repo,
		workspaces: workspaces,
	}
}

func (s *Service) Create(
	ctx context.Context,
	userID, workspaceID, projectID uuid.UUID,
	req CreateRequest,
) (*models.Job, error) {
	if _, err := s.workspaces.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanWrite); err != nil {
		return nil, err
	}

	title, err := utils.RequiredText("title", req.Title, 200)
	if err != nil {
		return nil, err
	}

	description, err := utils.OptionalText("description", req.Description, 10000)
	if err != nil {
		return nil, err
	}

	priority := req.Priority
	if priority == "" {
		priority = models.JobPriorityMedium
	}
	if !priority.Valid() {
		return nil, invalidPriority()
	}

	if req.AssigneeID != nil {
		if err := s.checkAssignee(ctx, workspaceID, *req.AssigneeID); err != nil {
			return nil, err
		}
	}

	return s.repo.Create(ctx, models.Job{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Priority:    priority,
		ReporterID:  userID,
		AssigneeID:  req.AssigneeID,
		DueAt:       req.DueAt,
	})
}

// List returns one page of a project's jobs, newest first. Pass the returned
// NextCursor back as params.Cursor to get the following page.
func (s *Service) List(
	ctx context.Context,
	userID, workspaceID, projectID uuid.UUID,
	params ListParams,
) (*utils.ListResponse[models.Job], error) {
	if _, err := s.workspaces.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanRead); err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit == 0 {
		limit = defaultPageSize
	}
	if limit < 1 || limit > maxPageSize {
		return nil, utils.Invalid("limit must be between 1 and 100")
	}

	if params.Status != nil && !params.Status.Valid() {
		return nil, invalidStatus()
	}

	after, err := decodeCursor(params.Cursor)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.ProjectExists(ctx, workspaceID, projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, utils.NotFound("project not found")
	}

	// Fetch one extra row to learn whether another page exists.
	jobs, err := s.repo.List(ctx, workspaceID, projectID, listFilter{
		Status: params.Status,
		After:  after,
		Limit:  limit + 1,
	})
	if err != nil {
		return nil, err
	}

	resp := &utils.ListResponse[models.Job]{Items: jobs}
	if len(jobs) > limit {
		resp.Items = jobs[:limit]
		resp.NextCursor = encodeCursor(jobs[limit-1])
	}
	return resp, nil
}

func (s *Service) Get(ctx context.Context, userID, workspaceID, jobID uuid.UUID) (*models.Job, error) {
	if _, err := s.workspaces.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanRead); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, workspaceID, jobID)
}

func (s *Service) Update(
	ctx context.Context,
	userID, workspaceID, jobID uuid.UUID,
	req UpdateRequest,
) (*models.Job, error) {
	if _, err := s.workspaces.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanWrite); err != nil {
		return nil, err
	}

	if req.Version < 1 {
		return nil, utils.Invalid("version is required: send the version from your last read of the job")
	}

	j, err := s.repo.Get(ctx, workspaceID, jobID)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		if j.Title, err = utils.RequiredText("title", *req.Title, 200); err != nil {
			return nil, err
		}
	}

	if j.Description, err = utils.OptionalText("description", req.Description.Apply(j.Description), 10000); err != nil {
		return nil, err
	}

	if req.Status != nil {
		if !req.Status.Valid() {
			return nil, invalidStatus()
		}
		j.Status = *req.Status
	}

	if req.Priority != nil {
		if !req.Priority.Valid() {
			return nil, invalidPriority()
		}
		j.Priority = *req.Priority
	}

	if req.AssigneeID.Set && req.AssigneeID.Value != nil {
		if err := s.checkAssignee(ctx, workspaceID, *req.AssigneeID.Value); err != nil {
			return nil, err
		}
	}
	j.AssigneeID = req.AssigneeID.Apply(j.AssigneeID)
	j.DueAt = req.DueAt.Apply(j.DueAt)

	return s.repo.Update(ctx, *j, req.Version)
}

func (s *Service) Delete(ctx context.Context, userID, workspaceID, jobID uuid.UUID) error {
	if _, err := s.workspaces.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanWrite); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, workspaceID, jobID)
}

// checkAssignee stops jobs from being assigned to someone outside the
// workspace. The foreign key only proves that the user exists.
func (s *Service) checkAssignee(ctx context.Context, workspaceID, assigneeID uuid.UUID) error {
	ok, err := s.workspaces.IsMember(ctx, workspaceID, assigneeID)
	if err != nil {
		return err
	}
	if !ok {
		return utils.Invalid("assignee must be a member of the workspace")
	}
	return nil
}

func invalidStatus() error {
	return utils.Invalid("status must be one of todo, in_progress, blocked, done or cancelled")
}

func invalidPriority() error {
	return utils.Invalid("priority must be one of low, medium, high or urgent")
}

// cursor marks the last job of the previous page.
type cursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

// encodeCursor makes an opaque page token, so clients cannot depend on what
// it contains and the format can change later.
func encodeCursor(j models.Job) string {
	raw := j.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + j.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(s string) (*cursor, error) {
	if s == "" {
		return nil, nil
	}

	invalid := utils.Invalid("invalid cursor")

	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, invalid
	}

	createdAt, id, ok := strings.Cut(string(raw), "|")
	if !ok {
		return nil, invalid
	}

	c := &cursor{}
	if c.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt); err != nil {
		return nil, invalid
	}
	if c.ID, err = uuid.Parse(id); err != nil {
		return nil, invalid
	}
	return c, nil
}
