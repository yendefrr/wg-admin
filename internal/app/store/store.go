package store

type Store interface {
	User() UserRepository
	Profile() ProfileRepository
}
