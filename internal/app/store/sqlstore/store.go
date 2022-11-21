package sqlstore

import (
	"database/sql"
	"go/wg-admin/internal/app/store"

	_ "github.com/lib/pq"
)

type Store struct {
	db                *sql.DB
	userRepository    *UserRepository
	profileRepository *ProfileRepository
}

func New(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
	}

	return s.userRepository
}

func (s *Store) Profile() store.ProfileRepository {
	s.profileRepository = &ProfileRepository{
		store: s,
	}

	return s.profileRepository
}
