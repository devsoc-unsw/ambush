package socroles

import (
	socrolesv1 "backend/gen/ambush/socroles/v1"
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	socrolesv1.UnimplementedSocRoleServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) Register(srv *grpc.Server) {
	socrolesv1.RegisterSocRoleServiceServer(srv, s)
}

func (s *GRPCServer) CreateSocRole(
	ctx context.Context, req *socrolesv1.CreateSocRoleRequest,
) (*socrolesv1.CreateSocRoleResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name required")
	}

	id, err := s.svc.CreateSocRole(ctx, NewSocRole{
		SocietyID:    req.GetSocietyId(),
		ParentRoleID: req.ParentRoleId, // field, not getter: keeps nil as unset
		Name:         req.GetName(),
	})
	if err != nil {
		log.Printf("create socrole: %v", err)
		return nil, status.Error(codes.Internal, "could not create socrole")
	}

	return &socrolesv1.CreateSocRoleResponse{RoleId: id}, nil
}
