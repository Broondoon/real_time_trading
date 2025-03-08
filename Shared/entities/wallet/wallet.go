package wallet

import (
	"Shared/entities/entity"
	"Shared/entities/user"
	"encoding/json"
)

type WalletInterface interface {
	GetUserID() string
	SetUserID(userID string)
	GetBalance() float64
	SetBalance(balance float64)
	ToParams() NewWalletParams
	entity.EntityInterface
}

type Wallet struct {
	UserID  string  `json:"user_id" gorm:"not null"`
	Balance float64 `json:"balance" gorm:"not null"`
	// The internal function fields have been commented out,
	// and the getters/setters below operate directly on the properties.
	/*
		GetUserIDInternal  func() string         `gorm:"-"`
		SetUserIDInternal  func(userID string)   `gorm:"-"`
		GetBalanceInternal func() float64        `gorm:"-"`
		SetBalanceInternal func(balance float64) `gorm:"-"`
	*/
	entity.Entity `json:"Entity" gorm:"embedded"`
}

func (w *Wallet) GetBalance() float64 {
	return w.Balance
}

func (w *Wallet) SetBalance(balance float64) {
	w.Balance = balance
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

type FakeWallet struct {
	entity.FakeEntity
	UserID  string `json:"UserId"`
	Balance float64
}

func (fw *FakeWallet) GetUserID() string          { return fw.UserID }
func (fw *FakeWallet) SetUserID(userID string)    { fw.UserID = userID }
func (fw *FakeWallet) GetBalance() float64        { return fw.Balance }
func (fw *FakeWallet) SetBalance(balance float64) { fw.Balance = balance }
func (fw *FakeWallet) ToParams() NewWalletParams  { return NewWalletParams{} }
func (fw *FakeWallet) ToJSON() ([]byte, error)    { return []byte{}, nil }

func (w *Wallet) SetDefaults() {
	if w.Balance == 0 {
		w.Balance = 0.00
	}
}
