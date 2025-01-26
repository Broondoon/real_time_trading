package entity

import (
	"time"
)

type EntityInterface interface {
	GetId() string
	SetId(id string)
	GetDateCreated() time.Time
	SetDateCreated(dateCreated time.Time)
	GetDateModified() time.Time
	SetDateModified(dateModified time.Time)
}

// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing Internal function.
// Instead, interact with the functions through the Entity Interface.
type Entity struct {
	id                      string
	dateCreated             time.Time
	dateModified            time.Time
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
	Id           string
	DateCreated  time.Time
	DateModified time.Time
}

func NewEntity(params NewEntityParams) *Entity {
	e := &Entity{
		id:           params.Id,
		dateCreated:  params.DateCreated,
		dateModified: params.DateModified,
	}
	e.SetIdInternal = func(id string) { e.id = id }
	e.GetIdInternal = func() string { return e.id }
	e.SetDateCreatedInternal = func(dateCreated time.Time) { e.dateCreated = dateCreated }
	e.GetDateCreatedInternal = func() time.Time { return e.dateCreated }
	e.SetDateModifiedInternal = func(dateModified time.Time) { e.dateModified = dateModified }
	e.GetDateModifiedInternal = func() time.Time { return e.dateModified }
	return e
}
