package stock

import (
	"Shared/entities/entity"
)

type StockInterface interface {
	GetName() string
	SetName(name string)
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
	Name string
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
