package databaseServiceStock

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/stock"
	"os"
	"time"

	"gorm.io/gorm"
)

type DatabaseServiceInterface interface {
	databaseService.EntityDataInterface[*stock.Stock, *gorm.DB]
}

type DatabaseService struct {
	databaseService.EntityDataInterface[*stock.Stock, *gorm.DB]
	Stocks *[]stock.StockInterface
}

type NewDatabaseServiceParams struct {
}

func NewDatabaseService(params *NewDatabaseServiceParams) DatabaseServiceInterface {
	dbConnection := databaseService.NewPostGresDatabase(&databaseService.NewPostGresDatabaseParams{})

	cachedStock := databaseService.NewCachedEntityData[*stock.Stock, *gorm.DB](&databaseService.NewCachedEntityDataParams[*stock.Stock, *gorm.DB]{
		RedisAddr:  os.Getenv("REDIS_STOCK_ADDR"),
		Password:   os.Getenv("REDIS_PASSWORD"),
		DefaultTTL: 5 * time.Minute,
		EntityData: databaseService.NewPostGresEntityData[*stock.Stock](&databaseService.NewPostGresEntityDataParams{
			Existing: dbConnection,
		}),
	})

	db := &DatabaseService{
		EntityDataInterface: cachedStock,
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
