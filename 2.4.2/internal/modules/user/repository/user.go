package repository

import (
	"app/internal/models"
	"context"
	"database/sql"
	"log"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) error
	GetByID(ctx context.Context, id string) (models.User, error)
	Update(ctx context.Context, user models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, c models.Conditions) ([]models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r userRepository) Create(ctx context.Context, user models.User) error {
	_, err := sq.Insert("users").
		Columns("name", "email").
		Values(user.Name, user.Email).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (r userRepository) GetByID(ctx context.Context, id string) (models.User, error) {
	var user models.User

	row := sq.Select("*").
		From("users").
		Where(sq.Eq{
			"id": id,
		}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).QueryRow()
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.DeletedAt)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r userRepository) Update(ctx context.Context, user models.User) error {
	res, err := sq.Update("users").
		SetMap(map[string]interface{}{
			"name":  user.Name,
			"email": user.Email,
		}).
		Where(sq.Eq{"id": user.ID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).Exec()
	if err != nil {
		return err
	}

	if v, err := res.RowsAffected(); err == nil {
		if v == 0 {
			return sql.ErrNoRows
		}
	}

	return nil
}

func (ur userRepository) Delete(ctx context.Context, id string) error {
	_, err := sq.Update("users").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		RunWith(ur.db).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (ur userRepository) List(ctx context.Context, c models.Conditions) ([]models.User, error) {
	if c.Limit == "" {
		c.Limit = "0"
	}

	if c.Offset == "" {
		c.Offset = "0"
	}

	limit, err := strconv.Atoi(c.Limit)
	if err != nil {
		return nil, err
	}

	offset, err := strconv.Atoi(c.Offset)
	if err != nil {
		return nil, err
	}

	var builder interface{} = sq.Select("*").From("users")

	if limit > 0 {
		builder = builder.(sq.SelectBuilder).Limit(uint64(limit))
	}

	if offset > 0 {
		builder = builder.(sq.SelectBuilder).Offset(uint64(offset))
	}

	rows, err := builder.(sq.SelectBuilder).
		PlaceholderFormat(sq.Dollar).
		RunWith(ur.db).Query()
	if err != nil {
		return nil, err
	}

	var users []models.User

	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.DeletedAt)
		if err != nil {
			log.Println(err)
		}
		users = append(users, user)
	}

	return users, nil
}
