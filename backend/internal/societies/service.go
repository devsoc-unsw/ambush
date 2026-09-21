package societies

import (
	"backend/db"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
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

	soc, err := q.CreateSociety(ctx, db.CreateSocietyParams{Name: in.Name, Slug: in.Slug})
	// TODO: err for inserting using a taken slug
	// NOTE: should return a pg errcode 23505
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
