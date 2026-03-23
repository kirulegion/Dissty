package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/kirulegion/Dissty/services/auth-service/internal/cache"
	"github.com/kirulegion/Dissty/services/auth-service/internal/domain"
	"github.com/kirulegion/Dissty/services/auth-service/internal/repository"
	"github.com/kirulegion/Dissty/services/auth-service/internal/token"
	userpb "github.com/kirulegion/Dissty/services/user-service/proto"
	"github.com/resend/resend-go/v2"
)

type AuthService interface {
	// OAuth + OTP combined login/register flow
	AuthenticateWithOAuth(ctx context.Context, provider domain.Provider, identifier, providerID, displayName, avatarURL string) (string, error)

	// Called after OAuth when user picks their username
	CompleteRegistration(ctx context.Context, accountID uuid.UUID, username, displayname string) (string, error)

	// OTP flow — step 1: send code to email/phone
	RequestOTP(ctx context.Context, provider domain.Provider, identifier string) error

	ValidateToken(ctx context.Context, token string) (*token.Claims, error)

	// OTP flow — step 2: verify code, return JWT
	VerifyOTP(ctx context.Context, provider domain.Provider, identifier, code string) (string, error)

	// Settings — link a new provider to existing account
	LinkProvider(ctx context.Context, accountID uuid.UUID, provider domain.Provider, identifier, providerID string) error

	// Settings — get all linked providers for this account
	GetLinkedProviders(ctx context.Context, accountID uuid.UUID) ([]*domain.IdentityProvider, error)
}

type authService struct {
	accountRepo  repository.AuthRepository
	providerRepo repository.IdentityProviderRepository
	userClient   userpb.UserServiceClient
	otpCache     *cache.OTPCache
	resendClient *resend.Client
	tokenSvc     token.TokenService
}

func NewAuthService(account repository.AuthRepository, provider repository.IdentityProviderRepository, userClient userpb.UserServiceClient, resendClient *resend.Client, otpCache *cache.OTPCache, tokenSvc token.TokenService) AuthService {
	return &authService{
		accountRepo:  account,
		providerRepo: provider,
		userClient:   userClient,
		resendClient: resendClient,
		otpCache:     otpCache,
		tokenSvc:     tokenSvc,
	}
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*token.Claims, error) {
	return s.tokenSvc.ValidateToken(tokenString)
}
