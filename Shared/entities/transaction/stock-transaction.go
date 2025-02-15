package transaction

import (
	"Shared/entities/entity"
	"Shared/entities/order"
	"Shared/entities/stock"
	"encoding/json"
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
	ToParams() NewStockTransactionParams
	entity.EntityInterface
}

type StockTransaction struct {
	StockID                  string  `json:"StockID" gorm:"not null"`
	ParentStockTransactionID string  `json:"ParentStockTransactionID"`
	WalletTransactionID      string  `json:"WalletTransactionID"`
	OrderStatus              string  `json:"OrderStatus" gorm:"not null"`
	IsBuy                    bool    `json:"IsBuy" gorm:"not null"`
	OrderType                string  `json:"OrderType" gorm:"not null"`
	StockPrice               float64 `json:"StockPrice" gorm:"not null"`
	Quantity                 int     `json:"Quantity" gorm:"not null"` // If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the StockTransaction Interface.
	GetStockIDInternal                  func() string                         `gorm:"-"`
	SetStockIDInternal                  func(stockID string)                  `gorm:"-"`
	GetParentStockTransactionIDInternal func() string                         `gorm:"-"`
	SetParentStockTransactionIDInternal func(parentStockTransactionID string) `gorm:"-"`
	GetWalletTransactionIDInternal      func() string                         `gorm:"-"`
	SetWalletTransactionIDInternal      func(walletTransactionID string)      `gorm:"-"`
	GetOrderStatusInternal              func() string                         `gorm:"-"`
	SetOrderStatusInternal              func(orderStatus string)              `gorm:"-"`
	GetIsBuyInternal                    func() bool                           `gorm:"-"`
	SetIsBuyInternal                    func(isBuy bool)                      `gorm:"-"`
	GetOrderTypeInternal                func() string                         `gorm:"-"`
	GetStockPriceInternal               func() float64                        `gorm:"-"`
	SetStockPriceInternal               func(stockPrice float64)              `gorm:"-"`
	GetQuantityInternal                 func() int                            `gorm:"-"`
	SetQuantityInternal                 func(quantity int)                    `gorm:"-"`
	entity.BaseEntityInterface          `gorm:"embedded"`
}

func (st *StockTransaction) GetStockID() string {
	return st.GetStockIDInternal()
}

func (st *StockTransaction) SetStockID(stockID string) {
	st.SetStockIDInternal(stockID)
}

func (st *StockTransaction) GetParentStockTransactionID() string {
	return st.GetParentStockTransactionIDInternal()
}

func (st *StockTransaction) SetParentStockTransactionID(parentStockTransactionID string) {
	st.SetParentStockTransactionIDInternal(parentStockTransactionID)
}

func (st *StockTransaction) GetWalletTransactionID() string {
	return st.GetWalletTransactionIDInternal()
}

func (st *StockTransaction) SetWalletTransactionID(walletTransactionID string) {
	st.SetWalletTransactionIDInternal(walletTransactionID)
}

func (st *StockTransaction) GetOrderStatus() string {
	return st.GetOrderStatusInternal()
}

func (st *StockTransaction) SetOrderStatus(orderStatus string) {
	st.SetOrderStatusInternal(orderStatus)
}

func (st *StockTransaction) GetIsBuy() bool {
	return st.GetIsBuyInternal()
}

func (st *StockTransaction) SetIsBuy(isBuy bool) {
	st.SetIsBuyInternal(isBuy)
}

func (st *StockTransaction) GetOrderType() string {
	return st.GetOrderTypeInternal()
}

func (st *StockTransaction) GetStockPrice() float64 {
	return st.GetStockPriceInternal()
}

func (st *StockTransaction) SetStockPrice(stockPrice float64) {
	st.SetStockPriceInternal(stockPrice)
}

func (st *StockTransaction) GetQuantity() int {
	return st.GetQuantityInternal()
}

func (st *StockTransaction) SetQuantity(quantity int) {
	st.SetQuantityInternal(quantity)
}

type NewStockTransactionParams struct {
	entity.NewEntityParams
	StockID                  string  `json:"StockID"`
	ParentStockTransactionID string  `json:"ParentStockTransactionID"`
	WalletTransactionID      string  `json:"WalletTransactionID"`
	OrderStatus              string  `json:"OrderStatus"`
	IsBuy                    bool    `json:"IsBuy"`
	OrderType                string  `json:"OrderType"`
	StockPrice               float64 `json:"StockPrice"`
	Quantity                 int     `json:"Quantity"`

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
	if params.ParentStockTransaction != nil {
		stockID = params.ParentStockTransaction.GetStockID()
		parentStockTransactionID = params.ParentStockTransaction.GetId()
		walletTransactionID = params.ParentStockTransaction.GetWalletTransactionID()
		isBuy = params.ParentStockTransaction.GetIsBuy()
		orderType = params.ParentStockTransaction.GetOrderType()
		stockPrice = params.ParentStockTransaction.GetStockPrice()
		quantity = params.ParentStockTransaction.GetQuantity()
	} else {
		parentStockTransactionID = params.ParentStockTransactionID
		if params.StockOrder != nil {
			stockID = params.StockOrder.GetStockID()
			isBuy = params.StockOrder.GetIsBuy()
			orderType = params.StockOrder.GetOrderType()
			stockPrice = params.StockOrder.GetPrice()
			quantity = params.StockOrder.GetQuantity()
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
		BaseEntityInterface:      e,
	}
	st.SetDefaults()
	return st
}

func (st *StockTransaction) SetDefaults() {
	st.GetStockIDInternal = func() string { return st.StockID }
	st.SetStockIDInternal = func(stockID string) { st.StockID = stockID }
	st.GetParentStockTransactionIDInternal = func() string { return st.ParentStockTransactionID }
	st.SetParentStockTransactionIDInternal = func(parentStockTransactionID string) { st.ParentStockTransactionID = parentStockTransactionID }
	st.GetWalletTransactionIDInternal = func() string { return st.WalletTransactionID }
	st.SetWalletTransactionIDInternal = func(walletTransactionID string) { st.WalletTransactionID = walletTransactionID }
	st.GetOrderStatusInternal = func() string { return st.OrderStatus }
	st.SetOrderStatusInternal = func(orderStatus string) { st.OrderStatus = orderStatus }
	st.GetIsBuyInternal = func() bool { return st.IsBuy }
	st.SetIsBuyInternal = func(isBuy bool) { st.IsBuy = isBuy }
	st.GetOrderTypeInternal = func() string { return st.OrderType }
	st.GetStockPriceInternal = func() float64 { return st.StockPrice }
	st.SetStockPriceInternal = func(stockPrice float64) { st.StockPrice = stockPrice }
	st.GetQuantityInternal = func() int { return st.Quantity }
	st.SetQuantityInternal = func(quantity int) { st.Quantity = quantity }
}

func ParseStockTransaction(jsonBytes []byte) (*StockTransaction, error) {
	var st NewStockTransactionParams
	if err := json.Unmarshal(jsonBytes, &st); err != nil {
		return nil, err
	}
	return NewStockTransaction(st), nil
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
		WalletTransaction:        nil,
		ParentStockTransaction:   nil,
		StockOrder:               nil,
		Stock:                    nil,
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
