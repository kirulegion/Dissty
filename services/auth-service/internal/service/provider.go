package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kirulegion/Dissty/services/auth-service/internal/domain"
)

func (s *authService) LinkProvider(ctx context.Context, accountID uuid.UUID, provider domain.Provider, identifier, providerID string) error {
	existing, err := s.providerRepo.FindProviderByNameAndID(ctx, string(provider), providerID)
	if existing != nil {
		return domain.ErrProviderLinkExists
	}

	if !errors.Is(err, domain.ErrAccountNotFound) && err != nil {
		return err
	}

	newProvider := &domain.IdentityProvider{
		ID:         uuid.New(),
		AccountID:  accountID,
		Provider:   domain.Provider(provider),
		ProviderID: providerID,
		Identifier: identifier,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	if err := s.providerRepo.Create(ctx, newProvider); err != nil {
		return err
	}

	return nil
}

func (s *authService) GetLinkedProviders(ctx context.Context, accountID uuid.UUID) ([]*domain.IdentityProvider, error) {
	linkedProviderList, err := s.providerRepo.FindAllProvidersByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return linkedProviderList, nil
}
