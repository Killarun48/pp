package modules

import (
	"database/sql"
	"app/internal/modules/user/repository"
)

type Repository struct {
	User repository.UserRepository
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		User: repository.NewUserRepository(db),
	}
}
