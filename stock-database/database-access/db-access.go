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
	AddNewStock(stock *stock.Stock) (stock.StockInterface, error)
}

type DatabaseAccess struct {
	EntityDataAccessInterface
	_networkManager network.NetworkInterface
}

type NewDatabaseAccessParams struct {
	*databaseAccess.NewEntityDataAccessNetworkParams[*stock.Stock]
	Network network.NetworkInterface
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	if params.NewEntityDataAccessNetworkParams == nil {
		params.NewEntityDataAccessNetworkParams = &databaseAccess.NewEntityDataAccessNetworkParams[*stock.Stock]{}
	}

	if params.Network == nil {
		panic("No network provided")
	}
	if params.NewEntityDataAccessNetworkParams.Client == nil {
		params.NewEntityDataAccessNetworkParams.Client = params.Network.Stocks()
	}
	if params.NewEntityDataAccessNetworkParams.DefaultRoute == "" {
		params.NewEntityDataAccessNetworkParams.DefaultRoute = os.Getenv("STOCK_DATABASE_SERVICE_ROUTE")
	}
	if params.NewEntityDataAccessNetworkParams.Parser == nil {
		params.NewEntityDataAccessNetworkParams.Parser = stock.Parse
	}
	if params.NewEntityDataAccessNetworkParams.ParserList == nil {
		params.NewEntityDataAccessNetworkParams.ParserList = stock.ParseList
	}

	dba := &DatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccessNetwork[*stock.Stock, stock.StockInterface](params.NewEntityDataAccessNetworkParams),
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

func (d *DatabaseAccess) AddNewStock(stockVal *stock.Stock) (stock.StockInterface, error) {
	json, err := d._networkManager.Stocks().Post("createStock", stockVal)
	if err != nil {
		return nil, err
	}
	stockVal, err = stock.Parse(json)
	if err != nil {
		return nil, err
	}
	return stockVal, nil

}
