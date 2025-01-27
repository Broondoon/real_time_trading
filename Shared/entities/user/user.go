package user

import "Shared/entities/entity"

type UserInterface interface {
	GetName() string
	SetName(name string)
	GetUsername() string
	SetUsername(username string)
	GetPassword() string
	SetPassword(password string)
	entity.EntityInterface
}

type User struct {
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	name     string
	username string
	password string
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the User Interface.
	GetNameInternal     func() string
	SetNameInternal     func(name string)
	GetUsernameInternal func() string
	SetUsernameInternal func(username string)
	GetPasswordInternal func() string
	SetPasswordInternal func(password string)
	entity.EntityInterface
}

func (u *User) GetName() string {
	return u.GetNameInternal()
}

func (u *User) SetName(name string) {
	u.SetNameInternal(name)
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
	Name            string
	Username        string
	Password        string
}

func NewUser(params NewUserParams) *User {
	e := entity.NewEntity(params.NewEntityParams)
	u := &User{
		EntityInterface: e,
		name:            params.Name,
		username:        params.Username,
		password:        params.Password,
	}
	u.GetNameInternal = func() string { return u.name }
	u.SetNameInternal = func(name string) { u.name = name }
	u.SetUsernameInternal = func(username string) { u.username = username }
	u.GetUsernameInternal = func() string { return u.username }
	u.SetPasswordInternal = func(password string) { u.password = password }
	u.GetPasswordInternal = func() string { return u.password }
	return u
}
