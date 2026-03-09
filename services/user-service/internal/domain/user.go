package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	DisplayName string
	UserName  string
	Email     string
	AvatarURL string
	Bio       string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
