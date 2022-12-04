package sqlstore

import (
	"database/sql"
	"fmt"
	"go/wg-admin/internal/app/model"
)

type ProfileRepository struct {
	store *Store
}

func (r *ProfileRepository) Create(p *model.Profile) error {
	u := &model.User{
		Username: p.Username,
	}

	if err := r.store.db.QueryRow(fmt.Sprintf("SELECT `id` FROM `users` WHERE `username` = '%s'", p.Username)).Scan(&u.ID); err != nil {
		if err != sql.ErrNoRows {
			return err
		}

		if err := r.store.User().Create(u); err != nil {
			return err
		}
	}

	r.store.db.QueryRow(fmt.Sprintf(
		"INSERT INTO `profiles` (`username`, `type`, `path`, `publickey`, `privatekey`) VALUES ('%s', '%s', '%s', '%s', '%s')",
		p.Username, p.Type, p.Path, p.Publickey, p.Privatekey))

	return r.store.db.QueryRow(fmt.Sprintf("SELECT `id` FROM `profiles` WHERE `path` = '%s'", p.Path)).Scan(&p.ID)
}

func (r *ProfileRepository) Delete(id int) error {
	r.store.db.QueryRow(fmt.Sprintf("DELETE FROM `profiles` WHERE `id` = %d", id))

	return nil
}

func (r *ProfileRepository) GetAll() ([]model.Profile, error) {
	res, err := r.store.db.Query("SELECT * FROM `profiles`")
	if err != nil {
		return nil, err
	}

	var profiles []model.Profile
	for res.Next() {
		var profile model.Profile
		err = res.Scan(
			&profile.ID,
			&profile.Username,
			&profile.Type,
			&profile.Path,
			&profile.Publickey,
			&profile.Privatekey,
			&profile.IsActive,
		)
		if err != nil {
			panic(err)
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (r *ProfileRepository) Find(id int) (*model.Profile, error) {
	p := &model.Profile{}

	if err := r.store.db.QueryRow(fmt.Sprintf("SELECT * FROM `profiles` WHERE `id` = %d", id)).Scan(
		&p.ID,
		&p.Username,
		&p.Type,
		&p.Path,
		&p.Publickey,
		&p.Privatekey,
		&p.IsActive,
	); err != nil {
		return nil, err
	}

	return p, nil
}

func (r *ProfileRepository) FindByUsername(username string) (*model.Profile, error) {
	return nil, nil
}
