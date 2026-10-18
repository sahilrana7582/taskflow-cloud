package project

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sahilrana7582/taskflow-cloud/internal/models"
	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
)

// keyPattern mirrors the projects_key_format CHECK constraint.
var keyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

// authorizer is the part of workspace.Service this package needs. Depending on
// a small interface keeps the packages decoupled and easy to fake in tests.
type authorizer interface {
	Authorize(ctx context.Context, workspaceID, userID uuid.UUID, allowed func(models.WorkspaceRole) bool) (models.WorkspaceRole, error)
}

type Service struct {
	repo       *Repository
	workspaces authorizer
}

func NewService(repo *Repository, workspaces authorizer) *Service {
	return &Service{
		repo:       repo,
		workspaces: workspaces,
	}
}

func (s *Service) Create(
	ctx context.Context,
	userID, workspaceID uuid.UUID,
	req CreateRequest,
) (*models.Project, error) {
	if _, err := s.workspaces.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanWrite); err != nil {
		return nil, err
	}

	key := strings.ToUpper(strings.TrimSpace(req.Key))
	if !keyPattern.MatchString(key) {
		return nil, utils.Invalid("key must be 2 to 10 letters or digits starting with a letter, like ENG")
	}

	name, err := utils.RequiredText("name", req.Name, 100)
	if err != nil {
		return nil, err
	}

	description, err := utils.OptionalText("description", req.Description, 5000)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, models.Project{
		WorkspaceID: workspaceID,
		Key:         key,
		Name:        name,
		Description: description,
		CreatedBy:   userID,
	})
}

func (s *Service) List(ctx context.Context, userID, workspaceID uuid.UUID) ([]models.Project, error) {
	if _, err := s.workspaces.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanRead); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, workspaceID)
}

func (s *Service) Get(ctx context.Context, userID, workspaceID, projectID uuid.UUID) (*models.Project, error) {
	if _, err := s.workspaces.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanRead); err != nil {
		return nil, err
	}
	return s.repo.Get(ctx, workspaceID, projectID)
}

func (s *Service) Update(
	ctx context.Context,
	userID, workspaceID, projectID uuid.UUID,
	req UpdateRequest,
) (*models.Project, error) {
	if _, err := s.workspaces.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanWrite); err != nil {
		return nil, err
	}

	p, err := s.repo.Get(ctx, workspaceID, projectID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if p.Name, err = utils.RequiredText("name", *req.Name, 100); err != nil {
			return nil, err
		}
	}

	if p.Description, err = utils.OptionalText("description", req.Description.Apply(p.Description), 5000); err != nil {
		return nil, err
	}

	if req.Archived != nil {
		switch {
		case *req.Archived && p.ArchivedAt == nil:
			now := time.Now()
			p.ArchivedAt = &now
		case !*req.Archived:
			p.ArchivedAt = nil
		}
	}

	return s.repo.Update(ctx, *p)
}

// Delete soft-deletes the project and its jobs. Only owners and admins may.
func (s *Service) Delete(ctx context.Context, userID, workspaceID, projectID uuid.UUID) error {
	if _, err := s.workspaces.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanManage); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, workspaceID, projectID)
}
