package domain

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	AccountID uuid.UUID
	UserID    *uuid.UUID
	IsComplete bool 
	CreatedAt time.Time
	UpdatedAt time.Time
}

type IdentityProvider struct {
	ID           uuid.UUID
	AccountID    uuid.UUID
	Provider     Provider
	ProviderID   string
	Identifier   string
	CreatedAt    time.Time
	LastUsedAt   time.Time
}
