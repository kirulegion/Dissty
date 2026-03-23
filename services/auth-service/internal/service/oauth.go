package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kirulegion/Dissty/services/auth-service/internal/cache"
	"github.com/kirulegion/Dissty/services/auth-service/internal/domain"
	"github.com/kirulegion/Dissty/services/auth-service/internal/repository"
	"github.com/resend/resend-go/v2"
)

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

func (s *authService) AuthenticateWithOAuth(ctx context.Context, provider, providerID, identifier, displayname, avatarurl string) (string, error) {
	exist, err := s.providerRepo.FindProviderByNameAndID(ctx, provider, providerID)

	if errors.Is(err, domain.ErrAccountNotFound) {
		//Since we did not found any provider linked to this email we are going to actually register the user.

		//Creating a new account but we don't have any userID yet since the user isn't registered on dissty rn.
		newAccount := &domain.Account{
			AccountID:  uuid.New(),
			UserID:     nil,
			IsComplete: false,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		//Let's now create a provider for this user.
		newProvider := &domain.IdentityProvider{
			ID:         uuid.New(),
			AccountID:  newAccount.AccountID,
			LastUsedAt: time.Now(),
			Provider:   provider,
			ProviderID: providerID,
			Identifier: identifier,
			CreatedAt:  time.Now(),
		}

		if err := s.accountRepo.Create(ctx, newAccount); err != nil {
			return "", err
		}

		if err := s.providerRepo.Create(ctx, newProvider); err != nil {
			return "", err
		}

		return "INCOMPLETE_JWT", nil
	}
	err = s.providerRepo.UpdateLastUsed(ctx, exist.ID)
	if err != nil {
		return "", err
	}
	return "JWT", nil
}

func (s *authService) CompleteRegistration(ctx context.Context, accountID uuid.UUID, username, displayname string) (string, error) {
	exist, err := s.providerRepo.FindAllProvidersByAccountID(ctx, accountID)
	if err != nil {
		return "", err
	}

	if len(exist) == 0 {
		return "", domain.ErrAccountNotFound
	}

	email := exist[0].Identifier
	user, err := s.userClient.CreateUser(ctx, &userpb.CreateUserRequest{
		Username:    username,
		Email:       email,
		DisplayName: displayname,
	})

	if err != nil {
		return "", err
	}

	account, err := s.accountRepo.FindAccountByID(ctx, accountID)
	if err != nil {
		return "", err
	}

	userID, err := uuid.Parse(user.User.Id)
	if err != nil {
		return "", err
	}
	account.UserID = &userID
	account.UpdatedAt = time.Now()
	account.IsComplete = true

	if err := s.accountRepo.UpdateAccount(ctx, account); err != nil {
		return "", err
	}
	return "COMPLETE_JWT", nil
}



