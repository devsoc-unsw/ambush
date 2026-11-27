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

type MembershipRepo struct {
	c *authzed.Client
}

func NewMembershipRepo(c *authzed.Client) *MembershipRepo {
	return &MembershipRepo{c: c}
}

func (r *MembershipRepo) AddHolder(
	ctx context.Context, roleID, userID uuid.UUID,
) error {
	_, err := r.c.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			create(object(typeRole, roleID), relHolder, object(typeUser, userID)),
		},
	})
	if status.Code(err) == codes.AlreadyExists {
		return repository.ErrAlreadyMember
	}
	if err != nil {
		return fmt.Errorf("add holder: %w", err)
	}
	return nil
}
