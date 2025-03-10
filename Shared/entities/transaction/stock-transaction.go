package transaction

import (
	"Shared/entities/entity"
	"Shared/entities/order"
	"Shared/entities/stock"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type StockTransactionInterface interface {
	GetStockID() *uuid.UUID
	GetStockIDString() string
	SetStockID(stockID *uuid.UUID)
	GetParentStockTransactionID() *uuid.UUID
	GetParentStockTransactionIDString() string
	SetParentStockTransactionID(parentStockTransactionID *uuid.UUID)
	GetWalletTransactionID() *uuid.UUID
	GetWalletTransactionIDString() string
	SetWalletTransactionID(walletTransactionID *uuid.UUID)
	GetOrderStatus() string
	SetOrderStatus(orderStatus string)
	GetIsBuy() bool
	SetIsBuy(isBuy bool)
	GetOrderType() string
	GetStockPrice() float64
	UpdateStockPrice(stockPrice float64)
	GetQuantity() int
	SetQuantity(quantity int)
	GetTimestamp() time.Time
	SetTimestamp(timestamp time.Time)
	SetStockTXID()
	GetUserID() *uuid.UUID
	GetUserIDString() string
	SetUserID(userID *uuid.UUID)
	ToParams() NewStockTransactionParams
	entity.EntityInterface
}

type StockTransaction struct {
	StockTXID                *uuid.UUID `json:"stock_tx_id" gorm:"-"` // Stock Transaction ID
	StockID                  *uuid.UUID `json:"stock_id" gorm:"column:stock_id;type:uuid;not null"`
	ParentStockTransactionID *uuid.UUID `json:"parent_stock_tx_id" gorm:"column:parent_stock_transaction_id;type:uuid"`
	WalletTransactionID      *uuid.UUID `json:"wallet_tx_id" gorm:"column:wallet_transaction_id;type:uuid"`
	OrderStatus              string     `json:"order_status" gorm:"not null"`
	IsBuy                    bool       `json:"is_buy" gorm:"not null"`
	OrderType                string     `json:"order_type" gorm:"not null"`
	StockPrice               float64    `json:"stock_price" gorm:"not null"`
	Quantity                 int        `json:"quantity" gorm:"not null"`
	Timestamp                time.Time  `json:"time_stamp"`
	UserID                   *uuid.UUID `json:"user_id" gorm:"type:uuid;column:user_id;not null"`
	entity.Entity            `json:"entity" gorm:"embedded"`
}

func (st *StockTransaction) GetStockID() *uuid.UUID {
	return st.StockID
}

func (st *StockTransaction) GetStockIDString() string {
	if st.StockID == nil {
		return ""
	}
	return st.StockID.String()
}

func (st *StockTransaction) SetStockID(stockID *uuid.UUID) {
	st.StockID = stockID
}

func (st *StockTransaction) GetParentStockTransactionID() *uuid.UUID {
	return st.ParentStockTransactionID
}

func (st *StockTransaction) GetParentStockTransactionIDString() string {
	if st.ParentStockTransactionID == nil {
		return ""
	}
	return st.ParentStockTransactionID.String()
}

func (st *StockTransaction) SetParentStockTransactionID(parentStockTransactionID *uuid.UUID) {
	st.ParentStockTransactionID = parentStockTransactionID
}

func (st *StockTransaction) GetWalletTransactionID() *uuid.UUID {
	return st.WalletTransactionID
}

func (st *StockTransaction) GetWalletTransactionIDString() string {
	if st.WalletTransactionID == nil {
		return ""
	}
	return st.WalletTransactionID.String()
}

func (st *StockTransaction) SetWalletTransactionID(walletTransactionID *uuid.UUID) {
	st.WalletTransactionID = walletTransactionID
	*st.GetUpdates() = append(*st.Updates, &entity.EntityUpdateData{
		ID:    st.GetId(),
		Field: "WalletTransactionID",
		NewValue: func() *string {
			if walletTransactionID != nil {
				s := walletTransactionID.String()
				return &s
			}
			return nil
		}(),
	})
}

func (st *StockTransaction) GetOrderStatus() string {
	return st.OrderStatus
}

func (st *StockTransaction) SetOrderStatus(orderStatus string) {
	st.OrderStatus = orderStatus
	*st.GetUpdates() = append(*st.Updates, &entity.EntityUpdateData{
		ID:       st.GetId(),
		Field:    "OrderStatus",
		NewValue: &orderStatus,
	})
}

func (st *StockTransaction) GetIsBuy() bool {
	return st.IsBuy
}

func (st *StockTransaction) SetIsBuy(isBuy bool) {
	st.IsBuy = isBuy
	*st.GetUpdates() = append(*st.Updates, &entity.EntityUpdateData{
		ID:       st.GetId(),
		Field:    "IsBuy",
		NewValue: func() *string { s := strconv.FormatBool(isBuy); return &s }(),
	})
}

func (st *StockTransaction) GetOrderType() string {
	return st.OrderType
}

func (st *StockTransaction) GetStockPrice() float64 {
	return st.StockPrice
}

func (st *StockTransaction) UpdateStockPrice(stockPrice float64) {
	st.StockPrice = stockPrice
	*st.GetUpdates() = append(*st.Updates, &entity.EntityUpdateData{
		ID:         st.GetId(),
		Field:      "StockPrice",
		AlterValue: func() *string { s := strconv.FormatFloat(stockPrice, 'f', -1, 64); return &s }(),
	})
}

func (st *StockTransaction) GetQuantity() int {
	return st.Quantity
}

func (st *StockTransaction) SetQuantity(quantity int) {
	st.Quantity = quantity
	*st.GetUpdates() = append(*st.Updates, &entity.EntityUpdateData{
		ID:       st.GetId(),
		Field:    "Quantity",
		NewValue: func() *string { s := strconv.Itoa(quantity); return &s }(),
	})
}

func (st *StockTransaction) GetTimestamp() time.Time {
	return st.Timestamp
}

func (st *StockTransaction) SetTimestamp(timestamp time.Time) {
	st.Timestamp = timestamp
	*st.GetUpdates() = append(*st.Updates, &entity.EntityUpdateData{
		ID:       st.GetId(),
		Field:    "Timestamp",
		NewValue: func() *string { s := timestamp.Format(time.RFC3339); return &s }(),
	})
}

func (st *StockTransaction) SetStockTXID() {
	st.StockTXID = st.GetId()
}

func (st *StockTransaction) GetUserID() *uuid.UUID {
	return st.UserID
}

func (st *StockTransaction) GetUserIDString() string {
	if st.UserID == nil {
		return ""
	}
	return st.UserID.String()
}

func (st *StockTransaction) SetUserID(userID *uuid.UUID) {
	st.UserID = userID
}

type NewStockTransactionParams struct {
	entity.NewEntityParams   `json:"entity"`
	StockID                  *uuid.UUID `json:"stock_id"`
	ParentStockTransactionID *uuid.UUID `json:"parent_stock_tx_id"`
	WalletTransactionID      *uuid.UUID `json:"wallet_tx_id"`
	OrderStatus              string     `json:"order_status"`
	IsBuy                    bool       `json:"is_buy"`
	OrderType                string     `json:"order_type"`
	StockPrice               float64    `json:"stock_price"`
	Quantity                 int        `json:"quantity"`
	TimeStamp                time.Time  `json:"time_stamp"`
	UserID                   *uuid.UUID `json:"user_id"`

	WalletTransaction WalletTransactionInterface // use this or WalletTransactionID or ParentStockTransaction
	//use one of the following
	ParentStockTransaction StockTransactionInterface
	//or
	// And one of the following
	StockOrder order.StockOrderInterface
	//or
	Stock stock.StockInterface // use this or StockID
}

func NewStockTransaction(params NewStockTransactionParams) *StockTransaction {
	e := entity.NewEntity(params.NewEntityParams)
	var stockID *uuid.UUID
	var parentStockTransactionID *uuid.UUID
	var walletTransactionID *uuid.UUID
	var isBuy bool
	var orderType string
	var stockPrice float64
	var quantity int
	var userID *uuid.UUID
	if params.ParentStockTransaction != nil {
		stockID = params.ParentStockTransaction.GetStockID()
		parentStockTransactionID = params.ParentStockTransaction.GetId()
		walletTransactionID = params.ParentStockTransaction.GetWalletTransactionID()
		isBuy = params.ParentStockTransaction.GetIsBuy()
		orderType = params.ParentStockTransaction.GetOrderType()
		stockPrice = params.ParentStockTransaction.GetStockPrice()
		quantity = params.ParentStockTransaction.GetQuantity()
		userID = params.ParentStockTransaction.GetUserID()
	} else {
		parentStockTransactionID = params.ParentStockTransactionID
		if params.StockOrder != nil {
			e.ID = params.StockOrder.GetId()
			stockID = params.StockOrder.GetStockID()
			isBuy = params.StockOrder.GetIsBuy()
			orderType = params.StockOrder.GetOrderType()
			stockPrice = params.StockOrder.GetPrice()
			quantity = params.StockOrder.GetQuantity()
			userID = params.StockOrder.GetUserID()

		} else {
			if params.Stock != nil {
				stockID = params.Stock.GetId()
			} else {
				stockID = params.StockID
			}
			isBuy = params.IsBuy
			orderType = params.OrderType
			stockPrice = params.StockPrice
			quantity = params.Quantity
			userID = params.UserID
		}
	}

	if params.WalletTransaction != nil {
		walletTransactionID = params.WalletTransaction.GetId()
	} else {
		walletTransactionID = params.WalletTransactionID
	}

	st := &StockTransaction{
		StockID:                  stockID,
		ParentStockTransactionID: parentStockTransactionID,
		WalletTransactionID:      walletTransactionID,
		OrderStatus:              params.OrderStatus,
		IsBuy:                    isBuy,
		OrderType:                orderType,
		StockPrice:               stockPrice,
		Quantity:                 quantity,
		Timestamp:                params.TimeStamp,
		UserID:                   userID,
		Entity:                   *e,
	}
	return st
}

func ParseStockTransaction(jsonBytes []byte) (*StockTransaction, error) {
	var st NewStockTransactionParams
	if err := json.Unmarshal(jsonBytes, &st); err != nil {
		return nil, err
	}
	return NewStockTransaction(st), nil
}

func ParseStockTransactionList(jsonBytes []byte) (*[]*StockTransaction, error) {
	var so []NewStockTransactionParams
	if err := json.Unmarshal(jsonBytes, &so); err != nil {
		return nil, err
	}
	soList := make([]*StockTransaction, len(so))
	for i, s := range so {
		soList[i] = NewStockTransaction(s)
	}
	return &soList, nil
}

func (st *StockTransaction) ToParams() NewStockTransactionParams {
	return NewStockTransactionParams{
		NewEntityParams:          st.EntityToParams(),
		StockID:                  st.GetStockID(),
		ParentStockTransactionID: st.GetParentStockTransactionID(),
		WalletTransactionID:      st.GetWalletTransactionID(),
		OrderStatus:              st.GetOrderStatus(),
		IsBuy:                    st.GetIsBuy(),
		OrderType:                st.GetOrderType(),
		StockPrice:               st.GetStockPrice(),
		Quantity:                 st.GetQuantity(),
		TimeStamp:                st.GetTimestamp(),
		UserID:                   st.GetUserID(),
	}
}

func (st *StockTransaction) ToJSON() ([]byte, error) {
	return json.Marshal(st.ToParams())
}
