package wallet

import (
	"Shared/entities/entity"
	"Shared/entities/user"
	"encoding/json"
	"strconv"
)

type WalletInterface interface {
	GetUserID() string
	SetUserID(userID string)
	GetBalance() float64
	UpdateBalance(balanceToAdd float64)
	ToParams() NewWalletParams
	entity.EntityInterface
}

type Wallet struct {
	UserID        string  `json:"user_id" gorm:"not null"`
	Balance       float64 `json:"balance" gorm:"not null"`
	entity.Entity `json:"Entity" gorm:"embedded"`
}

func (w *Wallet) GetBalance() float64 {
	return w.Balance
}

func (w *Wallet) UpdateBalance(balanceToAdd float64) {
	w.Balance += balanceToAdd
	*w.Updates = append(*w.Updates, &entity.EntityUpdateData{
		ID:         w.GetId(),
		Field:      "Balance",
		AlterValue: func() *string { s := strconv.FormatFloat(balanceToAdd, 'f', -1, 64); return &s }(),
	})
}

func (w *Wallet) GetUserID() string {
	return w.UserID
}

func (w *Wallet) SetUserID(userID string) {
	w.UserID = userID
}

type NewWalletParams struct {
	entity.NewEntityParams `json:"Entity"`
	UserID                 string             `json:"user_id" gorm:"not null"`
	Balance                float64            `json:"balance" gorm:"not null"`
	User                   user.UserInterface // use this or UserId
}

func New(params NewWalletParams) *Wallet {
	e := entity.NewEntity(params.NewEntityParams)
	var UserID string
	if params.User != nil {
		UserID = params.User.GetId()
	} else {
		UserID = params.UserID
	}
	e.SetId(UserID)

	wb := &Wallet{
		UserID:  UserID,
		Balance: params.Balance,
		Entity:  *e,
	}
	// Using direct field access; no need to set internal function defaults.
	return wb
}

func Parse(jsonBytes []byte) (*Wallet, error) {
	var w NewWalletParams
	if err := json.Unmarshal(jsonBytes, &w); err != nil {
		return nil, err
	}
	return New(w), nil
}

func ParseList(jsonBytes []byte) (*[]*Wallet, error) {
	var so []NewWalletParams
	if err := json.Unmarshal(jsonBytes, &so); err != nil {
		return nil, err
	}
	soList := make([]*Wallet, len(so))
	for i, s := range so {
		soList[i] = New(s)
	}
	return &soList, nil
}

func (w *Wallet) ToParams() NewWalletParams {
	return NewWalletParams{
		NewEntityParams: w.EntityToParams(),
		User:            nil,
		UserID:          w.GetUserID(),
		Balance:         w.GetBalance(),
	}
}

func (w *Wallet) ToJSON() ([]byte, error) {
	return json.Marshal(w.ToParams())
}
