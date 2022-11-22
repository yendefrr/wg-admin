package store

import "go/wg-admin/internal/app/model"

type UserRepository interface {
	Create(user *model.User) error
	Find(int) (*model.User, error)
	FindByEmail(string) (*model.User, error)
}

type ProfileRepository interface {
	Create(profile *model.Profile) error
	GetAll() ([]model.Profile, error)
	Find(int) (*model.Profile, error)
	FindByUsername(string) (*model.Profile, error)
}
