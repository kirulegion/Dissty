package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/kirulegion/Dissty/services/user-service/internal/domain"
)

// INFO: The user model is the database representation of the a user.
type UserModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	DisplayName string    `gorm:"not null"`
	Username    string    `gorm:"uniqueIndex;not null"`
	Email       string    `gorm:"uniqueIndex;not null"`
	AvatarURL   string
	Bio         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}

// INFO: Gorm names the tablesa in snake_case this table lets it know the name of the table is "users".
func (UserModel) TableName() string {
	return "users"
}

// INFO: After fetching a row from the database, GORM gives us back a UserModel. But the service layer doesn't know what a UserModel is — it only speaks domain.User. This function translates from database language to domain language.
func toDomain(m UserModel) *domain.User {
	return &domain.User{
		ID:          m.ID,
		UserName:    m.Username,
		DisplayName: m.DisplayName,
		Email:       m.Email,
		AvatarURL:   m.AvatarURL,
		Bio:         m.Bio,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}
}

// INFO: This is the reverse — when the service layer wants to save a user, it passes a `domain.User` down. The repository can't give that directly to GORM because GORM needs a `UserModel` with its tags to know how to write to the database. This function converts in the other direction.
func toModel(u *domain.User) UserModel {
	return UserModel{
		ID:          u.ID,
		Username:    u.UserName,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		AvatarURL:   u.AvatarURL,
		Bio:         u.Bio,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		DeletedAt:   u.DeletedAt,
	}
}
