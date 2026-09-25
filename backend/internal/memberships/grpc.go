package memberships

import (
	membershipsv1 "backend/gen/ambush/memberships/v1"
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a uuid")
	}

	id, err := s.svc.CreateMembership(ctx, NewMembership{
		SocietyID: req.GetSocietyId(),
		RoleID:    req.GetRoleId(),
		UserID:    userID,
	})
	if errors.Is(err, ErrRoleNotInSociety) {
		return nil, status.Error(codes.NotFound, "role not found in society")
	}
	if err != nil {
		log.Printf("create membership: %v", err)
		return nil, status.Error(codes.Internal, "could not create membership")
	}

	return &membershipsv1.CreateMembershipResponse{MembershipId: id}, nil
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
	for i, r := range rows {
		portfolios[i] = &membershipsv1.Portfolio{
			MembershipId:     r.MembershipID,
			JoinedAt:         timestamppb.New(r.JoinedAt.Time),
			RoleId:           r.RoleID,
			RoleName:         r.RoleName,
			RoleParentRoleId: int4Ptr(r.RoleParentRoleID),
			SocietyId:        r.SocietyID,
			SocietyName:      r.SocietyName,
			SocietySlug:      r.SocietySlug,
		}
	}

	return &membershipsv1.ListUserPortfoliosResponse{Portfolios: portfolios}, nil
}

// int4Ptr maps a nullable db int to an optional proto field
func int4Ptr(v pgtype.Int4) *int32 {
	if !v.Valid {
		return nil
	}
	return &v.Int32
}
