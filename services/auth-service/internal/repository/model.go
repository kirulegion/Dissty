package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/kirulegion/Dissty/services/auth-service/internal/domain"
)

type AccountModel struct {
	AccountID  uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID     *uuid.UUID `gorm:"type:uuid"`
	IsComplete bool       `gorm:"default:false;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (AccountModel) TableName() string {
	return "accounts"
}

func toModelAccount(d *domain.Account) AccountModel {
	return AccountModel{
		AccountID:  d.AccountID,
		UserID:     d.UserID,
		IsComplete: d.IsComeplete,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
}

func toDomainAccount(m AccountModel) *domain.Account {
	return &domain.Account{
		AccountID:  m.AccountID,
		UserID:     m.UserID,
		IsComeplete: m.IsComplete,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

/*



 */

type IdentityProviderModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	AccountID  uuid.UUID `gorm:"type:uuid;not null"`
	Provider   string    `gorm:"not null"`
	ProviderID string    `gorm:"not null"`
	Identifier string    `gorm:"not null"`
	CreatedAt  time.Time
	LastUsedAt time.Time
}

func (IdentityProviderModel) TableName() string {
	return "identity_providers"
}

func toDomainIdentityProvider(m IdentityProviderModel) *domain.IdentityProvider {
	return &domain.IdentityProvider{
		ID:         m.ID,
		AccountID:  m.AccountID,
		Provider:   m.Provider,
		ProviderID: m.ProviderID,
		Identifier: m.Identifier,
		CreatedAt:  m.CreatedAt,
		LastUsedAt: m.LastUsedAt,
	}
}

func toModelIdentityProvider(d *domain.IdentityProvider) IdentityProviderModel {
	return IdentityProviderModel{
		ID:         d.ID,
		AccountID:  d.AccountID,
		Provider:   d.Provider,
		ProviderID: d.ProviderID,
		Identifier: d.Identifier,
		CreatedAt:  d.CreatedAt,
		LastUsedAt: d.LastUsedAt,
	}
}
