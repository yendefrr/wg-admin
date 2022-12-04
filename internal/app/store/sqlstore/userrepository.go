package sqlstore

import (
	"database/sql"
	"fmt"
	"go/wg-admin/internal/app/model"
	"go/wg-admin/internal/app/store"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(u *model.User) error {
	if err := u.Validate(); err != nil {
		return err
	}
	if err := u.BeforeCreate(); err != nil {
		return err
	}
	r.store.db.QueryRow(fmt.Sprintf(
		"INSERT INTO users (username) VALUES ('%s')", u.Username))

	return r.store.db.QueryRow(fmt.Sprintf("SELECT `id` FROM `users` WHERE `username` = '%s'", u.Username)).Scan(&u.ID)
}

func (r *UserRepository) GetAll() ([]model.User, error) {
	res, err := r.store.db.Query("SELECT DISTINCT u.id, u.username FROM users as u JOIN profiles as p on u.username = p.username;")
	if err != nil {
		return nil, err
	}

	var users []model.User
	for res.Next() {
		var user model.User
		err = res.Scan(
			&user.ID,
			&user.Username,
		)
		if err != nil {
			panic(err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) Find(id int) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		fmt.Sprintf("SELECT `id`, `username` FROM `users` WHERE `id` = '%d'", id)).Scan(
		&u.ID,
		&u.Username,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		fmt.Sprintf("SELECT `id`, `username` FROM `users` WHERE `username` = '%s'", username)).Scan(
		&u.ID,
		&u.Username,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}
