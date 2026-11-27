package spicedb

import (
	"context"
	"fmt"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/google/uuid"
)

type RoleRepo struct {
	c *authzed.Client
}

func NewRoleRepo(c *authzed.Client) *RoleRepo {
	return &RoleRepo{c: c}
}

// LinkTeam writes both directions in one call, so either both exist or
// neither does:
//
//	ambush/role:<roleID>#team@ambush/team:<teamID>  (archive checks, manage_holder)
//	ambush/team:<teamID>#role@ambush/role:<roleID>  (team#member)
func (r *RoleRepo) LinkTeam(ctx context.Context, roleID, teamID uuid.UUID) error {
	role, team := object(typeRole, roleID), object(typeTeam, teamID)
	_, err := r.c.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			create(role, relTeam, team),
			create(team, relRole, role),
		},
	})
	if err != nil {
		return fmt.Errorf("link role to team: %w", err)
	}
	return nil
}
