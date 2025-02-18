package databaseAccessStockOrder

import (
	databaseAccess "Shared/database/database-access"
	databaseService "Shared/database/database-service"
	"Shared/entities/order"
	databaseServiceStockOrder "databaseServiceStockOrder"
)

type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	GetStockOrder(orderID string) order.StockOrderInterface
	GetStockOrders(orderIDs *[]string) *[]order.StockOrderInterface
	GetInitialStockOrdersForStock(stockID string) *[]order.StockOrderInterface
	CreateStockOrder(order order.StockOrderInterface) order.StockOrderInterface
	UpdateStockOrder(order order.StockOrderInterface) order.StockOrderInterface
	DeleteStockOrder(orderID string) order.StockOrderInterface
}

type DatabaseAccess struct {
	databaseAccess.BaseDatabaseAccessInterface
	databaseTEMP databaseServiceStockOrder.DatabaseServiceInterface
}

type NewDatabaseAccessParams struct {
	*databaseAccess.NewDatabaseAccessParams
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	dba := &DatabaseAccess{
		BaseDatabaseAccessInterface: databaseAccess.NewBaseDatabaseAccess(params.NewDatabaseAccessParams)}
	dba.Connect()
	return dba
}

func (d *DatabaseAccess) Connect() {
	//This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"
	//Cheap ignore of sepearation between access and database. Later on, we'd actually likely have a cache between here and the database, but for now, we'll just connect directly.
	//This would actually go in the proper main of the database. Since however we're currently just testing the database, we'll put it here.
	dbParams := databaseServiceStockOrder.NewDatabaseServiceParams{
		NewPostGresDatabaseParams: &databaseService.NewPostGresDatabaseParams{
			NewBaseDatabaseParams: &databaseService.NewBaseDatabaseParams{},
		},
	}
	d.databaseTEMP = databaseServiceStockOrder.NewDatabaseService(dbParams)
}

func (d *DatabaseAccess) Disconnect() {
	d.databaseTEMP.Disconnect()
}

// Dirty methods for database connection.
func (d *DatabaseAccess) GetStockOrder(orderID string) order.StockOrderInterface {
	stockOrder, err := d.databaseTEMP.GetStockOrder(orderID)
	if err != nil {
		return nil
	}
	return stockOrder

}

func (d *DatabaseAccess) GetStockOrders(orderIDs *[]string) *[]order.StockOrderInterface {
	stockOrders, err := d.databaseTEMP.GetStockOrders(orderIDs)
	if err != nil {
		return nil
	}
	return stockOrders
}

func (d *DatabaseAccess) GetInitialStockOrdersForStock(stockID string) *[]order.StockOrderInterface {
	stockOrders, err := d.databaseTEMP.GetInitialStockOrdersForStock(stockID)
	if err != nil {
		return nil
	}
	return stockOrders
}

func (d *DatabaseAccess) CreateStockOrder(order order.StockOrderInterface) order.StockOrderInterface {
	stockOrder, err := d.databaseTEMP.CreateStockOrder(order)
	if err != nil {
		return nil
	}
	return stockOrder
}

func (d *DatabaseAccess) UpdateStockOrder(order order.StockOrderInterface) order.StockOrderInterface {
	stockOrder, err := d.databaseTEMP.UpdateStockOrder(order)
	if err != nil {
		return nil
	}
	return stockOrder
}

func (d *DatabaseAccess) DeleteStockOrder(orderID string) order.StockOrderInterface {
	stockOrder, err := d.databaseTEMP.DeleteStockOrder(orderID)
	if err != nil {
		return nil
	}
	return stockOrder
}
