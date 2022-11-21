package sqlstore

import (
	"fmt"
	"go/wg-admin/internal/app/model"
)

type ProfileRepository struct {
	store *Store
}

func (r *ProfileRepository) Create(p *model.Profile) error {

	r.store.db.QueryRow(fmt.Sprintf(
		"INSERT INTO `profiles` (`username`, `type`, `path`, `publickey`, `privatekey`) VALUES ('%s', '%s', '%s', '%s', '%s')",
		p.Username, p.Type, p.Path, p.Publickey, p.Privatekey))

	return r.store.db.QueryRow(fmt.Sprintf("SELECT `id` FROM `profiles` WHERE `username` = '%s'", p.Username)).Scan(&p.ID)
}

func (r *ProfileRepository) Find(id int) (*model.Profile, error) {
	return nil, nil
}

func (r *ProfileRepository) FindByUsername(username string) (*model.Profile, error) {
	return nil, nil
}
