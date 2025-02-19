package databaseAccessStockTransaction

import (
	databaseAccess "Shared/database/database-access"
	"Shared/database/database-service"
	"Shared/entities/transaction"
	"databaseServiceStockTransaction"
)

type EntityDataAccessInterface = databaseAccess.EntityDataAccessInterface[*transaction.StockTransaction, transaction.StockTransactionInterface]

type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	EntityDataAccessInterface
}

type DatabaseAccess struct {
	EntityDataAccessInterface
	TEMPCONNECTION databaseServiceStockTransaction.DatabaseServiceInterface
}

type NewDatabaseAccessParams struct {
	*databaseAccess.NewDatabaseAccessParams
	*databaseAccess.NewEntityDataAccessParams[*transaction.StockTransaction]
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	//This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"
	dataServiceTemp := databaseServiceStockTransaction.NewDatabaseService(databaseServiceStockTransaction.NewDatabaseServiceParams{
		NewPostGresDatabaseParams: &database.NewPostGresDatabaseParams{
			NewBaseDatabaseParams: &database.NewBaseDatabaseParams{
				DATABASE_URL_ENV_OVERRIDE: "DATABASE_URL_STOCK_TRANSACTIONS",
			},
		},
	})

	dba := &DatabaseAccess{
		EntityDataAccessInterface: databaseAccess.NewEntityDataAccess[*transaction.StockTransaction, transaction.StockTransactionInterface](&databaseAccess.NewEntityDataAccessParams[*transaction.StockTransaction]{
			NewDatabaseAccessParams: params.NewDatabaseAccessParams,
			EntityDataServiceTemp:   dataServiceTemp, //This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"
		}),
	}
	dba.Connect()
	return dba
}
