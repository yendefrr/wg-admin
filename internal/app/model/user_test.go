package model_test

import (
	"github.com/stretchr/testify/assert"
	"go/wg-admin/internal/app/model"
	"testing"
)

func TestUser_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		u       func() *model.User
		isValid bool
	}{
		{
			name: "valid",
			u: func() *model.User {
				return model.TestUser()
			},
			isValid: true,
		},
		{
			name: "empty email",
			u: func() *model.User {
				u := model.TestUser()
				u.Email = ""
				return u
			},
			isValid: false,
		},
		{
			name: "empty password",
			u: func() *model.User {
				u := model.TestUser()
				u.Password = ""
				return u
			},
			isValid: false,
		},
		{
			name: "too short password",
			u: func() *model.User {
				u := model.TestUser()
				u.Password = "1234567"
				return u
			},
			isValid: false,
		},
		{
			name: "has hash, without password",
			u: func() *model.User {
				u := model.TestUser()
				u.Password = ""
				u.PasswordHash = "hash"
				return u
			},
			isValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				assert.NoError(t, tc.u().Validate())
			} else {
				assert.Error(t, tc.u().Validate())
			}
		})
	}
}

func TestUser_BeforeCreate(t *testing.T) {
	u := model.TestUser()
	assert.NoError(t, u.BeforeCreate())
	assert.NotEmpty(t, u.PasswordHash)
}
