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
	RootTeamName string
}

type CreatedOrg struct {
	OrgID      uuid.UUID
	RootTeamID uuid.UUID
}

// CreateOrg creates an org and its root team together, since an org without a
// root team breaks the team tree.
func (s *Service) CreateOrg(ctx context.Context, in NewOrg) (CreatedOrg, error) {
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
