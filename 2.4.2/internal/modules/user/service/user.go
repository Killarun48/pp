package service

import (
	"context"
	"app/internal/models"
)

type UserServicer interface {
	Create(ctx context.Context, user models.User) error
	GetByID(ctx context.Context, id string) (models.User, error)
	Update(ctx context.Context, user models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, c models.Conditions) ([]models.User, error)
}

type UserRepositoryer interface {
	Create(ctx context.Context, user models.User) error
	GetByID(ctx context.Context, id string) (models.User, error)
	Update(ctx context.Context, user models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, c models.Conditions) ([]models.User, error)
}

type UserService struct {
	userRepository UserRepositoryer
}

func NewUserService(userRepository UserRepositoryer) UserServicer {
	return &UserService{
		userRepository: userRepository,
	}
}

func (us UserService) Create(ctx context.Context, user models.User) error {
	return us.userRepository.Create(ctx, user)
}

func (us UserService) GetByID(ctx context.Context, id string) (models.User, error) {
	return us.userRepository.GetByID(ctx, id)
}

func (us UserService) Update(ctx context.Context, user models.User) error {
	return us.userRepository.Update(ctx, user)
}

func (us UserService) Delete(ctx context.Context, id string) error {
	return us.userRepository.Delete(ctx, id)
}

func (us UserService) List(ctx context.Context, c models.Conditions) ([]models.User, error) {
	return us.userRepository.List(ctx, c)
}
