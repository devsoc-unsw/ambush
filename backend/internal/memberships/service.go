package memberships

import (
	"backend/db"
	"backend/internal/pgerr"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRoleNotInOrg           = errors.New("role does not belong to org")
	ErrOrgNotFound            = errors.New("org not found")
	ErrAlreadyMember          = errors.New("user already holds this role")
	ErrPreviousHolderNotFound = errors.New("previous user does not hold this role")
	ErrMembershipNotFound     = errors.New("membership not found")
)

type Service struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, q: db.New(pool)}
}

type NewMembership struct {
	OrgID  uuid.UUID
	RoleID uuid.UUID
	UserID uuid.UUID
}

type MembershipPair struct {
	RoleID uuid.UUID
	UserID uuid.UUID
}

type AssignResult struct {
	Created int
	Skipped int
}

type Membership struct {
	ID       uuid.UUID
	RoleID   uuid.UUID
	UserID   uuid.UUID
	Term     *string
	JoinedAt time.Time
	EndedAt  *time.Time // nil = still current
}

func toMembership(m db.Membership) Membership {
	out := Membership{
		ID:       m.ID,
		RoleID:   m.RoleID,
		UserID:   m.UserID,
		JoinedAt: m.CreatedAt.Time,
	}
	if m.Term.Valid {
		out.Term = &m.Term.String
	}
	if m.EndedAt.Valid {
		out.EndedAt = &m.EndedAt.Time
	}
	return out
}

// Portfolio is one role a user currently holds, with its team and org.
type Portfolio struct {
	MembershipID uuid.UUID
	JoinedAt     time.Time
	RoleID       uuid.UUID
	RoleName     string
	TeamID       uuid.UUID
	TeamName     string
	OrgID        uuid.UUID
	OrgName      string
	OrgSlug      string
}

func (s *Service) CreateMembership(
	ctx context.Context, in NewMembership,
) (uuid.UUID, error) {
	m, err := s.q.CreateMembership(ctx, db.CreateMembershipParams{
		UserID: in.UserID,
		RoleID: in.RoleID,
		OrgID:  in.OrgID,
	})
	// the insert selects the role, so no row back means the role isn't in that
	// org, or the role, its team or the org is archived
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrRoleNotInOrg
	}
	if pgerr.IsUniqueViolation(err) {
		return uuid.Nil, ErrAlreadyMember
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("create membership: %w", err)
	}

	return m.ID, nil
}

func (s *Service) ListUserPortfolios(
	ctx context.Context, userID uuid.UUID,
) ([]Portfolio, error) {
	rows, err := s.q.ListUserPortfolios(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user portfolios: %w", err)
	}

	out := make([]Portfolio, len(rows))
	for i, r := range rows {
		out[i] = Portfolio{
			MembershipID: r.MembershipID,
			JoinedAt:     r.JoinedAt.Time,
			RoleID:       r.RoleID,
			RoleName:     r.RoleName,
			TeamID:       r.TeamID,
			TeamName:     r.TeamName,
			OrgID:        r.OrgID,
			OrgName:      r.OrgName,
			OrgSlug:      r.OrgSlug,
		}
	}

	return out, nil
}

func (s *Service) AssignMemberships(
	ctx context.Context, orgID uuid.UUID, pairs []MembershipPair,
) (AssignResult, error) {
	// same index in both slices so unnest pairs them back up
	roleIDs := make([]uuid.UUID, len(pairs))
	userIDs := make([]uuid.UUID, len(pairs))
	for i, p := range pairs {
		roleIDs[i] = p.RoleID
		userIDs[i] = p.UserID
	}

	// pairs whose role isn't usable in the org, or that already exist, are
	// silently dropped by the query, so they only show up as skipped
	rows, err := s.q.CreateMemberships(ctx, db.CreateMembershipsParams{
		RoleIds: roleIDs,
		UserIds: userIDs,
		OrgID:   orgID,
	})
	if err != nil {
		return AssignResult{}, fmt.Errorf("assign memberships: %w", err)
	}

	// nothing created could mean the org doesn't exist. only check then,
	// since any created row already proves it does
	if len(rows) == 0 {
		org, err := s.q.GetOrg(ctx, orgID)
		if errors.Is(err, pgx.ErrNoRows) || org.ArchivedAt.Valid {
			return AssignResult{}, ErrOrgNotFound
		}
		if err != nil {
			return AssignResult{}, fmt.Errorf("get org: %w", err)
		}
	}

	return AssignResult{Created: len(rows), Skipped: len(pairs) - len(rows)}, nil
}

type Handover struct {
	OrgID     uuid.UUID
	RoleID    uuid.UUID
	NewUserID uuid.UUID
	// nil ends every current holder of the role
	PreviousUserID *uuid.UUID
	Term           string
}

type HandoverResult struct {
	MembershipID uuid.UUID
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
	_, err = q.LockActiveRoleInOrg(ctx, db.LockActiveRoleInOrgParams{
		RoleID: in.RoleID,
		OrgID:  in.OrgID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return HandoverResult{}, ErrRoleNotInOrg
	}
	if err != nil {
		return HandoverResult{}, fmt.Errorf("lock role: %w", err)
	}

	end := db.EndRoleMembershipsParams{RoleID: in.RoleID}
	if in.PreviousUserID != nil {
		end.UserID = uuid.NullUUID{UUID: *in.PreviousUserID, Valid: true}
	}
	ended, err := q.EndRoleMemberships(ctx, end)
	if err != nil {
		return HandoverResult{}, fmt.Errorf("end memberships: %w", err)
	}
	if in.PreviousUserID != nil && len(ended) == 0 {
		return HandoverResult{}, ErrPreviousHolderNotFound
	}

	m, err := q.CreateMembership(ctx, db.CreateMembershipParams{
		UserID: in.NewUserID,
		Term:   pgtype.Text{String: in.Term, Valid: true},
		RoleID: in.RoleID,
		OrgID:  in.OrgID,
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

// ArchiveMembership ends a membership. Archiving it again is a no-op.
func (s *Service) ArchiveMembership(
	ctx context.Context, orgID, membershipID uuid.UUID,
) (Membership, error) {
	m, err := s.q.ArchiveMembership(ctx, db.ArchiveMembershipParams{
		MembershipID: membershipID,
		OrgID:        orgID,
	})
	// no row back covers a missing membership, one from another org, and an
	// archived org
	if errors.Is(err, pgx.ErrNoRows) {
		return Membership{}, ErrMembershipNotFound
	}
	if err != nil {
		return Membership{}, fmt.Errorf("archive membership: %w", err)
	}

	return toMembership(m), nil
}
