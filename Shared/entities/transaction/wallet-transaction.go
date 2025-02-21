package transaction

import (
	"Shared/entities/entity"
	"Shared/entities/wallet"
	"encoding/json"
	"time"
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
	GetTimestamp() time.Time
	SetTimestamp(timestamp time.Time)
	ToParams() NewWalletTransactionParams
	entity.EntityInterface
}

type WalletTransaction struct {
	WalletID           string    `json:"WalletID" gorm:"not null"`
	StockTransactionID string    `json:"StockTransactionID" gorm:"not null"`
	IsDebit            bool      `json:"IsDebit" gorm:"not null"`
	Amount             float64   `json:"Amount" gorm:"not null"`
	Timestamp          time.Time `json:"time_stamp"`

	// Internal functions have been commented out.
	// GetWalletIDInternal           func() string                   `gorm:"-"`
	// SetWalletIDInternal           func(walletID string)           `gorm:"-"`
	// GetStockTransactionIDInternal func() string                   `gorm:"-"`
	// SetStockTransactionIDInternal func(stockTransactionID string) `gorm:"-"`
	// GetIsDebitInternal            func() bool                     `gorm:"-"`
	// SetIsDebitInternal            func(isDebit bool)              `gorm:"-"`
	// GetAmountInternal             func() float64                  `gorm:"-"`
	// SetAmountInternal             func(amount float64)            `gorm:"-"`
	entity.Entity `json:"Entity" gorm:"embedded"`
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

func (wt *WalletTransaction) GetTimestamp() time.Time {
	return wt.Timestamp
}

func (wt *WalletTransaction) SetTimestamp(timestamp time.Time) {
	wt.Timestamp = timestamp
}

type NewWalletTransactionParams struct {
	entity.NewEntityParams
	WalletID           string    `json:"WalletID" gorm:"not null"`
	StockTransactionID string    `json:"StockTransactionID" gorm:"not null"`
	IsDebit            bool      `json:"IsDebit" gorm:"not null"`
	Amount             float64   `json:"Amount" gorm:"not null"`
	Timestamp          time.Time `json:"time_stamp"`
	Wallet             wallet.WalletInterface
	StockTransaction   StockTransactionInterface
}

func NewWalletTransaction(params NewWalletTransactionParams) *WalletTransaction {
	e := entity.NewEntity(params.NewEntityParams)
	wt := &WalletTransaction{
		Entity:    *e,
		IsDebit:   params.IsDebit,
		Amount:    params.Amount,
		Timestamp: params.Timestamp,
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

	return wt
}

func ParseWalletTransaction(jsonBytes []byte) (*WalletTransaction, error) {
	var wt NewWalletTransactionParams
	if err := json.Unmarshal(jsonBytes, &wt); err != nil {
		return nil, err
	}
	return NewWalletTransaction(wt), nil
}

func ParseWalletTransactionList(jsonBytes []byte) (*[]*WalletTransaction, error) {
	var so []NewWalletTransactionParams
	if err := json.Unmarshal(jsonBytes, &so); err != nil {
		return nil, err
	}
	soList := make([]*WalletTransaction, len(so))
	for i, s := range so {
		soList[i] = NewWalletTransaction(s)
	}
	return &soList, nil
}

func (wt *WalletTransaction) ToParams() NewWalletTransactionParams {
	return NewWalletTransactionParams{
		NewEntityParams:    wt.EntityToParams(),
		WalletID:           wt.GetWalletID(),
		StockTransactionID: wt.GetStockTransactionID(),
		IsDebit:            wt.GetIsDebit(),
		Amount:             wt.GetAmount(),
		Timestamp:          wt.GetTimestamp(),
	}
}

func (wt *WalletTransaction) ToJSON() ([]byte, error) {
	return json.Marshal(wt.ToParams())
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
