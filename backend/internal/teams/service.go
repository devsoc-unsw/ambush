package teams

import (
	"backend/db"
	"backend/internal/pgerr"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrParentNotFound = errors.New("parent team not found")
	ErrTeamNotFound   = errors.New("team not found")
	ErrRoleNotFound   = errors.New("role not found")
	ErrNameTaken      = errors.New("team name already taken in this org")
)

type Service struct {
	q *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{q: db.New(pool)}
}

type Team struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	ParentID     *uuid.UUID // nil = the org's root team
	LeaderRoleID *uuid.UUID // nil = no leader set
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ArchivedAt   *time.Time // nil = active
}

func toTeam(t db.Team) Team {
	out := Team{
		ID:        t.ID,
		OrgID:     t.OrgID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
	}
	if t.ParentID.Valid {
		out.ParentID = &t.ParentID.UUID
	}
	if t.LeaderRoleID.Valid {
		out.LeaderRoleID = &t.LeaderRoleID.UUID
	}
	if t.ArchivedAt.Valid {
		out.ArchivedAt = &t.ArchivedAt.Time
	}
	return out
}

type NewTeam struct {
	OrgID    uuid.UUID
	ParentID uuid.UUID
	Name     string
}

// CreateTeam adds a team under an existing one. Root teams are only made with
// their org, so every team here has a parent.
func (s *Service) CreateTeam(ctx context.Context, in NewTeam) (Team, error) {
	t, err := s.q.CreateTeam(ctx, db.CreateTeamParams{
		OrgID:    in.OrgID,
		ParentID: in.ParentID,
		Name:     in.Name,
	})
	// the insert selects the parent, so no row back means the parent isn't in
	// that org, or the parent or org is archived
	if errors.Is(err, pgx.ErrNoRows) {
		return Team{}, ErrParentNotFound
	}
	if pgerr.IsUniqueViolation(err) {
		return Team{}, ErrNameTaken
	}
	if err != nil {
		return Team{}, fmt.Errorf("create team: %w", err)
	}

	return toTeam(t), nil
}

// SetLeader sets which role leads a team, or clears it when roleID is nil.
// The role can be from any team in the same org.
func (s *Service) SetLeader(
	ctx context.Context, orgID, teamID uuid.UUID, roleID *uuid.UUID,
) (Team, error) {
	params := db.SetTeamLeaderParams{OrgID: orgID, TeamID: teamID}
	if roleID != nil {
		params.LeaderRoleID = uuid.NullUUID{UUID: *roleID, Valid: true}
	}

	t, err := s.q.SetTeamLeader(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return Team{}, s.whySetLeaderMissed(ctx, orgID, teamID)
	}
	if err != nil {
		return Team{}, fmt.Errorf("set team leader: %w", err)
	}

	return toTeam(t), nil
}

// the update checks the team and the role in one go, so when it misses, look
// at the team on its own to say which one was wrong
func (s *Service) whySetLeaderMissed(
	ctx context.Context, orgID, teamID uuid.UUID,
) error {
	t, err := s.q.GetTeam(ctx, teamID)
	if errors.Is(err, pgx.ErrNoRows) || t.OrgID != orgID || t.ArchivedAt.Valid {
		return ErrTeamNotFound
	}
	if err != nil {
		return fmt.Errorf("get team: %w", err)
	}

	org, err := s.q.GetOrg(ctx, orgID)
	if err != nil {
		return fmt.Errorf("get org: %w", err)
	}
	if org.ArchivedAt.Valid {
		return ErrTeamNotFound
	}

	return ErrRoleNotFound
}
