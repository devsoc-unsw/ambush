package audit

import (
	"backend/db"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("audit entry not found")

type Service struct {
	q *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{q: db.New(pool)}
}

func (s *Service) GetAuditEntry(ctx context.Context, id int64) (db.AuditLog, error) {
	entry, err := s.q.GetAuditEntry(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.AuditLog{}, ErrNotFound
	}
	if err != nil {
		return db.AuditLog{}, fmt.Errorf("get audit entry: %w", err)
	}

	return entry, nil
}
