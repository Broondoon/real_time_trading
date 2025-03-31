package databaseServiceStockOrder

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/order"

	"gorm.io/gorm"
)

type DatabaseServiceInterface interface {
	databaseService.EntityDataInterface[*order.StockOrder, *gorm.DB]
	GetInitialStockOrdersForStock(stockID string) (*[]*order.StockOrder, error)
}

type DatabaseService struct {
	databaseService.EntityDataInterface[*order.StockOrder, *gorm.DB]
	StockOrders *[]order.StockOrderInterface
}

type NewDatabaseServiceParams struct {
}

func NewDatabaseService(params NewDatabaseServiceParams) DatabaseServiceInterface {
	/* 	dbConnection := databaseService.NewPostGresDatabase(&databaseService.NewPostGresDatabaseParams{})

	   	cachedStockOrder := databaseService.NewCachedEntityData[*order.StockOrder, *gorm.DB](&databaseService.NewCachedEntityDataParams[*order.StockOrder, *gorm.DB]{
	   		RedisAddr:  os.Getenv("REDIS_STOCK_ORDER_ADDR"),
	   		Password:   os.Getenv("REDIS_PASSWORD"),
	   		DefaultTTL: 5 * time.Minute,
	   		EntityData: databaseService.NewPostGresEntityData[*order.StockOrder](&databaseService.NewPostGresEntityDataParams{
	   			Existing: dbConnection,
	   		}),
	   	})

	   	db := &DatabaseService{
	   		EntityDataInterface: cachedStockOrder,
	   	} */
	db := &DatabaseService{
		EntityDataInterface: databaseService.NewPostGresEntityData[*order.StockOrder](&databaseService.NewPostGresEntityDataParams{}),
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
func (d *DatabaseService) GetInitialStockOrdersForStock(stockID string) (*[]*order.StockOrder, error) {
	//disabling for now, since we don't have stocks orders in the db before, and it's screwing up sharding.
	//orders, err := d.GetByForeignID("StockID", stockID, "")
	// if err != nil {
	// 	return nil, err
	// }
	//return orders, nil
	emptyList := make([]*order.StockOrder, 0)
	return &emptyList, nil
}
