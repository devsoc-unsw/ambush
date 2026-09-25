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
	ParentRoleID *int32 // nil for a top-level role
	Name         string
}

func (s *Service) CreateSocRole(ctx context.Context, in NewSocRole) (int32, error) {
	params := db.CreateSocRoleParams{SocID: in.SocietyID, Name: in.Name}
	if in.ParentRoleID != nil {
		params.ParentRoleID = pgtype.Int4{Int32: *in.ParentRoleID, Valid: true}
	}

	role, err := s.q.CreateSocRole(ctx, params)
	// TODO: err for a taken role name in the society
	// NOTE: should return a pg errcode 23505
	// TODO: err for a missing society or parent role in another society
	// NOTE: should return a pg errcode 23503
	if err != nil {
		return 0, fmt.Errorf("create socrole: %w", err)
	}

	return role.ID, nil
}
