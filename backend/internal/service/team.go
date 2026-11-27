package service

import (
	"context"
	"strings"

	"backend/internal/repository"

	"github.com/google/uuid"
)

type TeamService struct {
	orgs  OrgRepository
	teams TeamRepository
	links TeamRelations
}

func NewTeamService(
	orgs OrgRepository, teams TeamRepository, links TeamRelations,
) *TeamService {
	return &TeamService{orgs: orgs, teams: teams, links: links}
}

// AddTeam creates a team in orgID: the postgres row first, then team#org in
// SpiceDB. If the SpiceDB write fails, the row is deleted again.
func (s *TeamService) AddTeam(
	ctx context.Context, orgID uuid.UUID, name string,
) (repository.Team, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return repository.Team{}, ErrEmptyName
	}

	org, err := s.orgs.Get(ctx, orgID)
	if err != nil {
		return repository.Team{}, err
	}
	if org.Archived() {
		return repository.Team{}, ErrOrgArchived
	}

	team, err := s.teams.Create(ctx, uuid.New(), name)
	if err != nil {
		return repository.Team{}, err
	}

	if err := s.links.LinkOrg(ctx, team.ID, orgID); err != nil {
		return repository.Team{}, rollback(ctx, err, func(ctx context.Context) error {
			return s.teams.Delete(ctx, team.ID)
		})
	}
	return team, nil
}
