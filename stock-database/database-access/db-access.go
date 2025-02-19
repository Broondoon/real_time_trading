package databaseAccessStock

import (
	databaseAccess "Shared/database/database-access"
	"Shared/entities/stock"
	"Shared/network"
	"os"
)

type EntityDataAccessInterface = databaseAccess.EntityDataAccessInterface[*stock.Stock, stock.StockInterface]

type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	GetStockIDs() *[]string
}

type DatabaseAccess struct {
	EntityDataAccessInterface
	_networkManager network.NetworkInterface
}

type NewDatabaseAccessParams struct {
	*databaseAccess.NewEntityDataAccessHTTPParams[*stock.Stock]
	network network.NetworkInterface
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	if params.NewEntityDataAccessHTTPParams == nil {
		params.NewEntityDataAccessHTTPParams = &databaseAccess.NewEntityDataAccessHTTPParams[*stock.Stock]{}
	}

	if params.network == nil {
		panic("No network provided")
	}
	if params.NewEntityDataAccessHTTPParams.Client == nil {
		params.NewEntityDataAccessHTTPParams.Client = params.network.Stocks()
	}
	if params.NewEntityDataAccessHTTPParams.DefaultRoute == "" {
		params.NewEntityDataAccessHTTPParams.DefaultRoute = os.Getenv("STOCK_DATABASE_SERVICE_ROUTE")
	}

	dba := &DatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccessHTTP[*stock.Stock, stock.StockInterface](params.NewEntityDataAccessHTTPParams),
		_networkManager:           params.network,
	}
	dba.Connect()
	return dba
}

func (d *DatabaseAccess) GetStockIDs() *[]string {
	stocks := d.GetAll()
	stockIDs := make([]string, len(*stocks))
	for i, stock := range *stocks {
		stockIDs[i] = stock.GetId()
	}
	return &stockIDs
}
