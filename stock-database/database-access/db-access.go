package databaseAccessStock

import (
	databaseAccess "Shared/database/database-access"
	"Shared/entities/stock"
	"Shared/network"
	"os"

	"github.com/google/uuid"
)

type EntityDataAccessInterface = databaseAccess.EntityDataAccessInterface[*stock.Stock, stock.StockInterface]

type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	EntityDataAccessInterface
	GetStockIDs() (*[]*uuid.UUID, error)
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

func (d *DatabaseAccess) GetStockIDs() (*[]*uuid.UUID, error) {
	stocks, err := d.GetAll()
	stockIDs := make([]*uuid.UUID, len(*stocks))
	for i, stock := range *stocks {
		stockIDs[i] = stock.GetId()
	}
	return &stockIDs, err
}
