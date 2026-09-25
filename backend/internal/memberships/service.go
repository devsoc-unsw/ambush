package memberships

import (
	"backend/db"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRoleNotInSociety = errors.New("role does not belong to society")

type Service struct {
	q *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{q: db.New(pool)}
}

type NewMembership struct {
	SocietyID int32
	RoleID    int32
	UserID    uuid.UUID
}

func (s *Service) CreateMembership(
	ctx context.Context, in NewMembership,
) (int32, error) {
	m, err := s.q.CreateMembership(ctx, db.CreateMembershipParams{
		UserID: pgtype.UUID{Bytes: in.UserID, Valid: true},
		RoleID: in.RoleID,
		SocID:  in.SocietyID,
	})
	// the insert selects the role from socroles, so no row back means the
	// role doesn't exist in that society
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrRoleNotInSociety
	}
	// TODO: err for a user who already holds this role
	// NOTE: should return a pg errcode 23505
	if err != nil {
		return 0, fmt.Errorf("create membership: %w", err)
	}

	return m.ID, nil
}

func (s *Service) ListUserPortfolios(
	ctx context.Context, userID uuid.UUID,
) ([]db.ListUserPortfoliosRow, error) {
	rows, err := s.q.ListUserPortfolios(ctx,
		pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list user portfolios: %w", err)
	}

	return rows, nil
}
