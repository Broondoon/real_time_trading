package userStock

import (
	"Shared/entities/entity"
	"Shared/entities/stock"
	"Shared/entities/user"
	"encoding/json"
)

// For tracking stocks owned per user.

type UserStockInterface interface {
	GetUserID() string
	SetUserID(userID string)
	GetStockID() string
	SetStockID(stockID string)
	GetStockName() string
	SetStockName(stockName string)
	GetQuantity() int
	SetQuantity(quantity int)
	ToParams() NewUserStockParams
	entity.EntityInterface
}

type UserStock struct {
	UserID    string `json:"UserID" gorm:"not null"`
	StockID   string `json:"StockID" gorm:"not null"`
	StockName string `json:"StockName" gorm:"not null"`
	Quantity  int    `json:"Quantity" gorm:"not null"`
	// The following internal functions have been commented out.
	// Instead, we use the fields directly in the getters and setters.
	/*
		GetUserIDInternal    func() string          `gorm:"-"`
		SetUserIDInternal    func(userID string)    `gorm:"-"`
		GetStockIDInternal   func() string          `gorm:"-"`
		SetStockIDInternal   func(stockID string)   `gorm:"-"`
		GetStockNameInternal func() string          `gorm:"-"`
		SetStockNameInternal func(stockName string) `gorm:"-"`
		GetQuantityInternal  func() int             `gorm:"-"`
		SetQuantityInternal  func(quantity int)     `gorm:"-"`
	*/
	entity.Entity `json:"Entity" gorm:"embedded"`
}

func (us *UserStock) GetQuantity() int {
	return us.Quantity
}

func (us *UserStock) SetQuantity(quantity int) {
	us.Quantity = quantity
}

func (us *UserStock) GetUserID() string {
	return us.UserID
}

func (us *UserStock) SetUserID(userID string) {
	us.UserID = userID
}

func (us *UserStock) GetStockID() string {
	return us.StockID
}

func (us *UserStock) SetStockID(stockID string) {
	us.StockID = stockID
}

func (us *UserStock) GetStockName() string {
	return us.StockName
}

func (us *UserStock) SetStockName(stockName string) {
	us.StockName = stockName
}

type NewUserStockParams struct {
	entity.NewEntityParams
	UserID    string               `json:"UserID"`
	StockID   string               `json:"StockID"`
	StockName string               `json:"StockName"`
	Quantity  int                  `json:"Quantity"`
	User      user.UserInterface   // use this or UserID
	Stock     stock.StockInterface // use this or StockID and StockName
}

func New(params NewUserStockParams) *UserStock {
	e := entity.NewEntity(params.NewEntityParams)
	var userId string
	if params.User != nil {
		userId = params.User.GetId()
	} else {
		userId = params.UserID
	}

	var stockId string
	var stockName string
	if params.Stock != nil {
		stockId = params.Stock.GetId()
		stockName = params.Stock.GetName()
	} else {
		stockId = params.StockID
		stockName = params.StockName
	}

	us := &UserStock{
		UserID:    userId,
		StockID:   stockId,
		StockName: stockName,
		Quantity:  params.Quantity,
		Entity:    *e,
	}
	// No need for internal function defaults, using direct field access now.
	return us
}

func Parse(jsonBytes []byte) (*UserStock, error) {
	var us NewUserStockParams
	if err := json.Unmarshal(jsonBytes, &us); err != nil {
		return nil, err
	}
	return New(us), nil
}

func ParseList(jsonBytes []byte) (*[]*UserStock, error) {
	var so []NewUserStockParams
	if err := json.Unmarshal(jsonBytes, &so); err != nil {
		return nil, err
	}
	soList := make([]*UserStock, len(so))
	for i, s := range so {
		soList[i] = New(s)
	}
	return &soList, nil
}

func (us *UserStock) ToParams() NewUserStockParams {
	return NewUserStockParams{
		NewEntityParams: us.EntityToParams(),
		UserID:          us.GetUserID(),
		StockID:         us.GetStockID(),
		StockName:       us.GetStockName(),
		Quantity:        us.GetQuantity(),
	}
}

func (us *UserStock) ToJSON() ([]byte, error) {
	return json.Marshal(us.ToParams())
}

type FakeUserStock struct {
	entity.FakeEntity
	UserID    string `json:"userID"`
	StockID   string `json:"stockID"`
	StockName string `json:"stockName"`
	Quantity  int    `json:"quantity"`
}

func (fus *FakeUserStock) GetUserID() string             { return fus.UserID }
func (fus *FakeUserStock) SetUserID(userID string)       { fus.UserID = userID }
func (fus *FakeUserStock) GetStockID() string            { return fus.StockID }
func (fus *FakeUserStock) SetStockID(stockID string)     { fus.StockID = stockID }
func (fus *FakeUserStock) GetStockName() string          { return fus.StockName }
func (fus *FakeUserStock) SetStockName(stockName string) { fus.StockName = stockName }
func (fus *FakeUserStock) GetQuantity() int              { return fus.Quantity }
func (fus *FakeUserStock) SetQuantity(quantity int)      { fus.Quantity = quantity }
func (fus *FakeUserStock) ToParams() NewUserStockParams  { return NewUserStockParams{} }
func (fus *FakeUserStock) ToJSON() ([]byte, error)       { return []byte{}, nil }
