package order

import (
	"Shared/entities/entity"
	"Shared/entities/stock"
	"encoding/json"
	"strconv"
)

const (
	OrderTypeMarket = "MARKET"
	OrderTypeLimit  = "LIMIT"
)

// // Set here so we can make sure we keep the price as something usable as both math and a key.
// // effectivly we should only be storing it as a string, but capable of processing it as a float64 or int64
// type PriceTypes interface {
// 	float64 | string | int64
// }

// type Price struct {
// 	Price            string                  `json:"Price" gorm:"not null"`
// 	GetPriceInternal func() string           `gorm:"-"`
// 	SetPriceInternal func(price interface{}) `gorm:"-"`
// }

type StockOrderInterface interface {
	GetStockID() string
	SetStockID(stockID string)
	GetIsBuy() bool
	SetIsBuy(isBuy bool)
	GetOrderType() string
	GetQuantity() int
	UpdateQuantity(quantityToAdd int)
	GetPrice() float64
	SetPrice(price float64)
	GetParentStockOrderID() string
	SetParentStockOrderID(parentStockOrderID string)
	GetUserID() string
	SetUserID(userID string)
	CreateChildOrder(parent StockOrderInterface, partner StockOrderInterface) StockOrderInterface
	ToParams() NewStockOrderParams
	entity.EntityInterface
}

type StockOrder struct {
	StockID            string  `json:"stock_id" gorm:"not null"` // use this or Stock
	ParentStockOrderID string  `json:"ParentStockOrderID"`
	IsBuy              bool    `json:"is_buy" gorm:"not null"`
	OrderType          string  `json:"order_type" gorm:"not null"` // MARKET or LIMIT. This can't be changed later.
	Quantity           int     `json:"quantity" gorm:"not null"`
	Price              float64 `json:"price" gorm:"not null"`
	UserID             string  `json:"user_id" gorm:"not null"`
	entity.Entity      `json:"Entity" gorm:"embedded"`
}

func (so *StockOrder) GetIsBuy() bool {
	return so.IsBuy
}

func (so *StockOrder) SetIsBuy(isBuy bool) {
	so.IsBuy = isBuy
	so.Updates = append(so.Updates, &entity.EntityUpdateData{
		ID:         so.GetId(),
		Field:      "IsBuy",
		AlterValue: func() *string { s := strconv.FormatBool(isBuy); return &s }(),
	})
}

func (so *StockOrder) GetOrderType() string {
	return so.OrderType
}

func (so *StockOrder) GetQuantity() int {
	return so.Quantity
}

func (so *StockOrder) UpdateQuantity(quantityToAdd int) {
	so.Quantity += quantityToAdd
	so.Updates = append(so.Updates, &entity.EntityUpdateData{
		ID:         so.GetId(),
		Field:      "Quantity",
		AlterValue: func() *string { s := strconv.Itoa(quantityToAdd); return &s }(),
	})
}

func (so *StockOrder) GetPrice() float64 {
	return so.Price
}
func (so *StockOrder) SetPrice(price float64) {
	so.Price = price
	so.Updates = append(so.Updates, &entity.EntityUpdateData{
		ID:         so.GetId(),
		Field:      "Price",
		AlterValue: func() *string { s := strconv.FormatFloat(price, 'f', -1, 64); return &s }(),
	})
}

func (so *StockOrder) GetStockID() string {
	return so.StockID
}

func (so *StockOrder) SetStockID(stockID string) {
	so.StockID = stockID
}

func (so *StockOrder) GetParentStockOrderID() string {
	return so.ParentStockOrderID
}

func (so *StockOrder) SetParentStockOrderID(parentStockOrderID string) {
	so.ParentStockOrderID = parentStockOrderID
}

func (so *StockOrder) GetUserID() string {
	return so.UserID
}

func (so *StockOrder) SetUserID(userID string) {
	so.UserID = userID
}

func (so *StockOrder) CreateChildOrder(parent StockOrderInterface, partner StockOrderInterface) StockOrderInterface {
	// Create a new Stock Order
	return New(NewStockOrderParams{
		NewEntityParams: entity.NewEntityParams{
			ID: parent.GetId(),
		},
		StockID:            parent.GetStockID(),
		Quantity:           partner.GetQuantity(),
		Price:              parent.GetPrice(),
		OrderType:          parent.GetOrderType(),
		IsBuy:              parent.GetIsBuy(),
		ParentStockOrderID: parent.GetId(),
		UserID:             parent.GetUserID(),
	})

}

type NewStockOrderParams struct {
	entity.NewEntityParams `json:"Entity"`
	Stock                  stock.StockInterface // use this or StockID
	StockID                string               `json:"stock_id"`
	IsBuy                  bool                 `json:"is_buy"`
	OrderType              string               `json:"order_type"` // MARKET or LIMIT. This can't be changed later.
	Quantity               int                  `json:"quantity"`
	Price                  float64              `json:"price"`
	ParentStockOrderID     string               `json:"ParentStockOrderID"`
	UserID                 string               `json:"user_id"`
}

func New(params NewStockOrderParams) *StockOrder {
	e := entity.NewEntity(params.NewEntityParams)
	var stockID string
	if params.Stock != nil {
		stockID = params.Stock.GetId()
	} else {
		stockID = params.StockID
	}

	so := &StockOrder{
		Entity:             *e,
		StockID:            stockID,
		IsBuy:              params.IsBuy,
		OrderType:          params.OrderType,
		Quantity:           params.Quantity,
		Price:              params.Price,
		ParentStockOrderID: params.ParentStockOrderID,
		UserID:             params.UserID,
	}
	return so
}

func Parse(jsonBytes []byte) (*StockOrder, error) {
	var so NewStockOrderParams
	if err := json.Unmarshal(jsonBytes, &so); err != nil {
		return nil, err
	}
	return New(so), nil
}

func ParseList(jsonBytes []byte) (*[]*StockOrder, error) {
	var so []NewStockOrderParams
	if err := json.Unmarshal(jsonBytes, &so); err != nil {
		return nil, err
	}
	soList := make([]*StockOrder, len(so))
	for i, s := range so {
		soList[i] = New(s)
	}
	return &soList, nil
}

func (so *StockOrder) ToParams() NewStockOrderParams {
	return NewStockOrderParams{
		NewEntityParams: so.EntityToParams(),
		StockID:         so.GetStockID(),
		IsBuy:           so.GetIsBuy(),
		OrderType:       so.GetOrderType(),
		Quantity:        so.GetQuantity(),
		Price:           so.GetPrice(),
		UserID:          so.GetUserID(),
	}
}

func (so *StockOrder) ToJSON() ([]byte, error) {
	return json.Marshal(so.ToParams())
}
