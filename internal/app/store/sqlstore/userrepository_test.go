package sqlstore_test

import (
	"github.com/stretchr/testify/assert"
	"go/wg-admin/internal/app/model"
	"go/wg-admin/internal/app/store"
	"go/wg-admin/internal/app/store/sqlstore"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestUserRepository_Create(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("users")

	s := sqlstore.New(db)
	u := model.TestUser()

	err := s.User().Create(u)
	assert.NoError(t, err)
	assert.NotNil(t, model.TestUser())
}

func TestUserRepository_Find(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("users")

	s := sqlstore.New(db)

	_, err := s.User().Find(9999)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	u := model.TestUser()
	err = s.User().Create(u)
	if err != nil {
		t.Fatal(err)
	}

	u, err = s.User().Find(u.ID)
	assert.NoError(t, err)
	assert.NotNil(t, u)
}

func TestUserRepository_FindByUsername(t *testing.T) {
	db, teardown := sqlstore.TestDB(t, databaseURL)
	defer teardown("users")

	s := sqlstore.New(db)

	username := "testusername"
	_, err := s.User().FindByUsername(username)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	u := model.TestUser()
	u.Username = username
	err = s.User().Create(u)
	if err != nil {
		t.Fatal(err)
	}

	u, err = s.User().FindByUsername(username)
	assert.NoError(t, err)
	assert.NotNil(t, u)
}
