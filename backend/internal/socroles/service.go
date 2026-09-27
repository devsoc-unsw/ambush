package socroles

import (
	"backend/db"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	q *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{q: db.New(pool)}
}

type NewSocRole struct {
	SocietyID    int32
	ParentRoleID *int32
	Name         string
}

func (s *Service) CreateSocRole(
	ctx context.Context, in NewSocRole,
) (int32, error) {
	params := db.CreateSocRoleParams{SocID: in.SocietyID, Name: in.Name}
	if in.ParentRoleID != nil {
		params.ParentRoleID = pgtype.Int4{Int32: *in.ParentRoleID, Valid: true}
	}

	socrole, err := s.q.CreateSocRole(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("create socrole: %w", err)
	}

	return socrole.ID, nil
}
