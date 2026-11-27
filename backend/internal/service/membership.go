package service

import (
	"context"

	"github.com/google/uuid"
)

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
