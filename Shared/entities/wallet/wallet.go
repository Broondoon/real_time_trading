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
	WalletToParams() NewWalletParams
	WalletToJSON() ([]byte, error)
	entity.EntityInterface
}

type Wallet struct {
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	UserID  string
	Balance float64
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the wallet Interface.
	GetUserIDInternal  func() string
	SetUserIDInternal  func(userID string)
	GetBalanceInternal func() float64
	SetBalanceInternal func(balance float64)
	entity.EntityInterface
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
	User    user.UserInterface // use this or UserId
	UserId  string             `json:"UserId"` // use this or User
	Balance float64            `json:"Balance"`
}

func NewWallet(params NewWalletParams) *Wallet {
	e := entity.NewEntity(params.NewEntityParams)
	var UserID string
	if params.User != nil {
		UserID = params.User.GetId()
	} else {
		UserID = params.UserId
	}

	wb := &Wallet{
		UserID:          UserID,
		Balance:         params.Balance,
		EntityInterface: e,
	}
	wb.GetUserIDInternal = func() string { return wb.UserID }
	wb.SetUserIDInternal = func(userID string) { wb.UserID = userID }
	wb.GetBalanceInternal = func() float64 { return wb.Balance }
	wb.SetBalanceInternal = func(balance float64) { wb.Balance = balance }
	return wb
}

func ParseWallet(jsonBytes []byte) (*Wallet, error) {
	var w NewWalletParams
	if err := json.Unmarshal(jsonBytes, &w); err != nil {
		return nil, err
	}
	return NewWallet(w), nil
}

func (w *Wallet) WalletToParams() NewWalletParams {
	return NewWalletParams{
		NewEntityParams: w.EntityToParams(),
		User:            nil,
		UserId:          w.GetUserID(),
		Balance:         w.GetBalance(),
	}
}

func (w *Wallet) WalletToJSON() ([]byte, error) {
	return json.Marshal(w.WalletToParams())
}

type FakeWallet struct {
	entity.FakeEntity
	UserID  string `json:"UserId"`
	Balance float64
}

func (fw *FakeWallet) GetUserID() string               { return fw.UserID }
func (fw *FakeWallet) SetUserID(userID string)         { fw.UserID = userID }
func (fw *FakeWallet) GetBalance() float64             { return fw.Balance }
func (fw *FakeWallet) SetBalance(balance float64)      { fw.Balance = balance }
func (fw *FakeWallet) WalletToParams() NewWalletParams { return NewWalletParams{} }
func (fw *FakeWallet) WalletToJSON() ([]byte, error)   { return []byte{}, nil }
