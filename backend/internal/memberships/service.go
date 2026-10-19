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
	ErrRootTeamRole           = errors.New("role is in the root team")
	ErrMembershipEnded        = errors.New("membership has already ended")
	ErrSameRole               = errors.New("membership already has this role")
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
	// rejects roles in the org's root (exec) team. set for grpc clients, which
	// can only staff the teams below exec
	NonRootOnly bool
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
		UserID:      in.UserID,
		RoleID:      in.RoleID,
		OrgID:       in.OrgID,
		NonRootOnly: in.NonRootOnly,
	})
	// the insert selects the role, so no row back means the role isn't in that
	// org, the role, its team or the org is archived, or (with NonRootOnly) the
	// role is in the root team
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, s.whyCreateMissed(ctx, in)
	}
	if pgerr.IsUniqueViolation(err) {
		return uuid.Nil, ErrAlreadyMember
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("create membership: %w", err)
	}

	return m.ID, nil
}

// only the root team rule needs telling apart, everything else is not found
func (s *Service) whyCreateMissed(ctx context.Context, in NewMembership) error {
	if !in.NonRootOnly {
		return ErrRoleNotInOrg
	}

	isRoot, err := s.q.IsRootTeamRole(ctx, db.IsRootTeamRoleParams{
		RoleID: in.RoleID,
		OrgID:  in.OrgID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRoleNotInOrg
	}
	if err != nil {
		return fmt.Errorf("check root team role: %w", err)
	}
	if isRoot {
		return ErrRootTeamRole
	}

	return ErrRoleNotInOrg
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

type ArchiveInput struct {
	OrgID        uuid.UUID
	MembershipID uuid.UUID
	// rejects memberships in the org's root (exec) team, for grpc clients
	NonRootOnly bool
}

// ArchiveMembership ends a membership. Archiving it again is a no-op.
func (s *Service) ArchiveMembership(
	ctx context.Context, in ArchiveInput,
) (Membership, error) {
	m, err := s.q.ArchiveMembership(ctx, db.ArchiveMembershipParams{
		MembershipID: in.MembershipID,
		OrgID:        in.OrgID,
		NonRootOnly:  in.NonRootOnly,
	})
	// no row back covers a missing membership, one from another org, an
	// archived org, and (with NonRootOnly) one in the root team
	if errors.Is(err, pgx.ErrNoRows) {
		return Membership{}, s.whyArchiveMissed(ctx, in)
	}
	if err != nil {
		return Membership{}, fmt.Errorf("archive membership: %w", err)
	}

	return toMembership(m), nil
}

// only the root team rule needs telling apart, everything else is not found
func (s *Service) whyArchiveMissed(ctx context.Context, in ArchiveInput) error {
	if !in.NonRootOnly {
		return ErrMembershipNotFound
	}

	m, err := s.q.LockMembershipInOrg(ctx, db.LockMembershipInOrgParams{
		MembershipID: in.MembershipID,
		OrgID:        in.OrgID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrMembershipNotFound
	}
	if err != nil {
		return fmt.Errorf("get membership: %w", err)
	}
	if m.InRootTeam {
		return ErrRootTeamRole
	}

	return ErrMembershipNotFound
}

type RoleChange struct {
	OrgID        uuid.UUID
	MembershipID uuid.UUID
	NewRoleID    uuid.UUID
	// nil keeps the old membership's term
	Term *string
	// rejects moving out of or into the org's root (exec) team, for grpc
	// clients
	NonRootOnly bool
}

type RoleChangeResult struct {
	EndedMembershipID uuid.UUID
	MembershipID      uuid.UUID
}

// ChangeMembershipRole moves someone to a different role, e.g. Subcom to
// Director. The old membership is ended and a new one started in the same
// transaction, so history keeps who held which role when.
func (s *Service) ChangeMembershipRole(
	ctx context.Context, in RoleChange,
) (RoleChangeResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return RoleChangeResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)

	old, err := q.LockMembershipInOrg(ctx, db.LockMembershipInOrgParams{
		MembershipID: in.MembershipID,
		OrgID:        in.OrgID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return RoleChangeResult{}, ErrMembershipNotFound
	}
	if err != nil {
		return RoleChangeResult{}, fmt.Errorf("lock membership: %w", err)
	}
	switch {
	case old.EndedAt.Valid:
		return RoleChangeResult{}, ErrMembershipEnded
	case in.NonRootOnly && old.InRootTeam:
		return RoleChangeResult{}, ErrRootTeamRole
	case old.RoleID == in.NewRoleID:
		return RoleChangeResult{}, ErrSameRole
	}

	// a user holds a role at most once at a time, so this ends exactly the
	// locked membership
	_, err = q.EndRoleMemberships(ctx, db.EndRoleMembershipsParams{
		RoleID: old.RoleID,
		UserID: uuid.NullUUID{UUID: old.UserID, Valid: true},
	})
	if err != nil {
		return RoleChangeResult{}, fmt.Errorf("end membership: %w", err)
	}

	term := old.Term
	if in.Term != nil {
		term = pgtype.Text{String: *in.Term, Valid: true}
	}
	newIn := NewMembership{
		OrgID:       in.OrgID,
		RoleID:      in.NewRoleID,
		UserID:      old.UserID,
		NonRootOnly: in.NonRootOnly,
	}
	m, err := q.CreateMembership(ctx, db.CreateMembershipParams{
		UserID:      newIn.UserID,
		Term:        term,
		RoleID:      newIn.RoleID,
		OrgID:       newIn.OrgID,
		NonRootOnly: newIn.NonRootOnly,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return RoleChangeResult{}, s.whyCreateMissed(ctx, newIn)
	}
	// the user already holds the new role in another membership
	if pgerr.IsUniqueViolation(err) {
		return RoleChangeResult{}, ErrAlreadyMember
	}
	if err != nil {
		return RoleChangeResult{}, fmt.Errorf("create membership: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return RoleChangeResult{}, fmt.Errorf("commit: %w", err)
	}

	return RoleChangeResult{EndedMembershipID: old.ID, MembershipID: m.ID}, nil
}
