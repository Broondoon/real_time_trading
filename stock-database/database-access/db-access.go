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
	EntityDataAccessInterface
	GetStockIDs() (*[]string, error)
}

type DatabaseAccess struct {
	EntityDataAccessInterface
	_networkManager network.NetworkInterface
}

type NewDatabaseAccessParams struct {
	*databaseAccess.NewEntityDataAccessHTTPParams[*stock.Stock]
	Network network.NetworkInterface
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	if params.NewEntityDataAccessHTTPParams == nil {
		params.NewEntityDataAccessHTTPParams = &databaseAccess.NewEntityDataAccessHTTPParams[*stock.Stock]{}
	}

	if params.Network == nil {
		panic("No network provided")
	}
	if params.NewEntityDataAccessHTTPParams.Client == nil {
		params.NewEntityDataAccessHTTPParams.Client = params.Network.Stocks()
	}
	if params.NewEntityDataAccessHTTPParams.DefaultRoute == "" {
		params.NewEntityDataAccessHTTPParams.DefaultRoute = os.Getenv("STOCK_DATABASE_SERVICE_ROUTE")
	}
	if params.NewEntityDataAccessHTTPParams.Parser == nil {
		params.NewEntityDataAccessHTTPParams.Parser = stock.Parse
	}
	if params.NewEntityDataAccessHTTPParams.ParserList == nil {
		params.NewEntityDataAccessHTTPParams.ParserList = stock.ParseList
	}

	dba := &DatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccessHTTP[*stock.Stock, stock.StockInterface](params.NewEntityDataAccessHTTPParams),
		_networkManager:           params.Network,
	}
	dba.Connect()
	return dba
}

func (d *DatabaseAccess) GetStockIDs() (*[]string, error) {
	stocks, err := d.GetAll()
	stockIDs := make([]string, len(*stocks))
	for i, stock := range *stocks {
		stockIDs[i] = stock.GetId()
	}
	return &stockIDs, err
}
