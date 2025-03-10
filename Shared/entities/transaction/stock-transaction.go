package transaction

import (
	"Shared/entities/entity"
	"Shared/entities/order"
	"Shared/entities/stock"
	"encoding/json"
	"time"
)

type StockTransactionInterface interface {
	GetStockID() string
	SetStockID(stockID string)
	GetParentStockTransactionID() string
	SetParentStockTransactionID(parentStockTransactionID string)
	GetWalletTransactionID() string
	SetWalletTransactionID(walletTransactionID string)
	GetOrderStatus() string
	SetOrderStatus(orderStatus string)
	GetIsBuy() bool
	SetIsBuy(isBuy bool)
	GetOrderType() string
	GetStockPrice() float64
	SetStockPrice(stockPrice float64)
	GetQuantity() int
	SetQuantity(quantity int)
	GetTimestamp() time.Time
	SetTimestamp(timestamp time.Time)
	SetStockTXID()
	GetUserID() string
	SetUserID(userID string)
	ToParams() NewStockTransactionParams
	entity.EntityInterface
}

type StockTransaction struct {
	StockTXID                string    `json:"stock_tx_id" gorm:"-"` // Stock Transaction ID
	StockID                  string    `json:"stock_id" gorm:"not null"`
	ParentStockTransactionID string    `json:"parent_stock_tx_id"`
	WalletTransactionID      string    `json:"wallet_tx_id"`
	OrderStatus              string    `json:"order_status" gorm:"not null"`
	IsBuy                    bool      `json:"is_buy" gorm:"not null"`
	OrderType                string    `json:"order_type" gorm:"not null"`
	StockPrice               float64   `json:"stock_price" gorm:"not null"`
	Quantity                 int       `json:"quantity" gorm:"not null"`
	Timestamp                time.Time `json:"time_stamp"`
	UserID                   string    `json:"user_id" gorm:"not null"`
	// Internal Functions (commented out)
	// GetStockIDInternal                  func() string                         `gorm:"-"`
	// SetStockIDInternal                  func(stockID string)                  `gorm:"-"`
	// GetParentStockTransactionIDInternal func() string                         `gorm:"-"`
	// SetParentStockTransactionIDInternal func(parentStockTransactionID string) `gorm:"-"`
	// GetWalletTransactionIDInternal      func() string                         `gorm:"-"`
	// SetWalletTransactionIDInternal      func(walletTransactionID string)      `gorm:"-"`
	// GetOrderStatusInternal              func() string                         `gorm:"-"`
	// SetOrderStatusInternal              func(orderStatus string)              `gorm:"-"`
	// GetIsBuyInternal                    func() bool                           `gorm:"-"`
	// SetIsBuyInternal                    func(isBuy bool)                      `gorm:"-"`
	// GetOrderTypeInternal                func() string                         `gorm:"-"`
	// GetStockPriceInternal               func() float64                        `gorm:"-"`
	// SetStockPriceInternal               func(stockPrice float64)              `gorm:"-"`
	// GetQuantityInternal                 func() int                            `gorm:"-"`
	// SetQuantityInternal                 func(quantity int)                    `gorm:"-"`
	entity.Entity `json:"entity" gorm:"embedded"`
}

func (st *StockTransaction) GetStockID() string {
	// return st.GetStockIDInternal()
	return st.StockID
}

func (st *StockTransaction) SetStockID(stockID string) {
	// st.SetStockIDInternal(stockID)
	st.StockID = stockID
}

func (st *StockTransaction) GetParentStockTransactionID() string {
	// return st.GetParentStockTransactionIDInternal()
	return st.ParentStockTransactionID
}

func (st *StockTransaction) SetParentStockTransactionID(parentStockTransactionID string) {
	// st.SetParentStockTransactionIDInternal(parentStockTransactionID)
	st.ParentStockTransactionID = parentStockTransactionID
}

func (st *StockTransaction) GetWalletTransactionID() string {
	// return st.GetWalletTransactionIDInternal()
	return st.WalletTransactionID
}

func (st *StockTransaction) SetWalletTransactionID(walletTransactionID string) {
	// st.SetWalletTransactionIDInternal(walletTransactionID)
	st.WalletTransactionID = walletTransactionID
}

func (st *StockTransaction) GetOrderStatus() string {
	// return st.GetOrderStatusInternal()
	return st.OrderStatus
}

func (st *StockTransaction) SetOrderStatus(orderStatus string) {
	// st.SetOrderStatusInternal(orderStatus)
	st.OrderStatus = orderStatus
}

func (st *StockTransaction) GetIsBuy() bool {
	// return st.GetIsBuyInternal()
	return st.IsBuy
}

