package databaseAccessTemp

import (
	databaseAccess "Shared/database/database-access"
	"Shared/entities/order"
)

type TempDatabaseAccessInterface interface {
	databaseAccess.BaseDatabaseAccessInterface
	GetStockOrder() order.StockOrderInterface
}

type TempDatabaseAccess struct {
	databaseAccess.BaseDatabaseAccess
}

type NewTempDatabaseAccessParams struct {
	databaseAccess.NewDatabaseAccessParams
}

func NewSampleDatabaseAccess(params NewTempDatabaseAccessParams) *TempDatabaseAccess {
	return &TempDatabaseAccess{
		BaseDatabaseAccess: *databaseAccess.NewBaseDatabaseAccess(params.NewDatabaseAccessParams)}
}

func (d *SampleDatabaseAccess) GetStockOrder() order.StockOrderInterface {

	return order.NewStockOrder(order.NewStockOrderParams{})
}
