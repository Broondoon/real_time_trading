package databaseServiceTemp

import (
	databaseService "Shared/database/database-service"
	"Shared/entities/order"
)

type TempDatabaseServiceInterface interface {
	databaseService.PostGresDatabaseInterface
	GetStockOrder() order.StockOrderInterface
}

type TempDatabaseService struct {
	databaseService.PostGresDatabaseInterface
	stockOrders []order.StockOrderInterface
}

type NewTempDatabaseServiceParams struct {
	databaseService.NewPostGresDatabaseParams
}

func NewTempDatabaseService(params NewTempDatabaseServiceParams) *TempDatabaseService {
	//This would actually go in the proper main. Since however we're currently just testing the database, we'll put it here.
	params.NewPostGresDatabaseParams.Db_ENV_PATH = "DATABASE_URL"
	return &TempDatabaseService{
		PostGresDatabaseInterface: databaseService.NewPostGresDatabase(params.NewPostGresDatabaseParams),
	}
}

func (d *TempDatabaseService) GetStockOrder(orderID string) order.StockOrderInterface {
	return d.GetDatabaseConnection().Where("id = ?", orderID).First(&order.StockOrder{})
}
