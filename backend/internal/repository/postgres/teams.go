package postgres

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/repository"
	"backend/internal/repository/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TeamRepo struct {
	q *db.Queries
}

// NewTeamRepo takes a pool or a pgx.Tx.
func NewTeamRepo(conn db.DBTX) *TeamRepo {
	return &TeamRepo{q: db.New(conn)}
}

func (r *TeamRepo) Get(ctx context.Context, id uuid.UUID) (repository.Team, error) {
	row, err := r.q.GetTeam(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.Team{}, repository.ErrTeamNotFound
	}
	if err != nil {
		return repository.Team{}, fmt.Errorf("get team %s: %w", id, err)
	}
	return toTeam(row), nil
}

func (r *TeamRepo) Create(
	ctx context.Context, id uuid.UUID, name string,
) (repository.Team, error) {
	row, err := r.q.CreateTeam(ctx, db.CreateTeamParams{ID: id, Name: name})
	if err != nil {
		return repository.Team{}, fmt.Errorf("create team: %w", err)
	}
	return toTeam(row), nil
}

func (r *TeamRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteTeam(ctx, id); err != nil {
		return fmt.Errorf("delete team %s: %w", id, err)
	}
	return nil
}

func toTeam(row db.Team) repository.Team {
	return repository.Team{
		ID:         row.ID,
		Name:       row.Name,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		ArchivedAt: row.ArchivedAt,
	}
}
