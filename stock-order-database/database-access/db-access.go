package databaseAccessStockOrder

import (
	databaseAccess "Shared/database/database-access"
	"Shared/database/database-service"
	"Shared/entities/order"
	"databaseServiceStockOrder"
)

type EntityDataAccessInterface = databaseAccess.EntityDataAccessInterface[*order.StockOrder, order.StockOrderInterface]

type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	EntityDataAccessInterface
	GetInitialStockOrdersForStock(stockID string) *[]order.StockOrderInterface
}

type DatabaseAccess struct {
	EntityDataAccessInterface
	TEMPCONNECTION databaseServiceStockOrder.DatabaseServiceInterface
}

type NewDatabaseAccessParams struct {
	*databaseAccess.NewDatabaseAccessParams
	*databaseAccess.NewEntityDataAccessParams[*order.StockOrder]
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	//This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"
	dataServiceTemp := databaseServiceStockOrder.NewDatabaseService(databaseServiceStockOrder.NewDatabaseServiceParams{
		NewPostGresDatabaseParams: &database.NewPostGresDatabaseParams{
			NewBaseDatabaseParams: &database.NewBaseDatabaseParams{},
		},
	})

	dba := &DatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccess[*order.StockOrder, order.StockOrderInterface](&databaseAccess.NewEntityDataAccessParams[*order.StockOrder]{
			NewDatabaseAccessParams: params.NewDatabaseAccessParams,
			EntityDataServiceTemp:   dataServiceTemp, //This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"
		}),
	}
	dba.Connect()
	return dba
}

func (d *DatabaseAccess) GetInitialStockOrdersForStock(stockID string) *[]order.StockOrderInterface {
	stockOrders, err := d.TEMPCONNECTION.GetInitialStockOrdersForStock(stockID)
	if err != nil {
		return nil
	}
	convertedStockOrders := make([]order.StockOrderInterface, len(*stockOrders))
	for i, o := range *stockOrders {
		convertedStockOrders[i] = &o
	}
	return &convertedStockOrders
}
