package societies

import (
	"backend/db"
	"backend/internal/pgerr"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSocietyNotFound = errors.New("society not found")
	ErrSlugTaken       = errors.New("slug already taken")
)

type Service struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, q: db.New(pool)}
}

type NewSociety struct {
	Name string
	Slug string
}

func (s *Service) CreateSociety(ctx context.Context, in NewSociety) (int32, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)

	soc, err := q.CreateSociety(ctx,
		db.CreateSocietyParams{Name: in.Name, Slug: in.Slug})
	if pgerr.IsUniqueViolation(err) {
		return 0, ErrSlugTaken
	}
	if err != nil {
		return 0, fmt.Errorf("create society: %w", err)
	}

	// TODO: include socrole and a membership for this

	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return soc.ID, nil
}

type Society struct {
	ID         int32
	Name       string
	Slug       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time // nil = active
}

func toSociety(soc db.Society) Society {
	out := Society{
		ID:        soc.ID,
		Name:      soc.Name,
		Slug:      soc.Slug,
		CreatedAt: soc.CreatedAt.Time,
		UpdatedAt: soc.UpdatedAt.Time,
	}
	if soc.ArchivedAt.Valid {
		out.ArchivedAt = &soc.ArchivedAt.Time
	}
	return out
}

// nil fields are left unchanged
type SocietyUpdate struct {
	Name *string
	Slug *string
}

func (s *Service) UpdateSociety(
	ctx context.Context, id int32, in SocietyUpdate,
) (Society, error) {
	params := db.UpdateSocietyParams{ID: id}
	if in.Name != nil {
		params.Name = pgtype.Text{String: *in.Name, Valid: true}
	}
	if in.Slug != nil {
		params.Slug = pgtype.Text{String: *in.Slug, Valid: true}
	}

	soc, err := s.q.UpdateSociety(ctx, params)
	// archived societies are read only, so they're not found here too
	if errors.Is(err, pgx.ErrNoRows) {
		return Society{}, ErrSocietyNotFound
	}
	if pgerr.IsUniqueViolation(err) {
		return Society{}, ErrSlugTaken
	}
	if err != nil {
		return Society{}, fmt.Errorf("update society: %w", err)
	}

	return toSociety(soc), nil
}

// ArchiveSociety makes a society read only. Archiving it again is a no-op.
func (s *Service) ArchiveSociety(ctx context.Context, id int32) (Society, error) {
	soc, err := s.q.ArchiveSociety(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Society{}, ErrSocietyNotFound
	}
	if err != nil {
		return Society{}, fmt.Errorf("archive society: %w", err)
	}

	return toSociety(soc), nil
}
