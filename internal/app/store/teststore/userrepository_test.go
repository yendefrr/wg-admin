package teststore_test

import (
	"github.com/stretchr/testify/assert"
	"go/wg-admin/internal/app/model"
	"go/wg-admin/internal/app/store"
	"go/wg-admin/internal/app/store/teststore"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestUserRepository_Create(t *testing.T) {
	s := teststore.New()
	u := model.TestUser()

	err := s.User().Create(u)
	assert.NoError(t, err)
	assert.NotNil(t, model.TestUser())
}

func TestUserRepository_Find(t *testing.T) {
	s := teststore.New()

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

func TestUserRepository_FindByEmail(t *testing.T) {
	s := teststore.New()

	email := "user@yendefr.xyz"
	_, err := s.User().FindByEmail(email)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	u := model.TestUser()
	u.Email = email
	err = s.User().Create(u)
	if err != nil {
		t.Fatal(err)
	}

	u, err = s.User().FindByEmail(email)
	assert.NoError(t, err)
	assert.NotNil(t, u)
}
