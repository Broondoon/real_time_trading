package transaction

import (
	"Shared/entities/entity"
	"Shared/entities/wallet"
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
	entity.EntityInterface
}

type WalletTransaction struct {
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	WalletID           string
	StockTransactionID string
	IsDebit            bool
	Amount             float64
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the WalletTransaction Interface.
	GetWalletIDInternal           func() string
	SetWalletIDInternal           func(walletID string)
	GetStockTransactionIDInternal func() string
	SetStockTransactionIDInternal func(stockTransactionID string)
	GetIsDebitInternal            func() bool
	SetIsDebitInternal            func(isDebit bool)
	GetAmountInternal             func() float64
	SetAmountInternal             func(amount float64)
	entity.EntityInterface
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
	WalletID           string // use this or Wallet
	Wallet             wallet.WalletInterface
	StockTransactionID string // use this or StockTransaction
	StockTransaction   StockTransactionInterface
	IsDebit            bool
	Amount             float64
}

func NewWalletTransaction(params NewWalletTransactionParams) *WalletTransaction {
	e := entity.NewEntity(params.NewEntityParams)
	wt := &WalletTransaction{
		EntityInterface: e,
		IsDebit:         params.IsDebit,
		Amount:          params.Amount,
	}
	var WalletID string
	if params.Wallet != nil {
		WalletID = params.Wallet.GetId()
	} else {
		WalletID = params.WalletID
	}
	wt.GetWalletIDInternal = func() string { return WalletID }
	wt.SetWalletIDInternal = func(walletID string) { WalletID = walletID }

	var StockTransactionID string
	if params.StockTransaction != nil {
		StockTransactionID = params.StockTransaction.GetId()
	} else {
		StockTransactionID = params.StockTransactionID
	}
	wt.GetStockTransactionIDInternal = func() string { return StockTransactionID }
	wt.SetStockTransactionIDInternal = func(stockTransactionID string) { StockTransactionID = stockTransactionID }

	wt.GetIsDebitInternal = func() bool { return wt.IsDebit }
	wt.SetIsDebitInternal = func(isDebit bool) { wt.IsDebit = isDebit }

	wt.GetAmountInternal = func() float64 { return wt.Amount }
	wt.SetAmountInternal = func(amount float64) { wt.Amount = amount }

	return wt
}
