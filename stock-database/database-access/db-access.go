package databaseAccessStock

import (
	databaseAccess "Shared/database/database-access"
	"Shared/entities/stock"
	"Shared/network"
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
	*databaseAccess.NewDatabaseAccessParams
	network network.NetworkInterface
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	//This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"

	dba := &DatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccessHTTP[*stock.Stock, stock.StockInterface](
			&databaseAccess.NewEntityDataAccessHTTPParams[*stock.Stock]{
				NewDatabaseAccessParams: params.NewDatabaseAccessParams,
				Client:                  params.network.Stocks(),
				PostRoute:               "/createStock",
			}),
		_networkManager: params.network,
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
