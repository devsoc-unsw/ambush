package audit

import (
	"backend/db"
	auditv1 "backend/gen/ambush/audit/v1"
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
	auditv1.UnimplementedAuditServiceServer
	svc *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) Register(srv *grpc.Server) {
	auditv1.RegisterAuditServiceServer(srv, s)
}

func (s *GRPCServer) GetAuditEntry(
	ctx context.Context, req *auditv1.GetAuditEntryRequest,
) (*auditv1.GetAuditEntryResponse, error) {
	entry, err := s.svc.GetAuditEntry(ctx, req.GetId())
	if errors.Is(err, ErrNotFound) {
		return nil, status.Error(codes.NotFound, "audit entry not found")
	}
	if err != nil {
		log.Printf("get audit entry: %v", err)
		return nil, status.Error(codes.Internal, "could not get audit entry")
	}

	return &auditv1.GetAuditEntryResponse{Entry: entryToProto(entry)}, nil
}

// db enums -> proto enums; anything unmapped falls back to UNSPECIFIED (0)
var entityToProto = map[db.AuditEntity]auditv1.AuditEntity{
	db.AuditEntitySociety:    auditv1.AuditEntity_AUDIT_ENTITY_SOCIETY,
	db.AuditEntitySocrole:    auditv1.AuditEntity_AUDIT_ENTITY_SOCROLE,
	db.AuditEntityMembership: auditv1.AuditEntity_AUDIT_ENTITY_MEMBERSHIP,
}

var actionToProto = map[db.AuditAction]auditv1.AuditAction{
	db.AuditActionCreate: auditv1.AuditAction_AUDIT_ACTION_CREATE,
	db.AuditActionUpdate: auditv1.AuditAction_AUDIT_ACTION_UPDATE,
	db.AuditActionDelete: auditv1.AuditAction_AUDIT_ACTION_DELETE,
}

func entryToProto(e db.AuditLog) *auditv1.AuditEntry {
	out := &auditv1.AuditEntry{
		Id:          e.ID,
		ChangeSetId: uuid.UUID(e.ChangeSetID.Bytes).String(),
		SocietyId:   e.SocID,
		EntityType:  entityToProto[e.EntityType],
		EntityId:    e.EntityID,
		Action:      actionToProto[e.Action],
		ActorId:     uuid.UUID(e.ActorID.Bytes).String(),
		CreatedAt:   timestamppb.New(e.CreatedAt.Time),
	}
	// jsonb columns are nil when NULL (before on create, after on delete)
	if e.Before != nil {
		before := string(e.Before)
		out.BeforeJson = &before
	}
	if e.After != nil {
		after := string(e.After)
		out.AfterJson = &after
	}
	if e.UndoesID.Valid {
		out.UndoesId = &e.UndoesID.Int64
	}

	return out
}
