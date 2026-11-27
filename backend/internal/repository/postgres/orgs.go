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

type OrgRepo struct {
	q *db.Queries
}

// NewOrgRepo takes a pool or a pgx.Tx.
func NewOrgRepo(conn db.DBTX) *OrgRepo {
	return &OrgRepo{q: db.New(conn)}
}

func (r *OrgRepo) Get(ctx context.Context, id uuid.UUID) (repository.Org, error) {
	row, err := r.q.GetOrg(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.Org{}, repository.ErrOrgNotFound
	}
	if err != nil {
		return repository.Org{}, fmt.Errorf("get org %s: %w", id, err)
	}
	return repository.Org{
		ID:         row.ID,
		Name:       row.Name,
		Slug:       row.Slug,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		ArchivedAt: row.ArchivedAt,
	}, nil
}
