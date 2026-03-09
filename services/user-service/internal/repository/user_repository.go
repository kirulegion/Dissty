package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kirulegion/Dissty/services/user-service/internal/domain"
	"gorm.io/gorm"
)

// UserRepository defines what operations are available.
// The service layer only knows about this interface.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// postgresUserRepository is the PostgreSQL implementation
type postgresUserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new PostgreSQL-backed repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &postgresUserRepository{db: db}
}



func (r *postgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	model := toModel(user)
	return r.db.WithContext(ctx).Create(&model).Error
}



func (r *postgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(model), nil
}




func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(model), nil
}



func (r *postgresUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(model), nil
}

func (r *postgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	model := toModel(user)
	return r.db.WithContext(ctx).Save(&model).Error
}

func (r *postgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id).Error
}
