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
		"INSERT INTO users (email, password_hash) VALUES ('%s', '%s')",
		u.Email,
		u.PasswordHash))

	return r.store.db.QueryRow(fmt.Sprintf("SELECT `id` FROM `users` WHERE `email` = '%s'", u.Email)).Scan(&u.ID)
}

func (r *UserRepository) Find(id int) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		fmt.Sprintf("SELECT `id`, `email`, `password_hash` FROM `users` WHERE `id` = '%d'", id)).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}
	if err := r.store.db.QueryRow(
		fmt.Sprintf("SELECT `id`, `email`, `password_hash` FROM `users` WHERE `email` = '%s'", email)).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}
