package userStock

import (
	"Shared/entities/entity"
	"Shared/entities/stock"
	"Shared/entities/user"
	"encoding/json"
)

//For tracking stocks owned per user.

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

type UserStockProps struct {
}

type UserStock struct {
	UserID    string `json:"UserID" gorm:"not null"`
	StockID   string `json:"StockID" gorm:"not null"`
	StockName string `json:"StockName" gorm:"not null"`
	Quantity  int    `json:"Quantity" gorm:"not null"`
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the UserStock Interface.
	GetUserIDInternal          func() string          `gorm:"-"`
	SetUserIDInternal          func(userID string)    `gorm:"-"`
	GetStockIDInternal         func() string          `gorm:"-"`
	SetStockIDInternal         func(stockID string)   `gorm:"-"`
	GetStockNameInternal       func() string          `gorm:"-"`
	SetStockNameInternal       func(stockName string) `gorm:"-"`
	GetQuantityInternal        func() int             `gorm:"-"`
	SetQuantityInternal        func(quantity int)     `gorm:"-"`
	entity.BaseEntityInterface `gorm:"embedded"`
}

func (us *UserStock) GetQuantity() int {
	return us.GetQuantityInternal()
}

func (us *UserStock) SetQuantity(quantity int) {
	us.SetQuantityInternal(quantity)
}

func (us *UserStock) GetUserID() string {
	return us.GetUserIDInternal()
}

func (us *UserStock) SetUserID(userID string) {
	us.SetUserIDInternal(userID)
}

func (us *UserStock) GetStockID() string {
	return us.GetStockIDInternal()
}

func (us *UserStock) SetStockID(stockID string) {
	us.SetStockIDInternal(stockID)
}

func (us *UserStock) GetStockName() string {
	return us.GetStockNameInternal()
}

func (us *UserStock) SetStockName(stockName string) {
	us.SetStockNameInternal(stockName)
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
		UserID:              userId,
		StockID:             stockId,
		StockName:           stockName,
		Quantity:            params.Quantity,
		BaseEntityInterface: e,
	}
	us.SetDefaults()
	return us
}

func (us *UserStock) SetDefaults() {
	us.GetQuantityInternal = func() int { return us.Quantity }
	us.SetQuantityInternal = func(quantity int) { us.Quantity = quantity }
	us.GetUserIDInternal = func() string { return us.UserID }
	us.SetUserIDInternal = func(userID string) { us.UserID = userID }
	us.GetStockIDInternal = func() string { return us.StockID }
	us.SetStockIDInternal = func(stockID string) { us.StockID = stockID }
	us.GetStockNameInternal = func() string { return us.StockName }
	us.SetStockNameInternal = func(stockName string) { us.StockName = stockName }
}

func Parse(jsonBytes []byte) (*UserStock, error) {
	var us NewUserStockParams
	if err := json.Unmarshal(jsonBytes, &us); err != nil {
		return nil, err
	}
	return New(us), nil
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
