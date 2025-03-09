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
	Name          string `json:"name" gorm:"not null"`
	Username      string `json:"user_name" gorm:"unique not null"`
	Password      string `json:"password" gorm:"not null"`
	entity.Entity `json:"Entity" gorm:"embedded"`
}

func (u *User) GetName() string {
	return u.Name
}

func (u *User) SetName(name string) {
	u.Name = name
	*u.GetUpdates() = append(*u.Updates, &entity.EntityUpdateData{
		ID:       u.GetId(),
		Field:    "Name",
		NewValue: &name,
	})
}

func (u *User) GetUsername() string {
	return u.Username
}

func (u *User) SetUsername(username string) {
	u.Username = username
	*u.GetUpdates() = append(*u.Updates, &entity.EntityUpdateData{
		ID:       u.GetId(),
		Field:    "Username",
		NewValue: &username,
	})
}

func (u *User) GetPassword() string {
	return u.Password
}

func (u *User) SetPassword(password string) {
	u.Password = password
	*u.GetUpdates() = append(*u.Updates, &entity.EntityUpdateData{
		ID:       u.GetId(),
		Field:    "Password",
		NewValue: &password,
	})
}

type NewUserParams struct {
	entity.NewEntityParams `json:"Entity"`
	Name                   string `json:"name"`
	Username               string `json:"user_name"`
	Password               string `json:"password"`
}

func New(params NewUserParams) *User {
	e := entity.NewEntity(params.NewEntityParams)
	u := &User{
		Entity:   *e,
		Name:     params.Name,
		Username: params.Username,
		Password: params.Password,
	}
	return u
}

func Parse(jsonBytes []byte) (*User, error) {
	var u NewUserParams
	if err := json.Unmarshal(jsonBytes, &u); err != nil {
		return nil, err
	}
	return New(u), nil
}

func ParseList(jsonBytes []byte) (*[]*User, error) {
	var so []NewUserParams
	if err := json.Unmarshal(jsonBytes, &so); err != nil {
		return nil, err
	}
	soList := make([]*User, len(so))
	for i, s := range so {
		soList[i] = New(s)
	}
	return &soList, nil
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
