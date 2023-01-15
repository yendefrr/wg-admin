package store

import "go/wg-admin/internal/app/model"

type UserRepository interface {
	Create(user *model.User) error
	GetAll() ([]model.User, error)
	Find(int) (*model.User, error)
	FindByUsername(string) (*model.User, error)
}

type ProfileRepository interface {
	Create(profile *model.Profile) error
	Update(profile *model.Profile) error
	Delete(int) error
	GetAll(bool) ([]model.Profile, error)
	Find(int) (*model.Profile, error)
	FindByUsername(string) (*model.Profile, error)
}
