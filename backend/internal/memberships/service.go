package memberships

import (
	"backend/db"
	"backend/internal/pgerr"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRoleNotInSociety       = errors.New("role does not belong to society")
	ErrSocietyNotFound        = errors.New("society not found")
	ErrAlreadyMember          = errors.New("user already holds this role")
	ErrPreviousHolderNotFound = errors.New("previous user does not hold this role")
)

type Service struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, q: db.New(pool)}
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
	if pgerr.IsUniqueViolation(err) {
		return 0, ErrAlreadyMember
	}
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
		soc, err := s.q.GetSociety(ctx, societyID)
		if errors.Is(err, pgx.ErrNoRows) || soc.ArchivedAt.Valid {
			return AssignResult{}, ErrSocietyNotFound
		}
		if err != nil {
			return AssignResult{}, fmt.Errorf("get society: %w", err)
		}
	}

	return AssignResult{Created: len(rows), Skipped: len(pairs) - len(rows)}, nil
}

type Handover struct {
	SocietyID int32
	RoleID    int32
	NewUserID uuid.UUID
	// nil ends every current holder of the role
	PreviousUserID *uuid.UUID
	Term           string
}

type HandoverResult struct {
	MembershipID int32
	Ended        int
}

// HandoverRole ends the current holder(s) of a role and gives it to a new user,
// all in one transaction so the role is never left half handed over.
func (s *Service) HandoverRole(
	ctx context.Context, in Handover,
) (HandoverResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return HandoverResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)

	// lock the role first. a second handover of the same role waits here until
	// we commit, so its update then sees our new holder instead of missing it
	_, err = q.LockSocRoleInSociety(ctx, db.LockSocRoleInSocietyParams{
		RoleID: in.RoleID,
		SocID:  in.SocietyID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return HandoverResult{}, ErrRoleNotInSociety
	}
	if err != nil {
		return HandoverResult{}, fmt.Errorf("lock role: %w", err)
	}

	end := db.EndRoleMembershipsParams{RoleID: in.RoleID}
	if in.PreviousUserID != nil {
		end.UserID = pgtype.UUID{Bytes: *in.PreviousUserID, Valid: true}
	}
	ended, err := q.EndRoleMemberships(ctx, end)
	if err != nil {
		return HandoverResult{}, fmt.Errorf("end memberships: %w", err)
	}
	if in.PreviousUserID != nil && len(ended) == 0 {
		return HandoverResult{}, ErrPreviousHolderNotFound
	}

	m, err := q.CreateMembership(ctx, db.CreateMembershipParams{
		UserID: pgtype.UUID{Bytes: in.NewUserID, Valid: true},
		Term:   pgtype.Text{String: in.Term, Valid: true},
		RoleID: in.RoleID,
		SocID:  in.SocietyID,
	})
	// only possible when the new user is a current holder who wasn't ended
	if pgerr.IsUniqueViolation(err) {
		return HandoverResult{}, ErrAlreadyMember
	}
	if err != nil {
		return HandoverResult{}, fmt.Errorf("create membership: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return HandoverResult{}, fmt.Errorf("commit: %w", err)
	}

	return HandoverResult{MembershipID: m.ID, Ended: len(ended)}, nil
}
