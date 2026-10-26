package orgs

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

// used when a new org doesn't name its root team
const defaultRootTeamName = "Exec"

var (
	ErrOrgNotFound = errors.New("org not found")
	ErrSlugTaken   = errors.New("slug already taken")
)

type Service struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, q: db.New(pool)}
}

type NewOrg struct {
	Name         string
	Slug         string
	RootTeamName string // optional, use "" if none
}

type CreatedOrg struct {
	OrgID      uuid.UUID
	RootTeamID uuid.UUID
}

// CreateOrg creates an org and its root team together, since an org without a
// root team breaks the team tree.
func (s *Service) CreateOrg(ctx context.Context, in NewOrg) (CreatedOrg, error) {
	if in.RootTeamName == "" {
		in.RootTeamName = defaultRootTeamName
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CreatedOrg{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)

	org, err := q.CreateOrg(ctx, db.CreateOrgParams{Name: in.Name, Slug: in.Slug})
	if pgerr.IsUniqueViolation(err) {
		return CreatedOrg{}, ErrSlugTaken
	}
	if err != nil {
		return CreatedOrg{}, fmt.Errorf("create org: %w", err)
	}

	root, err := q.CreateRootTeam(ctx, db.CreateRootTeamParams{
		OrgID: org.ID,
		Name:  in.RootTeamName,
	})
	if err != nil {
		return CreatedOrg{}, fmt.Errorf("create root team: %w", err)
	}

	// TODO: a leader role for the root team and a membership for the creator,
	// once the creator's user id comes from auth

	err = tx.Commit(ctx)
	if err != nil {
		return CreatedOrg{}, fmt.Errorf("commit: %w", err)
	}

	return CreatedOrg{OrgID: org.ID, RootTeamID: root.ID}, nil
}

type Org struct {
	ID         uuid.UUID
	Name       string
	Slug       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time // nil = active
}

func toOrg(org db.Org) Org {
	out := Org{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		CreatedAt: org.CreatedAt.Time,
		UpdatedAt: org.UpdatedAt.Time,
	}
	if org.ArchivedAt.Valid {
		out.ArchivedAt = &org.ArchivedAt.Time
	}
	return out
}

// nil fields are left unchanged
type OrgUpdate struct {
	Name *string
	Slug *string
}

func (s *Service) UpdateOrg(
	ctx context.Context, id uuid.UUID, in OrgUpdate,
) (Org, error) {
	params := db.UpdateOrgParams{ID: id}
	if in.Name != nil {
		params.Name = pgtype.Text{String: *in.Name, Valid: true}
	}
	if in.Slug != nil {
		params.Slug = pgtype.Text{String: *in.Slug, Valid: true}
	}

	org, err := s.q.UpdateOrg(ctx, params)
	// archived orgs are read only, so they're not found here too
	if errors.Is(err, pgx.ErrNoRows) {
		return Org{}, ErrOrgNotFound
	}
	if pgerr.IsUniqueViolation(err) {
		return Org{}, ErrSlugTaken
	}
	if err != nil {
		return Org{}, fmt.Errorf("update org: %w", err)
	}

	return toOrg(org), nil
}

// ArchiveOrg makes an org read only. Archiving it again is a no-op.
func (s *Service) ArchiveOrg(ctx context.Context, id uuid.UUID) (Org, error) {
	org, err := s.q.ArchiveOrg(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Org{}, ErrOrgNotFound
	}
	if err != nil {
		return Org{}, fmt.Errorf("archive org: %w", err)
	}

	return toOrg(org), nil
}

// the structure snapshot is what client apps (chaos, warchest) load into
// their own spicedb. it only has active teams and roles and current
// memberships

type OrgStructure struct {
	Org         Org
	Teams       []TeamStructure
	Roles       []RoleStructure
	Memberships []MembershipStructure
}

type TeamStructure struct {
	ID           uuid.UUID
	Name         string
	ParentID     *uuid.UUID // nil = the org's root (exec) team
	LeaderRoleID *uuid.UUID // nil = no leader set
}

type RoleStructure struct {
	ID     uuid.UUID
	TeamID uuid.UUID
	Name   string
}

type MembershipStructure struct {
	ID     uuid.UUID
	RoleID uuid.UUID
	UserID uuid.UUID
}

func (s *Service) GetOrgStructure(
	ctx context.Context, id uuid.UUID,
) (OrgStructure, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return OrgStructure{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)

	org, err := q.GetOrg(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrgStructure{}, ErrOrgNotFound
	}
	if err != nil {
		return OrgStructure{}, fmt.Errorf("get org: %w", err)
	}

	teams, err := q.ListTeamsByOrg(ctx, id)
	if err != nil {
		return OrgStructure{}, fmt.Errorf("list teams: %w", err)
	}
	roles, err := q.ListRolesByOrg(ctx, id)
	if err != nil {
		return OrgStructure{}, fmt.Errorf("list roles: %w", err)
	}
	memberships, err := q.ListCurrentMembershipsByOrg(ctx, id)
	if err != nil {
		return OrgStructure{}, fmt.Errorf("list memberships: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return OrgStructure{}, fmt.Errorf("commit: %w", err)
	}

	return OrgStructure{
		Org:         toOrg(org),
		Teams:       toTeamStructures(teams),
		Roles:       toRoleStructures(roles),
		Memberships: toMembershipStructures(memberships),
	}, nil
}

func toTeamStructures(rows []db.Team) []TeamStructure {
	out := make([]TeamStructure, len(rows))
	for i, t := range rows {
		out[i] = TeamStructure{ID: t.ID, Name: t.Name}
		if t.ParentID.Valid {
			out[i].ParentID = &t.ParentID.UUID
		}
		if t.LeaderRoleID.Valid {
			out[i].LeaderRoleID = &t.LeaderRoleID.UUID
		}
	}
	return out
}

func toRoleStructures(rows []db.Role) []RoleStructure {
	out := make([]RoleStructure, len(rows))
	for i, r := range rows {
		out[i] = RoleStructure{ID: r.ID, TeamID: r.TeamID, Name: r.Name}
	}
	return out
}

func toMembershipStructures(rows []db.Membership) []MembershipStructure {
	out := make([]MembershipStructure, len(rows))
	for i, m := range rows {
		out[i] = MembershipStructure{ID: m.ID, RoleID: m.RoleID, UserID: m.UserID}
	}
	return out
}
