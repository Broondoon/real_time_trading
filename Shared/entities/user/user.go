package user

import "Shared/entities/entity"

type UserInterface interface {
	GetUsername() string
	SetUsername(username string)
	GetPassword() string
	SetPassword(password string)
	entity.EntityInterface
}

// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
// Instead, interact with the functions through the User Interface.
type User struct {
	username            string
	password            string
	GetUsernameInternal func() string
	SetUsernameInternal func(username string)
	GetPasswordInternal func() string
	SetPasswordInternal func(password string)
	entity.EntityInterface
}

func (u *User) GetUsername() string {
	return u.GetUsernameInternal()
}

func (u *User) SetUsername(username string) {
	u.SetUsernameInternal(username)
}

func (u *User) GetPassword() string {
	return u.GetPasswordInternal()
}

func (u *User) SetPassword(password string) {
	u.SetPasswordInternal(password)
}

type NewUserParams struct {
	NewEntityParams entity.NewEntityParams
	Username        string
	Password        string
}

func NewUser(params NewUserParams) *User {
	e := entity.NewEntity(params.NewEntityParams)
	u := &User{
		EntityInterface: e,
		username:        params.Username,
		password:        params.Password,
	}
	u.SetUsernameInternal = func(username string) { u.username = username }
	u.GetUsernameInternal = func() string { return u.username }
	u.SetPasswordInternal = func(password string) { u.password = password }
	u.GetPasswordInternal = func() string { return u.password }
	return u
}
