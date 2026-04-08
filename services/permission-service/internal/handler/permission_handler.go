package handler

import (
    "context"

    "github.com/kirulegion/Dissty/services/permission-service/internal/domain"
    "github.com/kirulegion/Dissty/services/permission-service/internal/service"
    pb "github.com/kirulegion/Dissty/services/permission-service/proto/permissionpb"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type PermissionHandler struct {
    pb.UnimplementedPermissionServiceServer
    svc service.PermissionService
}

func NewPermissionHandler(svc service.PermissionService) *PermissionHandler {
    return &PermissionHandler{svc: svc}
}

func (h *PermissionHandler) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
    result, err := h.svc.CheckPermission(ctx, &domain.PermissionRequest{
        UserID:      req.UserId,
        Role:        req.Role,
        Action:      domain.Action(req.Action),
        ContextType: domain.ContextType(req.ContextType),
        ContextID:   req.ContextId,
    })
    if err != nil {
        return nil, toGRPCError(err)
    }

    return &pb.CheckPermissionResponse{
        Allowed: result.Allowed,
        Reason:  result.Reason,
    }, nil
}

func toGRPCError(err error) error {
    switch err {
    case domain.ErrUnauthorized:
        return status.Error(codes.PermissionDenied, err.Error())
    case domain.ErrInvalidAction, domain.ErrInvalidRole, domain.ErrInvalidContext:
        return status.Error(codes.InvalidArgument, err.Error())
    default:
        return status.Error(codes.Internal, err.Error())
    }
}
