package modules

import (
	"app/internal/modules/user/service"
)

type Service struct {
	User service.UserServicer
}

func NewService(repos *Repository) *Service {
	return &Service{
		User: service.NewUserService(repos.User),
	}
}