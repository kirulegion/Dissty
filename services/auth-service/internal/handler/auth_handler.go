package handler

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kirulegion/Dissty/services/auth-service/internal/domain"
	"github.com/kirulegion/Dissty/services/auth-service/internal/service"
	pb "github.com/kirulegion/Dissty/services/auth-service/proto"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) AuthenticateWithOAuth(ctx context.Context, req *pb.AuthenticateWithOAuthRequest) (*pb.TokenResponse, error) {
	token, err := h.svc.AuthenticateWithOAuth(
		ctx,
		domain.Provider(req.Provider),
		req.ProviderId,
		req.Identifier,
		req.DisplayName,
		req.AvatarUrl,
	)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.TokenResponse{Token: token}, nil
}

func (h *AuthHandler) CompleteRegistration(ctx context.Context, req *pb.CompleteRegistrationRequest) (*pb.TokenResponse, error) {
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account id: %v", err)
	}
	token, err := h.svc.CompleteRegistration(ctx, accountID, req.Username, req.DisplayName)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.TokenResponse{Token: token}, nil
}

func (h *AuthHandler) RequestOTP(ctx context.Context, req *pb.RequestOTPRequest) (*pb.RequestOTPResponse, error) {
	err := h.svc.RequestOTP(ctx, domain.Provider(req.Provider), req.Identifier)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RequestOTPResponse{Success: true}, nil
}

func (h *AuthHandler) VerifyOTP(ctx context.Context, req *pb.VerifyOTPRequest) (*pb.TokenResponse, error) {
	token, err := h.svc.VerifyOTP(ctx, domain.Provider(req.Provider), req.Identifier, req.Code)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.TokenResponse{Token: token}, nil
}

func (h *AuthHandler) LinkProvider(ctx context.Context, req *pb.LinkProviderRequest) (*pb.LinkProviderResponse, error) {
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account id: %v", err)
	}
	err = h.svc.LinkProvider(ctx, accountID, domain.Provider(req.Provider), req.Identifier, req.ProviderId)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.LinkProviderResponse{Success: true}, nil
}

func (h *AuthHandler) GetLinkedProviders(ctx context.Context, req *pb.GetLinkedProvidersRequest) (*pb.GetLinkedProvidersResponse, error) {
	accountID, err := uuid.Parse(req.AccountId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account id: %v", err)
	}
	providers, err := h.svc.GetLinkedProviders(ctx, accountID)
	if err != nil {
		return nil, toGRPCError(err)
	}
	var pbProviders []*pb.IdentityProvider
	for _, p := range providers {
		pbProviders = append(pbProviders, &pb.IdentityProvider{
			Id:         p.ID.String(),
			AccountId:  p.AccountID.String(),
			Provider:   string(p.Provider),
			ProviderId: p.ProviderID,
			Identifier: p.Identifier,
			CreatedAt:  p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastUsedAt: p.LastUsedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return &pb.GetLinkedProvidersResponse{Providers: pbProviders}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := h.svc.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}
	return &pb.ValidateTokenResponse{
		AccountId: claims.AccountID,
		UserId:    claims.UserID,
		Status:    claims.Status,
		Valid:     true,
	}, nil
}

func toGRPCError(err error) error {
	switch err {
	case domain.ErrAccountNotFound:
		return status.Errorf(codes.NotFound, "account not found")
	case domain.ErrProviderLinkExists:
		return status.Errorf(codes.AlreadyExists, "provider already linked")
	case domain.ErrProviderNotFound:
		return status.Errorf(codes.NotFound, "provider not found")
	case domain.ErrWrongOTP:
		return status.Errorf(codes.InvalidArgument, "wrong otp code")
	case domain.ErrExpiredOTP:
		return status.Errorf(codes.DeadlineExceeded, "otp has expired")
	case domain.ErrOTPAlreadyUsed:
		return status.Errorf(codes.AlreadyExists, "otp already used")
	default:
		return status.Errorf(codes.Internal, "internal error")
	}
}
