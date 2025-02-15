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
	UserID  string  `json:"UserId" gorm:"not null"`
	Balance float64 `json:"Balance" gorm:"not null"`
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the wallet Interface.
	GetUserIDInternal          func() string         `gorm:"-"`
	SetUserIDInternal          func(userID string)   `gorm:"-"`
	GetBalanceInternal         func() float64        `gorm:"-"`
	SetBalanceInternal         func(balance float64) `gorm:"-"`
	entity.BaseEntityInterface `gorm:"embedded"`
}

func (w *Wallet) GetBalance() float64 {
	return w.GetBalanceInternal()
}

func (w *Wallet) SetBalance(balance float64) {
	w.SetBalanceInternal(balance)
}

func (w *Wallet) GetUserID() string {
	return w.GetUserIDInternal()
}

func (w *Wallet) SetUserID(userID string) {
	w.SetUserIDInternal(userID)
}

type NewWalletParams struct {
	entity.NewEntityParams
	UserID  string             `json:"UserId" gorm:"not null"`
	Balance float64            `json:"Balance" gorm:"not null"`
	User    user.UserInterface // use this or UserId
}

func New(params NewWalletParams) *Wallet {
	e := entity.NewEntity(params.NewEntityParams)
	var UserID string
	if params.User != nil {
		UserID = params.User.GetId()
	} else {
		UserID = params.UserID
	}

	wb := &Wallet{
		UserID:              UserID,
		Balance:             params.Balance,
		BaseEntityInterface: e,
	}
	wb.SetDefaults()
	return wb
}

func (w *Wallet) SetDefaults() {
	w.GetUserIDInternal = func() string { return w.UserID }
	w.SetUserIDInternal = func(userID string) { w.UserID = userID }
	w.GetBalanceInternal = func() float64 { return w.Balance }
	w.SetBalanceInternal = func(balance float64) { w.Balance = balance }
}

func Parse(jsonBytes []byte) (*Wallet, error) {
	var w NewWalletParams
	if err := json.Unmarshal(jsonBytes, &w); err != nil {
		return nil, err
	}
	return New(w), nil
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
