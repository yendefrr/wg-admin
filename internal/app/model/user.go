package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Username, validation.Required, validation.Length(5, 32)),
	)
}

func (u *User) BeforeCreate() error {
	return nil
}
