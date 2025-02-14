package stock

import (
	"Shared/entities/entity"
	"encoding/json"
)

type StockInterface interface {
	GetName() string
	SetName(name string)
	StockToParams() NewStockParams
	StockToJSON() ([]byte, error)
	entity.EntityInterface
}

type Stock struct {
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	Name string
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the Stock Interface.
	GetNameInternal func() string
	SetNameInternal func(name string)
	entity.EntityInterface
}

func (s *Stock) GetName() string {
	return s.GetNameInternal()
}

func (s *Stock) SetName(name string) {
	s.SetNameInternal(name)
}

type NewStockParams struct {
	entity.NewEntityParams
	Name string `json:"Name"`
}

func NewStock(params NewStockParams) *Stock {
	e := entity.NewEntity(params.NewEntityParams)
	s := &Stock{
		Name:            params.Name,
		EntityInterface: e,
	}
	s.GetNameInternal = func() string { return s.Name }
	s.SetNameInternal = func(name string) { s.Name = name }
	return s
}

func ParseStock(jsonBytes []byte) (*Stock, error) {
	var s NewStockParams
	if err := json.Unmarshal(jsonBytes, &s); err != nil {
		return nil, err
	}
	return NewStock(s), nil
}

func (s *Stock) StockToParams() NewStockParams {
	return NewStockParams{
		Name:            s.GetName(),
		NewEntityParams: s.EntityToParams(),
	}
}

func (s *Stock) StockToJSON() ([]byte, error) {
	return json.Marshal(s.StockToParams())
}

type FakeStock struct {
	entity.FakeEntity
	Name string `json:"name"`
}

func (fs *FakeStock) GetName() string               { return fs.Name }
func (fs *FakeStock) SetName(name string)           { fs.Name = name }
func (fs *FakeStock) StockToParams() NewStockParams { return NewStockParams{} }
func (fs *FakeStock) StockToJSON() ([]byte, error)  { return []byte{}, nil }
