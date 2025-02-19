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
	Name string `json:"Name" gorm:"not null"`
	// If you need to access a property, please use the Get and Set functions, not the property itself. It is only exposed in case you need to interact with it when altering internal functions.
	// Internal Functions should not be interacted with directly. if you need to change functionality, set a new function to the existing internal function.
	// Instead, interact with the functions through the Stock Interface.
	// GetNameInternal func() string     `gorm:"-"`
	// SetNameInternal func(name string) `gorm:"-"`
	entity.Entity `json:"Entity" gorm:"embedded"`
}

func (s *Stock) GetName() string {
	// return s.GetNameInternal()
	return s.Name
}

func (s *Stock) SetName(name string) {
	// s.SetNameInternal(name)
	s.Name = name
}

type NewStockParams struct {
	entity.NewEntityParams
	Name string `json:"Name"`
}

func New(params NewStockParams) *Stock {
	e := entity.NewEntity(params.NewEntityParams)
	s := &Stock{
		Name:   params.Name,
		Entity: *e,
	}
	s.SetDefaults()
	return s
}

func (s *Stock) SetDefaults() {
	// s.GetNameInternal = func() string { return s.Name }
	// s.SetNameInternal = func(name string) { s.Name = name }
}

func Parse(jsonBytes []byte) (*Stock, error) {
	var s NewStockParams
	if err := json.Unmarshal(jsonBytes, &s); err != nil {
		return nil, err
	}
	return New(s), nil
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

// FakeStock is a fake stock mock for testing purposes
type FakeStock struct {
	entity.FakeEntity
	Name string `json:"name"`
}

func (fs *FakeStock) GetName() string          { return fs.Name }
func (fs *FakeStock) SetName(name string)      { fs.Name = name }
func (fs *FakeStock) ToParams() NewStockParams { return NewStockParams{} }
func (fs *FakeStock) ToJSON() ([]byte, error)  { return []byte{}, nil }
