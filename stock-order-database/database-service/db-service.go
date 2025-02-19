package databaseServiceStockOrder

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/order"
)

type DatabaseServiceInterface interface {
	databaseService.EntityDataInterface[*order.StockOrder]
	GetInitialStockOrdersForStock(stockID string) (*[]order.StockOrder, error)
}

type DatabaseService struct {
	databaseService.EntityDataInterface[*order.StockOrder]
	StockOrders *[]order.StockOrderInterface
}

type NewDatabaseServiceParams struct {
	*databaseService.NewPostGresDatabaseParams
}

func NewDatabaseService(params NewDatabaseServiceParams) DatabaseServiceInterface {
	db := &DatabaseService{
		EntityDataInterface: databaseService.NewEntityData[*order.StockOrder](params.NewPostGresDatabaseParams),
	}
	db.Connect()
	db.GetDatabaseSession().AutoMigrate(&order.StockOrder{})
	return db
}

func (d *DatabaseService) Connect() {
	d.EntityDataInterface.Connect()
}

func (d *DatabaseService) Disconnect() {
	d.EntityDataInterface.Disconnect()
}

// Right now, we're just gonna get all stocksOrders for a given stock. Later, we need to limit this to a specific subset of orders.
func (d *DatabaseService) GetInitialStockOrdersForStock(stockID string) (*[]order.StockOrder, error) {
	var orders []order.StockOrder
	d.GetDatabaseSession().Find(&orders, "StockID = ? ", stockID)
	for _, o := range orders {
		o.SetDefaults()
	}
	return &orders, nil
}
