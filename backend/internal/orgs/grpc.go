package orgs

import (
	orgsv1 "backend/gen/ambush/orgs/v1"
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer serves the read-only org structure to client apps
type GRPCServer struct {
	orgsv1.UnimplementedOrgStructureServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) Register(srv *grpc.Server) {
	orgsv1.RegisterOrgStructureServiceServer(srv, s)
}

func (s *GRPCServer) GetOrgStructure(
	ctx context.Context, req *orgsv1.GetOrgStructureRequest,
) (*orgsv1.GetOrgStructureResponse, error) {
	orgID, err := uuid.Parse(req.GetOrgId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "org_id must be a uuid")
	}

	st, err := s.svc.GetOrgStructure(ctx, orgID)
	if errors.Is(err, ErrOrgNotFound) {
		return nil, status.Error(codes.NotFound, "org not found")
	}
	if err != nil {
		log.Printf("get org structure: %v", err)
		return nil, status.Error(codes.Internal, "could not get org structure")
	}

	return toStructurePB(st), nil
}

func toStructurePB(st OrgStructure) *orgsv1.GetOrgStructureResponse {
	out := &orgsv1.GetOrgStructureResponse{
		Org: &orgsv1.Org{
			Id:       st.Org.ID.String(),
			Name:     st.Org.Name,
			Slug:     st.Org.Slug,
			Archived: st.Org.ArchivedAt != nil,
		},
		Teams:       make([]*orgsv1.Team, len(st.Teams)),
		Roles:       make([]*orgsv1.Role, len(st.Roles)),
		Memberships: make([]*orgsv1.Membership, len(st.Memberships)),
	}

	for i, t := range st.Teams {
		out.Teams[i] = &orgsv1.Team{
			Id:           t.ID.String(),
			Name:         t.Name,
			ParentId:     uuidPtrString(t.ParentID),
			LeaderRoleId: uuidPtrString(t.LeaderRoleID),
		}
	}
	for i, r := range st.Roles {
		out.Roles[i] = &orgsv1.Role{
			Id:     r.ID.String(),
			TeamId: r.TeamID.String(),
			Name:   r.Name,
		}
	}
	for i, m := range st.Memberships {
		out.Memberships[i] = &orgsv1.Membership{
			Id:     m.ID.String(),
			RoleId: m.RoleID.String(),
			UserId: m.UserID.String(),
		}
	}

	return out
}

// optional proto fields are *string, so nil stays unset
func uuidPtrString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}
