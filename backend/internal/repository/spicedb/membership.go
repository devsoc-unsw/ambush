package spicedb

import (
	"context"
	"fmt"

	"backend/internal/repository"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	typeUser = "user"
	typeRole = "ambush/role"

	relHolder = "holder"
)

type MembershipRepo struct {
	c *authzed.Client
}

func NewMembershipRepo(c *authzed.Client) *MembershipRepo {
	return &MembershipRepo{c: c}
}

func (r *MembershipRepo) AddHolder(
	ctx context.Context, roleID, userID uuid.UUID,
) error {
	rel := &v1.Relationship{
		Resource: object(typeRole, roleID),
		Relation: relHolder,
		Subject:  &v1.SubjectReference{Object: object(typeUser, userID)},
	}
	_, err := r.c.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{{
			Operation:    v1.RelationshipUpdate_OPERATION_CREATE,
			Relationship: rel,
		}},
	})
	if status.Code(err) == codes.AlreadyExists {
		return repository.ErrAlreadyMember
	}
	if err != nil {
		return fmt.Errorf("add holder: %w", err)
	}
	return nil
}

func object(objectType string, id uuid.UUID) *v1.ObjectReference {
	return &v1.ObjectReference{ObjectType: objectType, ObjectId: id.String()}
}
