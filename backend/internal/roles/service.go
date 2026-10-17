package roles

import (
	"backend/db"
	"backend/internal/pgerr"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTeamNotFound = errors.New("team not found")
	ErrNameTaken    = errors.New("role name already taken in this team")
)

type Service struct {
	q *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{q: db.New(pool)}
}

type NewRole struct {
	OrgID  uuid.UUID
	TeamID uuid.UUID
	Name   string
}

func (s *Service) CreateRole(ctx context.Context, in NewRole) (uuid.UUID, error) {
	role, err := s.q.CreateRole(ctx, db.CreateRoleParams{
		OrgID:  in.OrgID,
		TeamID: in.TeamID,
		Name:   in.Name,
	})
	// the insert selects the team, so no row back means the team isn't in that
	// org, or the team or org is archived
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrTeamNotFound
	}
	if pgerr.IsUniqueViolation(err) {
		return uuid.Nil, ErrNameTaken
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("create role: %w", err)
	}

	return role.ID, nil
}
