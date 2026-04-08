package service

import (
    "context"
    "github.com/kirulegion/Dissty/services/permission-service/internal/domain"
)

type PermissionService interface {
    CheckPermission(ctx context.Context, req *domain.PermissionRequest) (*domain.PermissionResult, error)
}

type permissionService struct{}

func NewPermissionService() PermissionService {
    return &permissionService{}
}

func (s *permissionService) CheckPermission(ctx context.Context, req *domain.PermissionRequest) (*domain.PermissionResult, error) {
    if req.Role == "" || req.Action == "" {
        return nil, domain.ErrInvalidAction
    }

    allowed := s.evaluate(req.Role, req.Action)
    reason := "allowed"
    if !allowed {
        reason = "insufficient permissions"
    }

    return &domain.PermissionResult{Allowed: allowed, Reason: reason}, nil
}

func (s *permissionService) evaluate(role string, action domain.Action) bool {
    switch role {
    case "owner":
        return true
    case "moderator":
        switch action {
        case domain.ActionSendMessage,
            domain.ActionDeleteMessage,
            domain.ActionBanMember,
            domain.ActionKickMember,
            domain.ActionStartSession,
            domain.ActionViewChannel:
            return true
        default:
            return false
        }
    case "member":
        switch action {
        case domain.ActionSendMessage,
            domain.ActionViewChannel:
            return true
        default:
            return false
        }
    case "guest":
        switch action {
        case domain.ActionViewChannel:
            return true
        default:
            return false
        }
    default:
        return false
    }
}
