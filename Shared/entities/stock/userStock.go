package stock

import (
	"Shared/entities/entity"
	"Shared/entities/user"
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
	entity.EntityInterface
}

type UserStock struct {
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	UserID    string
	StockID   string
	StockName string
	Quantity  int
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the UserStock Interface.
	GetUserIDInternal    func() string
	SetUserIDInternal    func(userID string)
	GetStockIDInternal   func() string
	SetStockIDInternal   func(stockID string)
	GetStockNameInternal func() string
	SetStockNameInternal func(stockName string)
	GetQuantityInternal  func() int
	SetQuantityInternal  func(quantity int)
	entity.EntityInterface
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
	UserID    string             // use this or User
	User      user.UserInterface // use this or UserID
	StockID   string             // use this or Stock
	StockName string             // use this or Stock
	Stock     StockInterface     // use this or StockID and StockName
	Quantity  int
}

func NewUserStock(params NewUserStockParams) *UserStock {
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
		UserID:          userId,
		StockID:         stockId,
		StockName:       stockName,
		Quantity:        params.Quantity,
		EntityInterface: e,
	}
	us.GetQuantityInternal = func() int { return us.Quantity }
	us.SetQuantityInternal = func(quantity int) { us.Quantity = quantity }
	us.GetUserIDInternal = func() string { return us.UserID }
	us.SetUserIDInternal = func(userID string) { us.UserID = userID }
	us.GetStockIDInternal = func() string { return us.StockID }
	us.SetStockIDInternal = func(stockID string) { us.StockID = stockID }
	us.GetStockNameInternal = func() string { return us.StockName }
	us.SetStockNameInternal = func(stockName string) { us.StockName = stockName }
	return us
}
