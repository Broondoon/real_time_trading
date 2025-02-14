package user

import (
	"Shared/entities/entity"
	"encoding/json"
)

type UserInterface interface {
	GetName() string
	SetName(name string)
	GetUsername() string
	SetUsername(username string)
	GetPassword() string
	SetPassword(password string)
	ToParams() NewUserParams
	entity.EntityInterface
}

type User struct {
	Name     string `json:"Name" gorm:"not null"`
	Username string `json:"Username" gorm:"unique not null"`
	Password string `json:"Password" gorm:"not null"` // If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the User Interface.
	GetNameInternal     func() string         `gorm:"-"`
	SetNameInternal     func(name string)     `gorm:"-"`
	GetUsernameInternal func() string         `gorm:"-"`
	SetUsernameInternal func(username string) `gorm:"-"`
	GetPasswordInternal func() string         `gorm:"-"`
	SetPasswordInternal func(password string) `gorm:"-"`
	entity.BaseEntityInterface
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
	entity.NewEntityParams
	Name     string `json:"Name"`
	Username string `json:"Username"`
	Password string `json:"Password"`
}

func New(params NewUserParams) *User {
	e := entity.NewEntity(params.NewEntityParams)
	u := &User{
		BaseEntityInterface: e,
		Name:                params.Name,
		Username:            params.Username,
		Password:            params.Password,
	}

	u.GetNameInternal = func() string { return u.Name }
	u.SetNameInternal = func(name string) { u.Name = name }
	u.SetUsernameInternal = func(username string) { u.Username = username }
	u.GetUsernameInternal = func() string { return u.Username }
	u.SetPasswordInternal = func(password string) { u.Password = password }
	u.GetPasswordInternal = func() string { return u.Password }
	return u
}

func Parse(jsonBytes []byte) (*User, error) {
	var u NewUserParams
	if err := json.Unmarshal(jsonBytes, &u); err != nil {
		return nil, err
	}
	return New(u), nil
}

func (u *User) ToParams() NewUserParams {
	return NewUserParams{
		NewEntityParams: u.EntityToParams(),
		Name:            u.GetName(),
		Username:        u.GetUsername(),
		Password:        u.GetPassword(),
	}
}

func (u *User) ToJSON() ([]byte, error) {
	return json.Marshal(u.ToParams())
}

type FakeUser struct {
	entity.FakeEntity
	Name     string `json:"Name"`
	Username string `json:"Username"`
	Password string `json:"Password"`
}

func (fu *FakeUser) GetName() string             { return fu.Name }
func (fu *FakeUser) SetName(name string)         { fu.Name = name }
func (fu *FakeUser) GetUsername() string         { return fu.Username }
func (fu *FakeUser) SetUsername(username string) { fu.Username = username }
func (fu *FakeUser) GetPassword() string         { return fu.Password }
func (fu *FakeUser) SetPassword(password string) { fu.Password = password }
func (fu *FakeUser) ToParams() NewUserParams     { return NewUserParams{} }
func (fu *FakeUser) ToJSON() ([]byte, error)     { return []byte{}, nil }
