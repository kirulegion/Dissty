package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kirulegion/Dissty/services/auth-service/internal/domain"
	"gorm.io/gorm"
)

// INFO: Here there are two task which need to be handled therefore we are creating two interfaces , one interface can handle both but it makes the code messy.
//
// INFO:The AuthRepository handles the accounts part of the dealing.
type AuthRepository interface {
	Create(ctx context.Context, account *domain.Account) error
	FindAccountByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	FindAccountByUserID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	UpdateAccount(ctx context.Context, account *domain.Account) error
}

/*
INFO: why are we creating a struct ???

	-> because it's a simple backpack for carring the database connection
	   & since we have two interfaces and both have there share of functions
	   both of them need a different struct.
*/
type postgresAccountRepository struct {
	db *gorm.DB
}

// INFO: A constructor to safely create the struct.
func NewAccountRepository(db *gorm.DB) AuthRepository {
	return &postgresAccountRepository{db: db}
}

func (ap *postgresAccountRepository) Create(ctx context.Context, account *domain.Account) error {
	model := toModelAccount(account)

	return ap.db.WithContext(ctx).Create(&model).Error
}

func (ap *postgresAccountRepository) FindAccountByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	var model AccountModel
	err := ap.db.WithContext(ctx).Where("account_id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrAccountNotFound
	}

	if err != nil {
		return nil, err
	}

	return toDomainAccount(model), nil
}

func (ap *postgresAccountRepository) FindAccountByUserID(ctx context.Context, userid uuid.UUID) (*domain.Account, error) {
	var model AccountModel
	err := ap.db.WithContext(ctx).Where("user_id = ?", userid).First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrAccountNotFound
	}

	if err != nil {
		return nil, err
	}

	return toDomainAccount(model), nil
}

func (ap *postgresAccountRepository) UpdateAccount(ctx context.Context, account *domain.Account) error {
	model := toModelAccount(account)
	return ap.db.WithContext(ctx).Save(&model).Error
}

/*




 */

// INFO: Where as the IdentityProviderRepository handles the IdentityProvider side of the deal.
type IdentityProviderRepository interface {
	Create(ctx context.Context, provider *domain.IdentityProvider) error
	FindProviderByNameAndID(ctx context.Context, provider, providerID string) (*domain.IdentityProvider, error)
	FindAllProvidersByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.IdentityProvider, error)
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

// INFO: Same goes here this is the Identity Provider's backpack for carrying the db connection.
type postgresIdentityProviderRepository struct {
	db *gorm.DB
}

// INFO: A constructor to safely create the struct.
func NewIdentityProviderRepository(db *gorm.DB) IdentityProviderRepository {
	return &postgresIdentityProviderRepository{db: db}
}

/*
NOTE: Create the IdentityProvide , when creating the New user this Create a new auth for the user.
*/
func (ap *postgresIdentityProviderRepository) Create(ctx context.Context, provider *domain.IdentityProvider) error {
	model := toModelIdentityProvider(provider)

	return ap.db.WithContext(ctx).Create(&model).Error
}

/*
NOTE: Finding the Provider using identifier & provider it will be needed when we are signing an user and we need to check weather this user has has already linked his or her google account or not.
*/
func (ap *postgresIdentityProviderRepository) FindProviderByNameAndID(ctx context.Context, provider, providerID string) (*domain.IdentityProvider, error) {
	var model IdentityProviderModel
	err := ap.db.WithContext(ctx).Where("provider = ?", provider).Where("provider_id = ?", providerID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrAccountNotFound
	}

	if err != nil {
		return nil, err
	}

	return toDomainIdentityProvider(model), nil
}

/*
IDEA: This function finds all the provides the user has linked to Dissty , we can use this while showing user the settings tab of all linked platforms.
*/
func (ap *postgresIdentityProviderRepository) FindAllProvidersByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.IdentityProvider, error) {
	var model []IdentityProviderModel
	err := ap.db.WithContext(ctx).Where("account_id = ?", accountID).Find(&model).Error

	if err != nil {
		return nil, err
	}

	var result []*domain.IdentityProvider
	for _, m := range model {
		result = append(result, toDomainIdentityProvider(m))
	}

	return result, nil
}

/*
IDEA: The id -> providerID here , to update the lastused time.
*/
func (ap *postgresIdentityProviderRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return ap.db.WithContext(ctx).
		Model(&IdentityProviderModel{}).
		Where("id = ?", id).
		Update("last_used_at", time.Now()).Error
}








