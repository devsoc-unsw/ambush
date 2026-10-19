package memberships

import (
	membershipsv1 "backend/gen/ambush/memberships/v1"
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// grpc clients can only manage memberships in teams below exec, so every
// write here sets NonRootOnly. the frontend uses the http routes, which have
// no such limit

type GRPCServer struct {
	membershipsv1.UnimplementedMembershipServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) Register(srv *grpc.Server) {
	membershipsv1.RegisterMembershipServiceServer(srv, s)
}

func (s *GRPCServer) CreateMembership(
	ctx context.Context, req *membershipsv1.CreateMembershipRequest,
) (*membershipsv1.CreateMembershipResponse, error) {
	orgID, err := parseUUID("org_id", req.GetOrgId())
	if err != nil {
		return nil, err
	}
	roleID, err := parseUUID("role_id", req.GetRoleId())
	if err != nil {
		return nil, err
	}
	userID, err := parseUUID("user_id", req.GetUserId())
	if err != nil {
		return nil, err
	}

	id, err := s.svc.CreateMembership(ctx, NewMembership{
		OrgID:       orgID,
		RoleID:      roleID,
		UserID:      userID,
		NonRootOnly: true,
	})
	if err != nil {
		return nil, toStatus("create membership", err)
	}

	return &membershipsv1.CreateMembershipResponse{MembershipId: id.String()}, nil
}

func (s *GRPCServer) ChangeMembershipRole(
	ctx context.Context, req *membershipsv1.ChangeMembershipRoleRequest,
) (*membershipsv1.ChangeMembershipRoleResponse, error) {
	orgID, err := parseUUID("org_id", req.GetOrgId())
	if err != nil {
		return nil, err
	}
	membershipID, err := parseUUID("membership_id", req.GetMembershipId())
	if err != nil {
		return nil, err
	}
	roleID, err := parseUUID("role_id", req.GetRoleId())
	if err != nil {
		return nil, err
	}
	// term is optional in the proto, so the generated field is nil when unset
	if req.Term != nil && req.GetTerm() == "" {
		return nil, status.Error(codes.InvalidArgument,
			"term can't be empty, leave it out to keep the old term")
	}

	res, err := s.svc.ChangeMembershipRole(ctx, RoleChange{
		OrgID:        orgID,
		MembershipID: membershipID,
		NewRoleID:    roleID,
		Term:         req.Term,
		NonRootOnly:  true,
	})
	if err != nil {
		return nil, toStatus("change membership role", err)
	}

	return &membershipsv1.ChangeMembershipRoleResponse{
		EndedMembershipId: res.EndedMembershipID.String(),
		MembershipId:      res.MembershipID.String(),
	}, nil
}

func (s *GRPCServer) ArchiveMembership(
	ctx context.Context, req *membershipsv1.ArchiveMembershipRequest,
) (*membershipsv1.ArchiveMembershipResponse, error) {
	orgID, err := parseUUID("org_id", req.GetOrgId())
	if err != nil {
		return nil, err
	}
	membershipID, err := parseUUID("membership_id", req.GetMembershipId())
	if err != nil {
		return nil, err
	}

	m, err := s.svc.ArchiveMembership(ctx, ArchiveInput{
		OrgID:        orgID,
		MembershipID: membershipID,
		NonRootOnly:  true,
	})
	if err != nil {
		return nil, toStatus("archive membership", err)
	}

	// archiving always sets ended_at, so it's never nil here
	return &membershipsv1.ArchiveMembershipResponse{
		MembershipId: m.ID.String(),
		EndedAt:      timestamppb.New(*m.EndedAt),
	}, nil
}

func (s *GRPCServer) ListUserPortfolios(
	ctx context.Context, req *membershipsv1.ListUserPortfoliosRequest,
) (*membershipsv1.ListUserPortfoliosResponse, error) {
	userID, err := parseUUID("user_id", req.GetUserId())
	if err != nil {
		return nil, err
	}

	rows, err := s.svc.ListUserPortfolios(ctx, userID)
	if err != nil {
		return nil, toStatus("list user portfolios", err)
	}

	portfolios := make([]*membershipsv1.Portfolio, len(rows))
	for i, p := range rows {
		portfolios[i] = &membershipsv1.Portfolio{
			MembershipId: p.MembershipID.String(),
			JoinedAt:     timestamppb.New(p.JoinedAt),
			RoleId:       p.RoleID.String(),
			RoleName:     p.RoleName,
			TeamId:       p.TeamID.String(),
			TeamName:     p.TeamName,
			OrgId:        p.OrgID.String(),
			OrgName:      p.OrgName,
			OrgSlug:      p.OrgSlug,
		}
	}

	return &membershipsv1.ListUserPortfoliosResponse{Portfolios: portfolios}, nil
}

func parseUUID(field, v string) (uuid.UUID, error) {
	id, err := uuid.Parse(v)
	if err != nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, field+" must be a uuid")
	}
	return id, nil
}

// toStatus maps service errors to grpc codes. anything unexpected is logged
// and hidden behind Internal
func toStatus(op string, err error) error {
	switch {
	case errors.Is(err, ErrRoleNotInOrg):
		return status.Error(codes.NotFound, "role not found in org")
	case errors.Is(err, ErrMembershipNotFound):
		return status.Error(codes.NotFound, "membership not found in org")
	case errors.Is(err, ErrRootTeamRole):
		return status.Error(codes.PermissionDenied,
			"memberships in the root team can't be managed over grpc")
	case errors.Is(err, ErrAlreadyMember):
		return status.Error(codes.AlreadyExists, "user already holds this role")
	case errors.Is(err, ErrMembershipEnded):
		return status.Error(codes.FailedPrecondition,
			"membership has already ended")
	case errors.Is(err, ErrSameRole):
		return status.Error(codes.InvalidArgument,
			"membership already has this role")
	default:
		log.Printf("%s: %v", op, err)
		return status.Error(codes.Internal, "could not "+op)
	}
}
