package service

import (
	"context"

	"backend/internal/repository"

	"github.com/google/uuid"
)

// The service defines the interfaces it needs; the repositories satisfy them.

// Postgres (metadata).

type OrgRepository interface {
	Get(ctx context.Context, id uuid.UUID) (repository.Org, error)
}

type TeamRepository interface {
	Get(ctx context.Context, id uuid.UUID) (repository.Team, error)
	Create(ctx context.Context, id uuid.UUID, name string) (repository.Team, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type RoleRepository interface {
	Get(ctx context.Context, id uuid.UUID) (repository.Role, error)
	Create(ctx context.Context, id uuid.UUID, name string) (repository.Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// SpiceDB (relationships).

type TeamRelations interface {
	LinkOrg(ctx context.Context, teamID, orgID uuid.UUID) error
}

type RoleRelations interface {
	LinkTeam(ctx context.Context, roleID, teamID uuid.UUID) error
}

type MembershipRepository interface {
	AddHolder(ctx context.Context, roleID, userID uuid.UUID) error
}
