package model

func TestUser() *User {
	return &User{
		Email:    "user@yendefr.xyz",
		Password: "password",
	}
}
