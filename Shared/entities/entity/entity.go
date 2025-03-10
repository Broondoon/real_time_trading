package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type BaseEntityInterface interface {
	GetId() *uuid.UUID
	GetIdString() string
	SetId(id *uuid.UUID)
	GetDateCreated() time.Time
	SetDateCreated(dateCreated time.Time)
	GetDateModified() time.Time
	SetDateModified(dateModified time.Time)
	GetUpdates() *[]*EntityUpdateData
	EntityToParams() NewEntityParams
	EntityToJSON() ([]byte, error)
	GenUniquePairing() *uuid.UUID
	GetUniquePairing() *uuid.UUID
	SetUnqiuePairing(uniquePairing *uuid.UUID)
}

type EntityInterface interface {
	ToJSON() ([]byte, error)
	BaseEntityInterface
}

type Entity struct {
	ID            *uuid.UUID           `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"` // gorm:"primaryKey" is used to set the primary key in the database.
	DateCreated   time.Time            `json:"DateCreated" gorm:"autoCreateTime:milli"`                   // gorm:"autoCreateTime:milli" is used to set the time the entity was created in the database.
	DateModified  time.Time            `json:"DateModified" gorm:"autoUpdateTime:milli"`                  // gorm:"autoUpdateTime:milli" is used to set the time the entity was last modified in the database.
	Updates       *[]*EntityUpdateData `json:"-" gorm:"-"`
	UniquePairing *uuid.UUID           `json:"temp" gorm:"-"`
}

type EntityUpdateData struct {
	ID         *uuid.UUID
	Field      string
	NewValue   *string // New value to set the field; set as nil to not update.
	AlterValue *string // Value to add to the existing value NOT the new value; set as nil to not update.
}

func (e *Entity) GenUniquePairing() *uuid.UUID {
	temp := uuid.New()
	e.UniquePairing = &temp
	return e.UniquePairing
}

func (e *Entity) SetUnqiuePairing(uniquePairing *uuid.UUID) {
	e.UniquePairing = uniquePairing
}

func (e *Entity) GetUniquePairing() *uuid.UUID {
	if e.UniquePairing == nil {
		e.UniquePairing = e.GenUniquePairing()
	}
	return e.UniquePairing
}

func (e *Entity) GetId() *uuid.UUID {
	return e.ID
}

func (e *Entity) GetIdString() string {
	if e.ID == nil {
		return ""
	}
	return e.ID.String()
}

func (e *Entity) SetId(id *uuid.UUID) {
	e.ID = id
}

func (e *Entity) GetDateCreated() time.Time {
	return e.DateCreated
}

func (e *Entity) SetDateCreated(dateCreated time.Time) {
	e.DateCreated = dateCreated
}

func (e *Entity) GetDateModified() time.Time {
	return e.DateModified
}

func (e *Entity) SetDateModified(dateModified time.Time) {
	e.DateModified = dateModified
	*e.GetUpdates() = append(*e.Updates, &EntityUpdateData{
		ID:       e.GetId(),
		Field:    "DateModified",
		NewValue: func() *string { s := dateModified.Format(time.RFC3339); return &s }(),
	})
}

func (e *Entity) GetUpdates() *[]*EntityUpdateData {
	if e.Updates == nil {
		tmp := make([]*EntityUpdateData, 0)
		e.Updates = &tmp
	}
	return e.Updates
}

type NewEntityParams struct {
	ID            *uuid.UUID `json:"ID"`
	DateCreated   time.Time  `json:"DateCreated"`
	DateModified  time.Time  `json:"DateModified"`
	UniquePairing *uuid.UUID `json:"temp"` //Do not set. This is for pairing with bulk stuff.
}

func NewEntity(params NewEntityParams) *Entity {
	tmp := make([]*EntityUpdateData, 0)

	e := &Entity{
		ID:            params.ID,
		DateCreated:   params.DateCreated,
		DateModified:  params.DateModified,
		Updates:       &tmp,
		UniquePairing: params.UniquePairing,
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
		ID:            e.GetId(),
		DateCreated:   e.GetDateCreated(),
		DateModified:  e.GetDateModified(),
		UniquePairing: e.GetUniquePairing(),
	}
}

func (e *Entity) EntityToJSON() ([]byte, error) {
	return json.Marshal(e.EntityToParams())
}
