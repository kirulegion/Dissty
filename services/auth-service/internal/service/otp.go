package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/kirulegion/Dissty/services/auth-service/internal/domain"
	"github.com/resend/resend-go/v2"
)

func (s *authService) RequestOTP(ctx context.Context, provider domain.Provider, identifier string) error {
	if !provider.IsValid() || !provider.IsOTP() {
		return domain.ErrProviderNotFound
	}

	code := fmt.Sprintf("%06d", rand.Intn(999999))

	if err := s.otpCache.Set(ctx, identifier, code); err != nil {
		return err
	}

	params := &resend.SendEmailRequest{
		From:    "Dissty <noreply@yourdomain.com>",
		To:      []string{identifier},
		Subject: "Your Dissty login code",
		Html:    fmt.Sprintf("<p>Your login code is: <strong>%s</strong></p><p>Expires in 10 minutes.</p>", code),
	}

	_, err := s.resendClient.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}

func (s *authService) VerifyOTP(ctx context.Context, provider domain.Provider, identifier, code string) (string, error) {
	otp, err := s.otpCache.Get(ctx, identifier)

	if err != nil {
		return "", domain.ErrExpiredOTP
	}

	if otp != code {
		return "", domain.ErrWrongOTP
	}

	//Removing the otp once there is no use of it.
	s.otpCache.Delete(ctx, identifier)

	//Checking if the user is old or not.
	var existing *domain.IdentityProvider
	existing, err = s.providerRepo.FindProviderByNameAndID(ctx, string(provider), identifier)

	if errors.Is(err, domain.ErrAccountNotFound) {
		newAccount := &domain.Account{
			AccountID:  uuid.New(),
			IsComplete: false,
			UpdatedAt:  time.Now(),
			CreatedAt:  time.Now(),
		}

		newProvider := &domain.IdentityProvider{
			ID:         uuid.New(),
			AccountID:  newAccount.AccountID,
			Provider:   provider,
			ProviderID: identifier,
			Identifier: identifier,
			CreatedAt:  time.Now(),
			LastUsedAt: time.Now(),
		}

		if err := s.accountRepo.Create(ctx, newAccount); err != nil {
			return "", err
		}

		if err := s.providerRepo.Create(ctx, newProvider); err != nil {
			return "", err
		}

		return s.tokenSvc.GenerateIncompleteToken(newAccount.AccountID)

	}

	if err != nil {
		return "", err
	}

	// returning user
	if err = s.providerRepo.UpdateLastUsed(ctx, existing.ID); err != nil {
		return "", err
	}

	// returning user — check IsComplete first
	account, err := s.accountRepo.FindAccountByID(ctx, existing.AccountID)
	if err != nil {
		return "", err
	}
	if !account.IsComplete {
		return s.tokenSvc.GenerateIncompleteToken(account.AccountID)
	}
	return s.tokenSvc.GenerateCompleteToken(account.AccountID, *account.UserID)
}
