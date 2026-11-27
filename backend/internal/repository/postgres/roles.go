// Package postgres reads and writes metadata in postgres. It wraps the
// sqlc-generated code in ./db so the service layer never sees sqlc or pgx.
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

type RoleRepo struct {
	q *db.Queries
}

// NewRoleRepo takes a pool or a pgx.Tx.
func NewRoleRepo(conn db.DBTX) *RoleRepo {
	return &RoleRepo{q: db.New(conn)}
}

func (r *RoleRepo) Get(ctx context.Context, id uuid.UUID) (repository.Role, error) {
	row, err := r.q.GetRole(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.Role{}, repository.ErrRoleNotFound
	}
	if err != nil {
		return repository.Role{}, fmt.Errorf("get role %s: %w", id, err)
	}
	return toRole(row), nil
}

func (r *RoleRepo) Create(
	ctx context.Context, id uuid.UUID, name string,
) (repository.Role, error) {
	row, err := r.q.CreateRole(ctx, db.CreateRoleParams{ID: id, Name: name})
	if err != nil {
		return repository.Role{}, fmt.Errorf("create role: %w", err)
	}
	return toRole(row), nil
}

func (r *RoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteRole(ctx, id); err != nil {
		return fmt.Errorf("delete role %s: %w", id, err)
	}
	return nil
}

func toRole(row db.Role) repository.Role {
	return repository.Role{
		ID:         row.ID,
		Name:       row.Name,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		ArchivedAt: row.ArchivedAt,
	}
}
