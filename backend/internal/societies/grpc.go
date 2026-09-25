package societies

import (
	societiesv1 "backend/gen/ambush/societies/v1"
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	societiesv1.UnimplementedSocietyServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) Register(srv *grpc.Server) {
	societiesv1.RegisterSocietyServiceServer(srv, s)
}

func (s *GRPCServer) CreateSociety(
	ctx context.Context, req *societiesv1.CreateSocietyRequest,
) (*societiesv1.CreateSocietyResponse, error) {
	if req.GetName() == "" || req.GetSlug() == "" {
		return nil, status.Error(codes.InvalidArgument, "name and slug required")
	}

	id, err := s.svc.CreateSociety(ctx, NewSociety{
		Name: req.GetName(),
		Slug: req.GetSlug(),
	})
	if err != nil {
		log.Printf("create society: %v", err)
		return nil, status.Error(codes.Internal, "could not create society")
	}

	return &societiesv1.CreateSocietyResponse{SocietyId: id}, nil
}
