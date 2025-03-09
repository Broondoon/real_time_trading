package userStock

import (
	"Shared/entities/entity"
	"Shared/entities/stock"
	"Shared/entities/user"
	"encoding/json"
	"strconv"
	"time"
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
	UpdateQuantity(quantityToAdd int)
	SetUpdatedAt(time.Time)
	GetUpdatedAt() time.Time
	ToParams() NewUserStockParams
	entity.EntityInterface
}

type UserStock struct {
	UserID        string `json:"user_id" gorm:"not null"`
	StockID       string `json:"stock_id" gorm:"not null"`
	StockName     string `json:"stock_name" gorm:"not null"`
	Quantity      int    `json:"quantity_owned" gorm:"not null"`
	entity.Entity `json:"Entity" gorm:"embedded"`
}

func (us *UserStock) GetQuantity() int {
	return us.Quantity
}

func (us *UserStock) UpdateQuantity(quantityToAdd int) {
	us.Quantity += quantityToAdd
	*us.Updates = append(*us.Updates, &entity.EntityUpdateData{
		ID:         us.GetId(),
		Field:      "Quantity",
		AlterValue: func() *string { s := strconv.Itoa(quantityToAdd); return &s }(),
	})
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
	*us.Updates = append(*us.Updates, &entity.EntityUpdateData{
		ID:       us.GetId(),
		Field:    "StockName",
		NewValue: &stockName,
	})
}

func (us *UserStock) GetUpdatedAt() time.Time {
	return us.DateModified
}

func (us *UserStock) SetUpdatedAt(updatedAt time.Time) {
	us.SetDateModified(updatedAt)
}

type NewUserStockParams struct {
	entity.NewEntityParams `json:"Entity"`
	UserID                 string               `json:"user_id"`
	StockID                string               `json:"stock_id"`
	StockName              string               `json:"stock_name"`
	Quantity               int                  `json:"quantity_owned"`
	User                   user.UserInterface   // use this or UserID
	Stock                  stock.StockInterface // use this or StockID and StockName
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
