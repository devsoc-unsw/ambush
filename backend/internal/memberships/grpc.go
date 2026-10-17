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
	orgID, err := uuid.Parse(req.GetOrgId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "org_id must be a uuid")
	}
	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "role_id must be a uuid")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a uuid")
	}

	id, err := s.svc.CreateMembership(ctx, NewMembership{
		OrgID:  orgID,
		RoleID: roleID,
		UserID: userID,
	})
	if errors.Is(err, ErrRoleNotInOrg) {
		return nil, status.Error(codes.NotFound, "role not found in org")
	}
	if errors.Is(err, ErrAlreadyMember) {
		return nil, status.Error(codes.AlreadyExists, "user already holds this role")
	}
	if err != nil {
		log.Printf("create membership: %v", err)
		return nil, status.Error(codes.Internal, "could not create membership")
	}

	return &membershipsv1.CreateMembershipResponse{MembershipId: id.String()}, nil
}

func (s *GRPCServer) ListUserPortfolios(
	ctx context.Context, req *membershipsv1.ListUserPortfoliosRequest,
) (*membershipsv1.ListUserPortfoliosResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a uuid")
	}

	rows, err := s.svc.ListUserPortfolios(ctx, userID)
	if err != nil {
		log.Printf("list user portfolios: %v", err)
		return nil, status.Error(codes.Internal, "could not list portfolios")
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
