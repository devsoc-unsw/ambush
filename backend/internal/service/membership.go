package service

import (
	"context"
	"errors"

	"backend/internal/repository"

	"github.com/google/uuid"
)

var ErrRoleArchived = errors.New("role is archived")

// The service defines the interfaces it needs; the repositories satisfy them.
type RoleRepository interface {
	Get(ctx context.Context, id uuid.UUID) (repository.Role, error)
}

type MembershipRepository interface {
	AddHolder(ctx context.Context, roleID, userID uuid.UUID) error
}

type MembershipService struct {
	roles       RoleRepository
	memberships MembershipRepository
}

func NewMembershipService(
	roles RoleRepository, memberships MembershipRepository,
) *MembershipService {
	return &MembershipService{roles: roles, memberships: memberships}
}

func (s *MembershipService) AddMembership(
	ctx context.Context, userID, roleID uuid.UUID,
) error {
	role, err := s.roles.Get(ctx, roleID)
	if err != nil {
		return err
	}
	if role.Archived() {
		return ErrRoleArchived
	}
	return s.memberships.AddHolder(ctx, roleID, userID)
}
