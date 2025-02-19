package databaseServiceStockTransaction

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/transaction"
)

type DatabaseServiceInterface interface {
	databaseService.EntityDataInterface[*transaction.StockTransaction]
}

type DatabaseService struct {
	databaseService.EntityDataInterface[*transaction.StockTransaction]
	StockTransactions *[]transaction.StockTransactionInterface
}

type NewDatabaseServiceParams struct {
	*databaseService.NewPostGresDatabaseParams
}

func NewDatabaseService(params NewDatabaseServiceParams) DatabaseServiceInterface {
	db := &DatabaseService{
		EntityDataInterface: databaseService.NewEntityData[*transaction.StockTransaction](params.NewPostGresDatabaseParams),
	}
	db.Connect()
	db.GetDatabaseSession().AutoMigrate(&transaction.StockTransaction{})
	return db
}

func (d *DatabaseService) Connect() {
	d.EntityDataInterface.Connect()
}

func (d *DatabaseService) Disconnect() {
	d.EntityDataInterface.Disconnect()
}
