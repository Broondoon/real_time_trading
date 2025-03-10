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
	*databaseService.NewEntityDataParams
}

func NewDatabaseService(params NewDatabaseServiceParams) DatabaseServiceInterface {
	if params.NewEntityDataParams == nil {
		params.NewEntityDataParams = &databaseService.NewEntityDataParams{}
	}

	//CACHE IMPLEMENTATION
	//THIS ONE IS DISABLED BECAUSE AS OF MARCH 9/25 THIS DOES NOT IMPROVE PERFORMANCE DUE TO STOCKORDER NOT BEING
	//STOCKORDER BEING CREATED BUT NOT RETRIEVED FROM THE CACHE
	/* cachedStockOrder := databaseService.NewCachedEntityData[*order.StockOrder](&databaseService.NewCachedEntityDataParams{
		NewEntityDataParams: params.NewEntityDataParams,
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		Password:            os.Getenv("REDIS_PASSWORD"),
		DefaultTTL:          5 * time.Minute,
	})

	db := &DatabaseService{
		EntityDataInterface: cachedStockOrder,
	} */

	db := &DatabaseService{
		EntityDataInterface: databaseService.NewEntityData[*order.StockOrder](params.NewEntityDataParams),
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
	d.GetDatabaseSession().Find(&orders, "stock_id = ? ", stockID)
	return &orders, nil
}
