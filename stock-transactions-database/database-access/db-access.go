package databaseAccessStockTransaction

import (
	databaseAccess "Shared/database/database-access"
	databaseService "Shared/database/database-service"
	"Shared/entities/transaction"
	databaseServiceStockTransaction "databaseServiceStockTransaction"
)

type DatabaseAccessInterface interface {
	databaseAccess.DatabaseAccessInterface
	GetStockTransaction(transactionID string) transaction.StockTransactionInterface
	GetStockTransactions(transactionID *[]string) *[]transaction.StockTransactionInterface
	CreateStockTransaction(transactionID transaction.StockTransactionInterface) transaction.StockTransactionInterface
	UpdateStockTransaction(transactionID transaction.StockTransactionInterface) transaction.StockTransactionInterface
	DeleteStockTransaction(transactionID string) transaction.StockTransactionInterface
}

type DatabaseAccess struct {
	databaseAccess.BaseDatabaseAccessInterface
	databaseTEMP databaseServiceStockTransaction.DatabaseServiceInterface
}

type NewDatabaseAccessParams struct {
	*databaseAccess.NewDatabaseAccessParams
}

func NewDatabaseAccess(params *NewDatabaseAccessParams) DatabaseAccessInterface {
	dba := &DatabaseAccess{
		BaseDatabaseAccessInterface: databaseAccess.NewBaseDatabaseAccess(params.NewDatabaseAccessParams)}
	dba.Connect()
	return dba
}

func (d *DatabaseAccess) Connect() {
	//This is our dirty temporary implementation of this. Ideallily, this access has no idea what sort of database setup there is. It just knows "SEND HERE TO GET DATA"
	//Cheap ignore of sepearation between access and database. Later on, we'd actually likely have a cache between here and the database, but for now, we'll just connect directly.
	//This would actually go in the proper main of the database. Since however we're currently just testing the database, we'll put it here.
	dbParams := databaseServiceStockTransaction.NewDatabaseServiceParams{
		NewPostGresDatabaseParams: &databaseService.NewPostGresDatabaseParams{
			NewBaseDatabaseParams: &databaseService.NewBaseDatabaseParams{
				DATABASE_URL_ENV_OVERRIDE: "DATABASE_URL_STOCK_TRANSACTIONS", // Here because we don't have seperate database Service services yet, so everyone is technically creating a new service that connects to the same database. Remove when this is no longer the case.
			},
		},
	}
	d.databaseTEMP = databaseServiceStockTransaction.NewDatabaseService(dbParams)
}

func (d *DatabaseAccess) Disconnect() {
	d.databaseTEMP.Disconnect()
}

// Dirty methods for database connection.
func (d *DatabaseAccess) GetStockTransaction(transactionID string) transaction.StockTransactionInterface {
	StockTransaction, err := d.databaseTEMP.GetStockTransaction(transactionID)
	if err != nil {
		return nil
	}
	return StockTransaction

}

func (d *DatabaseAccess) GetStockTransactions(transactionIDs *[]string) *[]transaction.StockTransactionInterface {
	StockTransactions, err := d.databaseTEMP.GetStockTransactions(transactionIDs)
	if err != nil {
		return nil
	}
	return StockTransactions
}

// func (d *DatabaseAccess) GetInitialStockTransactionsForStock(stockID string) *[]transaction.StockTransactionInterface {
// 	StockTransactions, err := d.databaseTEMP.GetInitialStockTransactionsForStock(stockID)
// 	if err != nil {
// 		return nil
// 	}
// 	return StockTransactions
// }

func (d *DatabaseAccess) CreateStockTransaction(transaction transaction.StockTransactionInterface) transaction.StockTransactionInterface {
	StockTransaction, err := d.databaseTEMP.CreateStockTransaction(transaction)
	if err != nil {
		return nil
	}
	return StockTransaction
}

func (d *DatabaseAccess) UpdateStockTransaction(transaction transaction.StockTransactionInterface) transaction.StockTransactionInterface {
	StockTransaction, err := d.databaseTEMP.UpdateStockTransaction(transaction)
	if err != nil {
		return nil
	}
	return StockTransaction
}

func (d *DatabaseAccess) DeleteStockTransaction(transactionID string) transaction.StockTransactionInterface {
	StockTransaction, err := d.databaseTEMP.DeleteStockTransaction(transactionID)
	if err != nil {
		return nil
	}
	return StockTransaction
}
