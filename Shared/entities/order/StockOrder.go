package order

import (
	"Shared/entities/entity"
	"Shared/entities/stock"
	"encoding/json"
)

const (
	OrderTypeMarket = "MARKET"
	OrderTypeLimit  = "LIMIT"
)

type StockOrderInterface interface {
	GetStockID() string
	SetStockID(stockID string)
	GetIsBuy() bool
	SetIsBuy(isBuy bool)
	GetOrderType() string
	GetQuantity() int
	SetQuantity(quantity int)
	GetPrice() float64
	SetPrice(price float64)
	StockOrderToParams() NewStockOrderParams
	StockOrderToJSON() ([]byte, error)
	entity.EntityInterface
}

type StockOrder struct {
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	StockID   string
	IsBuy     bool
	OrderType string
	Quantity  int
	Price     float64
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the Interface.
	GetStockIDInternal   func() string
	SetStockIDInternal   func(stockID string)
	GetIsBuyInternal     func() bool
	SetIsBuyInternal     func(isBuy bool)
	GetOrderTypeInternal func() string
	GetQuantityInternal  func() int
	SetQuantityInternal  func(quantity int)
	GetPriceInternal     func() float64
	SetPriceInternal     func(price float64)
	entity.EntityInterface
}

func (so *StockOrder) GetIsBuy() bool {
	return so.GetIsBuyInternal()
}

func (so *StockOrder) SetIsBuy(isBuy bool) {
	so.SetIsBuyInternal(isBuy)
}

func (so *StockOrder) GetOrderType() string {
	return so.GetOrderTypeInternal()
}

func (so *StockOrder) GetQuantity() int {
	return so.GetQuantityInternal()
}

func (so *StockOrder) SetQuantity(quantity int) {
	so.SetQuantityInternal(quantity)
}

func (so *StockOrder) GetPrice() float64 {
	return so.GetPriceInternal()
}

func (so *StockOrder) SetPrice(price float64) {
	so.SetPriceInternal(price)
}

func (so *StockOrder) GetStockID() string {
	return so.GetStockIDInternal()
}

func (so *StockOrder) SetStockID(stockID string) {
	so.SetStockIDInternal(stockID)
}

type NewStockOrderParams struct {
	entity.NewEntityParams
	StockID   string               `json:"StockID"` // use this or Stock
	Stock     stock.StockInterface // use this or StockID
	IsBuy     bool                 `json:"IsBuy"`
	OrderType string               `json:"OrderType"` // MARKET or LIMIT. This can't be changed later.
	Quantity  int                  `json:"Quantity"`
	Price     float64              `json:"Price"`
}

func NewStockOrder(params NewStockOrderParams) *StockOrder {
	e := entity.NewEntity(params.NewEntityParams)
	var stockID string
	if params.Stock != nil {
		stockID = params.Stock.GetId()
	} else {
		stockID = params.StockID
	}

	sob := &StockOrder{
		StockID:         stockID,
		IsBuy:           params.IsBuy,
		OrderType:       params.OrderType,
		Quantity:        params.Quantity,
		Price:           params.Price,
		EntityInterface: e,
	}
	sob.GetStockIDInternal = func() string { return sob.StockID }
	sob.SetStockIDInternal = func(stockID string) { sob.StockID = stockID }
	sob.GetIsBuyInternal = func() bool { return sob.IsBuy }
	sob.SetIsBuyInternal = func(isBuy bool) { sob.IsBuy = isBuy }
	sob.GetOrderTypeInternal = func() string { return sob.OrderType }
	sob.GetQuantityInternal = func() int { return sob.Quantity }
	sob.SetQuantityInternal = func(quantity int) { sob.Quantity = quantity }
	sob.GetPriceInternal = func() float64 { return sob.Price }
	sob.SetPriceInternal = func(price float64) { sob.Price = price }
	return sob
}

func ParseStockOrder(jsonBytes []byte) (*StockOrder, error) {
	var so NewStockOrderParams
	if err := json.Unmarshal(jsonBytes, &so); err != nil {
		return nil, err
	}
	return NewStockOrder(so), nil
}

func (so *StockOrder) StockOrderToParams() NewStockOrderParams {
	return NewStockOrderParams{
		NewEntityParams: so.EntityToParams(),
		StockID:         so.GetStockID(),
		IsBuy:           so.GetIsBuy(),
		OrderType:       so.GetOrderType(),
		Quantity:        so.GetQuantity(),
		Price:           so.GetPrice(),
	}
}

func (so *StockOrder) StockOrderToJSON() ([]byte, error) {
	return json.Marshal(so.StockOrderToParams())
}
