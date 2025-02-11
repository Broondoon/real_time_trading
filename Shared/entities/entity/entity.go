package entity

import (
	"encoding/json"
	"time"
)

type EntityInterface interface {
	GetId() string
	SetId(id string)
	GetDateCreated() time.Time
	SetDateCreated(dateCreated time.Time)
	GetDateModified() time.Time
	SetDateModified(dateModified time.Time)
	EntityToParams() NewEntityParams
	EntityToJSON() ([]byte, error)
}

type Entity struct {
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	Id           string
	DateCreated  time.Time
	DateModified time.Time
	// Internal Functions should not be interacted with directly, but if you need to change functionality, set a new function to the existing function.
	// Instead, interact with the functions through the Entity Interface.
	SetIdInternal           func(id string)
	GetIdInternal           func() string
	SetDateCreatedInternal  func(dateCreated time.Time)
	GetDateCreatedInternal  func() time.Time
	SetDateModifiedInternal func(dateModified time.Time)
	GetDateModifiedInternal func() time.Time
}

func (e *Entity) GetId() string {
	return e.GetIdInternal()
}

func (e *Entity) SetId(id string) {
	e.SetIdInternal(id)
}

func (e *Entity) GetDateCreated() time.Time {
	return e.GetDateCreatedInternal()
}

func (e *Entity) SetDateCreated(dateCreated time.Time) {
	e.SetDateCreatedInternal(dateCreated)
}

func (e *Entity) GetDateModified() time.Time {
	return e.GetDateModifiedInternal()
}

func (e *Entity) SetDateModified(dateModified time.Time) {
	e.SetDateModifiedInternal(dateModified)
}

type NewEntityParams struct {
	Id           string    `json:"Id"`
	DateCreated  time.Time `json:"DateCreated"`
	DateModified time.Time `json:"DateModified"`
}

func NewEntity(params NewEntityParams) *Entity {
	e := &Entity{
		Id:           params.Id,
		DateCreated:  params.DateCreated,
		DateModified: params.DateModified,
	}
	e.SetIdInternal = func(id string) { e.Id = id }
	e.GetIdInternal = func() string { return e.Id }
	e.SetDateCreatedInternal = func(dateCreated time.Time) { e.DateCreated = dateCreated }
	e.GetDateCreatedInternal = func() time.Time { return e.DateCreated }
	e.SetDateModifiedInternal = func(dateModified time.Time) { e.DateModified = dateModified }
	e.GetDateModifiedInternal = func() time.Time { return e.DateModified }
	return e
}

func ParseEntity(jsonBytes []byte) (*Entity, error) {
	var e NewEntityParams
	if err := json.Unmarshal(jsonBytes, &e); err != nil {
		return nil, err
	}
	return NewEntity(e), nil
}

func (e *Entity) EntityToParams() NewEntityParams {
	return NewEntityParams{
		Id:           e.GetId(),
		DateCreated:  e.GetDateCreated(),
		DateModified: e.GetDateModified(),
	}
}

func (e *Entity) EntityToJSON() ([]byte, error) {
	return json.Marshal(e.EntityToParams())
}
