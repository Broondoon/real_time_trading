package databaseAccessStockOrder

import (
	databaseAccess "Shared/database/database-access"
	"Shared/entities/order"
	databaseServiceStockOrder "databaseServiceStockOrder/database-connection"
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
	*databaseAccess.NewEntityDataAccessParams[*order.StockOrder]
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	if params.NewEntityDataAccessParams == nil {
		params.NewEntityDataAccessParams = &databaseAccess.NewEntityDataAccessParams[*order.StockOrder]{
			NewDatabaseAccessParams: &databaseAccess.NewDatabaseAccessParams{},
		}
	}

	//This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"
	var tempDatabaseService databaseServiceStockOrder.DatabaseServiceInterface
	if params.NewEntityDataAccessParams.EntityDataServiceTemp == nil {
		tempDatabaseService = databaseServiceStockOrder.NewDatabaseService(databaseServiceStockOrder.NewDatabaseServiceParams{})
		params.NewEntityDataAccessParams.EntityDataServiceTemp = tempDatabaseService
	} else {
		tempDatabaseService = params.NewEntityDataAccessParams.EntityDataServiceTemp.(databaseServiceStockOrder.DatabaseServiceInterface)
	}

	dba := &DatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccess[*order.StockOrder, order.StockOrderInterface](params.NewEntityDataAccessParams),
		TEMPCONNECTION:            tempDatabaseService,
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
		convertedStockOrders[i] = o
	}
	return &convertedStockOrders
}
