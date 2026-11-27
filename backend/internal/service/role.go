package service

import (
	"context"
	"strings"

	"backend/internal/repository"

	"github.com/google/uuid"
)

type RoleService struct {
	teams TeamRepository
	roles RoleRepository
	links RoleRelations
}

func NewRoleService(
	teams TeamRepository, roles RoleRepository, links RoleRelations,
) *RoleService {
	return &RoleService{teams: teams, roles: roles, links: links}
}

// AddRole creates a role in teamID: the postgres row first, then role#team and
// team#role in SpiceDB. If the SpiceDB write fails, the row is deleted again.
func (s *RoleService) AddRole(
	ctx context.Context, teamID uuid.UUID, name string,
) (repository.Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return repository.Role{}, ErrEmptyName
	}

	team, err := s.teams.Get(ctx, teamID)
	if err != nil {
		return repository.Role{}, err
	}
	if team.Archived() {
		return repository.Role{}, ErrTeamArchived
	}

	role, err := s.roles.Create(ctx, uuid.New(), name)
	if err != nil {
		return repository.Role{}, err
	}

	if err := s.links.LinkTeam(ctx, role.ID, teamID); err != nil {
		return repository.Role{}, rollback(ctx, err, func(ctx context.Context) error {
			return s.roles.Delete(ctx, role.ID)
		})
	}
	return role, nil
}
