package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/kirulegion/Dissty/services/auth-service/internal/cache"
	"github.com/kirulegion/Dissty/services/auth-service/internal/domain"
	"github.com/kirulegion/Dissty/services/auth-service/internal/repository"
	userpb "github.com/kirulegion/Dissty/services/user-service/proto"
	"github.com/resend/resend-go/v2"
)

type AuthService interface {
	// OAuth + OTP combined login/register flow
	AuthenticateWithOAuth(ctx context.Context, provider, identifier, providerID, displayName, avatarURL string) (string, error)

	// Called after OAuth when user picks their username
	CompleteRegistration(ctx context.Context, accountID uuid.UUID, username string) (string, error)

	// OTP flow — step 1: send code to email/phone
	RequestOTP(ctx context.Context, provider, identifier string) error

	// OTP flow — step 2: verify code, return JWT
	VerifyOTP(ctx context.Context, identifier, code string, provider domain.Provider) (string, error)

	// Settings — link a new provider to existing account
	LinkProvider(ctx context.Context, accountID uuid.UUID, provider, identifier, providerID string) error

	// Settings — get all linked providers for this account
	GetLinkedProviders(ctx context.Context, accountID uuid.UUID) ([]*domain.IdentityProvider, error)
}

type authService struct {
	accountRepo  repository.AuthRepository
	providerRepo repository.IdentityProviderRepository
	userClient   userpb.UserServiceClient
	otpCache     *cache.OTPCache
	resendClient *resend.Client
}

func NewAuthService(account repository.AuthRepository, provider repository.IdentityProviderRepository, userClient userpb.UserServiceClient, resendClient *resend.Client,
	otpCache *cache.OTPCache) AuthService {
	return &authService{
		accountRepo:  account,
		providerRepo: provider,
		userClient:   userClient,
		resendClient: resendClient,
		otpCache:     otpCache,
	}
}
