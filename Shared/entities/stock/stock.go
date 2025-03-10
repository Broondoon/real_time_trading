package stock

import (
	"Shared/entities/entity"
	"encoding/json"
)

type StockInterface interface {
	GetName() string
	SetName(name string)
	ToParams() NewStockParams
	entity.EntityInterface
}

type Stock struct {
	Name          string `json:"stock_name" gorm:"not null"`
	entity.Entity `json:"Entity" gorm:"embedded"`
}

func (s *Stock) GetName() string {
	return s.Name
}

func (s *Stock) SetName(name string) {
	s.Name = name
}

type NewStockParams struct {
	entity.NewEntityParams `json:"Entity"`
	Name                   string `json:"stock_name"`
}

func New(params NewStockParams) *Stock {
	e := entity.NewEntity(params.NewEntityParams)
	s := &Stock{
		Name:   params.Name,
		Entity: *e,
	}
	return s
}

func Parse(jsonBytes []byte) (*Stock, error) {
	var s NewStockParams
	if err := json.Unmarshal(jsonBytes, &s); err != nil {
		return nil, err
	}
	return New(s), nil
}

func ParseList(jsonBytes []byte) (*[]*Stock, error) {
	var so []NewStockParams
	if err := json.Unmarshal(jsonBytes, &so); err != nil {
		return nil, err
	}
	soList := make([]*Stock, len(so))
	for i, s := range so {
		soList[i] = New(s)
	}
	return &soList, nil
}

func (s *Stock) ToParams() NewStockParams {
	return NewStockParams{
		NewEntityParams: s.EntityToParams(),
		Name:            s.GetName(),
	}
}

func (s *Stock) ToJSON() ([]byte, error) {
	return json.Marshal(s.ToParams())
}
