package spicedb

import (
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/google/uuid"
)

// Object types and relations from ambush.zed.
const (
	typeUser = "user"
	typeOrg  = "ambush/org"
	typeTeam = "ambush/team"
	typeRole = "ambush/role"

	relOrg    = "org"
	relTeam   = "team"
	relRole   = "role"
	relHolder = "holder"
)

func object(objectType string, id uuid.UUID) *v1.ObjectReference {
	return &v1.ObjectReference{ObjectType: objectType, ObjectId: id.String()}
}

// create is a CREATE update for resource#relation@subject.
func create(
	resource *v1.ObjectReference, relation string, subject *v1.ObjectReference,
) *v1.RelationshipUpdate {
	return &v1.RelationshipUpdate{
		Operation: v1.RelationshipUpdate_OPERATION_CREATE,
		Relationship: &v1.Relationship{
			Resource: resource,
			Relation: relation,
			Subject:  &v1.SubjectReference{Object: subject},
		},
	}
}
