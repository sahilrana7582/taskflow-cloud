package workspace

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/sahilrana7582/taskflow-cloud/internal/models"
	"github.com/sahilrana7582/taskflow-cloud/internal/utils"
)

// slugPattern mirrors the workspaces_slug_format CHECK constraint.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// Authorize checks that the user's role in the workspace passes allowed, for
// example models.WorkspaceRole.CanWrite. Non-members get "workspace not
// found" rather than "forbidden", so workspace IDs cannot be probed. The
// project and job services call this before touching any data.
func (s *Service) Authorize(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	allowed func(models.WorkspaceRole) bool,
) (models.WorkspaceRole, error) {
	role, err := s.repo.MemberRole(ctx, workspaceID, userID)
	if err != nil {
		return "", err
	}
	if !allowed(role) {
		return "", utils.Forbidden("your role in this workspace does not allow this action")
	}
	return role, nil
}

// IsMember reports whether the user belongs to the workspace.
func (s *Service) IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	_, err := s.repo.MemberRole(ctx, workspaceID, userID)
	if errors.Is(err, utils.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateRequest) (*WorkspaceResponse, error) {
	name, err := utils.RequiredText("name", req.Name, 100)
	if err != nil {
		return nil, err
	}

	slug, err := validateSlug(req.Slug)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, name, slug, userID)
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]WorkspaceResponse, error) {
	return s.repo.ListForUser(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID, workspaceID uuid.UUID) (*WorkspaceResponse, error) {
	return s.repo.GetForUser(ctx, workspaceID, userID)
}

func (s *Service) Update(
	ctx context.Context,
	userID, workspaceID uuid.UUID,
	req UpdateRequest,
) (*WorkspaceResponse, error) {
	role, err := s.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanManage)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		name, err := utils.RequiredText("name", *req.Name, 100)
		if err != nil {
			return nil, err
		}
		req.Name = &name
	}

	if req.Slug != nil {
		slug, err := validateSlug(*req.Slug)
		if err != nil {
			return nil, err
		}
		req.Slug = &slug
	}

	ws, err := s.repo.Update(ctx, workspaceID, req.Name, req.Slug)
	if err != nil {
		return nil, err
	}
	return &WorkspaceResponse{Workspace: *ws, Role: role}, nil
}

// Delete soft-deletes the workspace. Only its owner may do this.
func (s *Service) Delete(ctx context.Context, userID, workspaceID uuid.UUID) error {
	if _, err := s.Authorize(ctx, workspaceID, userID, isOwner); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, workspaceID)
}

func (s *Service) ListMembers(ctx context.Context, userID, workspaceID uuid.UUID) ([]MemberResponse, error) {
	if _, err := s.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanRead); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, workspaceID)
}

func (s *Service) AddMember(
	ctx context.Context,
	userID, workspaceID uuid.UUID,
	req AddMemberRequest,
) (*MemberResponse, error) {
	if _, err := s.Authorize(ctx, workspaceID, userID, models.WorkspaceRole.CanManage); err != nil {
		return nil, err
	}

	// Each workspace has exactly one owner: the user who created it.
	if !req.Role.Valid() || req.Role == models.WorkspaceRoleOwner {
		return nil, utils.Invalid("role must be one of admin, member or viewer")
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if err := utils.ValidateEmail(email); err != nil {
		return nil, utils.Invalid("invalid email")
	}

	return s.repo.AddMember(ctx, workspaceID, email, req.Role, userID)
}

// RemoveMember removes memberID from the workspace. Managers can remove other
// members, anyone can remove themselves, and the owner can never be removed.
func (s *Service) RemoveMember(ctx context.Context, userID, workspaceID, memberID uuid.UUID) error {
	allowed := models.WorkspaceRole.CanManage
	if memberID == userID {
		allowed = models.WorkspaceRole.CanRead
	}
	if _, err := s.Authorize(ctx, workspaceID, userID, allowed); err != nil {
		return err
	}

	role, err := s.repo.MemberRole(ctx, workspaceID, memberID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return utils.NotFound("member not found")
		}
		return err
	}
	if role == models.WorkspaceRoleOwner {
		return utils.Forbidden("the workspace owner cannot be removed")
	}

	return s.repo.RemoveMember(ctx, workspaceID, memberID)
}

func isOwner(role models.WorkspaceRole) bool {
	return role == models.WorkspaceRoleOwner
}

func validateSlug(slug string) (string, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if len(slug) < 3 || len(slug) > 50 || !slugPattern.MatchString(slug) {
		return "", utils.Invalid("slug must be 3 to 50 lowercase letters, digits or single hyphens, like acme-corp")
	}
	return slug, nil
}