func (st *StockTransaction) SetIsBuy(isBuy bool) {
	// st.SetIsBuyInternal(isBuy)
	st.IsBuy = isBuy
}

func (st *StockTransaction) GetOrderType() string {
	// return st.GetOrderTypeInternal()
	return st.OrderType
}

func (st *StockTransaction) GetStockPrice() float64 {
	// return st.GetStockPriceInternal()
	return st.StockPrice
}

func (st *StockTransaction) SetStockPrice(stockPrice float64) {
	// st.SetStockPriceInternal(stockPrice)
	st.StockPrice = stockPrice
}

func (st *StockTransaction) GetQuantity() int {
	// return st.GetQuantityInternal()
	return st.Quantity
}

func (st *StockTransaction) SetQuantity(quantity int) {
	// st.SetQuantityInternal(quantity)
	st.Quantity = quantity
}

func (st *StockTransaction) GetTimestamp() time.Time {
	return st.Timestamp
}

func (st *StockTransaction) SetTimestamp(timestamp time.Time) {
	st.Timestamp = timestamp
}

func (st *StockTransaction) SetStockTXID() {
	st.StockTXID = st.GetId()
}

func (st *StockTransaction) GetUserID() string {
	return st.UserID
}

func (st *StockTransaction) SetUserID(userID string) {
	st.UserID = userID
}

type NewStockTransactionParams struct {
	entity.NewEntityParams   `json:"entity"`
	StockID                  string    `json:"stock_id"`
	ParentStockTransactionID string    `json:"parent_stock_tx_id"`
	WalletTransactionID      string    `json:"wallet_tx_id"`
	OrderStatus              string    `json:"order_status"`
	IsBuy                    bool      `json:"is_buy"`
	OrderType                string    `json:"order_type"`
	StockPrice               float64   `json:"stock_price"`
	Quantity                 int       `json:"quantity"`
	TimeStamp                time.Time `json:"time_stamp"`
	UserID                   string    `json:"user_id"`

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
	var stockID string
	var parentStockTransactionID string
	var walletTransactionID string
	var isBuy bool
	var orderType string
	var stockPrice float64
	var quantity int
	var userID string
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

type FakeStockTransaction struct {
	entity.FakeEntity
	StockID                  string `json:"stockID"`
	ParentStockTransactionID string `json:"parentStockTransactionID"`
	WalletTransactionID      string `json:"walletTransactionID"`
	OrderStatus              string `json:"orderStatus"`
	IsBuy                    bool   `json:"isBuy"`
	OrderType                string `json:"orderType"`
	StockPrice               float64
	Quantity                 int
}

func (fst *FakeStockTransaction) GetStockID() string        { return fst.StockID }
func (fst *FakeStockTransaction) SetStockID(stockID string) { fst.StockID = stockID }
func (fst *FakeStockTransaction) GetParentStockTransactionID() string {
	return fst.ParentStockTransactionID
}
func (fst *FakeStockTransaction) SetParentStockTransactionID(parentStockTransactionID string) {
	fst.ParentStockTransactionID = parentStockTransactionID
}
func (fst *FakeStockTransaction) GetWalletTransactionID() string { return fst.WalletTransactionID }
func (fst *FakeStockTransaction) SetWalletTransactionID(walletTransactionID string) {
	fst.WalletTransactionID = walletTransactionID
}
func (fst *FakeStockTransaction) GetOrderStatus() string            { return fst.OrderStatus }
func (fst *FakeStockTransaction) SetOrderStatus(orderStatus string) { fst.OrderStatus = orderStatus }
func (fst *FakeStockTransaction) GetIsBuy() bool                    { return fst.IsBuy }
func (fst *FakeStockTransaction) SetIsBuy(isBuy bool)               { fst.IsBuy = isBuy }
func (fst *FakeStockTransaction) GetOrderType() string              { return fst.OrderType }
func (fst *FakeStockTransaction) SetOrderType(orderType string)     { fst.OrderType = orderType }
func (fst *FakeStockTransaction) GetStockPrice() float64            { return fst.StockPrice }
func (fst *FakeStockTransaction) SetStockPrice(stockPrice float64)  { fst.StockPrice = stockPrice }
func (fst *FakeStockTransaction) GetQuantity() int                  { return fst.Quantity }
func (fst *FakeStockTransaction) SetQuantity(quantity int)          { fst.Quantity = quantity }
func (fst *FakeStockTransaction) ToParams() NewStockTransactionParams {
	return NewStockTransactionParams{}
}
func (fst *FakeStockTransaction) ToJSON() ([]byte, error) { return []byte{}, nil }