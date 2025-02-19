package transaction

import (
	"Shared/entities/entity"
	"Shared/entities/wallet"
	"encoding/json"
)

type WalletTransactionInterface interface {
	GetWalletID() string
	SetWalletID(walletID string)
	GetStockTransactionID() string
	SetStockTransactionID(stockTransactionID string)
	GetIsDebit() bool
	SetIsDebit(isDebit bool)
	GetAmount() float64
	SetAmount(amount float64)
	ToParams() NewWalletTransactionParams
	ToJSON() ([]byte, error)
	entity.EntityInterface
}

type WalletTransaction struct {
	WalletID           string  `json:"WalletID" gorm:"not null"`
	StockTransactionID string  `json:"StockTransactionID" gorm:"not null"`
	IsDebit            bool    `json:"IsDebit" gorm:"not null"`
	Amount             float64 `json:"Amount" gorm:"not null"`
	// Internal functions have been commented out.
	// GetWalletIDInternal           func() string                   `gorm:"-"`
	// SetWalletIDInternal           func(walletID string)           `gorm:"-"`
	// GetStockTransactionIDInternal func() string                   `gorm:"-"`
	// SetStockTransactionIDInternal func(stockTransactionID string) `gorm:"-"`
	// GetIsDebitInternal            func() bool                     `gorm:"-"`
	// SetIsDebitInternal            func(isDebit bool)              `gorm:"-"`
	// GetAmountInternal             func() float64                  `gorm:"-"`
	// SetAmountInternal             func(amount float64)            `gorm:"-"`
	entity.Entity `gorm:"embedded"`
}

func (wt *WalletTransaction) GetWalletID() string {
	return wt.WalletID
}

func (wt *WalletTransaction) SetWalletID(walletID string) {
	wt.WalletID = walletID
}

func (wt *WalletTransaction) GetStockTransactionID() string {
	return wt.StockTransactionID
}

func (wt *WalletTransaction) SetStockTransactionID(stockTransactionID string) {
	wt.StockTransactionID = stockTransactionID
}

func (wt *WalletTransaction) GetIsDebit() bool {
	return wt.IsDebit
}

func (wt *WalletTransaction) SetIsDebit(isDebit bool) {
	wt.IsDebit = isDebit
}

func (wt *WalletTransaction) GetAmount() float64 {
	return wt.Amount
}

func (wt *WalletTransaction) SetAmount(amount float64) {
	wt.Amount = amount
}

type NewWalletTransactionParams struct {
	entity.NewEntityParams
	WalletID           string  `json:"WalletID" gorm:"not null"`
	StockTransactionID string  `json:"StockTransactionID" gorm:"not null"`
	IsDebit            bool    `json:"IsDebit" gorm:"not null"`
	Amount             float64 `json:"Amount" gorm:"not null"`
	Wallet             wallet.WalletInterface
	StockTransaction   StockTransactionInterface
}

func NewWalletTransaction(params NewWalletTransactionParams) *WalletTransaction {
	e := entity.NewEntity(params.NewEntityParams)
	wt := &WalletTransaction{
		Entity:  *e,
		IsDebit: params.IsDebit,
		Amount:  params.Amount,
	}
	if params.Wallet != nil {
		wt.WalletID = params.Wallet.GetId()
	} else {
		wt.WalletID = params.WalletID
	}

	if params.StockTransaction != nil {
		wt.StockTransactionID = params.StockTransaction.GetId()
	} else {
		wt.StockTransactionID = params.StockTransactionID
	}

	wt.SetDefaults()

	return wt
}

func (wt *WalletTransaction) SetDefaults() {
	// Internal function setters and getters were removed,
	// so this function is kept for compatibility or future use.
	// It is now empty.
}

func ParseWalletTransaction(jsonBytes []byte) (*WalletTransaction, error) {
	var wt NewWalletTransactionParams
	if err := json.Unmarshal(jsonBytes, &wt); err != nil {
		return nil, err
	}
	return NewWalletTransaction(wt), nil
}

func (wt *WalletTransaction) ToParams() NewWalletTransactionParams {
	return NewWalletTransactionParams{
		NewEntityParams:    wt.EntityToParams(),
		WalletID:           wt.GetWalletID(),
		StockTransactionID: wt.GetStockTransactionID(),
		IsDebit:            wt.GetIsDebit(),
		Amount:             wt.GetAmount(),
	}
}

type FakeWalletTransaction struct {
	entity.FakeEntity
	WalletID           string  `json:"walletID"`
	StockTransactionID string  `json:"stockTransactionID"`
	IsDebit            bool    `json:"isDebit"`
	Amount             float64 `json:"amount"`
}

func (fwt *FakeWalletTransaction) GetWalletID() string           { return fwt.WalletID }
func (fwt *FakeWalletTransaction) SetWalletID(walletID string)   { fwt.WalletID = walletID }
func (fwt *FakeWalletTransaction) GetStockTransactionID() string { return fwt.StockTransactionID }
func (fwt *FakeWalletTransaction) SetStockTransactionID(stockTransactionID string) {
	fwt.StockTransactionID = stockTransactionID
}
func (fwt *FakeWalletTransaction) GetIsDebit() bool         { return fwt.IsDebit }
func (fwt *FakeWalletTransaction) SetIsDebit(isDebit bool)  { fwt.IsDebit = isDebit }
func (fwt *FakeWalletTransaction) GetAmount() float64       { return fwt.Amount }
func (fwt *FakeWalletTransaction) SetAmount(amount float64) { fwt.Amount = amount }
func (fwt *FakeWalletTransaction) ToParams() NewWalletTransactionParams {
	return NewWalletTransactionParams{}
}
func (fwt *FakeWalletTransaction) ToJSON() ([]byte, error) { return []byte{}, nil }
