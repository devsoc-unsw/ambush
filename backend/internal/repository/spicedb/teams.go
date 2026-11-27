package spicedb

import (
	"context"
	"fmt"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/google/uuid"
)

type TeamRepo struct {
	c *authzed.Client
}

func NewTeamRepo(c *authzed.Client) *TeamRepo {
	return &TeamRepo{c: c}
}

// LinkOrg writes ambush/team:<teamID>#org@ambush/org:<orgID>.
func (r *TeamRepo) LinkOrg(ctx context.Context, teamID, orgID uuid.UUID) error {
	_, err := r.c.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			create(object(typeTeam, teamID), relOrg, object(typeOrg, orgID)),
		},
	})
	if err != nil {
		return fmt.Errorf("link team to org: %w", err)
	}
	return nil
}
