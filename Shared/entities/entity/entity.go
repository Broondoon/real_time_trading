package entity

import (
	"encoding/json"
	"time"
)

type BaseEntityInterface interface {
	GetId() string
	SetId(id string)
	GetDateCreated() time.Time
	SetDateCreated(dateCreated time.Time)
	GetDateModified() time.Time
	SetDateModified(dateModified time.Time)
	EntityToParams() NewEntityParams
	EntityToJSON() ([]byte, error)
}

type EntityInterface interface {
	ToJSON() ([]byte, error)
	BaseEntityInterface
}

type Entity struct {
	ID           string    `json:"ID" gorm:"primaryKey"`                     // gorm:"primaryKey" is used to set the primary key in the database.
	DateCreated  time.Time `json:"DateCreated" gorm:"autoCreateTime:milli"`  // gorm:"autoCreateTime:milli" is used to set the time the entity was created in the database.
	DateModified time.Time `json:"DateModified" gorm:"autoUpdateTime:milli"` // gorm:"autoUpdateTime:milli" is used to set the time the entity was last modified in the database.
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	// Internal Functions should not be interacted with directly, but if you need to change functionality, set a new function to the existing function.
	// Instead, interact with the functions through the Entity Interface.
	// SetIdInternal           func(id string)              `gorm:"-"`
	// GetIdInternal           func() string                `gorm:"-"`
	// SetDateCreatedInternal  func(dateCreated time.Time)  `gorm:"-"`
	// GetDateCreatedInternal  func() time.Time             `gorm:"-"`
	// SetDateModifiedInternal func(dateModified time.Time) `gorm:"-"`
	// GetDateModifiedInternal func() time.Time             `gorm:"-"`
}

func (e *Entity) GetId() string {
	return e.ID //e.GetIdInternal()
}

func (e *Entity) SetId(id string) {
	e.ID = id //e.SetIdInternal(id)
}

func (e *Entity) GetDateCreated() time.Time {
	return e.DateCreated //e.GetDateCreatedInternal()
}

func (e *Entity) SetDateCreated(dateCreated time.Time) {
	e.DateCreated = dateCreated //e.SetDateCreatedInternal(dateCreated)
}

func (e *Entity) GetDateModified() time.Time {
	return e.DateModified //e.GetDateModifiedInternal()
}

func (e *Entity) SetDateModified(dateModified time.Time) {
	e.DateModified = dateModified //e.SetDateModifiedInternal(dateModified)
}

type NewEntityParams struct {
	ID           string    `json:"ID"`
	DateCreated  time.Time `json:"DateCreated"`
	DateModified time.Time `json:"DateModified"`
}

func NewEntity(params NewEntityParams) *Entity {
	e := &Entity{
		ID:           params.ID,
		DateCreated:  params.DateCreated,
		DateModified: params.DateModified,
	}
	return e
}

func ParseEntity(jsonBytes []byte) (BaseEntityInterface, error) {
	var e NewEntityParams
	if err := json.Unmarshal(jsonBytes, &e); err != nil {
		return nil, err
	}
	return NewEntity(e), nil
}

func (e *Entity) EntityToParams() NewEntityParams {
	return NewEntityParams{
		ID:           e.GetId(),
		DateCreated:  e.GetDateCreated(),
		DateModified: e.GetDateModified(),
	}
}

func (e *Entity) EntityToJSON() ([]byte, error) {
	return json.Marshal(e.EntityToParams())
}

type FakeEntity struct {
	Id           string    `json:"ID"`
	DateCreated  time.Time `json:"dateCreated"`
	DateModified time.Time `json:"dateModified"`
}

func (fe *FakeEntity) GetId() string                          { return fe.Id }
func (fe *FakeEntity) SetId(id string)                        { fe.Id = id }
func (fe *FakeEntity) GetDateCreated() time.Time              { return fe.DateCreated }
func (fe *FakeEntity) SetDateCreated(dateCreated time.Time)   { fe.DateCreated = dateCreated }
func (fe *FakeEntity) GetDateModified() time.Time             { return fe.DateModified }
func (fe *FakeEntity) SetDateModified(dateModified time.Time) { fe.DateModified = dateModified }
func (fe *FakeEntity) EntityToParams() NewEntityParams        { return NewEntityParams{} }
func (fe *FakeEntity) EntityToJSON() ([]byte, error)          { return []byte{}, nil }
