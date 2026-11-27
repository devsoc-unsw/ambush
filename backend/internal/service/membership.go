package service

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/repository/postgres/db"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrRoleNotFound  = errors.New("role not found")
	ErrRoleArchived  = errors.New("role is archived")
	ErrAlreadyMember = errors.New("user already holds this role")
)

type MembershipService struct {
	q     *db.Queries
	spice *authzed.Client
}

func NewMembershipService(q *db.Queries, spice *authzed.Client) *MembershipService {
	return &MembershipService{q: q, spice: spice}
}

func (s *MembershipService) AddMembership(ctx context.Context, userID, roleID uuid.UUID) error {
	role, err := s.q.GetRole(ctx, roleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrRoleNotFound
	}
	if err != nil {
		return fmt.Errorf("look up role: %w", err)
	}
	if role.ArchivedAt != nil {
		return ErrRoleArchived
	}

	_, err = s.spice.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{{
			Operation: v1.RelationshipUpdate_OPERATION_CREATE,
			Relationship: &v1.Relationship{
				Resource: &v1.ObjectReference{ObjectType: "ambush/role", ObjectId: roleID.String()},
				Relation: "holder",
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{ObjectType: "user", ObjectId: userID.String()},
				},
			},
		}},
	})
	if status.Code(err) == codes.AlreadyExists {
		return ErrAlreadyMember
	}
	if err != nil {
		return fmt.Errorf("write holder: %w", err)
	}
	return nil
}
