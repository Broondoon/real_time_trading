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
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the WalletTransaction Interface.
	GetWalletIDInternal           func() string                   `gorm:"-"`
	SetWalletIDInternal           func(walletID string)           `gorm:"-"`
	GetStockTransactionIDInternal func() string                   `gorm:"-"`
	SetStockTransactionIDInternal func(stockTransactionID string) `gorm:"-"`
	GetIsDebitInternal            func() bool                     `gorm:"-"`
	SetIsDebitInternal            func(isDebit bool)              `gorm:"-"`
	GetAmountInternal             func() float64                  `gorm:"-"`
	SetAmountInternal             func(amount float64)            `gorm:"-"`
	entity.BaseEntityInterface    `gorm:"embedded"`
}

func (wt *WalletTransaction) GetWalletID() string {
	return wt.GetWalletIDInternal()
}

func (wt *WalletTransaction) SetWalletID(walletID string) {
	wt.SetWalletIDInternal(walletID)
}

func (wt *WalletTransaction) GetStockTransactionID() string {
	return wt.GetStockTransactionIDInternal()
}

func (wt *WalletTransaction) SetStockTransactionID(stockTransactionID string) {
	wt.SetStockTransactionIDInternal(stockTransactionID)
}

func (wt *WalletTransaction) GetIsDebit() bool {
	return wt.GetIsDebitInternal()
}

func (wt *WalletTransaction) SetIsDebit(isDebit bool) {
	wt.SetIsDebitInternal(isDebit)
}

func (wt *WalletTransaction) GetAmount() float64 {
	return wt.GetAmountInternal()
}

func (wt *WalletTransaction) SetAmount(amount float64) {
	wt.SetAmountInternal(amount)
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
		BaseEntityInterface: e,
		IsDebit:             params.IsDebit,
		Amount:              params.Amount,
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
	wt.GetWalletIDInternal = func() string { return wt.WalletID }
	wt.SetWalletIDInternal = func(walletID string) { wt.WalletID = walletID }
	wt.GetStockTransactionIDInternal = func() string { return wt.StockTransactionID }
	wt.SetStockTransactionIDInternal = func(stockTransactionID string) { wt.StockTransactionID = stockTransactionID }

	wt.GetIsDebitInternal = func() bool { return wt.IsDebit }
	wt.SetIsDebitInternal = func(isDebit bool) { wt.IsDebit = isDebit }

	wt.GetAmountInternal = func() float64 { return wt.Amount }
	wt.SetAmountInternal = func(amount float64) { wt.Amount = amount }
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
