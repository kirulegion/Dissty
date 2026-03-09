package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kirulegion/Dissty/services/user-service/internal/domain"
	"github.com/kirulegion/Dissty/services/user-service/internal/repository"
)

// UserService defines what operations exist on users.
// The handler layer depends on this interface — never the concrete struct.
type UserService interface {
	CreateUser(ctx context.Context, displayname, username, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// userService is the private concrete implementation of UserService.
// It holds a reference to the repository so it can delegate storage.
type userService struct {
	repo repository.UserRepository
}

// NewUserService is the only way to create a userService.
// It takes a UserRepository interface — not a concrete PostgreSQL type.
// This means you can pass any implementation — real DB, mock, in-memory.
// It returns UserService interface so callers never see the concrete type.
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, displayname, username, email string) (*domain.User, error) {
	// INFO: RULE 1 : The email must unique.
	existing , err := s.repo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// INFO: RULE 2 : The username must unique.
	existing, err = s.repo.FindByUsername(ctx, username)
	if existing != nil {
		return nil, domain.ErrUsernameAlreadyExists
	}

	// Silence the "err declared but not used" compiler error.
	// err was assigned above but we only checked existing, not err directly.
	_ = err
	
	if username == "" || email == "" || displayname == "" {
		return nil , domain.ErrInvalidInput
	}

	newUser := &domain.User{
		ID:          uuid.New(),
		UserName:    username,
		DisplayName: displayname,
		Email:       email,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail fetches a user by their email address.
// Used primarily by auth-service during login —
// auth-service finds the user by email, then verifies their password.
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser applies changes to an existing user.
// Rule 1 — verify the user exists before attempting an update.
// Without this check, updating a non-existent user would silently do nothing.
// Rule 2 — always refresh UpdatedAt so we know when the last change happened.
func (s *userService) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	// Verify existence first.
	// We discard the returned user with _ because we only care that it exists.
	// The actual update uses the user passed in as a parameter.
	_, err := s.repo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser removes a user from Dissty permanently.
// Rule — verify existence before deleting.
// Without this, deleting a non-existent UUID would silently succeed,user_ser
// making it impossible to tell "deleted" from "never existed."
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Verify existence. Discard the user — we only need to confirm it's there.
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.ErrUserNotFound
	}

	// Delegate the actual deletion to the repository.
	// Only returns error — nothing meaningful to return after a deletion.
	return s.repo.Delete(ctx, id)
}
