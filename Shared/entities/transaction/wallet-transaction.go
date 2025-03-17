package transaction

import (
	"Shared/entities/entity"
	"Shared/entities/wallet"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type WalletTransactionInterface interface {
	GetWalletID() *uuid.UUID
	GetWalletIDString() string
	SetWalletID(walletID *uuid.UUID)
	GetStockTransactionID() *uuid.UUID
	GetStockTransactionIDString() string
	SetStockTransactionID(stockTransactionID *uuid.UUID)
	GetIsDebit() bool
	SetIsDebit(isDebit bool)
	GetAmount() float64
	SetAmount(amount float64)
	GetTimestamp() time.Time
	SetTimestamp(timestamp time.Time)
	SetWalletTXID()
	GetUserID() *uuid.UUID
	GetUserIDString() string
	SetUserID(userID *uuid.UUID)
	ToParams() NewWalletTransactionParams
	entity.EntityInterface
}

type WalletTransaction struct {
	WalletID           *uuid.UUID `json:"wallet_id" gorm:"column:wallet_id;type:uuid;not null"`
	WalletTXID         *uuid.UUID `json:"wallet_tx_id" gorm:"-"`
	StockTransactionID *uuid.UUID `json:"stock_tx_id" gorm:"column:stock_transaction_id;type:uuid;not null"`
	IsDebit            bool       `json:"is_debit" gorm:"not null"`
	Amount             float64    `json:"amount" gorm:"not null"`
	Timestamp          time.Time  `json:"time_stamp"`
	UserID             *uuid.UUID `json:"user_id" gorm:"type:uuid;column:user_id;not null"`
	entity.Entity      `json:"Entity" gorm:"embedded"`
}

func (wt *WalletTransaction) GetWalletID() *uuid.UUID {
	return wt.WalletID
}

func (wt *WalletTransaction) GetWalletIDString() string {
	if wt.WalletID == nil {
		return ""
	}
	return wt.WalletID.String()
}

func (wt *WalletTransaction) SetWalletID(walletID *uuid.UUID) {
	wt.WalletID = walletID
}

func (wt *WalletTransaction) GetStockTransactionID() *uuid.UUID {
	return wt.StockTransactionID
}

func (wt *WalletTransaction) GetStockTransactionIDString() string {
	if wt.StockTransactionID == nil {
		return ""
	}
	return wt.StockTransactionID.String()
}

func (wt *WalletTransaction) SetStockTransactionID(stockTransactionID *uuid.UUID) {
	wt.StockTransactionID = stockTransactionID
	*wt.GetUpdates() = append(*wt.Updates, &entity.EntityUpdateData{
		ID:       wt.GetId(),
		Field:    "StockTransactionID",
		NewValue: func() *string { s := stockTransactionID.String(); return &s }(),
	})
}

func (wt *WalletTransaction) GetIsDebit() bool {
	return wt.IsDebit
}

func (wt *WalletTransaction) SetIsDebit(isDebit bool) {
	wt.IsDebit = isDebit
	*wt.GetUpdates() = append(*wt.Updates, &entity.EntityUpdateData{
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
	*wt.GetUpdates() = append(*wt.Updates, &entity.EntityUpdateData{
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
	*wt.GetUpdates() = append(*wt.Updates, &entity.EntityUpdateData{
		ID:       wt.GetId(),
		Field:    "Timestamp",
		NewValue: func() *string { s := timestamp.Format(time.RFC3339); return &s }(),
	})
}

func (wt *WalletTransaction) SetWalletTXID() {
	wt.WalletTXID = wt.GetId()
}

func (st *WalletTransaction) GetUserID() *uuid.UUID {
	return st.UserID
}

func (st *WalletTransaction) GetUserIDString() string {
	if st.UserID == nil {
		return ""
	}
	return st.UserID.String()
}

func (st *WalletTransaction) SetUserID(userID *uuid.UUID) {
	st.UserID = userID
}

type NewWalletTransactionParams struct {
	entity.NewEntityParams `json:"Entity"`
	WalletID               *uuid.UUID `json:"wallet_id" gorm:"not null"`
	StockTransactionID     *uuid.UUID `json:"stock_tx_id" gorm:"not null"`
	IsDebit                bool       `json:"is_debit" gorm:"not null"`
	Amount                 float64    `json:"amount" gorm:"not null"`
	Timestamp              time.Time  `json:"time_stamp"`
	Wallet                 wallet.WalletInterface
	StockTransaction       StockTransactionInterface
	UserID                 *uuid.UUID `json:"user_id"`
}

func NewWalletTransaction(params NewWalletTransactionParams) *WalletTransaction {
	e := entity.NewEntity(params.NewEntityParams)
	wt := &WalletTransaction{
		Entity:    *e,
		IsDebit:   params.IsDebit,
		Amount:    params.Amount,
		Timestamp: params.Timestamp,
		UserID:    params.UserID,
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
		UserID:             wt.GetUserID(),
	}
}

func (wt *WalletTransaction) ToJSON() ([]byte, error) {
	return json.Marshal(wt.ToParams())
}
