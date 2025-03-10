package databaseServiceStock

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/stock"
)

type DatabaseServiceInterface interface {
	databaseService.EntityDataInterface[*stock.Stock]
}

type DatabaseService struct {
	databaseService.EntityDataInterface[*stock.Stock]
	Stocks *[]stock.StockInterface
}

type NewDatabaseServiceParams struct {
	*databaseService.NewEntityDataParams //leave nil for default
}

func NewDatabaseService(params *NewDatabaseServiceParams) DatabaseServiceInterface {
	if params.NewEntityDataParams == nil {
		params.NewEntityDataParams = &databaseService.NewEntityDataParams{}
	}

	//CACHE IMPLEMENTATION
	/* cachedStock := databaseService.NewCachedEntityData[*stock.Stock](&databaseService.NewCachedEntityDataParams{
		NewEntityDataParams: params.NewEntityDataParams,
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		Password:            os.Getenv("REDIS_PASSWORD"),
		DefaultTTL:          5 * time.Minute,
	})

	db := &DatabaseService{
		EntityDataInterface: cachedStock,
	} */

	db := &DatabaseService{
		EntityDataInterface: databaseService.NewEntityData[*stock.Stock](params.NewEntityDataParams),
	}

	db.Connect()
	db.GetDatabaseSession().AutoMigrate(&stock.Stock{})
	return db
}

func (d *DatabaseService) Connect() {
	d.EntityDataInterface.Connect()
}

func (d *DatabaseService) Disconnect() {
	d.EntityDataInterface.Disconnect()
}
