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
	*databaseService.NewPostGresDatabaseParams
}

func NewDatabaseService(params *NewDatabaseServiceParams) DatabaseServiceInterface {
	db := &DatabaseService{
		EntityDataInterface: databaseService.NewEntityData[*stock.Stock](params.NewPostGresDatabaseParams),
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
