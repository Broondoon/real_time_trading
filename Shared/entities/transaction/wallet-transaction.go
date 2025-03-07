package transaction

import (
	"Shared/entities/entity"
	"Shared/entities/wallet"
	"encoding/json"
	"strconv"
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
	SetWalletTXID()
	ToParams() NewWalletTransactionParams
	entity.EntityInterface
}

type WalletTransaction struct {
	WalletID           string    `json:"wallet_id" gorm:"not null"`
	WalletTXID         string    `json:"wallet_tx_id" gorm:"-"`
	StockTransactionID string    `json:"stock_tx_id" gorm:"not null"`
	IsDebit            bool      `json:"is_debit" gorm:"not null"`
	Amount             float64   `json:"amount" gorm:"not null"`
	Timestamp          time.Time `json:"time_stamp"`
	entity.Entity      `json:"Entity" gorm:"embedded"`
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
	wt.Updates = append(wt.Updates, &entity.EntityUpdateData{
		ID:       wt.GetId(),
		Field:    "StockTransactionID",
		NewValue: &stockTransactionID,
	})
}

func (wt *WalletTransaction) GetIsDebit() bool {
	return wt.IsDebit
}

func (wt *WalletTransaction) SetIsDebit(isDebit bool) {
	wt.IsDebit = isDebit
	wt.Updates = append(wt.Updates, &entity.EntityUpdateData{
		ID:       wt.GetId(),
		Field:    "IsDebit",
		NewValue: func() *string { s := strconv.FormatBool(isDebit); return &s }(),
	})
}

func (wt *WalletTransaction) GetAmount() float64 {
	return wt.Amount
}

func (wt *WalletTransaction) SetAmount(amount float64) {
	wt.Amount = amount
	wt.Updates = append(wt.Updates, &entity.EntityUpdateData{
		ID:       wt.GetId(),
		Field:    "Amount",
		NewValue: func() *string { s := strconv.FormatFloat(amount, 'f', -1, 64); return &s }(),
	})
}

func (wt *WalletTransaction) GetTimestamp() time.Time {
	return wt.Timestamp
}

func (wt *WalletTransaction) SetTimestamp(timestamp time.Time) {
	wt.Timestamp = timestamp
	wt.Updates = append(wt.Updates, &entity.EntityUpdateData{
		ID:       wt.GetId(),
		Field:    "Timestamp",
		NewValue: func() *string { s := timestamp.Format(time.RFC3339); return &s }(),
	})
}

func (wt *WalletTransaction) SetWalletTXID() {
	wt.WalletTXID = wt.GetId()
	wt.Updates = append(wt.Updates, &entity.EntityUpdateData{
		ID:       wt.GetId(),
		Field:    "WalletTXID",
		NewValue: &wt.WalletTXID,
	})
}

type NewWalletTransactionParams struct {
	entity.NewEntityParams `json:"Entity"`
	WalletID               string    `json:"wallet_id" gorm:"not null"`
	StockTransactionID     string    `json:"stock_tx_id" gorm:"not null"`
	IsDebit                bool      `json:"is_debit" gorm:"not null"`
	Amount                 float64   `json:"amount" gorm:"not null"`
	Timestamp              time.Time `json:"time_stamp"`
	Wallet                 wallet.WalletInterface
	StockTransaction       StockTransactionInterface
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
