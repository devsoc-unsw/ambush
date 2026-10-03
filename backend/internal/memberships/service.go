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

var (
	ErrRoleNotInSociety = errors.New("role does not belong to society")
	ErrSocietyNotFound  = errors.New("society not found")
)

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

type MembershipPair struct {
	RoleID int32
	UserID uuid.UUID
}

type AssignResult struct {
	Created int
	Skipped int
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

func (s *Service) AssignMemberships(
	ctx context.Context, societyID int32, pairs []MembershipPair,
) (AssignResult, error) {
	// same index in both slices so unnest pairs them back up
	roleIDs := make([]int32, len(pairs))
	userIDs := make([]pgtype.UUID, len(pairs))
	for i, p := range pairs {
		roleIDs[i] = p.RoleID
		userIDs[i] = pgtype.UUID{Bytes: p.UserID, Valid: true}
	}

	// pairs whose role isn't in the society, or that already exist, are
	// silently dropped by the query, so they only show up as skipped
	rows, err := s.q.CreateMemberships(ctx, db.CreateMembershipsParams{
		RoleIds: roleIDs,
		UserIds: userIDs,
		SocID:   societyID,
	})
	if err != nil {
		return AssignResult{}, fmt.Errorf("assign memberships: %w", err)
	}

	// nothing created could mean the society doesn't exist. only check then,
	// since any created row already proves it does
	if len(rows) == 0 {
		_, err = s.q.GetSociety(ctx, societyID)
		if errors.Is(err, pgx.ErrNoRows) {
			return AssignResult{}, ErrSocietyNotFound
		}
		if err != nil {
			return AssignResult{}, fmt.Errorf("get society: %w", err)
		}
	}

	return AssignResult{Created: len(rows), Skipped: len(pairs) - len(rows)}, nil
}
