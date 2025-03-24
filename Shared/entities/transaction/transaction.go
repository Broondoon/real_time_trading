package transaction

import (
	"Shared/entities/entity"
	"time"

	"github.com/google/uuid"
)

type TransactionInterface interface {
	entity.EntityInterface
	GetTimestamp() time.Time
	SetTimestamp(timestamp time.Time)
	GetUserID() *uuid.UUID
	GetUserIDString() string
	SetUserID(userID *uuid.UUID)
}

type Transaction struct {
	entity.Entity `json:"entity" gorm:"embedded"`
	Timestamp     time.Time  `json:"time_stamp" gorm:"type:timestamp;autoUpdateTime:milli"`
	UserID        *uuid.UUID `json:"user_id" gorm:"type:uuid;column:user_id;not null"`
}

func (st *Transaction) GetTimestamp() time.Time {
	return st.Timestamp
}

func (st *Transaction) SetTimestamp(timestamp time.Time) {
	st.Timestamp = timestamp
	*st.GetUpdates() = append(*st.Updates, &entity.EntityUpdateData{
		ID:       st.GetId(),
		Field:    "Timestamp",
		NewValue: func() *string { s := timestamp.Format(time.RFC3339); return &s }(),
	})
}

func (st *Transaction) GetUserID() *uuid.UUID {
	return st.UserID
}

func (st *Transaction) GetUserIDString() string {
	if st.UserID == nil {
		return ""
	}
	return st.UserID.String()
}

func (st *Transaction) SetUserID(userID *uuid.UUID) {
	st.UserID = userID
}

type NewTransactionParams struct {
	*entity.NewEntityParams `json:"entity"`
	TimeStamp               time.Time  `json:"time_stamp"`
	UserID                  *uuid.UUID `json:"user_id"`
}

func New(params *NewTransactionParams) *Transaction {
	if params.NewEntityParams == nil {
		params.NewEntityParams = &entity.NewEntityParams{}
	}
	e := entity.NewEntity(params.NewEntityParams)
	return &Transaction{
		Entity:    *e,
		Timestamp: params.TimeStamp,
		UserID:    params.UserID,
	}
}

func (st *Transaction) ToParams() NewTransactionParams {
	eparams := st.EntityToParams()
	return NewTransactionParams{
		NewEntityParams: &eparams,
		TimeStamp:       st.GetTimestamp(),
		UserID:          st.GetUserID(),
	}
}
